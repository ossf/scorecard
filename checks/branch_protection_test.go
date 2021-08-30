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
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-github/v38/github"

	"github.com/ossf/scorecard/v2/checker"
	"github.com/ossf/scorecard/v2/clients"
	"github.com/ossf/scorecard/v2/clients/mockrepo"
	sce "github.com/ossf/scorecard/v2/errors"
	scut "github.com/ossf/scorecard/v2/utests"
)

type mockRepos struct {
	releases []*string
	nonadmin bool
}

func (m mockRepos) ListReleases(ctx context.Context, owner string,
	repo string, opts *github.ListOptions) ([]*github.RepositoryRelease, *github.Response, error) {
	res := make([]*github.RepositoryRelease, len(m.releases))
	for i, rel := range m.releases {
		res[i] = &github.RepositoryRelease{TargetCommitish: rel}
	}
	return res, nil, nil
}

func getBranch(branches []*clients.BranchRef, name string) *clients.BranchRef {
	for _, branch := range branches {
		if branch.GetName() == name {
			return branch
		}
	}
	return nil
}

func scrubBranch(branch *clients.BranchRef) *clients.BranchRef {
	ret := branch
	ret.BranchProtectionRule = nil
	return ret
}

func scrubBranches(branches []*clients.BranchRef) []*clients.BranchRef {
	ret := make([]*clients.BranchRef, len(branches))
	for i, branch := range branches {
		ret[i] = scrubBranch(branch)
	}
	return ret
}

