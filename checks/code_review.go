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
	"errors"
	"strings"

	"github.com/google/go-github/v32/github"

	"github.com/ossf/scorecard/checker"
)

const (
	// CheckCodeReview is the registered name for DoesCodeReview.
	CheckCodeReview = "Code-Review"
	crPassThreshold = .75
)

// ErrorNoReviews indicates no reviews were found for this repo.
var ErrorNoReviews = errors.New("no reviews found")

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
	return checker.MultiCheckOr(
		IsPrReviewRequired,
		GithubCodeReview,
		ProwCodeReview,
		CommitMessageHints,
	)(c)
}

func GithubCodeReview(c *checker.CheckRequest) checker.CheckResult {
	// Look at some merged PRs to see if they were reviewed.
	totalMerged := 0
	totalReviewed := 0
	for _, pr := range c.RepoClient.ListMergedPRs() {
		if pr.MergedAt.IsZero() {
			continue
		}
		totalMerged++

		// check if the PR is approved by a reviewer
		foundApprovedReview := false
		for _, r := range pr.Reviews {
			if r.State == "APPROVED" {
				c.Logf("found review approved pr: %d", pr.Number)
				totalReviewed++
				foundApprovedReview = true
				break
			}
		}

		// check if the PR is committed by someone other than author. this is kind
		// of equivalent to a review and is done several times on small prs to save
		// time on clicking the approve button.
		if !foundApprovedReview {
			if !pr.MergeCommit.AuthoredByCommitter {
				c.Logf("found pr with committer different than author: %d", pr.Number)
				totalReviewed++
			}
		}
	}

	if totalReviewed > 0 {
		c.Logf("github code reviews found")
	}
	return checker.MakeProportionalResult(CheckCodeReview, totalReviewed, totalMerged, crPassThreshold)
}

func IsPrReviewRequired(c *checker.CheckRequest) checker.CheckResult {
	// Look to see if review is enforced.
	// Check the branch protection rules, we may not be able to get these though.
	if c.RepoClient.GetDefaultBranch().BranchProtectionRule.RequiredApprovingReviewCount >= 1 {
		c.Logf("pr review policy enforced")
		const confidence = 5
		return checker.CheckResult{
			Name:       CheckCodeReview,
			Pass:       true,
			Confidence: confidence,
		}
	}
	return checker.MakeInconclusiveResult(CheckCodeReview, nil)
}

func ProwCodeReview(c *checker.CheckRequest) checker.CheckResult {
	// Look at some merged PRs to see if they were reviewed
	totalMerged := 0
	totalReviewed := 0
	for _, pr := range c.RepoClient.ListMergedPRs() {
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

	if totalReviewed == 0 {
		return checker.MakeInconclusiveResult(CheckCodeReview, ErrorNoReviews)
	}
	c.Logf("prow code reviews found")
	return checker.MakeProportionalResult(CheckCodeReview, totalReviewed, totalMerged, crPassThreshold)
}

func CommitMessageHints(c *checker.CheckRequest) checker.CheckResult {
	commits, _, err := c.Client.Repositories.ListCommits(c.Ctx, c.Owner, c.Repo, &github.CommitsListOptions{})
	if err != nil {
		return checker.MakeRetryResult(CheckCodeReview, err)
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
		return checker.MakeInconclusiveResult(CheckCodeReview, ErrorNoReviews)
	}
	c.Logf("code reviews found")
	return checker.MakeProportionalResult(CheckCodeReview, totalReviewed, total, crPassThreshold)
}
