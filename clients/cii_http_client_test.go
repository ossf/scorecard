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

package clients

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// defaultClientMu serializes access to http.DefaultClient.Transport across unit tests.
var defaultClientMu sync.Mutex

type trackingReadCloser struct {
	io.ReadCloser
	onClose func()
	mu      sync.Mutex
	closed  bool
}

func newTrackingBody(content string) *trackingReadCloser {
	return &trackingReadCloser{
		ReadCloser: io.NopCloser(strings.NewReader(content)),
	}
}

func (t *trackingReadCloser) Close() error {
	t.mu.Lock()
	t.closed = true
	callback := t.onClose
	t.mu.Unlock()

	if callback != nil {
		callback()
	}
	if t.ReadCloser != nil {
		return t.ReadCloser.Close()
	}
	return nil
}

func (t *trackingReadCloser) isClosed() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.closed
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func mustNewRequest(t *testing.T, method, url string, body io.Reader) *http.Request {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), method, url, body)
	if err != nil {
		t.Fatalf("failed to create http request: %v", err)
	}
	return req
}

//nolint:paralleltest // touches process-wide http.DefaultClient.Transport
func TestExpBackoffTransport_Cancellation(t *testing.T) {
	defaultClientMu.Lock()
	defer defaultClientMu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	origTransport := http.DefaultClient.Transport
	http.DefaultClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		body := newTrackingBody("rate limited")
		body.onClose = func() {
			// As soon as RoundTrip closes the 429 response body and is about
			// to enter backoff, cancel the request context.
			cancel()
		}
		return &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Body:       body,
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})
	defer func() {
		http.DefaultClient.Transport = origTransport
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost/cancellation", nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext() unexpected error: %v", err)
	}

	transport := &expBackoffTransport{numRetries: 3}
	start := time.Now()
	resp, err := transport.RoundTrip(req)
	elapsed := time.Since(start)

	if resp != nil {
		_ = resp.Body.Close()
		t.Fatalf("expected nil response on cancellation, got: %v", resp)
	}
	if err == nil {
		t.Fatal("expected error on cancellation, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected error to wrap context.Canceled, got: %v", err)
	}
	if !strings.Contains(err.Error(), "waiting for CII Best Practices retry") {
		t.Errorf("expected error to contain 'waiting for CII Best Practices retry', got: %v", err)
	}
	if elapsed >= 800*time.Millisecond {
		t.Errorf("expected immediate cancellation abort (< 800ms), took: %v", elapsed)
	}
}

//nolint:paralleltest // touches process-wide http.DefaultClient.Transport
func TestExpBackoffTransport_DeadlineExceeded(t *testing.T) {
	defaultClientMu.Lock()
	defer defaultClientMu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	origTransport := http.DefaultClient.Transport
	http.DefaultClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Body:       newTrackingBody("rate limited"),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})
	defer func() {
		http.DefaultClient.Transport = origTransport
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost/deadline", nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext() unexpected error: %v", err)
	}

	transport := &expBackoffTransport{numRetries: 3}
	start := time.Now()
	resp, err := transport.RoundTrip(req)
	elapsed := time.Since(start)

	if resp != nil {
		_ = resp.Body.Close()
		t.Fatalf("expected nil response on deadline exceeded, got: %v", resp)
	}
	if err == nil {
		t.Fatal("expected error on deadline exceeded, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected error to wrap context.DeadlineExceeded, got: %v", err)
	}
	if !strings.Contains(err.Error(), "waiting for CII Best Practices retry") {
		t.Errorf("expected error to contain 'waiting for CII Best Practices retry', got: %v", err)
	}
	if elapsed >= 800*time.Millisecond {
		t.Errorf("expected immediate timeout abort (< 800ms), took: %v", elapsed)
	}
}

