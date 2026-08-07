// Copyright 2023 OpenSSF Scorecard Authors
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

package securityPolicyEnablesPrivateReporting

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
)

func Test_Run(t *testing.T) {
	t.Parallel()
	
	trueVal := true
	falseVal := false

	tests := []struct {
		name     string
		raw      *checker.RawResults
		expected []finding.Finding
	}{
		{
			name: "Private vulnerability reporting is enabled",
			raw: &checker.RawResults{
				SecurityPolicyResults: checker.SecurityPolicyData{
					PrivateVulnerabilityReportingEnabled: &trueVal,
				},
			},
			expected: []finding.Finding{
				{
					Probe:   Probe,
					Outcome: finding.OutcomeTrue,
					Message: "Private vulnerability reporting is enabled",
				},
			},
		},
		{
			name: "Private vulnerability reporting is disabled",
			raw: &checker.RawResults{
				SecurityPolicyResults: checker.SecurityPolicyData{
					PrivateVulnerabilityReportingEnabled: &falseVal,
				},
			},
			expected: []finding.Finding{
				{
					Probe:   Probe,
					Outcome: finding.OutcomeFalse,
					Message: "Private vulnerability reporting is disabled",
				},
			},
		},
		{
			name: "Private vulnerability reporting is unknown",
			raw: &checker.RawResults{
				SecurityPolicyResults: checker.SecurityPolicyData{
					PrivateVulnerabilityReportingEnabled: nil,
				},
			},
			expected: []finding.Finding{
				{
					Probe:   Probe,
					Outcome: finding.OutcomeNotApplicable,
					Message: "Private vulnerability reporting status is unknown or unsupported on this platform",
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			findings, _, err := Run(tt.raw)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.expected, findings, cmpopts.IgnoreFields(finding.Finding{}, "Location")); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
