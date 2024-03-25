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
package requiresApproversForPullRequests

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
	var zeroVal int32
	var oneVal int32 = 1
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
			name: "1 branch requires 1 reviewer = 1 positive outcome = 1 positive outcome",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.BranchRef{
						{
							Name: &branchVal1,
							BranchProtectionRule: clients.BranchProtectionRule{
								PullRequestRule: clients.PullRequestRule{
									RequiredApprovingReviewCount: &oneVal,
								},
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomePositive,
			},
		},
		{
			name: "2 branch require 1 reviewer each = 2 positive outcomes = 2 positive outcomes",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.BranchRef{
						{
							Name: &branchVal1,
							BranchProtectionRule: clients.BranchProtectionRule{
								PullRequestRule: clients.PullRequestRule{
									RequiredApprovingReviewCount: &oneVal,
								},
							},
						},
						{
							Name: &branchVal2,
							BranchProtectionRule: clients.BranchProtectionRule{
								PullRequestRule: clients.PullRequestRule{
									RequiredApprovingReviewCount: &oneVal,
								},
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomePositive, finding.OutcomePositive,
			},
		},
		{
			name: "1 branch requires 1 reviewer and 1 branch requires 0 reviewers = 1 positive and 1 negative",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.BranchRef{
						{
							Name: &branchVal1,
							BranchProtectionRule: clients.BranchProtectionRule{
								PullRequestRule: clients.PullRequestRule{
									RequiredApprovingReviewCount: &oneVal,
								},
							},
						},
						{
							Name: &branchVal2,
							BranchProtectionRule: clients.BranchProtectionRule{
								PullRequestRule: clients.PullRequestRule{
									RequiredApprovingReviewCount: &zeroVal,
								},
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomePositive, finding.OutcomeNegative,
			},
		},
		{
			name: "1 branch requires 0 reviewers and 1 branch requires 1 reviewer = 1 negative and 1 positive",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.BranchRef{
						{
							Name: &branchVal1,
							BranchProtectionRule: clients.BranchProtectionRule{
								PullRequestRule: clients.PullRequestRule{
									RequiredApprovingReviewCount: &zeroVal,
								},
							},
						},
						{
							Name: &branchVal2,
							BranchProtectionRule: clients.BranchProtectionRule{
								PullRequestRule: clients.PullRequestRule{
									RequiredApprovingReviewCount: &oneVal,
								},
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeNegative, finding.OutcomePositive,
			},
		},
		{
			name: "1 branch requires 0 reviewers and 1 branch lacks data = 1 negative and 1 unavailable",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.BranchRef{
						{
							Name: &branchVal1,
							BranchProtectionRule: clients.BranchProtectionRule{
								PullRequestRule: clients.PullRequestRule{
									RequiredApprovingReviewCount: &zeroVal,
								},
							},
						},
						{
							Name: &branchVal2,
							BranchProtectionRule: clients.BranchProtectionRule{
								PullRequestRule: clients.PullRequestRule{
									RequiredApprovingReviewCount: nil,
								},
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeNegative, finding.OutcomeNotAvailable,
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
