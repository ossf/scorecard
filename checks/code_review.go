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
	"sort"
	"strings"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
)

type scoreAndReason struct {
	reason string
	score  int
}

// CheckCodeReview is the registered name for DoesCodeReview.
const CheckCodeReview = "Code-Review"

// nolint:gochecknoinits
func init() {
	if err := registerCheck(CheckCodeReview, DoesCodeReview); err != nil {
		// this should never happen
		panic(err)
	}
}

// DoesCodeReview attempts to determine whether a project requires review before code gets merged.
// It uses a set of heuristics:
// - Looking at the repo configuration to see if reviews are required.
// - Checking if most of the recent merged PRs were "Approved".
// - Looking for other well-known review labels.
func DoesCodeReview(c *checker.CheckRequest) checker.CheckResult {
	commits, err := c.RepoClient.ListCommits()
	if err != nil {
		return checker.CreateRuntimeErrorResult(CheckCodeReview,
			sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("RepoClient.Commits: %v", err)))
	}

	totalNonBotCommits := 0
	totalMergedPRs := 0
	totalGerritReviewed := 0
	totalGitHubReviewed := 0
	totalProwReviewed := 0
	for i := range commits {
		commit := &commits[i]
		if !isBotCommitted(commit, c.Dlogger) {
			totalNonBotCommits++
			if isGerritReviewed(commit, c.Dlogger) {
				totalGerritReviewed++
			}
		}

		pr := commit.AssociatedMergeRequest
		// TODO(#575): We ignore associated PRs if Scorecard is being run on a fork
		// but the PR was created in the original repo.
		if pr.MergedAt.IsZero() || pr.Repository != c.Repo.String() {
			continue
		}
		totalMergedPRs++

		if isGitHubReviewed(commit, c.Dlogger) {
			totalGitHubReviewed++
		}

		if isProwReviewed(commit) {
			totalProwReviewed++
		}
	}

	gerritScoreAndReason := createReturn("Gerrit", totalGerritReviewed, totalNonBotCommits)
	githubScoreAndReason := createReturn("GitHub", totalGitHubReviewed, totalMergedPRs)
	prowScoreAndReason := createReturn("Prow", totalProwReviewed, totalMergedPRs)

	bestScoreAndReason := selectBestScoreAndReason([]scoreAndReason{
		gerritScoreAndReason, githubScoreAndReason, prowScoreAndReason,
	}, c.Dlogger)
	score := bestScoreAndReason.score
	reason := bestScoreAndReason.reason
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

func isGitHubReviewed(commit *clients.Commit, logger checker.DetailLogger) bool {
	// Check if the associatedPR is approved by a reviewer.
	pr := commit.AssociatedMergeRequest
	foundApprovedReview := false
	for _, r := range pr.Reviews {
		if r.State == "APPROVED" {
			logger.Debug3(&checker.LogMessage{
				Text: fmt.Sprintf("found review approved pr: %d", pr.Number),
			})
			foundApprovedReview = true
			break
		}
	}

	// Check if the PR is committed by someone other than author. this is kind
	// of equivalent to a review and is done several times on small prs to save
	// time on clicking the approve button.
	if !foundApprovedReview &&
		commit.Committer.Login != "" &&
		commit.Committer.Login != pr.Author.Login {
		logger.Debug3(&checker.LogMessage{
			Text: fmt.Sprintf("found PR#%d with committer (%s) different from author (%s)",
				pr.Number, pr.Author.Login, commit.Committer.Login),
		})
		foundApprovedReview = true
	}

	if !foundApprovedReview {
		logger.Debug3(&checker.LogMessage{
			Text: fmt.Sprintf("merged PR without code review: %d", pr.Number),
		})
	}

	return foundApprovedReview
}

func isProwReviewed(commit *clients.Commit) bool {
	for _, l := range commit.AssociatedMergeRequest.Labels {
		if l.Name == "lgtm" || l.Name == "approved" {
			return true
		}
	}
	return false
}

// Check for gerrit use via Reviewed-on and Reviewed-by.
func isGerritReviewed(commit *clients.Commit, logger checker.DetailLogger) bool {
	commitMessage := commit.Message
	isReviewed := strings.Contains(commitMessage, "\nReviewed-on: ") &&
		strings.Contains(commitMessage, "\nReviewed-by: ")
	if isReviewed {
		logger.Debug3(&checker.LogMessage{
			Text: fmt.Sprintf("Gerrit review found for commit '%s'", commit.SHA),
		})
	}

	return isReviewed
}

func isBotCommitted(commit *clients.Commit, logger checker.DetailLogger) bool {
	committer := commit.Committer.Login
	isBot := strings.Contains(committer, "bot") ||
		strings.Contains(committer, "gardener")
	if isBot {
		logger.Debug3(&checker.LogMessage{
			Text: fmt.Sprintf("committed from bot account: %s", commit.Committer.Login),
		})
	}

	return isBot
}

func createReturn(reviewName string, reviewed, total int) scoreAndReason {
	if total > 0 {
		return scoreAndReason{
			score: checker.CreateProportionalScore(reviewed, total),
			reason: fmt.Sprintf(
				"%s code reviews found for %v commits out of the last %v", reviewName, reviewed, total),
		}
	}
	return scoreAndReason{
		score:  checker.InconclusiveResultScore,
		reason: fmt.Sprintf("no %v commits found", reviewName),
	}
}

func selectBestScoreAndReason(scoresAndReasons []scoreAndReason, logger checker.DetailLogger) scoreAndReason {
	// Sort descending.
	sort.SliceStable(scoresAndReasons, func(i, j int) bool {
		return scoresAndReasons[i].score > scoresAndReasons[j].score
	})

	// Log every reason except the highest score.
	for _, sr := range scoresAndReasons[1:] {
		logger.Info3(&checker.LogMessage{
			Text: sr.reason,
		})
	}
	return scoresAndReasons[0]
}
