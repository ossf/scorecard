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
	sce "github.com/ossf/scorecard/v4/errors"
)

// CodeReview retrieves the raw data for the Code-Review check.
func CodeReview(c clients.RepoClient) (checker.CodeReviewData, error) {
	results := []checker.DefaultBranchCommit{}

	// 1. Look at the latest commits
	commits, err := c.ListCommits()
	if err != nil {
		return checker.CodeReviewData{}, fmt.Errorf("%w", err)
	}

	oc := make(map[string]checker.DefaultBranchCommit)
	for _, commit := range commits {
		com := commitRequest(commit)
		results = append(results, com)
		// Keep an index of commits by SHA.
		oc[commit.SHA] = com
		fmt.Println("adding", commit.SHA, commit.Committer.Login, com.Committer.Login)
	}

	// 2. Look at merge requests.
	mrs, err := c.ListMergedPRs()
	if err != nil {
		return checker.CodeReviewData{}, fmt.Errorf("%w", err)
	}

	for i := range mrs {
		mr := mrs[i]
		if mr.MergedAt.IsZero() {
			continue
		}

		// If the merge request is not a recent commit, skip.
		com, exists := oc[mr.MergeCommit.SHA]
		if !exists {
			fmt.Println("skipping", mr.MergeCommit.SHA, mr.Number)
			continue
		}

		// Sanity checks the logins are the same.
		if com.Committer.Login != mr.MergeCommit.Committer.Login {
			fmt.Println(mr.MergeCommit.SHA, mr.Number, com.Committer.Login, mr.MergeCommit.Committer.Login)
			return checker.CodeReviewData{}, sce.WithMessage(sce.ErrScorecardInternal,
				fmt.Sprintf("commit login (%s) different from merge request commit login (%s)",
					com.Committer.Login, mr.MergeCommit.Committer.Login))
		}

		// We have a recent merge request: add other fields.
		com.ApprovedReviews = approvedReviews(&mr, &com)
		com.MergeRequest = mergeRequest(&mr)

		results = append(results, com)
	}

	if len(results) > 0 {
		return checker.CodeReviewData{DefaultBranchCommits: results}, nil
	}

	return checker.CodeReviewData{DefaultBranchCommits: results}, nil
}

func reviewPlatform(platform string) checker.ApprovedReviews {
	mr := checker.ApprovedReviews{
		Platform: checker.ReviewPlatform{
			Name: platform,
		},
	}
	return mr
}

func commitRequest(c clients.Commit) checker.DefaultBranchCommit {
	r := checker.DefaultBranchCommit{
		Committer: checker.User{
			Login: c.Committer.Login,
		},
		SHA:           c.SHA,
		CommitMessage: c.Message,
	}

	return r
}

func mergeRequest(mr *clients.PullRequest) *checker.MergeRequest {
	r := checker.MergeRequest{
		Number: mr.Number,
		Author: checker.User{
			Login: mr.Author.Login,
		},

		Labels:  labels(mr),
		Reviews: reviews(mr),
	}
	return &r
}

func labels(mr *clients.PullRequest) []string {
	labels := []string{}
	for _, l := range mr.Labels {
		labels = append(labels, l.Name)
	}
	return labels
}

func approvedReviews(mr *clients.PullRequest, c *checker.DefaultBranchCommit) *checker.ApprovedReviews {
	var review checker.ApprovedReviews

	// Review platform.
	switch {
	case isReviewedOnGitHub(mr):
		review = reviewPlatform(checker.ReviewPlatformGitHub)
		review.MaintainerReviews = maintainerReviews(mr)

	case isReviewedOnProw(mr):
		review = reviewPlatform(checker.ReviewPlatformProw)

	case isReviewedOnGerrit(c):
		review = reviewPlatform(checker.ReviewPlatformGerrit)

	}

	return &review
}

func reviews(mr *clients.PullRequest) []checker.Review {
	reviews := []checker.Review{}
	for _, m := range mr.Reviews {
		r := checker.Review{
			State: m.State,
		}

		if m.Author != nil &&
			m.Author.Login != "" {
			r.Reviewer = checker.User{
				Login: m.Author.Login,
			}
		}
		reviews = append(reviews, r)
	}
	return reviews
}

func maintainerReviews(mr *clients.PullRequest) []checker.Review {
	reviews := []checker.Review{}
	mauthors := make(map[string]bool)
	for _, m := range mr.Reviews {
		if !(m.State == "APPROVED" &&
			m.Author != nil &&
			m.Author.Login != "") {
			continue
		}

		if _, exists := mauthors[m.Author.Login]; !exists {
			reviews = append(reviews, checker.Review{
				Reviewer: checker.User{
					Login: m.Author.Login,
				},
				State: checker.ReviewStateApproved,
			})
			// Needed because it's is possible for a user
			// to approve a merge request multiple times.
			mauthors[m.Author.Login] = true
		}
	}

	if len(reviews) > 0 {
		return reviews
	}

	// Check if the merge request is committed by someone other than author. This is kind
	// of equivalent to a review and is done several times on small prs to save
	// time on clicking the approve button.
	if mr.MergeCommit.Committer.Login != "" &&
		mr.MergeCommit.Committer.Login != mr.Author.Login {
		reviews = append(reviews, checker.Review{
			State: checker.ReviewStateApproved,
			Reviewer: checker.User{
				Login: mr.MergeCommit.Committer.Login,
			},
		})
	}
	return reviews
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

func isReviewedOnGerrit(c *checker.DefaultBranchCommit) bool {
	m := c.CommitMessage
	return strings.Contains(m, "\nReviewed-on: ") &&
		strings.Contains(m, "\nReviewed-by: ")
}
