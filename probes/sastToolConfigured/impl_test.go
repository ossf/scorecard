// Copyright 2024 OpenSSF Scorecard Authors
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

//nolint:stylecheck
package sastToolConfigured

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/internal/utils/test"
	"github.com/ossf/scorecard/v5/probes/internal/utils/uerror"
)

func Test_Run(t *testing.T) {
	t.Parallel()
	tests := []struct {
		err      error
		raw      *checker.RawResults
		name     string
		outcomes []finding.Outcome
	}{
		{
			name:     "no raw data",
			raw:      nil,
			err:      uerror.ErrNil,
			outcomes: nil,
		},
		{
			name: "no SAST tools detected",
			raw: &checker.RawResults{
				SASTResults: checker.SASTData{
					Workflows: []checker.SASTWorkflow{},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeFalse,
			},
		},
		{
			name: "multiple tools detected",
			err:  nil,
			raw: &checker.RawResults{
				SASTResults: checker.SASTData{
					Workflows: []checker.SASTWorkflow{
						{
							Type: checker.CodeQLWorkflow,
						},
						{
							Type: checker.QodanaWorkflow,
						},
						{
							Type: checker.PysaWorkflow,
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue,
				finding.OutcomeTrue,
				finding.OutcomeTrue,
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			findings, s, err := Run(tt.raw)
			if !cmp.Equal(tt.err, err, cmpopts.EquateErrors()) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(tt.err, err, cmpopts.EquateErrors()))
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(Probe, s); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
			test.AssertOutcomes(t, findings, tt.outcomes)
		})
	}
}

func Test_Run_tools(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		raw   *checker.RawResults
		tools []string
	}{
		{
			name: "one tool",
			raw: &checker.RawResults{
				SASTResults: checker.SASTData{
					Workflows: []checker.SASTWorkflow{
						{
							Type: checker.CodeQLWorkflow,
						},
					},
				},
			},
			tools: []string{"CodeQL"},
		},
		{
			name: "one tool, multiple times",
			raw: &checker.RawResults{
				SASTResults: checker.SASTData{
					Workflows: []checker.SASTWorkflow{
						{
							Type: checker.CodeQLWorkflow,
						},
						{
							Type: checker.CodeQLWorkflow,
						},
					},
				},
			},
			tools: []string{"CodeQL", "CodeQL"},
		},
		{
			name: "multiple tools",
			raw: &checker.RawResults{
				SASTResults: checker.SASTData{
					Workflows: []checker.SASTWorkflow{
						{
							Type: checker.SonarWorkflow,
						},
						{
							Type: checker.PysaWorkflow,
						},
					},
				},
			},
			tools: []string{"Sonar", "Pysa"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			findings, s, err := Run(tt.raw)
			if err != nil {
				t.Fatalf("expected no error: %v", err)
			}
			if diff := cmp.Diff(Probe, s); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
			assertTools(t, findings, tt.tools)
		})
	}
}

func assertTools(tb testing.TB, findings []finding.Finding, tools []string) {
	tb.Helper()
	if len(findings) != len(tools) {
		tb.Fatalf("mismatch between number of finding (%d) and tools (%d)", len(findings), len(tools))
	}
	for i, f := range findings {
		if f.Outcome != finding.OutcomeTrue {
			tb.Errorf("outcome (%v) shouldn't have a tool field", f.Outcome)
		}
		tool, ok := f.Values[ToolKey]
		if !ok {
			tb.Errorf("no tool present")
		}
		if tool != tools[i] {
			tb.Errorf("got: %s, wanted: %s", tool, tools[i])
		}
	}
}
