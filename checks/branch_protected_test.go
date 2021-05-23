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
	"fmt"
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

//nolint:dupl // repeating test cases that are slightly different is acceptable
func TestIsBranchProtected(t *testing.T) {
	t.Parallel()
	type args struct {
		protection *github.Protection
		c          checker.CheckRequest
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
				c: checker.CheckRequest{Logf: l.Logf},
			},
			want: checker.CheckResult{
				Name:        branchProtectionStr,
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
				c: checker.CheckRequest{Logf: l.Logf},
			},
			want: checker.CheckResult{
				Name:        branchProtectionStr,
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
				c: checker.CheckRequest{Logf: l.Logf},
			},
			want: checker.CheckResult{
				Name:        branchProtectionStr,
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
				c: checker.CheckRequest{Logf: l.Logf},
			},
			want: checker.CheckResult{
				Name:        branchProtectionStr,
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
				c: checker.CheckRequest{Logf: l.Logf},
			},
			want: checker.CheckResult{
				Name:        branchProtectionStr,
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
				c: checker.CheckRequest{Logf: l.Logf},
			},
			want: checker.CheckResult{
				Name:        branchProtectionStr,
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
				c: checker.CheckRequest{Logf: l.Logf},
			},
			want: checker.CheckResult{
				Name:        branchProtectionStr,
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
				c: checker.CheckRequest{Logf: l.Logf},
			},
			want: checker.CheckResult{
				Name:        branchProtectionStr,
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
				c: checker.CheckRequest{Logf: l.Logf},
			},
			want: checker.CheckResult{
				Name:        branchProtectionStr,
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
			got := IsBranchProtected(tt.args.protection, &tt.args.c)
			got.Details = l.messages
			if got.Confidence != tt.want.Confidence || got.Pass != tt.want.Pass {
				t.Errorf("IsBranchProtected() = %s, %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
