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
	"net/http"
	"testing"

	"github.com/google/go-github/v32/github"

	"github.com/ossf/scorecard/checker"
	sce "github.com/ossf/scorecard/errors"
	scut "github.com/ossf/scorecard/utests"
)

type mockRepos struct {
	branches      []*string
	protections   map[string]*github.Protection
	defaultBranch *string
	releases      []*string
}

func (m mockRepos) Get(ctx context.Context, o, r string) (
	*github.Repository, *github.Response, error) {
	return &github.Repository{
		DefaultBranch: m.defaultBranch,
	}, nil, nil
}

func (m mockRepos) ListReleases(ctx context.Context, owner string,
	repo string, opts *github.ListOptions) ([]*github.RepositoryRelease, *github.Response, error) {
	res := make([]*github.RepositoryRelease, len(m.releases))
	for i, rel := range m.releases {
		res[i] = &github.RepositoryRelease{TargetCommitish: rel}
	}
	return res, nil, nil
}

func (m mockRepos) GetBranchProtection(ctx context.Context, o string, r string,
	b string) (*github.Protection, *github.Response, error) {
	p, ok := m.protections[b]
	if ok {
		return p, &github.Response{
			Response: &http.Response{StatusCode: http.StatusAccepted},
		}, nil
	}
	return nil, &github.Response{
			Response: &http.Response{StatusCode: http.StatusNotFound},
		},
		//nolint
		sce.Create(sce.ErrScorecardInternal, errInternalBranchNotFound.Error())
}

