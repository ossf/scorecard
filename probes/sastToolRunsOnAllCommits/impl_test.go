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

//nolint:stylecheck
package sastToolRunsOnAllCommits

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/internal/utils/test"
)

func Test_Run(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		raw              *checker.RawResults
		outcomes         []finding.Outcome
		err              error
		expectedFindings []finding.Finding
	}{
		{
			name: "sonar present",
			err:  nil,
			raw: &checker.RawResults{
				SASTResults: checker.SASTData{
					Commits: []checker.SASTCommit{
						{
							Compliant: false,
						},
						{
							Compliant: true,
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeNegative,
			},
			expectedFindings: []finding.Finding{
				{
					Probe:   Probe,
					Message: "1 commits out of 2 are checked with a SAST tool",
					Values: map[string]string{
						AnalyzedPRsKey: "1",
						TotalPRsKey:    "2",
					},
				},
			},
		},
		{
			name: "sonar present",
			err:  nil,
			raw: &checker.RawResults{
				SASTResults: checker.SASTData{
					Commits: []checker.SASTCommit{
						{
							Compliant: true,
						},
						{
							Compliant: true,
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomePositive,
			},
			expectedFindings: []finding.Finding{
				{
					Probe:   Probe,
					Message: "all commits (2) are checked with a SAST tool",
					Outcome: finding.OutcomePositive,
					Values: map[string]string{
						AnalyzedPRsKey: "2",
						TotalPRsKey:    "2",
					},
				},
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
			if !cmp.Equal(tt.expectedFindings, findings, cmpopts.EquateErrors()) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(tt.expectedFindings, findings, cmpopts.EquateErrors()))
			}
		})
	}
}
