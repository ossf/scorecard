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

package githubrepo

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// tokenTransport sets a token on every call, as the PAT transport and
// ghinstallation do.
type tokenTransport struct{}

func (tokenTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Header.Set("Authorization", "Bearer secret")
	return http.DefaultTransport.RoundTrip(r)
}

// The tarball download follows the archive redirect with the client's
// configured transport.
func TestTokenNotForwardedOnRedirect(t *testing.T) {
	t.Parallel()
	var originAuth, otherAuth string
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		otherAuth = r.Header.Get("Authorization")
	}))
	t.Cleanup(other.Close)
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		originAuth = r.Header.Get("Authorization")
		http.Redirect(w, r, other.URL, http.StatusFound)
	}))
	t.Cleanup(origin.Close)

	rc, err := NewRepoClient(t.Context(), WithRoundTripper(tokenTransport{}))
	if err != nil {
		t.Fatal(err)
	}
	client, ok := rc.(*Client)
	if !ok {
		t.Fatalf("NewRepoClient returned %T", rc)
	}
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, origin.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.tarball.httpClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if originAuth != "Bearer secret" {
		t.Fatalf("Authorization at the origin: got %q, want the token", originAuth)
	}
	if otherAuth != "" {
		t.Errorf("Authorization at the redirect target: got %q, want none", otherAuth)
	}
}
