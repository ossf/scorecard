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

// Package redirectguard keeps forge credentials on the origin they were
// issued for when an HTTP client follows a redirect.
package redirectguard

import (
	"net/http"
	"net/url"
	"strings"
)

// credentialHeaders are the request headers forge clients carry credentials
// in: bearer tokens (GitHub, GitLab GraphQL) and GitLab personal and job
// tokens.
var credentialHeaders = []string{"Authorization", "Private-Token", "Job-Token"}

// New returns a RoundTripper that sends each request through rt, except a
// redirect that has left the origin (scheme and host) of the request that
// started the chain. That request goes through http.DefaultTransport with the
// credential headers removed.
//
// http.Client drops only a fixed set of headers on a cross-host redirect, and
// it calls the transport again for every hop. A transport that adds a token on
// each call, or a library that sets a custom token header on the request,
// would otherwise send the credential to whatever host the redirect names.
func New(rt http.RoundTripper) http.RoundTripper {
	if rt == nil {
		rt = http.DefaultTransport
	}
	return &transport{credentialed: rt, plain: http.DefaultTransport}
}

type transport struct {
	credentialed http.RoundTripper
	plain        http.RoundTripper
}

func (t *transport) RoundTrip(r *http.Request) (*http.Response, error) {
	if sameOrigin(r.URL, first(r).URL) {
		return t.credentialed.RoundTrip(r) //nolint:wrapcheck // pass through unchanged
	}
	stripped := r.Clone(r.Context())
	for _, h := range credentialHeaders {
		stripped.Header.Del(h)
	}
	return t.plain.RoundTrip(stripped) //nolint:wrapcheck // pass through unchanged
}

// first returns the request that started r's redirect chain. http.Client sets
// Response on each request it sends to follow a redirect.
func first(r *http.Request) *http.Request {
	for r.Response != nil && r.Response.Request != nil {
		r = r.Response.Request
	}
	return r
}

func sameOrigin(a, b *url.URL) bool {
	return strings.EqualFold(a.Scheme, b.Scheme) && strings.EqualFold(a.Host, b.Host)
}
