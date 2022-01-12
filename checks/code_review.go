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

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
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

	score, reason := selectBestScoreAndReason(hintScore, ghScore, hintReason, ghReason, c.Dlogger)

	// Prow CI/CD.
	prowScore, prowReason, err := prowCodeReview(c)
	if err != nil {
		return checker.CreateRuntimeErrorResult(CheckCodeReview, err)
	}

	score, reason = selectBestScoreAndReason(prowScore, score, prowReason, reason, c.Dlogger)
	if score == checker.MinResultScore {
		c.Dlogger.Info3(&checker.LogMessage{
			Text: reason,
		})
		return checker.CreateResultWithScore(CheckCodeReview, "no reviews detected", score)
	}

	if score == checker.InconclusiveResultScore {
		return checker.CreateInconclusiveResult(CheckCodeReview, "no reviews detected")
	}

	return checker.CreateResultWithScore(CheckCodeReview, checker.NormalizeReason(reason, score), score)
}

//nolint
func selectBestScoreAndReason(s1, s2 int, r1, r2 string,
	dl checker.DetailLogger) (int, string) {
	if s1 > s2 {
		dl.Info3(&checker.LogMessage{
			Text: r2,
		})
		return s1, r1
	}

	dl.Info3(&checker.LogMessage{
		Text: r1,
	})
	return s2, r2
}

//nolint
func githubCodeReview(c *checker.CheckRequest) (int, string, error) {
	// Look at some merged PRs to see if they were reviewed.
	totalMerged := 0
	totalReviewed := 0
	prs, err := c.RepoClient.ListMergedPRs()
	if err != nil {
		return 0, "", sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("RepoClient.ListMergedPRs: %v", err))
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
				c.Dlogger.Debug3(&checker.LogMessage{
					Text: fmt.Sprintf("found review approved pr: %d", pr.Number),
				})
				totalReviewed++
				foundApprovedReview = true
				break
			}
		}

		// Check if the PR is committed by someone other than author. this is kind
		// of equivalent to a review and is done several times on small prs to save
		// time on clicking the approve button.
		if !foundApprovedReview &&
			pr.MergeCommit.Committer.Login != "" &&
			pr.MergeCommit.Committer.Login != pr.Author.Login {
			c.Dlogger.Debug3(&checker.LogMessage{
				Text: fmt.Sprintf("found PR#%d with committer (%s) different from author (%s)",
					pr.Number, pr.Author.Login, pr.MergeCommit.Committer.Login),
			})
			totalReviewed++
			foundApprovedReview = true
		}

		if !foundApprovedReview {
			c.Dlogger.Debug3(&checker.LogMessage{
				Text: fmt.Sprintf("merged PR without code review: %d", pr.Number),
			})
		}

	}

	return createReturn("GitHub", totalReviewed, totalMerged)
}

//nolint
func prowCodeReview(c *checker.CheckRequest) (int, string, error) {
	// Look at some merged PRs to see if they were reviewed
	totalMerged := 0
	totalReviewed := 0
	prs, err := c.RepoClient.ListMergedPRs()
	if err != nil {
		sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("RepoClient.ListMergedPRs: %v", err))
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
	commits, err := c.RepoClient.ListCommits()
	if err != nil {
		return checker.InconclusiveResultScore, "",
			sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("Client.Repositories.ListCommits: %v", err))
	}

	total := 0
	totalReviewed := 0
	for _, commit := range commits {
		isBot := false
		committer := commit.Committer.Login
		for _, substring := range []string{"bot", "gardener"} {
			if strings.Contains(committer, substring) {
				isBot = true
				break
			}
		}
		if isBot {
			c.Dlogger.Debug3(&checker.LogMessage{
				Text: fmt.Sprintf("skip commit from bot account: %s", committer),
			})
			continue
		}

		total++

		// check for gerrit use via Reviewed-on and Reviewed-by
		commitMessage := commit.Message
		if strings.Contains(commitMessage, "\nReviewed-on: ") &&
			strings.Contains(commitMessage, "\nReviewed-by: ") {
			c.Dlogger.Debug3(&checker.LogMessage{
				Text: fmt.Sprintf("Gerrit review found for commit '%s'", commit.SHA),
			})
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
