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
package requiresCodeOwnersReview

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
			name: "1 branch requires code owner reviews with files = 1 positive outcome",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.BranchRef{
						{
							Name: &branchVal1,
							BranchProtectionRule: clients.BranchProtectionRule{
								RequiredPullRequestReviews: clients.PullRequestReviewRule{
									RequireCodeOwnerReviews: &trueVal,
								},
							},
						},
					},
					CodeownersFiles: []string{"file"},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomePositive,
			},
		},
		{
			name: "1 branch requires code owner reviews without files = 1 negative outcome",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.BranchRef{
						{
							Name: &branchVal1,
							BranchProtectionRule: clients.BranchProtectionRule{
								RequiredPullRequestReviews: clients.PullRequestReviewRule{
									RequireCodeOwnerReviews: &trueVal,
								},
							},
						},
					},
					CodeownersFiles: []string{},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeNegative,
			},
		},
		{
			name: "2 branches require code owner reviews with files = 2 positive outcomes",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.BranchRef{
						{
							Name: &branchVal1,
							BranchProtectionRule: clients.BranchProtectionRule{
								RequiredPullRequestReviews: clients.PullRequestReviewRule{
									RequireCodeOwnerReviews: &trueVal,
								},
							},
						},
						{
							Name: &branchVal2,
							BranchProtectionRule: clients.BranchProtectionRule{
								RequiredPullRequestReviews: clients.PullRequestReviewRule{
									RequireCodeOwnerReviews: &trueVal,
								},
							},
						},
					},
					CodeownersFiles: []string{"file"},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomePositive, finding.OutcomePositive,
			},
		},
		{
			name: "2 branches require code owner reviews with files = 2 negative outcomes",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.BranchRef{
						{
							Name: &branchVal1,
							BranchProtectionRule: clients.BranchProtectionRule{
								RequiredPullRequestReviews: clients.PullRequestReviewRule{
									RequireCodeOwnerReviews: &trueVal,
								},
							},
						},
						{
							Name: &branchVal2,
							BranchProtectionRule: clients.BranchProtectionRule{
								RequiredPullRequestReviews: clients.PullRequestReviewRule{
									RequireCodeOwnerReviews: &trueVal,
								},
							},
						},
					},
					CodeownersFiles: []string{},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeNegative, finding.OutcomeNegative,
			},
		},
		{
			name: "1 branches require code owner reviews and 1 branch doesn't with files = 1 positive 1 negative",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.BranchRef{
						{
							Name: &branchVal1,
							BranchProtectionRule: clients.BranchProtectionRule{
								RequiredPullRequestReviews: clients.PullRequestReviewRule{
									RequireCodeOwnerReviews: &trueVal,
								},
							},
						},
						{
							Name: &branchVal2,
							BranchProtectionRule: clients.BranchProtectionRule{
								RequiredPullRequestReviews: clients.PullRequestReviewRule{
									RequireCodeOwnerReviews: &falseVal,
								},
							},
						},
					},
					CodeownersFiles: []string{"file"},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomePositive, finding.OutcomeNegative,
			},
		},
		{
			name: "Requires code owner reviews on 1/2 branches - without files = 2 negative outcomes",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.BranchRef{
						{
							Name: &branchVal1,
							BranchProtectionRule: clients.BranchProtectionRule{
								RequiredPullRequestReviews: clients.PullRequestReviewRule{
									RequireCodeOwnerReviews: &trueVal,
								},
							},
						},
						{
							Name: &branchVal2,
							BranchProtectionRule: clients.BranchProtectionRule{
								RequiredPullRequestReviews: clients.PullRequestReviewRule{
									RequireCodeOwnerReviews: &falseVal,
								},
							},
						},
					},
					CodeownersFiles: []string{},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeNegative, finding.OutcomeNegative,
			},
		},
		{
			name: "Requires code owner reviews on 1/2 branches - with files = 1 negative and 1 positive",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.BranchRef{
						{
							Name: &branchVal1,
							BranchProtectionRule: clients.BranchProtectionRule{
								RequiredPullRequestReviews: clients.PullRequestReviewRule{
									RequireCodeOwnerReviews: &falseVal,
								},
							},
						},
						{
							Name: &branchVal2,
							BranchProtectionRule: clients.BranchProtectionRule{
								RequiredPullRequestReviews: clients.PullRequestReviewRule{
									RequireCodeOwnerReviews: &trueVal,
								},
							},
						},
					},
					CodeownersFiles: []string{"file"},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeNegative, finding.OutcomePositive,
			},
		},
		{
			name: "Requires code owner reviews on 1/2 branches - without files = 2 negative outcomes",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.BranchRef{
						{
							Name: &branchVal1,
							BranchProtectionRule: clients.BranchProtectionRule{
								RequiredPullRequestReviews: clients.PullRequestReviewRule{
									RequireCodeOwnerReviews: &falseVal,
								},
							},
						},
						{
							Name: &branchVal2,
							BranchProtectionRule: clients.BranchProtectionRule{
								RequiredPullRequestReviews: clients.PullRequestReviewRule{
									RequireCodeOwnerReviews: &trueVal,
								},
							},
						},
					},
					CodeownersFiles: []string{},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeNegative, finding.OutcomeNegative,
			},
		},
		{
			name: "1 branch does not require code owner review and 1 lacks data = 1 negative and 1 unavailable",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.BranchRef{
						{
							Name: &branchVal1,
							BranchProtectionRule: clients.BranchProtectionRule{
								RequiredPullRequestReviews: clients.PullRequestReviewRule{
									RequireCodeOwnerReviews: &falseVal,
								},
							},
						},
						{
							Name: &branchVal2,
							BranchProtectionRule: clients.BranchProtectionRule{
								RequiredPullRequestReviews: clients.PullRequestReviewRule{
									RequireCodeOwnerReviews: nil,
								},
							},
						},
					},
					CodeownersFiles: []string{"file"},
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
