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

package config

const (
	DEFAULT_SERVER_PORT = "8000"
)

const (
	// JWT validation & claims extraction
	JWT_ENABLED         = "jwt_validation"
	JWKS_URL            = "jwks_url"
	JWKS_CACHE_DURATION = "jwks_cache_duration" //TODO
	JWT_ISSUER          = "jwt_issuer"
	JWT_AUDIENCE        = "jwt_audience"
	JWT_FORWARD_TOKEN   = "jwt_forward"
	JWT_EXTRACT_CLAIMS  = "jwt_extract_claims"
)
