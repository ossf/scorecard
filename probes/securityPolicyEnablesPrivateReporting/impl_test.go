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

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
)

func TestRun(t *testing.T) {
	t.Run("returns not applicable when nil", func(t *testing.T) {
		raw := &checker.RawResults{
			SecurityPolicyResults: checker.SecurityPolicyData{
				PrivateVulnerabilityReportingEnabled: nil,
			},
			Metadata: checker.MetadataData{Metadata: map[string]string{}},
		}
		findings, probe, err := Run(raw)
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
		if probe != Probe {
			t.Errorf("Run() probe = %v, want %v", probe, Probe)
		}
		if len(findings) != 1 {
			t.Fatalf("Run() findings length = %v, want 1", len(findings))
		}
		if findings[0].Outcome != finding.OutcomeNotApplicable {
			t.Errorf("Run() outcome = %v, want %v", findings[0].Outcome, finding.OutcomeNotApplicable)
		}
	})

	t.Run("returns true when enabled", func(t *testing.T) {
		enabled := true
		raw := &checker.RawResults{
			SecurityPolicyResults: checker.SecurityPolicyData{
				PrivateVulnerabilityReportingEnabled: &enabled,
			},
			Metadata: checker.MetadataData{Metadata: map[string]string{}},
		}
		findings, _, err := Run(raw)
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
		if findings[0].Outcome != finding.OutcomeTrue {
			t.Errorf("Run() outcome = %v, want %v", findings[0].Outcome, finding.OutcomeTrue)
		}
	})

	t.Run("returns false when disabled", func(t *testing.T) {
		disabled := false
		raw := &checker.RawResults{
			SecurityPolicyResults: checker.SecurityPolicyData{
				PrivateVulnerabilityReportingEnabled: &disabled,
			},
			Metadata: checker.MetadataData{Metadata: map[string]string{}},
		}
		findings, _, err := Run(raw)
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
		if findings[0].Outcome != finding.OutcomeFalse {
			t.Errorf("Run() outcome = %v, want %v", findings[0].Outcome, finding.OutcomeFalse)
		}
	})
}