func (m mockRepos) ListBranches(ctx context.Context, owner string, repo string,
	opts *github.BranchListOptions) ([]*github.Branch, *github.Response, error) {
	res := make([]*github.Branch, len(m.branches))
	for i, rel := range m.branches {
		_, protected := m.protections[*rel]
		res[i] = &github.Branch{Name: rel, Protected: &protected}
	}
	return res, nil, nil
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
		branches      []*string
		defaultBranch *string
		releases      []*string
		protections   map[string]*github.Protection
	}{
		{
			name: "Only development branch",
			expected: scut.TestReturn{
				Errors:        nil,
				Score:         2,
				NumberOfWarn:  6,
				NumberOfInfo:  2,
				NumberOfDebug: 0,
			},
			defaultBranch: &main,
			branches:      []*string{&rel1, &main},
			releases:      nil,
			protections: map[string]*github.Protection{
				"main": {
					RequiredStatusChecks: &github.RequiredStatusChecks{
						Strict:   false,
						Contexts: nil,
					},
					RequiredPullRequestReviews: &github.PullRequestReviewsEnforcement{
						DismissalRestrictions: &github.DismissalRestrictions{
							Users: nil,
							Teams: nil,
						},
						DismissStaleReviews:          false,
						RequireCodeOwnerReviews:      false,
						RequiredApprovingReviewCount: 0,
					},
					EnforceAdmins: &github.AdminEnforcement{
						URL:     nil,
						Enabled: false,
					},
					Restrictions: &github.BranchRestrictions{
						Users: nil,
						Teams: nil,
						Apps:  nil,
					},
					RequireLinearHistory: &github.RequireLinearHistory{
						Enabled: false,
					},
					AllowForcePushes: &github.AllowForcePushes{
						Enabled: false,
					},
					AllowDeletions: &github.AllowDeletions{
						Enabled: false,
					},
				},
			},
		},
		{
			name: "Take worst of release and development",
			expected: scut.TestReturn{
				Errors:        nil,
				Score:         2,
				NumberOfWarn:  9,
				NumberOfInfo:  7,
				NumberOfDebug: 0,
			},
			defaultBranch: &main,
			branches:      []*string{&rel1, &main},
			releases:      []*string{&rel1},
			protections: map[string]*github.Protection{
				"main": {
					RequiredStatusChecks: &github.RequiredStatusChecks{
						Strict:   true,
						Contexts: []string{"foo"},
					},
					RequiredPullRequestReviews: &github.PullRequestReviewsEnforcement{
						DismissalRestrictions: &github.DismissalRestrictions{
							Users: nil,
							Teams: nil,
						},
						DismissStaleReviews:          true,
						RequireCodeOwnerReviews:      true,
						RequiredApprovingReviewCount: 1,
					},
					EnforceAdmins: &github.AdminEnforcement{
						URL:     nil,
						Enabled: true,
					},
					Restrictions: &github.BranchRestrictions{
						Users: nil,
						Teams: nil,
						Apps:  nil,
					},
					RequireLinearHistory: &github.RequireLinearHistory{
						Enabled: true,
					},
					AllowForcePushes: &github.AllowForcePushes{
						Enabled: false,
					},
					AllowDeletions: &github.AllowDeletions{
						Enabled: false,
					},
				},
				"release/v.1": {
					RequiredStatusChecks: &github.RequiredStatusChecks{
						Strict:   false,
						Contexts: nil,
					},
					RequiredPullRequestReviews: &github.PullRequestReviewsEnforcement{
						DismissalRestrictions: &github.DismissalRestrictions{
							Users: nil,
							Teams: nil,
						},
						DismissStaleReviews:          false,
						RequireCodeOwnerReviews:      false,
						RequiredApprovingReviewCount: 0,
					},
					EnforceAdmins: &github.AdminEnforcement{
						URL:     nil,
						Enabled: false,
					},
					Restrictions: &github.BranchRestrictions{
						Users: nil,
						Teams: nil,
						Apps:  nil,
					},
					RequireLinearHistory: &github.RequireLinearHistory{
						Enabled: false,
					},
					AllowForcePushes: &github.AllowForcePushes{
						Enabled: false,
					},
					AllowDeletions: &github.AllowDeletions{
						Enabled: false,
					},
				},
			},
		},
		{
			name: "Both release and development are OK",
			expected: scut.TestReturn{
				Errors:        nil,
				Score:         5,
				NumberOfWarn:  6,
				NumberOfInfo:  10,
				NumberOfDebug: 0,
			},
			defaultBranch: &main,
			branches:      []*string{&rel1, &main},
			releases:      []*string{&rel1},
			protections: map[string]*github.Protection{
				"main": {
					RequiredStatusChecks: &github.RequiredStatusChecks{
						Strict:   true,
						Contexts: []string{"foo"},
					},
					RequiredPullRequestReviews: &github.PullRequestReviewsEnforcement{
						DismissalRestrictions: &github.DismissalRestrictions{
							Users: nil,
							Teams: nil,
						},
						DismissStaleReviews:          true,
						RequireCodeOwnerReviews:      true,
						RequiredApprovingReviewCount: 1,
					},
					EnforceAdmins: &github.AdminEnforcement{
						URL:     nil,
						Enabled: true,
					},
					Restrictions: &github.BranchRestrictions{
						Users: nil,
						Teams: nil,
						Apps:  nil,
					},
					RequireLinearHistory: &github.RequireLinearHistory{
						Enabled: true,
					},
					AllowForcePushes: &github.AllowForcePushes{
						Enabled: false,
					},
					AllowDeletions: &github.AllowDeletions{
						Enabled: false,
					},
				},
				"release/v.1": {
					RequiredStatusChecks: &github.RequiredStatusChecks{
						Strict:   true,
						Contexts: []string{"foo"},
					},
					RequiredPullRequestReviews: &github.PullRequestReviewsEnforcement{
						DismissalRestrictions: &github.DismissalRestrictions{
							Users: nil,
							Teams: nil,
						},
						DismissStaleReviews:          true,
						RequireCodeOwnerReviews:      true,
						RequiredApprovingReviewCount: 1,
					},
					EnforceAdmins: &github.AdminEnforcement{
						URL:     nil,
						Enabled: true,
					},
					Restrictions: &github.BranchRestrictions{
						Users: nil,
						Teams: nil,
						Apps:  nil,
					},
					RequireLinearHistory: &github.RequireLinearHistory{
						Enabled: true,
					},
					AllowForcePushes: &github.AllowForcePushes{
						Enabled: false,
					},
					AllowDeletions: &github.AllowDeletions{
						Enabled: false,
					},
				},
			},
		},
		{
			name: "Ignore a non-branch targetcommitish",
			expected: scut.TestReturn{
				Errors:        nil,
				Score:         2,
				NumberOfWarn:  6,
				NumberOfInfo:  2,
				NumberOfDebug: 0,
			},
			defaultBranch: &main,
			branches:      []*string{&rel1, &main},
			releases:      []*string{&sha},
			protections: map[string]*github.Protection{
				"main": {
					RequiredStatusChecks: &github.RequiredStatusChecks{
						Strict:   false,
						Contexts: nil,
					},
					RequiredPullRequestReviews: &github.PullRequestReviewsEnforcement{
						DismissalRestrictions: &github.DismissalRestrictions{
							Users: nil,
							Teams: nil,
						},
						DismissStaleReviews:          false,
						RequireCodeOwnerReviews:      false,
						RequiredApprovingReviewCount: 0,
					},
					EnforceAdmins: &github.AdminEnforcement{
						URL:     nil,
						Enabled: false,
					},
					Restrictions: &github.BranchRestrictions{
						Users: nil,
						Teams: nil,
						Apps:  nil,
					},
					RequireLinearHistory: &github.RequireLinearHistory{
						Enabled: false,
					},
					AllowForcePushes: &github.AllowForcePushes{
						Enabled: false,
					},
					AllowDeletions: &github.AllowDeletions{
						Enabled: false,
					},
				},
			},
		},
		{
			name: "TargetCommittish nil",
			expected: scut.TestReturn{
				Errors:        []error{sce.ErrScorecardInternal},
				Score:         checker.InconclusiveResultScore,
				NumberOfWarn:  6,
				NumberOfInfo:  2,
				NumberOfDebug: 0,
			},
			defaultBranch: &main,
			branches:      []*string{&main},
			releases:      []*string{nil},
			protections: map[string]*github.Protection{
				"main": {
					RequiredStatusChecks: &github.RequiredStatusChecks{
						Strict:   false,
						Contexts: nil,
					},
					RequiredPullRequestReviews: &github.PullRequestReviewsEnforcement{
						DismissalRestrictions: &github.DismissalRestrictions{
							Users: nil,
							Teams: nil,
						},
						DismissStaleReviews:          false,
						RequireCodeOwnerReviews:      false,
						RequiredApprovingReviewCount: 0,
					},
					EnforceAdmins: &github.AdminEnforcement{
						URL:     nil,
						Enabled: false,
					},
					Restrictions: &github.BranchRestrictions{
						Users: nil,
						Teams: nil,
						Apps:  nil,
					},
					RequireLinearHistory: &github.RequireLinearHistory{
						Enabled: false,
					},
					AllowForcePushes: &github.AllowForcePushes{
						Enabled: false,
					},
					AllowDeletions: &github.AllowDeletions{
						Enabled: false,
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
				defaultBranch: tt.defaultBranch,
				branches:      tt.branches,
				releases:      tt.releases,
				protections:   tt.protections,
			}
			dl := scut.TestDetailLogger{}
			r := checkReleaseAndDevBranchProtection(context.Background(), m,
				&dl, "testowner", "testrepo")
			scut.ValidateTestReturn(t, tt.name, &tt.expected, &r, &dl)
		})
	}
}

