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
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

var Meter = otel.Meter("ossf-scorecard")

var Metrics struct {
	RemainingTokens metric.Int64Histogram
	RetryAfter      metric.Int64Histogram
}

var once sync.Once

func InitMetrics() error {
	var err error
	once.Do(func() {
		c, err := Meter.Int64Histogram("RetryAfter",
			metric.WithDescription("Measures the retry delay when dealing with secondary rate limits"),
			metric.WithUnit("s"))
		if err != nil {
			return
		}
		Metrics.RetryAfter = c

		c, err = Meter.Int64Histogram("RemainingTokens", metric.WithDescription("Measures the remaining count of API tokens"))
		if err != nil {
			return
		}
		Metrics.RemainingTokens = c
	})

	return err
}

var (
// // GithubTokens tracks the usage/remaining stats per token per resource-type.
//
//	GithubTokens = view.View{
//		Name:        "GithubTokens",
//		Description: "Token usage/remaining stats for Github API tokens",
//		Measure:     RemainingTokens,
//		TagKeys:     []tag.Key{TokenIndex, ResourceType},
//		Aggregation: view.LastValue(),
//	}
)
