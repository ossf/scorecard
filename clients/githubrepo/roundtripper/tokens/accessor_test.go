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
	"net/http/httptest"
	"net/rpc"
	"os"
	"strings"
	"testing"
)

//nolint:paralleltest // test uses t.Setenv indirectly
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
	unsetTokens(t)
	t.Setenv(githubAuthServer, "")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch {
			case tt.useGitHubToken:
				testToken(t)
			case tt.useServer:
				testServer(t)
			default:
				got := MakeTokenAccessor()
				if got != nil {
					t.Errorf("MakeTokenAccessor() = %v, want nil", got)
				}
			}
		})
	}
}

func testToken(t *testing.T) {
	t.Helper()
	token := "test"
	t.Setenv("GITHUB_AUTH_TOKEN", token)
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
}

func testServer(t *testing.T) {
	t.Helper()
	server := httptest.NewServer(nil)
	serverURL := strings.TrimPrefix(server.URL, "http://")
	t.Setenv("GITHUB_AUTH_SERVER", serverURL)
	t.Cleanup(server.Close)
	rpc.HandleHTTP()
	got := MakeTokenAccessor()
	if got == nil {
		t.Errorf("MakeTokenAccessor() = nil, want not nil")
	}
}

//nolint:paralleltest // test uses t.Setenv indirectly
func TestClashingTokensDisplayWarning(t *testing.T) {
	unsetTokens(t)

	someToken := "test_token"
	otherToken := "clashing_token"
	t.Setenv("GITHUB_AUTH_TOKEN", someToken)
	t.Setenv("GITHUB_TOKEN", otherToken)

	warningCalled := false
	originalLogWarning := logDuplicateTokenWarning
	logDuplicateTokenWarning = func(firstName string, clashingName string) {
		warningCalled = true
	}
	defer func() { logDuplicateTokenWarning = originalLogWarning }()

	token, exists := readGitHubTokens()

	if token != someToken {
		t.Errorf("Received wrong token")
	}
	if !exists {
		t.Errorf("Token is expected to exist")
	}
	if !warningCalled {
		t.Errorf("Expected logWarning to be called for clashing tokens, but it was not.")
	}
}

//nolint:paralleltest // test uses t.Setenv indirectly
func TestConsistentTokensDoNotDisplayWarning(t *testing.T) {
	unsetTokens(t)

	someToken := "test_token"
	t.Setenv("GITHUB_AUTH_TOKEN", someToken)
	t.Setenv("GITHUB_TOKEN", someToken)

	warningCalled := false
	originalLogWarning := logDuplicateTokenWarning
	logDuplicateTokenWarning = func(firstName string, clashingName string) {
		warningCalled = true
	}
	defer func() { logDuplicateTokenWarning = originalLogWarning }()

	token, exists := readGitHubTokens()

	if token != someToken {
		t.Errorf("Received wrong token")
	}
	if !exists {
		t.Errorf("Token is expected to exist")
	}
	if warningCalled {
		t.Errorf("Expected logWarning to not have been called for consistent tokens, but it was.")
	}
}

//nolint:paralleltest // test uses t.Setenv indirectly
func TestNoTokensDoNoDisplayWarning(t *testing.T) {
	unsetTokens(t)

	warningCalled := false
	originalLogWarning := logDuplicateTokenWarning
	logDuplicateTokenWarning = func(firstName string, clashingName string) {
		warningCalled = true
	}
	defer func() { logDuplicateTokenWarning = originalLogWarning }()

	token, exists := readGitHubTokens()

	if token != "" {
		t.Errorf("Scorecard found a token somewhere")
	}
	if exists {
		t.Errorf("Token is not expected to exist")
	}
	if warningCalled {
		t.Errorf("Expected logWarning to not have been called for no set tokens, but it was not.")
	}
}

// temporarily unset all of the github token env vars,
// as tests may otherwise fail depending on the local environment.
func unsetTokens(t *testing.T) {
	t.Helper()
	for _, name := range githubAuthTokenEnvVars {
		// equivalent to t.Unsetenv (which does not exist)
		t.Setenv(name, "")
		os.Unsetenv(name)
	}
}