func TestReleaseAndDevBranchProtected(t *testing.T) {
	t.Parallel()

	rel1 := "release/v.1"
	sha := "8fb3cb86082b17144a80402f5367ae65f06083bd"
	main := "main"
	//nolint
	tests := []struct {
		name          string
		expected      scut.TestReturn
		branches      []*clients.BranchRef
		defaultBranch string
		releases      []*string
		nonadmin      bool
	}{
		{
			name: "Only development branch",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         1,
				NumberOfWarn:  6,
				NumberOfInfo:  2,
				NumberOfDebug: 0,
			},
			defaultBranch: main,
			branches: []*clients.BranchRef{
				{
					Name:      rel1,
					Protected: false,
				},
				{
					Name:      main,
					Protected: true,
					BranchProtectionRule: &clients.BranchProtectionRule{
						RequiredStatusChecks: &clients.StatusChecksRule{
							Strict:   false,
							Contexts: nil,
						},
						RequiredPullRequestReviews: &clients.PullRequestReviewRule{
							DismissStaleReviews:          false,
							RequireCodeOwnerReviews:      false,
							RequiredApprovingReviewCount: 0,
						},
						EnforceAdmins: &clients.EnforceAdmins{
							Enabled: false,
						},
						RequireLinearHistory: &clients.RequireLinearHistory{
							Enabled: false,
						},
						AllowForcePushes: &clients.AllowForcePushes{
							Enabled: false,
						},
						AllowDeletions: &clients.AllowDeletions{
							Enabled: false,
						},
					},
				},
			},
			releases: nil,
		},
		{
			name: "Take worst of release and development",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         5,
				NumberOfWarn:  8,
				NumberOfInfo:  9,
				NumberOfDebug: 0,
			},
			defaultBranch: main,
			branches: []*clients.BranchRef{
				{
					Name:      main,
					Protected: true,
					BranchProtectionRule: &clients.BranchProtectionRule{
						RequiredStatusChecks: &clients.StatusChecksRule{
							Strict:   true,
							Contexts: []string{"foo"},
						},
						RequiredPullRequestReviews: &clients.PullRequestReviewRule{
							DismissStaleReviews:          true,
							RequireCodeOwnerReviews:      true,
							RequiredApprovingReviewCount: 1,
						},
						EnforceAdmins: &clients.EnforceAdmins{
							Enabled: true,
						},
						RequireLinearHistory: &clients.RequireLinearHistory{
							Enabled: true,
						},
						AllowForcePushes: &clients.AllowForcePushes{
							Enabled: false,
						},
						AllowDeletions: &clients.AllowDeletions{
							Enabled: false,
						},
					},
				},
				{
					Name:      rel1,
					Protected: true,
					BranchProtectionRule: &clients.BranchProtectionRule{
						RequiredStatusChecks: &clients.StatusChecksRule{
							Strict:   false,
							Contexts: nil,
						},
						RequiredPullRequestReviews: &clients.PullRequestReviewRule{
							DismissStaleReviews:          false,
							RequireCodeOwnerReviews:      false,
							RequiredApprovingReviewCount: 0,
						},
						EnforceAdmins: &clients.EnforceAdmins{
							Enabled: false,
						},
						RequireLinearHistory: &clients.RequireLinearHistory{
							Enabled: false,
						},
						AllowForcePushes: &clients.AllowForcePushes{
							Enabled: false,
						},
						AllowDeletions: &clients.AllowDeletions{
							Enabled: false,
						},
					},
				},
			},
			releases: []*string{&rel1},
		},
		{
			name: "Both release and development are OK",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         9,
				NumberOfWarn:  4,
				NumberOfInfo:  14,
				NumberOfDebug: 0,
			},
			defaultBranch: main,
			branches: []*clients.BranchRef{
				{
					Name:      main,
					Protected: true,
					BranchProtectionRule: &clients.BranchProtectionRule{
						RequiredStatusChecks: &clients.StatusChecksRule{
							Strict:   true,
							Contexts: []string{"foo"},
						},
						RequiredPullRequestReviews: &clients.PullRequestReviewRule{
							DismissStaleReviews:          true,
							RequireCodeOwnerReviews:      true,
							RequiredApprovingReviewCount: 1,
						},
						EnforceAdmins: &clients.EnforceAdmins{
							Enabled: true,
						},
						RequireLinearHistory: &clients.RequireLinearHistory{
							Enabled: true,
						},
						AllowForcePushes: &clients.AllowForcePushes{
							Enabled: false,
						},
						AllowDeletions: &clients.AllowDeletions{
							Enabled: false,
						},
					},
				},
				{
					Name:      rel1,
					Protected: true,
					BranchProtectionRule: &clients.BranchProtectionRule{
						RequiredStatusChecks: &clients.StatusChecksRule{
							Strict:   true,
							Contexts: []string{"foo"},
						},
						RequiredPullRequestReviews: &clients.PullRequestReviewRule{
							DismissStaleReviews:          true,
							RequireCodeOwnerReviews:      true,
							RequiredApprovingReviewCount: 1,
						},
						EnforceAdmins: &clients.EnforceAdmins{
							Enabled: true,
						},
						RequireLinearHistory: &clients.RequireLinearHistory{
							Enabled: true,
						},
						AllowForcePushes: &clients.AllowForcePushes{
							Enabled: false,
						},
						AllowDeletions: &clients.AllowDeletions{
							Enabled: false,
						},
					},
				},
			},
			releases: []*string{&rel1},
		},
		{
			name: "Ignore a non-branch targetcommitish",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         1,
				NumberOfWarn:  6,
				NumberOfInfo:  2,
				NumberOfDebug: 0,
			},
			defaultBranch: main,
			releases:      []*string{&sha},
			branches: []*clients.BranchRef{
				{
					Name:      main,
					Protected: true,
					BranchProtectionRule: &clients.BranchProtectionRule{
						RequiredStatusChecks: &clients.StatusChecksRule{
							Strict:   false,
							Contexts: nil,
						},
						RequiredPullRequestReviews: &clients.PullRequestReviewRule{
							DismissStaleReviews:          false,
							RequireCodeOwnerReviews:      false,
							RequiredApprovingReviewCount: 0,
						},
						EnforceAdmins: &clients.EnforceAdmins{
							Enabled: false,
						},
						RequireLinearHistory: &clients.RequireLinearHistory{
							Enabled: false,
						},
						AllowForcePushes: &clients.AllowForcePushes{
							Enabled: false,
						},
						AllowDeletions: &clients.AllowDeletions{
							Enabled: false,
						},
					},
				}, {
					Name:      rel1,
					Protected: false,
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
			releases:      []*string{nil},
			branches: []*clients.BranchRef{
				{
					Name:      main,
					Protected: true,
					BranchProtectionRule: &clients.BranchProtectionRule{
						RequiredStatusChecks: &clients.StatusChecksRule{
							Strict:   false,
							Contexts: nil,
						},
						RequiredPullRequestReviews: &clients.PullRequestReviewRule{
							DismissStaleReviews:          false,
							RequireCodeOwnerReviews:      false,
							RequiredApprovingReviewCount: 0,
						},
						EnforceAdmins: &clients.EnforceAdmins{
							Enabled: false,
						},
						RequireLinearHistory: &clients.RequireLinearHistory{
							Enabled: false,
						},
						AllowForcePushes: &clients.AllowForcePushes{
							Enabled: false,
						},
						AllowDeletions: &clients.AllowDeletions{
							Enabled: false,
						},
					},
				},
			},
		},
		{
			name: "Non-admin check with protected release and development",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         1,
				NumberOfWarn:  2,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
			nonadmin:      true,
			defaultBranch: main,
			// branches:      []*string{&rel1, &main},
			releases: []*string{&rel1},
			branches: []*clients.BranchRef{
				{
					Name:      main,
					Protected: true,
					BranchProtectionRule: &clients.BranchProtectionRule{
						RequiredStatusChecks: &clients.StatusChecksRule{
							Strict:   true,
							Contexts: []string{"foo"},
						},
					},
				},
				{
					Name:      rel1,
					Protected: true,
					BranchProtectionRule: &clients.BranchProtectionRule{
						RequiredStatusChecks: &clients.StatusChecksRule{
							Strict:   true,
							Contexts: []string{"foo"},
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
			m := mockRepos{
				releases: tt.releases,
				nonadmin: tt.nonadmin,
			}

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
			mockRepoClient.EXPECT().ListBranches().
				DoAndReturn(func() ([]*clients.BranchRef, error) {
					if tt.nonadmin {
						return scrubBranches(tt.branches), nil
					}
					return tt.branches, nil
				}).AnyTimes()
			dl := scut.TestDetailLogger{}
			r := checkReleaseAndDevBranchProtection(context.Background(), mockRepoClient, m,
				&dl, "testowner", "testrepo")
			if !scut.ValidateTestReturn(t, tt.name, &tt.expected, &r, &dl) {
				t.Fail()
			}
			ctrl.Finish()
		})
	}
}

func TestIsBranchProtected(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		protection *clients.BranchProtectionRule
		expected   scut.TestReturn
	}{
		{
			name: "Nothing is enabled",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         1,
				NumberOfWarn:  6,
				NumberOfInfo:  2,
				NumberOfDebug: 0,
			},
			protection: &clients.BranchProtectionRule{
				RequiredStatusChecks: &clients.StatusChecksRule{
					Strict:   false,
					Contexts: nil,
				},
				RequiredPullRequestReviews: &clients.PullRequestReviewRule{
					DismissStaleReviews:          false,
					RequireCodeOwnerReviews:      false,
					RequiredApprovingReviewCount: 0,
				},
				EnforceAdmins: &clients.EnforceAdmins{
					Enabled: false,
				},
				RequireLinearHistory: &clients.RequireLinearHistory{
					Enabled: false,
				},
				AllowForcePushes: &clients.AllowForcePushes{
					Enabled: false,
				},
				AllowDeletions: &clients.AllowDeletions{
					Enabled: false,
				},
			},
		},
		{
			name: "Nothing is enabled and values in github.Protection are nil",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         1,
				NumberOfWarn:  4,
				NumberOfInfo:  2,
				NumberOfDebug: 0,
			},
			protection: &clients.BranchProtectionRule{},
		},
		{
			name: "Required status check enabled",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         2,
				NumberOfWarn:  6,
				NumberOfInfo:  3,
				NumberOfDebug: 0,
			},
			protection: &clients.BranchProtectionRule{
				RequiredStatusChecks: &clients.StatusChecksRule{
					Strict:   true,
					Contexts: []string{"foo"},
				},
				RequiredPullRequestReviews: &clients.PullRequestReviewRule{
					DismissStaleReviews:          false,
					RequireCodeOwnerReviews:      false,
					RequiredApprovingReviewCount: 0,
				},
				EnforceAdmins: &clients.EnforceAdmins{
					Enabled: false,
				},
				RequireLinearHistory: &clients.RequireLinearHistory{
					Enabled: false,
				},
				AllowForcePushes: &clients.AllowForcePushes{
					Enabled: false,
				},
				AllowDeletions: &clients.AllowDeletions{
					Enabled: false,
				},
			},
		},
		{
			name: "Required status check enabled without checking for status string",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         2,
				NumberOfWarn:  6,
				NumberOfInfo:  3,
				NumberOfDebug: 0,
			},
			protection: &clients.BranchProtectionRule{
				RequiredStatusChecks: &clients.StatusChecksRule{
					Strict:   true,
					Contexts: nil,
				},
				RequiredPullRequestReviews: &clients.PullRequestReviewRule{
					DismissStaleReviews:          false,
					RequireCodeOwnerReviews:      false,
					RequiredApprovingReviewCount: 0,
				},
				EnforceAdmins: &clients.EnforceAdmins{
					Enabled: false,
				},
				RequireLinearHistory: &clients.RequireLinearHistory{
					Enabled: false,
				},
				AllowForcePushes: &clients.AllowForcePushes{
					Enabled: false,
				},
				AllowDeletions: &clients.AllowDeletions{
					Enabled: false,
				},
			},
		},
		{
			name: "Required pull request enabled",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         2,
				NumberOfWarn:  5,
				NumberOfInfo:  3,
				NumberOfDebug: 0,
			},
			protection: &clients.BranchProtectionRule{
				RequiredStatusChecks: &clients.StatusChecksRule{
					Strict:   false,
					Contexts: []string{"foo"},
				},
				RequiredPullRequestReviews: &clients.PullRequestReviewRule{
					DismissStaleReviews:          false,
					RequireCodeOwnerReviews:      false,
					RequiredApprovingReviewCount: 1,
				},
				EnforceAdmins: &clients.EnforceAdmins{
					Enabled: false,
				},
				RequireLinearHistory: &clients.RequireLinearHistory{
					Enabled: true,
				},
				AllowForcePushes: &clients.AllowForcePushes{
					Enabled: false,
				},
				AllowDeletions: &clients.AllowDeletions{
					Enabled: false,
				},
			},
		},
		{
			name: "Required admin enforcement enabled",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         3,
				NumberOfWarn:  5,
				NumberOfInfo:  3,
				NumberOfDebug: 0,
			},
			protection: &clients.BranchProtectionRule{
				RequiredStatusChecks: &clients.StatusChecksRule{
					Strict:   false,
					Contexts: []string{"foo"},
				},
				RequiredPullRequestReviews: &clients.PullRequestReviewRule{
					DismissStaleReviews:          false,
					RequireCodeOwnerReviews:      false,
					RequiredApprovingReviewCount: 0,
				},
				EnforceAdmins: &clients.EnforceAdmins{
					Enabled: true,
				},
				RequireLinearHistory: &clients.RequireLinearHistory{
					Enabled: false,
				},
				AllowForcePushes: &clients.AllowForcePushes{
					Enabled: false,
				},
				AllowDeletions: &clients.AllowDeletions{
					Enabled: false,
				},
			},
		},
		{
			name: "Required linear history enabled",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         2,
				NumberOfWarn:  5,
				NumberOfInfo:  3,
				NumberOfDebug: 0,
			},
			protection: &clients.BranchProtectionRule{
				RequiredStatusChecks: &clients.StatusChecksRule{
					Strict:   false,
					Contexts: []string{"foo"},
				},
				RequiredPullRequestReviews: &clients.PullRequestReviewRule{
					DismissStaleReviews:          false,
					RequireCodeOwnerReviews:      false,
					RequiredApprovingReviewCount: 0,
				},
				EnforceAdmins: &clients.EnforceAdmins{
					Enabled: false,
				},
				RequireLinearHistory: &clients.RequireLinearHistory{
					Enabled: true,
				},
				AllowForcePushes: &clients.AllowForcePushes{
					Enabled: false,
				},
				AllowDeletions: &clients.AllowDeletions{
					Enabled: false,
				},
			},
		},
		{
			name: "Allow force push enabled",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         0,
				NumberOfWarn:  7,
				NumberOfInfo:  1,
				NumberOfDebug: 0,
			},
			protection: &clients.BranchProtectionRule{
				RequiredStatusChecks: &clients.StatusChecksRule{
					Strict:   false,
					Contexts: []string{"foo"},
				},
				RequiredPullRequestReviews: &clients.PullRequestReviewRule{
					DismissStaleReviews:          false,
					RequireCodeOwnerReviews:      false,
					RequiredApprovingReviewCount: 0,
				},
				EnforceAdmins: &clients.EnforceAdmins{
					Enabled: false,
				},
				RequireLinearHistory: &clients.RequireLinearHistory{
					Enabled: false,
				},
				AllowForcePushes: &clients.AllowForcePushes{
					Enabled: true,
				},
				AllowDeletions: &clients.AllowDeletions{
					Enabled: false,
				},
			},
		},
		{
			name: "Allow deletions enabled",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         0,
				NumberOfWarn:  7,
				NumberOfInfo:  1,
				NumberOfDebug: 0,
			},
			protection: &clients.BranchProtectionRule{
				RequiredStatusChecks: &clients.StatusChecksRule{
					Strict:   false,
					Contexts: []string{"foo"},
				},
				RequiredPullRequestReviews: &clients.PullRequestReviewRule{
					DismissStaleReviews:          false,
					RequireCodeOwnerReviews:      false,
					RequiredApprovingReviewCount: 0,
				},
				EnforceAdmins: &clients.EnforceAdmins{
					Enabled: false,
				},
				RequireLinearHistory: &clients.RequireLinearHistory{
					Enabled: false,
				},
				AllowForcePushes: &clients.AllowForcePushes{
					Enabled: false,
				},
				AllowDeletions: &clients.AllowDeletions{
					Enabled: true,
				},
			},
		},
		{
			name: "Branches are protected",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         9,
				NumberOfWarn:  2,
				NumberOfInfo:  7,
				NumberOfDebug: 0,
			},
			protection: &clients.BranchProtectionRule{
				RequiredStatusChecks: &clients.StatusChecksRule{
					Strict:   true,
					Contexts: []string{"foo"},
				},
				RequiredPullRequestReviews: &clients.PullRequestReviewRule{
					DismissStaleReviews:          true,
					RequireCodeOwnerReviews:      true,
					RequiredApprovingReviewCount: 1,
				},
				EnforceAdmins: &clients.EnforceAdmins{
					Enabled: true,
				},
				RequireLinearHistory: &clients.RequireLinearHistory{
					Enabled: true,
				},
				AllowForcePushes: &clients.AllowForcePushes{
					Enabled: false,
				},
				AllowDeletions: &clients.AllowDeletions{
					Enabled: false,
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			actual := &checker.CheckResult{
				Score: isBranchProtected(tt.protection, "test", &dl),
			}
			if !scut.ValidateTestReturn(t, tt.name, &tt.expected, actual, &dl) {
				t.Fail()
			}
		})
	}
}
