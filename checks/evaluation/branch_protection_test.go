// Copyright 2020 OpenSSF Scorecard Authors
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

package evaluation

import (
	"testing"

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
	scut "github.com/ossf/scorecard/v4/utests"
)

func TestBranchProtection(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		findings []finding.Finding
		result   scut.TestReturn
	}{
		{
			name: "Required status check enabled",
			findings: []finding.Finding{
				{
					Probe:   "blocksDeleteOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"": 1,
					},
				},
				{
					Probe:   "blocksForcePushOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"": 1,
					},
				},
				{
					Probe:   "branchesAreProtected",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"": 1,
					},
				},
				{
					Probe:   "branchProtectionAppliesToAdmins",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"": 1,
					},
				},
				{
					Probe:   "dismissesStaleReviews",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"": 1,
					},
				},
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"":                          1,
						"numberOfRequiredReviewers": 0,
					},
				},
				{
					Probe:   "requiresCodeOwnersReview",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"": 1,
					},
				},
				{
					Probe:   "requiresLastPushApproval",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"": 1,
					},
				},
				{
					Probe:   "requiresUpToDateBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"": 1,
					},
				},
				{
					Probe:   "runsStatusChecksBeforeMerging",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"": 1,
					},
				},
				{
					Probe:   "requiresPRsToChangeCode",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"": 1,
					},
				},
			},
			result: scut.TestReturn{
				Error: sce.ErrScorecardInternal,
				Score: checker.InconclusiveResultScore,
			},
		},
		{
			name: "Required status check enabled",
			findings: []finding.Finding{
				{
					Probe:   "blocksDeleteOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "blocksForcePushOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchesAreProtected",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchProtectionAppliesToAdmins",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "dismissesStaleReviews",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main":                      1,
						"numberOfRequiredReviewers": 0,
					},
				},
				{
					Probe:   "requiresCodeOwnersReview",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresLastPushApproval",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresUpToDateBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "runsStatusChecksBeforeMerging",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresPRsToChangeCode",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
			},
			result: scut.TestReturn{
				Score:        4,
				NumberOfInfo: 5,
				NumberOfWarn: 5,
			},
		},
		{
			name: "Required status check enabled without checking for status string",
			findings: []finding.Finding{
				{
					Probe:   "blocksDeleteOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "blocksForcePushOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchesAreProtected",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchProtectionAppliesToAdmins",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "dismissesStaleReviews",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main":                      1,
						"numberOfRequiredReviewers": 0,
					},
				},
				{
					Probe:   "requiresCodeOwnersReview",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresLastPushApproval",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresUpToDateBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "runsStatusChecksBeforeMerging",
					Outcome: finding.OutcomeNotAvailable,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresPRsToChangeCode",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
			},
			result: scut.TestReturn{
				Score:        4,
				NumberOfInfo: 4,
				NumberOfWarn: 6,
			},
		},
		{
			name: "Admin run only preventing force pushes and deletions",
			findings: []finding.Finding{
				{
					Probe:   "blocksDeleteOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "blocksForcePushOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchesAreProtected",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchProtectionAppliesToAdmins",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "dismissesStaleReviews",
					Outcome: finding.OutcomeNotAvailable,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNotAvailable,
					Values: map[string]int{
						"main":                      1,
						"numberOfRequiredReviewers": 0,
					},
				},
				{
					Probe:   "requiresCodeOwnersReview",
					Outcome: finding.OutcomeNotAvailable,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresLastPushApproval",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresUpToDateBranches",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "runsStatusChecksBeforeMerging",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresPRsToChangeCode",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
			},
			result: scut.TestReturn{
				Score:         3,
				NumberOfWarn:  6,
				NumberOfInfo:  2,
				NumberOfDebug: 1,
			},
		},
		{
			name: "Admin run with all tier 2 requirements except require PRs and reviewers",
			findings: []finding.Finding{
				{
					Probe:   "blocksDeleteOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "blocksForcePushOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchesAreProtected",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchProtectionAppliesToAdmins",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "dismissesStaleReviews",
					Outcome: finding.OutcomeNotAvailable,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNotAvailable,
					Values: map[string]int{
						"main":                      1,
						"numberOfRequiredReviewers": 0,
					},
				},
				{
					Probe:   "requiresCodeOwnersReview",
					Outcome: finding.OutcomeNotAvailable,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresLastPushApproval",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresUpToDateBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "runsStatusChecksBeforeMerging",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresPRsToChangeCode",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
			},
			result: scut.TestReturn{
				Score:         4,
				NumberOfWarn:  2,
				NumberOfInfo:  6,
				NumberOfDebug: 1,
			},
		},
		{
			name: "Admin run on project requiring pull requests but without approver -- best a single maintainer can do",
			findings: []finding.Finding{
				{
					Probe:   "blocksDeleteOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "blocksForcePushOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchesAreProtected",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchProtectionAppliesToAdmins",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "dismissesStaleReviews",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main":                      1,
						"numberOfRequiredReviewers": 0,
					},
				},
				{
					Probe:   "requiresCodeOwnersReview",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresLastPushApproval",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresUpToDateBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "runsStatusChecksBeforeMerging",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresPRsToChangeCode",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
			},
			result: scut.TestReturn{
				Score:        4,
				NumberOfWarn: 1,
				NumberOfInfo: 9,
			},
		},
		{
			name: "Admin run on project with all tier 2 requirements",
			findings: []finding.Finding{
				{
					Probe:   "blocksDeleteOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "blocksForcePushOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchesAreProtected",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchProtectionAppliesToAdmins",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "dismissesStaleReviews",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main":                      1,
						"numberOfRequiredReviewers": 1,
					},
				},
				{
					Probe:   "requiresCodeOwnersReview",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresLastPushApproval",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresUpToDateBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "runsStatusChecksBeforeMerging",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresPRsToChangeCode",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
			},
			result: scut.TestReturn{
				Score:        6,
				NumberOfWarn: 4,
				NumberOfInfo: 6,
			},
		},
		{
			name: "Non-admin run on project that require zero reviewer (or don't require PRs at all, we can't differentiate it)",
			findings: []finding.Finding{
				{
					Probe:   "blocksDeleteOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "blocksForcePushOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchesAreProtected",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchProtectionAppliesToAdmins",
					Outcome: finding.OutcomeNotAvailable,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "dismissesStaleReviews",
					Outcome: finding.OutcomeNotAvailable,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNotAvailable,
					Values: map[string]int{
						"main":                      1,
						"numberOfRequiredReviewers": 0,
					},
				},
				{
					Probe:   "requiresCodeOwnersReview",
					Outcome: finding.OutcomeNotAvailable,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresLastPushApproval",
					Outcome: finding.OutcomeNotAvailable,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresUpToDateBranches",
					Outcome: finding.OutcomeNotAvailable,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "runsStatusChecksBeforeMerging",
					Outcome: finding.OutcomeNotAvailable,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresPRsToChangeCode",
					Outcome: finding.OutcomeNotAvailable,
					Values: map[string]int{
						"main": 1,
					},
				},
			},
			result: scut.TestReturn{
				Score:         3,
				NumberOfWarn:  3,
				NumberOfInfo:  2,
				NumberOfDebug: 4,
			},
		},
		{
			name: "Non-admin run on project that require 1 reviewer",
			findings: []finding.Finding{
				{
					Probe:   "blocksDeleteOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "blocksForcePushOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchesAreProtected",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchProtectionAppliesToAdmins",
					Outcome: finding.OutcomeNotAvailable,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "dismissesStaleReviews",
					Outcome: finding.OutcomeNotAvailable,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main":                      1,
						"numberOfRequiredReviewers": 1,
					},
				},
				{
					Probe:   "requiresCodeOwnersReview",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresLastPushApproval",
					Outcome: finding.OutcomeNotAvailable,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresUpToDateBranches",
					Outcome: finding.OutcomeNotAvailable,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "runsStatusChecksBeforeMerging",
					Outcome: finding.OutcomeNotAvailable,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresPRsToChangeCode",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
			},
			result: scut.TestReturn{
				Score:         6,
				NumberOfWarn:  3,
				NumberOfInfo:  3,
				NumberOfDebug: 4,
			},
		},
		{
			name: "Required admin enforcement enabled",
			findings: []finding.Finding{
				{
					Probe:   "blocksDeleteOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "blocksForcePushOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchesAreProtected",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchProtectionAppliesToAdmins",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "dismissesStaleReviews",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main":                      1,
						"numberOfRequiredReviewers": 0,
					},
				},
				{
					Probe:   "requiresCodeOwnersReview",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresLastPushApproval",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresUpToDateBranches",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "runsStatusChecksBeforeMerging",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresPRsToChangeCode",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
			},
			result: scut.TestReturn{
				Score:        3,
				NumberOfWarn: 5,
				NumberOfInfo: 5,
			},
		},
		{
			name: "Required linear history enabled",
			findings: []finding.Finding{
				{
					Probe:   "blocksDeleteOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "blocksForcePushOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchesAreProtected",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchProtectionAppliesToAdmins",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "dismissesStaleReviews",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main":                      1,
						"numberOfRequiredReviewers": 0,
					},
				},
				{
					Probe:   "requiresCodeOwnersReview",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresLastPushApproval",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresUpToDateBranches",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "runsStatusChecksBeforeMerging",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresPRsToChangeCode",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
			},
			result: scut.TestReturn{
				Score:        3,
				NumberOfWarn: 6,
				NumberOfInfo: 4,
			},
		},
		{
			name: "Allow force push enabled",
			findings: []finding.Finding{
				{
					Probe:   "blocksDeleteOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "blocksForcePushOnBranches",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchesAreProtected",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchProtectionAppliesToAdmins",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "dismissesStaleReviews",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main":                      1,
						"numberOfRequiredReviewers": 0,
					},
				},
				{
					Probe:   "requiresCodeOwnersReview",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresLastPushApproval",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresUpToDateBranches",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "runsStatusChecksBeforeMerging",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresPRsToChangeCode",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
			},
			result: scut.TestReturn{
				Score:        1,
				NumberOfWarn: 7,
				NumberOfInfo: 3,
			},
		},
		{
			name: "Allow deletions enabled",
			findings: []finding.Finding{
				{
					Probe:   "blocksDeleteOnBranches",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "blocksForcePushOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchesAreProtected",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchProtectionAppliesToAdmins",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "dismissesStaleReviews",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main":                      1,
						"numberOfRequiredReviewers": 0,
					},
				},
				{
					Probe:   "requiresCodeOwnersReview",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresLastPushApproval",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresUpToDateBranches",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "runsStatusChecksBeforeMerging",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresPRsToChangeCode",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
			},
			result: scut.TestReturn{
				Score:        1,
				NumberOfWarn: 7,
				NumberOfInfo: 3,
			},
		},
		{
			name: "Branches are protected",
			findings: []finding.Finding{
				{
					Probe:   "blocksDeleteOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "blocksForcePushOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchesAreProtected",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchProtectionAppliesToAdmins",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "dismissesStaleReviews",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main":                      1,
						"numberOfRequiredReviewers": 1,
					},
				},
				{
					Probe:   "requiresCodeOwnersReview",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresLastPushApproval",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresUpToDateBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "runsStatusChecksBeforeMerging",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresPRsToChangeCode",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
			},
			result: scut.TestReturn{
				Score:        8,
				NumberOfWarn: 1,
				NumberOfInfo: 9,
			},
		},
		{
			name: "Branches are protected and require codeowner review",
			findings: []finding.Finding{
				{
					Probe:   "blocksDeleteOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "blocksForcePushOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchesAreProtected",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchProtectionAppliesToAdmins",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "dismissesStaleReviews",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main":                      1,
						"numberOfRequiredReviewers": 1,
					},
				},
				{
					Probe:   "requiresCodeOwnersReview",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresLastPushApproval",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresUpToDateBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "runsStatusChecksBeforeMerging",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresPRsToChangeCode",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
			},
			result: scut.TestReturn{
				Score:        8,
				NumberOfWarn: 1,
				NumberOfInfo: 9,
			},
		},
		{
			name: "Branches are protected and require codeowner review, but file is not present",
			findings: []finding.Finding{
				{
					Probe:   "blocksDeleteOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "blocksForcePushOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchesAreProtected",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchProtectionAppliesToAdmins",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "dismissesStaleReviews",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main":                      1,
						"numberOfRequiredReviewers": 1,
					},
				},
				{
					Probe:   "requiresCodeOwnersReview",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresLastPushApproval",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresUpToDateBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "runsStatusChecksBeforeMerging",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresPRsToChangeCode",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
			},
			result: scut.TestReturn{
				Score:        5,
				NumberOfWarn: 2,
				NumberOfInfo: 8,
			},
		},
		{
			name: "2 branches, one is protected, one is not",
			findings: []finding.Finding{
				{
					Probe:   "blocksDeleteOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "blocksForcePushOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchesAreProtected",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchProtectionAppliesToAdmins",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "dismissesStaleReviews",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main":                      1,
						"numberOfRequiredReviewers": 1,
					},
				},
				{
					Probe:   "requiresCodeOwnersReview",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresLastPushApproval",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresUpToDateBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "runsStatusChecksBeforeMerging",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresPRsToChangeCode",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main2": 1,
					},
				},
				{
					Probe:   "blocksDeleteOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main2": 1,
					},
				},
				{
					Probe:   "blocksForcePushOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main2": 1,
					},
				},
				{
					Probe:   "branchesAreProtected",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main2": 1,
					},
				},
				{
					Probe:   "branchProtectionAppliesToAdmins",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main2": 1,
					},
				},
				{
					Probe:   "dismissesStaleReviews",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main2": 1,
					},
				},
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main2":                     1,
						"numberOfRequiredReviewers": 1,
					},
				},
				{
					Probe:   "requiresCodeOwnersReview",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main2": 1,
					},
				},
				{
					Probe:   "requiresLastPushApproval",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main2": 1,
					},
				},
				{
					Probe:   "requiresUpToDateBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main2": 1,
					},
				},
				{
					Probe:   "runsStatusChecksBeforeMerging",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main2": 1,
					},
				},
				{
					Probe:   "requiresPRsToChangeCode",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main2": 1,
					},
				},
			},
			result: scut.TestReturn{
				Score:        8,
				NumberOfWarn: 3,
				NumberOfInfo: 7,
			},
		},
		{
			name: "1 branch that is not protected",
			findings: []finding.Finding{
				{
					Probe:   "blocksDeleteOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "blocksForcePushOnBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchesAreProtected",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "branchProtectionAppliesToAdmins",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "dismissesStaleReviews",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main":                      1,
						"numberOfRequiredReviewers": 1,
					},
				},
				{
					Probe:   "requiresCodeOwnersReview",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresLastPushApproval",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresUpToDateBranches",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "runsStatusChecksBeforeMerging",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
				{
					Probe:   "requiresPRsToChangeCode",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"main": 1,
					},
				},
			},
			result: scut.TestReturn{
				Score:        5,
				NumberOfWarn: 1,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			got := BranchProtection(tt.name, tt.findings, &dl)
			scut.ValidateTestReturn(t, tt.name, &tt.result, &got, &dl)
		})
	}
}