//nolint:paralleltest // touches process-wide http.DefaultClient.Transport
func TestExpBackoffTransport_ClosesDiscarded429Bodies(t *testing.T) {
	defaultClientMu.Lock()
	defer defaultClientMu.Unlock()

	var mu sync.Mutex
	var bodies []*trackingReadCloser
	attempts := 0

	origTransport := http.DefaultClient.Transport
	http.DefaultClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		mu.Lock()
		defer mu.Unlock()

		// Verify every previously created body was already closed before this attempt started.
		for idx, prevBody := range bodies {
			if !prevBody.isClosed() {
				t.Errorf("attempt %d started, but previous body %d was not closed", attempts, idx)
			}
		}

		body := newTrackingBody(fmt.Sprintf("rate limited attempt %d", attempts))
		bodies = append(bodies, body)
		attempts++

		return &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Body:       body,
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})
	defer func() {
		http.DefaultClient.Transport = origTransport
	}()

	transport := &expBackoffTransport{numRetries: 2}
	resp, err := transport.RoundTrip(mustNewRequest(t, http.MethodGet, "http://localhost/retry", nil))

	if resp != nil {
		_ = resp.Body.Close()
		t.Fatalf("expected nil response on exhaustion, got: %v", resp)
	}
	if !errors.Is(err, errTooManyRequests) {
		t.Fatalf("expected errTooManyRequests, got: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
	if len(bodies) != 2 {
		t.Fatalf("expected 2 tracking bodies, got %d", len(bodies))
	}

	for idx, b := range bodies {
		if !b.isClosed() {
			t.Errorf("expected body %d to be closed, but it was not", idx)
		}
	}
}

//nolint:paralleltest // touches process-wide http.DefaultClient.Transport
func TestExpBackoffTransport_RetrySuccess(t *testing.T) {
	defaultClientMu.Lock()
	defer defaultClientMu.Unlock()

	var mu sync.Mutex
	var bodies []*trackingReadCloser
	attempts := 0

	origTransport := http.DefaultClient.Transport
	http.DefaultClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		mu.Lock()
		defer mu.Unlock()

		attempts++
		if attempts == 1 {
			body := newTrackingBody("rate limited")
			bodies = append(bodies, body)
			return &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Body:       body,
				Header:     make(http.Header),
				Request:    req,
			}, nil
		}

		// Ensure previous 429 response body was closed before attempt 2
		if len(bodies) > 0 && !bodies[0].isClosed() {
			t.Errorf("expected discarded 429 body to be closed before attempt 2")
		}

		body := newTrackingBody(`{"status":"success"}`)
		bodies = append(bodies, body)
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       body,
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})
	defer func() {
		http.DefaultClient.Transport = origTransport
	}()

	transport := &expBackoffTransport{numRetries: 2}
	req := mustNewRequest(t, http.MethodGet, "http://localhost/retry-success", nil)

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("expected nil error on retry success, got: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	mu.Lock()
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
	if len(bodies) != 2 {
		t.Fatalf("expected 2 tracked bodies, got %d", len(bodies))
	}
	if !bodies[0].isClosed() {
		t.Errorf("expected discarded 429 body to be closed")
	}
	if bodies[1].isClosed() {
		t.Errorf("expected 200 response body to remain open for caller")
	}
	mu.Unlock()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	if string(data) != `{"status":"success"}` {
		t.Errorf("unexpected body content: got %q, want %q", string(data), `{"status":"success"}`)
	}

	if err := resp.Body.Close(); err != nil {
		t.Fatalf("failed to close response body: %v", err)
	}
	if !bodies[1].isClosed() {
		t.Errorf("expected 200 body to be closed after caller Close()")
	}
}

//nolint:paralleltest // touches process-wide http.DefaultClient.Transport
func TestExpBackoffTransport_Exhaustion(t *testing.T) {
	defaultClientMu.Lock()
	defer defaultClientMu.Unlock()

	var mu sync.Mutex
	attempts := 0

	origTransport := http.DefaultClient.Transport
	http.DefaultClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		mu.Lock()
		attempts++
		mu.Unlock()
		return &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Body:       newTrackingBody("rate limited"),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})
	defer func() {
		http.DefaultClient.Transport = origTransport
	}()

	configuredRetries := uint8(1)
	transport := &expBackoffTransport{numRetries: configuredRetries}
	req := mustNewRequest(t, http.MethodGet, "http://localhost/exhaustion", nil)

	resp, err := transport.RoundTrip(req)
	if resp != nil {
		_ = resp.Body.Close()
		t.Fatalf("expected nil response on exhaustion, got: %v", resp)
	}
	if err == nil {
		t.Fatal("expected error on exhaustion, got nil")
	}
	if !errors.Is(err, errTooManyRequests) {
		t.Errorf("expected error to match errTooManyRequests, got: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if attempts != int(configuredRetries) {
		t.Errorf("expected %d attempts, got %d", configuredRetries, attempts)
	}

	// Also verify that 0 retries immediately returns errTooManyRequests without requests
	attempts = 0
	transport0 := &expBackoffTransport{numRetries: 0}
	resp0, err0 := transport0.RoundTrip(req)
	if resp0 != nil {
		_ = resp0.Body.Close()
		t.Fatalf("expected nil response on 0 retries, got: %v", resp0)
	}
	if !errors.Is(err0, errTooManyRequests) {
		t.Errorf("expected errTooManyRequests on 0 retries, got: %v", err0)
	}
	if attempts != 0 {
		t.Errorf("expected 0 attempts on 0 retries, got %d", attempts)
	}
}

