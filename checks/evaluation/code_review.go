// Copyright 2021 Security Scorecard Authors
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
	"fmt"
	"strings"

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
)

var (
	reviewPlatformGitHub = "GitHub"
	reviewPlatformProw   = "Prow"
	reviewPlatformGerrit = "Gerrit"
)

// CodeReview applies the score policy for the Code-Review check.
func CodeReview(name string, dl checker.DetailLogger,
	r *checker.CodeReviewData) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	if len(r.DefaultBranchCommits) == 0 {
		return checker.CreateInconclusiveResult(name, "no commits found")
	}

	totalReviewed := map[string]int{
		// The 3 platforms we support.
		reviewPlatformGitHub: 0,
		reviewPlatformProw:   0,
		reviewPlatformGerrit: 0,
	}

	for i := range r.DefaultBranchCommits {
		commit := r.DefaultBranchCommits[i]

		rs := getApprovedReviewSystem(&commit, dl)
		if rs == "" {
			dl.Warn(&checker.LogMessage{
				Text: fmt.Sprintf("no reviews found for commit: %s", commit.SHA),
			})
			continue
		}

		totalReviewed[rs]++
	}

	if totalReviewed[reviewPlatformGitHub] == 0 &&
		totalReviewed[reviewPlatformGerrit] == 0 &&
		totalReviewed[reviewPlatformProw] == 0 {
		return checker.CreateMinScoreResult(name, "no reviews found")
	}

	totalCommits := len(r.DefaultBranchCommits)
	// Consider a single review system.
	nbReviews, reviewSystem := computeReviews(totalReviewed)
	if nbReviews == totalCommits {
		return checker.CreateMaxScoreResult(name,
			fmt.Sprintf("all last %v commits are reviewed through %s", totalCommits, reviewSystem))
	}

	reason := fmt.Sprintf("%s code reviews found for %v commits out of the last %v", reviewSystem, nbReviews, totalCommits)
	return checker.CreateProportionalScoreResult(name, reason, nbReviews, totalCommits)
}

func computeReviews(m map[string]int) (int, string) {
	n := 0
	s := ""
	for k, v := range m {
		if v > n {
			n = v
			s = k
		}
	}
	return n, s
}

func isBot(name string) bool {
	for _, substring := range []string{"bot", "gardener"} {
		if strings.Contains(name, substring) {
			return true
		}
	}
	return false
}

func getApprovedReviewSystem(c *checker.DefaultBranchCommit, dl checker.DetailLogger) string {
	switch {
	case isReviewedOnGitHub(c, dl):
		return reviewPlatformGitHub

	case isReviewedOnProw(c, dl):
		return reviewPlatformProw

	case isReviewedOnGerrit(c, dl):
		return reviewPlatformGerrit
	}

	return ""
}

func isReviewedOnGitHub(c *checker.DefaultBranchCommit, dl checker.DetailLogger) bool {
	mr := c.MergeRequest
	if mr == nil || mr.MergedAt.IsZero() {
		return false
	}

	for _, r := range mr.Reviews {
		if r.State == "APPROVED" {
			dl.Debug(&checker.LogMessage{
				Text: fmt.Sprintf("commit %s was reviewed through %s #%d approved merge request",
					c.SHA, reviewPlatformGitHub, mr.Number),
			})
			return true
		}
	}

	// Check if the merge request is committed by someone other than author. This is kind
	// of equivalent to a review and is done several times on small prs to save
	// time on clicking the approve button.
	if c.Committer.Login != "" &&
		c.Committer.Login != mr.Author.Login {
		dl.Debug(&checker.LogMessage{
			Text: fmt.Sprintf("commit %s was reviewed through %s #%d merge request",
				c.SHA, reviewPlatformGitHub, mr.Number),
		})
		return true
	}

	return false
}

func isReviewedOnProw(c *checker.DefaultBranchCommit, dl checker.DetailLogger) bool {
	if isBot(c.Committer.Login) {
		dl.Debug(&checker.LogMessage{
			Text: fmt.Sprintf("skip commit %s from bot account: %s", c.SHA, c.Committer.Login),
		})
		return true
	}

	if c.MergeRequest != nil && !c.MergeRequest.MergedAt.IsZero() {
		for _, l := range c.MergeRequest.Labels {
			if l == "lgtm" || l == "approved" {
				dl.Debug(&checker.LogMessage{
					Text: fmt.Sprintf("commit %s review was through %s #%d approved merge request",
						c.SHA, reviewPlatformProw, c.MergeRequest.Number),
				})
				return true
			}
		}
	}
	return false
}

func isReviewedOnGerrit(c *checker.DefaultBranchCommit, dl checker.DetailLogger) bool {
	if isBot(c.Committer.Login) {
		dl.Debug(&checker.LogMessage{
			Text: fmt.Sprintf("skip commit %s from bot account: %s", c.SHA, c.Committer.Login),
		})
		return true
	}

	m := c.CommitMessage
	if strings.Contains(m, "\nReviewed-on: ") &&
		strings.Contains(m, "\nReviewed-by: ") {
		dl.Debug(&checker.LogMessage{
			Text: fmt.Sprintf("commit %s was approved through %s", c.SHA, reviewPlatformGerrit),
		})
		return true
	}
	return false
}
