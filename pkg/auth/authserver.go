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

package auth

import (
	"context"
	"log"
	"strconv"

	"github.com/asaintsever/gloo-oss-extauth/pkg/auth/common"
	"github.com/asaintsever/gloo-oss-extauth/pkg/auth/jwt"
	"github.com/asaintsever/gloo-oss-extauth/pkg/config"
	extauth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v2"
)

// Server is used to implement extauth.AuthorizationServer interface:
// - https://github.com/envoyproxy/data-plane-api/blob/master/envoy/service/auth/v2/external_auth.proto
// - https://github.com/envoyproxy/go-control-plane/blob/master/envoy/service/auth/v2/external_auth.pb.go
type Server struct{}

// Check ...
func (srv *Server) Check(ctx context.Context, in *extauth.CheckRequest) (*extauth.CheckResponse, error) {
	log.Println(">>> check incoming request")
	var err error

	checkResponse := &extauth.CheckResponse{Status: common.ALLOW_STATUS}
	ctxExtensions := in.GetAttributes().GetContextExtensions()
	headers := in.GetAttributes().GetRequest().GetHttp().GetHeaders()

	// For info & debug
	log.Printf("ctxExtensions: %+v", ctxExtensions)
	log.Printf("headers: %+v", headers)

	jwtValidationEnabled, _ := strconv.ParseBool(ctxExtensions[config.JWT_ENABLED])
	if jwtValidationEnabled {
		log.Println("Ask for JWT validation")
		checkResponse, err = (&jwt.Context{ctxExtensions, headers}).Check()

		if err != nil {
			log.Printf("Error: %s", err.Error())
		}
	}

	return checkResponse, err
}
