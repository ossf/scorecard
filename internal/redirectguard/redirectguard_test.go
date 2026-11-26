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

package redirectguard

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// addToken mimics a transport that sets a credential on every call, as the
// GitHub token transport, ghinstallation, and oauth2.Transport do.
type addToken struct{}

func (addToken) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Header.Set("Authorization", "Bearer secret")
	return http.DefaultTransport.RoundTrip(r)
}

// recorder is a server that records the credential headers it receives.
type recorder struct {
	*httptest.Server
	auth, privateToken string
}

func newRecorder(t *testing.T, next http.Handler) *recorder {
	t.Helper()
	rec := &recorder{}
	rec.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.auth = r.Header.Get("Authorization")
		rec.privateToken = r.Header.Get("PRIVATE-TOKEN")
		if next != nil {
			next.ServeHTTP(w, r)
		}
	}))
	t.Cleanup(rec.Close)
	return rec
}

func redirectTo(url string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, url, http.StatusFound)
	})
}

func get(t *testing.T, client *http.Client, url string, header http.Header) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range header {
		req.Header[k] = v
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
}

func TestCrossOriginRedirect(t *testing.T) {
	t.Parallel()
	tests := []struct {
		transport http.RoundTripper
		header    http.Header
		name      string
		wantLeak  bool
	}{
		{
			name:      "transport token, unguarded",
			transport: addToken{},
			wantLeak:  true,
		},
		{
			name:      "transport token, guarded",
			transport: New(addToken{}),
		},
		{
			name:      "request header token, unguarded",
			transport: http.DefaultTransport,
			header:    http.Header{"Private-Token": {"secret"}},
			wantLeak:  true,
		},
		{
			name:      "request header token, guarded",
			transport: New(nil),
			header:    http.Header{"Private-Token": {"secret"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// httptest servers differ by port, so the redirect leaves the origin.
			other := newRecorder(t, nil)
			origin := newRecorder(t, redirectTo(other.URL))

			get(t, &http.Client{Transport: tt.transport}, origin.URL, tt.header)

			if origin.auth == "" && origin.privateToken == "" {
				t.Fatal("origin got no credential")
			}
			leaked := other.auth != "" || other.privateToken != ""
			if leaked != tt.wantLeak {
				t.Errorf("credential at the redirect target: got %v, want %v (Authorization %q, PRIVATE-TOKEN %q)",
					leaked, tt.wantLeak, other.auth, other.privateToken)
			}
		})
	}
}

func TestSameOriginRedirectKeepsCredential(t *testing.T) {
	t.Parallel()
	origin := newRecorder(t, nil)
	mux := http.NewServeMux()
	mux.Handle("/start", redirectTo("/end"))
	mux.HandleFunc("/end", func(w http.ResponseWriter, r *http.Request) {
		origin.auth = r.Header.Get("Authorization")
	})
	origin.Config.Handler = mux

	get(t, &http.Client{Transport: New(addToken{})}, origin.URL+"/start", nil)

	if origin.auth != "Bearer secret" {
		t.Errorf("Authorization after a same-origin redirect: got %q, want the token", origin.auth)
	}
}
