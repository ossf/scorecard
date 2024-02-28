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

const (
	emptyBranchName   = ""
	defaultBranchName = "main"
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
				branchFinding("blocksDeleteOnBranches", emptyBranchName, finding.OutcomePositive),
				branchFinding("blocksForcePushOnBranches", emptyBranchName, finding.OutcomePositive),
				branchFinding("branchesAreProtected", emptyBranchName, finding.OutcomePositive),
				branchFinding("branchProtectionAppliesToAdmins", emptyBranchName, finding.OutcomeNegative),
				branchFinding("dismissesStaleReviews", emptyBranchName, finding.OutcomeNegative),
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNegative,
					Values: map[string]string{
						"branchName":                emptyBranchName,
						"numberOfRequiredReviewers": "0",
					},
				},
				branchFinding("requiresCodeOwnersReview", emptyBranchName, finding.OutcomeNegative),
				branchFinding("requiresLastPushApproval", emptyBranchName, finding.OutcomeNegative),
				branchFinding("requiresUpToDateBranches", emptyBranchName, finding.OutcomePositive),
				branchFinding("runsStatusChecksBeforeMerging", emptyBranchName, finding.OutcomePositive),
				branchFinding("requiresPRsToChangeCode", emptyBranchName, finding.OutcomePositive),
			},
			result: scut.TestReturn{
				Error: sce.ErrScorecardInternal,
				Score: checker.InconclusiveResultScore,
			},
		},
		{
			name: "Required status check enabled",
			findings: []finding.Finding{
				branchFinding("blocksDeleteOnBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("blocksForcePushOnBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("branchesAreProtected", defaultBranchName, finding.OutcomePositive),
				branchFinding("branchProtectionAppliesToAdmins", defaultBranchName, finding.OutcomeNegative),
				branchFinding("dismissesStaleReviews", defaultBranchName, finding.OutcomeNegative),
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNegative,
					Values: map[string]string{
						"branchName":                defaultBranchName,
						"numberOfRequiredReviewers": "0",
					},
				},
				branchFinding("requiresCodeOwnersReview", defaultBranchName, finding.OutcomeNegative),
				branchFinding("requiresLastPushApproval", defaultBranchName, finding.OutcomeNegative),
				branchFinding("requiresUpToDateBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("runsStatusChecksBeforeMerging", defaultBranchName, finding.OutcomePositive),
				branchFinding("requiresPRsToChangeCode", defaultBranchName, finding.OutcomePositive),
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
				branchFinding("blocksDeleteOnBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("blocksForcePushOnBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("branchesAreProtected", defaultBranchName, finding.OutcomePositive),
				branchFinding("branchProtectionAppliesToAdmins", defaultBranchName, finding.OutcomeNegative),
				branchFinding("dismissesStaleReviews", defaultBranchName, finding.OutcomeNegative),
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNegative,
					Values: map[string]string{
						"branchName":                defaultBranchName,
						"numberOfRequiredReviewers": "0",
					},
				},
				branchFinding("requiresCodeOwnersReview", defaultBranchName, finding.OutcomeNegative),
				branchFinding("requiresLastPushApproval", defaultBranchName, finding.OutcomeNegative),
				branchFinding("requiresUpToDateBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("runsStatusChecksBeforeMerging", defaultBranchName, finding.OutcomeNotAvailable),
				branchFinding("requiresPRsToChangeCode", defaultBranchName, finding.OutcomePositive),
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
				branchFinding("blocksDeleteOnBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("blocksForcePushOnBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("branchesAreProtected", defaultBranchName, finding.OutcomePositive),
				branchFinding("branchProtectionAppliesToAdmins", defaultBranchName, finding.OutcomeNegative),
				branchFinding("dismissesStaleReviews", defaultBranchName, finding.OutcomeNotAvailable),
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNotAvailable,
					Values: map[string]string{
						"branchName":                defaultBranchName,
						"numberOfRequiredReviewers": "0",
					},
				},
				branchFinding("requiresCodeOwnersReview", defaultBranchName, finding.OutcomeNotAvailable),
				branchFinding("requiresLastPushApproval", defaultBranchName, finding.OutcomeNegative),
				branchFinding("requiresUpToDateBranches", defaultBranchName, finding.OutcomeNegative),
				branchFinding("runsStatusChecksBeforeMerging", defaultBranchName, finding.OutcomeNegative),
				branchFinding("requiresPRsToChangeCode", defaultBranchName, finding.OutcomeNegative),
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
				branchFinding("blocksDeleteOnBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("blocksForcePushOnBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("branchesAreProtected", defaultBranchName, finding.OutcomePositive),
				branchFinding("branchProtectionAppliesToAdmins", defaultBranchName, finding.OutcomePositive),
				branchFinding("dismissesStaleReviews", defaultBranchName, finding.OutcomeNotAvailable),
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNotAvailable,
					Values: map[string]string{
						"branchName":                defaultBranchName,
						"numberOfRequiredReviewers": "0",
					},
				},
				branchFinding("requiresCodeOwnersReview", defaultBranchName, finding.OutcomeNotAvailable),
				branchFinding("requiresLastPushApproval", defaultBranchName, finding.OutcomePositive),
				branchFinding("requiresUpToDateBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("runsStatusChecksBeforeMerging", defaultBranchName, finding.OutcomePositive),
				branchFinding("requiresPRsToChangeCode", defaultBranchName, finding.OutcomeNegative),
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
				branchFinding("blocksDeleteOnBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("blocksForcePushOnBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("branchesAreProtected", defaultBranchName, finding.OutcomePositive),
				branchFinding("branchProtectionAppliesToAdmins", defaultBranchName, finding.OutcomePositive),
				branchFinding("dismissesStaleReviews", defaultBranchName, finding.OutcomePositive),
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNegative,
					Values: map[string]string{
						"branchName":                defaultBranchName,
						"numberOfRequiredReviewers": "0",
					},
				},
				branchFinding("requiresCodeOwnersReview", defaultBranchName, finding.OutcomePositive),
				branchFinding("requiresLastPushApproval", defaultBranchName, finding.OutcomePositive),
				branchFinding("requiresUpToDateBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("runsStatusChecksBeforeMerging", defaultBranchName, finding.OutcomePositive),
				branchFinding("requiresPRsToChangeCode", defaultBranchName, finding.OutcomePositive),
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
				branchFinding("blocksDeleteOnBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("blocksForcePushOnBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("branchesAreProtected", defaultBranchName, finding.OutcomePositive),
				branchFinding("branchProtectionAppliesToAdmins", defaultBranchName, finding.OutcomePositive),
				branchFinding("dismissesStaleReviews", defaultBranchName, finding.OutcomeNegative),
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomePositive,
					Values: map[string]string{
						"branchName":                defaultBranchName,
						"numberOfRequiredReviewers": "1",
					},
				},
				branchFinding("requiresCodeOwnersReview", defaultBranchName, finding.OutcomeNegative),
				branchFinding("requiresLastPushApproval", defaultBranchName, finding.OutcomePositive),
				branchFinding("requiresUpToDateBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("runsStatusChecksBeforeMerging", defaultBranchName, finding.OutcomeNegative),
				branchFinding("requiresPRsToChangeCode", defaultBranchName, finding.OutcomePositive),
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
				branchFinding("blocksDeleteOnBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("blocksForcePushOnBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("branchesAreProtected", defaultBranchName, finding.OutcomePositive),
				branchFinding("branchProtectionAppliesToAdmins", defaultBranchName, finding.OutcomeNotAvailable),
				branchFinding("dismissesStaleReviews", defaultBranchName, finding.OutcomeNotAvailable),
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNotAvailable,
					Values: map[string]string{
						"branchName":                defaultBranchName,
						"numberOfRequiredReviewers": "0",
					},
				},
				branchFinding("requiresCodeOwnersReview", defaultBranchName, finding.OutcomeNotAvailable),
				branchFinding("requiresLastPushApproval", defaultBranchName, finding.OutcomeNotAvailable),
				branchFinding("requiresUpToDateBranches", defaultBranchName, finding.OutcomeNotAvailable),
				branchFinding("runsStatusChecksBeforeMerging", defaultBranchName, finding.OutcomeNotAvailable),
				branchFinding("requiresPRsToChangeCode", defaultBranchName, finding.OutcomeNotAvailable),
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
				branchFinding("blocksDeleteOnBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("blocksForcePushOnBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("branchesAreProtected", defaultBranchName, finding.OutcomePositive),
				branchFinding("branchProtectionAppliesToAdmins", defaultBranchName, finding.OutcomeNotAvailable),
				branchFinding("dismissesStaleReviews", defaultBranchName, finding.OutcomeNotAvailable),
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomePositive,
					Values: map[string]string{
						"branchName":                defaultBranchName,
						"numberOfRequiredReviewers": "1",
					},
				},
				branchFinding("requiresCodeOwnersReview", defaultBranchName, finding.OutcomeNegative),
				branchFinding("requiresLastPushApproval", defaultBranchName, finding.OutcomeNotAvailable),
				branchFinding("requiresUpToDateBranches", defaultBranchName, finding.OutcomeNotAvailable),
				branchFinding("runsStatusChecksBeforeMerging", defaultBranchName, finding.OutcomeNotAvailable),
				branchFinding("requiresPRsToChangeCode", defaultBranchName, finding.OutcomePositive),
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
				branchFinding("blocksDeleteOnBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("blocksForcePushOnBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("branchesAreProtected", defaultBranchName, finding.OutcomePositive),
				branchFinding("branchProtectionAppliesToAdmins", defaultBranchName, finding.OutcomePositive),
				branchFinding("dismissesStaleReviews", defaultBranchName, finding.OutcomeNegative),
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNegative,
					Values: map[string]string{
						"branchName":                defaultBranchName,
						"numberOfRequiredReviewers": "0",
					},
				},
				branchFinding("requiresCodeOwnersReview", defaultBranchName, finding.OutcomeNegative),
				branchFinding("requiresLastPushApproval", defaultBranchName, finding.OutcomeNegative),
				branchFinding("requiresUpToDateBranches", defaultBranchName, finding.OutcomeNegative),
				branchFinding("runsStatusChecksBeforeMerging", defaultBranchName, finding.OutcomePositive),
				branchFinding("requiresPRsToChangeCode", defaultBranchName, finding.OutcomePositive),
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
				branchFinding("blocksDeleteOnBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("blocksForcePushOnBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("branchesAreProtected", defaultBranchName, finding.OutcomePositive),
				branchFinding("branchProtectionAppliesToAdmins", defaultBranchName, finding.OutcomeNegative),
				branchFinding("dismissesStaleReviews", defaultBranchName, finding.OutcomeNegative),
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNegative,
					Values: map[string]string{
						"branchName":                defaultBranchName,
						"numberOfRequiredReviewers": "0",
					},
				},
				branchFinding("requiresCodeOwnersReview", defaultBranchName, finding.OutcomeNegative),
				branchFinding("requiresLastPushApproval", defaultBranchName, finding.OutcomeNegative),
				branchFinding("requiresUpToDateBranches", defaultBranchName, finding.OutcomeNegative),
				branchFinding("runsStatusChecksBeforeMerging", defaultBranchName, finding.OutcomePositive),
				branchFinding("requiresPRsToChangeCode", defaultBranchName, finding.OutcomePositive),
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
				branchFinding("blocksDeleteOnBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("blocksForcePushOnBranches", defaultBranchName, finding.OutcomeNegative),
				branchFinding("branchesAreProtected", defaultBranchName, finding.OutcomePositive),
				branchFinding("branchProtectionAppliesToAdmins", defaultBranchName, finding.OutcomeNegative),
				branchFinding("dismissesStaleReviews", defaultBranchName, finding.OutcomeNegative),
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNegative,
					Values: map[string]string{
						"branchName":                defaultBranchName,
						"numberOfRequiredReviewers": "0",
					},
				},
				branchFinding("requiresCodeOwnersReview", defaultBranchName, finding.OutcomeNegative),
				branchFinding("requiresLastPushApproval", defaultBranchName, finding.OutcomeNegative),
				branchFinding("requiresUpToDateBranches", defaultBranchName, finding.OutcomeNegative),
				branchFinding("runsStatusChecksBeforeMerging", defaultBranchName, finding.OutcomePositive),
				branchFinding("requiresPRsToChangeCode", defaultBranchName, finding.OutcomePositive),
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
				branchFinding("blocksDeleteOnBranches", defaultBranchName, finding.OutcomeNegative),
				branchFinding("blocksForcePushOnBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("branchesAreProtected", defaultBranchName, finding.OutcomePositive),
				branchFinding("branchProtectionAppliesToAdmins", defaultBranchName, finding.OutcomeNegative),
				branchFinding("dismissesStaleReviews", defaultBranchName, finding.OutcomeNegative),
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomeNegative,
					Values: map[string]string{
						"branchName":                defaultBranchName,
						"numberOfRequiredReviewers": "0",
					},
				},
				branchFinding("requiresCodeOwnersReview", defaultBranchName, finding.OutcomeNegative),
				branchFinding("requiresLastPushApproval", defaultBranchName, finding.OutcomeNegative),
				branchFinding("requiresUpToDateBranches", defaultBranchName, finding.OutcomeNegative),
				branchFinding("runsStatusChecksBeforeMerging", defaultBranchName, finding.OutcomePositive),
				branchFinding("requiresPRsToChangeCode", defaultBranchName, finding.OutcomePositive),
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
				branchFinding("blocksDeleteOnBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("blocksForcePushOnBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("branchesAreProtected", defaultBranchName, finding.OutcomePositive),
				branchFinding("branchProtectionAppliesToAdmins", defaultBranchName, finding.OutcomePositive),
				branchFinding("dismissesStaleReviews", defaultBranchName, finding.OutcomePositive),
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomePositive,
					Values: map[string]string{
						"branchName":                defaultBranchName,
						"numberOfRequiredReviewers": "1",
					},
				},
				branchFinding("requiresCodeOwnersReview", defaultBranchName, finding.OutcomePositive),
				branchFinding("requiresLastPushApproval", defaultBranchName, finding.OutcomePositive),
				branchFinding("requiresUpToDateBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("runsStatusChecksBeforeMerging", defaultBranchName, finding.OutcomePositive),
				branchFinding("requiresPRsToChangeCode", defaultBranchName, finding.OutcomePositive),
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
				branchFinding("blocksDeleteOnBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("blocksForcePushOnBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("branchesAreProtected", defaultBranchName, finding.OutcomePositive),
				branchFinding("branchProtectionAppliesToAdmins", defaultBranchName, finding.OutcomePositive),
				branchFinding("dismissesStaleReviews", defaultBranchName, finding.OutcomePositive),
				{
					Probe:   "requiresApproversForPullRequests",
					Outcome: finding.OutcomePositive,
					Values: map[string]string{
						"branchName":                defaultBranchName,
						"numberOfRequiredReviewers": "1",
					},
				},
				branchFinding("requiresCodeOwnersReview", defaultBranchName, finding.OutcomePositive),
				branchFinding("requiresLastPushApproval", defaultBranchName, finding.OutcomePositive),
				branchFinding("requiresUpToDateBranches", defaultBranchName, finding.OutcomePositive),
				branchFinding("runsStatusChecksBeforeMerging", defaultBranchName, finding.OutcomePositive),
				branchFinding("requiresPRsToChangeCode", defaultBranchName, finding.OutcomePositive),
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
