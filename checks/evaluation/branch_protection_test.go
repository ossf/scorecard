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

	"github.com/ossf/scorecard/v5/checker"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/blocksDeleteOnBranches"
	"github.com/ossf/scorecard/v5/probes/blocksForcePushOnBranches"
	"github.com/ossf/scorecard/v5/probes/branchProtectionAppliesToAdmins"
	"github.com/ossf/scorecard/v5/probes/branchesAreProtected"
	"github.com/ossf/scorecard/v5/probes/dismissesStaleReviews"
	"github.com/ossf/scorecard/v5/probes/requiresApproversForPullRequests"
	"github.com/ossf/scorecard/v5/probes/requiresCodeOwnersReview"
	"github.com/ossf/scorecard/v5/probes/requiresLastPushApproval"
	"github.com/ossf/scorecard/v5/probes/requiresPRsToChangeCode"
	"github.com/ossf/scorecard/v5/probes/requiresUpToDateBranches"
	"github.com/ossf/scorecard/v5/probes/runsStatusChecksBeforeMerging"
	scut "github.com/ossf/scorecard/v5/utests"
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
				branchFinding(blocksDeleteOnBranches.Probe, emptyBranchName, finding.OutcomeTrue),
				branchFinding(blocksForcePushOnBranches.Probe, emptyBranchName, finding.OutcomeTrue),
				branchFinding(branchesAreProtected.Probe, emptyBranchName, finding.OutcomeTrue),
				branchFinding(branchProtectionAppliesToAdmins.Probe, emptyBranchName, finding.OutcomeFalse),
				branchFinding(dismissesStaleReviews.Probe, emptyBranchName, finding.OutcomeFalse),
				withValue(
					branchFinding(requiresApproversForPullRequests.Probe, emptyBranchName, finding.OutcomeFalse),
					requiresApproversForPullRequests.RequiredReviewersKey, "0",
				),
				branchFinding(requiresCodeOwnersReview.Probe, emptyBranchName, finding.OutcomeFalse),
				branchFinding(requiresLastPushApproval.Probe, emptyBranchName, finding.OutcomeFalse),
				branchFinding(requiresUpToDateBranches.Probe, emptyBranchName, finding.OutcomeTrue),
				branchFinding(runsStatusChecksBeforeMerging.Probe, emptyBranchName, finding.OutcomeTrue),
				branchFinding(requiresPRsToChangeCode.Probe, emptyBranchName, finding.OutcomeTrue),
			},
			result: scut.TestReturn{
				Error: sce.ErrScorecardInternal,
				Score: checker.InconclusiveResultScore,
			},
		},
		{
			name: "Required status check enabled",
			findings: []finding.Finding{
				branchFinding(blocksDeleteOnBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(blocksForcePushOnBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(branchesAreProtected.Probe, "main", finding.OutcomeTrue),
				branchFinding(branchProtectionAppliesToAdmins.Probe, "main", finding.OutcomeFalse),
				branchFinding(dismissesStaleReviews.Probe, "main", finding.OutcomeFalse),
				withValue(
					branchFinding(requiresApproversForPullRequests.Probe, "main", finding.OutcomeFalse),
					requiresApproversForPullRequests.RequiredReviewersKey, "0",
				),
				branchFinding(requiresCodeOwnersReview.Probe, "main", finding.OutcomeFalse),
				branchFinding(requiresLastPushApproval.Probe, "main", finding.OutcomeFalse),
				branchFinding(requiresUpToDateBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(runsStatusChecksBeforeMerging.Probe, "main", finding.OutcomeTrue),
				branchFinding(requiresPRsToChangeCode.Probe, "main", finding.OutcomeTrue),
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
				branchFinding(blocksDeleteOnBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(blocksForcePushOnBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(branchesAreProtected.Probe, "main", finding.OutcomeTrue),
				branchFinding(branchProtectionAppliesToAdmins.Probe, "main", finding.OutcomeFalse),
				branchFinding(dismissesStaleReviews.Probe, "main", finding.OutcomeFalse),
				withValue(
					branchFinding(requiresApproversForPullRequests.Probe, "main", finding.OutcomeFalse),
					requiresApproversForPullRequests.RequiredReviewersKey, "0",
				),
				branchFinding(requiresCodeOwnersReview.Probe, "main", finding.OutcomeFalse),
				branchFinding(requiresLastPushApproval.Probe, "main", finding.OutcomeFalse),
				branchFinding(requiresUpToDateBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(runsStatusChecksBeforeMerging.Probe, "main", finding.OutcomeNotAvailable),
				branchFinding(requiresPRsToChangeCode.Probe, "main", finding.OutcomeTrue),
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
				branchFinding(blocksDeleteOnBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(blocksForcePushOnBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(branchesAreProtected.Probe, "main", finding.OutcomeTrue),
				branchFinding(branchProtectionAppliesToAdmins.Probe, "main", finding.OutcomeFalse),
				branchFinding(dismissesStaleReviews.Probe, "main", finding.OutcomeNotAvailable),
				withValue(
					branchFinding(requiresApproversForPullRequests.Probe, "main", finding.OutcomeNotAvailable),
					requiresApproversForPullRequests.RequiredReviewersKey, "0",
				),
				branchFinding(requiresCodeOwnersReview.Probe, "main", finding.OutcomeNotAvailable),
				branchFinding(requiresLastPushApproval.Probe, "main", finding.OutcomeFalse),
				branchFinding(requiresUpToDateBranches.Probe, "main", finding.OutcomeFalse),
				branchFinding(runsStatusChecksBeforeMerging.Probe, "main", finding.OutcomeFalse),
				branchFinding(requiresPRsToChangeCode.Probe, "main", finding.OutcomeFalse),
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
				branchFinding(blocksDeleteOnBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(blocksForcePushOnBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(branchesAreProtected.Probe, "main", finding.OutcomeTrue),
				branchFinding(branchProtectionAppliesToAdmins.Probe, "main", finding.OutcomeTrue),
				branchFinding(dismissesStaleReviews.Probe, "main", finding.OutcomeNotAvailable),
				withValue(
					branchFinding(requiresApproversForPullRequests.Probe, "main", finding.OutcomeNotAvailable),
					requiresApproversForPullRequests.RequiredReviewersKey, "0",
				),
				branchFinding(requiresCodeOwnersReview.Probe, "main", finding.OutcomeNotAvailable),
				branchFinding(requiresLastPushApproval.Probe, "main", finding.OutcomeTrue),
				branchFinding(requiresUpToDateBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(runsStatusChecksBeforeMerging.Probe, "main", finding.OutcomeTrue),
				branchFinding(requiresPRsToChangeCode.Probe, "main", finding.OutcomeFalse),
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
				branchFinding(blocksDeleteOnBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(blocksForcePushOnBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(branchesAreProtected.Probe, "main", finding.OutcomeTrue),
				branchFinding(branchProtectionAppliesToAdmins.Probe, "main", finding.OutcomeTrue),
				branchFinding(dismissesStaleReviews.Probe, "main", finding.OutcomeTrue),
				withValue(
					branchFinding(requiresApproversForPullRequests.Probe, "main", finding.OutcomeFalse),
					requiresApproversForPullRequests.RequiredReviewersKey, "0",
				),
				branchFinding(requiresCodeOwnersReview.Probe, "main", finding.OutcomeTrue),
				branchFinding(requiresLastPushApproval.Probe, "main", finding.OutcomeTrue),
				branchFinding(requiresUpToDateBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(runsStatusChecksBeforeMerging.Probe, "main", finding.OutcomeTrue),
				branchFinding(requiresPRsToChangeCode.Probe, "main", finding.OutcomeTrue),
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
				branchFinding(blocksDeleteOnBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(blocksForcePushOnBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(branchesAreProtected.Probe, "main", finding.OutcomeTrue),
				branchFinding(branchProtectionAppliesToAdmins.Probe, "main", finding.OutcomeTrue),
				branchFinding(dismissesStaleReviews.Probe, "main", finding.OutcomeFalse),
				withValue(
					branchFinding(requiresApproversForPullRequests.Probe, "main", finding.OutcomeTrue),
					requiresApproversForPullRequests.RequiredReviewersKey, "1",
				),
				branchFinding(requiresCodeOwnersReview.Probe, "main", finding.OutcomeFalse),
				branchFinding(requiresLastPushApproval.Probe, "main", finding.OutcomeTrue),
				branchFinding(requiresUpToDateBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(runsStatusChecksBeforeMerging.Probe, "main", finding.OutcomeFalse),
				branchFinding(requiresPRsToChangeCode.Probe, "main", finding.OutcomeTrue),
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
				branchFinding(blocksDeleteOnBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(blocksForcePushOnBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(branchesAreProtected.Probe, "main", finding.OutcomeTrue),
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
				branchFinding(blocksDeleteOnBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(blocksForcePushOnBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(branchesAreProtected.Probe, "main", finding.OutcomeTrue),
				branchFinding(branchProtectionAppliesToAdmins.Probe, "main", finding.OutcomeNotAvailable),
				branchFinding(dismissesStaleReviews.Probe, "main", finding.OutcomeNotAvailable),
				withValue(
					branchFinding(requiresApproversForPullRequests.Probe, "main", finding.OutcomeTrue),
					requiresApproversForPullRequests.RequiredReviewersKey, "1",
				),
				branchFinding(requiresCodeOwnersReview.Probe, "main", finding.OutcomeFalse),
				branchFinding(requiresLastPushApproval.Probe, "main", finding.OutcomeNotAvailable),
				branchFinding(requiresUpToDateBranches.Probe, "main", finding.OutcomeNotAvailable),
				branchFinding(runsStatusChecksBeforeMerging.Probe, "main", finding.OutcomeNotAvailable),
				branchFinding(requiresPRsToChangeCode.Probe, "main", finding.OutcomeTrue),
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
				branchFinding(blocksDeleteOnBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(blocksForcePushOnBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(branchesAreProtected.Probe, "main", finding.OutcomeTrue),
				branchFinding(branchProtectionAppliesToAdmins.Probe, "main", finding.OutcomeTrue),
				branchFinding(dismissesStaleReviews.Probe, "main", finding.OutcomeFalse),
				withValue(
					branchFinding(requiresApproversForPullRequests.Probe, "main", finding.OutcomeFalse),
					requiresApproversForPullRequests.RequiredReviewersKey, "0",
				),
				branchFinding(requiresCodeOwnersReview.Probe, "main", finding.OutcomeFalse),
				branchFinding(requiresLastPushApproval.Probe, "main", finding.OutcomeFalse),
				branchFinding(requiresUpToDateBranches.Probe, "main", finding.OutcomeFalse),
				branchFinding(runsStatusChecksBeforeMerging.Probe, "main", finding.OutcomeTrue),
				branchFinding(requiresPRsToChangeCode.Probe, "main", finding.OutcomeTrue),
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
				branchFinding(blocksDeleteOnBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(blocksForcePushOnBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(branchesAreProtected.Probe, "main", finding.OutcomeTrue),
				branchFinding(branchProtectionAppliesToAdmins.Probe, "main", finding.OutcomeFalse),
				branchFinding(dismissesStaleReviews.Probe, "main", finding.OutcomeFalse),
				withValue(
					branchFinding(requiresApproversForPullRequests.Probe, "main", finding.OutcomeFalse),
					requiresApproversForPullRequests.RequiredReviewersKey, "0",
				),
				branchFinding(requiresCodeOwnersReview.Probe, "main", finding.OutcomeFalse),
				branchFinding(requiresLastPushApproval.Probe, "main", finding.OutcomeFalse),
				branchFinding(requiresUpToDateBranches.Probe, "main", finding.OutcomeFalse),
				branchFinding(runsStatusChecksBeforeMerging.Probe, "main", finding.OutcomeTrue),
				branchFinding(requiresPRsToChangeCode.Probe, "main", finding.OutcomeTrue),
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
				branchFinding(blocksDeleteOnBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(blocksForcePushOnBranches.Probe, "main", finding.OutcomeFalse),
				branchFinding(branchesAreProtected.Probe, "main", finding.OutcomeTrue),
				branchFinding(branchProtectionAppliesToAdmins.Probe, "main", finding.OutcomeFalse),
				branchFinding(dismissesStaleReviews.Probe, "main", finding.OutcomeFalse),
				withValue(
					branchFinding(requiresApproversForPullRequests.Probe, "main", finding.OutcomeFalse),
					requiresApproversForPullRequests.RequiredReviewersKey, "0",
				),
				branchFinding(requiresCodeOwnersReview.Probe, "main", finding.OutcomeFalse),
				branchFinding(requiresLastPushApproval.Probe, "main", finding.OutcomeFalse),
				branchFinding(requiresUpToDateBranches.Probe, "main", finding.OutcomeFalse),
				branchFinding(runsStatusChecksBeforeMerging.Probe, "main", finding.OutcomeTrue),
				branchFinding(requiresPRsToChangeCode.Probe, "main", finding.OutcomeTrue),
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
				branchFinding(blocksDeleteOnBranches.Probe, "main", finding.OutcomeFalse),
				branchFinding(blocksForcePushOnBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(branchesAreProtected.Probe, "main", finding.OutcomeTrue),
				branchFinding(branchProtectionAppliesToAdmins.Probe, "main", finding.OutcomeFalse),
				branchFinding(dismissesStaleReviews.Probe, "main", finding.OutcomeFalse),
				withValue(
					branchFinding(requiresApproversForPullRequests.Probe, "main", finding.OutcomeFalse),
					requiresApproversForPullRequests.RequiredReviewersKey, "0",
				),
				branchFinding(requiresCodeOwnersReview.Probe, "main", finding.OutcomeFalse),
				branchFinding(requiresLastPushApproval.Probe, "main", finding.OutcomeFalse),
				branchFinding(requiresUpToDateBranches.Probe, "main", finding.OutcomeFalse),
				branchFinding(runsStatusChecksBeforeMerging.Probe, "main", finding.OutcomeTrue),
				branchFinding(requiresPRsToChangeCode.Probe, "main", finding.OutcomeTrue),
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
				branchFinding(blocksDeleteOnBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(blocksForcePushOnBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(branchesAreProtected.Probe, "main", finding.OutcomeTrue),
				branchFinding(branchProtectionAppliesToAdmins.Probe, "main", finding.OutcomeTrue),
				branchFinding(dismissesStaleReviews.Probe, "main", finding.OutcomeTrue),
				withValue(
					branchFinding(requiresApproversForPullRequests.Probe, "main", finding.OutcomeTrue),
					requiresApproversForPullRequests.RequiredReviewersKey, "1",
				),
				branchFinding(requiresCodeOwnersReview.Probe, "main", finding.OutcomeTrue),
				branchFinding(requiresLastPushApproval.Probe, "main", finding.OutcomeTrue),
				branchFinding(requiresUpToDateBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(runsStatusChecksBeforeMerging.Probe, "main", finding.OutcomeTrue),
				branchFinding(requiresPRsToChangeCode.Probe, "main", finding.OutcomeTrue),
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
				branchFinding(blocksDeleteOnBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(blocksForcePushOnBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(branchesAreProtected.Probe, "main", finding.OutcomeTrue),
				branchFinding(branchProtectionAppliesToAdmins.Probe, "main", finding.OutcomeTrue),
				branchFinding(dismissesStaleReviews.Probe, "main", finding.OutcomeTrue),
				withValue(
					branchFinding(requiresApproversForPullRequests.Probe, "main", finding.OutcomeTrue),
					requiresApproversForPullRequests.RequiredReviewersKey, "1",
				),
				branchFinding(requiresCodeOwnersReview.Probe, "main", finding.OutcomeTrue),
				branchFinding(requiresLastPushApproval.Probe, "main", finding.OutcomeTrue),
				branchFinding(requiresUpToDateBranches.Probe, "main", finding.OutcomeTrue),
				branchFinding(runsStatusChecksBeforeMerging.Probe, "main", finding.OutcomeTrue),
				branchFinding(requiresPRsToChangeCode.Probe, "main", finding.OutcomeTrue),
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

//nolint:gocritic // not worried about param size / efficiency since this is a test
func withValue(f finding.Finding, k, v string) finding.Finding {
	f.Values[k] = v
	return f
}
