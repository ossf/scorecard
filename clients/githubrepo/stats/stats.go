// Copyright 2021 OpenSSF Scorecard Authors
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

const (
	meterName = "github.com/ossf/scorecard/v5/clients/githubrepo"

	// TokenIndex is the attribute key for specifying a unique token.
	TokenIndex = attribute.Key("tokenIndex")
	// ResourceType specifies the type of GitHub resource.
	ResourceType = attribute.Key("resourceType")
)

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

var (
	meter = otel.GetMeterProvider().Meter(meterName)

	// RemainingTokens measures the remaining number of API tokens.
	RemainingTokens = must(meter.Int64Gauge(
		"scorecard.github.tokens.remaining",
		metric.WithDescription("Measures the remaining count of API tokens"),
		metric.WithUnit("{token}"),
	))
	// RetryAfter measures the retry delay when dealing with secondary rate limits.
	RetryAfter = must(meter.Int64Histogram(
		"scorecard.github.rate_limit.retry_after",
		metric.WithDescription("Measures the retry delay when dealing with secondary rate limits"),
		metric.WithUnit("s"),
	))
)

// SetMeterProvider updates the meter provider and re-initializes instruments.
func SetMeterProvider(mp metric.MeterProvider) {
	meter = mp.Meter(meterName)
	RemainingTokens = must(meter.Int64Gauge(
		"scorecard.github.tokens.remaining",
		metric.WithDescription("Measures the remaining count of API tokens"),
		metric.WithUnit("{token}"),
	))
	RetryAfter = must(meter.Int64Histogram(
		"scorecard.github.rate_limit.retry_after",
		metric.WithDescription("Measures the retry delay when dealing with secondary rate limits"),
		metric.WithUnit("s"),
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

// RecordRemainingTokens records the remaining token count with context attributes.
func RecordRemainingTokens(ctx context.Context, remaining int64, extraAttrs ...attribute.KeyValue) {
	attrs := combineAttrs(ctx, extraAttrs)
	RemainingTokens.Record(ctx, remaining, metric.WithAttributes(attrs...))
}

// RecordRetryAfter records the retry-after duration with context attributes.
func RecordRetryAfter(ctx context.Context, retryAfterSecs int64, extraAttrs ...attribute.KeyValue) {
	attrs := combineAttrs(ctx, extraAttrs)
	RetryAfter.Record(ctx, retryAfterSecs, metric.WithAttributes(attrs...))
}
