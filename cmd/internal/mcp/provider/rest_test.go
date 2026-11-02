package provider

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

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
	var gotPath, gotQuery string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotQuery = r.URL.Path, r.URL.RawQuery
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
	if res.Complete {
		t.Error("cached result should be marked incomplete")
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
