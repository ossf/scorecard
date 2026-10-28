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

package stats

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestContextHelpers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	if attrs := FromContext(ctx); len(attrs) != 0 {
		t.Fatalf("expected 0 attrs, got %d", len(attrs))
	}

	ctx = NewContext(ctx, CheckName.String("Binary-Artifacts"), RepoHost.String("github.com"))
	attrs := FromContext(ctx)
	if len(attrs) != 2 {
		t.Fatalf("expected 2 attrs, got %d", len(attrs))
	}
	if attrs[0].Key != CheckName || attrs[0].Value.AsString() != "Binary-Artifacts" {
		t.Errorf("unexpected attribute: %v", attrs[0])
	}
	if attrs[1].Key != RepoHost || attrs[1].Value.AsString() != "github.com" {
		t.Errorf("unexpected attribute: %v", attrs[1])
	}

	// Appending more attributes
	ctx = NewContext(ctx, RequestTag.String("from-cache"))
	attrs = FromContext(ctx)
	if len(attrs) != 3 {
		t.Fatalf("expected 3 attrs, got %d", len(attrs))
	}
}

//nolint:paralleltest // modifies package-level meter provider
func TestMetricsRecording(t *testing.T) {
	origMP := otel.GetMeterProvider()
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	SetMeterProvider(mp)
	t.Cleanup(func() {
		SetMeterProvider(origMP)
	})

	ctx := context.Background()
	ctx = NewContext(ctx, CheckName.String("Packaging"), RepoHost.String("github.com"))

	RecordCheckRuntime(ctx, 42)
	RecordCheckError(ctx, ErrorName.String("internal-error"))
	RecordHTTPRequests(ctx, 5, RequestTag.String("requested"))

	var rm metricdata.ResourceMetrics
	if err := reader.Collect(ctx, &rm); err != nil {
		t.Fatalf("failed to collect metrics: %v", err)
	}

	verifyRuntimeMetric(t, &rm)
	verifyErrorMetric(t, &rm)
	verifyHTTPMetric(t, &rm)
}

func verifyRuntimeMetric(t *testing.T, rm *metricdata.ResourceMetrics) {
	t.Helper()
	runtimeMetric, ok := findMetric(rm, "scorecard.check.runtime")
	if !ok {
		t.Error("scorecard.check.runtime metric not found in collection")
		return
	}
	hist, ok := runtimeMetric.Data.(metricdata.Histogram[int64])
	if !ok {
		return
	}
	for _, dp := range hist.DataPoints {
		if val, exists := dp.Attributes.Value(CheckName); !exists || val.AsString() != "Packaging" {
			t.Errorf("expected checkName Packaging, got %v", val)
		}
		if val, exists := dp.Attributes.Value(RepoHost); !exists || val.AsString() != "github.com" {
			t.Errorf("expected repoHost github.com, got %v", val)
		}
	}
}

func verifyErrorMetric(t *testing.T, rm *metricdata.ResourceMetrics) {
	t.Helper()
	errorMetric, ok := findMetric(rm, "scorecard.check.errors")
	if !ok {
		t.Error("scorecard.check.errors metric not found in collection")
		return
	}
	sum, ok := errorMetric.Data.(metricdata.Sum[int64])
	if !ok {
		return
	}
	for _, dp := range sum.DataPoints {
		if dp.Value != 1 {
			t.Errorf("expected error count 1, got %d", dp.Value)
		}
		if val, exists := dp.Attributes.Value(CheckName); !exists || val.AsString() != "Packaging" {
			t.Errorf("expected checkName Packaging, got %v", val)
		}
		if val, exists := dp.Attributes.Value(ErrorName); !exists || val.AsString() != "internal-error" {
			t.Errorf("expected errorName internal-error, got %v", val)
		}
	}
}

func verifyHTTPMetric(t *testing.T, rm *metricdata.ResourceMetrics) {
	t.Helper()
	httpMetric, ok := findMetric(rm, "scorecard.http.requests")
	if !ok {
		t.Error("scorecard.http.requests metric not found in collection")
		return
	}
	sum, ok := httpMetric.Data.(metricdata.Sum[int64])
	if !ok {
		return
	}
	for _, dp := range sum.DataPoints {
		if dp.Value != 5 {
			t.Errorf("expected http request count 5, got %d", dp.Value)
		}
		if val, exists := dp.Attributes.Value(RequestTag); !exists || val.AsString() != "requested" {
			t.Errorf("expected requestTag requested, got %v", val)
		}
	}
}

func findMetric(rm *metricdata.ResourceMetrics, name string) (metricdata.Metrics, bool) {
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == name {
				return m, true
			}
		}
	}
	return metricdata.Metrics{}, false
}

func TestViews(t *testing.T) {
	t.Parallel()

	views := RecommendedViews()
	if len(views) == 0 {
		t.Fatal("expected at least one recommended view")
	}

	reader := sdkmetric.NewManualReader()
	opts := make([]sdkmetric.Option, 0, len(views)+1)
	opts = append(opts, sdkmetric.WithReader(reader))
	for _, v := range views {
		opts = append(opts, sdkmetric.WithView(v))
	}
	mp := sdkmetric.NewMeterProvider(opts...)
	if mp == nil {
		t.Fatal("expected non-nil MeterProvider with recommended views")
	}
}
