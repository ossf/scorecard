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
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
	scut "github.com/ossf/scorecard/v4/utests"
)

func TestSecurityPolicy(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name     string
		findings []finding.Finding
		result   scut.TestReturn
	}{
		{
			name: "missing findings links",
			findings: []finding.Finding{
				{
					Probe:   "securityPolicyContainsVulnerabilityDisclosure",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "securityPolicyContainsText",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "securityPolicyPresent",
					Outcome: finding.OutcomeNegative,
				},
			},
			result: scut.TestReturn{
				Score: checker.InconclusiveResultScore,
				Error: sce.ErrScorecardInternal,
			},
		},
		{
			name: "invalid probe name",
			findings: []finding.Finding{
				{
					Probe:   "securityPolicyContainsVulnerabilityDisclosure",
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
					Probe:   "securityPolicyPresent",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "securityPolicyInvalidProbeName",
					Outcome: finding.OutcomeNegative,
				},
			},
			result: scut.TestReturn{
				Score: checker.InconclusiveResultScore,
				Error: sce.ErrScorecardInternal,
			},
		},
		{
			name: "file found only",
			findings: []finding.Finding{
				{
					Probe:   "securityPolicyContainsVulnerabilityDisclosure",
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
					Probe:   "securityPolicyPresent",
					Outcome: finding.OutcomePositive,
				},
			},
			result: scut.TestReturn{
				Score:        checker.MinResultScore,
				NumberOfInfo: 1,
				NumberOfWarn: 3,
			},
		},
		{
			name: "file not found with positive probes",
			findings: []finding.Finding{
				{
					Probe:   "securityPolicyContainsVulnerabilityDisclosure",
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
					Probe:   "securityPolicyPresent",
					Outcome: finding.OutcomeNegative,
				},
			},
			result: scut.TestReturn{
				Score: checker.InconclusiveResultScore,
				Error: sce.ErrScorecardInternal,
			},
		},
		{
			name: "file found with no disclosure and text",
			findings: []finding.Finding{
				{
					Probe:   "securityPolicyContainsVulnerabilityDisclosure",
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
					Probe:   "securityPolicyPresent",
					Outcome: finding.OutcomePositive,
				},
			},
			result: scut.TestReturn{
				Score:        6,
				NumberOfInfo: 2,
				NumberOfWarn: 2,
			},
		},
		{
			name: "file found all positive",
			findings: []finding.Finding{
				{
					Probe:   "securityPolicyContainsVulnerabilityDisclosure",
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
					Probe:   "securityPolicyPresent",
					Outcome: finding.OutcomePositive,
				},
			},
			result: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 4,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			got := SecurityPolicy(tt.name, tt.findings, &dl)
			if !scut.ValidateTestReturn(t, tt.name, &tt.result, &got, &dl) {
				t.Errorf("got %v, expected %v", got, tt.result)
			}
		})
	}
}
