// Copyright 2024 OpenSSF Scorecard Authors
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

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
	scut "github.com/ossf/scorecard/v5/utests"
)

func TestChangelog(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		findings []finding.Finding
		result   scut.TestReturn
	}{
		{
			name: "Path A: changelog file + all releases covered. 10/10",
			findings: []finding.Finding{
				{
					Probe:   "hasChangelogFile",
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   "releasesHaveChangelog",
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						"releasesWithChangelog": "5",
						"releasesTotal":         "5",
					},
				},
			},
			result: scut.TestReturn{
				Score:        checker.MaxResultScore, // 3 + 7 = 10
				NumberOfInfo: 2,
			},
		},
		{
			name: "Path A: changelog file + 3/5 releases covered. 7/10",
			findings: []finding.Finding{
				{
					Probe:   "hasChangelogFile",
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   "releasesHaveChangelog",
					Outcome: finding.OutcomeFalse,
					Values: map[string]string{
						"releasesWithChangelog": "3",
						"releasesTotal":         "5",
					},
				},
			},
			result: scut.TestReturn{
				Score:        changelogFileScore + 4,
				NumberOfInfo: 1,
				NumberOfWarn: 1,
			},
		},
		{
			name: "Path A: changelog file + no releases covered. 3/10",
			findings: []finding.Finding{
				{
					Probe:   "hasChangelogFile",
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   "releasesHaveChangelog",
					Outcome: finding.OutcomeFalse,
					Values: map[string]string{
						"releasesWithChangelog": "0",
						"releasesTotal":         "5",
					},
				},
			},
			result: scut.TestReturn{
				Score:        changelogFileScore, // 3 + 0 = 3
				NumberOfInfo: 1,
				NumberOfWarn: 1,
			},
		},
		{
			name: "Path A: changelog file + no releases. 3/10",
			findings: []finding.Finding{
				{
					Probe:   "hasChangelogFile",
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   "releasesHaveChangelog",
					Outcome: finding.OutcomeNotApplicable,
				},
			},
			result: scut.TestReturn{
				Score:         changelogFileScore, // 3
				NumberOfInfo:  1,
				NumberOfDebug: 1,
			},
		},
		{
			name: "Path B: no file, all releases have notes. 10/10",
			findings: []finding.Finding{
				{
					Probe:   "hasChangelogFile",
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   "releasesHaveChangelog",
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						"releasesWithChangelog": "5",
						"releasesTotal":         "5",
					},
				},
			},
			result: scut.TestReturn{
				Score:        checker.MaxResultScore, // 10 * 5/5 = 10
				NumberOfWarn: 1,
				NumberOfInfo: 1,
			},
		},
		{
			name: "Path B: no file, 3/5 releases have notes. 6/10",
			findings: []finding.Finding{
				{
					Probe:   "hasChangelogFile",
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   "releasesHaveChangelog",
					Outcome: finding.OutcomeFalse,
					Values: map[string]string{
						"releasesWithChangelog": "3",
						"releasesTotal":         "5",
					},
				},
			},
			result: scut.TestReturn{
				Score:        6,
				NumberOfWarn: 2,
			},
		},
		{
			name: "Inconclusive: no file, no releases",
			findings: []finding.Finding{
				{
					Probe:   "hasChangelogFile",
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   "releasesHaveChangelog",
					Outcome: finding.OutcomeNotApplicable,
				},
			},
			result: scut.TestReturn{
				Score:         checker.InconclusiveResultScore,
				NumberOfWarn:  1,
				NumberOfDebug: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			got := Changelog(tt.name, tt.findings, &dl)
			scut.ValidateTestReturn(t, tt.name, &tt.result, &got, &dl)
		})
	}
}
