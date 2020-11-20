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
	"strings"

	"github.com/google/go-github/v32/github"
	"github.com/ossf/scorecard/checker"
)

func init() {
	registerCheck("Code-Review", DoesCodeReview)
}

// DoesCodeReview attempts to determine whether a project requires review before code gets merged.
// It uses a set of heuristics:
// - Looking at the repo configuration to see if reviews are required
// - Checking if most of the recent merged PRs were "Approved"
// - Looking for other well-known review labels
func DoesCodeReview(c checker.Checker) checker.CheckResult {
	return checker.MultiCheck(
		IsPrReviewRequired,
		GithubCodeReview,
		ProwCodeReview,
		GerritCodeReview,
	)(c)
}

func GithubCodeReview(c checker.Checker) checker.CheckResult {
	// Look at some merged PRs to see if they were reviewed
	prs, _, err := c.Client.PullRequests.List(c.Ctx, c.Owner, c.Repo, &github.PullRequestListOptions{
		State: "closed",
	})
	if err != nil {
		return checker.InconclusiveResult
	}

	totalMerged := 0
	totalReviewed := 0
	for _, pr := range prs {
		if pr.MergedAt == nil {
			continue
		}
		totalMerged++

		// check if the PR is approved by a reviewer
		foundApprovedReview := false
		reviews, _, err := c.Client.PullRequests.ListReviews(c.Ctx, c.Owner, c.Repo, pr.GetNumber(), &github.ListOptions{})
		if err != nil {
			continue
		}
		for _, r := range reviews {
			if r.GetState() == "APPROVED" {
				c.Logf("found review approved pr: %d", pr.GetNumber())
				totalReviewed++
				foundApprovedReview = true
				break
			}
		}

		// check if the PR is committed by someone other than author. this is kind
		// of equivalent to a review and is done several times on small prs to save
		// time on clicking the approve button.
		if !foundApprovedReview {
			commit, _, err := c.Client.Repositories.GetCommit(c.Ctx, c.Owner, c.Repo, pr.GetMergeCommitSHA())
			if err != nil {
				commitAuthor := commit.GetAuthor().GetLogin()
				commitCommitter := commit.GetCommitter().GetLogin()
				if commitAuthor != "" && commitCommitter != "" && commitAuthor != commitCommitter {
					c.Logf("found different author and committer for pr: %d", pr.GetNumber())
					totalReviewed++
				}
			}
		}

	}

	if totalReviewed > 0 {
		c.Logf("github code reviews found")
	}
	return checker.ProportionalResult(totalReviewed, totalMerged, .75)
}

func IsPrReviewRequired(c checker.Checker) checker.CheckResult {
	// Look to see if review is enforced.
	r, _, err := c.Client.Repositories.Get(c.Ctx, c.Owner, c.Repo)
	if err != nil {
		return checker.RetryResult(err)
	}

	// Check the branch protection rules, we may not be able to get these though.
	bp, _, err := c.Client.Repositories.GetBranchProtection(c.Ctx, c.Owner, c.Repo, r.GetDefaultBranch())
	if err != nil {
		return checker.InconclusiveResult
	}
	if bp.GetRequiredPullRequestReviews().RequiredApprovingReviewCount >= 1 {
		c.Logf("pr review policy enforced")
		return checker.CheckResult{
			Pass:       true,
			Confidence: 5,
		}
	}
	return checker.InconclusiveResult
}

func ProwCodeReview(c checker.Checker) checker.CheckResult {
	// Look at some merged PRs to see if they were reviewed
	prs, _, err := c.Client.PullRequests.List(c.Ctx, c.Owner, c.Repo, &github.PullRequestListOptions{
		State: "closed",
	})
	if err != nil {
		return checker.InconclusiveResult
	}

	totalMerged := 0
	totalReviewed := 0
	for _, pr := range prs {
		if pr.MergedAt == nil {
			continue
		}
		totalMerged++
		for _, l := range pr.Labels {
			if l.GetName() == "lgtm" || l.GetName() == "approved" {
				totalReviewed++
				break
			}
		}
	}

	if totalReviewed == 0 {
		return checker.InconclusiveResult
	}
	c.Logf("prow code reviews found")
	return checker.ProportionalResult(totalReviewed, totalMerged, .75)
}

func GerritCodeReview(c checker.Checker) checker.CheckResult {
	commits, _, err := c.Client.Repositories.ListCommits(c.Ctx, c.Owner, c.Repo, &github.CommitsListOptions{})
	if err != nil {
		return checker.RetryResult(err)
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
			c.Logf("skip commit from bot account: %s", committer)
			continue
		}

		total++

		// check for gerrit use via Reviewed-on and Reviewed-by
		commitMessage := commit.GetCommit().GetMessage()
		if strings.Contains(commitMessage, "\nReviewed-on: ") &&
			strings.Contains(commitMessage, "\nReviewed-by: ") {
			totalReviewed++
			continue
		}
	}

	if totalReviewed == 0 {
		return checker.InconclusiveResult
	}
	c.Logf("gerrit code reviews found")
	return checker.ProportionalResult(totalReviewed, total, .75)
}
