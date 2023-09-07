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
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
	scut "github.com/ossf/scorecard/v4/utests"
)

func TestDependencyUpdateTool(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name     string
		findings []finding.Finding
		result   scut.TestReturn
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
			result: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 1,
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
			result: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 1,
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
			result: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 1,
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
				{
					Probe:   "toolRenovateInstalled",
					Outcome: finding.OutcomeNegative,
				},
			},
			result: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 1,
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
			result: scut.TestReturn{
				Score:        checker.MinResultScore,
				NumberOfWarn: 4,
			},
		},
		{
			name: "missing probes renovate",
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
					Probe:   "toolSonatypeInstalled",
					Outcome: finding.OutcomeNegative,
				},
			},
			result: scut.TestReturn{
				Score: checker.InconclusiveResultScore,
				Error: sce.ErrScorecardInternal,
			},
		},
		{
			name: "invalid probe name",
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
					Probe:   "toolSonatypeInstalled",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "toolInvalidProbeName",
					Outcome: finding.OutcomeNegative,
				},
			},
			result: scut.TestReturn{
				Score: checker.InconclusiveResultScore,
				Error: sce.ErrScorecardInternal,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dl := scut.TestDetailLogger{}
			got := DependencyUpdateTool(tt.name, tt.findings, &dl)
			if !scut.ValidateTestReturn(t, tt.name, &tt.result, &got, &dl) {
				t.Fatalf(tt.name)
			}
		})
	}
}
