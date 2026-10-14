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

package githubrepo

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ossf/scorecard/v5/clients"
)

func Test_rulesMatchingBranch(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name                  string
		defaultBranchNames    map[string]bool
		nonDefaultBranchNames map[string]bool
		condition             ruleSetCondition
	}{
		{
			name: "including all branches",
			condition: ruleSetCondition{
				RefName: ruleSetConditionRefs{
					Include: []string{ruleConditionAllBranches},
				},
			},
			defaultBranchNames: map[string]bool{
				"main": true,
				"foo":  true,
			},
			nonDefaultBranchNames: map[string]bool{
				"main": true,
				"foo":  true,
			},
		},
		{
			name: "empty include and empty exclude applies to no branches",
			condition: ruleSetCondition{
				RefName: ruleSetConditionRefs{},
			},
			defaultBranchNames: map[string]bool{
				"main": false,
				"foo":  false,
			},
			nonDefaultBranchNames: map[string]bool{
				"main": false,
				"foo":  false,
			},
		},
		{
			name: "empty include honors exclude patterns",
			condition: ruleSetCondition{
				RefName: ruleSetConditionRefs{
					Exclude: []string{"refs/heads/foo"},
				},
			},
			defaultBranchNames: map[string]bool{
				"main": true,
				"foo":  false,
			},
			nonDefaultBranchNames: map[string]bool{
				"main": true,
				"foo":  false,
			},
		},
		{
			name: "including default branch",
			condition: ruleSetCondition{
				RefName: ruleSetConditionRefs{
					Include: []string{ruleConditionDefaultBranch},
				},
			},
			defaultBranchNames: map[string]bool{
				"main": true,
				"foo":  true,
			},
			nonDefaultBranchNames: map[string]bool{
				"main": false,
				"foo":  false,
			},
		},
		{
			name: "including branch by name",
			condition: ruleSetCondition{
				RefName: ruleSetConditionRefs{
					Include: []string{"refs/heads/foo"},
				},
			},
			defaultBranchNames: map[string]bool{
				"main": false,
				"foo":  true,
			},
			nonDefaultBranchNames: map[string]bool{
				"main": false,
				"foo":  true,
			},
		},
		{
			name: "including branch by fnmatch",
			condition: ruleSetCondition{
				RefName: ruleSetConditionRefs{
					Include: []string{"refs/heads/foo/**"},
				},
			},
			defaultBranchNames: map[string]bool{
				"main":    false,
				"foo":     false,
				"foo/bar": true,
			},
			nonDefaultBranchNames: map[string]bool{
				"main":    false,
				"foo":     false,
				"foo/bar": true,
			},
		},
		{
			name: "include+exclude branch by fnmatch",
			condition: ruleSetCondition{
				RefName: ruleSetConditionRefs{
					Include: []string{"refs/heads/foo/**"},
					Exclude: []string{"refs/heads/foo/bar"},
				},
			},
			defaultBranchNames: map[string]bool{
				"foo/bar": false,
				"foo/baz": true,
			},
			nonDefaultBranchNames: map[string]bool{
				"foo/bar": false,
				"foo/baz": true,
			},
		},
	}

	active := "ACTIVE"
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()
			inputRules := []*repoRuleSet{{Enforcement: &active, Conditions: testcase.condition}}
			for branchName, expected := range testcase.defaultBranchNames {
				t.Run(fmt.Sprintf("default branch %s", branchName), func(t *testing.T) {
					t.Parallel()
					matching, err := rulesMatchingBranch(inputRules, branchName, true)
					if err != nil {
						t.Fatalf("expected - no error, got: %v", err)
					}
					if matched := len(matching) == 1; matched != expected {
						t.Errorf("expected %v, got %v", expected, matched)
					}
				})
			}
			for branchName, expected := range testcase.nonDefaultBranchNames {
				t.Run(fmt.Sprintf("non-default branch %s", branchName), func(t *testing.T) {
					t.Parallel()
					matching, err := rulesMatchingBranch(inputRules, branchName, false)
					if err != nil {
						t.Fatalf("expected - no error, got: %v", err)
					}
					if matched := len(matching) == 1; matched != expected {
						t.Errorf("expected %v, got %v", expected, matched)
					}
				})
			}
		})
	}
}

type ruleSetOpt func(*repoRuleSet)

