// Copyright 2025 OpenSSF Scorecard Authors
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

package MTTUDependenciesIsLow

import (
	"testing"
	"time"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
)

var epoch = time.Unix(0, 0).UTC()

func enc(d time.Duration) time.Time { return epoch.Add(d) }
func boolp(b bool) *bool            { return &b }

//nolint:paralleltest // probe test is independent
func TestRun_LowTrue_WhenMeanInWindow(t *testing.T) {
	// mean ~ 30d → in [14d, 180d)
	raw := &checker.RawResults{
		MTTUDependenciesResults: checker.MTTUDependenciesData{
			Dependencies: []checker.LockDependency{
				{Name: "a", Version: "1.0.0", IsLatest: boolp(false), TimeSinceOldestReleast: enc(30 * 24 * time.Hour)},
				{Name: "b", Version: "1.0.0", IsLatest: boolp(true)},
			},
		},
	}
	ff, _, err := Run(raw)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if ff[0].Outcome != finding.OutcomeTrue {
		t.Fatalf("expected OutcomeTrue, got %v", ff[0].Outcome)
	}
}

//nolint:paralleltest // probe test is independent
func TestRun_LowFalse_WhenMeanBelow2Weeks(t *testing.T) {
	raw := &checker.RawResults{
		MTTUDependenciesResults: checker.MTTUDependenciesData{
			Dependencies: []checker.LockDependency{
				{Name: "a", Version: "1.0.0", IsLatest: boolp(false), TimeSinceOldestReleast: enc(10 * 24 * time.Hour)},
			},
		},
	}
	ff, _, err := Run(raw)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if ff[0].Outcome != finding.OutcomeFalse {
		t.Fatalf("expected OutcomeFalse, got %v", ff[0].Outcome)
	}
}

//nolint:paralleltest // probe test is independent
func TestRun_LowFalse_WhenMeanGE6Months(t *testing.T) {
	raw := &checker.RawResults{
		MTTUDependenciesResults: checker.MTTUDependenciesData{
			Dependencies: []checker.LockDependency{
				{Name: "a", Version: "1.0.0", IsLatest: boolp(false), TimeSinceOldestReleast: enc(200 * 24 * time.Hour)},
			},
		},
	}
	ff, _, err := Run(raw)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if ff[0].Outcome != finding.OutcomeFalse {
		t.Fatalf("expected OutcomeFalse, got %v", ff[0].Outcome)
	}
}
