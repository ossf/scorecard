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
package requiresLastPushApproval

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/internal/utils/test"
)

func Test_Run(t *testing.T) {
	t.Parallel()
	trueVal := true
	falseVal := false
	branchVal1 := "branch-name1"
	branchVal2 := "branch-name1"
	//nolint:govet
	tests := []struct {
		name     string
		raw      *checker.RawResults
		outcomes []finding.Outcome
		err      error
	}{
		{
			name: "1 branch requires last push approval = 1 true outcome",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.RepoRef{
						{
							Name: &branchVal1,
							ProtectionRule: clients.ProtectionRule{
								RequireLastPushApproval: &trueVal,
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue,
			},
		},
		{
			name: "2 branches requires last push approval = 2 true outcomes",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.RepoRef{
						{
							Name: &branchVal1,
							ProtectionRule: clients.ProtectionRule{
								RequireLastPushApproval: &trueVal,
							},
						},
						{
							Name: &branchVal2,
							ProtectionRule: clients.ProtectionRule{
								RequireLastPushApproval: &trueVal,
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue, finding.OutcomeTrue,
			},
		},
		{
			name: "Last push approval enabled on 1/2 branches = 1 true and 1 false outcomes",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.RepoRef{
						{
							Name: &branchVal1,
							ProtectionRule: clients.ProtectionRule{
								RequireLastPushApproval: &trueVal,
							},
						},
						{
							Name: &branchVal2,
							ProtectionRule: clients.ProtectionRule{
								RequireLastPushApproval: &falseVal,
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue, finding.OutcomeFalse,
			},
		},
		{
			name: "Last push approval enabled on 1/2 branches = 1 false and 1 true outcomes",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.RepoRef{
						{
							Name: &branchVal1,
							ProtectionRule: clients.ProtectionRule{
								RequireLastPushApproval: &falseVal,
							},
						},
						{
							Name: &branchVal2,
							ProtectionRule: clients.ProtectionRule{
								RequireLastPushApproval: &trueVal,
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeFalse, finding.OutcomeTrue,
			},
		},
		{
			name: "1 branch does not require last push approval and 1 lacks data = 1 false and 1 unavailable",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.RepoRef{
						{
							Name: &branchVal1,
							ProtectionRule: clients.ProtectionRule{
								RequireLastPushApproval: &falseVal,
							},
						},
						{
							Name: &branchVal2,
							ProtectionRule: clients.ProtectionRule{
								RequireLastPushApproval: nil,
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeFalse, finding.OutcomeNotAvailable,
			},
		},
	}
	for _, tt := range tests {
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