func ruleSet(opts ...ruleSetOpt) *repoRuleSet {
	r := &repoRuleSet{}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func withRules(rules ...*repoRule) ruleSetOpt {
	return func(r *repoRuleSet) {
		r.Rules.Nodes = append(r.Rules.Nodes, rules...)
	}
}

func withBypass() ruleSetOpt {
	return func(r *repoRuleSet) {
		r.BypassActors.Nodes = append(r.BypassActors.Nodes, &ruleSetBypass{})
	}
}

func Test_applyRepoRules(t *testing.T) {
	t.Parallel()
	trueVal := true
	falseVal := false
	zeroVal := int32(0)
	twoVal := int32(2)

	testcases := []struct {
		base       *clients.BranchRef
		expected   *clients.BranchRef
		ruleBypass *ruleSetBypass
		name       string
		ruleSets   []*repoRuleSet
	}{
		{
			name: "unchecked checkboxes have consistent values",
			base: &clients.BranchRef{},
			ruleSets: []*repoRuleSet{
				ruleSet(),
			},
			expected: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions:   &trueVal,
					AllowForcePushes: &trueVal,
					CheckRules: clients.StatusChecksRule{
						// nil values mean that the CheckRules checkbox wasn't checked
						UpToDateBeforeMerge:  nil,
						RequiresStatusChecks: nil,
						Contexts:             nil,
					},
					EnforceAdmins:           &trueVal,
					RequireLastPushApproval: nil, // this checkbox is enabled only if require status checks
					RequireLinearHistory:    &falseVal,
					PullRequestRule: clients.PullRequestRule{
						Required: &falseVal,
					},
				},
			},
		},
		{
			name: "block deletion no bypass",
			base: &clients.BranchRef{},
			ruleSets: []*repoRuleSet{
				ruleSet(withRules(&repoRule{Type: ruleDeletion})),
			},
			expected: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions:       &falseVal,
					AllowForcePushes:     &trueVal,
					RequireLinearHistory: &falseVal,
					EnforceAdmins:        &trueVal,
					PullRequestRule: clients.PullRequestRule{
						Required: &falseVal,
					},
				},
			},
		},
		{
			name: "block deletion with bypass",
			base: &clients.BranchRef{},
			ruleSets: []*repoRuleSet{
				ruleSet(withRules(&repoRule{Type: ruleDeletion}), withBypass()),
			},
			expected: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions:       &falseVal,
					AllowForcePushes:     &trueVal,
					RequireLinearHistory: &falseVal,
					EnforceAdmins:        &falseVal,
					PullRequestRule: clients.PullRequestRule{
						Required: &falseVal,
					},
				},
			},
		},
		{
			name: "block deletion and force push with bypass when block force push no bypass",
			base: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowForcePushes: &falseVal,
					EnforceAdmins:    &trueVal,
				},
			},
			ruleSets: []*repoRuleSet{
				ruleSet(withRules(&repoRule{Type: ruleDeletion}, &repoRule{Type: ruleForcePush}), withBypass()),
			},
			expected: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions:       &falseVal,
					AllowForcePushes:     &falseVal,
					EnforceAdmins:        &falseVal, // Downgrade: deletion does not enforce admins
					RequireLinearHistory: &falseVal,
					PullRequestRule: clients.PullRequestRule{
						Required: &falseVal,
					},
				},
			},
		},
		{
			name: "block deletion no bypass while force push is blocked with bypass",
			base: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowForcePushes:     &falseVal,
					EnforceAdmins:        &falseVal,
					RequireLinearHistory: &falseVal,
				},
			},
			ruleSets: []*repoRuleSet{
				ruleSet(withRules(&repoRule{Type: ruleDeletion})),
			},
			expected: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions:       &falseVal,
					AllowForcePushes:     &falseVal,
					EnforceAdmins:        &falseVal, // Maintain: deletion enforces but force-push does not
					RequireLinearHistory: &falseVal,
					PullRequestRule: clients.PullRequestRule{
						Required: &falseVal,
					},
				},
			},
		},
		{
			name: "block deletion no bypass while force push is blocked no bypass",
			base: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowForcePushes: &falseVal,
					EnforceAdmins:    &trueVal,
				},
			},
			ruleSets: []*repoRuleSet{
				ruleSet(withRules(&repoRule{Type: ruleDeletion})),
			},
			expected: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions:       &falseVal,
					AllowForcePushes:     &falseVal,
					EnforceAdmins:        &trueVal, // Maintain: base and rule are equal strictness
					RequireLinearHistory: &falseVal,
					PullRequestRule: clients.PullRequestRule{
						Required: &falseVal,
					},
				},
			},
		},
		{
			name: "block force push no bypass",
			base: &clients.BranchRef{},
			ruleSets: []*repoRuleSet{
				ruleSet(withRules(&repoRule{Type: ruleForcePush})),
			},
			expected: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions:       &trueVal,
					AllowForcePushes:     &falseVal,
					EnforceAdmins:        &trueVal,
					RequireLinearHistory: &falseVal,
					PullRequestRule: clients.PullRequestRule{
						Required: &falseVal,
					},
				},
			},
		},
		{
			name: "require linear history no bypass",
			base: &clients.BranchRef{},
			ruleSets: []*repoRuleSet{
				ruleSet(withRules(&repoRule{Type: ruleLinear})),
			},
			expected: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions:       &trueVal,
					AllowForcePushes:     &trueVal,
					RequireLinearHistory: &trueVal,
					EnforceAdmins:        &trueVal,
					PullRequestRule: clients.PullRequestRule{
						Required: &falseVal,
					},
				},
			},
		},
		{
			name: "require pull request but no reviewers and no bypass",
			base: &clients.BranchRef{},
			ruleSets: []*repoRuleSet{
				ruleSet(withRules(&repoRule{
					Type: rulePullRequest,
					Parameters: repoRulesParameters{
						PullRequestParameters: pullRequestRuleParameters{
							RequireLastPushApproval:      asPtr(true),
							RequiredApprovingReviewCount: &zeroVal,
						},
					},
				})),
			},
			expected: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions:          &trueVal,
					AllowForcePushes:        &trueVal,
					EnforceAdmins:           &trueVal,
					RequireLastPushApproval: &trueVal,
					RequireLinearHistory:    &falseVal,
					PullRequestRule: clients.PullRequestRule{
						Required:                     &trueVal,
						RequiredApprovingReviewCount: &zeroVal,
					},
				},
			},
		},
		{
			name: "require pull request with 2 reviewers no bypass",
			base: &clients.BranchRef{},
			ruleSets: []*repoRuleSet{
				ruleSet(withRules(&repoRule{
					Type: rulePullRequest,
					Parameters: repoRulesParameters{
						PullRequestParameters: pullRequestRuleParameters{
							DismissStaleReviewsOnPush:      &trueVal,
							RequireCodeOwnerReview:         &trueVal,
							RequireLastPushApproval:        &trueVal,
							RequiredApprovingReviewCount:   &twoVal,
							RequiredReviewThreadResolution: &trueVal,
						},
					},
				})),
			},
			expected: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions:          &trueVal,
					AllowForcePushes:        &trueVal,
					EnforceAdmins:           &trueVal,
					RequireLinearHistory:    &falseVal,
					RequireLastPushApproval: &trueVal,
					PullRequestRule: clients.PullRequestRule{
						Required:                     &trueVal,
						DismissStaleReviews:          &trueVal,
						RequireCodeOwnerReviews:      &trueVal,
						RequiredApprovingReviewCount: &twoVal,
					},
				},
			},
		},
		{
			name: "required status checks no bypass",
			base: &clients.BranchRef{},
			ruleSets: []*repoRuleSet{
				ruleSet(withRules(&repoRule{
					Type: ruleStatusCheck,
					Parameters: repoRulesParameters{
						StatusCheckParameters: requiredStatusCheckParameters{
							StrictRequiredStatusChecksPolicy: &trueVal,
							RequiredStatusChecks: []statusCheck{
								{
									Context: asPtr("foo"),
								},
							},
						},
					},
				})),
			},
			expected: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions:       &trueVal,
					AllowForcePushes:     &trueVal,
					EnforceAdmins:        &trueVal,
					RequireLinearHistory: &falseVal,
					CheckRules: clients.StatusChecksRule{
						UpToDateBeforeMerge:  &trueVal,
						RequiresStatusChecks: &trueVal,
						Contexts:             []string{"foo"},
					},
					PullRequestRule: clients.PullRequestRule{
						Required: &falseVal,
					},
				},
			},
		},
		{
			name: "Multiple rules sets impacting a branch",
			base: &clients.BranchRef{},
			ruleSets: []*repoRuleSet{
				ruleSet(withRules( // first a restrictive rule set, let's suppose it was built only for main.
					&repoRule{Type: ruleDeletion},
					&repoRule{Type: ruleForcePush},
					&repoRule{Type: ruleLinear},
					&repoRule{
						Type: ruleStatusCheck,
						Parameters: repoRulesParameters{
							StatusCheckParameters: requiredStatusCheckParameters{
								StrictRequiredStatusChecksPolicy: &trueVal,
								RequiredStatusChecks: []statusCheck{
									{
										Context: asPtr("foo"),
									},
								},
							},
						},
					},
					&repoRule{
						Type: rulePullRequest,
						Parameters: repoRulesParameters{
							PullRequestParameters: pullRequestRuleParameters{
								DismissStaleReviewsOnPush:      &trueVal,
								RequireCodeOwnerReview:         &trueVal,
								RequireLastPushApproval:        &trueVal,
								RequiredApprovingReviewCount:   &twoVal,
								RequiredReviewThreadResolution: &trueVal,
							},
						},
					},
				)),
				ruleSet(withRules( // Then a more permissive rule set, that might be applied to a broader range of branches.
					&repoRule{Type: ruleDeletion},
					&repoRule{
						Type: rulePullRequest,
						Parameters: repoRulesParameters{
							PullRequestParameters: pullRequestRuleParameters{
								DismissStaleReviewsOnPush:      &falseVal,
								RequireCodeOwnerReview:         &falseVal,
								RequireLastPushApproval:        &falseVal,
								RequiredApprovingReviewCount:   &zeroVal,
								RequiredReviewThreadResolution: &falseVal,
							},
						},
					},
				)),
			},
			expected: &clients.BranchRef{ // We expect to see dominance of restrictive rules.
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions:          &falseVal,
					AllowForcePushes:        &falseVal,
					EnforceAdmins:           &trueVal,
					RequireLinearHistory:    &trueVal,
					RequireLastPushApproval: &trueVal,
					CheckRules: clients.StatusChecksRule{
						UpToDateBeforeMerge:  &trueVal,
						RequiresStatusChecks: &trueVal,
						Contexts:             []string{"foo"},
					},
					PullRequestRule: clients.PullRequestRule{
						Required:                     &trueVal,
						RequiredApprovingReviewCount: &twoVal,
						DismissStaleReviews:          &trueVal,
						RequireCodeOwnerReviews:      &trueVal,
					},
				},
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()
			applyRepoRules(testcase.base, testcase.ruleSets)

			if !cmp.Equal(testcase.base, testcase.expected) {
				diff := cmp.Diff(testcase.base, testcase.expected)
				t.Errorf("test failed: expected - %v, got - %v. \n%s", testcase.expected, testcase.base, diff)
			}
		})
	}
}

