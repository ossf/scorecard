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
		checker.ReviewPlatformGitHub: 0,
		checker.ReviewPlatformProw:   0,
		checker.ReviewPlatformGerrit: 0,
	}

	for _, commit := range r.DefaultBranchCommits {
		if commit.ApprovedReviews == nil && isBot(commit.Committer.Login) {
			dl.Debug3(&checker.LogMessage{
				Text: fmt.Sprintf("skip commit from bot account: %s", commit.Committer),
			})
			continue
		}

		// New commit to consider.
		totalCommits++

		// No reviews.
		if commit.ApprovedReviews == nil {
			continue
		}

		totalReviewed[commit.ApprovedReviews.Platform.Name]++

		switch commit.ApprovedReviews.Platform.Name {
		// GitHub reviews.
		case checker.ReviewPlatformGitHub:
			dl.Debug3(&checker.LogMessage{
				Text: fmt.Sprintf("%s #%d merge request approved",
					checker.ReviewPlatformGitHub, commit.MergeRequest.Number),
			})

		// Prow reviews.
		case checker.ReviewPlatformProw:
			dl.Debug3(&checker.LogMessage{
				Text: fmt.Sprintf("%s #%d merge request approved",
					checker.ReviewPlatformProw, commit.MergeRequest.Number),
			})

		// Gerrit reviews.
		case checker.ReviewPlatformGerrit:
			dl.Debug3(&checker.LogMessage{
				Text: fmt.Sprintf("%s commit approved", checker.ReviewPlatformGerrit),
			})
		}
	}

	if totalCommits == 0 {
		return checker.CreateInconclusiveResult(name, "no commits found")
	}

	if totalReviewed[checker.ReviewPlatformGitHub] == 0 &&
		totalReviewed[checker.ReviewPlatformGerrit] == 0 &&
		totalReviewed[checker.ReviewPlatformProw] == 0 {
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
	nbReviews, reviewSystem := reviews(totalReviewed)
	if nbReviews == totalCommits {
		return checker.CreateMaxScoreResult(name,
			fmt.Sprintf("all last %v commits are reviewed through %s", totalCommits, reviewSystem))
	}

	reason := fmt.Sprintf("%s code reviews found for %v commits out of the last %v", reviewSystem, nbReviews, totalCommits)
	return checker.CreateProportionalScoreResult(name, reason, nbReviews, totalCommits)
}

func reviews(m map[string]int) (int, string) {
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
