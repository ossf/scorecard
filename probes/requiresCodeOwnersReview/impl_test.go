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
			name: "1 branch requires code owner reviews with files = 1 true outcome",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.BranchRef{
						{
							Name: &branchVal1,
							BranchProtectionRule: clients.BranchProtectionRule{
								PullRequestRule: clients.PullRequestRule{
									RequireCodeOwnerReviews: &trueVal,
								},
							},
						},
					},
					CodeownersFiles: []string{"file"},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue,
			},
		},
		{
			name: "1 branch requires code owner reviews without files = 1 false outcome",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.BranchRef{
						{
							Name: &branchVal1,
							BranchProtectionRule: clients.BranchProtectionRule{
								PullRequestRule: clients.PullRequestRule{
									RequireCodeOwnerReviews: &trueVal,
								},
							},
						},
					},
					CodeownersFiles: []string{},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeFalse,
			},
		},
		{
			name: "2 branches require code owner reviews with files = 2 true outcomes",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.BranchRef{
						{
							Name: &branchVal1,
							BranchProtectionRule: clients.BranchProtectionRule{
								PullRequestRule: clients.PullRequestRule{
									RequireCodeOwnerReviews: &trueVal,
								},
							},
						},
						{
							Name: &branchVal2,
							BranchProtectionRule: clients.BranchProtectionRule{
								PullRequestRule: clients.PullRequestRule{
									RequireCodeOwnerReviews: &trueVal,
								},
							},
						},
					},
					CodeownersFiles: []string{"file1"},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue, finding.OutcomeTrue,
			},
		},
		{
			name: "2 branches require code owner reviews with files = 2 true outcomes",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.BranchRef{
						{
							Name: &branchVal1,
							BranchProtectionRule: clients.BranchProtectionRule{
								PullRequestRule: clients.PullRequestRule{
									RequireCodeOwnerReviews: &trueVal,
								},
							},
						},
						{
							Name: &branchVal2,
							BranchProtectionRule: clients.BranchProtectionRule{
								PullRequestRule: clients.PullRequestRule{
									RequireCodeOwnerReviews: &trueVal,
								},
							},
						},
					},
					CodeownersFiles: []string{"file1", "file2"},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue, finding.OutcomeTrue,
			},
		},
		{
			name: "1 branches require code owner reviews and 1 branch doesn't with files = 1 true 1 false",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.BranchRef{
						{
							Name: &branchVal1,
							BranchProtectionRule: clients.BranchProtectionRule{
								PullRequestRule: clients.PullRequestRule{
									RequireCodeOwnerReviews: &trueVal,
								},
							},
						},
						{
							Name: &branchVal2,
							BranchProtectionRule: clients.BranchProtectionRule{
								PullRequestRule: clients.PullRequestRule{
									RequireCodeOwnerReviews: &falseVal,
								},
							},
						},
					},
					CodeownersFiles: []string{"file"},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue, finding.OutcomeFalse,
			},
		},
		{
			name: "Requires code owner reviews on 1/2 branches - without files = 1 true and 1 false",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.BranchRef{
						{
							Name: &branchVal1,
							BranchProtectionRule: clients.BranchProtectionRule{
								PullRequestRule: clients.PullRequestRule{
									RequireCodeOwnerReviews: &trueVal,
								},
							},
						},
						{
							Name: &branchVal2,
							BranchProtectionRule: clients.BranchProtectionRule{
								PullRequestRule: clients.PullRequestRule{
									RequireCodeOwnerReviews: &falseVal,
								},
							},
						},
					},
					CodeownersFiles: []string{"file"},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue, finding.OutcomeFalse,
			},
		},
		{
			name: "Requires code owner reviews on 1/2 branches - with files = 1 false and 1 true",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.BranchRef{
						{
							Name: &branchVal1,
							BranchProtectionRule: clients.BranchProtectionRule{
								PullRequestRule: clients.PullRequestRule{
									RequireCodeOwnerReviews: &falseVal,
								},
							},
						},
						{
							Name: &branchVal2,
							BranchProtectionRule: clients.BranchProtectionRule{
								PullRequestRule: clients.PullRequestRule{
									RequireCodeOwnerReviews: &trueVal,
								},
							},
						},
					},
					CodeownersFiles: []string{"file"},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeFalse, finding.OutcomeTrue,
			},
		},
		{
			name: "Requires code owner reviews on 1/2 branches - without files = 2 false",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.BranchRef{
						{
							Name: &branchVal1,
							BranchProtectionRule: clients.BranchProtectionRule{
								PullRequestRule: clients.PullRequestRule{
									RequireCodeOwnerReviews: &falseVal,
								},
							},
						},
						{
							Name: &branchVal2,
							BranchProtectionRule: clients.BranchProtectionRule{
								PullRequestRule: clients.PullRequestRule{
									RequireCodeOwnerReviews: &trueVal,
								},
							},
						},
					},
					CodeownersFiles: []string{},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeFalse, finding.OutcomeFalse,
			},
		},
		{
			name: "1 branch does not require code owner review and 1 lacks data = 1 false and 1 unavailable",
			raw: &checker.RawResults{
				BranchProtectionResults: checker.BranchProtectionsData{
					Branches: []clients.BranchRef{
						{
							Name: &branchVal1,
							BranchProtectionRule: clients.BranchProtectionRule{
								PullRequestRule: clients.PullRequestRule{
									RequireCodeOwnerReviews: &falseVal,
								},
							},
						},
						{
							Name: &branchVal2,
							BranchProtectionRule: clients.BranchProtectionRule{
								PullRequestRule: clients.PullRequestRule{
									RequireCodeOwnerReviews: nil,
								},
							},
						},
					},
					CodeownersFiles: []string{"file"},
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
