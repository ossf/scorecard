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

var (
	testSnippet   = "other/checkout@a81bbbf8298c0fa03ea29cdc473d45769f953675"
	testLineStart = uint(123)
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
					Location: &finding.Location{
						Type:      finding.FileTypeText,
						Path:      "./github/workflows/dangerous-workflow.yml",
						LineStart: &testLineStart,
						Snippet:   &testSnippet,
					},
				},
			},
			result: scut.TestReturn{
				Score:        0,
				NumberOfWarn: 1,
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
					Location: &finding.Location{
						Type:      finding.FileTypeText,
						Path:      "./github/workflows/dangerous-workflow.yml",
						LineStart: &testLineStart,
						Snippet:   &testSnippet,
					},
				},
			},
			result: scut.TestReturn{
				Score:        0,
				NumberOfWarn: 1,
			},
		},
		{
			name: "DangerousWorkflow - Script injection detected",
			findings: []finding.Finding{
				{
					Probe:   "hasDangerousWorkflowScriptInjection",
					Outcome: finding.OutcomeNegative,
					Location: &finding.Location{
						Type:      finding.FileTypeText,
						Path:      "./github/workflows/dangerous-workflow.yml",
						LineStart: &testLineStart,
						Snippet:   &testSnippet,
					},
				}, {
					Probe:   "hasDangerousWorkflowUntrustedCheckout",
					Outcome: finding.OutcomePositive,
				},
			},
			result: scut.TestReturn{
				Score:        0,
				NumberOfWarn: 1,
			},
		},
		{
			name: "DangerousWorkflow - 3 script injection workflows detected",
			findings: []finding.Finding{
				{
					Probe:   "hasDangerousWorkflowScriptInjection",
					Outcome: finding.OutcomeNegative,
					Location: &finding.Location{
						Type:      finding.FileTypeText,
						Path:      "./github/workflows/dangerous-workflow.yml",
						LineStart: &testLineStart,
						Snippet:   &testSnippet,
					},
				}, {
					Probe:   "hasDangerousWorkflowScriptInjection",
					Outcome: finding.OutcomeNegative,
					Location: &finding.Location{
						Type:      finding.FileTypeText,
						Path:      "./github/workflows/dangerous-workflow2.yml",
						LineStart: &testLineStart,
						Snippet:   &testSnippet,
					},
				}, {
					Probe:   "hasDangerousWorkflowUntrustedCheckout",
					Outcome: finding.OutcomePositive,
				},
			},
			result: scut.TestReturn{
				Score:        0,
				NumberOfWarn: 2,
			},
		},
		{
			name: "DangerousWorkflow - 8 script injection workflows detected",
			findings: []finding.Finding{
				{
					Probe:   "hasDangerousWorkflowScriptInjection",
					Outcome: finding.OutcomeNegative,
					Location: &finding.Location{
						Type:      finding.FileTypeText,
						Path:      "./github/workflows/dangerous-workflow.yml",
						LineStart: &testLineStart,
						Snippet:   &testSnippet,
					},
				}, {
					Probe:   "hasDangerousWorkflowScriptInjection",
					Outcome: finding.OutcomeNegative,
					Location: &finding.Location{
						Type:      finding.FileTypeText,
						Path:      "./github/workflows/dangerous-workflow2.yml",
						LineStart: &testLineStart,
						Snippet:   &testSnippet,
					},
				}, {
					Probe:   "hasDangerousWorkflowScriptInjection",
					Outcome: finding.OutcomeNegative,
					Location: &finding.Location{
						Type:      finding.FileTypeText,
						Path:      "./github/workflows/dangerous-workflow3.yml",
						LineStart: &testLineStart,
						Snippet:   &testSnippet,
					},
				}, {
					Probe:   "hasDangerousWorkflowScriptInjection",
					Outcome: finding.OutcomeNegative,
					Location: &finding.Location{
						Type:      finding.FileTypeText,
						Path:      "./github/workflows/dangerous-workflow4.yml",
						LineStart: &testLineStart,
						Snippet:   &testSnippet,
					},
				}, {
					Probe:   "hasDangerousWorkflowScriptInjection",
					Outcome: finding.OutcomeNegative,
					Location: &finding.Location{
						Type:      finding.FileTypeText,
						Path:      "./github/workflows/dangerous-workflow5.yml",
						LineStart: &testLineStart,
						Snippet:   &testSnippet,
					},
				}, {
					Probe:   "hasDangerousWorkflowScriptInjection",
					Outcome: finding.OutcomeNegative,
					Location: &finding.Location{
						Type:      finding.FileTypeText,
						Path:      "./github/workflows/dangerous-workflow6.yml",
						LineStart: &testLineStart,
						Snippet:   &testSnippet,
					},
				}, {
					Probe:   "hasDangerousWorkflowScriptInjection",
					Outcome: finding.OutcomeNegative,
					Location: &finding.Location{
						Type:      finding.FileTypeText,
						Path:      "./github/workflows/dangerous-workflow7.yml",
						LineStart: &testLineStart,
						Snippet:   &testSnippet,
					},
				}, {
					Probe:   "hasDangerousWorkflowScriptInjection",
					Outcome: finding.OutcomeNegative,
					Location: &finding.Location{
						Type:      finding.FileTypeText,
						Path:      "./github/workflows/dangerous-workflow8.yml",
						LineStart: &testLineStart,
						Snippet:   &testSnippet,
					},
				}, {
					Probe:   "hasDangerousWorkflowUntrustedCheckout",
					Outcome: finding.OutcomePositive,
				},
			},
			result: scut.TestReturn{
				Score:        0,
				NumberOfWarn: 8,
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
