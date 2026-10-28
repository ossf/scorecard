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

// Package stats defines monitoring metrics and views exposed by Scorecard.
package stats

import (
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// CheckRuntimeView returns the recommended OpenTelemetry View for check runtime histogram buckets.
func CheckRuntimeView() sdkmetric.View {
	return sdkmetric.NewView(
		sdkmetric.Instrument{Name: "scorecard.check.runtime"},
		sdkmetric.Stream{
			Aggregation: sdkmetric.AggregationExplicitBucketHistogram{
				Boundaries: []float64{
					0,
					1 << 2,
					1 << 3,
					1 << 4,
					1 << 5,
					1 << 6,
					1 << 7,
					1 << 8,
					1 << 9,
					1 << 10,
					1 << 11,
					1 << 12,
					1 << 13,
					1 << 14,
					1 << 15,
				},
			},
		},
	)
}

// RecommendedViews returns all recommended OpenTelemetry SDK Views for Scorecard.
func RecommendedViews() []sdkmetric.View {
	return []sdkmetric.View{
		CheckRuntimeView(),
	}
}
