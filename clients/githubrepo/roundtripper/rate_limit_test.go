// Copyright 2023 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package roundtripper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ossf/scorecard/v4/log"
)

func TestRoundTrip(t *testing.T) {
	t.Parallel()
	var requestCount int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Customize the response headers and body based on the test scenario
		switch r.URL.Path {
		case "/error":
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error")) // nolint: errcheck
		case "/retry":
			requestCount++
			if requestCount == 2 {
				// Second request: Return successful response
				w.Header().Set("X-RateLimit-Remaining", "10")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Success")) // nolint: errcheck
			} else {
				// First request: Return Retry-After header
				w.Header().Set("Retry-After", "1")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte("Rate Limit Exceeded")) // nolint: errcheck
			}
		case "/success":
			w.Header().Set("X-RateLimit-Remaining", "10")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Success")) // nolint: errcheck
		}
	}))
	t.Cleanup(func() {
		defer ts.Close()
	})

	// Create the rateLimitTransport with the test server as the inner transport and a default logger
	transport := &rateLimitTransport{
		innerTransport: ts.Client().Transport,
		logger:         log.NewLogger(log.DefaultLevel),
	}

	t.Run("Successful response", func(t *testing.T) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, ts.URL+"/success", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		resp, err := transport.RoundTrip(req)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("Retry-After header set", func(t *testing.T) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, ts.URL+"/retry", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		resp, err := transport.RoundTrip(req)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}
		if requestCount != 2 {
			t.Errorf("Expected 2 requests, got %d", requestCount)
		}
	})
}
