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
	"github.com/ossf/scorecard/v4/probes/blocksDeleteOnBranches"
	"github.com/ossf/scorecard/v4/probes/blocksForcePushOnBranches"
	"github.com/ossf/scorecard/v4/probes/branchProtectionAppliesToAdmins"
	"github.com/ossf/scorecard/v4/probes/branchesAreProtected"
	"github.com/ossf/scorecard/v4/probes/dismissesStaleReviews"
	"github.com/ossf/scorecard/v4/probes/requiresApproversForPullRequests"
	"github.com/ossf/scorecard/v4/probes/requiresCodeOwnersReview"
	"github.com/ossf/scorecard/v4/probes/requiresLastPushApproval"
	"github.com/ossf/scorecard/v4/probes/requiresPRsToChangeCode"
	"github.com/ossf/scorecard/v4/probes/requiresUpToDateBranches"
	"github.com/ossf/scorecard/v4/probes/runsStatusChecksBeforeMerging"
	scut "github.com/ossf/scorecard/v4/utests"
)

const emptyBranchName = ""

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
				branchFinding(blocksDeleteOnBranches.Probe, emptyBranchName, finding.OutcomePositive),
				branchFinding(blocksForcePushOnBranches.Probe, emptyBranchName, finding.OutcomePositive),
				branchFinding(branchesAreProtected.Probe, emptyBranchName, finding.OutcomePositive),
				branchFinding(branchProtectionAppliesToAdmins.Probe, emptyBranchName, finding.OutcomeNegative),
				branchFinding(dismissesStaleReviews.Probe, emptyBranchName, finding.OutcomeNegative),
				withValue(
					branchFinding(requiresApproversForPullRequests.Probe, emptyBranchName, finding.OutcomeNegative),
					requiresApproversForPullRequests.RequiredReviewersKey, "0",
				),
				branchFinding(requiresCodeOwnersReview.Probe, emptyBranchName, finding.OutcomeNegative),
				branchFinding(requiresLastPushApproval.Probe, emptyBranchName, finding.OutcomeNegative),
				branchFinding(requiresUpToDateBranches.Probe, emptyBranchName, finding.OutcomePositive),
				branchFinding(runsStatusChecksBeforeMerging.Probe, emptyBranchName, finding.OutcomePositive),
				branchFinding(requiresPRsToChangeCode.Probe, emptyBranchName, finding.OutcomePositive),
			},
			result: scut.TestReturn{
				Error: sce.ErrScorecardInternal,
				Score: checker.InconclusiveResultScore,
			},
		},
		{
			name: "Required status check enabled",
			findings: []finding.Finding{
				branchFinding(blocksDeleteOnBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(blocksForcePushOnBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(branchesAreProtected.Probe, "main", finding.OutcomePositive),
				branchFinding(branchProtectionAppliesToAdmins.Probe, "main", finding.OutcomeNegative),
				branchFinding(dismissesStaleReviews.Probe, "main", finding.OutcomeNegative),
				withValue(
					branchFinding(requiresApproversForPullRequests.Probe, "main", finding.OutcomeNegative),
					requiresApproversForPullRequests.RequiredReviewersKey, "0",
				),
				branchFinding(requiresCodeOwnersReview.Probe, "main", finding.OutcomeNegative),
				branchFinding(requiresLastPushApproval.Probe, "main", finding.OutcomeNegative),
				branchFinding(requiresUpToDateBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(runsStatusChecksBeforeMerging.Probe, "main", finding.OutcomePositive),
				branchFinding(requiresPRsToChangeCode.Probe, "main", finding.OutcomePositive),
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
				branchFinding(blocksDeleteOnBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(blocksForcePushOnBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(branchesAreProtected.Probe, "main", finding.OutcomePositive),
				branchFinding(branchProtectionAppliesToAdmins.Probe, "main", finding.OutcomeNegative),
				branchFinding(dismissesStaleReviews.Probe, "main", finding.OutcomeNegative),
				withValue(
					branchFinding(requiresApproversForPullRequests.Probe, "main", finding.OutcomeNegative),
					requiresApproversForPullRequests.RequiredReviewersKey, "0",
				),
				branchFinding(requiresCodeOwnersReview.Probe, "main", finding.OutcomeNegative),
				branchFinding(requiresLastPushApproval.Probe, "main", finding.OutcomeNegative),
				branchFinding(requiresUpToDateBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(runsStatusChecksBeforeMerging.Probe, "main", finding.OutcomeNotAvailable),
				branchFinding(requiresPRsToChangeCode.Probe, "main", finding.OutcomePositive),
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
				branchFinding(blocksDeleteOnBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(blocksForcePushOnBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(branchesAreProtected.Probe, "main", finding.OutcomePositive),
				branchFinding(branchProtectionAppliesToAdmins.Probe, "main", finding.OutcomeNegative),
				branchFinding(dismissesStaleReviews.Probe, "main", finding.OutcomeNotAvailable),
				withValue(
					branchFinding(requiresApproversForPullRequests.Probe, "main", finding.OutcomeNotAvailable),
					requiresApproversForPullRequests.RequiredReviewersKey, "0",
				),
				branchFinding(requiresCodeOwnersReview.Probe, "main", finding.OutcomeNotAvailable),
				branchFinding(requiresLastPushApproval.Probe, "main", finding.OutcomeNegative),
				branchFinding(requiresUpToDateBranches.Probe, "main", finding.OutcomeNegative),
				branchFinding(runsStatusChecksBeforeMerging.Probe, "main", finding.OutcomeNegative),
				branchFinding(requiresPRsToChangeCode.Probe, "main", finding.OutcomeNegative),
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
				branchFinding(blocksDeleteOnBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(blocksForcePushOnBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(branchesAreProtected.Probe, "main", finding.OutcomePositive),
				branchFinding(branchProtectionAppliesToAdmins.Probe, "main", finding.OutcomePositive),
				branchFinding(dismissesStaleReviews.Probe, "main", finding.OutcomeNotAvailable),
				withValue(
					branchFinding(requiresApproversForPullRequests.Probe, "main", finding.OutcomeNotAvailable),
					requiresApproversForPullRequests.RequiredReviewersKey, "0",
				),
				branchFinding(requiresCodeOwnersReview.Probe, "main", finding.OutcomeNotAvailable),
				branchFinding(requiresLastPushApproval.Probe, "main", finding.OutcomePositive),
				branchFinding(requiresUpToDateBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(runsStatusChecksBeforeMerging.Probe, "main", finding.OutcomePositive),
				branchFinding(requiresPRsToChangeCode.Probe, "main", finding.OutcomeNegative),
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
				branchFinding(blocksDeleteOnBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(blocksForcePushOnBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(branchesAreProtected.Probe, "main", finding.OutcomePositive),
				branchFinding(branchProtectionAppliesToAdmins.Probe, "main", finding.OutcomePositive),
				branchFinding(dismissesStaleReviews.Probe, "main", finding.OutcomePositive),
				withValue(
					branchFinding(requiresApproversForPullRequests.Probe, "main", finding.OutcomeNegative),
					requiresApproversForPullRequests.RequiredReviewersKey, "0",
				),
				branchFinding(requiresCodeOwnersReview.Probe, "main", finding.OutcomePositive),
				branchFinding(requiresLastPushApproval.Probe, "main", finding.OutcomePositive),
				branchFinding(requiresUpToDateBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(runsStatusChecksBeforeMerging.Probe, "main", finding.OutcomePositive),
				branchFinding(requiresPRsToChangeCode.Probe, "main", finding.OutcomePositive),
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
				branchFinding(blocksDeleteOnBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(blocksForcePushOnBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(branchesAreProtected.Probe, "main", finding.OutcomePositive),
				branchFinding(branchProtectionAppliesToAdmins.Probe, "main", finding.OutcomePositive),
				branchFinding(dismissesStaleReviews.Probe, "main", finding.OutcomeNegative),
				withValue(
					branchFinding(requiresApproversForPullRequests.Probe, "main", finding.OutcomePositive),
					requiresApproversForPullRequests.RequiredReviewersKey, "1",
				),
				branchFinding(requiresCodeOwnersReview.Probe, "main", finding.OutcomeNegative),
				branchFinding(requiresLastPushApproval.Probe, "main", finding.OutcomePositive),
				branchFinding(requiresUpToDateBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(runsStatusChecksBeforeMerging.Probe, "main", finding.OutcomeNegative),
				branchFinding(requiresPRsToChangeCode.Probe, "main", finding.OutcomePositive),
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
				branchFinding(blocksDeleteOnBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(blocksForcePushOnBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(branchesAreProtected.Probe, "main", finding.OutcomePositive),
				branchFinding(branchProtectionAppliesToAdmins.Probe, "main", finding.OutcomeNotAvailable),
				branchFinding(dismissesStaleReviews.Probe, "main", finding.OutcomeNotAvailable),
				withValue(
					branchFinding(requiresApproversForPullRequests.Probe, "main", finding.OutcomeNotAvailable),
					requiresApproversForPullRequests.RequiredReviewersKey, "0",
				),
				branchFinding(requiresCodeOwnersReview.Probe, "main", finding.OutcomeNotAvailable),
				branchFinding(requiresLastPushApproval.Probe, "main", finding.OutcomeNotAvailable),
				branchFinding(requiresUpToDateBranches.Probe, "main", finding.OutcomeNotAvailable),
				branchFinding(runsStatusChecksBeforeMerging.Probe, "main", finding.OutcomeNotAvailable),
				branchFinding(requiresPRsToChangeCode.Probe, "main", finding.OutcomeNotAvailable),
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
				branchFinding(blocksDeleteOnBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(blocksForcePushOnBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(branchesAreProtected.Probe, "main", finding.OutcomePositive),
				branchFinding(branchProtectionAppliesToAdmins.Probe, "main", finding.OutcomeNotAvailable),
				branchFinding(dismissesStaleReviews.Probe, "main", finding.OutcomeNotAvailable),
				withValue(
					branchFinding(requiresApproversForPullRequests.Probe, "main", finding.OutcomePositive),
					requiresApproversForPullRequests.RequiredReviewersKey, "1",
				),
				branchFinding(requiresCodeOwnersReview.Probe, "main", finding.OutcomeNegative),
				branchFinding(requiresLastPushApproval.Probe, "main", finding.OutcomeNotAvailable),
				branchFinding(requiresUpToDateBranches.Probe, "main", finding.OutcomeNotAvailable),
				branchFinding(runsStatusChecksBeforeMerging.Probe, "main", finding.OutcomeNotAvailable),
				branchFinding(requiresPRsToChangeCode.Probe, "main", finding.OutcomePositive),
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
				branchFinding(blocksDeleteOnBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(blocksForcePushOnBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(branchesAreProtected.Probe, "main", finding.OutcomePositive),
				branchFinding(branchProtectionAppliesToAdmins.Probe, "main", finding.OutcomePositive),
				branchFinding(dismissesStaleReviews.Probe, "main", finding.OutcomeNegative),
				withValue(
					branchFinding(requiresApproversForPullRequests.Probe, "main", finding.OutcomeNegative),
					requiresApproversForPullRequests.RequiredReviewersKey, "0",
				),
				branchFinding(requiresCodeOwnersReview.Probe, "main", finding.OutcomeNegative),
				branchFinding(requiresLastPushApproval.Probe, "main", finding.OutcomeNegative),
				branchFinding(requiresUpToDateBranches.Probe, "main", finding.OutcomeNegative),
				branchFinding(runsStatusChecksBeforeMerging.Probe, "main", finding.OutcomePositive),
				branchFinding(requiresPRsToChangeCode.Probe, "main", finding.OutcomePositive),
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
				branchFinding(blocksDeleteOnBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(blocksForcePushOnBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(branchesAreProtected.Probe, "main", finding.OutcomePositive),
				branchFinding(branchProtectionAppliesToAdmins.Probe, "main", finding.OutcomeNegative),
				branchFinding(dismissesStaleReviews.Probe, "main", finding.OutcomeNegative),
				withValue(
					branchFinding(requiresApproversForPullRequests.Probe, "main", finding.OutcomeNegative),
					requiresApproversForPullRequests.RequiredReviewersKey, "0",
				),
				branchFinding(requiresCodeOwnersReview.Probe, "main", finding.OutcomeNegative),
				branchFinding(requiresLastPushApproval.Probe, "main", finding.OutcomeNegative),
				branchFinding(requiresUpToDateBranches.Probe, "main", finding.OutcomeNegative),
				branchFinding(runsStatusChecksBeforeMerging.Probe, "main", finding.OutcomePositive),
				branchFinding(requiresPRsToChangeCode.Probe, "main", finding.OutcomePositive),
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
				branchFinding(blocksDeleteOnBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(blocksForcePushOnBranches.Probe, "main", finding.OutcomeNegative),
				branchFinding(branchesAreProtected.Probe, "main", finding.OutcomePositive),
				branchFinding(branchProtectionAppliesToAdmins.Probe, "main", finding.OutcomeNegative),
				branchFinding(dismissesStaleReviews.Probe, "main", finding.OutcomeNegative),
				withValue(
					branchFinding(requiresApproversForPullRequests.Probe, "main", finding.OutcomeNegative),
					requiresApproversForPullRequests.RequiredReviewersKey, "0",
				),
				branchFinding(requiresCodeOwnersReview.Probe, "main", finding.OutcomeNegative),
				branchFinding(requiresLastPushApproval.Probe, "main", finding.OutcomeNegative),
				branchFinding(requiresUpToDateBranches.Probe, "main", finding.OutcomeNegative),
				branchFinding(runsStatusChecksBeforeMerging.Probe, "main", finding.OutcomePositive),
				branchFinding(requiresPRsToChangeCode.Probe, "main", finding.OutcomePositive),
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
				branchFinding(blocksDeleteOnBranches.Probe, "main", finding.OutcomeNegative),
				branchFinding(blocksForcePushOnBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(branchesAreProtected.Probe, "main", finding.OutcomePositive),
				branchFinding(branchProtectionAppliesToAdmins.Probe, "main", finding.OutcomeNegative),
				branchFinding(dismissesStaleReviews.Probe, "main", finding.OutcomeNegative),
				withValue(
					branchFinding(requiresApproversForPullRequests.Probe, "main", finding.OutcomeNegative),
					requiresApproversForPullRequests.RequiredReviewersKey, "0",
				),
				branchFinding(requiresCodeOwnersReview.Probe, "main", finding.OutcomeNegative),
				branchFinding(requiresLastPushApproval.Probe, "main", finding.OutcomeNegative),
				branchFinding(requiresUpToDateBranches.Probe, "main", finding.OutcomeNegative),
				branchFinding(runsStatusChecksBeforeMerging.Probe, "main", finding.OutcomePositive),
				branchFinding(requiresPRsToChangeCode.Probe, "main", finding.OutcomePositive),
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
				branchFinding(blocksDeleteOnBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(blocksForcePushOnBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(branchesAreProtected.Probe, "main", finding.OutcomePositive),
				branchFinding(branchProtectionAppliesToAdmins.Probe, "main", finding.OutcomePositive),
				branchFinding(dismissesStaleReviews.Probe, "main", finding.OutcomePositive),
				withValue(
					branchFinding(requiresApproversForPullRequests.Probe, "main", finding.OutcomePositive),
					requiresApproversForPullRequests.RequiredReviewersKey, "1",
				),
				branchFinding(requiresCodeOwnersReview.Probe, "main", finding.OutcomePositive),
				branchFinding(requiresLastPushApproval.Probe, "main", finding.OutcomePositive),
				branchFinding(requiresUpToDateBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(runsStatusChecksBeforeMerging.Probe, "main", finding.OutcomePositive),
				branchFinding(requiresPRsToChangeCode.Probe, "main", finding.OutcomePositive),
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
				branchFinding(blocksDeleteOnBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(blocksForcePushOnBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(branchesAreProtected.Probe, "main", finding.OutcomePositive),
				branchFinding(branchProtectionAppliesToAdmins.Probe, "main", finding.OutcomePositive),
				branchFinding(dismissesStaleReviews.Probe, "main", finding.OutcomePositive),
				withValue(
					branchFinding(requiresApproversForPullRequests.Probe, "main", finding.OutcomePositive),
					requiresApproversForPullRequests.RequiredReviewersKey, "1",
				),
				branchFinding(requiresCodeOwnersReview.Probe, "main", finding.OutcomePositive),
				branchFinding(requiresLastPushApproval.Probe, "main", finding.OutcomePositive),
				branchFinding(requiresUpToDateBranches.Probe, "main", finding.OutcomePositive),
				branchFinding(runsStatusChecksBeforeMerging.Probe, "main", finding.OutcomePositive),
				branchFinding(requiresPRsToChangeCode.Probe, "main", finding.OutcomePositive),
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

// helper function to create findings for branch protection probes.
func branchFinding(probe, branch string, outcome finding.Outcome) finding.Finding {
	return finding.Finding{
		Probe:   probe,
		Outcome: outcome,
		Values: map[string]string{
			"branchName": branch,
		},
	}
}

func withValue(f finding.Finding, k, v string) finding.Finding {
	f.Values[k] = v
	return f
}
