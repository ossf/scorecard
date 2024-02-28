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
			name: "Branch name is an empty string which is not allowed and will error",
			findings: []finding.Finding{
				branchFinding("blocksDeleteOnBranches", "", finding.OutcomePositive),
				branchFinding("blocksForcePushOnBranches", "", finding.OutcomePositive),
				branchFinding("branchesAreProtected", "", finding.OutcomePositive),
				branchFinding("branchProtectionAppliesToAdmins", "", finding.OutcomeNegative),
				branchFinding("dismissesStaleReviews", "", finding.OutcomeNegative),
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNegative,
					Values: map[string]string{
						"branchName":                "",
						"numberOfRequiredReviewers": "0",
					},
				},
				branchFinding("requiresCodeOwnersReview", "", finding.OutcomeNegative),
				branchFinding("requiresLastPushApproval", "", finding.OutcomeNegative),
				branchFinding("requiresUpToDateBranches", "", finding.OutcomePositive),
				branchFinding("runsStatusChecksBeforeMerging", "", finding.OutcomePositive),
				branchFinding("requiresPRsToChangeCode", "", finding.OutcomePositive),
			},
			result: scut.TestReturn{
				Error: sce.ErrScorecardInternal,
				Score: checker.InconclusiveResultScore,
			},
		},
		{
			name: "Required status check enabled",
			findings: []finding.Finding{
				branchFinding("blocksDeleteOnBranches", "main", finding.OutcomePositive),
				branchFinding("blocksForcePushOnBranches", "main", finding.OutcomePositive),
				branchFinding("branchesAreProtected", "main", finding.OutcomePositive),
				branchFinding("branchProtectionAppliesToAdmins", "main", finding.OutcomeNegative),
				branchFinding("dismissesStaleReviews", "main", finding.OutcomeNegative),
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNegative,
					Values: map[string]string{
						"branchName":                "main",
						"numberOfRequiredReviewers": "0",
					},
				},
				branchFinding("requiresCodeOwnersReview", "main", finding.OutcomeNegative),
				branchFinding("requiresLastPushApproval", "main", finding.OutcomeNegative),
				branchFinding("requiresUpToDateBranches", "main", finding.OutcomePositive),
				branchFinding("runsStatusChecksBeforeMerging", "main", finding.OutcomePositive),
				branchFinding("requiresPRsToChangeCode", "main", finding.OutcomePositive),
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
				branchFinding("blocksDeleteOnBranches", "main", finding.OutcomePositive),
				branchFinding("blocksForcePushOnBranches", "main", finding.OutcomePositive),
				branchFinding("branchesAreProtected", "main", finding.OutcomePositive),
				branchFinding("branchProtectionAppliesToAdmins", "main", finding.OutcomeNegative),
				branchFinding("dismissesStaleReviews", "main", finding.OutcomeNegative),
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNegative,
					Values: map[string]string{
						"branchName":                "main",
						"numberOfRequiredReviewers": "0",
					},
				},
				branchFinding("requiresCodeOwnersReview", "main", finding.OutcomeNegative),
				branchFinding("requiresLastPushApproval", "main", finding.OutcomeNegative),
				branchFinding("requiresUpToDateBranches", "main", finding.OutcomePositive),
				branchFinding("runsStatusChecksBeforeMerging", "main", finding.OutcomeNotAvailable),
				branchFinding("requiresPRsToChangeCode", "main", finding.OutcomePositive),
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
				branchFinding("blocksDeleteOnBranches", "main", finding.OutcomePositive),
				branchFinding("blocksForcePushOnBranches", "main", finding.OutcomePositive),
				branchFinding("branchesAreProtected", "main", finding.OutcomePositive),
				branchFinding("branchProtectionAppliesToAdmins", "main", finding.OutcomeNegative),
				branchFinding("dismissesStaleReviews", "main", finding.OutcomeNotAvailable),
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNotAvailable,
					Values: map[string]string{
						"branchName":                "main",
						"numberOfRequiredReviewers": "0",
					},
				},
				branchFinding("requiresCodeOwnersReview", "main", finding.OutcomeNotAvailable),
				branchFinding("requiresLastPushApproval", "main", finding.OutcomeNegative),
				branchFinding("requiresUpToDateBranches", "main", finding.OutcomeNegative),
				branchFinding("runsStatusChecksBeforeMerging", "main", finding.OutcomeNegative),
				branchFinding("requiresPRsToChangeCode", "main", finding.OutcomeNegative),
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
				branchFinding("blocksDeleteOnBranches", "main", finding.OutcomePositive),
				branchFinding("blocksForcePushOnBranches", "main", finding.OutcomePositive),
				branchFinding("branchesAreProtected", "main", finding.OutcomePositive),
				branchFinding("branchProtectionAppliesToAdmins", "main", finding.OutcomePositive),
				branchFinding("dismissesStaleReviews", "main", finding.OutcomeNotAvailable),
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNotAvailable,
					Values: map[string]string{
						"branchName":                "main",
						"numberOfRequiredReviewers": "0",
					},
				},
				branchFinding("requiresCodeOwnersReview", "main", finding.OutcomeNotAvailable),
				branchFinding("requiresLastPushApproval", "main", finding.OutcomePositive),
				branchFinding("requiresUpToDateBranches", "main", finding.OutcomePositive),
				branchFinding("runsStatusChecksBeforeMerging", "main", finding.OutcomePositive),
				branchFinding("requiresPRsToChangeCode", "main", finding.OutcomeNegative),
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
				branchFinding("blocksDeleteOnBranches", "main", finding.OutcomePositive),
				branchFinding("blocksForcePushOnBranches", "main", finding.OutcomePositive),
				branchFinding("branchesAreProtected", "main", finding.OutcomePositive),
				branchFinding("branchProtectionAppliesToAdmins", "main", finding.OutcomePositive),
				branchFinding("dismissesStaleReviews", "main", finding.OutcomePositive),
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNegative,
					Values: map[string]string{
						"branchName":                "main",
						"numberOfRequiredReviewers": "0",
					},
				},
				branchFinding("requiresCodeOwnersReview", "main", finding.OutcomePositive),
				branchFinding("requiresLastPushApproval", "main", finding.OutcomePositive),
				branchFinding("requiresUpToDateBranches", "main", finding.OutcomePositive),
				branchFinding("runsStatusChecksBeforeMerging", "main", finding.OutcomePositive),
				branchFinding("requiresPRsToChangeCode", "main", finding.OutcomePositive),
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
				branchFinding("blocksDeleteOnBranches", "main", finding.OutcomePositive),
				branchFinding("blocksForcePushOnBranches", "main", finding.OutcomePositive),
				branchFinding("branchesAreProtected", "main", finding.OutcomePositive),
				branchFinding("branchProtectionAppliesToAdmins", "main", finding.OutcomePositive),
				branchFinding("dismissesStaleReviews", "main", finding.OutcomeNegative),
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomePositive,
					Values: map[string]string{
						"branchName":                "main",
						"numberOfRequiredReviewers": "1",
					},
				},
				branchFinding("requiresCodeOwnersReview", "main", finding.OutcomeNegative),
				branchFinding("requiresLastPushApproval", "main", finding.OutcomePositive),
				branchFinding("requiresUpToDateBranches", "main", finding.OutcomePositive),
				branchFinding("runsStatusChecksBeforeMerging", "main", finding.OutcomeNegative),
				branchFinding("requiresPRsToChangeCode", "main", finding.OutcomePositive),
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
				branchFinding("blocksDeleteOnBranches", "main", finding.OutcomePositive),
				branchFinding("blocksForcePushOnBranches", "main", finding.OutcomePositive),
				branchFinding("branchesAreProtected", "main", finding.OutcomePositive),
				branchFinding("branchProtectionAppliesToAdmins", "main", finding.OutcomeNotAvailable),
				branchFinding("dismissesStaleReviews", "main", finding.OutcomeNotAvailable),
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNotAvailable,
					Values: map[string]string{
						"branchName":                "main",
						"numberOfRequiredReviewers": "0",
					},
				},
				branchFinding("requiresCodeOwnersReview", "main", finding.OutcomeNotAvailable),
				branchFinding("requiresLastPushApproval", "main", finding.OutcomeNotAvailable),
				branchFinding("requiresUpToDateBranches", "main", finding.OutcomeNotAvailable),
				branchFinding("runsStatusChecksBeforeMerging", "main", finding.OutcomeNotAvailable),
				branchFinding("requiresPRsToChangeCode", "main", finding.OutcomeNotAvailable),
			},
			result: scut.TestReturn{
				Score:         3,
				NumberOfWarn:  2,
				NumberOfInfo:  2,
				NumberOfDebug: 5,
			},
		},
		{
			name: "Non-admin run on project that require 1 reviewer",
			findings: []finding.Finding{
				branchFinding("blocksDeleteOnBranches", "main", finding.OutcomePositive),
				branchFinding("blocksForcePushOnBranches", "main", finding.OutcomePositive),
				branchFinding("branchesAreProtected", "main", finding.OutcomePositive),
				branchFinding("branchProtectionAppliesToAdmins", "main", finding.OutcomeNotAvailable),
				branchFinding("dismissesStaleReviews", "main", finding.OutcomeNotAvailable),
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomePositive,
					Values: map[string]string{
						"branchName":                "main",
						"numberOfRequiredReviewers": "1",
					},
				},
				branchFinding("requiresCodeOwnersReview", "main", finding.OutcomeNegative),
				branchFinding("requiresLastPushApproval", "main", finding.OutcomeNotAvailable),
				branchFinding("requiresUpToDateBranches", "main", finding.OutcomeNotAvailable),
				branchFinding("runsStatusChecksBeforeMerging", "main", finding.OutcomeNotAvailable),
				branchFinding("requiresPRsToChangeCode", "main", finding.OutcomePositive),
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
				branchFinding("blocksDeleteOnBranches", "main", finding.OutcomePositive),
				branchFinding("blocksForcePushOnBranches", "main", finding.OutcomePositive),
				branchFinding("branchesAreProtected", "main", finding.OutcomePositive),
				branchFinding("branchProtectionAppliesToAdmins", "main", finding.OutcomePositive),
				branchFinding("dismissesStaleReviews", "main", finding.OutcomeNegative),
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNegative,
					Values: map[string]string{
						"branchName":                "main",
						"numberOfRequiredReviewers": "0",
					},
				},
				branchFinding("requiresCodeOwnersReview", "main", finding.OutcomeNegative),
				branchFinding("requiresLastPushApproval", "main", finding.OutcomeNegative),
				branchFinding("requiresUpToDateBranches", "main", finding.OutcomeNegative),
				branchFinding("runsStatusChecksBeforeMerging", "main", finding.OutcomePositive),
				branchFinding("requiresPRsToChangeCode", "main", finding.OutcomePositive),
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
				branchFinding("blocksDeleteOnBranches", "main", finding.OutcomePositive),
				branchFinding("blocksForcePushOnBranches", "main", finding.OutcomePositive),
				branchFinding("branchesAreProtected", "main", finding.OutcomePositive),
				branchFinding("branchProtectionAppliesToAdmins", "main", finding.OutcomeNegative),
				branchFinding("dismissesStaleReviews", "main", finding.OutcomeNegative),
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNegative,
					Values: map[string]string{
						"branchName":                "main",
						"numberOfRequiredReviewers": "0",
					},
				},
				branchFinding("requiresCodeOwnersReview", "main", finding.OutcomeNegative),
				branchFinding("requiresLastPushApproval", "main", finding.OutcomeNegative),
				branchFinding("requiresUpToDateBranches", "main", finding.OutcomeNegative),
				branchFinding("runsStatusChecksBeforeMerging", "main", finding.OutcomePositive),
				branchFinding("requiresPRsToChangeCode", "main", finding.OutcomePositive),
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
				branchFinding("blocksDeleteOnBranches", "main", finding.OutcomePositive),
				branchFinding("blocksForcePushOnBranches", "main", finding.OutcomeNegative),
				branchFinding("branchesAreProtected", "main", finding.OutcomePositive),
				branchFinding("branchProtectionAppliesToAdmins", "main", finding.OutcomeNegative),
				branchFinding("dismissesStaleReviews", "main", finding.OutcomeNegative),
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNegative,
					Values: map[string]string{
						"branchName":                "main",
						"numberOfRequiredReviewers": "0",
					},
				},
				branchFinding("requiresCodeOwnersReview", "main", finding.OutcomeNegative),
				branchFinding("requiresLastPushApproval", "main", finding.OutcomeNegative),
				branchFinding("requiresUpToDateBranches", "main", finding.OutcomeNegative),
				branchFinding("runsStatusChecksBeforeMerging", "main", finding.OutcomePositive),
				branchFinding("requiresPRsToChangeCode", "main", finding.OutcomePositive),
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
				branchFinding("blocksDeleteOnBranches", "main", finding.OutcomeNegative),
				branchFinding("blocksForcePushOnBranches", "main", finding.OutcomePositive),
				branchFinding("branchesAreProtected", "main", finding.OutcomePositive),
				branchFinding("branchProtectionAppliesToAdmins", "main", finding.OutcomeNegative),
				branchFinding("dismissesStaleReviews", "main", finding.OutcomeNegative),
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNegative,
					Values: map[string]string{
						"branchName":                "main",
						"numberOfRequiredReviewers": "0",
					},
				},
				branchFinding("requiresCodeOwnersReview", "main", finding.OutcomeNegative),
				branchFinding("requiresLastPushApproval", "main", finding.OutcomeNegative),
				branchFinding("requiresUpToDateBranches", "main", finding.OutcomeNegative),
				branchFinding("runsStatusChecksBeforeMerging", "main", finding.OutcomePositive),
				branchFinding("requiresPRsToChangeCode", "main", finding.OutcomePositive),
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
				branchFinding("blocksDeleteOnBranches", "main", finding.OutcomePositive),
				branchFinding("blocksForcePushOnBranches", "main", finding.OutcomePositive),
				branchFinding("branchesAreProtected", "main", finding.OutcomePositive),
				branchFinding("branchProtectionAppliesToAdmins", "main", finding.OutcomePositive),
				branchFinding("dismissesStaleReviews", "main", finding.OutcomePositive),
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomePositive,
					Values: map[string]string{
						"branchName":                "main",
						"numberOfRequiredReviewers": "1",
					},
				},
				branchFinding("requiresCodeOwnersReview", "main", finding.OutcomePositive),
				branchFinding("requiresLastPushApproval", "main", finding.OutcomePositive),
				branchFinding("requiresUpToDateBranches", "main", finding.OutcomePositive),
				branchFinding("runsStatusChecksBeforeMerging", "main", finding.OutcomePositive),
				branchFinding("requiresPRsToChangeCode", "main", finding.OutcomePositive),
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
				branchFinding("blocksDeleteOnBranches", "main", finding.OutcomePositive),
				branchFinding("blocksForcePushOnBranches", "main", finding.OutcomePositive),
				branchFinding("branchesAreProtected", "main", finding.OutcomePositive),
				branchFinding("branchProtectionAppliesToAdmins", "main", finding.OutcomePositive),
				branchFinding("dismissesStaleReviews", "main", finding.OutcomePositive),
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomePositive,
					Values: map[string]string{
						"branchName":                "main",
						"numberOfRequiredReviewers": "1",
					},
				},
				branchFinding("requiresCodeOwnersReview", "main", finding.OutcomePositive),
				branchFinding("requiresLastPushApproval", "main", finding.OutcomePositive),
				branchFinding("requiresUpToDateBranches", "main", finding.OutcomePositive),
				branchFinding("runsStatusChecksBeforeMerging", "main", finding.OutcomePositive),
				branchFinding("requiresPRsToChangeCode", "main", finding.OutcomePositive),
			},
			result: scut.TestReturn{
				Score:        8,
				NumberOfWarn: 1,
				NumberOfInfo: 9,
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
