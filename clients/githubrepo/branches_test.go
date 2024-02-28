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

	"github.com/ossf/scorecard/v4/clients"
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
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()
			inputRules := []*repoRuleSet{{Enforcement: &active, Conditions: testcase.condition}}
			for branchName, expected := range testcase.defaultBranchNames {
				branchName := branchName
				expected := expected
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
				branchName := branchName
				expected := expected
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
					RequiredPullRequestReviews: clients.PullRequestReviewRule{
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
					RequiredPullRequestReviews: clients.PullRequestReviewRule{
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
					RequiredPullRequestReviews: clients.PullRequestReviewRule{
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
					RequiredPullRequestReviews: clients.PullRequestReviewRule{
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
					RequiredPullRequestReviews: clients.PullRequestReviewRule{
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
					RequiredPullRequestReviews: clients.PullRequestReviewRule{
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
					RequiredPullRequestReviews: clients.PullRequestReviewRule{
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
					RequiredPullRequestReviews: clients.PullRequestReviewRule{
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
					RequiredPullRequestReviews: clients.PullRequestReviewRule{
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
					RequiredPullRequestReviews: clients.PullRequestReviewRule{
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
					RequiredPullRequestReviews: clients.PullRequestReviewRule{
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
					RequiredPullRequestReviews: clients.PullRequestReviewRule{
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
		testcase := testcase
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
					RequiredPullRequestReviews: clients.PullRequestReviewRule{
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
					RequiredPullRequestReviews: clients.PullRequestReviewRule{
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
		testcase := testcase
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
