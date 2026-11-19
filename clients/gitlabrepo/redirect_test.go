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
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// redirectPair starts an origin server that redirects every request to a
// second server, and returns the origin URL and a func reporting the
// credential headers the second server received.
func redirectPair(t *testing.T) (string, func() (auth, privateToken string)) {
	t.Helper()
	var auth, privateToken string
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		privateToken = r.Header.Get("PRIVATE-TOKEN")
	}))
	t.Cleanup(other.Close)
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, other.URL, http.StatusFound)
	}))
	t.Cleanup(origin.Close)
	return origin.URL, func() (string, string) { return auth, privateToken }
}

func TestRESTClientTokenNotForwardedOnRedirect(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		opts     []gitlab.ClientOptionFunc
		wantLeak bool
	}{
		{name: "client-go default client", wantLeak: true},
		{name: "newHTTPClient", opts: []gitlab.ClientOptionFunc{gitlab.WithHTTPClient(newHTTPClient())}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			originURL, received := redirectPair(t)
			opts := append([]gitlab.ClientOptionFunc{gitlab.WithBaseURL(originURL)}, tt.opts...)
			client, err := gitlab.NewClient("secret", opts...)
			if err != nil {
				t.Fatal(err)
			}
			// The redirect target returns an empty body, so the decode error is expected.
			//nolint:errcheck
			client.Projects.GetProject("owner/project", nil)

			_, privateToken := received()
			if leaked := privateToken != ""; leaked != tt.wantLeak {
				t.Errorf("PRIVATE-TOKEN at the redirect target: got %q, want leak=%v", privateToken, tt.wantLeak)
			}
		})
	}
}

func TestGraphQLClientTokenNotForwardedOnRedirect(t *testing.T) {
	originURL, received := redirectPair(t)
	u, err := url.Parse(originURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("GL_HOST", u.Host)
	t.Setenv("GITLAB_AUTH_TOKEN", "secret")

	var handler graphqlHandler
	handler.init(t.Context(), &Repo{host: u.Host})
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, originURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := handler.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if auth, _ := received(); auth != "" {
		t.Errorf("Authorization at the redirect target: got %q, want none", auth)
	}
}
