// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package jwt

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/asaintsever/gloo-oss-extauth/pkg/auth/common"
	"github.com/asaintsever/gloo-oss-extauth/pkg/config"
	jwt_go "github.com/dgrijalva/jwt-go"
	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	extauth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v2"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type"
)

// Check ...
func (ctx *Context) Check() (*extauth.CheckResponse, error) {
	authHeader, ok := ctx.Headers["authorization"]
	var splitToken []string

	if ok {
		splitToken = strings.Split(authHeader, "Bearer ")
	}
	if len(splitToken) == 2 {
		token := splitToken[1]

		// Check it is a JWT and a valid one (issuer, audience, not expired, signature ok)
		parsedToken, err := jwt_go.Parse(token, ctx.validationKeyGetter)
		if err != nil || !parsedToken.Valid {
			if err == nil {
				log.Println("JWT is not valid")
			}
			return &extauth.CheckResponse{
				Status: common.DENY_STATUS,
				HttpResponse: &extauth.CheckResponse_DeniedResponse{
					DeniedResponse: &extauth.DeniedHttpResponse{
						Status: &envoy_type.HttpStatus{Code: envoy_type.StatusCode_Unauthorized},
						Body:   "PERMISSION_DENIED",
					},
				},
			}, err
		}

		var headersToAdd []*core.HeaderValueOption

		// Forward token to upstream service?
		jwtForward, _ := strconv.ParseBool(ctx.AuthServerCfg[config.JWT_FORWARD_TOKEN])
		if !jwtForward {
			headersToAdd = append(headersToAdd, &core.HeaderValueOption{Header: &core.HeaderValue{Key: "authorization", Value: ""}})
		}

		// Do we have claims to extract?
		if ctx.AuthServerCfg[config.JWT_EXTRACT_CLAIMS] != "" {
			claims := strings.Split(ctx.AuthServerCfg[config.JWT_EXTRACT_CLAIMS], ",")
			mapClaims := parsedToken.Claims.(jwt_go.MapClaims)
			for _, claim := range claims {
				jwtClaimValue, ok := mapClaims[claim].(string) // Expect and handle string
				if ok {
					headersToAdd = append(headersToAdd, &core.HeaderValueOption{Header: &core.HeaderValue{Key: "x-jwt-" + claim, Value: jwtClaimValue}})
				}
			}
		}

		return &extauth.CheckResponse{
			Status: common.ALLOW_STATUS,
			HttpResponse: &extauth.CheckResponse_OkResponse{
				OkResponse: &extauth.OkHttpResponse{Headers: headersToAdd},
			},
		}, nil
	}
	return &extauth.CheckResponse{
		Status: common.UNAUTHENTICATED_STATUS,
		HttpResponse: &extauth.CheckResponse_DeniedResponse{
			DeniedResponse: &extauth.DeniedHttpResponse{
				Status: &envoy_type.HttpStatus{Code: envoy_type.StatusCode_Unauthorized},
				Body:   "Authorization Header malformed or not provided",
			},
		},
	}, nil
}

func (ctx *Context) validationKeyGetter(token *jwt_go.Token) (interface{}, error) {
	// Verify 'aud' claim
	aud := ctx.AuthServerCfg[config.JWT_AUDIENCE]
	if aud != "" {
		// Do *not* use 'token.Claims.(jwt_go.MapClaims).VerifyAudience(...)' method has it does not deal with multiple audiences!!!! So fail to verify Auth0 token's audience
		if !verifyAudience(aud, token) {
			return token, errors.New("Invalid audience")
		}
	}

	// Verify 'iss' claim
	if !token.Claims.(jwt_go.MapClaims).VerifyIssuer(ctx.AuthServerCfg[config.JWT_ISSUER], true) {
		return token, errors.New("Invalid issuer")
	}

	// Get public key from JWKS
	pubKey, err := getKey(ctx.AuthServerCfg[config.JWKS_URL], token)
	if err != nil {
		return token, err
	}

	return pubKey, nil
}

func verifyAudience(audience string, token *jwt_go.Token) bool {
	bAudVerified := false
	mapClaims := token.Claims.(jwt_go.MapClaims)
	//log.Printf("**** mapClaims: %+v", mapClaims)

	jwtAud, ok := mapClaims["aud"].(string) // here we use type assertion (https://golang.org/ref/spec#Type_assertions)
	if ok {                                 // test if audience is a simple string
		if jwtAud == audience {
			bAudVerified = true
		}
	} else { // or if we have multiple audiences here (as Auth0 always add '/userinfo' endpoint audience in addition to ours)
		jwtAudArray, ok := mapClaims["aud"].([]interface{}) // here we use type assertion (https://golang.org/ref/spec#Type_assertions)

		if ok {
			for _, jwtAud := range jwtAudArray {
				jwtAudStr, ok := jwtAud.(string)
				if ok {
					if jwtAudStr == audience {
						bAudVerified = true
						break
					}
				}
			}
		}
	}

	return bAudVerified
}

func getKey(jwksURL string, token *jwt_go.Token) (*rsa.PublicKey, error) {
	//TODO: cache JWKS for a duration (to define as a new key/value in customAuth.contextExtensions)
	cert := ""
	resp, err := http.Get(jwksURL)

	if err != nil {
		log.Printf("Fail fetching JWKS from URL %s: %s", jwksURL, err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	var jwks = Jwks{}
	err = json.NewDecoder(resp.Body).Decode(&jwks)

	if err != nil {
		log.Printf("Fail decoding JWKS fetched from URL %s: %s", jwksURL, err.Error())
		return nil, err
	}

	// Caution: here we only consider JWKS with certificates. Improve code to handle all JWKS.
	for k := range jwks.Keys {
		if token.Header["kid"] == jwks.Keys[k].Kid {
			cert = "-----BEGIN CERTIFICATE-----\n" + jwks.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"
		}
	}

	if cert == "" {
		err := errors.New("Unable to find appropriate key")
		return nil, err
	}

	// Caution: here we only deal with RSA Pubkey. To improve.
	pubKey, _ := jwt_go.ParseRSAPublicKeyFromPEM([]byte(cert))
	return pubKey, nil
}
