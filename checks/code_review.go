// Copyright 2020 Security Scorecard Authors
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

package checks

import (
	"fmt"
	"strings"

	"github.com/google/go-github/v32/github"

	"github.com/ossf/scorecard/v2/checker"
	sce "github.com/ossf/scorecard/v2/errors"
)

// CheckCodeReview is the registered name for DoesCodeReview.
const CheckCodeReview = "Code-Review"

//nolint:gochecknoinits
func init() {
	registerCheck(CheckCodeReview, DoesCodeReview)
}

// DoesCodeReview attempts to determine whether a project requires review before code gets merged.
// It uses a set of heuristics:
// - Looking at the repo configuration to see if reviews are required.
// - Checking if most of the recent merged PRs were "Approved".
// - Looking for other well-known review labels.
func DoesCodeReview(c *checker.CheckRequest) checker.CheckResult {
	// Branch protection.
	score, reason := isPrReviewRequired(c)
	if score == checker.MaxResultScore {
		return checker.CreateMaxScoreResult(CheckCodeReview, reason)
	}

	// GitHub reviews.
	ghScore, ghReason, err := githubCodeReview(c)
	if err != nil {
		return checker.CreateRuntimeErrorResult(CheckCodeReview, err)
	}

	// Review messages.
	hintScore, hintReason, err := commitMessageHints(c)
	if err != nil {
		return checker.CreateRuntimeErrorResult(CheckCodeReview, err)
	}

	score, reason = selectBestScoreAndReason(hintScore, ghScore, hintReason, ghReason, c.Dlogger)

	// Prow CI/CD.
	prowScore, prowReason, err := prowCodeReview(c)
	if err != nil {
		return checker.CreateRuntimeErrorResult(CheckCodeReview, err)
	}

	score, reason = selectBestScoreAndReason(prowScore, score, prowReason, reason, c.Dlogger)
	if score == checker.InconclusiveResultScore {
		return checker.CreateInconclusiveResult(CheckCodeReview, "no reviews detected")
	}

	return checker.CreateResultWithScore(CheckCodeReview, checker.NormalizeReason(reason, score), score)
}

//nolint
func selectBestScoreAndReason(s1, s2 int, r1, r2 string,
	dl checker.DetailLogger) (int, string) {
	if s1 > s2 {
		dl.Info(r2)
		return s1, r1
	}

	dl.Info(r1)
	return s2, r2
}

//nolint
func githubCodeReview(c *checker.CheckRequest) (int, string, error) {
	// Look at some merged PRs to see if they were reviewed.
	totalMerged := 0
	totalReviewed := 0
	prs, err := c.RepoClient.ListMergedPRs()
	if err != nil {
		return 0, "", sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("RepoClient.ListMergedPRs: %v", err))
	}
	for _, pr := range prs {
		if pr.MergedAt.IsZero() {
			continue
		}
		totalMerged++

		// Check if the PR is approved by a reviewer.
		foundApprovedReview := false
		for _, r := range pr.Reviews {
			if r.State == "APPROVED" {
				c.Dlogger.Debug("found review approved pr: %d", pr.Number)
				totalReviewed++
				foundApprovedReview = true
				break
			}
		}

		// Check if the PR is committed by someone other than author. this is kind
		// of equivalent to a review and is done several times on small prs to save
		// time on clicking the approve button.
		if !foundApprovedReview {
			if !pr.MergeCommit.AuthoredByCommitter {
				c.Dlogger.Debug("found pr with committer different than author: %d", pr.Number)
				totalReviewed++
			}
		}
	}

	return createReturn("GitHub", totalReviewed, totalMerged)
}

//nolint
func isPrReviewRequired(c *checker.CheckRequest) (int, string) {
	// Look to see if review is enforced.
	// Check the branch protection rules, we may not be able to get these though.
	branch, err := c.RepoClient.GetDefaultBranch()
	if err != nil {
		sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("RepoClient.GetDefaultBranch: %v", err))
	}
	if branch.BranchProtectionRule.RequiredApprovingReviewCount >= 1 {
		// If the default value is 0 when we cannot retrieve the value,
		// a non-zero value means we're confident it's enabled.
		return checker.MaxResultScore, "branch protection for default branch is enabled"
	}
	return checker.InconclusiveResultScore, "cannot determine if branch protection is enabled"
}

//nolint
func prowCodeReview(c *checker.CheckRequest) (int, string, error) {
	// Look at some merged PRs to see if they were reviewed
	totalMerged := 0
	totalReviewed := 0
	prs, err := c.RepoClient.ListMergedPRs()
	if err != nil {
		sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("RepoClient.ListMergedPRs: %v", err))
	}
	for _, pr := range prs {
		if pr.MergedAt.IsZero() {
			continue
		}
		totalMerged++
		for _, l := range pr.Labels {
			if l.Name == "lgtm" || l.Name == "approved" {
				totalReviewed++
				break
			}
		}
	}

	return createReturn("Prow", totalReviewed, totalMerged)
}

//nolint
func commitMessageHints(c *checker.CheckRequest) (int, string, error) {
	commits, _, err := c.Client.Repositories.ListCommits(c.Ctx, c.Owner, c.Repo, &github.CommitsListOptions{})
	if err != nil {
		//nolint
		return checker.InconclusiveResultScore, "",
			sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("Client.Repositories.ListCommits: %v", err))
	}

	total := 0
	totalReviewed := 0
	for _, commit := range commits {
		isBot := false
		committer := commit.GetCommitter().GetLogin()
		for _, substring := range []string{"bot", "gardener"} {
			if strings.Contains(committer, substring) {
				isBot = true
				break
			}
		}
		if isBot {
			c.Dlogger.Debug("skip commit from bot account: %s", committer)
			continue
		}

		total++

		// check for gerrit use via Reviewed-on and Reviewed-by
		commitMessage := commit.GetCommit().GetMessage()
		if strings.Contains(commitMessage, "\nReviewed-on: ") &&
			strings.Contains(commitMessage, "\nReviewed-by: ") {
			c.Dlogger.Debug("Gerrit review found for commit '%s'", commit.GetCommit().GetSHA())
			totalReviewed++
			continue
		}
	}

	return createReturn("Gerrit", totalReviewed, total)
}

//nolint
func createReturn(reviewName string, reviewed, total int) (int, string, error) {
	if total > 0 {
		reason := fmt.Sprintf("%s code reviews found for %v commits out of the last %v", reviewName, reviewed, total)
		return checker.CreateProportionalScore(reviewed, total), reason, nil
	}

	return checker.InconclusiveResultScore, fmt.Sprintf("no %v commits found", reviewName), nil
}
