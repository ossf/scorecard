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
					Probe:   "commitsInLast90Days",
					Outcome: finding.OutcomePositive,
				}, {
					Probe:   "commitsInLast90Days",
					Outcome: finding.OutcomePositive,
				}, {
					Probe:   "activityOnIssuesByCollaboratorsMembersOrOwnersInLast90Days",
					Outcome: finding.OutcomePositive,
				}, {
					Probe:   "archived",
					Outcome: finding.OutcomePositive,
				}, {
					Probe:   "wasCreatedInLast90Days",
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
					Probe:   "commitsInLast90Days",
					Outcome: finding.OutcomeNegative,
				}, {
					Probe:   "activityOnIssuesByCollaboratorsMembersOrOwnersInLast90Days",
					Outcome: finding.OutcomeNegative,
				}, {
					Probe:   "archived",
					Outcome: finding.OutcomePositive,
				}, {
					Probe:   "wasCreatedInLast90Days",
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
					Probe:   "commitsInLast90Days",
					Outcome: finding.OutcomeNegative,
				}, {
					Probe:   "activityOnIssuesByCollaboratorsMembersOrOwnersInLast90Days",
					Outcome: finding.OutcomeNegative,
				}, {
					Probe:   "archvied", /*misspelling*/
					Outcome: finding.OutcomePositive,
				}, {
					Probe:   "wasCreatedInLast90Days",
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
					Probe:   "commitsInLast90Days",
					Outcome: finding.OutcomeNegative,
				}, {
					Probe:   "activityOnIssuesByCollaboratorsMembersOrOwnersInLast90Days",
					Outcome: finding.OutcomeNegative,
				}, {
					Probe:   "archived",
					Outcome: finding.OutcomeNegative,
				}, {
					Probe:   "wasCreatedInLast90Days",
					Outcome: finding.OutcomePositive,
				},
			},
			result: scut.TestReturn{
				Score:        0,
				NumberOfInfo: 1,
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			got := Maintained(tt.name, tt.findings, &dl)
			if !scut.ValidateTestReturn(t, tt.name, &tt.result, &got, &dl) {
				t.Errorf("got %v, expected %v", got, tt.result)
			}
		})
	}
}
