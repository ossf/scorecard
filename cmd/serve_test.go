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

package cmd

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ossf/scorecard/v5/log"
	"github.com/ossf/scorecard/v5/options"
)

func TestListenAddress(t *testing.T) {
	t.Parallel()
	tests := []struct {
		listen, port, wantHost, wantAddr string
	}{
		{listen: "", port: "", wantHost: "127.0.0.1", wantAddr: "127.0.0.1:8080"},
		{listen: "", port: "9000", wantHost: "127.0.0.1", wantAddr: "127.0.0.1:9000"},
		{listen: "0.0.0.0", port: "9000", wantHost: "0.0.0.0", wantAddr: "0.0.0.0:9000"},
		{listen: "0.0.0.0:9001", port: "9000", wantHost: "0.0.0.0", wantAddr: "0.0.0.0:9001"},
		{listen: "0.0.0.0:", port: "9000", wantHost: "0.0.0.0", wantAddr: "0.0.0.0:9000"},
		{listen: ":9001", port: "", wantHost: "", wantAddr: ":9001"},
		{listen: "localhost", port: "", wantHost: "localhost", wantAddr: "localhost:8080"},
		{listen: "::1", port: "", wantHost: "::1", wantAddr: "[::1]:8080"},
		{listen: "[::1]", port: "", wantHost: "::1", wantAddr: "[::1]:8080"},
		{listen: "[::1]:9001", port: "", wantHost: "::1", wantAddr: "[::1]:9001"},
	}
	for _, tt := range tests {
		host, addr := listenAddress(tt.listen, tt.port)
		if host != tt.wantHost || addr != tt.wantAddr {
			t.Errorf("listenAddress(%q, %q) = %q, %q; want %q, %q",
				tt.listen, tt.port, host, addr, tt.wantHost, tt.wantAddr)
		}
	}
}

func TestIsLoopback(t *testing.T) {
	t.Parallel()
	tests := map[string]bool{
		"localhost":     true,
		"LocalHost":     true,
		"127.0.0.1":     true,
		"127.0.0.2":     true,
		"::1":           true,
		"":              false,
		"0.0.0.0":       false,
		"::":            false,
		"10.0.0.1":      false,
		"other.example": false,
		"localhost.":    false,
	}
	for host, want := range tests {
		if got := isLoopback(host); got != want {
			t.Errorf("isLoopback(%q) = %v, want %v", host, got, want)
		}
	}
}

