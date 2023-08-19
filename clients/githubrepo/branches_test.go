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
	twoVal := int32(2)

	testcases := []struct {
		base       *clients.BranchRef
		ruleSet    *repoRuleSet
		expected   *clients.BranchRef
		ruleBypass *ruleSetBypass
		name       string
	}{
		{
			name:    "block deletion no bypass",
			base:    &clients.BranchRef{},
			ruleSet: ruleSet(withRules(&repoRule{Type: "DELETION"})),
			expected: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions: &falseVal,
					EnforceAdmins:  &trueVal,
				},
			},
		},
		{
			name:    "block deletion with bypass",
			base:    &clients.BranchRef{},
			ruleSet: ruleSet(withRules(&repoRule{Type: "DELETION"}), withBypass()),
			expected: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions: &falseVal,
					EnforceAdmins:  &falseVal,
				},
			},
		},
		{
			name: "block deletion with bypass when block deletion no bypass",
			base: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions: &falseVal,
					EnforceAdmins:  &trueVal,
				},
			},
			ruleSet: ruleSet(withRules(&repoRule{Type: "DELETION"}), withBypass()),
			expected: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions: &falseVal,
					EnforceAdmins:  &trueVal, // Maintain: base is more strict than rule
				},
			},
		},
		{
			name: "block deletion no bypass when block deletion with bypass",
			base: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions: &falseVal,
					EnforceAdmins:  &falseVal,
				},
			},
			ruleSet: ruleSet(withRules(&repoRule{Type: "DELETION"})),
			expected: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions: &falseVal,
					EnforceAdmins:  &trueVal, // Upgrade: rule is more strict than base
				},
			},
		},
		{
			name: "block deletion no bypass when block force push no bypass",
			base: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowForcePushes: &falseVal,
					EnforceAdmins:    &trueVal,
				},
			},
			ruleSet: ruleSet(withRules(&repoRule{Type: "DELETION"})),
			expected: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions:   &falseVal,
					AllowForcePushes: &falseVal,
					EnforceAdmins:    &trueVal, // Maintain: base and rule are equal strictness
				},
			},
		},
		{
			name: "block deletion and force push no bypass when block deletion with bypass",
			base: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions: &falseVal,
					EnforceAdmins:  &falseVal,
				},
			},
			ruleSet: ruleSet(withRules(&repoRule{Type: "DELETION"}, &repoRule{Type: "NON_FAST_FORWARD"})),
			expected: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions:   &falseVal,
					AllowForcePushes: &falseVal,
					EnforceAdmins:    &trueVal, // Upgrade: rule is more strict than base
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
			ruleSet: ruleSet(withRules(&repoRule{Type: "DELETION"}, &repoRule{Type: "NON_FAST_FORWARD"}), withBypass()),
			expected: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions:   &falseVal,
					AllowForcePushes: &falseVal,
					EnforceAdmins:    &falseVal, // Downgrade: deletion does not enforce admins
				},
			},
		},
		{
			name: "block deletion no bypass while force push is blocked with bypass",
			base: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowForcePushes: &falseVal,
					EnforceAdmins:    &falseVal,
				},
			},
			ruleSet: ruleSet(withRules(&repoRule{Type: "DELETION"})),
			expected: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions:   &falseVal,
					AllowForcePushes: &falseVal,
					EnforceAdmins:    &falseVal, // Maintain: deletion enforces but forcepush does not
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
			ruleSet: ruleSet(withRules(&repoRule{Type: "DELETION"})),
			expected: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions:   &falseVal,
					AllowForcePushes: &falseVal,
					EnforceAdmins:    &trueVal, // Maintain: base and rule are equal strictness
				},
			},
		},
		{
			name:    "block force push no bypass",
			base:    &clients.BranchRef{},
			ruleSet: ruleSet(withRules(&repoRule{Type: "NON_FAST_FORWARD"})),
			expected: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowForcePushes: &falseVal,
					EnforceAdmins:    &trueVal,
				},
			},
		},
		{
			name: "require pull request no bypass",
			base: &clients.BranchRef{},
			ruleSet: ruleSet(withRules(&repoRule{
				Type: "PULL_REQUEST",
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
			expected: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					EnforceAdmins:           &trueVal,
					RequireLastPushApproval: &trueVal,
					RequiredPullRequestReviews: clients.PullRequestReviewRule{
						DismissStaleReviews:          &trueVal,
						RequireCodeOwnerReviews:      &trueVal,
						RequiredApprovingReviewCount: &twoVal,
					},
				},
			},
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()
			applyRepoRules(testcase.base, []*repoRuleSet{testcase.ruleSet})

			if !cmp.Equal(testcase.base, testcase.expected) {
				diff := cmp.Diff(testcase.base, testcase.expected)
				t.Errorf("test failed: expected - %v, got - %v. \n%s", testcase.expected, testcase.base, diff)
			}
		})
	}
}
