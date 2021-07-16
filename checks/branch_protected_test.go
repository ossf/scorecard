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

//nolint:gci
import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-github/v32/github"
	"github.com/ossf/scorecard/checker"
)

// TODO: these logging functions are repeated from lib/check_fn.go. Reuse code.
type log struct {
	messages []string
}

func (l *log) Logf(s string, f ...interface{}) {
	l.messages = append(l.messages, fmt.Sprintf(s, f...))
}

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
		ErrBranchNotFound
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

func TestReleaseAndDevBranchProtected(t *testing.T) { //nolint:tparallel // mocks return different results per test case
	t.Parallel()
	l := log{}

	rel1 := "release/v.1"
	sha := "8fb3cb86082b17144a80402f5367ae65f06083bd"
	main := "main"
	//nolint
	tests := []struct {
		name          string
		branches      []*string
		defaultBranch *string
		releases      []*string
		protections   map[string]*github.Protection
		want          checker.CheckResult
	}{
		{
			name:          "Only development branch",
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
			want: checker.CheckResult{
				Name:        CheckBranchProtection,
				Pass:        false,
				Details:     nil,
				Confidence:  7,
				ShouldRetry: false,
				Error:       nil,
			},
		},
		{
			name:          "Take worst of release and development",
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
			want: checker.CheckResult{
				Name:        CheckBranchProtection,
				Pass:        false,
				Details:     nil,
				Confidence:  7,
				ShouldRetry: false,
				Error:       nil,
			},
		},
		{
			name:          "Both release and development are OK",
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
			want: checker.CheckResult{
				Name:        CheckBranchProtection,
				Pass:        true,
				Details:     nil,
				Confidence:  10,
				ShouldRetry: false,
				Error:       nil,
			},
		},
		{
			name:          "Ignore a non-branch targetcommitish",
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
			want: checker.CheckResult{
				Name:        CheckBranchProtection,
				Pass:        false,
				Details:     nil,
				Confidence:  7,
				ShouldRetry: false,
				Error:       nil,
			},
		},
		{
			name:          "TargetCommittish nil",
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
			want: checker.CheckResult{
				Name:        CheckBranchProtection,
				Pass:        false,
				Details:     nil,
				Confidence:  10,
				ShouldRetry: false,
				Error:       ErrCommitishNil,
			},
		},
	}

	for _, tt := range tests { //nolint:paralleltest // mocks return different results per test case
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		l.messages = []string{}

		t.Run(tt.name, func(t *testing.T) {
			m := mockRepos{
				defaultBranch: tt.defaultBranch,
				branches:      tt.branches,
				releases:      tt.releases,
				protections:   tt.protections,
			}
			got := checkReleaseAndDevBranchProtection(context.Background(), m,
				l.Logf, "testowner", "testrepo")
			got.Details = l.messages
			if got.Confidence != tt.want.Confidence || got.Pass != tt.want.Pass {
				t.Errorf("IsBranchProtected() = %s, %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestIsBranchProtected(t *testing.T) {
	t.Parallel()
	type args struct {
		protection *github.Protection
	}

	l := log{}
	tests := []struct {
		name string
		args args
		want checker.CheckResult
	}{
		{
			name: "Nothing is enabled",
			args: args{
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
			want: checker.CheckResult{
				Name:        CheckBranchProtection,
				Pass:        false,
				Details:     nil,
				Confidence:  7,
				ShouldRetry: false,
				Error:       nil,
			},
		},
		{
			name: "Required status check enabled",
			args: args{
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
			want: checker.CheckResult{
				Name:        CheckBranchProtection,
				Pass:        false,
				Details:     nil,
				Confidence:  5,
				ShouldRetry: false,
				Error:       nil,
			},
		},
		{
			name: "Required status check enabled without checking for status string",
			args: args{
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
			want: checker.CheckResult{
				Name:        CheckBranchProtection,
				Pass:        false,
				Details:     nil,
				Confidence:  7,
				ShouldRetry: false,
				Error:       nil,
			},
		},

		{
			name: "Required pull request enabled",
			args: args{
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
			want: checker.CheckResult{
				Name:        CheckBranchProtection,
				Pass:        false,
				Details:     nil,
				Confidence:  5,
				ShouldRetry: false,
				Error:       nil,
			},
		},
		{
			name: "Required admin enforcement enabled",
			args: args{
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
			want: checker.CheckResult{
				Name:        CheckBranchProtection,
				Pass:        false,
				Details:     nil,
				Confidence:  5,
				ShouldRetry: false,
				Error:       nil,
			},
		},
		{
			name: "Required linear history enabled",
			args: args{
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
			want: checker.CheckResult{
				Name:        CheckBranchProtection,
				Pass:        false,
				Details:     nil,
				Confidence:  5,
				ShouldRetry: false,
				Error:       nil,
			},
		},
		{
			name: "Allow force push enabled",
			args: args{
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
			want: checker.CheckResult{
				Name:        CheckBranchProtection,
				Pass:        false,
				Details:     nil,
				Confidence:  9,
				ShouldRetry: false,
				Error:       nil,
			},
		},
		{
			name: "Allow deletions enabled",
			args: args{
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
			want: checker.CheckResult{
				Name:        CheckBranchProtection,
				Pass:        false,
				Details:     nil,
				Confidence:  9,
				ShouldRetry: false,
				Error:       nil,
			},
		},
		{
			name: "Branches are protected",
			args: args{
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
			want: checker.CheckResult{
				Name:        CheckBranchProtection,
				Pass:        true,
				Details:     nil,
				Confidence:  10,
				ShouldRetry: false,
				Error:       nil,
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		l.messages = []string{}
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := IsBranchProtected(tt.args.protection, "test", l.Logf)
			got.Details = l.messages
			if got.Confidence != tt.want.Confidence || got.Pass != tt.want.Pass {
				t.Errorf("IsBranchProtected() = %s, %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
