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

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	scut "github.com/ossf/scorecard/v4/utests"
)

func TestDangerousWorkflow(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		findings []finding.Finding
		result   scut.TestReturn
	}{
		{
			name: "Has untrusted checkout workflow",
			findings: []finding.Finding{
				{
					Probe:   "hasDangerousWorkflowScriptInjection",
					Outcome: finding.OutcomePositive,
				}, {
					Probe:   "hasDangerousWorkflowUntrustedCheckout",
					Outcome: finding.OutcomeNegative,
				},
			},
			result: scut.TestReturn{
				Score: 0,
			},
		},
		{
			name: "DangerousWorkflow - no worklflows",
			findings: []finding.Finding{
				{
					Probe:   "hasDangerousWorkflowScriptInjection",
					Outcome: finding.OutcomeNotApplicable,
				}, {
					Probe:   "hasDangerousWorkflowUntrustedCheckout",
					Outcome: finding.OutcomeNotApplicable,
				},
			},
			result: scut.TestReturn{
				Score: checker.InconclusiveResultScore,
			},
		},
		{
			name: "DangerousWorkflow - found workflows, none dangerous",
			findings: []finding.Finding{
				{
					Probe:   "hasDangerousWorkflowScriptInjection",
					Outcome: finding.OutcomePositive,
				}, {
					Probe:   "hasDangerousWorkflowUntrustedCheckout",
					Outcome: finding.OutcomePositive,
				},
			},
			result: scut.TestReturn{
				Score: 10,
			},
		},
		{
			name: "DangerousWorkflow - Dangerous workflow detected",
			findings: []finding.Finding{
				{
					Probe:   "hasDangerousWorkflowScriptInjection",
					Outcome: finding.OutcomePositive,
				}, {
					Probe:   "hasDangerousWorkflowUntrustedCheckout",
					Outcome: finding.OutcomeNegative,
				},
			},
			result: scut.TestReturn{
				Score: 0,
			},
		},
		{
			name: "DangerousWorkflow - Script injection detected",
			findings: []finding.Finding{
				{
					Probe:   "hasDangerousWorkflowScriptInjection",
					Outcome: finding.OutcomeNegative,
				}, {
					Probe:   "hasDangerousWorkflowUntrustedCheckout",
					Outcome: finding.OutcomePositive,
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
			got := DangerousWorkflow(tt.name, tt.findings, &dl)
			if !scut.ValidateTestReturn(t, tt.name, &tt.result, &got, &dl) {
				t.Errorf("got %v, expected %v", got, tt.result)
			}
		})
	}
}