func Test_translationFromGithubAPIBranchProtectionData(t *testing.T) {
	t.Parallel()
	trueVal := true
	falseVal := false
	zeroVal := int32(0)

	testcases := []struct {
		branch   *branch
		ruleSet  *repoRuleSet
		expected *clients.BranchRef
		name     string
	}{
		{
			name: "Non-admin Branch Protection rule with insufficient data about requiring PRs",
			branch: &branch{
				RefUpdateRule: &refUpdateRule{
					AllowsDeletions:              &falseVal,
					AllowsForcePushes:            &falseVal,
					RequiredApprovingReviewCount: &zeroVal,
					RequiresCodeOwnerReviews:     &falseVal,
					RequiresLinearHistory:        &falseVal,
					RequiredStatusCheckContexts:  nil,
				},
				BranchProtectionRule: nil,
			},
			ruleSet: nil,
			expected: &clients.BranchRef{
				Protected: &trueVal,
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions:       &falseVal,
					AllowForcePushes:     &falseVal,
					RequireLinearHistory: &falseVal,
					CheckRules: clients.StatusChecksRule{
						UpToDateBeforeMerge:  nil,
						RequiresStatusChecks: nil,
						Contexts:             []string{},
					},
					PullRequestRule: clients.PullRequestRule{
						RequiredApprovingReviewCount: asPtr[int32](0),
						RequireCodeOwnerReviews:      &falseVal,
					},
				},
			},
		},
		{
			name: "Admin Branch Protection rule nothing selected",
			branch: &branch{
				BranchProtectionRule: &branchProtectionRule{
					DismissesStaleReviews:        &falseVal,
					IsAdminEnforced:              &falseVal,
					RequiresStrictStatusChecks:   &falseVal,
					RequiresStatusChecks:         &falseVal,
					AllowsDeletions:              &falseVal,
					AllowsForcePushes:            &falseVal,
					RequiredApprovingReviewCount: nil,
					RequiresCodeOwnerReviews:     &falseVal,
					RequiresLinearHistory:        &falseVal,
					RequireLastPushApproval:      &falseVal,
					RequiredStatusCheckContexts:  []string{},
				},
			},
			ruleSet: nil,
			expected: &clients.BranchRef{
				Protected: &trueVal,
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions:          &falseVal,
					AllowForcePushes:        &falseVal,
					EnforceAdmins:           &falseVal,
					RequireLastPushApproval: &falseVal,
					RequireLinearHistory:    &falseVal,
					CheckRules: clients.StatusChecksRule{
						UpToDateBeforeMerge:  &falseVal,
						RequiresStatusChecks: &falseVal,
						Contexts:             []string{},
					},
					PullRequestRule: clients.PullRequestRule{
						Required:                     &falseVal,
						RequireCodeOwnerReviews:      &falseVal,
						DismissStaleReviews:          &falseVal,
						RequiredApprovingReviewCount: nil,
					},
				},
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			var repoRules []*repoRuleSet
			if testcase.ruleSet == nil {
				repoRules = []*repoRuleSet{}
			} else {
				repoRules = []*repoRuleSet{testcase.ruleSet}
			}

			result := getBranchRefFrom(testcase.branch, repoRules)

			if !cmp.Equal(result, testcase.expected) {
				diff := cmp.Diff(result, testcase.expected)
				t.Errorf("test failed: expected - %v, got - %v. \n%s", testcase.expected, result, diff)
			}
		})
	}
}

