// Copyright 2020 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package packagemanager implements a packagemanager client
package packagemanager

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ossf/scorecard/v5/clients/githubrepo/roundtripper/tokens"
)

func Test_GetURI_calls_client_get_with_input(t *testing.T) {
	t.Parallel()
	type args struct {
		inputURL string
	}
	tests := []struct {
		name         string
		args         args
		wantURL      string
		wantResponse string
	}{
		{
			name: "GetURI_input_is_the_same_as_get_uri",

			args: args{
				inputURL: "test",
			},
			wantURL:      "/test",
			wantResponse: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.wantURL {
					t.Errorf("Expected to request '%s', got: %s", tt.wantURL, r.URL.Path)
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.wantResponse)) //nolint:errcheck
			}))
			defer server.Close()
			client := PackageManagerClient{}
			got, err := client.GetURI(server.URL + "/" + tt.args.inputURL)
			if err != nil {
				t.Errorf("Test_GetURI_calls_client_get_with_input() error in Get= %v", err)
				return
			}
			defer got.Body.Close()
			body, err := io.ReadAll(got.Body)
			if err != nil {
				t.Errorf("Test_GetURI_calls_client_get_with_input() error in ReadAll= %v", err)
				return
			}
			if string(body) != tt.wantResponse {
				t.Errorf("GetURI() = %v, want %v", got, tt.wantResponse)
			}
		})
	}
}

func Test_Get_calls_client_get_with_input(t *testing.T) {
	t.Parallel()
	type args struct {
		inputURL    string
		packageName string
	}
	tests := []struct {
		name         string
		args         args
		wantURL      string
		wantResponse string
	}{
		{
			name: "Get_input_adds_package_name_for_get_uri",

			args: args{
				inputURL:    "test-%s-test",
				packageName: "test_package",
			},
			wantURL:      "/test-test_package-test",
			wantResponse: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.wantURL {
					t.Errorf("Expected to request '%s', got: %s", tt.wantURL, r.URL.Path)
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.wantResponse)) //nolint:errcheck
			}))
			defer server.Close()
			client := PackageManagerClient{}
			got, err := client.Get(server.URL+"/"+tt.args.inputURL, tt.args.packageName)
			if err != nil {
				t.Errorf("Test_Get_calls_client_get_with_input() error in Get = %v", err)
				return
			}
			defer got.Body.Close()
			body, err := io.ReadAll(got.Body)
			if err != nil {
				t.Errorf("Test_Get_calls_client_get_with_input() error in ReadAll = %v", err)
				return
			}
			if string(body) != tt.wantResponse {
				t.Errorf("GetURI() = %v, want %v", got, tt.wantResponse)
			}
		})
	}
}

func TestGetRemoteURLTokenOnlyForGitHubAPI(t *testing.T) {
	t.Setenv("GITHUB_AUTH_TOKEN", "secret")
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("Authorization")
	}))
	defer srv.Close()
	// URL contains the string "api.github.com" but the host is not GitHub.
	resp, err := (&PackageManagerClient{}).getRemoteURL(srv.URL + "/?x=api.github.com")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if got != "" {
		t.Errorf("token leaked to non-GitHub host: %q", got)
	}
}

type fakeAccessor struct {
	token    string
	released []uint64
}

func (a *fakeAccessor) Next() (uint64, string) { return 7, a.token }
func (a *fakeAccessor) Release(id uint64)      { a.released = append(a.released, id) }

type recordingTransport struct {
	req *http.Request
}

func (r *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r.req = req
	return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody, Request: req}, nil
}

func TestGitHubAuthTransport(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		url         string
		accessor    tokens.TokenAccessor
		wantAuth    string
		wantRelease bool
	}{
		{
			name:        "GitHub API over https gets the token",
			url:         "https://api.github.com/repos/microsoft/winget-pkgs/contents/manifests/g/Git/Git",
			accessor:    &fakeAccessor{token: "secret"},
			wantAuth:    "Bearer secret",
			wantRelease: true,
		},
		{
			name:     "other host gets no token",
			url:      "https://raw.githubusercontent.com/microsoft/winget-pkgs/master/manifests/g/Git/Git",
			accessor: &fakeAccessor{token: "secret"},
		},
		{
			name:     "plain http to the GitHub API gets no token",
			url:      "http://api.github.com/repos",
			accessor: &fakeAccessor{token: "secret"},
		},
		{
			name: "no tokens configured",
			url:  "https://api.github.com/repos",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			base := &recordingTransport{}
			transport := githubAuthTransport{
				base:      base,
				getTokens: func() tokens.TokenAccessor { return tt.accessor },
			}
			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, tt.url, nil)
			if err != nil {
				t.Fatal(err)
			}
			resp, err := transport.RoundTrip(req)
			if err != nil {
				t.Fatal(err)
			}
			resp.Body.Close()
			if got := base.req.Header.Get("Authorization"); got != tt.wantAuth {
				t.Errorf("Authorization = %q, want %q", got, tt.wantAuth)
			}
			if got := req.Header.Get("Authorization"); got != "" {
				t.Errorf("caller's request was modified: Authorization = %q", got)
			}
			if fa, ok := tt.accessor.(*fakeAccessor); ok {
				if released := len(fa.released) == 1 && fa.released[0] == 7; released != tt.wantRelease {
					t.Errorf("token released = %v (ids %v), want %v", released, fa.released, tt.wantRelease)
				}
			}
		})
	}
}
