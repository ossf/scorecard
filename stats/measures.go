// Copyright 2020 OpenSSF Scorecard Authors
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

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const meterName = "github.com/ossf/scorecard/v5"

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

var (
	meter = otel.GetMeterProvider().Meter(meterName)

	// CheckRuntimeInSec measures the CPU runtime in seconds per check.
	CheckRuntimeInSec = must(meter.Int64Histogram(
		"scorecard.check.runtime",
		metric.WithDescription("Measures the CPU runtime in seconds for a check"),
		metric.WithUnit("s"),
	))
	// CheckErrors measures the count of errors per check.
	CheckErrors = must(meter.Int64Counter(
		"scorecard.check.errors",
		metric.WithDescription("Measures the count of errors"),
		metric.WithUnit("{error}"),
	))
	// HTTPRequests measures the count of HTTP requests.
	HTTPRequests = must(meter.Int64Counter(
		"scorecard.http.requests",
		metric.WithDescription("Measures the count of HTTP requests"),
		metric.WithUnit("{request}"),
	))
)

// SetMeterProvider updates the meter provider and re-initializes instruments.
func SetMeterProvider(mp metric.MeterProvider) {
	meter = mp.Meter(meterName)
	CheckRuntimeInSec = must(meter.Int64Histogram(
		"scorecard.check.runtime",
		metric.WithDescription("Measures the CPU runtime in seconds for a check"),
		metric.WithUnit("s"),
	))
	CheckErrors = must(meter.Int64Counter(
		"scorecard.check.errors",
		metric.WithDescription("Measures the count of errors"),
		metric.WithUnit("{error}"),
	))
	HTTPRequests = must(meter.Int64Counter(
		"scorecard.http.requests",
		metric.WithDescription("Measures the count of HTTP requests"),
		metric.WithUnit("{request}"),
	))
}

type contextKey struct{}

// NewContext returns a new context with the provided attributes attached.
func NewContext(ctx context.Context, attrs ...attribute.KeyValue) context.Context {
	if len(attrs) == 0 {
		return ctx
	}
	var combined []attribute.KeyValue
	if existing, ok := ctx.Value(contextKey{}).([]attribute.KeyValue); ok {
		combined = make([]attribute.KeyValue, 0, len(existing)+len(attrs))
		combined = append(combined, existing...)
	} else {
		combined = make([]attribute.KeyValue, 0, len(attrs))
	}
	combined = append(combined, attrs...)
	return context.WithValue(ctx, contextKey{}, combined)
}

// FromContext extracts the attributes stored in context.
func FromContext(ctx context.Context) []attribute.KeyValue {
	if attrs, ok := ctx.Value(contextKey{}).([]attribute.KeyValue); ok {
		return attrs
	}
	return nil
}

func combineAttrs(ctx context.Context, extraAttrs []attribute.KeyValue) []attribute.KeyValue {
	ctxAttrs := FromContext(ctx)
	if len(extraAttrs) == 0 {
		return ctxAttrs
	}
	combined := make([]attribute.KeyValue, 0, len(ctxAttrs)+len(extraAttrs))
	combined = append(combined, ctxAttrs...)
	combined = append(combined, extraAttrs...)
	return combined
}

// RecordCheckRuntime records check runtime duration with context attributes.
func RecordCheckRuntime(ctx context.Context, runtimeSecs int64, extraAttrs ...attribute.KeyValue) {
	attrs := combineAttrs(ctx, extraAttrs)
	CheckRuntimeInSec.Record(ctx, runtimeSecs, metric.WithAttributes(attrs...))
}

// RecordCheckError records a check error with context attributes.
func RecordCheckError(ctx context.Context, extraAttrs ...attribute.KeyValue) {
	attrs := combineAttrs(ctx, extraAttrs)
	CheckErrors.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// RecordHTTPRequests records an HTTP request count with context attributes.
func RecordHTTPRequests(ctx context.Context, count int64, extraAttrs ...attribute.KeyValue) {
	attrs := combineAttrs(ctx, extraAttrs)
	HTTPRequests.Add(ctx, count, metric.WithAttributes(attrs...))
}
