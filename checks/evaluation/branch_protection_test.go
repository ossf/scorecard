// Copyright 2020 Security Scorecard Authors
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
	scut "github.com/ossf/scorecard/v4/utests"
)

func testScore(branch *checker.BranchProtectionData, dl checker.DetailLogger) (int, error) {
	var score levelScore
	score.scores.basic, score.maxes.basic = basicNonAdminProtection(branch, dl)
	score.scores.adminBasic, score.maxes.adminBasic = basicAdminProtection(branch, dl)
	score.scores.review, score.maxes.review = nonAdminReviewProtection(branch)
	score.scores.adminReview, score.maxes.adminReview = adminReviewProtection(branch, dl)
	score.scores.context, score.maxes.context = nonAdminContextProtection(branch, dl)
	score.scores.thoroughReview, score.maxes.thoroughReview = nonAdminThoroughReviewProtection(branch, dl)
	score.scores.adminThoroughReview, score.maxes.adminThoroughReview = adminThoroughReviewProtection(branch, dl)

	return computeScore([]levelScore{score})
}

func TestIsBranchProtected(t *testing.T) {
	t.Parallel()
	trueVal := true
	falseVal := false
	var zeroVal int
	oneVal := 1

	tests := []struct {
		name     string
		branch   *checker.BranchProtectionData
		expected scut.TestReturn
	}{
		{
			name: "Nothing is enabled",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         2,
				NumberOfWarn:  5,
				NumberOfInfo:  2,
				NumberOfDebug: 0,
			},
			branch: &checker.BranchProtectionData{
				Name:                                "branch-name",
				Protected:                           &trueVal,
				RequiresStatusChecks:                &trueVal,
				RequiresUpToDateBranchBeforeMerging: &falseVal,
				StatusCheckContexts:                 nil,
				DismissesStaleReviews:               &falseVal,
				RequiresCodeOwnerReviews:            &falseVal,
				RequiredApprovingReviewCount:        &zeroVal,
				EnforcesAdmins:                      &falseVal,
				RequiresLinearHistory:               &falseVal,
				AllowsForcePushes:                   &falseVal,
				AllowsDeletions:                     &falseVal,
			},
		},
		{
			name: "Nothing is enabled and values are nil",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         0,
				NumberOfWarn:  2,
				NumberOfInfo:  0,
				NumberOfDebug: 3,
			},
			branch: &checker.BranchProtectionData{
				Name:      "branch-name",
				Protected: &trueVal,
			},
		},
		{
			name: "Required status check enabled",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         2,
				NumberOfWarn:  3,
				NumberOfInfo:  4,
				NumberOfDebug: 0,
			},
			branch: &checker.BranchProtectionData{
				Name:                                "branch-name",
				Protected:                           &trueVal,
				RequiresStatusChecks:                &trueVal,
				RequiresUpToDateBranchBeforeMerging: &trueVal,
				StatusCheckContexts:                 []string{"foo"},
				DismissesStaleReviews:               &falseVal,
				RequiresCodeOwnerReviews:            &falseVal,
				RequiredApprovingReviewCount:        &zeroVal,
				EnforcesAdmins:                      &falseVal,
				RequiresLinearHistory:               &falseVal,
				AllowsForcePushes:                   &falseVal,
				AllowsDeletions:                     &falseVal,
			},
		},
		{
			name: "Required status check enabled without checking for status string",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         2,
				NumberOfWarn:  4,
				NumberOfInfo:  3,
				NumberOfDebug: 0,
			},
			branch: &checker.BranchProtectionData{
				Name:                                "branch-name",
				Protected:                           &trueVal,
				RequiresStatusChecks:                &trueVal,
				RequiresUpToDateBranchBeforeMerging: &trueVal,
				StatusCheckContexts:                 nil,
				DismissesStaleReviews:               &falseVal,
				RequiresCodeOwnerReviews:            &falseVal,
				RequiredApprovingReviewCount:        &zeroVal,
				EnforcesAdmins:                      &falseVal,
				RequiresLinearHistory:               &falseVal,
				AllowsForcePushes:                   &falseVal,
				AllowsDeletions:                     &falseVal,
			},
		},
		{
			name: "Required pull request enabled",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         2,
				NumberOfWarn:  4,
				NumberOfInfo:  3,
				NumberOfDebug: 0,
			},
			branch: &checker.BranchProtectionData{
				Name:                                "branch-name",
				Protected:                           &trueVal,
				RequiresStatusChecks:                &trueVal,
				RequiresUpToDateBranchBeforeMerging: &falseVal,
				StatusCheckContexts:                 []string{"foo"},
				DismissesStaleReviews:               &falseVal,
				RequiresCodeOwnerReviews:            &falseVal,
				RequiredApprovingReviewCount:        &oneVal,
				EnforcesAdmins:                      &falseVal,
				RequiresLinearHistory:               &trueVal,
				AllowsForcePushes:                   &falseVal,
				AllowsDeletions:                     &falseVal,
			},
		},
		{
			name: "Required admin enforcement enabled",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         3,
				NumberOfWarn:  3,
				NumberOfInfo:  4,
				NumberOfDebug: 0,
			},
			branch: &checker.BranchProtectionData{
				Name:                                "branch-name",
				Protected:                           &trueVal,
				RequiresStatusChecks:                &falseVal,
				RequiresUpToDateBranchBeforeMerging: &falseVal,
				StatusCheckContexts:                 []string{"foo"},
				DismissesStaleReviews:               &falseVal,
				RequiresCodeOwnerReviews:            &falseVal,
				RequiredApprovingReviewCount:        &zeroVal,
				EnforcesAdmins:                      &trueVal,
				RequiresLinearHistory:               &trueVal,
				AllowsForcePushes:                   &falseVal,
				AllowsDeletions:                     &falseVal,
			},
		},
		{
			name: "Required linear history enabled",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         2,
				NumberOfWarn:  4,
				NumberOfInfo:  3,
				NumberOfDebug: 0,
			},
			branch: &checker.BranchProtectionData{
				Name:                                "branch-name",
				Protected:                           &trueVal,
				RequiresStatusChecks:                &falseVal,
				RequiresUpToDateBranchBeforeMerging: &falseVal,
				StatusCheckContexts:                 []string{"foo"},
				DismissesStaleReviews:               &falseVal,
				RequiresCodeOwnerReviews:            &falseVal,
				RequiredApprovingReviewCount:        &zeroVal,
				EnforcesAdmins:                      &falseVal,
				RequiresLinearHistory:               &trueVal,
				AllowsForcePushes:                   &falseVal,
				AllowsDeletions:                     &falseVal,
			},
		},
		{
			name: "Allow force push enabled",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         1,
				NumberOfWarn:  5,
				NumberOfInfo:  2,
				NumberOfDebug: 0,
			},
			branch: &checker.BranchProtectionData{
				Name:                                "branch-name",
				Protected:                           &trueVal,
				RequiresStatusChecks:                &falseVal,
				RequiresUpToDateBranchBeforeMerging: &falseVal,
				StatusCheckContexts:                 []string{"foo"},
				DismissesStaleReviews:               &falseVal,
				RequiresCodeOwnerReviews:            &falseVal,
				RequiredApprovingReviewCount:        &zeroVal,
				EnforcesAdmins:                      &falseVal,
				RequiresLinearHistory:               &falseVal,
				AllowsForcePushes:                   &trueVal,
				AllowsDeletions:                     &falseVal,
			},
		},
		{
			name: "Allow deletions enabled",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         1,
				NumberOfWarn:  5,
				NumberOfInfo:  2,
				NumberOfDebug: 0,
			},
			branch: &checker.BranchProtectionData{
				Name:                                "branch-name",
				Protected:                           &trueVal,
				RequiresStatusChecks:                &falseVal,
				RequiresUpToDateBranchBeforeMerging: &falseVal,
				StatusCheckContexts:                 []string{"foo"},
				DismissesStaleReviews:               &falseVal,
				RequiresCodeOwnerReviews:            &falseVal,
				RequiredApprovingReviewCount:        &zeroVal,
				EnforcesAdmins:                      &falseVal,
				RequiresLinearHistory:               &falseVal,
				AllowsForcePushes:                   &falseVal,
				AllowsDeletions:                     &trueVal,
			},
		},
		{
			name: "Branches are protected",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         8,
				NumberOfWarn:  1,
				NumberOfInfo:  6,
				NumberOfDebug: 0,
			},
			branch: &checker.BranchProtectionData{
				Name:                                "branch-name",
				Protected:                           &trueVal,
				RequiresStatusChecks:                &falseVal,
				RequiresUpToDateBranchBeforeMerging: &trueVal,
				StatusCheckContexts:                 []string{"foo"},
				DismissesStaleReviews:               &trueVal,
				RequiresCodeOwnerReviews:            &trueVal,
				RequiredApprovingReviewCount:        &oneVal,
				EnforcesAdmins:                      &trueVal,
				RequiresLinearHistory:               &trueVal,
				AllowsForcePushes:                   &falseVal,
				AllowsDeletions:                     &falseVal,
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			score, err := testScore(tt.branch, &dl)
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
