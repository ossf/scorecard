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

// TestWebhooks tests the webhooks check.
func TestWebhooks(t *testing.T) {
	t.Parallel()
	//nolint:govet
	type args struct {
		name string
		dl   checker.DetailLogger
		r    *checker.WebhooksData
	}
	tests := []struct {
		name     string
		findings []finding.Finding
		result   scut.TestReturn
	}{
		{
			name: "no webhooks",
			findings: []finding.Finding{
				{
					Probe:   "webhooksWithoutTokenAuth",
					Outcome: finding.OutcomeNotApplicable,
				},
			},
			result: scut.TestReturn{
				Score: 10,
			},
		},
		{
			name: "1 webhook with no secret",
			findings: []finding.Finding{
				{
					Probe:   "webhooksWithoutTokenAuth",
					Outcome: finding.OutcomeNegative,
				},
			},
			result: scut.TestReturn{
				Score: 0,
			},
		},
		{
			name: "1 webhook with secret",
			findings: []finding.Finding{
				{
					Probe:   "webhooksWithoutTokenAuth",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"totalWebhooks":         1,
						"webhooksWithoutSecret": 0,
					},
				},
			},
			result: scut.TestReturn{
				Score: 10,
			},
		},
		{
			name: "2 webhooks one of which has secret",
			findings: []finding.Finding{
				{
					Probe:   "webhooksWithoutTokenAuth",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"totalWebhooks":         2,
						"webhooksWithoutSecret": 1,
					},
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
					Probe:   "webhooksWithoutTokenAuth",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"totalWebhooks":         5,
						"webhooksWithoutSecret": 2,
					},
				},
			},
			result: scut.TestReturn{
				Score: 4,
			},
		},
		{
			name: "Twelve webhooks none of which have secrets",
			findings: []finding.Finding{
				{
					Probe:   "webhooksWithoutTokenAuth",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"totalWebhooks":         12,
						"webhooksWithoutSecret": 12,
					},
				},
			},
			result: scut.TestReturn{
				Score: 0,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			got := Webhooks(tt.name, tt.findings, &dl)
			if !scut.ValidateTestReturn(t, tt.name, &tt.result, &got, &dl) {
				t.Errorf("got %v, expected %v", got, tt.result)
			}
		})
	}
}