func TestApplyRepoRulesRedundantBypass(t *testing.T) {
	t.Parallel()

	for _, bypassFirst := range []bool{false, true} {
		t.Run(fmt.Sprintf("bypass_first_%t", bypassFirst), func(t *testing.T) {
			t.Parallel()

			enforced := ruleSet(withRules(&repoRule{Type: ruleDeletion}))
			bypass := ruleSet(
				withRules(&repoRule{Type: ruleDeletion}),
				withBypass(),
			)

			rules := []*repoRuleSet{enforced, bypass}
			if bypassFirst {
				rules = []*repoRuleSet{bypass, enforced}
			}

			got := &clients.BranchRef{}
			applyRepoRules(got, rules)

			want := &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions:       asPtr(false),
					AllowForcePushes:     asPtr(true),
					RequireLinearHistory: asPtr(false),
					EnforceAdmins:        asPtr(true),
					PullRequestRule: clients.PullRequestRule{
						Required: asPtr(false),
					},
				},
			}

			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("branch protection mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestApplyRepoRulesAdditionalBypassProtection(t *testing.T) {
	t.Parallel()

	for _, bypassFirst := range []bool{false, true} {
		t.Run(fmt.Sprintf("bypass_first_%t", bypassFirst), func(t *testing.T) {
			t.Parallel()

			enforced := ruleSet(withRules(&repoRule{Type: ruleDeletion}))
			bypass := ruleSet(
				withRules(&repoRule{Type: ruleForcePush}),
				withBypass(),
			)

			rules := []*repoRuleSet{enforced, bypass}
			if bypassFirst {
				rules = []*repoRuleSet{bypass, enforced}
			}

			got := &clients.BranchRef{}
			applyRepoRules(got, rules)

			want := &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions:       asPtr(false),
					AllowForcePushes:     asPtr(false),
					RequireLinearHistory: asPtr(false),
					EnforceAdmins:        asPtr(false),
					PullRequestRule: clients.PullRequestRule{
						Required: asPtr(false),
					},
				},
			}

			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("branch protection mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestApplyRepoRulesBypassParameters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		enforcedRule *repoRule
		bypassRule   *repoRule
		name         string
		wantEnforced bool
	}{
		{
			name:         "fewer required approvals are covered",
			enforcedRule: reviewCountRule(2),
			bypassRule:   reviewCountRule(1),
			wantEnforced: true,
		},
		{
			name:         "equal required approvals are covered",
			enforcedRule: reviewCountRule(2),
			bypassRule:   reviewCountRule(2),
			wantEnforced: true,
		},
		{
			name:         "more required approvals add protection",
			enforcedRule: reviewCountRule(1),
			bypassRule:   reviewCountRule(2),
			wantEnforced: false,
		},
		{
			name:         "existing status check is covered",
			enforcedRule: statusContextsRule("build", "test"),
			bypassRule:   statusContextsRule("build"),
			wantEnforced: true,
		},
		{
			name:         "different status check adds protection",
			enforcedRule: statusContextsRule("build"),
			bypassRule:   statusContextsRule("test"),
			wantEnforced: false,
		},
		{
			name:         "additional status check adds protection",
			enforcedRule: statusContextsRule("build"),
			bypassRule:   statusContextsRule("build", "test"),
			wantEnforced: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			for _, bypassFirst := range []bool{false, true} {
				t.Run(fmt.Sprintf("bypass_first_%t", bypassFirst), func(t *testing.T) {
					t.Parallel()

					enforced := ruleSet(withRules(tt.enforcedRule))
					bypass := ruleSet(withRules(tt.bypassRule), withBypass())

					rules := []*repoRuleSet{enforced, bypass}
					if bypassFirst {
						rules = []*repoRuleSet{bypass, enforced}
					}

					got := &clients.BranchRef{}
					applyRepoRules(got, rules)

					admins := got.BranchProtectionRule.EnforceAdmins
					if admins == nil {
						t.Fatal("EnforceAdmins is nil")
					}
					if *admins != tt.wantEnforced {
						t.Errorf("EnforceAdmins = %t, want %t",
							*admins, tt.wantEnforced)
					}
				})
			}
		})
	}
}

func reviewCountRule(count int32) *repoRule {
	return &repoRule{
		Type: rulePullRequest,
		Parameters: repoRulesParameters{
			PullRequestParameters: pullRequestRuleParameters{
				RequiredApprovingReviewCount: asPtr(count),
			},
		},
	}
}

func statusContextsRule(contexts ...string) *repoRule {
	rule := &repoRule{Type: ruleStatusCheck}
	for _, context := range contexts {
		rule.Parameters.StatusCheckParameters.RequiredStatusChecks = append(
			rule.Parameters.StatusCheckParameters.RequiredStatusChecks,
			statusCheck{Context: asPtr(context)},
		)
	}
	return rule
}

func TestApplyRepoRulesClassicAdminCoverage(t *testing.T) {
	t.Parallel()

	for _, adminEnforced := range []bool{false, true} {
		t.Run(fmt.Sprintf("classic_admin_enforced_%t", adminEnforced), func(t *testing.T) {
			t.Parallel()

			got := &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions: asPtr(false),
					EnforceAdmins:  asPtr(adminEnforced),
				},
			}
			rules := []*repoRuleSet{
				ruleSet(withRules(&repoRule{Type: ruleDeletion}), withBypass()),
			}

			applyRepoRules(got, rules)

			admins := got.BranchProtectionRule.EnforceAdmins
			if admins == nil {
				t.Fatal("EnforceAdmins is nil")
			}
			if *admins != adminEnforced {
				t.Errorf("EnforceAdmins = %t, want %t", *admins, adminEnforced)
			}
			if valueOrZero(got.BranchProtectionRule.AllowDeletions) {
				t.Error("branch deletion protection was lost")
			}
		})
	}
}

