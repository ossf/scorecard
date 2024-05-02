// Copyright 2021 OpenSSF Scorecard Authors
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
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/codeApproved"
	scut "github.com/ossf/scorecard/v5/utests"
)

func TestCodeReview(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		findings []finding.Finding
		expected scut.TestReturn
	}{
		{
			name: "no findings is an error",
			expected: scut.TestReturn{
				Error: sce.ErrScorecardInternal,
				Score: checker.InconclusiveResultScore,
			},
			findings: nil,
		},
		{
			name: "no changesets",
			expected: scut.TestReturn{
				Score: checker.InconclusiveResultScore,
			},
			findings: []finding.Finding{
				{
					Probe:   codeApproved.Probe,
					Outcome: finding.OutcomeNotApplicable,
					Message: "no changesets detected",
				},
			},
		},
		{
			name: "unreviewed changes result in minimum score",
			expected: scut.TestReturn{
				Score:         checker.MinResultScore,
				NumberOfDebug: 0, // TODO
			},
			findings: []finding.Finding{
				{
					Probe:   codeApproved.Probe,
					Outcome: finding.OutcomeFalse,
					Message: "Found 0/2 approved changesets",
					Values: map[string]string{
						codeApproved.NumApprovedKey: "0",
						codeApproved.NumTotalKey:    "2",
					},
				},
			},
		},
		{
			name: "all changesets reviewed",
			expected: scut.TestReturn{
				Score: checker.MaxResultScore,
			},
			findings: []finding.Finding{
				{
					Probe:   codeApproved.Probe,
					Outcome: finding.OutcomeTrue,
					Message: "All changesets approved",
					Values: map[string]string{
						codeApproved.NumApprovedKey: "2",
						codeApproved.NumTotalKey:    "2",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := &scut.TestDetailLogger{}
			res := CodeReview(tt.name, tt.findings, dl)
			scut.ValidateTestReturn(t, tt.name, &tt.expected, &res, dl)
		})
	}
}
