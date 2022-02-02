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

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
)

// CodeReview retrieves the raw data for the Code-Review check.
func CodeReview(c clients.RepoClient) (checker.CodeReviewData, error) {
	results := []checker.DefaultBranchCommit{}

	// Look at the latest commits.
	commits, err := c.ListCommits()
	if err != nil {
		return checker.CodeReviewData{}, fmt.Errorf("%w", err)
	}

	oc := make(map[string]checker.DefaultBranchCommit)
	for _, commit := range commits {
		com := commitRequest(commit)
		// Keep an index of commits by SHA.
		oc[commit.SHA] = com
	}

	// Look at merge requests.
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

		if exists {
			// Sanity checks the logins are the same.
			// TODO(#1543): re-enable this code.
			/*if com.Committer.Login != mr.MergeCommit.Committer.Login {
				return checker.CodeReviewData{}, sce.WithMessage(sce.ErrScorecardInternal,
					fmt.Sprintf("commit login (%s) different from merge request commit login (%s)",
						com.Committer.Login, mr.MergeCommit.Committer.Login))
			}*/

			// We have a recent merge request, update it.
			com.MergeRequest = mergeRequest(&mr)
			oc[com.SHA] = com
		}
	}

	for _, v := range oc {
		results = append(results, v)
	}

	if len(results) > 0 {
		return checker.CodeReviewData{DefaultBranchCommits: results}, nil
	}

	return checker.CodeReviewData{DefaultBranchCommits: results}, nil
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
