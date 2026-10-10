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

	resp, ok := h.responses[req.URL.Host]
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