//nolint:paralleltest // touches process-wide http.DefaultClient.Transport
func TestExpBackoffTransport_Non429PassThrough(t *testing.T) {
	defaultClientMu.Lock()
	defer defaultClientMu.Unlock()

	testCases := []struct {
		name       string
		body       string
		statusCode int
	}{
		{
			name:       "HTTP 200 OK",
			body:       "ok response",
			statusCode: http.StatusOK,
		},
		{
			name:       "HTTP 500 Internal Server Error",
			body:       "server error",
			statusCode: http.StatusInternalServerError,
		},
		{
			name:       "HTTP 404 Not Found",
			body:       "not found",
			statusCode: http.StatusNotFound,
		},
	}

	//nolint:paralleltest // touches process-wide http.DefaultClient.Transport
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			attempts := 0
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				attempts++
				w.WriteHeader(tc.statusCode)
				if _, err := w.Write([]byte(tc.body)); err != nil {
					t.Errorf("failed to write test response: %v", err)
				}
			}))
			defer ts.Close()

			req := mustNewRequest(t, http.MethodGet, ts.URL, nil)
			transport := &expBackoffTransport{numRetries: 3}

			start := time.Now()
			resp, err := transport.RoundTrip(req)
			elapsed := time.Since(start)

			if err != nil {
				t.Fatalf("expected nil error, got: %v", err)
			}
			if resp == nil {
				t.Fatal("expected non-nil response")
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.statusCode {
				t.Errorf("expected status %d, got %d", tc.statusCode, resp.StatusCode)
			}

			data, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("failed to read response body: %v", err)
			}
			if string(data) != tc.body {
				t.Errorf("expected body %q, got %q", tc.body, string(data))
			}

			if attempts != 1 {
				t.Errorf("expected exactly 1 attempt without retrying, got %d", attempts)
			}
			if elapsed >= 500*time.Millisecond {
				t.Errorf("expected immediate return without retrying (< 500ms), took: %v", elapsed)
			}
		})
	}
}

//nolint:paralleltest // touches process-wide http.DefaultClient.Transport
func TestHttpClientCIIBestPractices_GetBadgeLevel(t *testing.T) {
	defaultClientMu.Lock()
	defer defaultClientMu.Unlock()

	//nolint:paralleltest // touches process-wide http.DefaultClient.Transport
	t.Run("SuccessPassing", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`[{"badge_level":"passing"}]`)); err != nil {
				t.Errorf("failed to write test response: %v", err)
			}
		}))
		defer ts.Close()

		client := &httpClientCIIBestPractices{baseURL: ts.URL}
		level, err := client.GetBadgeLevel(context.Background(), "github.com/ossf/scorecard")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if level != Passing {
			t.Errorf("expected Passing, got %v", level)
		}
	})

	//nolint:paralleltest // touches process-wide http.DefaultClient.Transport
	t.Run("NotFound", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`[]`)); err != nil {
				t.Errorf("failed to write test response: %v", err)
			}
		}))
		defer ts.Close()

		client := &httpClientCIIBestPractices{baseURL: ts.URL}
		level, err := client.GetBadgeLevel(context.Background(), "github.com/ossf/scorecard")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if level != NotFound {
			t.Errorf("expected NotFound, got %v", level)
		}
	})

	//nolint:paralleltest // touches process-wide http.DefaultClient.Transport
	t.Run("ContextCancelledDuringRetry", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var once sync.Once
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
			once.Do(func() {
				go func() {
					time.Sleep(10 * time.Millisecond)
					cancel()
				}()
			})
		}))
		defer ts.Close()

		client := &httpClientCIIBestPractices{baseURL: ts.URL}
		start := time.Now()
		level, err := client.GetBadgeLevel(ctx, "github.com/ossf/scorecard")
		elapsed := time.Since(start)

		if level != Unknown {
			t.Errorf("expected Unknown on error, got %v", level)
		}
		if err == nil {
			t.Fatal("expected error on cancelled context, got nil")
		}
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected error to wrap context.Canceled, got: %v", err)
		}
		if elapsed >= 800*time.Millisecond {
			t.Errorf("expected immediate abort on cancellation (< 800ms), took: %v", elapsed)
		}
	})
}
