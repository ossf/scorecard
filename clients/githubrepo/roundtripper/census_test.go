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

package roundtripper

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	githubstats "github.com/ossf/scorecard/v5/clients/githubrepo/stats"
	"github.com/ossf/scorecard/v5/log"
	"github.com/ossf/scorecard/v5/stats"
)

//nolint:paralleltest // modifies package-level meter provider
func TestOTelTransport(t *testing.T) {
	origMP := otel.GetMeterProvider()
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	stats.SetMeterProvider(mp)
	t.Cleanup(func() {
		stats.SetMeterProvider(origMP)
	})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cached" {
			w.Header().Set(fromCacheHeader, "true")
		}
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("ok")); err != nil {
			t.Errorf("w.Write failed: %v", err)
		}
	}))
	t.Cleanup(func() {
		ts.Close()
	})

	transport := MakeOTelTransport(ts.Client().Transport)

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/test", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}
	defer resp.Body.Close()

	cachedReq, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/cached", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	cachedResp, err := transport.RoundTrip(cachedReq)
	if err != nil {
		t.Fatalf("RoundTrip cached failed: %v", err)
	}
	defer cachedResp.Body.Close()

	var rm metricdata.ResourceMetrics
	if err := reader.Collect(ctx, &rm); err != nil {
		t.Fatalf("failed to collect metrics: %v", err)
	}

	foundHTTP := false
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "scorecard.http.requests" {
				foundHTTP = true
			}
		}
	}

	if !foundHTTP {
		t.Error("scorecard.http.requests metric not found after round trips")
	}
}

func TestMakeCensusTransportDeprecated(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(func() {
		ts.Close()
	})

	transport := MakeCensusTransport(ts.Client().Transport)
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, ts.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

type testTokenAccessor struct {
	token string
	id    uint64
}

func (m *testTokenAccessor) Next() (uint64, string) {
	return m.id, m.token
}

func (m *testTokenAccessor) Release(id uint64) {}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestGitHubTransport_AuthAndContext(t *testing.T) {
	t.Parallel()

	var gotAuth string
	var gotTokenIndex string

	mockTransport := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		gotAuth = r.Header.Get("Authorization")
		for _, attr := range githubstats.FromContext(r.Context()) {
			if attr.Key == githubstats.TokenIndex {
				gotTokenIndex = attr.Value.AsString()
			}
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("ok")),
		}, nil
	})

	accessor := &testTokenAccessor{id: 42, token: "test-token"}
	transport := makeGitHubTransport(mockTransport, accessor)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://api.github.com/test", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}
	defer resp.Body.Close()

	if gotAuth != "Bearer test-token" {
		t.Errorf("expected Bearer test-token, got %s", gotAuth)
	}
	if gotTokenIndex != "42" {
		t.Errorf("expected tokenIndex 42, got %s", gotTokenIndex)
	}
}

//nolint:paralleltest // modifies package-level meter provider
func TestRateLimitTransport_Metrics(t *testing.T) {
	origMP := otel.GetMeterProvider()
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	githubstats.SetMeterProvider(mp)
	t.Cleanup(func() {
		githubstats.SetMeterProvider(origMP)
	})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "42")
		w.Header().Set("X-RateLimit-Resource", "core")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("ok")); err != nil {
			t.Errorf("w.Write failed: %v", err)
		}
	}))
	t.Cleanup(func() {
		ts.Close()
	})

	transport := &rateLimitTransport{
		innerTransport: ts.Client().Transport,
		logger:         log.NewLogger(log.DefaultLevel),
	}

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}
	defer resp.Body.Close()

	var rm metricdata.ResourceMetrics
	if err := reader.Collect(ctx, &rm); err != nil {
		t.Fatalf("failed to collect metrics: %v", err)
	}

	foundGauge := false
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "scorecard.github.tokens.remaining" {
				foundGauge = true
			}
		}
	}
	if !foundGauge {
		t.Error("scorecard.github.tokens.remaining gauge metric not recorded")
	}
}
