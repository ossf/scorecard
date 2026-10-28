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

package checker

import (
	"context"
	"errors"
	"testing"

	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/clients/localdir"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/stats"
)

//nolint:paralleltest // modifies package-level meter provider
func TestRunner_Run_Metrics(t *testing.T) {
	origMP := otel.GetMeterProvider()
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	stats.SetMeterProvider(mp)
	t.Cleanup(func() {
		stats.SetMeterProvider(origMP)
	})

	repo, err := localdir.MakeLocalDirRepo(".")
	if err != nil {
		t.Fatalf("MakeLocalDirRepo failed: %v", err)
	}

	runner := &Runner{
		CheckName: "Test-Check",
		Repo:      "local",
		CheckRequest: CheckRequest{
			Repo: repo,
			RequiredTypes: []RequestType{
				FileBased,
			},
		},
	}

	checkSuccess := Check{
		SupportedRequestTypes: []RequestType{FileBased},
		Fn: func(c *CheckRequest) CheckResult {
			return CreateMaxScoreResult(runner.CheckName, "all good")
		},
	}

	ctx := context.Background()
	res := runner.Run(ctx, checkSuccess)
	if res.Score != MaxResultScore {
		t.Errorf("expected score %d, got %d", MaxResultScore, res.Score)
	}

	checkError := Check{
		SupportedRequestTypes: []RequestType{FileBased},
		Fn: func(c *CheckRequest) CheckResult {
			return CreateRuntimeErrorResult(runner.CheckName, sce.WithMessage(sce.ErrScorecardInternal, "test error"))
		},
	}

	resErr := runner.Run(ctx, checkError)
	if resErr.Score != InconclusiveResultScore {
		t.Errorf("expected inconclusive score, got %d", resErr.Score)
	}

	var rm metricdata.ResourceMetrics
	if err := reader.Collect(ctx, &rm); err != nil {
		t.Fatalf("failed to collect metrics: %v", err)
	}

	runtimeMetric, ok := findMetric(&rm, "scorecard.check.runtime")
	if !ok {
		t.Error("scorecard.check.runtime metric not recorded by Runner.Run")
	} else if hist, ok := runtimeMetric.Data.(metricdata.Histogram[int64]); ok {
		for _, dp := range hist.DataPoints {
			if val, exists := dp.Attributes.Value(stats.CheckName); !exists || val.AsString() != runner.CheckName {
				t.Errorf("expected checkName %s, got %v", runner.CheckName, val)
			}
		}
	}

	errorMetric, ok := findMetric(&rm, "scorecard.check.errors")
	if !ok {
		t.Error("scorecard.check.errors metric not recorded for failed check")
	} else if sum, ok := errorMetric.Data.(metricdata.Sum[int64]); ok {
		for _, dp := range sum.DataPoints {
			if val, exists := dp.Attributes.Value(stats.CheckName); !exists || val.AsString() != runner.CheckName {
				t.Errorf("expected checkName %s, got %v", runner.CheckName, val)
			}
			if val, exists := dp.Attributes.Value(stats.ErrorName); !exists || val.AsString() != "ErrScorecardInternal" {
				t.Errorf("expected errorName ErrScorecardInternal, got %v", val)
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

func TestRunner_Run_Unsupported(t *testing.T) {
	t.Parallel()

	runner := &Runner{
		CheckName: "Unsupported-Check",
		CheckRequest: CheckRequest{
			RequiredTypes: []RequestType{
				CommitBased,
			},
		},
	}

	check := Check{
		SupportedRequestTypes: []RequestType{FileBased},
		Fn: func(c *CheckRequest) CheckResult {
			return CreateMaxScoreResult(runner.CheckName, "all good")
		},
	}

	res := runner.Run(context.Background(), check)
	if !errors.Is(res.Error, sce.ErrUnsupportedCheck) {
		t.Errorf("expected ErrUnsupportedCheck, got %v", res.Error)
	}
}

var _ clients.Repo = (*localdir.Repo)(nil)
