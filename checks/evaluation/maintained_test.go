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

	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
	scut "github.com/ossf/scorecard/v4/utests"
)

func TestMaintained(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		findings []finding.Finding
		result   scut.TestReturn
	}{
		{
			name: "Two commits in last 90 days",
			findings: []finding.Finding{
				{
					Probe:   "hasRecentCommits",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"commitsWithinThreshold": 2,
					},
				}, {
					Probe:   "issueActivityByProjectMember",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"numberOfIssuesUpdatedWithinThreshold": 1,
					},
				}, {
					Probe:   "notArchived",
					Outcome: finding.OutcomePositive,
				}, {
					Probe:   "notCreatedRecently",
					Outcome: finding.OutcomePositive,
				},
			},
			result: scut.TestReturn{
				Score: 2,
			},
		},
		{
			name: "No issues, no commits and not archived",
			findings: []finding.Finding{
				{
					Probe:   "hasRecentCommits",
					Outcome: finding.OutcomeNegative,
				}, {
					Probe:   "issueActivityByProjectMember",
					Outcome: finding.OutcomeNegative,
				}, {
					Probe:   "notArchived",
					Outcome: finding.OutcomePositive,
				}, {
					Probe:   "notCreatedRecently",
					Outcome: finding.OutcomePositive,
				},
			},
			result: scut.TestReturn{
				Score: 0,
			},
		},
		{
			name: "Wrong probe name",
			findings: []finding.Finding{
				{
					Probe:   "hasRecentCommits",
					Outcome: finding.OutcomeNegative,
				}, {
					Probe:   "issueActivityByProjectMember",
					Outcome: finding.OutcomeNegative,
				}, {
					Probe:   "archvied", /*misspelling*/
					Outcome: finding.OutcomePositive,
				}, {
					Probe:   "notCreatedRecently",
					Outcome: finding.OutcomePositive,
				},
			},
			result: scut.TestReturn{
				Score: -1,
				Error: sce.ErrScorecardInternal,
			},
		},
		{
			name: "Project is archived",
			findings: []finding.Finding{
				{
					Probe:   "hasRecentCommits",
					Outcome: finding.OutcomeNegative,
				}, {
					Probe:   "issueActivityByProjectMember",
					Outcome: finding.OutcomeNegative,
				}, {
					Probe:   "notArchived",
					Outcome: finding.OutcomeNegative,
				}, {
					Probe:   "notCreatedRecently",
					Outcome: finding.OutcomePositive,
				},
			},
			result: scut.TestReturn{
				Score:        0,
				NumberOfWarn: 3,
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			got := Maintained(tt.name, tt.findings, &dl)
			scut.ValidateTestReturn(t, tt.name, &tt.result, &got, &dl)
		})
	}
}
