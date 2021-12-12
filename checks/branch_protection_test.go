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

package checks

import (
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/ossf/scorecard/v3/checker"
	"github.com/ossf/scorecard/v3/clients"
	mockrepo "github.com/ossf/scorecard/v3/clients/mockclients"
	sce "github.com/ossf/scorecard/v3/errors"
	scut "github.com/ossf/scorecard/v3/utests"
)

func getBranch(branches []*clients.BranchRef, name string) *clients.BranchRef {
	for _, branch := range branches {
		branchName := getBranchName(branch)
		if branchName == name {
			return branch
		}
	}
	return nil
}

func scrubBranch(branch *clients.BranchRef) *clients.BranchRef {
	ret := branch
	ret.BranchProtectionRule = clients.BranchProtectionRule{}
	return ret
}

func scrubBranches(branches []*clients.BranchRef) []*clients.BranchRef {
	ret := make([]*clients.BranchRef, len(branches))
	for i, branch := range branches {
		ret[i] = scrubBranch(branch)
	}
	return ret
}

func testScore(protection *clients.BranchProtectionRule,
	branch string, dl checker.DetailLogger) (int, error) {
	var score levelScore
	score.scores.basic, score.maxes.basic = basicNonAdminProtection(protection, branch, dl, true)
	score.scores.adminBasic, score.maxes.adminBasic = basicAdminProtection(protection, branch, dl, true)
	score.scores.review, score.maxes.review = nonAdminReviewProtection(protection)
	score.scores.adminReview, score.maxes.adminReview = adminReviewProtection(protection, branch, dl, true)
	score.scores.context, score.maxes.context = nonAdminContextProtection(protection, branch, dl, true)
	score.scores.thoroughReview, score.maxes.thoroughReview =
		nonAdminThoroughReviewProtection(protection, branch, dl, true)
	score.scores.adminThoroughReview, score.maxes.adminThoroughReview =
		adminThoroughReviewProtection(protection, branch, dl, true)

	return computeScore([]levelScore{score})
}

