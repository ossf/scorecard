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

	"github.com/ossf/scorecard/v5/checker"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/archived"
	"github.com/ossf/scorecard/v5/probes/createdRecently"
	"github.com/ossf/scorecard/v5/probes/hasInactiveMaintainers"
	"github.com/ossf/scorecard/v5/probes/hasRecentCommits"
	"github.com/ossf/scorecard/v5/probes/issueActivityByProjectMember"
	scut "github.com/ossf/scorecard/v5/utests"
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
					Probe:   hasRecentCommits.Probe,
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						hasRecentCommits.NumCommitsKey: "2",
					},
				}, {
					Probe:   issueActivityByProjectMember.Probe,
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						issueActivityByProjectMember.NumIssuesKey: "1",
					},
				}, {
					Probe:   archived.Probe,
					Outcome: finding.OutcomeFalse,
				}, {
					Probe:   createdRecently.Probe,
					Outcome: finding.OutcomeFalse,
				}, {
					Probe:   hasInactiveMaintainers.Probe,
					Outcome: finding.OutcomeFalse,
				},
			},
			result: scut.TestReturn{
				Score:        2,
				NumberOfInfo: 1,
			},
		},
		{
			name: "No issues, no commits and not archived",
			findings: []finding.Finding{
				{
					Probe:   hasRecentCommits.Probe,
					Outcome: finding.OutcomeFalse,
				}, {
					Probe:   issueActivityByProjectMember.Probe,
					Outcome: finding.OutcomeFalse,
				}, {
					Probe:   archived.Probe,
					Outcome: finding.OutcomeFalse,
				}, {
					Probe:   createdRecently.Probe,
					Outcome: finding.OutcomeFalse,
				}, {
					Probe:   hasInactiveMaintainers.Probe,
					Outcome: finding.OutcomeFalse,
				},
			},
			result: scut.TestReturn{
				Score:        0,
				NumberOfInfo: 1,
			},
		},
		{
			name: "Wrong probe name",
			findings: []finding.Finding{
				{
					Probe:   hasRecentCommits.Probe,
					Outcome: finding.OutcomeFalse,
				}, {
					Probe:   issueActivityByProjectMember.Probe,
					Outcome: finding.OutcomeFalse,
				}, {
					Probe:   "archvied", /*misspelling*/
					Outcome: finding.OutcomeTrue,
				}, {
					Probe:   createdRecently.Probe,
					Outcome: finding.OutcomeFalse,
				}, {
					Probe:   hasInactiveMaintainers.Probe,
					Outcome: finding.OutcomeFalse,
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
					Probe:   hasRecentCommits.Probe,
					Outcome: finding.OutcomeFalse,
				}, {
					Probe:   issueActivityByProjectMember.Probe,
					Outcome: finding.OutcomeFalse,
				}, {
					Probe:   archived.Probe,
					Outcome: finding.OutcomeTrue,
				}, {
					Probe:   createdRecently.Probe,
					Outcome: finding.OutcomeFalse,
				}, {
					Probe:   hasInactiveMaintainers.Probe,
					Outcome: finding.OutcomeFalse,
				},
			},
			result: scut.TestReturn{
				Score:        0,
				NumberOfWarn: 1,
				NumberOfInfo: 1,
			},
		},
		{
			name: "recently created projects get min score",
			findings: []finding.Finding{
				{
					Probe:   hasRecentCommits.Probe,
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						hasRecentCommits.NumCommitsKey: "20",
					},
				}, {
					Probe:   issueActivityByProjectMember.Probe,
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						issueActivityByProjectMember.NumIssuesKey: "10",
					},
				}, {
					Probe:   archived.Probe,
					Outcome: finding.OutcomeFalse,
				}, {
					Probe:   createdRecently.Probe,
					Outcome: finding.OutcomeTrue,
				}, {
					Probe:   hasInactiveMaintainers.Probe,
					Outcome: finding.OutcomeFalse,
				},
			},
			result: scut.TestReturn{
				Score:        checker.MinResultScore,
				NumberOfWarn: 1,
				NumberOfInfo: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			got := Maintained(tt.name, tt.findings, &dl)
			scut.ValidateTestReturn(t, tt.name, &tt.result, &got, &dl)
		})
	}
}
