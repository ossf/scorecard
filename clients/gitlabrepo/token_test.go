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
	"sync/atomic"
	"testing"

	sce "github.com/ossf/scorecard/v5/errors"
)

const testToken = "glpat-test-token"

func TestTokenForHost(t *testing.T) {
	tests := []struct {
		name   string
		glHost string
		host   string
		want   string
	}{
		{name: "gitlab.com by default", host: "gitlab.com", want: testToken},
		{name: "case-insensitive", host: "GitLab.com", want: testToken},
		{name: "other gitlab instance", host: "gitlab.example.com", want: ""},
		{name: "gitlab.com as a prefix", host: "gitlab.com.example.net", want: ""},
		{name: "gitlab.com as a suffix", host: "evilgitlab.com", want: ""},
		{name: "gitlab.com with a port", host: "gitlab.com:8443", want: ""},
		{name: "GL_HOST", glHost: "gitlab.example.com", host: "gitlab.example.com", want: testToken},
		{name: "GL_HOST excludes gitlab.com", glHost: "gitlab.example.com", host: "gitlab.com", want: ""},
		{name: "GL_HOST with scheme and path", glHost: "https://example.com/gitlab/", host: "example.com/gitlab", want: testToken},
		{name: "GL_HOST path required", glHost: "https://example.com/gitlab/", host: "example.com", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GITLAB_AUTH_TOKEN", testToken)
			t.Setenv("GL_HOST", tt.glHost)
			if got := tokenForHost(tt.host); got != tt.want {
				t.Errorf("tokenForHost(%q) = %q, want %q", tt.host, got, tt.want)
			}
		})
	}
}

func TestApiFunctionTokenScopedToHost(t *testing.T) {
	tests := []struct {
		name   string
		glHost string
		host   string
		want   string
	}{
		{name: "token host", host: "gitlab.com", want: testToken},
		{name: "other host", host: "gitlab.example.com", want: ""},
		{name: "gitlab.com when GL_HOST is set", glHost: "gitlab.example.com", host: "gitlab.com", want: ""},
		{name: "GL_HOST", glHost: "gitlab.example.com", host: "gitlab.example.com", want: testToken},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GITLAB_AUTH_TOKEN", testToken)
			t.Setenv("GL_HOST", tt.glHost)

			var got string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = r.Header.Get("PRIVATE-TOKEN")
			}))
			defer srv.Close()

			tempDir := t.TempDir()
			repoFile, err := os.CreateTemp(tempDir, repoFilename)
			if err != nil {
				t.Fatalf("os.CreateTemp: %v", err)
			}
			defer repoFile.Close()

			handler := tarballHandler{ctx: context.Background(), once: new(sync.Once), repourl: &Repo{host: tt.host}}
			if err := handler.apiFunction(srv.URL, tempDir, repoFile); err != nil {
				t.Fatalf("apiFunction: %v", err)
			}
			if got != tt.want {
				t.Errorf("PRIVATE-TOKEN = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGraphqlTokenScopedToHost(t *testing.T) {
	tests := []struct {
		name string
		host string
		want string
	}{
		{name: "token host", host: "gitlab.com", want: "Bearer " + testToken},
		{name: "other host", host: "gitlab.example.com", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GITLAB_AUTH_TOKEN", testToken)
			t.Setenv("GL_HOST", "")

			var got string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = r.Header.Get("Authorization")
			}))
			defer srv.Close()

			var handler graphqlHandler
			handler.init(context.Background(), &Repo{host: tt.host})
			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
			if err != nil {
				t.Fatalf("http.NewRequestWithContext: %v", err)
			}
			resp, err := handler.client.Do(req)
			if err != nil {
				t.Fatalf("client.Do: %v", err)
			}
			resp.Body.Close()
			if got != tt.want {
				t.Errorf("Authorization = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsValidRequiresHTTPS(t *testing.T) {
	t.Parallel()
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
	}))
	defer srv.Close()

	for _, input := range []string{
		srv.URL + "/owner/project",
		"http://gitlab.example.com/owner/project",
	} {
		var r Repo
		if err := r.parse(input); err != nil {
			t.Fatalf("parse(%q): %v", input, err)
		}
		if err := r.IsValid(); !errors.Is(err, sce.ErrInvalidURL) {
			t.Errorf("IsValid(%q) = %v, want %v", input, err, sce.ErrInvalidURL)
		}
	}
	if n := hits.Load(); n != 0 {
		t.Errorf("instance check contacted the http host %d times, want 0", n)
	}

	var r Repo
	if err := r.parse("https://gitlab.example.com/owner/project"); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := r.IsValid(); err != nil {
		t.Errorf("IsValid(https gitlab.example.com) = %v, want nil", err)
	}
}
