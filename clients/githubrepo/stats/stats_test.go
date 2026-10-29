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

	ctx = NewContext(ctx, TokenIndex.String("0"), ResourceType.String("core"))
	attrs := FromContext(ctx)
	if len(attrs) != 2 {
		t.Fatalf("expected 2 attrs, got %d", len(attrs))
	}
	if attrs[0].Key != TokenIndex || attrs[0].Value.AsString() != "0" {
		t.Errorf("unexpected attribute: %v", attrs[0])
	}
	if attrs[1].Key != ResourceType || attrs[1].Value.AsString() != "core" {
		t.Errorf("unexpected attribute: %v", attrs[1])
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
	ctx = NewContext(ctx, TokenIndex.String("1"))

	RecordRemainingTokens(ctx, 4999, ResourceType.String("search"))
	RecordRetryAfter(ctx, 60, ResourceType.String("graphql"))

	var rm metricdata.ResourceMetrics
	if err := reader.Collect(ctx, &rm); err != nil {
		t.Fatalf("failed to collect metrics: %v", err)
	}

	remainingMetric, ok := findMetric(&rm, "scorecard.github.tokens.remaining")
	if !ok {
		t.Error("scorecard.github.tokens.remaining metric not found in collection")
	} else if gauge, ok := remainingMetric.Data.(metricdata.Gauge[int64]); ok {
		for _, dp := range gauge.DataPoints {
			if dp.Value != 4999 {
				t.Errorf("expected value 4999, got %d", dp.Value)
			}
			if val, exists := dp.Attributes.Value(TokenIndex); !exists || val.AsString() != "1" {
				t.Errorf("expected tokenIndex 1, got %v", val)
			}
			if val, exists := dp.Attributes.Value(ResourceType); !exists || val.AsString() != "search" {
				t.Errorf("expected resourceType search, got %v", val)
			}
		}
	}

	retryMetric, ok := findMetric(&rm, "scorecard.github.rate_limit.retry_after")
	if !ok {
		t.Error("scorecard.github.rate_limit.retry_after metric not found in collection")
	} else if hist, ok := retryMetric.Data.(metricdata.Histogram[int64]); ok {
		for _, dp := range hist.DataPoints {
			if val, exists := dp.Attributes.Value(TokenIndex); !exists || val.AsString() != "1" {
				t.Errorf("expected tokenIndex 1, got %v", val)
			}
			if val, exists := dp.Attributes.Value(ResourceType); !exists || val.AsString() != "graphql" {
				t.Errorf("expected resourceType graphql, got %v", val)
			}
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