func TestApplyRepoRulesBypassCannotCoverBypass(t *testing.T) {
	t.Parallel()

	got := &clients.BranchRef{}
	rules := []*repoRuleSet{
		ruleSet(withRules(&repoRule{Type: ruleDeletion})),
		ruleSet(withRules(&repoRule{Type: ruleForcePush}), withBypass()),
		ruleSet(withRules(&repoRule{Type: ruleForcePush}), withBypass()),
	}

	applyRepoRules(got, rules)

	admins := got.BranchProtectionRule.EnforceAdmins
	if admins == nil {
		t.Fatal("EnforceAdmins is nil")
	}
	if *admins {
		t.Error("bypassable force-push protection must not enforce admins")
	}
	if valueOrZero(got.BranchProtectionRule.AllowDeletions) ||
		valueOrZero(got.BranchProtectionRule.AllowForcePushes) {
		t.Error("merged branch protections were lost")
	}
}

func TestApplyRepoRulesUnrepresentedBypassProtection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		enforcedRule *repoRule
		bypassRule   *repoRule
		name         string
	}{
		{
			enforcedRule: &repoRule{Type: ruleDeletion},
			bypassRule:   &repoRule{Type: "REQUIRED_SIGNATURES"},
			name:         "untranslated rule",
		},
		{
			enforcedRule: reviewCountRule(2),
			bypassRule: &repoRule{
				Type: rulePullRequest,
				Parameters: repoRulesParameters{
					PullRequestParameters: pullRequestRuleParameters{
						RequiredApprovingReviewCount:   asPtr[int32](2),
						RequiredReviewThreadResolution: asPtr(true),
					},
				},
			},
			name: "review thread resolution",
		},
		{
			enforcedRule: statusContextsRule("build"),
			bypassRule: &repoRule{
				Type: ruleStatusCheck,
				Parameters: repoRulesParameters{
					StatusCheckParameters: requiredStatusCheckParameters{
						RequiredStatusChecks: []statusCheck{
							{
								Context:       asPtr("build"),
								IntegrationID: asPtr[int64](123),
							},
						},
					},
				},
			},
			name: "status check tied to an integration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			for _, bypassFirst := range []bool{false, true} {
				t.Run(fmt.Sprintf("bypass_first_%t", bypassFirst), func(t *testing.T) {
					t.Parallel()

					enforced := ruleSet(withRules(tt.enforcedRule))
					bypass := ruleSet(withRules(tt.bypassRule), withBypass())
					rules := []*repoRuleSet{enforced, bypass}
					if bypassFirst {
						rules = []*repoRuleSet{bypass, enforced}
					}

					got := &clients.BranchRef{}
					applyRepoRules(got, rules)

					admins := got.BranchProtectionRule.EnforceAdmins
					if admins == nil {
						t.Fatal("EnforceAdmins is nil")
					}
					if *admins {
						t.Error("unrepresented bypass protection must not enforce admins")
					}
				})
			}
		})
	}
}