func TestReleaseAndDevBranchProtected(t *testing.T) {
	t.Parallel()

	rel1 := "release/v.1"
	sha := "8fb3cb86082b17144a80402f5367ae65f06083bd"
	main := "main"
	trueVal := true
	falseVal := false
	var zeroVal int32

	var oneVal int32 = 1

	//nolint
	tests := []struct {
		name          string
		expected      scut.TestReturn
		branches      []*clients.BranchRef
		defaultBranch string
		releases      []string
		nonadmin      bool
	}{
		{
			name: "Nil release and main branch names",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.InconclusiveResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
			defaultBranch: main,
			branches: []*clients.BranchRef{
				{
					Name:      nil,
					Protected: &trueVal,
					BranchProtectionRule: clients.BranchProtectionRule{
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
						EnforceAdmins:        &trueVal,
						RequireLinearHistory: &trueVal,
						AllowForcePushes:     &falseVal,
						AllowDeletions:       &falseVal,
					},
				},
				{
					Name:      nil,
					Protected: &trueVal,
					BranchProtectionRule: clients.BranchProtectionRule{
						CheckRules: clients.StatusChecksRule{
							RequiresStatusChecks: &trueVal,
							UpToDateBeforeMerge:  &falseVal,
							Contexts:             nil,
						},
						RequiredPullRequestReviews: clients.PullRequestReviewRule{
							DismissStaleReviews:          &falseVal,
							RequireCodeOwnerReviews:      &falseVal,
							RequiredApprovingReviewCount: &zeroVal,
						},
						EnforceAdmins:        &falseVal,
						RequireLinearHistory: &falseVal,
						AllowForcePushes:     &falseVal,
						AllowDeletions:       &falseVal,
					},
				},
				nil,
			},
			releases: []string{},
		},
		{
			name: "Only development branch",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         2,
				NumberOfWarn:  5,
				NumberOfInfo:  2,
				NumberOfDebug: 0,
			},
			defaultBranch: main,
			branches: []*clients.BranchRef{
				{
					Name:      &rel1,
					Protected: &falseVal,
				},
				{
					Name:      &main,
					Protected: &trueVal,
					BranchProtectionRule: clients.BranchProtectionRule{
						CheckRules: clients.StatusChecksRule{
							RequiresStatusChecks: &trueVal,
							UpToDateBeforeMerge:  &falseVal,
							Contexts:             nil,
						},
						RequiredPullRequestReviews: clients.PullRequestReviewRule{
							DismissStaleReviews:          &falseVal,
							RequireCodeOwnerReviews:      &falseVal,
							RequiredApprovingReviewCount: &zeroVal,
						},
						EnforceAdmins:        &falseVal,
						RequireLinearHistory: &falseVal,
						AllowForcePushes:     &falseVal,
						AllowDeletions:       &falseVal,
					},
				},
			},
			releases: nil,
		},
		{
			name: "Take worst of release and development",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         2,
				NumberOfWarn:  6,
				NumberOfInfo:  8,
				NumberOfDebug: 0,
			},
			defaultBranch: main,
			branches: []*clients.BranchRef{
				{
					Name:      &main,
					Protected: &trueVal,
					BranchProtectionRule: clients.BranchProtectionRule{
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
						EnforceAdmins:        &trueVal,
						RequireLinearHistory: &trueVal,
						AllowForcePushes:     &falseVal,
						AllowDeletions:       &falseVal,
					},
				},
				{
					Name:      &rel1,
					Protected: &trueVal,
					BranchProtectionRule: clients.BranchProtectionRule{
						CheckRules: clients.StatusChecksRule{
							RequiresStatusChecks: &trueVal,
							UpToDateBeforeMerge:  &falseVal,
							Contexts:             nil,
						},
						RequiredPullRequestReviews: clients.PullRequestReviewRule{
							DismissStaleReviews:          &falseVal,
							RequireCodeOwnerReviews:      &falseVal,
							RequiredApprovingReviewCount: &zeroVal,
						},
						EnforceAdmins:        &falseVal,
						RequireLinearHistory: &falseVal,
						AllowForcePushes:     &falseVal,
						AllowDeletions:       &falseVal,
					},
				},
			},
			releases: []string{rel1},
		},
		{
			name: "Both release and development are OK",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         8,
				NumberOfWarn:  2,
				NumberOfInfo:  12,
				NumberOfDebug: 0,
			},
			defaultBranch: main,
			branches: []*clients.BranchRef{
				{
					Name:      &main,
					Protected: &trueVal,
					BranchProtectionRule: clients.BranchProtectionRule{
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
						EnforceAdmins:        &trueVal,
						RequireLinearHistory: &trueVal,
						AllowForcePushes:     &falseVal,
						AllowDeletions:       &falseVal,
					},
				},
				{
					Name:      &rel1,
					Protected: &trueVal,
					BranchProtectionRule: clients.BranchProtectionRule{
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
						EnforceAdmins:        &trueVal,
						RequireLinearHistory: &trueVal,
						AllowForcePushes:     &falseVal,
						AllowDeletions:       &falseVal,
					},
				},
			},
			releases: []string{rel1},
		},
		{
			name: "Ignore a non-branch targetcommitish",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         2,
				NumberOfWarn:  5,
				NumberOfInfo:  2,
				NumberOfDebug: 0,
			},
			defaultBranch: main,
			releases:      []string{sha},
			branches: []*clients.BranchRef{
				{
					Name:      &main,
					Protected: &trueVal,
					BranchProtectionRule: clients.BranchProtectionRule{
						CheckRules: clients.StatusChecksRule{
							RequiresStatusChecks: &trueVal,
							UpToDateBeforeMerge:  &falseVal,
							Contexts:             nil,
						},
						RequiredPullRequestReviews: clients.PullRequestReviewRule{
							DismissStaleReviews:          &falseVal,
							RequireCodeOwnerReviews:      &falseVal,
							RequiredApprovingReviewCount: &zeroVal,
						},
						EnforceAdmins:        &falseVal,
						RequireLinearHistory: &falseVal,
						AllowForcePushes:     &falseVal,
						AllowDeletions:       &falseVal,
					},
				}, {
					Name:      &rel1,
					Protected: &falseVal,
				},
			},
		},
		{
			name: "TargetCommittish nil",
			expected: scut.TestReturn{
				Error:         sce.ErrScorecardInternal,
				Score:         checker.InconclusiveResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
			defaultBranch: main,
			releases:      []string{""},
			branches: []*clients.BranchRef{
				{
					Name:      &main,
					Protected: &trueVal,
					BranchProtectionRule: clients.BranchProtectionRule{
						CheckRules: clients.StatusChecksRule{
							RequiresStatusChecks: &trueVal,
							UpToDateBeforeMerge:  &falseVal,
							Contexts:             nil,
						},
						RequiredPullRequestReviews: clients.PullRequestReviewRule{
							DismissStaleReviews:          &falseVal,
							RequireCodeOwnerReviews:      &falseVal,
							RequiredApprovingReviewCount: &zeroVal,
						},
						EnforceAdmins:        &falseVal,
						RequireLinearHistory: &falseVal,
						AllowForcePushes:     &falseVal,
						AllowDeletions:       &falseVal,
					},
				},
			},
		},
		{
			name: "Non-admin check with protected release and development",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         0,
				NumberOfWarn:  4,
				NumberOfInfo:  0,
				NumberOfDebug: 6,
			},
			nonadmin:      true,
			defaultBranch: main,
			// branches:      []*string{&rel1, &main},
			releases: []string{rel1},
			branches: []*clients.BranchRef{
				{
					Name:      &main,
					Protected: &trueVal,
					BranchProtectionRule: clients.BranchProtectionRule{
						CheckRules: clients.StatusChecksRule{
							RequiresStatusChecks: &trueVal,
							UpToDateBeforeMerge:  &trueVal,
							Contexts:             []string{"foo"},
						},
					},
				},
				{
					Name:      &rel1,
					Protected: &trueVal,
					BranchProtectionRule: clients.BranchProtectionRule{
						CheckRules: clients.StatusChecksRule{
							RequiresStatusChecks: &trueVal,
							UpToDateBeforeMerge:  &trueVal,
							Contexts:             []string{"foo"},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
			mockRepoClient.EXPECT().GetDefaultBranch().
				DoAndReturn(func() (*clients.BranchRef, error) {
					defaultBranch := getBranch(tt.branches, tt.defaultBranch)
					if defaultBranch != nil && tt.nonadmin {
						return scrubBranch(defaultBranch), nil
					}
					return defaultBranch, nil
				}).AnyTimes()
			mockRepoClient.EXPECT().ListReleases().
				DoAndReturn(func() ([]clients.Release, error) {
					var ret []clients.Release
					for _, rel := range tt.releases {
						ret = append(ret, clients.Release{
							TargetCommitish: rel,
						})
					}
					return ret, nil
				}).AnyTimes()
			mockRepoClient.EXPECT().ListBranches().
				DoAndReturn(func() ([]*clients.BranchRef, error) {
					if tt.nonadmin {
						return scrubBranches(tt.branches), nil
					}
					return tt.branches, nil
				}).AnyTimes()
			dl := scut.TestDetailLogger{}
			r := checkReleaseAndDevBranchProtection(mockRepoClient, &dl)
			if !scut.ValidateTestReturn(t, tt.name, &tt.expected, &r, &dl) {
				t.Fail()
			}
			ctrl.Finish()
		})
	}
}

func TestIsBranchProtected(t *testing.T) {
	t.Parallel()
	trueVal := true
	falseVal := false
	var zeroVal int32
	var oneVal int32 = 1

	tests := []struct {
		name       string
		protection *clients.BranchProtectionRule
		expected   scut.TestReturn
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
			protection: &clients.BranchProtectionRule{
				CheckRules: clients.StatusChecksRule{
					RequiresStatusChecks: &trueVal,
					UpToDateBeforeMerge:  &falseVal,
					Contexts:             nil,
				},
				RequiredPullRequestReviews: clients.PullRequestReviewRule{
					DismissStaleReviews:          &falseVal,
					RequireCodeOwnerReviews:      &falseVal,
					RequiredApprovingReviewCount: &zeroVal,
				},
				EnforceAdmins:        &falseVal,
				RequireLinearHistory: &falseVal,
				AllowForcePushes:     &falseVal,
				AllowDeletions:       &falseVal,
			},
		},
		{
			name: "Nothing is enabled and values in github.Protection are nil",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         0,
				NumberOfWarn:  2,
				NumberOfInfo:  0,
				NumberOfDebug: 3,
			},
			protection: &clients.BranchProtectionRule{},
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
			protection: &clients.BranchProtectionRule{
				CheckRules: clients.StatusChecksRule{
					RequiresStatusChecks: &trueVal,
					UpToDateBeforeMerge:  &trueVal,
					Contexts:             []string{"foo"},
				},
				RequiredPullRequestReviews: clients.PullRequestReviewRule{
					DismissStaleReviews:          &falseVal,
					RequireCodeOwnerReviews:      &falseVal,
					RequiredApprovingReviewCount: &zeroVal,
				},
				EnforceAdmins:        &falseVal,
				RequireLinearHistory: &falseVal,
				AllowForcePushes:     &falseVal,
				AllowDeletions:       &falseVal,
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
			protection: &clients.BranchProtectionRule{
				CheckRules: clients.StatusChecksRule{
					RequiresStatusChecks: &trueVal,
					UpToDateBeforeMerge:  &trueVal,
					Contexts:             nil,
				},
				RequiredPullRequestReviews: clients.PullRequestReviewRule{
					DismissStaleReviews:          &falseVal,
					RequireCodeOwnerReviews:      &falseVal,
					RequiredApprovingReviewCount: &zeroVal,
				},
				EnforceAdmins:        &falseVal,
				RequireLinearHistory: &falseVal,
				AllowForcePushes:     &falseVal,
				AllowDeletions:       &falseVal,
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
			protection: &clients.BranchProtectionRule{
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
				EnforceAdmins:        &falseVal,
				RequireLinearHistory: &trueVal,
				AllowForcePushes:     &falseVal,
				AllowDeletions:       &falseVal,
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
			protection: &clients.BranchProtectionRule{
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
				EnforceAdmins:        &trueVal,
				RequireLinearHistory: &falseVal,
				AllowForcePushes:     &falseVal,
				AllowDeletions:       &falseVal,
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
			protection: &clients.BranchProtectionRule{
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
				EnforceAdmins:        &falseVal,
				RequireLinearHistory: &trueVal,
				AllowForcePushes:     &falseVal,
				AllowDeletions:       &falseVal,
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
			protection: &clients.BranchProtectionRule{
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
				EnforceAdmins:        &falseVal,
				RequireLinearHistory: &falseVal,
				AllowForcePushes:     &trueVal,
				AllowDeletions:       &falseVal,
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
			protection: &clients.BranchProtectionRule{
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
				EnforceAdmins:        &falseVal,
				RequireLinearHistory: &falseVal,
				AllowForcePushes:     &falseVal,
				AllowDeletions:       &trueVal,
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
			protection: &clients.BranchProtectionRule{
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
				EnforceAdmins:        &trueVal,
				RequireLinearHistory: &trueVal,
				AllowForcePushes:     &falseVal,
				AllowDeletions:       &falseVal,
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			score, err := testScore(tt.protection, "test", &dl)
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
