// Copyright 2026 OpenSSF Scorecard Authors
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

package gitlabrepo

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

func TestApiFunctionTokenNotForwardedOnRedirect(t *testing.T) {
	const token = "super-secret-pat"
	t.Setenv("GITLAB_AUTH_TOKEN", token)

	var forwarded string
	// Stands in for the redirect target (object storage / CDN / attacker host).
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		forwarded = r.Header.Get("PRIVATE-TOKEN")
		w.WriteHeader(http.StatusOK)
	}))
	defer other.Close()

	// Stands in for the GitLab instance; redirects the archive request elsewhere.
	instance := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, other.URL, http.StatusFound)
	}))
	defer instance.Close()

	tempDir := t.TempDir()
	repoFile, err := os.CreateTemp(tempDir, repoFilename)
	if err != nil {
		t.Fatalf("os.CreateTemp: %v", err)
	}
	defer repoFile.Close()

	handler := tarballHandler{ctx: context.Background(), once: new(sync.Once)}
	if err := handler.apiFunction(instance.URL, tempDir, repoFile); err != nil {
		t.Fatalf("apiFunction: %v", err)
	}

	if forwarded != "" {
		t.Fatalf("PRIVATE-TOKEN forwarded to redirect host: got %q, want it dropped", forwarded)
	}
}

func TestCheckTokenRedirect(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		from      string
		to        string
		wantToken bool
	}{
		{name: "same host and scheme", from: "https://gitlab.com/a", to: "https://gitlab.com/b", wantToken: true},
		{name: "cross host", from: "https://gitlab.com/a", to: "https://cdn.example.com/b", wantToken: false},
		{name: "same host downgrade to http", from: "https://gitlab.com/a", to: "http://gitlab.com/b", wantToken: false},
		{name: "same host different port", from: "https://gitlab.com/a", to: "https://gitlab.com:8443/b", wantToken: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			first, err := http.NewRequestWithContext(context.Background(), http.MethodGet, tt.from, nil)
			if err != nil {
				t.Fatalf("http.NewRequestWithContext: %v", err)
			}
			next, err := http.NewRequestWithContext(context.Background(), http.MethodGet, tt.to, nil)
			if err != nil {
				t.Fatalf("http.NewRequestWithContext: %v", err)
			}
			next.Header.Set("PRIVATE-TOKEN", "secret")
			if err := checkTokenRedirect(next, []*http.Request{first}); err != nil {
				t.Fatalf("checkTokenRedirect: %v", err)
			}
			if got := next.Header.Get("PRIVATE-TOKEN") != ""; got != tt.wantToken {
				t.Errorf("token kept = %v, want %v", got, tt.wantToken)
			}
		})
	}
}

func TestCheckTokenRedirectLimit(t *testing.T) {
	t.Parallel()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://gitlab.com/a", nil)
	if err != nil {
		t.Fatalf("http.NewRequestWithContext: %v", err)
	}
	via := make([]*http.Request, 10)
	for i := range via {
		via[i] = req
	}
	if err := checkTokenRedirect(req, via); !errors.Is(err, errTooManyRedirects) {
		t.Errorf("checkTokenRedirect: got %v, want %v", err, errTooManyRedirects)
	}
	if err := checkTokenRedirect(req, via[:9]); err != nil {
		t.Errorf("checkTokenRedirect with 9 redirects: %v", err)
	}
}