func TestApplyRepoRulesBypassBooleanSettings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		makeRule func(bool) *repoRule
		name     string
	}{
		{
			makeRule: func(enabled bool) *repoRule {
				return &repoRule{
					Type: rulePullRequest,
					Parameters: repoRulesParameters{
						PullRequestParameters: pullRequestRuleParameters{
							DismissStaleReviewsOnPush: asPtr(enabled),
						},
					},
				}
			},
			name: "dismiss stale reviews",
		},
		{
			makeRule: func(enabled bool) *repoRule {
				return &repoRule{
					Type: rulePullRequest,
					Parameters: repoRulesParameters{
						PullRequestParameters: pullRequestRuleParameters{
							RequireCodeOwnerReview: asPtr(enabled),
						},
					},
				}
			},
			name: "code owner reviews",
		},
		{
			makeRule: func(enabled bool) *repoRule {
				return &repoRule{
					Type: rulePullRequest,
					Parameters: repoRulesParameters{
						PullRequestParameters: pullRequestRuleParameters{
							RequireLastPushApproval: asPtr(enabled),
						},
					},
				}
			},
			name: "last push approval",
		},
		{
			makeRule: func(enabled bool) *repoRule {
				rule := statusContextsRule("build")
				rule.Parameters.StatusCheckParameters.StrictRequiredStatusChecksPolicy = asPtr(enabled)
				return rule
			},
			name: "strict status checks",
		},
		{
			makeRule: func(enabled bool) *repoRule {
				ruleTypes := map[bool]string{
					false: ruleDeletion,
					true:  ruleLinear,
				}
				return &repoRule{Type: ruleTypes[enabled]}
			},
			name: "linear history",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			for _, alreadyEnforced := range []bool{false, true} {
				t.Run(fmt.Sprintf("already_enforced_%t", alreadyEnforced), func(t *testing.T) {
					t.Parallel()

					for _, bypassFirst := range []bool{false, true} {
						t.Run(fmt.Sprintf("bypass_first_%t", bypassFirst), func(t *testing.T) {
							t.Parallel()

							enforced := ruleSet(withRules(tt.makeRule(alreadyEnforced)))
							bypass := ruleSet(withRules(tt.makeRule(true)), withBypass())
							rules := []*repoRuleSet{enforced, bypass}
							if bypassFirst {
								rules = []*repoRuleSet{bypass, enforced}
							}

							got := &clients.BranchRef{}
							applyRepoRules(got, rules)

							admins := got.BranchProtectionRule.EnforceAdmins
							if admins == nil {
								t.Fatal("EnforceAdmins is nil")
							}
							if *admins != alreadyEnforced {
								t.Errorf("EnforceAdmins = %t, want %t",
									*admins, alreadyEnforced)
							}
						})
					}
				})
			}
		})
	}
}

