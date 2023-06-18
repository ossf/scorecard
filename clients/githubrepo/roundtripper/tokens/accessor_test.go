// Copyright 2023 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package tokens

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/rpc"
	"testing"
)

//nolint:paralleltest
func TestMakeTokenAccessor(t *testing.T) {
	tests := []struct {
		name           string
		useGitHubToken bool
		useServer      bool
	}{
		{
			name:           "GitHub Token",
			useGitHubToken: true,
		},
		{
			name:           "No GitHub Token",
			useGitHubToken: false,
		},
		{
			name:      "Server",
			useServer: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch {
			case tt.useGitHubToken:
				token := "test"
				t.Setenv("GITHUB_TOKEN", token)
				got := MakeTokenAccessor()
				if got == nil {
					t.Errorf("MakeTokenAccessor() = nil, want not nil")
				}
				raccess, ok := got.(*roundRobinAccessor)
				if !ok {
					t.Errorf("MakeTokenAccessor() = %v, want *roundRobinAccessor", got)
				}
				if raccess.accessTokens[0] != token {
					t.Errorf("accessTokens[0] = %v, want %v", raccess.accessTokens[0], token)
				}
			case tt.useServer:
				t.Setenv("GITHUB_AUTH_SERVER", "localhost:8080")
				server := startTestServer()
				defer serverShutdown(server)
				myRPCService := &MyRPCService{}
				rpc.Register(myRPCService) //nolint:errcheck
				server.Handler = nil
				rpc.HandleHTTP()
				got := MakeTokenAccessor()
				if got == nil {
					t.Errorf("MakeTokenAccessor() = nil, want not nil")
				}
			default:
				got := MakeTokenAccessor()
				if got != nil {
					t.Errorf("MakeTokenAccessor() = %v, want nil", got)
				}
			}
		})
	}
}

type MyRPCService struct {
	// Define your RPC service methods here
}

func startTestServer() *http.Server {
	// Create a new server
	server := &http.Server{ //nolint:gosec
		Addr:    ":8080", // Use any available port
		Handler: nil,     // Use the default handler
	}

	// Start the server in a separate goroutine
	go func() {
		fmt.Println("Starting server on http://localhost:8080")
		if err := server.ListenAndServe(); err != nil && !errors.Is(http.ErrServerClosed, err) {
			log.Fatal(err)
		}
	}()

	return server
}

func serverShutdown(server *http.Server) {
	err := server.Close()
	if err != nil {
		log.Fatalf("Error shutting down server: %s\n", err.Error())
	}

	fmt.Println("Server gracefully stopped")
}
