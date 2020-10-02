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

package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/asaintsever/gloo-oss-extauth/pkg/auth"
	"github.com/asaintsever/gloo-oss-extauth/pkg/config"
	extauth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// create a TCP listener
	lis, err := net.Listen("tcp", ":"+config.GetPort())
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	extauth.RegisterAuthorizationServer(grpcServer, &auth.Server{})

	// Optional, useful to be able to test the service using grpcurl for eg
	reflection.Register(grpcServer)

	// Listening OS shutdown singal
	signalChan := make(chan os.Signal, 1)

	// Start GRPC server
	go func() {
		log.Printf("Starting Custom Auth server on %s ...", lis.Addr().String())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
			signalChan <- syscall.SIGINT
		}
	}()

	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	grpcServer.GracefulStop()
	log.Println("Custom Auth server stopped")
}
