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
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/shurcooL/githubv4"

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
			name: "merge queue no bypass",
			base: &clients.BranchRef{},
			ruleSets: []*repoRuleSet{
				ruleSet(withRules(&repoRule{Type: ruleMergeQueue})),
			},
			expected: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions:       &trueVal,
					AllowForcePushes:     &trueVal,
					EnforceAdmins:        &trueVal,
					RequireLinearHistory: &falseVal,
					CheckRules: clients.StatusChecksRule{
						UpToDateBeforeMerge: &trueVal,
					},
					PullRequestRule: clients.PullRequestRule{
						Required: &falseVal,
					},
				},
			},
		},
		{
			name: "merge queue with bypass",
			base: &clients.BranchRef{},
			ruleSets: []*repoRuleSet{
				ruleSet(withRules(&repoRule{Type: ruleMergeQueue}), withBypass()),
			},
			expected: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions:       &trueVal,
					AllowForcePushes:     &trueVal,
					EnforceAdmins:        &falseVal,
					RequireLinearHistory: &falseVal,
					CheckRules: clients.StatusChecksRule{
						UpToDateBeforeMerge: &trueVal,
					},
					PullRequestRule: clients.PullRequestRule{
						Required: &falseVal,
					},
				},
			},
		},
		{
			name: "merge queue overrides non-strict status checks in the same rule set",
			base: &clients.BranchRef{},
			ruleSets: []*repoRuleSet{
				ruleSet(withRules(
					&repoRule{Type: ruleMergeQueue},
					&repoRule{
						Type: ruleStatusCheck,
						Parameters: repoRulesParameters{
							StatusCheckParameters: requiredStatusCheckParameters{
								StrictRequiredStatusChecksPolicy: &falseVal,
								RequiredStatusChecks: []statusCheck{
									{
										Context: asPtr("foo"),
									},
								},
							},
						},
					},
				)),
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
			name: "merge queue rule set and separate non-strict status checks rule set",
			base: &clients.BranchRef{},
			ruleSets: []*repoRuleSet{
				ruleSet(withRules(&repoRule{Type: ruleMergeQueue})),
				ruleSet(withRules(&repoRule{
					Type: ruleStatusCheck,
					Parameters: repoRulesParameters{
						StatusCheckParameters: requiredStatusCheckParameters{
							StrictRequiredStatusChecksPolicy: &falseVal,
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

// branchQueryRoundTripper answers the branchesHandler GraphQL queries
// with canned data, serving refJSON for the per-branch ref query.
type branchQueryRoundTripper struct {
	refJSON string
}

func (b branchQueryRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %w", err)
	}
	var data string
	switch {
	case strings.Contains(string(reqBody), "branchRefName"):
		data = `"repository": {"ref": ` + b.refJSON + `}`
	case strings.Contains(string(reqBody), "rulesets"):
		data = `"repository": {"defaultBranchRef": {"name": "main"}, "rulesets": {"nodes": []}}`
	default:
		data = `"repository": {"defaultBranchRef": {"name": "main", "refUpdateRule": null, "branchProtectionRule": null}}, ` +
			`"rateLimit": {"cost": 1}`
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       io.NopCloser(strings.NewReader(`{"data": {` + data + `}}`)),
		Header:     make(http.Header),
	}, nil
}

func Test_branchesQuery_branchNotFound(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name    string
		refJSON string
		wantErr error
		wantRef string
	}{
		{
			// A deleted/unlocatable branch comes back as a null ref and
			// must surface as ErrBranchNotFound (#2161).
			name:    "null-ref",
			refJSON: `null`,
			wantErr: clients.ErrBranchNotFound,
		},
		{
			name:    "existing-ref",
			refJSON: `{"name": "release-x", "refUpdateRule": null, "branchProtectionRule": null}`,
			wantRef: "release-x",
		},
	}
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()
			handler := branchesHandler{
				graphClient: githubv4.NewClient(&http.Client{
					Transport: branchQueryRoundTripper{refJSON: testcase.refJSON},
				}),
			}
			handler.init(t.Context(), &Repo{owner: "o", repo: "r", commitSHA: clients.HeadSHA})
			ref, err := handler.query("some-branch")
			if !errors.Is(err, testcase.wantErr) {
				t.Fatalf("query(): got err %v, want %v", err, testcase.wantErr)
			}
			if testcase.wantRef == "" {
				return
			}
			if ref == nil || ref.Name == nil || *ref.Name != testcase.wantRef {
				t.Errorf("query(): got %+v, want name %q", ref, testcase.wantRef)
			}
		})
	}
}