func TestIsCrossOrigin(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		secFetchSite string
		origin       string
		want         bool
	}{
		{name: "non-browser client", want: false},
		{name: "same origin", secFetchSite: "same-origin", want: false},
		{name: "typed address", secFetchSite: "none", want: false},
		{name: "cross site", secFetchSite: "cross-site", want: true},
		{name: "same site", secFetchSite: "same-site", want: true},
		{name: "cross site with matching origin", secFetchSite: "cross-site", origin: "http://127.0.0.1:8080", want: true},
		{name: "matching origin only", origin: "http://127.0.0.1:8080", want: false},
		{name: "mismatched origin only", origin: "http://example.com", want: true},
		{name: "different port", origin: "http://127.0.0.1:9000", want: true},
		{name: "null origin", origin: "null", want: true},
		{name: "unparsable origin", origin: "http://[::1", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
			r.Host = "127.0.0.1:8080"
			if tt.secFetchSite != "" {
				r.Header.Set("Sec-Fetch-Site", tt.secFetchSite)
			}
			if tt.origin != "" {
				r.Header.Set("Origin", tt.origin)
			}
			if got := isCrossOrigin(r); got != tt.want {
				t.Errorf("isCrossOrigin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServeHandler(t *testing.T) {
	t.Parallel()
	tests := []struct {
		headers  map[string]string
		name     string
		method   string
		target   string
		host     string
		body     string
		wantBody string
		wantCode int
		loopback bool
	}{
		{
			name:     "OPTIONS gets 404",
			method:   http.MethodOptions,
			target:   "/",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "cross-site GET",
			method:   http.MethodGet,
			target:   "/?repo=owner/repo",
			headers:  map[string]string{"Sec-Fetch-Site": "cross-site"},
			wantCode: http.StatusForbidden,
			wantBody: "cross-origin request rejected",
		},
		{
			name:     "cross-site preflight",
			method:   http.MethodOptions,
			target:   "/",
			headers:  map[string]string{"Sec-Fetch-Site": "cross-site", "Origin": "http://example.com"},
			wantCode: http.StatusForbidden,
		},
		{
			name:     "mismatched Origin on POST",
			method:   http.MethodPost,
			target:   "/",
			headers:  map[string]string{"Origin": "http://example.com"},
			body:     `{"repo": "owner/repo"}`,
			wantCode: http.StatusForbidden,
			wantBody: "cross-origin request rejected",
		},
		{
			name:     "typed address",
			method:   http.MethodGet,
			target:   "/health",
			headers:  map[string]string{"Sec-Fetch-Site": "none"},
			wantCode: http.StatusOK,
		},
		{
			name:     "non-browser client",
			method:   http.MethodGet,
			target:   "/health",
			wantCode: http.StatusOK,
		},
		{
			name:     "other host on loopback",
			method:   http.MethodGet,
			target:   "/?repo=owner/repo",
			host:     "other.example:8080",
			loopback: true,
			wantCode: http.StatusForbidden,
			wantBody: "host not allowed",
		},
		{
			name:     "localhost on loopback",
			method:   http.MethodGet,
			target:   "/health",
			host:     "LOCALHOST:8080",
			loopback: true,
			wantCode: http.StatusOK,
		},
		{
			name:     "IPv4 loopback on loopback",
			method:   http.MethodGet,
			target:   "/health",
			host:     "127.0.0.1:8080",
			loopback: true,
			wantCode: http.StatusOK,
		},
		{
			name:     "IPv6 loopback on loopback",
			method:   http.MethodGet,
			target:   "/health",
			host:     "[::1]:8080",
			loopback: true,
			wantCode: http.StatusOK,
		},
		{
			name:     "IPv6 loopback without port",
			method:   http.MethodGet,
			target:   "/health",
			host:     "[::1]",
			loopback: true,
			wantCode: http.StatusOK,
		},
		{
			name:     "other host on non-loopback",
			method:   http.MethodGet,
			target:   "/health",
			host:     "scorecard.internal",
			wantCode: http.StatusOK,
		},
		{
			name:     "oversized body",
			method:   http.MethodPost,
			target:   "/",
			body:     `{"repo": "` + strings.Repeat("a", 2<<20) + `"}`,
			wantCode: http.StatusRequestEntityTooLarge,
			wantBody: "request body too large",
		},
		{
			name:     "valid JSON followed by padding",
			method:   http.MethodPost,
			target:   "/",
			body:     `{"repo": "owner/repo"}` + strings.Repeat(" ", 2<<20),
			wantCode: http.StatusRequestEntityTooLarge,
			wantBody: "request body too large",
		},
		{
			name:     "malformed body",
			method:   http.MethodPost,
			target:   "/",
			body:     `{not json`,
			wantCode: http.StatusBadRequest,
			wantBody: "invalid request body",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			handler := newHandler(newServer(log.NewLogger(log.InfoLevel)), tt.loopback)
			var body io.Reader
			if tt.body != "" {
				body = strings.NewReader(tt.body)
			}
			r := httptest.NewRequestWithContext(t.Context(), tt.method, tt.target, body)
			r.Host = "127.0.0.1:8080"
			if tt.host != "" {
				r.Host = tt.host
			}
			for k, v := range tt.headers {
				r.Header.Set(k, v)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)

			if w.Code != tt.wantCode {
				t.Errorf("status = %d, want %d (body %q)", w.Code, tt.wantCode, w.Body.String())
			}
			if tt.wantBody != "" && strings.TrimSpace(w.Body.String()) != tt.wantBody {
				t.Errorf("body = %q, want %q", w.Body.String(), tt.wantBody)
			}
			if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
				t.Errorf("Access-Control-Allow-Origin = %q, want none", got)
			}
		})
	}
}

func TestServeListenFailure(t *testing.T) {
	var lc net.ListenConfig
	ln, err := lc.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	t.Cleanup(func() {
		if err := ln.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	_, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatalf("SplitHostPort() error = %v", err)
	}
	t.Setenv("PORT", port)

	cmd := serveCmd(options.New())
	cmd.SetArgs([]string{"--listen", "127.0.0.1"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	err = cmd.ExecuteContext(t.Context())
	if want := "listening on 127.0.0.1:" + port; err == nil || !strings.Contains(err.Error(), want) {
		t.Errorf("Execute() error = %v, want one containing %q", err, want)
	}
}
