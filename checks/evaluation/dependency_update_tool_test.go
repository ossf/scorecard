// Copyright 2022 OpenSSF Scorecard Authors
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

package evaluation

import (
	"testing"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	scut "github.com/ossf/scorecard/v4/utests"
)

func TestDependencyUpdateTool(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name     string
		findings []finding.Finding
		err      bool
		want     checker.CheckResult
		expected scut.TestReturn
	}{
		{
			name: "dependabot",
			findings: []finding.Finding{
				{
					Probe:   "toolDependabotInstalled",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "toolPyUpInstalled",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "toolRenovateInstalled",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "toolSonatypeLiftInstalled",
					Outcome: finding.OutcomeNegative,
				},
			},
			want: checker.CheckResult{
				Score: 10,
			},
		},
		{
			name: "renovate",
			findings: []finding.Finding{
				{
					Probe:   "toolDependabotInstalled",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "toolPyUpInstalled",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "toolRenovateInstalled",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "toolSonatypeLiftInstalled",
					Outcome: finding.OutcomeNegative,
				},
			},
			want: checker.CheckResult{
				Score: 10,
			},
		},
		{
			name: "pyup",
			findings: []finding.Finding{
				{
					Probe:   "toolDependabotInstalled",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "toolPyUpInstalled",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "toolRenovateInstalled",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "toolSonatypeLiftInstalled",
					Outcome: finding.OutcomeNegative,
				},
			},
			want: checker.CheckResult{
				Score: 10,
			},
		},
		{
			name: "sonatype",
			findings: []finding.Finding{
				{
					Probe:   "toolDependabotInstalled",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "toolPyUpInstalled",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "toolRenovateInstalled",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "toolSonatypeLiftInstalled",
					Outcome: finding.OutcomePositive,
				},
			},
			want: checker.CheckResult{
				Score: 10,
			},
		},
		{
			name: "none",
			findings: []finding.Finding{
				{
					Probe:   "toolDependabotInstalled",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "toolRenovateInstalled",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "toolPyUpInstalled",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "toolSonatypeLiftInstalled",
					Outcome: finding.OutcomeNegative,
				},
			},
			want: checker.CheckResult{
				Score: 0,
			},
		},
		{
			name: "empty tool list",
			want: checker.CheckResult{
				Score: -1,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := DependencyUpdateTool(tt.name, tt.findings)
			if tt.want.Score != got.Score {
				t.Errorf("DependencyUpdateTool() got Score = %v, want %v for %v", got.Score, tt.want.Score, tt.name)
			}
			if tt.err && got.Error == nil {
				t.Errorf("DependencyUpdateTool() error = %v, want %v for %v", got.Error, tt.want.Error, tt.name)
				return
			}
		})
	}
}
