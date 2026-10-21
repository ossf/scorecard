// Copyright 2026 OpenSSF Scorecard Authors
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

package clients

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/synctest"
	"time"
)

//nolint:paralleltest // These tests replace http.DefaultClient.
func TestExpBackoffTransportContext(t *testing.T) {
	for _, deadline := range []bool{false, true} {
		name := "cancellation"
		wantErr := context.Canceled
		if deadline {
			name = "deadline"
			wantErr = context.DeadlineExceeded
		}
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				tracking := trackCIIResponses(t, &http.Client{Transport: &ciiStubTransport{
					statuses: []int{http.StatusTooManyRequests},
				}})
				wantWait := 100 * time.Millisecond
				ctx, cancel := context.WithCancel(t.Context())
				if deadline {
					cancel()
					wantWait = 250 * time.Millisecond
					ctx, cancel = context.WithTimeout(t.Context(), wantWait)
				} else {
					tracking.onResponse = func() {
						// Cancel during the backoff, after receiving the 429 response.
						timer := time.AfterFunc(wantWait, cancel)
						t.Cleanup(func() { timer.Stop() })
					}
				}
				t.Cleanup(cancel)
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://cii.test", nil)
				if err != nil {
					t.Fatal(err)
				}

				start := time.Now()
				resp, err := (&expBackoffTransport{numRetries: 3}).RoundTrip(req)
				if resp != nil {
					resp.Body.Close()
					t.Error("expected no response after context cancellation")
				}
				if !errors.Is(err, wantErr) {
					t.Errorf("RoundTrip() error = %v, want %v", err, wantErr)
				}
				if elapsed := time.Since(start); elapsed != wantWait {
					t.Errorf("context interrupted backoff after %s, want %s", elapsed, wantWait)
				}
				checkCIIResponses(t, tracking, 1)
			})
		})
	}
}

//nolint:paralleltest // These tests replace http.DefaultClient.
func TestExpBackoffTransportRetries(t *testing.T) {
	tests := []struct {
		wantErr    error
		name       string
		statuses   []int
		wantWait   time.Duration
		numRetries uint8
	}{
		{
			name:       "success after retries",
			statuses:   []int{http.StatusTooManyRequests, http.StatusTooManyRequests, http.StatusOK},
			numRetries: 3,
			wantWait:   3 * time.Second,
		},
		{
			name:       "exhausted retries",
			statuses:   []int{http.StatusTooManyRequests, http.StatusTooManyRequests, http.StatusTooManyRequests},
			numRetries: 3,
			wantErr:    errTooManyRequests,
			wantWait:   7 * time.Second,
		},
		{
			name:       "non-429 response",
			statuses:   []int{http.StatusInternalServerError},
			numRetries: 3,
		},
		{
			name:    "no attempts",
			wantErr: errTooManyRequests,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				tracking := trackCIIResponses(t, &http.Client{Transport: &ciiStubTransport{statuses: tt.statuses}})
				req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://cii.test", nil)
				if err != nil {
					t.Fatal(err)
				}

				start := time.Now()
				resp, err := (&expBackoffTransport{numRetries: tt.numRetries}).RoundTrip(req)
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("RoundTrip() error = %v, want %v", err, tt.wantErr)
				}
				if elapsed := time.Since(start); elapsed != tt.wantWait {
					t.Errorf("backoff took %s, want %s", elapsed, tt.wantWait)
				}
				checkCIIResponses(t, tracking, len(tt.statuses))
				if tt.wantErr != nil {
					if resp != nil {
						resp.Body.Close()
						t.Error("expected no response after retry exhaustion")
					}
					return
				}
				checkCIIReturnedResponse(t, resp, tt.statuses[len(tt.statuses)-1])
			})
		})
	}
}

