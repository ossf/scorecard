// Copyright 2023 OpenSSF Scorecard Authors
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
package evaluation

import (
	"testing"

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
	scut "github.com/ossf/scorecard/v4/utests"
)

func TestSAST(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		findings []finding.Finding
		result   scut.TestReturn
	}{
		{
			name: "SAST - Missing a probe",
			findings: []finding.Finding{
				{
					Probe:   "sastToolCodeQLInstalled",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "sastToolRunsOnAllCommits",
					Outcome: finding.OutcomePositive,
				},
			},
			result: scut.TestReturn{
				Score: checker.InconclusiveResultScore,
				Error: sce.ErrScorecardInternal,
			},
		},
		{
			name: "Sonar and CodeCQ is installed",
			findings: []finding.Finding{
				{
					Probe:   "sastToolCodeQLInstalled",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "sastToolRunsOnAllCommits",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"totalTested": 1,
						"totalMerged": 2,
					},
				},
				{
					Probe:   "sastToolSonarInstalled",
					Outcome: finding.OutcomePositive,
				},
			},
			result: scut.TestReturn{
				Score:        10,
				NumberOfInfo: 3,
			},
		},
		{
			name: `Sonar is installed. CodeQL is not installed.
					Does not have info about whether SAST runs
					on every commit.`,
			findings: []finding.Finding{
				{
					Probe:   "sastToolCodeQLInstalled",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "sastToolRunsOnAllCommits",
					Outcome: finding.OutcomeNotApplicable,
				},
				{
					Probe:   "sastToolSonarInstalled",
					Outcome: finding.OutcomePositive,
				},
			},
			result: scut.TestReturn{
				Score:        10,
				NumberOfInfo: 1,
				NumberOfWarn: 2,
			},
		},
		{
			name: "Sonar and CodeQL are not installed",
			findings: []finding.Finding{
				{
					Probe:   "sastToolCodeQLInstalled",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "sastToolRunsOnAllCommits",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"totalTested": 1,
						"totalMerged": 3,
					},
				},
				{
					Probe:   "sastToolSonarInstalled",
					Outcome: finding.OutcomeNegative,
				},
			},
			result: scut.TestReturn{
				Score:        3,
				NumberOfWarn: 3,
				NumberOfInfo: 0,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			got := SAST(tt.name, tt.findings, &dl)
			if !scut.ValidateTestReturn(t, tt.name, &tt.result, &got, &dl) {
				t.Errorf("got %v, expected %v", got, tt.result)
			}
		})
	}
}
