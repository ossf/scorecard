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

	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/hasOpenSSFBadge"
	scut "github.com/ossf/scorecard/v5/utests"
)

func TestCIIBestPractices(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		findings []finding.Finding
		result   scut.TestReturn
	}{
		{
			name: "Unsupported badge found with false finding",
			findings: []finding.Finding{
				{
					Probe:   "hasOpenSSFBadge",
					Outcome: finding.OutcomeFalse,
					Values: map[string]string{
						hasOpenSSFBadge.LevelKey: "Unsupported",
					},
				},
			},
			result: scut.TestReturn{
				Score: 0,
			},
		},
		{
			name: "Unsupported badge found with true finding",
			findings: []finding.Finding{
				{
					Probe:   "hasOpenSSFBadge",
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						hasOpenSSFBadge.LevelKey: "Unsupported",
					},
				},
			},
			result: scut.TestReturn{
				Score: -1,
				Error: sce.ErrScorecardInternal,
			},
		},
		{
			name: "Has InProgress Badge",
			findings: []finding.Finding{
				{
					Probe:   "hasOpenSSFBadge",
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						hasOpenSSFBadge.LevelKey: hasOpenSSFBadge.InProgressLevel,
					},
				},
			},
			result: scut.TestReturn{
				Score: 2,
			},
		},
		{
			name: "Has Passing Badge",
			findings: []finding.Finding{
				{
					Probe:   "hasOpenSSFBadge",
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						hasOpenSSFBadge.LevelKey: hasOpenSSFBadge.PassingLevel,
					},
				},
			},
			result: scut.TestReturn{
				Score: 5,
			},
		},
		{
			name: "Has Silver Badge",
			findings: []finding.Finding{
				{
					Probe:   "hasOpenSSFBadge",
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						hasOpenSSFBadge.LevelKey: hasOpenSSFBadge.SilverLevel,
					},
				},
			},
			result: scut.TestReturn{
				Score: 7,
			},
		},
		{
			name: "Has Gold Badge",
			findings: []finding.Finding{
				{
					Probe:   "hasOpenSSFBadge",
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						hasOpenSSFBadge.LevelKey: hasOpenSSFBadge.GoldLevel,
					},
				},
			},
			result: scut.TestReturn{
				Score: 10,
			},
		},
		{
			name: "Has Unknown Badge",
			findings: []finding.Finding{
				{
					Probe:   "hasOpenSSFBadge",
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						hasOpenSSFBadge.LevelKey: hasOpenSSFBadge.UnknownLevel,
					},
				},
			},
			result: scut.TestReturn{
				Score: -1,
				Error: sce.ErrScorecardInternal,
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			got := CIIBestPractices(tt.name, tt.findings, &dl)
			scut.ValidateTestReturn(t, tt.name, &tt.result, &got, &dl)
		})
	}
}
