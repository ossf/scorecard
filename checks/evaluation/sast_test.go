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

	"github.com/ossf/scorecard/v5/checker"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/sastToolConfigured"
	"github.com/ossf/scorecard/v5/probes/sastToolRunsOnAllCommits"
	scut "github.com/ossf/scorecard/v5/utests"
)

func TestSAST(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		findings []finding.Finding
		result   scut.TestReturn
	}{
		{
			name: "SAST - Missing a probe (sastToolConfigured)",
			findings: []finding.Finding{
				{
					Probe:   sastToolRunsOnAllCommits.Probe,
					Outcome: finding.OutcomeTrue,
				},
			},
			result: scut.TestReturn{
				Score: checker.InconclusiveResultScore,
				Error: sce.ErrScorecardInternal,
			},
		},
		{
			name: "Sonar and codeQL is installed",
			findings: []finding.Finding{
				tool(checker.SonarWorkflow),
				tool(checker.CodeQLWorkflow),
				{
					Probe:   sastToolRunsOnAllCommits.Probe,
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						sastToolRunsOnAllCommits.AnalyzedPRsKey: "1",
						sastToolRunsOnAllCommits.TotalPRsKey:    "2",
					},
				},
			},
			result: scut.TestReturn{
				Score:        10,
				NumberOfInfo: 3,
				NumberOfWarn: 0,
			},
		},
		{
			name: "Pysa is installed. No other SAST tools are installed.",
			findings: []finding.Finding{
				tool(checker.PysaWorkflow),
				{
					Probe:   sastToolRunsOnAllCommits.Probe,
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						sastToolRunsOnAllCommits.AnalyzedPRsKey: "1",
						sastToolRunsOnAllCommits.TotalPRsKey:    "2",
					},
				},
			},
			result: scut.TestReturn{
				Score:        10,
				NumberOfInfo: 2,
				NumberOfWarn: 0,
			},
		},
		{
			name: `Sonar is installed. No other SAST tools are installed.
					Does not have info about whether SAST runs
					on every commit.`,
			findings: []finding.Finding{
				tool(checker.SonarWorkflow),
				{
					Probe:   sastToolRunsOnAllCommits.Probe,
					Outcome: finding.OutcomeNotApplicable,
				},
			},
			result: scut.TestReturn{
				Score:        10,
				NumberOfInfo: 1,
				NumberOfWarn: 1,
			},
		},
		{
			name: "No SAST tools are installed",
			findings: []finding.Finding{
				{
					Probe:   sastToolConfigured.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   sastToolRunsOnAllCommits.Probe,
					Outcome: finding.OutcomeFalse,
					Values: map[string]string{
						sastToolRunsOnAllCommits.AnalyzedPRsKey: "1",
						sastToolRunsOnAllCommits.TotalPRsKey:    "3",
					},
				},
			},
			result: scut.TestReturn{
				Score:        3,
				NumberOfWarn: 1,
				NumberOfInfo: 0,
			},
		},
		{
			name: "Snyk is installed, no other SAST tools are installed",
			findings: []finding.Finding{
				tool(checker.SnykWorkflow),
				{
					Probe:   sastToolRunsOnAllCommits.Probe,
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						sastToolRunsOnAllCommits.AnalyzedPRsKey: "1",
						sastToolRunsOnAllCommits.TotalPRsKey:    "3",
					},
				},
			},
			result: scut.TestReturn{
				Score:        10,
				NumberOfWarn: 0,
				NumberOfInfo: 2,
			},
		},
		{
			name: "Qodana is installed, no other SAST tools are installed",
			findings: []finding.Finding{
				tool(checker.QodanaWorkflow),
				{
					Probe:   sastToolRunsOnAllCommits.Probe,
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						sastToolRunsOnAllCommits.AnalyzedPRsKey: "1",
						sastToolRunsOnAllCommits.TotalPRsKey:    "3",
					},
				},
			},
			result: scut.TestReturn{
				Score:        10,
				NumberOfWarn: 0,
				NumberOfInfo: 2,
			},
		},
		{
			name: "Hadolint is installed, no other SAST tools are installed",
			findings: []finding.Finding{
				tool(checker.HadolintWorkflow),
				{
					Probe:   sastToolRunsOnAllCommits.Probe,
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						sastToolRunsOnAllCommits.AnalyzedPRsKey: "1",
						sastToolRunsOnAllCommits.TotalPRsKey:    "3",
					},
				},
			},
			result: scut.TestReturn{
				Score:        10,
				NumberOfWarn: 0,
				NumberOfInfo: 2,
			},
		},
		{
			name: "Claude Code Security is installed, no other SAST tools are installed",
			findings: []finding.Finding{
				tool(checker.ClaudeCodeSecurityWorkflow),
				{
					Probe:   sastToolRunsOnAllCommits.Probe,
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						sastToolRunsOnAllCommits.AnalyzedPRsKey: "1",
						sastToolRunsOnAllCommits.TotalPRsKey:    "3",
					},
				},
			},
			result: scut.TestReturn{
				Score:        10,
				NumberOfWarn: 0,
				NumberOfInfo: 2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			got := SAST(tt.name, tt.findings, &dl)
			scut.ValidateTestReturn(t, tt.name, &tt.result, &got, &dl)
		})
	}
}

func tool(name checker.SASTWorkflowType) finding.Finding {
	return finding.Finding{
		Probe:   sastToolConfigured.Probe,
		Outcome: finding.OutcomeTrue,
		Values: map[string]string{
			sastToolConfigured.ToolKey: string(name),
		},
	}
}
