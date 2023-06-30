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
)

func TestSecurityPolicy(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name     string
		findings []finding.Finding
		err      bool
		want     checker.CheckResult
	}{
		{
			name: "missing findings links",
			findings: []finding.Finding{
				{
					Probe:   "securityPolicyContainsDisclosure",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "securityPolicyContainsText",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "securityPolicyPresentInOrg",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "securityPolicyPresentInRepo",
					Outcome: finding.OutcomeNegative,
				},
			},
			want: checker.CheckResult{
				Score: -1,
			},
		},
		{
			name: "invalid findings",
			findings: []finding.Finding{
				{
					Probe:   "securityPolicyContainsDisclosure",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "securityPolicyContainsLinks",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "securityPolicyContainsText",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "securityPolicyPresentInOrg",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "securityPolicyPresentInRepo",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "securityPolicyInvalidProbeName",
					Outcome: finding.OutcomeNegative,
				},
			},
			want: checker.CheckResult{
				Score: -1,
			},
		},
		{
			name: "file found",
			findings: []finding.Finding{
				{
					Probe:   "securityPolicyContainsDisclosure",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "securityPolicyContainsLinks",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "securityPolicyContainsText",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "securityPolicyPresentInOrg",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "securityPolicyPresentInRepo",
					Outcome: finding.OutcomePositive,
				},
			},
			want: checker.CheckResult{
				Score: 0,
			},
		},
		{
			name: "file not found with positive probes",
			findings: []finding.Finding{
				{
					Probe:   "securityPolicyContainsDisclosure",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "securityPolicyContainsLinks",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "securityPolicyContainsText",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "securityPolicyPresentInOrg",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "securityPolicyPresentInRepo",
					Outcome: finding.OutcomeNegative,
				},
			},
			want: checker.CheckResult{
				Score: -1,
			},
		},
		{
			name: "file found with text",
			findings: []finding.Finding{
				{
					Probe:   "securityPolicyContainsDisclosure",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "securityPolicyContainsLinks",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "securityPolicyContainsText",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "securityPolicyPresentInOrg",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "securityPolicyPresentInRepo",
					Outcome: finding.OutcomePositive,
				},
			},
			want: checker.CheckResult{
				Score: 6,
			},
		},
		{
			name: "file found all positive",
			findings: []finding.Finding{
				{
					Probe:   "securityPolicyContainsDisclosure",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "securityPolicyContainsLinks",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "securityPolicyContainsText",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "securityPolicyPresentInOrg",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "securityPolicyPresentInRepo",
					Outcome: finding.OutcomePositive,
				},
			},
			want: checker.CheckResult{
				Score: 10,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := SecurityPolicy("SecurityPolicy", tt.findings)
			if tt.err {
				if got.Score != -1 {
					t.Errorf("SecurityPolicy() = %v, want %v", got, tt.want)
				}
			}
			if got.Score != tt.want.Score {
				t.Errorf("SecurityPolicy() = %v, want %v for %v", got.Score, tt.want.Score, tt.name)
			}
		})
	}
}
