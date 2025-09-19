package org

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ossf/scorecard/v5/clients/githubrepo/roundtripper"
	"github.com/ossf/scorecard/v5/log"
)

func TestParseOrgName(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want string
	}{
		{"http://github.com/owner", "owner"},
		{"https://github.com/owner", "owner"},
		{"github.com/owner", "owner"},
		{"owner", "owner"},
		{"", ""},
	}
	for _, c := range cases {
		if got := parseOrgName(c.in); got != c.want {
			t.Fatalf("parseOrgName(%q) = %q; want %q", c.in, got, c.want)
		}
	}
}

// Test ListOrgRepos handles pagination and filters archived repos.
func TestListOrgRepos_PaginationAndArchived(t *testing.T) {
	t.Parallel()
	// Single page: one archived repo and two active repos; expect active ones returned.
	body := `[
        {"html_url": "https://github.com/owner/repo1", "archived": true},
        {"html_url": "https://github.com/owner/repo2", "archived": false},
        {"html_url": "https://github.com/owner/repo3", "archived": false}
    ]`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	defer srv.Close()

	// Override TransportFactory to redirect requests to our test server.
	orig := roundtripper.TransportFactory
	roundtripper.TransportFactory = func(ctx context.Context, logger *log.Logger) http.RoundTripper {
		return roundTripperToServer(srv.URL)
	}
	defer func() { roundtripper.TransportFactory = orig }()

	repos, err := ListOrgRepos(context.Background(), "owner")
	if err != nil {
		t.Fatalf("ListOrgRepos returned error: %v", err)
	}
	// Expect repo2 and repo3 (repo1 archived)
	if len(repos) != 2 {
		t.Fatalf("expected 2 repos, got %d: %v", len(repos), repos)
	}
	if !strings.Contains(repos[0], "repo2") || !strings.Contains(repos[1], "repo3") {
		t.Fatalf("unexpected repos: %v", repos)
	}
}

// roundTripperToServer returns an http.RoundTripper that rewrites requests
// to the given serverURL, keeping the path and query intact.
func roundTripperToServer(serverURL string) http.RoundTripper {
	return http.RoundTripper(httpTransportFunc(func(req *http.Request) (*http.Response, error) {
		// rewrite target
		req.URL.Scheme = "http"
		req.URL.Host = strings.TrimPrefix(serverURL, "http://")
		return http.DefaultTransport.RoundTrip(req)
	}))
}

// httpTransportFunc converts a function into an http.RoundTripper.
type httpTransportFunc func(*http.Request) (*http.Response, error)

func (f httpTransportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
