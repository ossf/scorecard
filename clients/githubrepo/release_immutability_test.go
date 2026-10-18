// Copyright 2025 OpenSSF Scorecard Authors
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

package githubrepo

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-github/v82/github"
)

// hostRoundTripper routes requests to canned responses based on the
// request's host, simulating a GitHub Enterprise Server (ghes) and
// github.com (dotcom) that disagree about whether a release exists/is
// immutable.
type hostRoundTripper struct {
	// responses maps a host (req.URL.Host) to a status code and body to
	// return for any request to that host.
	responses map[string]hostResponse
	// pathResponses maps host+path to a response, taking precedence over responses.
	pathResponses map[string]hostResponse
	// calls counts the number of requests made per host.
	calls map[string]int
}

type hostResponse struct {
	status int
	body   string
}

func (h *hostRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if h.calls == nil {
		h.calls = map[string]int{}
	}
	h.calls[req.URL.Host]++

	resp, ok := h.pathResponses[req.URL.Host+req.URL.Path]
	if !ok {
		resp, ok = h.responses[req.URL.Host]
	}
	if !ok {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Status:     "404 Not Found",
			Body:       io.NopCloser(strings.NewReader(`{"message":"Not Found"}`)),
			Header:     http.Header{},
		}, nil
	}
	return &http.Response{
		StatusCode: resp.status,
		Status:     fmt.Sprintf("%d", resp.status),
		Body:       io.NopCloser(strings.NewReader(resp.body)),
		Header:     http.Header{},
	}, nil
}

func TestIsReleaseImmutable_GitHubConnectFallback(t *testing.T) {
	t.Parallel()

	const (
		ghesHost   = "ghes.example.com"
		dotcomHost = "api.github.com"
	)

	immutableBody := `{"tag_name":"v1.0.0","immutable":true}`
	notFoundBody := `{"message":"Not Found"}`

	tests := []struct {
		name          string
		ghesResponse  hostResponse
		wantImmutable bool
		wantErr       bool
		wantDotcomHit bool
	}{
		{
			name:          "found and immutable on enterprise server",
			ghesResponse:  hostResponse{status: http.StatusOK, body: immutableBody},
			wantImmutable: true,
			wantDotcomHit: false,
		},
		{
			name:          "not found on enterprise server falls back to dotcom",
			ghesResponse:  hostResponse{status: http.StatusNotFound, body: notFoundBody},
			wantImmutable: true, // dotcom response below is immutable
			wantDotcomHit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rt := &hostRoundTripper{
				responses: map[string]hostResponse{
					ghesHost:   tt.ghesResponse,
					dotcomHost: {status: http.StatusOK, body: immutableBody},
				},
			}
			httpClient := &http.Client{Transport: rt}

			enterpriseClient, err := github.NewClient(httpClient).WithEnterpriseURLs(
				"https://"+ghesHost+"/api/v3", "https://"+ghesHost+"/api/v3")
			if err != nil {
				t.Fatalf("WithEnterpriseURLs: %v", err)
			}
			dotcomClient := github.NewClient(httpClient)

			client := &Client{
				ctx:          t.Context(),
				repoClient:   enterpriseClient,
				dotcomClient: dotcomClient,
			}

			immutable, err := client.IsReleaseImmutable("owner", "repo", "v1.0.0")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if immutable != tt.wantImmutable {
				t.Errorf("IsReleaseImmutable() = %v, want %v", immutable, tt.wantImmutable)
			}

			gotDotcomHit := rt.calls[dotcomHost] > 0
			if gotDotcomHit != tt.wantDotcomHit {
				t.Errorf("dotcom fallback called = %v, want %v", gotDotcomHit, tt.wantDotcomHit)
			}
		})
	}
}

func TestIsReleaseImmutable_NoDotcomFallbackConfigured(t *testing.T) {
	t.Parallel()

	rt := &hostRoundTripper{
		responses: map[string]hostResponse{
			"api.github.com": {status: http.StatusNotFound, body: `{"message":"Not Found"}`},
		},
	}
	client := &Client{
		ctx:        t.Context(),
		repoClient: github.NewClient(&http.Client{Transport: rt}),
		// dotcomClient intentionally left nil: non-enterprise configuration.
	}

	immutable, err := client.IsReleaseImmutable("owner", "repo", "v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if immutable {
		t.Error("expected immutable=false when release isn't found and no fallback is configured")
	}
}

