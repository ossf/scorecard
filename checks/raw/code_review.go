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

package raw

import (
	"fmt"
	"strings"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
)

// CodeReview retrieves the raw data for the Code-Review check.
func CodeReview(c clients.RepoClient) (checker.CodeReviewData, error) {
	results := []checker.Commit{}

	// 1. Look at merge requests.
	mrs, err := c.ListMergedPRs()
	if err != nil {
		return checker.CodeReviewData{}, fmt.Errorf("%w", err)
	}

	for _, mr := range mrs {
		if mr.MergedAt.IsZero() {
			continue
		}

		// We have a merge request.
		com := checker.Commit{
			Committer: checker.User{
				Login: mr.MergeCommit.Committer.Login,
			},
			Review: ReviewData(&mr),
		}
		results = append(results, com)
	}

	if len(results) > 0 {
		return checker.CodeReviewData{Commits: results}, nil
	}

	// 2. Look at commits.
	commits, err := c.ListCommits()
	if err != nil {
		return checker.CodeReviewData{}, fmt.Errorf("%w", err)
	}

	for _, commit := range commits {
		com := commitRequestData(&commit)
		results = append(results, com)
	}

	return checker.CodeReviewData{Commits: results}, nil
}

func reviewPlatform(platform string) checker.Review {
	review := checker.Review{
		Platform: checker.ReviewPlatform{
			Name: platform,
		},
	}
	return review
}

func commitRequestData(c *clients.Commit) checker.Commit {
	r := checker.Commit{
		Committer: checker.User{
			Login: c.Committer.Login,
		},
	}

	if isReviewedOnGerrit(c) {
		review := reviewPlatform(checker.ReviewPlatformGerrit)
		r.Review = &review
	}

	return r
}

func ReviewData(mr *clients.PullRequest) *checker.Review {
	var review checker.Review

	// Review platform.
	// Note: Gerrit does not use merge requests and is checked
	// in for commits only via commitRequestData().
	switch {
	case isReviewedOnGitHub(mr):
		review = reviewPlatform(checker.ReviewPlatformGitHub)
		review.Authors = reviewAuthors(mr)

	case isReviewedOnProw(mr):
		review = reviewPlatform(checker.ReviewPlatformProw)
	}

	// Add merge request.
	r := checker.MergeRequest{
		Number: mr.Number,
		Author: checker.User{
			Login: mr.Author.Login,
		},
	}
	review.MergeRequest = &r

	return &review
}

func reviewAuthors(mr *clients.PullRequest) []checker.User {
	authors := []checker.User{}
	mauthors := make(map[string]bool, 0)
	for _, m := range mr.Reviews {
		if !(m.State == "APPROVED" &&
			m.Author != nil &&
			m.Author.Login != "") {
			continue
		}

		if _, exists := mauthors[m.Author.Login]; !exists {
			authors = append(authors, checker.User{
				Login: m.Author.Login,
			})
			// Needed because it's is possible for a user
			// to approve a merge request multiple times.
			mauthors[m.Author.Login] = true
		}
	}

	if len(authors) > 0 {
		return authors
	}

	// Check if the merge request is committed by someone other than author. This is kind
	// of equivalent to a review and is done several times on small prs to save
	// time on clicking the approve button.
	if mr.MergeCommit.Committer.Login != "" &&
		mr.MergeCommit.Committer.Login != mr.Author.Login {
		authors = append(authors, checker.User{
			Login: mr.MergeCommit.Committer.Login,
		})
	}

	return authors
}

func isReviewedOnGitHub(mr *clients.PullRequest) bool {
	for _, m := range mr.Reviews {
		if m.State == "APPROVED" {
			return true
		}
	}

	// Check if the merge request is committed by someone other than author. This is kind
	// of equivalent to a review and is done several times on small prs to save
	// time on clicking the approve button.
	if mr.MergeCommit.Committer.Login != "" &&
		mr.MergeCommit.Committer.Login != mr.Author.Login {
		return true
	}

	return false
}

func isReviewedOnProw(mr *clients.PullRequest) bool {
	for _, l := range mr.Labels {
		if l.Name == "lgtm" || l.Name == "approved" {
			return true
		}
	}
	return false
}

func isReviewedOnGerrit(c *clients.Commit) bool {
	commitMessage := c.Message
	return strings.Contains(commitMessage, "\nReviewed-on: ") &&
		strings.Contains(commitMessage, "\nReviewed-by: ")
}
