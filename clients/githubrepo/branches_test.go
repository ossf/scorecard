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

func Test_applyRepoRules(t *testing.T) {
	t.Parallel()

	main := "main"
	trueVal := true
	falseVal := false

	twoVal := int32(2)

	testcases := []struct {
		base       *clients.BranchRef
		rule       *repoRule
		expected   *clients.BranchRef
		ruleBypass *ruleSetBypass
		name       string
	}{
		{
			name:     "no repo rules",
			base:     &clients.BranchRef{Name: &main, Protected: &trueVal},
			expected: &clients.BranchRef{Name: &main, Protected: &trueVal},
		},
		{
			name: "rule blocks deletion without bypass",
			base: &clients.BranchRef{},
			rule: &repoRule{Type: "DELETION"},
			expected: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions: &falseVal,
					EnforceAdmins:  &trueVal,
				},
			},
		},
		{
			name:       "rule blocks deletion with bypass",
			base:       &clients.BranchRef{},
			rule:       &repoRule{Type: "DELETION"},
			ruleBypass: &ruleSetBypass{},
			expected: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions: &falseVal,
					EnforceAdmins:  &falseVal,
				},
			},
		},
		{
			name: "rule blocks deletion with bypass, but traditional branch protection blocks without bypass",
			base: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions: &falseVal,
					EnforceAdmins:  &trueVal,
				},
			},
			rule:       &repoRule{Type: "DELETION"},
			ruleBypass: &ruleSetBypass{},
			expected: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions: &falseVal,
					EnforceAdmins:  &trueVal,
				},
			},
		},
		{
			name: "rule blocks deletion without bypass, while traditional branch protection blocks with bypass",
			base: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions: &falseVal,
					EnforceAdmins:  &falseVal,
				},
			},
			rule: &repoRule{Type: "DELETION"},
			expected: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions: &falseVal,
					EnforceAdmins:  &trueVal,
				},
			},
		},
		{
			name: "rule blocks force pushes without bypass",
			base: &clients.BranchRef{},
			rule: &repoRule{Type: "NON_FAST_FORWARD"},
			expected: &clients.BranchRef{
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowForcePushes: &falseVal,
					EnforceAdmins:    &trueVal,
				},
			},
		},
		{
			name: "rule requires pull request without bypass",
			base: &clients.BranchRef{},
			rule: &repoRule{
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
			},
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

			var rules []*repoRuleSet
			if testcase.rule != nil {
				ruleSet := &repoRuleSet{
					Rules: struct{ Nodes []*repoRule }{
						Nodes: []*repoRule{testcase.rule},
					},
				}
				if testcase.ruleBypass != nil {
					ruleSet.BypassActors.Nodes = []*ruleSetBypass{testcase.ruleBypass}
				}
				rules = append(rules, ruleSet)
			}
			applyRepoRules(testcase.base, rules)

			if !cmp.Equal(testcase.base, testcase.expected) {
				diff := cmp.Diff(testcase.base, testcase.expected)
				t.Errorf("test failed: expected - %v, got - %v. \n%s", testcase.expected, testcase.base, diff)
			}
		})
	}
}