func TestIsBranchProtected(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		protection *github.Protection
		expected   scut.TestReturn
	}{
		{
			name: "Nothing is enabled",
			expected: scut.TestReturn{
				Errors:        nil,
				Score:         2,
				NumberOfWarn:  6,
				NumberOfInfo:  2,
				NumberOfDebug: 0,
			},
			protection: &github.Protection{
				RequiredStatusChecks: &github.RequiredStatusChecks{
					Strict:   false,
					Contexts: nil,
				},
				RequiredPullRequestReviews: &github.PullRequestReviewsEnforcement{
					DismissalRestrictions: &github.DismissalRestrictions{
						Users: nil,
						Teams: nil,
					},
					DismissStaleReviews:          false,
					RequireCodeOwnerReviews:      false,
					RequiredApprovingReviewCount: 0,
				},
				EnforceAdmins: &github.AdminEnforcement{
					URL:     nil,
					Enabled: false,
				},
				Restrictions: &github.BranchRestrictions{
					Users: nil,
					Teams: nil,
					Apps:  nil,
				},
				RequireLinearHistory: &github.RequireLinearHistory{
					Enabled: false,
				},
				AllowForcePushes: &github.AllowForcePushes{
					Enabled: false,
				},
				AllowDeletions: &github.AllowDeletions{
					Enabled: false,
				},
			},
		},
		{
			name: "Required status check enabled",
			expected: scut.TestReturn{
				Errors:        nil,
				Score:         3,
				NumberOfWarn:  5,
				NumberOfInfo:  3,
				NumberOfDebug: 0,
			},
			protection: &github.Protection{
				RequiredStatusChecks: &github.RequiredStatusChecks{
					Strict:   true,
					Contexts: []string{"foo"},
				},
				RequiredPullRequestReviews: &github.PullRequestReviewsEnforcement{
					DismissalRestrictions: &github.DismissalRestrictions{
						Users: nil,
						Teams: nil,
					},
					DismissStaleReviews:          false,
					RequireCodeOwnerReviews:      false,
					RequiredApprovingReviewCount: 0,
				},
				EnforceAdmins: &github.AdminEnforcement{
					URL:     nil,
					Enabled: false,
				},
				Restrictions: &github.BranchRestrictions{
					Users: nil,
					Teams: nil,
					Apps:  nil,
				},
				RequireLinearHistory: &github.RequireLinearHistory{
					Enabled: false,
				},
				AllowForcePushes: &github.AllowForcePushes{
					Enabled: false,
				},
				AllowDeletions: &github.AllowDeletions{
					Enabled: false,
				},
			},
		},
		{
			name: "Required status check enabled without checking for status string",
			expected: scut.TestReturn{
				Errors:        nil,
				Score:         2,
				NumberOfWarn:  6,
				NumberOfInfo:  2,
				NumberOfDebug: 0,
			},
			protection: &github.Protection{
				RequiredStatusChecks: &github.RequiredStatusChecks{
					Strict:   true,
					Contexts: nil,
				},
				RequiredPullRequestReviews: &github.PullRequestReviewsEnforcement{
					DismissalRestrictions: &github.DismissalRestrictions{
						Users: nil,
						Teams: nil,
					},
					DismissStaleReviews:          false,
					RequireCodeOwnerReviews:      false,
					RequiredApprovingReviewCount: 0,
				},
				EnforceAdmins: &github.AdminEnforcement{
					URL:     nil,
					Enabled: false,
				},
				Restrictions: &github.BranchRestrictions{
					Users: nil,
					Teams: nil,
					Apps:  nil,
				},
				RequireLinearHistory: &github.RequireLinearHistory{
					Enabled: false,
				},
				AllowForcePushes: &github.AllowForcePushes{
					Enabled: false,
				},
				AllowDeletions: &github.AllowDeletions{
					Enabled: false,
				},
			},
		},
		{
			name: "Required pull request enabled",
			expected: scut.TestReturn{
				Errors:        nil,
				Score:         3,
				NumberOfWarn:  5,
				NumberOfInfo:  3,
				NumberOfDebug: 0,
			},
			protection: &github.Protection{
				RequiredStatusChecks: &github.RequiredStatusChecks{
					Strict:   false,
					Contexts: []string{"foo"},
				},
				RequiredPullRequestReviews: &github.PullRequestReviewsEnforcement{
					DismissalRestrictions: &github.DismissalRestrictions{
						Users: nil,
						Teams: nil,
					},
					DismissStaleReviews:          false,
					RequireCodeOwnerReviews:      false,
					RequiredApprovingReviewCount: 1,
				},
				EnforceAdmins: &github.AdminEnforcement{
					URL:     nil,
					Enabled: false,
				},
				Restrictions: &github.BranchRestrictions{
					Users: nil,
					Teams: nil,
					Apps:  nil,
				},
				RequireLinearHistory: &github.RequireLinearHistory{
					Enabled: true,
				},
				AllowForcePushes: &github.AllowForcePushes{
					Enabled: false,
				},
				AllowDeletions: &github.AllowDeletions{
					Enabled: false,
				},
			},
		},
		{
			name: "Required admin enforcement enabled",
			expected: scut.TestReturn{
				Errors:        nil,
				Score:         3,
				NumberOfWarn:  5,
				NumberOfInfo:  3,
				NumberOfDebug: 0,
			},
			protection: &github.Protection{
				RequiredStatusChecks: &github.RequiredStatusChecks{
					Strict:   false,
					Contexts: []string{"foo"},
				},
				RequiredPullRequestReviews: &github.PullRequestReviewsEnforcement{
					DismissalRestrictions: &github.DismissalRestrictions{
						Users: nil,
						Teams: nil,
					},
					DismissStaleReviews:          false,
					RequireCodeOwnerReviews:      false,
					RequiredApprovingReviewCount: 0,
				},
				EnforceAdmins: &github.AdminEnforcement{
					URL:     nil,
					Enabled: true,
				},
				Restrictions: &github.BranchRestrictions{
					Users: nil,
					Teams: nil,
					Apps:  nil,
				},
				RequireLinearHistory: &github.RequireLinearHistory{
					Enabled: false,
				},
				AllowForcePushes: &github.AllowForcePushes{
					Enabled: false,
				},
				AllowDeletions: &github.AllowDeletions{
					Enabled: false,
				},
			},
		},
		{
			name: "Required linear history enabled",
			expected: scut.TestReturn{
				Errors:        nil,
				Score:         3,
				NumberOfWarn:  5,
				NumberOfInfo:  3,
				NumberOfDebug: 0,
			},
			protection: &github.Protection{
				RequiredStatusChecks: &github.RequiredStatusChecks{
					Strict:   false,
					Contexts: []string{"foo"},
				},
				RequiredPullRequestReviews: &github.PullRequestReviewsEnforcement{
					DismissalRestrictions: &github.DismissalRestrictions{
						Users: nil,
						Teams: nil,
					},
					DismissStaleReviews:          false,
					RequireCodeOwnerReviews:      false,
					RequiredApprovingReviewCount: 0,
				},
				EnforceAdmins: &github.AdminEnforcement{
					URL:     nil,
					Enabled: false,
				},
				Restrictions: &github.BranchRestrictions{
					Users: nil,
					Teams: nil,
					Apps:  nil,
				},
				RequireLinearHistory: &github.RequireLinearHistory{
					Enabled: true,
				},
				AllowForcePushes: &github.AllowForcePushes{
					Enabled: false,
				},
				AllowDeletions: &github.AllowDeletions{
					Enabled: false,
				},
			},
		},
		{
			name: "Allow force push enabled",
			expected: scut.TestReturn{
				Errors:        nil,
				Score:         1,
				NumberOfWarn:  7,
				NumberOfInfo:  1,
				NumberOfDebug: 0,
			},
			protection: &github.Protection{
				RequiredStatusChecks: &github.RequiredStatusChecks{
					Strict:   false,
					Contexts: []string{"foo"},
				},
				RequiredPullRequestReviews: &github.PullRequestReviewsEnforcement{
					DismissalRestrictions: &github.DismissalRestrictions{
						Users: nil,
						Teams: nil,
					},
					DismissStaleReviews:          false,
					RequireCodeOwnerReviews:      false,
					RequiredApprovingReviewCount: 0,
				},
				EnforceAdmins: &github.AdminEnforcement{
					URL:     nil,
					Enabled: false,
				},
				Restrictions: &github.BranchRestrictions{
					Users: nil,
					Teams: nil,
					Apps:  nil,
				},
				RequireLinearHistory: &github.RequireLinearHistory{
					Enabled: false,
				},
				AllowForcePushes: &github.AllowForcePushes{
					Enabled: true,
				},
				AllowDeletions: &github.AllowDeletions{
					Enabled: false,
				},
			},
		},
		{
			name: "Allow deletions enabled",
			expected: scut.TestReturn{
				Errors:        nil,
				Score:         1,
				NumberOfWarn:  7,
				NumberOfInfo:  1,
				NumberOfDebug: 0,
			},
			protection: &github.Protection{
				RequiredStatusChecks: &github.RequiredStatusChecks{
					Strict:   false,
					Contexts: []string{"foo"},
				},
				RequiredPullRequestReviews: &github.PullRequestReviewsEnforcement{
					DismissalRestrictions: &github.DismissalRestrictions{
						Users: nil,
						Teams: nil,
					},
					DismissStaleReviews:          false,
					RequireCodeOwnerReviews:      false,
					RequiredApprovingReviewCount: 0,
				},
				EnforceAdmins: &github.AdminEnforcement{
					URL:     nil,
					Enabled: false,
				},
				Restrictions: &github.BranchRestrictions{
					Users: nil,
					Teams: nil,
					Apps:  nil,
				},
				RequireLinearHistory: &github.RequireLinearHistory{
					Enabled: false,
				},
				AllowForcePushes: &github.AllowForcePushes{
					Enabled: false,
				},
				AllowDeletions: &github.AllowDeletions{
					Enabled: true,
				},
			},
		},
		{
			name: "Branches are protected",
			expected: scut.TestReturn{
				Errors:        nil,
				Score:         5,
				NumberOfWarn:  3,
				NumberOfInfo:  5,
				NumberOfDebug: 0,
			},
			protection: &github.Protection{
				RequiredStatusChecks: &github.RequiredStatusChecks{
					Strict:   true,
					Contexts: []string{"foo"},
				},
				RequiredPullRequestReviews: &github.PullRequestReviewsEnforcement{
					DismissalRestrictions: &github.DismissalRestrictions{
						Users: nil,
						Teams: nil,
					},
					DismissStaleReviews:          true,
					RequireCodeOwnerReviews:      true,
					RequiredApprovingReviewCount: 1,
				},
				EnforceAdmins: &github.AdminEnforcement{
					URL:     nil,
					Enabled: true,
				},
				Restrictions: &github.BranchRestrictions{
					Users: nil,
					Teams: nil,
					Apps:  nil,
				},
				RequireLinearHistory: &github.RequireLinearHistory{
					Enabled: true,
				},
				AllowForcePushes: &github.AllowForcePushes{
					Enabled: false,
				},
				AllowDeletions: &github.AllowDeletions{
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
			r := IsBranchProtected(tt.protection, "test", &dl)
			scut.ValidateTestReturn(t, tt.name, &tt.expected, &r, &dl)
		})
	}
}