func TestApplyRepoRulesCombinedCoverage(t *testing.T) {
	t.Parallel()

	orders := [][]int{
		{0, 1, 2},
		{0, 2, 1},
		{1, 0, 2},
		{1, 2, 0},
		{2, 0, 1},
		{2, 1, 0},
	}

	for _, order := range orders {
		t.Run(fmt.Sprint(order), func(t *testing.T) {
			t.Parallel()

			input := []*repoRuleSet{
				ruleSet(withRules(&repoRule{Type: ruleDeletion})),
				ruleSet(withRules(&repoRule{Type: ruleForcePush})),
				ruleSet(
					withRules(
						&repoRule{Type: ruleDeletion},
						&repoRule{Type: ruleForcePush},
					),
					withBypass(),
				),
			}

			rules := make([]*repoRuleSet, 0, len(order))
			for _, index := range order {
				rules = append(rules, input[index])
			}

			got := &clients.BranchRef{}
			applyRepoRules(got, rules)

			want := &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions:       asPtr(false),
					AllowForcePushes:     asPtr(false),
					RequireLinearHistory: asPtr(false),
					EnforceAdmins:        asPtr(true),
					PullRequestRule: clients.PullRequestRule{
						Required: asPtr(false),
					},
				},
			}

			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("branch protection mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestApplyRepoRulesCoversClassicAdminBypass(t *testing.T) {
	t.Parallel()

	for _, bypassFirst := range []bool{false, true} {
		t.Run(fmt.Sprintf("bypass_first_%t", bypassFirst), func(t *testing.T) {
			t.Parallel()

			got := &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions:       asPtr(true),
					AllowForcePushes:     asPtr(true),
					RequireLinearHistory: asPtr(false),
					EnforceAdmins:        asPtr(false),
					PullRequestRule: clients.PullRequestRule{
						Required:                     asPtr(true),
						RequiredApprovingReviewCount: asPtr[int32](1),
					},
				},
			}

			enforced := ruleSet(withRules(reviewCountRule(2)))
			bypass := ruleSet(withRules(reviewCountRule(1)), withBypass())
			rules := []*repoRuleSet{enforced, bypass}
			if bypassFirst {
				rules = []*repoRuleSet{bypass, enforced}
			}

			applyRepoRules(got, rules)

			want := &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions:       asPtr(true),
					AllowForcePushes:     asPtr(true),
					RequireLinearHistory: asPtr(false),
					EnforceAdmins:        asPtr(true),
					PullRequestRule: clients.PullRequestRule{
						Required:                     asPtr(true),
						RequiredApprovingReviewCount: asPtr[int32](2),
					},
				},
			}

			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("branch protection mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestApplyRepoRulesPartialClassicAdminCoverage(t *testing.T) {
	t.Parallel()

	for _, bypassFirst := range []bool{false, true} {
		t.Run(fmt.Sprintf("bypass_first_%t", bypassFirst), func(t *testing.T) {
			t.Parallel()

			got := &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions:       asPtr(true),
					AllowForcePushes:     asPtr(true),
					RequireLinearHistory: asPtr(false),
					EnforceAdmins:        asPtr(false),
					PullRequestRule: clients.PullRequestRule{
						Required:                     asPtr(true),
						RequiredApprovingReviewCount: asPtr[int32](1),
					},
				},
			}

			enforced := ruleSet(withRules(&repoRule{Type: ruleDeletion}))
			bypass := ruleSet(
				withRules(&repoRule{Type: ruleDeletion}),
				withBypass(),
			)
			rules := []*repoRuleSet{enforced, bypass}
			if bypassFirst {
				rules = []*repoRuleSet{bypass, enforced}
			}

			applyRepoRules(got, rules)

			want := &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions:       asPtr(false),
					AllowForcePushes:     asPtr(true),
					RequireLinearHistory: asPtr(false),
					EnforceAdmins:        asPtr(false),
					PullRequestRule: clients.PullRequestRule{
						Required:                     asPtr(true),
						RequiredApprovingReviewCount: asPtr[int32](1),
					},
				},
			}

			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("branch protection mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestApplyRepoRulesCoveredDetailedParameters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		makeRule func() *repoRule
		name     string
	}{
		{
			makeRule: func() *repoRule {
				rule := reviewCountRule(2)
				rule.Parameters.PullRequestParameters.RequiredReviewThreadResolution = asPtr(true)
				return rule
			},
			name: "review thread resolution",
		},
		{
			makeRule: func() *repoRule {
				rule := statusContextsRule("build")
				rule.Parameters.StatusCheckParameters.RequiredStatusChecks[0].IntegrationID = asPtr[int64](123)
				return rule
			},
			name: "same status check integration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			for _, bypassFirst := range []bool{false, true} {
				t.Run(fmt.Sprintf("bypass_first_%t", bypassFirst), func(t *testing.T) {
					t.Parallel()

					enforced := ruleSet(withRules(tt.makeRule()))
					bypass := ruleSet(withRules(tt.makeRule()), withBypass())
					rules := []*repoRuleSet{enforced, bypass}
					if bypassFirst {
						rules = []*repoRuleSet{bypass, enforced}
					}

					got := &clients.BranchRef{}
					applyRepoRules(got, rules)

					admins := got.BranchProtectionRule.EnforceAdmins
					if admins == nil {
						t.Fatal("EnforceAdmins is nil")
					}
					if !*admins {
						t.Error("covered bypass protection must enforce admins")
					}
				})
			}
		})
	}
}

