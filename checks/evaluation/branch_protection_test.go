// Copyright 2020 OpenSSF Scorecard Authors
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

package evaluation

import (
	"testing"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	scut "github.com/ossf/scorecard/v4/utests"
)

func testScore(branch *clients.BranchRef, codeownersFiles []string, dl checker.DetailLogger) (int, error) {
	var score levelScore
	score.scores.basic, score.maxes.basic = basicNonAdminProtection(branch, dl)
	score.scores.review, score.maxes.review = nonAdminReviewProtection(branch)
	score.scores.adminReview, score.maxes.adminReview = adminReviewProtection(branch, dl)
	score.scores.context, score.maxes.context = nonAdminContextProtection(branch, dl)
	score.scores.thoroughReview, score.maxes.thoroughReview = nonAdminThoroughReviewProtection(branch, dl)
	score.scores.adminThoroughReview, score.maxes.adminThoroughReview = adminThoroughReviewProtection(branch, dl)
	score.scores.codeownerReview, score.maxes.codeownerReview = codeownerBranchProtection(branch, codeownersFiles, dl)

	return computeScore([]levelScore{score})
}

func TestIsBranchProtected(t *testing.T) {
	t.Parallel()
	trueVal := true
	falseVal := false
	var zeroVal int32
	var oneVal int32 = 1
	branchVal := "branch-name"
	tests := []struct {
		name            string
		branch          *clients.BranchRef
		codeownersFiles []string
		expected        scut.TestReturn
	}{
		{
			name: "Nothing is enabled",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         3,
				NumberOfWarn:  7,
				NumberOfInfo:  2,
				NumberOfDebug: 0,
			},
			branch: &clients.BranchRef{
				Name:      &branchVal,
				Protected: &trueVal,
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions:          &falseVal,
					AllowForcePushes:        &falseVal,
					RequireLinearHistory:    &falseVal,
					EnforceAdmins:           &falseVal,
					RequireLastPushApproval: &falseVal,
					RequiredPullRequestReviews: clients.PullRequestReviewRule{
						DismissStaleReviews:          &falseVal,
						RequireCodeOwnerReviews:      &falseVal,
						RequiredApprovingReviewCount: &zeroVal,
					},
					CheckRules: clients.StatusChecksRule{
						RequiresStatusChecks: &trueVal,
						Contexts:             nil,
						UpToDateBeforeMerge:  &falseVal,
					},
				},
			},
		},
		{
			name: "Nothing is enabled and values are nil",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         0,
				NumberOfWarn:  2,
				NumberOfInfo:  0,
				NumberOfDebug: 4,
			},
			branch: &clients.BranchRef{
				Name:      &branchVal,
				Protected: &trueVal,
			},
		},
		{
			name: "Required status check enabled",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         4,
				NumberOfWarn:  5,
				NumberOfInfo:  4,
				NumberOfDebug: 0,
			},
			branch: &clients.BranchRef{
				Name:      &branchVal,
				Protected: &trueVal,
				BranchProtectionRule: clients.BranchProtectionRule{
					RequiredPullRequestReviews: clients.PullRequestReviewRule{
						DismissStaleReviews:          &falseVal,
						RequireCodeOwnerReviews:      &falseVal,
						RequiredApprovingReviewCount: &zeroVal,
					},
					CheckRules: clients.StatusChecksRule{
						RequiresStatusChecks: &trueVal,
						UpToDateBeforeMerge:  &trueVal,
						Contexts:             []string{"foo"},
					},
					EnforceAdmins:           &falseVal,
					RequireLastPushApproval: &falseVal,
					RequireLinearHistory:    &falseVal,
					AllowForcePushes:        &falseVal,
					AllowDeletions:          &falseVal,
				},
			},
		},
		{
			name: "Required status check enabled without checking for status string",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         4,
				NumberOfWarn:  6,
				NumberOfInfo:  3,
				NumberOfDebug: 0,
			},
			branch: &clients.BranchRef{
				Name:      &branchVal,
				Protected: &trueVal,
				BranchProtectionRule: clients.BranchProtectionRule{
					EnforceAdmins:           &falseVal,
					RequireLastPushApproval: &falseVal,
					RequireLinearHistory:    &falseVal,
					AllowForcePushes:        &falseVal,
					AllowDeletions:          &falseVal,
					RequiredPullRequestReviews: clients.PullRequestReviewRule{
						DismissStaleReviews:          &falseVal,
						RequireCodeOwnerReviews:      &falseVal,
						RequiredApprovingReviewCount: &zeroVal,
					},
					CheckRules: clients.StatusChecksRule{
						RequiresStatusChecks: &trueVal,
						UpToDateBeforeMerge:  &trueVal,
						Contexts:             nil,
					},
				},
			},
		},
		{
			name: "Required pull request enabled",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         4,
				NumberOfWarn:  6,
				NumberOfInfo:  3,
				NumberOfDebug: 0,
			},
			branch: &clients.BranchRef{
				Name:      &branchVal,
				Protected: &trueVal,
				BranchProtectionRule: clients.BranchProtectionRule{
					EnforceAdmins:           &falseVal,
					RequireLastPushApproval: &falseVal,
					RequireLinearHistory:    &trueVal,
					AllowForcePushes:        &falseVal,
					AllowDeletions:          &falseVal,
					CheckRules: clients.StatusChecksRule{
						RequiresStatusChecks: &trueVal,
						UpToDateBeforeMerge:  &falseVal,
						Contexts:             []string{"foo"},
					},
					RequiredPullRequestReviews: clients.PullRequestReviewRule{
						DismissStaleReviews:          &falseVal,
						RequireCodeOwnerReviews:      &falseVal,
						RequiredApprovingReviewCount: &oneVal,
					},
				},
			},
		},
		{
			name: "Required admin enforcement enabled",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         3,
				NumberOfWarn:  5,
				NumberOfInfo:  4,
				NumberOfDebug: 0,
			},
			branch: &clients.BranchRef{
				Name:      &branchVal,
				Protected: &trueVal,
				BranchProtectionRule: clients.BranchProtectionRule{
					EnforceAdmins:           &trueVal,
					RequireLastPushApproval: &falseVal,
					RequireLinearHistory:    &trueVal,
					AllowForcePushes:        &falseVal,
					AllowDeletions:          &falseVal,
					CheckRules: clients.StatusChecksRule{
						RequiresStatusChecks: &falseVal,
						UpToDateBeforeMerge:  &falseVal,
						Contexts:             []string{"foo"},
					},
					RequiredPullRequestReviews: clients.PullRequestReviewRule{
						DismissStaleReviews:          &falseVal,
						RequireCodeOwnerReviews:      &falseVal,
						RequiredApprovingReviewCount: &zeroVal,
					},
				},
			},
		},
		{
			name: "Required linear history enabled",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         3,
				NumberOfWarn:  6,
				NumberOfInfo:  3,
				NumberOfDebug: 0,
			},
			branch: &clients.BranchRef{
				Name:      &branchVal,
				Protected: &trueVal,
				BranchProtectionRule: clients.BranchProtectionRule{
					EnforceAdmins:           &falseVal,
					RequireLastPushApproval: &falseVal,
					RequireLinearHistory:    &trueVal,
					AllowForcePushes:        &falseVal,
					AllowDeletions:          &falseVal,
					CheckRules: clients.StatusChecksRule{
						RequiresStatusChecks: &falseVal,
						UpToDateBeforeMerge:  &falseVal,
						Contexts:             []string{"foo"},
					},
					RequiredPullRequestReviews: clients.PullRequestReviewRule{
						DismissStaleReviews:          &falseVal,
						RequireCodeOwnerReviews:      &falseVal,
						RequiredApprovingReviewCount: &zeroVal,
					},
				},
			},
		},
		{
			name: "Allow force push enabled",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         1,
				NumberOfWarn:  7,
				NumberOfInfo:  2,
				NumberOfDebug: 0,
			},
			branch: &clients.BranchRef{
				Name:      &branchVal,
				Protected: &trueVal,
				BranchProtectionRule: clients.BranchProtectionRule{
					EnforceAdmins:           &falseVal,
					RequireLastPushApproval: &falseVal,
					RequireLinearHistory:    &falseVal,
					AllowForcePushes:        &trueVal,
					AllowDeletions:          &falseVal,

					CheckRules: clients.StatusChecksRule{
						RequiresStatusChecks: &falseVal,
						UpToDateBeforeMerge:  &falseVal,
						Contexts:             []string{"foo"},
					},
					RequiredPullRequestReviews: clients.PullRequestReviewRule{
						DismissStaleReviews:          &falseVal,
						RequireCodeOwnerReviews:      &falseVal,
						RequiredApprovingReviewCount: &zeroVal,
					},
				},
			},
		},
		{
			name: "Allow deletions enabled",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         1,
				NumberOfWarn:  7,
				NumberOfInfo:  2,
				NumberOfDebug: 0,
			},
			branch: &clients.BranchRef{
				Name:      &branchVal,
				Protected: &trueVal,
				BranchProtectionRule: clients.BranchProtectionRule{
					EnforceAdmins:           &falseVal,
					RequireLastPushApproval: &falseVal,
					RequireLinearHistory:    &falseVal,
					AllowForcePushes:        &falseVal,
					AllowDeletions:          &trueVal,
					CheckRules: clients.StatusChecksRule{
						RequiresStatusChecks: &falseVal,
						UpToDateBeforeMerge:  &falseVal,
						Contexts:             []string{"foo"},
					},
					RequiredPullRequestReviews: clients.PullRequestReviewRule{
						DismissStaleReviews:          &falseVal,
						RequireCodeOwnerReviews:      &falseVal,
						RequiredApprovingReviewCount: &zeroVal,
					},
				},
			},
		},
		{
			name: "Branches are protected",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         8,
				NumberOfWarn:  2,
				NumberOfInfo:  8,
				NumberOfDebug: 0,
			},
			branch: &clients.BranchRef{
				Name:      &branchVal,
				Protected: &trueVal,
				BranchProtectionRule: clients.BranchProtectionRule{
					EnforceAdmins:           &trueVal,
					RequireLinearHistory:    &trueVal,
					RequireLastPushApproval: &trueVal,
					AllowForcePushes:        &falseVal,
					AllowDeletions:          &falseVal,
					CheckRules: clients.StatusChecksRule{
						RequiresStatusChecks: &falseVal,
						UpToDateBeforeMerge:  &trueVal,
						Contexts:             []string{"foo"},
					},
					RequiredPullRequestReviews: clients.PullRequestReviewRule{
						DismissStaleReviews:          &trueVal,
						RequireCodeOwnerReviews:      &trueVal,
						RequiredApprovingReviewCount: &oneVal,
					},
				},
			},
		},
		{
			name: "Branches are protected and require codeowner review",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         8,
				NumberOfWarn:  1,
				NumberOfInfo:  8,
				NumberOfDebug: 0,
			},
			branch: &clients.BranchRef{
				Name:      &branchVal,
				Protected: &trueVal,
				BranchProtectionRule: clients.BranchProtectionRule{
					EnforceAdmins:           &trueVal,
					RequireLinearHistory:    &trueVal,
					RequireLastPushApproval: &trueVal,
					AllowForcePushes:        &falseVal,
					AllowDeletions:          &falseVal,
					CheckRules: clients.StatusChecksRule{
						RequiresStatusChecks: &trueVal,
						UpToDateBeforeMerge:  &trueVal,
						Contexts:             []string{"foo"},
					},
					RequiredPullRequestReviews: clients.PullRequestReviewRule{
						DismissStaleReviews:          &trueVal,
						RequireCodeOwnerReviews:      &trueVal,
						RequiredApprovingReviewCount: &oneVal,
					},
				},
			},
			codeownersFiles: []string{".github/CODEOWNERS"},
		},
		{
			name: "Branches are protected and require codeowner review, but file is not present",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         5,
				NumberOfWarn:  3,
				NumberOfInfo:  7,
				NumberOfDebug: 0,
			},
			branch: &clients.BranchRef{
				Name:      &branchVal,
				Protected: &trueVal,
				BranchProtectionRule: clients.BranchProtectionRule{
					EnforceAdmins:           &trueVal,
					RequireLastPushApproval: &falseVal,
					RequireLinearHistory:    &trueVal,
					AllowForcePushes:        &falseVal,
					AllowDeletions:          &falseVal,
					CheckRules: clients.StatusChecksRule{
						RequiresStatusChecks: &falseVal,
						UpToDateBeforeMerge:  &trueVal,
						Contexts:             []string{"foo"},
					},
					RequiredPullRequestReviews: clients.PullRequestReviewRule{
						DismissStaleReviews:          &trueVal,
						RequireCodeOwnerReviews:      &trueVal,
						RequiredApprovingReviewCount: &oneVal,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			score, err := testScore(tt.branch, tt.codeownersFiles, &dl)
			actual := &checker.CheckResult{
				Score: score,
				Error: err,
			}
			if !scut.ValidateTestReturn(t, tt.name, &tt.expected, actual, &dl) {
				t.Fail()
			}
		})
	}
}