func TestIsReleaseImmutable_FallbackDecisions(t *testing.T) {
	t.Parallel()

	const (
		ghesHost   = "ghes.example.com"
		dotcomHost = "api.github.com"
		releasePth = "/repos/owner/repo/releases/tags/v1.0.0"
		repoPath   = "/repos/owner/repo"
	)
	ok := func(body string) hostResponse { return hostResponse{status: http.StatusOK, body: body} }
	status := func(code int) hostResponse { return hostResponse{status: code, body: `{"message":"x"}`} }
	immutableBody := `{"tag_name":"v1.0.0","immutable":true}`

	tests := []struct {
		name          string
		ghesRelease   hostResponse
		ghesRepo      hostResponse
		dotcomRelease hostResponse
		wantImmutable bool
		wantErr       bool
		wantDotcomHit bool
	}{
		{
			name:        "enterprise release lookup failure is an error without fallback",
			ghesRelease: status(http.StatusInternalServerError),
			ghesRepo:    ok(`{}`),
			wantErr:     true,
		},
		{
			name:          "release missing on an existing enterprise repo never falls back to dotcom",
			ghesRelease:   status(http.StatusNotFound),
			ghesRepo:      ok(`{"name":"repo"}`),
			dotcomRelease: ok(immutableBody),
			wantImmutable: false,
			wantDotcomHit: false,
		},
		{
			name:          "enterprise repo lookup failure is an error without fallback",
			ghesRelease:   status(http.StatusNotFound),
			ghesRepo:      status(http.StatusInternalServerError),
			dotcomRelease: ok(immutableBody),
			wantErr:       true,
		},
		{
			name:          "dotcom failure is an error",
			ghesRelease:   status(http.StatusNotFound),
			ghesRepo:      status(http.StatusNotFound),
			dotcomRelease: status(http.StatusInternalServerError),
			wantErr:       true,
			wantDotcomHit: true,
		},
		{
			name:          "missing on both hosts is not immutable",
			ghesRelease:   status(http.StatusNotFound),
			ghesRepo:      status(http.StatusNotFound),
			dotcomRelease: status(http.StatusNotFound),
			wantDotcomHit: true,
		},
		{
			name:          "repo absent on enterprise and release immutable on dotcom",
			ghesRelease:   status(http.StatusNotFound),
			ghesRepo:      status(http.StatusNotFound),
			dotcomRelease: ok(immutableBody),
			wantImmutable: true,
			wantDotcomHit: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rt := &hostRoundTripper{
				pathResponses: map[string]hostResponse{
					ghesHost + "/api/v3" + releasePth: tt.ghesRelease,
					ghesHost + "/api/v3" + repoPath:   tt.ghesRepo,
					dotcomHost + releasePth:           tt.dotcomRelease,
				},
			}
			httpClient := &http.Client{Transport: rt}
			enterpriseClient, err := github.NewClient(httpClient).WithEnterpriseURLs(
				"https://"+ghesHost+"/api/v3", "https://"+ghesHost+"/api/v3")
			if err != nil {
				t.Fatalf("WithEnterpriseURLs: %v", err)
			}
			client := &Client{
				ctx:          t.Context(),
				repoClient:   enterpriseClient,
				dotcomClient: github.NewClient(httpClient),
			}

			got, err := client.IsReleaseImmutable("owner", "repo", "v1.0.0")
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.wantImmutable {
				t.Errorf("immutable = %v, want %v", got, tt.wantImmutable)
			}
			if hit := rt.calls[dotcomHost] > 0; hit != tt.wantDotcomHit {
				t.Errorf("dotcom hit = %v, want %v", hit, tt.wantDotcomHit)
			}
		})
	}
}

// authHeaderRecorder records the Authorization header of each request per host, and answers 404.
type authHeaderRecorder struct {
	auth map[string][]string
}

func (r *authHeaderRecorder) RoundTrip(req *http.Request) (*http.Response, error) {
	r.auth[req.URL.Host] = append(r.auth[req.URL.Host], req.Header.Get("Authorization"))
	return &http.Response{
		StatusCode: http.StatusNotFound,
		Status:     "404 Not Found",
		Body:       io.NopCloser(strings.NewReader(`{"message":"Not Found"}`)),
		Header:     http.Header{},
		Request:    req,
	}, nil
}

// authInjector mimics the configured transport by adding credentials to every request.
type authInjector struct{ next http.RoundTripper }

func (a authInjector) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer enterprise-secret")
	return a.next.RoundTrip(req)
}

//nolint:paralleltest // uses t.Setenv
func TestIsReleaseImmutable_DotcomFallbackNeverSendsEnterpriseCredentials(t *testing.T) {
	t.Setenv("GH_HOST", "ghes.example.com")

	enterprise := &authHeaderRecorder{auth: map[string][]string{}}
	dotcom := &authHeaderRecorder{auth: map[string][]string{}}
	rc, err := NewRepoClient(t.Context(),
		WithRoundTripper(authInjector{next: enterprise}),
		WithDotcomRoundTripper(dotcom),
	)
	if err != nil {
		t.Fatalf("NewRepoClient: %v", err)
	}

	// 404 on both hosts forces the lookup through the fallback path.
	if immutable, err := rc.IsReleaseImmutable("owner", "repo", "v1"); err != nil || immutable {
		t.Fatalf("IsReleaseImmutable = %v, %v; want false, nil", immutable, err)
	}

	if got := enterprise.auth["ghes.example.com"]; len(got) == 0 || got[0] != "Bearer enterprise-secret" {
		t.Errorf("enterprise requests should carry credentials, got %v", got)
	}
	if len(dotcom.auth["api.github.com"]) == 0 {
		t.Fatal("expected the lookup to fall back to github.com")
	}
	for host, headers := range dotcom.auth {
		for _, h := range headers {
			if h != "" {
				t.Errorf("request to %s carried credentials: %q", host, h)
			}
		}
	}
	if len(enterprise.auth["api.github.com"]) != 0 {
		t.Error("github.com requests must not go through the enterprise transport")
	}
}

//nolint:paralleltest // uses t.Setenv
func TestNewRepoClient_NoDotcomFallbackOnGithubCom(t *testing.T) {
	t.Setenv("GH_HOST", "github.com")
	rc, err := NewRepoClient(t.Context())
	if err != nil {
		t.Fatalf("NewRepoClient: %v", err)
	}
	if c, ok := rc.(*Client); !ok || c.dotcomClient != nil {
		t.Fatalf("expected no dotcom fallback client, got %+v", rc)
	}
}
