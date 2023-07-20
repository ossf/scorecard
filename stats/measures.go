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
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

var Meter = otel.Meter("ossf-scorecard")

var Metrics struct {
	CheckRuntimeInSec metric.Int64Histogram
	CheckErrors       metric.Int64Histogram
	HTTPRequests      metric.Int64Histogram
}
var once sync.Once

func InitMetrics() error {
	var err error
	once.Do(func() {
		c, err := Meter.Int64Histogram("CheckRuntimeInSec",
			metric.WithDescription("Measures the CPU runtime in seconds for a check"),
			metric.WithUnit("s"))
		if err != nil {
			return
		}
		Metrics.CheckRuntimeInSec = c

		c, err = Meter.Int64Histogram("CheckErrors", metric.WithDescription("Measures the count of errors"))
		if err != nil {
			return
		}
		Metrics.CheckErrors = c

		c, err = Meter.Int64Histogram("HTTPRequests", metric.WithDescription("Measures the count of HTTP requests"))
		if err != nil {
			return
		}
		Metrics.HTTPRequests = c
	})

	return err
}