//nolint:paralleltest // These tests replace http.DefaultClient.
func TestExpBackoffTransportHTTPResponses(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusTooManyRequests, http.StatusInternalServerError} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(status)
				if _, err := io.WriteString(w, "response"); err != nil {
					t.Errorf("writing response: %v", err)
				}
			}))
			t.Cleanup(server.Close)
			tracking := trackCIIResponses(t, server.Client())
			ctx, cancel := context.WithCancel(t.Context())
			t.Cleanup(cancel)
			if status == http.StatusTooManyRequests {
				// Interrupt the wait after an actual HTTP 429 response.
				tracking.onResponse = cancel
			}
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
			if err != nil {
				t.Fatal(err)
			}
			resp, err := (&expBackoffTransport{numRetries: 3}).RoundTrip(req)
			checkCIIResponses(t, tracking, 1)
			if status == http.StatusTooManyRequests {
				if !errors.Is(err, context.Canceled) {
					t.Errorf("RoundTrip() error = %v, want context.Canceled", err)
				}
				if resp != nil {
					resp.Body.Close()
					t.Error("expected no response after context cancellation")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			checkCIIReturnedResponse(t, resp, status)
		})
	}
}

func checkCIIResponses(t *testing.T, tracking *ciiTrackingTransport, wantResponses int) {
	t.Helper()
	if len(tracking.bodies) != wantResponses {
		t.Fatalf("received %d responses, want %d", len(tracking.bodies), wantResponses)
	}
	for i, body := range tracking.bodies {
		wantClosed := body.statusCode == http.StatusTooManyRequests
		if body.closed != wantClosed {
			t.Errorf("response %d: body closed = %t, want %t", i, body.closed, wantClosed)
		}
		if i > 0 {
			wantDelay := time.Second << (i - 1)
			if delay := tracking.attempts[i].Sub(tracking.attempts[i-1]); delay != wantDelay {
				t.Errorf("attempt %d: backoff took %s, want %s", i, delay, wantDelay)
			}
		}
	}
}

func checkCIIReturnedResponse(t *testing.T, resp *http.Response, wantStatus int) {
	t.Helper()
	if resp == nil {
		t.Fatal("expected a response")
	}
	defer resp.Body.Close()
	if resp.StatusCode != wantStatus {
		t.Errorf("status = %d, want %d", resp.StatusCode, wantStatus)
	}
	if body, err := io.ReadAll(resp.Body); err != nil || string(body) != "response" {
		t.Errorf("response body = %q, error = %v", body, err)
	}
}

type ciiTrackingBody struct {
	io.ReadCloser
	statusCode int
	closed     bool
}

func (body *ciiTrackingBody) Close() error {
	body.closed = true
	return body.ReadCloser.Close()
}

type ciiTrackingTransport struct {
	http.RoundTripper
	t          *testing.T
	onResponse func()
	bodies     []*ciiTrackingBody
	attempts   []time.Time
}

func (transport *ciiTrackingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if n := len(transport.bodies); n > 0 && !transport.bodies[n-1].closed {
		transport.t.Error("discarded response body was still open before the next attempt")
	}
	transport.attempts = append(transport.attempts, time.Now())
	resp, err := transport.RoundTripper.RoundTrip(req)
	if err != nil {
		return resp, err
	}
	body := &ciiTrackingBody{ReadCloser: resp.Body, statusCode: resp.StatusCode}
	resp.Body = body
	transport.bodies = append(transport.bodies, body)
	if transport.onResponse != nil {
		transport.onResponse()
	}
	return resp, nil
}

func trackCIIResponses(t *testing.T, client *http.Client) *ciiTrackingTransport {
	t.Helper()
	transport := &ciiTrackingTransport{RoundTripper: client.Transport, t: t}
	client.Transport = transport
	original := http.DefaultClient
	http.DefaultClient = client
	t.Cleanup(func() {
		http.DefaultClient = original
		for _, body := range transport.bodies {
			body.Close()
		}
	})
	return transport
}

type ciiStubTransport struct {
	statuses []int
	requests int
}

func (transport *ciiStubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := req.Context().Err(); err != nil {
		return nil, err
	}
	status := http.StatusTooManyRequests
	if transport.requests < len(transport.statuses) {
		status = transport.statuses[transport.requests]
	}
	transport.requests++
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader("response")),
		Request:    req,
	}, nil
}
