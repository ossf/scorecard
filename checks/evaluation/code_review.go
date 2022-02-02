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
	reviewStateApproved  = "approved"
)

// ApprovedReviews represents the LGTMs associated with a commit
// to the default branch.
type approvedReviews struct {
	Platform string
	// Note: this field is only populated for GitHub and Prow.
	MaintainerReviews []review
}

// Review represent a single-maintainer's review.
type review struct {
	Reviewer string
	State    string
}

// CodeReview applies the score policy for the Code-Review check.
func CodeReview(name string, dl checker.DetailLogger,
	r *checker.CodeReviewData) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	totalCommits := 0
	totalReviewed := map[string]int{
		// The 3 platforms we support.
		reviewPlatformGitHub: 0,
		reviewPlatformProw:   0,
		reviewPlatformGerrit: 0,
	}

	for _, commit := range r.DefaultBranchCommits {
		// New commit to consider.
		totalCommits++

		ar := getApprovedReviews(&commit, dl)
		// No commits.
		if ar.Platform == "" {
			continue
		}

		totalReviewed[ar.Platform]++

		switch ar.Platform {
		// GitHub reviews.
		case reviewPlatformGitHub:
			dl.Debug3(&checker.LogMessage{
				Text: fmt.Sprintf("%s #%d merge request approved",
					reviewPlatformGitHub, commit.MergeRequest.Number),
			})

		// Prow reviews.
		case reviewPlatformProw:
			if commit.MergeRequest != nil {
				dl.Debug3(&checker.LogMessage{
					Text: fmt.Sprintf("%s #%d merge request approved",
						reviewPlatformProw, commit.MergeRequest.Number),
				})
			}

		// Gerrit reviews.
		case reviewPlatformGerrit:
			dl.Debug3(&checker.LogMessage{
				Text: fmt.Sprintf("%s commit approved", reviewPlatformGerrit),
			})
		}
	}

	if totalCommits == 0 {
		return checker.CreateInconclusiveResult(name, "no commits found")
	}

	if totalReviewed[reviewPlatformGitHub] == 0 &&
		totalReviewed[reviewPlatformGerrit] == 0 &&
		totalReviewed[reviewPlatformProw] == 0 {
		// Only show all warnings if all fail.
		// We should not show warning if at least one succeeds, as this is confusing.
		for k := range totalReviewed {
			dl.Warn3(&checker.LogMessage{
				Text: fmt.Sprintf("no %s reviews found", k),
			})
		}

		return checker.CreateMinScoreResult(name, "no reviews found")
	}

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

func getApprovedReviews(c *checker.DefaultBranchCommit, dl checker.DetailLogger) approvedReviews {
	var review approvedReviews

	// Review platform.
	switch {
	case isReviewedOnGitHub(c):
		review.Platform = reviewPlatformGitHub

	case isReviewedOnProw(c, dl):
		review.Platform = reviewPlatformProw

	case isReviewedOnGerrit(c, dl):
		review.Platform = reviewPlatformGerrit

	}

	return review
}

func maintainerReviews(c *checker.DefaultBranchCommit) []review {
	reviews := []review{}
	mauthors := make(map[string]bool)
	mr := c.MergeRequest
	for _, r := range mr.Reviews {

		if !(r.State == "APPROVED" &&
			r.Reviewer.Login != "") {
			continue
		}

		if _, exists := mauthors[r.Reviewer.Login]; !exists {
			reviews = append(reviews, review{
				Reviewer: r.Reviewer.Login,
				State:    reviewStateApproved,
			})
			// Needed because it's is possible for a user
			// to approve a merge request multiple times.
			mauthors[r.Reviewer.Login] = true
		}

	}

	if len(reviews) > 0 {
		return reviews
	}

	// Check if the merge request is committed by someone other than author. This is kind
	// of equivalent to a review and is done several times on small prs to save
	// time on clicking the approve button.
	if c.Committer.Login != "" &&
		c.Committer.Login != mr.Author.Login {
		reviews = append(reviews, review{
			State:    reviewStateApproved,
			Reviewer: c.Committer.Login,
		})
	}
	return reviews
}

func isReviewedOnGitHub(c *checker.DefaultBranchCommit) bool {
	mr := c.MergeRequest
	if mr == nil {
		return false
	}

	for _, r := range mr.Reviews {
		if r.State == "APPROVED" {
			return true
		}
	}

	// Check if the merge request is committed by someone other than author. This is kind
	// of equivalent to a review and is done several times on small prs to save
	// time on clicking the approve button.
	if c.Committer.Login != "" &&
		c.Committer.Login != mr.Author.Login {
		return true
	}

	return false
}

func isReviewedOnProw(c *checker.DefaultBranchCommit, dl checker.DetailLogger) bool {
	if isBot(c.Committer.Login) {
		dl.Debug3(&checker.LogMessage{
			Text: fmt.Sprintf("skip commit from bot account: %s", c.Committer.Login),
		})
		return true
	}

	if c.MergeRequest != nil {
		for _, l := range c.MergeRequest.Labels {
			if l == "lgtm" || l == "approved" {
				return true
			}
		}
	}
	return false
}

func isReviewedOnGerrit(c *checker.DefaultBranchCommit, dl checker.DetailLogger) bool {
	if isBot(c.Committer.Login) {
		dl.Debug3(&checker.LogMessage{
			Text: fmt.Sprintf("skip commit from bot account: %s", c.Committer.Login),
		})
		return true
	}

	m := c.CommitMessage
	return strings.Contains(m, "\nReviewed-on: ") &&
		strings.Contains(m, "\nReviewed-by: ")
}
