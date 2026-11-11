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

package provider

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sigs.k8s.io/release-utils/version"

	"github.com/ossf/scorecard/v5/cmd/internal/mcp/model"
)

const sampleResult = `{
  "date":"2024-01-02T03:04:05Z",
  "repo":{"name":"github.com/ossf/scorecard","commit":"64febf8c5229f0a5f0a6d2a0f0d4b6a7c8d9e0f1"},
  "scorecard":{"version":"v5.5.0","commit":"abc123"},
  "score":8.7,
  "checks":[
    {"name":"Code-Review","score":10,"reason":"all good","details":["d1"],"documentation":{"short":"cr","url":"http://x"}},
    {"name":"Fuzzing","score":-1,"reason":"inconclusive"}
  ],
  "metadata":null
}`

func TestCachedRESTProviderGetResult(t *testing.T) {
	t.Parallel()
	const commit = "64febf8c5229f0a5f0a6d2a0f0d4b6a7c8d9e0f1"
	var gotPath, gotQuery, gotUserAgent string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotQuery, gotUserAgent = r.URL.Path, r.URL.RawQuery, r.UserAgent()
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(sampleResult)); err != nil {
			t.Errorf("writing response: %v", err)
		}
	}))
	defer ts.Close()

	p := NewCachedREST(ts.URL)
	ref := model.RepoRef{Platform: "github.com", Org: "ossf", Name: "scorecard"}
	res, err := p.GetResult(context.Background(), ref, commit)
	if err != nil {
		t.Fatalf("GetResult: %v", err)
	}

	if gotPath != "/projects/github.com/ossf/scorecard" {
		t.Errorf("path = %q", gotPath)
	}
	if gotQuery != "commit="+commit {
		t.Errorf("query = %q", gotQuery)
	}
	if want := "scorecard/" + version.GetVersionInfo().GitVersion; gotUserAgent != want {
		t.Errorf("User-Agent = %q, want %q", gotUserAgent, want)
	}
	if res.Source != model.SourceCachedREST {
		t.Errorf("source = %q, want cached-rest", res.Source)
	}
	if res.Score != 8.7 {
		t.Errorf("score = %v, want 8.7", res.Score)
	}
	if res.Commit != commit {
		t.Errorf("commit = %q, want %q", res.Commit, commit)
	}
	if len(res.Checks) != 2 {
		t.Fatalf("checks = %d, want 2", len(res.Checks))
	}
	if res.Checks[1].Score != model.InconclusiveScore {
		t.Errorf("Fuzzing score = %d, want -1 (inconclusive)", res.Checks[1].Score)
	}
	if len(res.Caveats) == 0 {
		t.Error("expected cached-data caveats")
	}
	if res.Attribution == nil || res.Attribution.DataLicense != "CDLA-Permissive-2.0" {
		t.Errorf("attribution = %+v, want CDLA-Permissive-2.0", res.Attribution)
	}
}

func TestCachedRESTProviderTrimsTrailingSlash(t *testing.T) {
	t.Parallel()
	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(sampleResult)); err != nil {
			t.Errorf("writing response: %v", err)
		}
	}))
	defer ts.Close()

	p := NewCachedREST(ts.URL + "/")
	if p.BaseURL != ts.URL {
		t.Errorf("BaseURL = %q, want %q", p.BaseURL, ts.URL)
	}

	ref := model.RepoRef{Platform: "github.com", Org: "ossf", Name: "scorecard"}
	if _, err := p.GetResult(context.Background(), ref, ""); err != nil {
		t.Fatalf("GetResult: %v", err)
	}
	if gotPath != "/projects/github.com/ossf/scorecard" {
		t.Errorf("path = %q, want no doubled slash", gotPath)
	}
}

func TestCachedRESTProviderNotFound(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	p := NewCachedREST(ts.URL)
	_, err := p.GetResult(context.Background(), model.RepoRef{Platform: "github.com", Org: "o", Name: "r"}, "")
	var nf *NotFoundError
	if !errors.As(err, &nf) {
		t.Fatalf("expected *NotFoundError, got %v", err)
	}
}

func TestCachedRESTProviderBadRequest(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	p := NewCachedREST(ts.URL)
	if _, err := p.GetResult(context.Background(), model.RepoRef{Platform: "github.com", Org: "o", Name: "r"}, ""); err == nil {
		t.Fatal("expected an error on HTTP 400")
	}
}

func TestNewCachedRESTDefaultBaseURL(t *testing.T) {
	t.Parallel()
	if got := NewCachedREST("").BaseURL; got != DefaultRESTBaseURL {
		t.Errorf("BaseURL = %q, want %q", got, DefaultRESTBaseURL)
	}
}

func TestNotFoundErrorMessage(t *testing.T) {
	t.Parallel()
	err := &NotFoundError{Ref: model.RepoRef{Platform: "github.com", Org: "o", Name: "r"}}
	msg := err.Error()
	for _, want := range []string{"github.com/o/r", "publish_results: true", "`scorecard --repo=github.com/o/r`"} {
		if !strings.Contains(msg, want) {
			t.Errorf("Error() = %q, want it to contain %q", msg, want)
		}
	}
	if strings.Contains(msg, "future") {
		t.Errorf("Error() = %q, want no promise of future features", msg)
	}
}

func TestCachedRESTProviderErrors(t *testing.T) {
	t.Parallel()
	closed := httptest.NewServer(http.NotFoundHandler())
	closed.Close()

	tests := []struct {
		handler http.HandlerFunc
		wantErr error
		name    string
		baseURL string
		want    string
	}{
		{
			name: "unexpected status",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantErr: errUnexpectedStatus,
			want:    "500",
		},
		{
			name: "invalid JSON",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				if _, err := w.Write([]byte("{")); err != nil {
					t.Errorf("writing response: %v", err)
				}
			},
			want: "decoding Scorecard results",
		},
		{
			name:    "unreachable server",
			baseURL: closed.URL,
			want:    "requesting Scorecard results",
		},
		{
			name:    "invalid base URL",
			baseURL: "http://invalid\x7f",
			want:    "building request",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			baseURL := tt.baseURL
			if tt.handler != nil {
				ts := httptest.NewServer(tt.handler)
				defer ts.Close()
				baseURL = ts.URL
			}
			_, err := NewCachedREST(baseURL).GetResult(context.Background(),
				model.RepoRef{Platform: "github.com", Org: "o", Name: "r"}, "")
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("GetResult error = %v, want one containing %q", err, tt.want)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("GetResult error = %v, want errors.Is %v", err, tt.wantErr)
			}
		})
	}
}

func TestCachedRESTProviderFallsBackToRequestedCommit(t *testing.T) {
	t.Parallel()
	const commit = "0123456789abcdef0123456789abcdef01234567"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write([]byte(`{"repo":{"name":"github.com/o/r"},"score":5,"checks":[]}`)); err != nil {
			t.Errorf("writing response: %v", err)
		}
	}))
	defer ts.Close()

	res, err := NewCachedREST(ts.URL).GetResult(context.Background(),
		model.RepoRef{Platform: "github.com", Org: "o", Name: "r"}, commit)
	if err != nil {
		t.Fatalf("GetResult: %v", err)
	}
	if res.Commit != commit {
		t.Errorf("Commit = %q, want the requested %q when the API omits it", res.Commit, commit)
	}
}
