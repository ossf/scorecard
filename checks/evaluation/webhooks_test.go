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

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
	scut "github.com/ossf/scorecard/v5/utests"
)

// TestWebhooks tests the webhooks check.
func TestWebhooks(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		findings []finding.Finding
		result   scut.TestReturn
	}{
		{
			name: "no webhooks",
			findings: []finding.Finding{
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeNotApplicable,
				},
			},
			result: scut.TestReturn{
				Score: checker.MaxResultScore,
			},
		},
		{
			name: "1 webhook with no secret",
			findings: []finding.Finding{
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeFalse,
				},
			},
			result: scut.TestReturn{
				Score: checker.MinResultScore,
			},
		},
		{
			name: "1 webhook with secret",
			findings: []finding.Finding{
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeTrue,
				},
			},
			result: scut.TestReturn{
				Score: checker.MaxResultScore,
			},
		},
		{
			name: "2 webhooks one of which has secret",
			findings: []finding.Finding{
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeTrue,
				},
			},
			result: scut.TestReturn{
				Score: 5,
			},
		},
		{
			name: "Five webhooks three of which have secrets",
			findings: []finding.Finding{
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeFalse,
				},
			},
			result: scut.TestReturn{
				Score: 6,
			},
		},
		{
			name: "One of 12 webhooks does not have secrets",
			findings: []finding.Finding{
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeTrue,
				},
			},
			result: scut.TestReturn{
				Score: 9,
			},
		},
		{
			name: "Score should not drop below min score",
			findings: []finding.Finding{
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   "webhooksUseSecrets",
					Outcome: finding.OutcomeFalse,
				},
			},
			result: scut.TestReturn{
				Score: checker.MinResultScore,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			got := Webhooks(tt.name, tt.findings, &dl)
			scut.ValidateTestReturn(t, tt.name, &tt.result, &got, &dl)
		})
	}
}