func TestApplyRepoRulesDifferentStatusCheckIntegration(t *testing.T) {
	t.Parallel()

	for _, bypassFirst := range []bool{false, true} {
		t.Run(fmt.Sprintf("bypass_first_%t", bypassFirst), func(t *testing.T) {
			t.Parallel()

			enforcedRule := statusContextsRule("build")
			enforcedRule.Parameters.StatusCheckParameters.
				RequiredStatusChecks[0].IntegrationID = asPtr[int64](123)

			bypassRule := statusContextsRule("build")
			bypassRule.Parameters.StatusCheckParameters.
				RequiredStatusChecks[0].IntegrationID = asPtr[int64](456)

			enforced := ruleSet(withRules(enforcedRule))
			bypass := ruleSet(withRules(bypassRule), withBypass())
			rules := []*repoRuleSet{enforced, bypass}
			if bypassFirst {
				rules = []*repoRuleSet{bypass, enforced}
			}

			got := &clients.BranchRef{}
			applyRepoRules(got, rules)

			admins := got.BranchProtectionRule.EnforceAdmins
			if admins == nil {
				t.Fatal("EnforceAdmins is nil")
			}
			if *admins {
				t.Error("a different integration must not count as covered")
			}
		})
	}
}

func TestApplyRepoRulesParameterlessCoverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		ruleType string
		covered  bool
	}{
		{ruleType: "CREATION", covered: false},
		{ruleType: "CREATION", covered: true},
		{ruleType: "REQUIRED_SIGNATURES", covered: false},
		{ruleType: "REQUIRED_SIGNATURES", covered: true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s/covered_%t", tt.ruleType, tt.covered), func(t *testing.T) {
			t.Parallel()

			for _, bypassFirst := range []bool{false, true} {
				t.Run(fmt.Sprintf("bypass_first_%t", bypassFirst), func(t *testing.T) {
					t.Parallel()

					enforced := ruleSet(
						withRules(&repoRule{Type: ruleDeletion}),
					)
					if tt.covered {
						enforced.Rules.Nodes = append(
							enforced.Rules.Nodes,
							&repoRule{Type: tt.ruleType},
						)
					}

					bypass := ruleSet(
						withRules(&repoRule{Type: tt.ruleType}),
						withBypass(),
					)
					rules := []*repoRuleSet{enforced, bypass}
					if bypassFirst {
						rules = []*repoRuleSet{bypass, enforced}
					}

					got := &clients.BranchRef{}
					applyRepoRules(got, rules)

					admins := got.BranchProtectionRule.EnforceAdmins
					if admins == nil {
						t.Fatal("EnforceAdmins is nil")
					}
					if *admins != tt.covered {
						t.Errorf("EnforceAdmins = %t, want %t", *admins, tt.covered)
					}
				})
			}
		})
	}
}
