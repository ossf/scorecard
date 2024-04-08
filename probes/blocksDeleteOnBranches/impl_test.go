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
package blocksDeleteOnBranches

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/internal/utils/test"
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
			name: "One branch blocks branch deletion should result in one true outcome",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.BranchRef{
						{
							Name: &branchVal1,
							BranchProtectionRule: clients.BranchProtectionRule{
								AllowDeletions: &falseVal,
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
			name: "Two branches that block branch deletions should result in two true outcomes",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.BranchRef{
						{
							Name: &branchVal1,
							BranchProtectionRule: clients.BranchProtectionRule{
								AllowDeletions: &falseVal,
							},
						},
						{
							Name: &branchVal2,
							BranchProtectionRule: clients.BranchProtectionRule{
								AllowDeletions: &falseVal,
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
			name: "Two branches in total: One blocks branch deletion and one doesn't = 1 true & 1 false",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.BranchRef{
						{
							Name: &branchVal1,
							BranchProtectionRule: clients.BranchProtectionRule{
								AllowDeletions: &falseVal,
							},
						},
						{
							Name: &branchVal2,
							BranchProtectionRule: clients.BranchProtectionRule{
								AllowDeletions: &trueVal,
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
			name: "Two branches in total: One blocks branch deletion and one doesn't = 1 false & 1 true",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.BranchRef{
						{
							Name: &branchVal1,
							BranchProtectionRule: clients.BranchProtectionRule{
								AllowDeletions: &trueVal,
							},
						},
						{
							Name: &branchVal2,
							BranchProtectionRule: clients.BranchProtectionRule{
								AllowDeletions: &falseVal,
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
			name: "Two branches in total: One blocks branch deletion and one lacks data = 1 false & 1 unavailable",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.BranchRef{
						{
							Name: &branchVal1,
							BranchProtectionRule: clients.BranchProtectionRule{
								AllowDeletions: &trueVal,
							},
						},
						{
							Name: &branchVal2,
							BranchProtectionRule: clients.BranchProtectionRule{
								AllowDeletions: nil,
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
