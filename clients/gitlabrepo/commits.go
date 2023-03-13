// Copyright 2022 OpenSSF Scorecard Authors
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

package gitlabrepo

import (
	"fmt"
	"strings"
	"sync"

	"github.com/xanzy/go-gitlab"

	"github.com/ossf/scorecard/v4/clients"
)

type commitsHandler struct {
	glClient *gitlab.Client
	once     *sync.Once
	errSetup error
	repourl  *repoURL
	commits  []clients.Commit
}

func (handler *commitsHandler) init(repourl *repoURL) {
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
}

//nolint:gocognit
func (handler *commitsHandler) setup() error {
	handler.once.Do(func() {
		commits, _, err := handler.glClient.Commits.ListCommits(handler.repourl.project, &gitlab.ListCommitsOptions{})
		if err != nil {
			handler.errSetup = fmt.Errorf("request for commits failed with %w", err)
			return
		}

		for _, commit := range commits {
			var err error

			// Commits are able to be a part of multiple merge requests, but the only one that will be important
			// here is the earliest one.
			mergeRequests, _, err := handler.glClient.Commits.ListMergeRequestsByCommit(handler.repourl.project, commit.ID)
			if err != nil {
				handler.errSetup = fmt.Errorf("unable to find merge requests associated with commit: %w", err)
				return
			}
			var mergeRequest *gitlab.MergeRequest
			if len(mergeRequests) > 0 {
				mergeRequest = mergeRequests[0]
				for i := range mergeRequests {
					if mergeRequests[i] == nil || mergeRequests[i].MergedAt == nil {
						continue
					}
					if mergeRequests[i].CreatedAt.Before(*mergeRequest.CreatedAt) {
						mergeRequest = mergeRequests[i]
					}
				}
			}

			if mergeRequest == nil || mergeRequest.MergedAt == nil {
				handler.commits = append(handler.commits, clients.Commit{
					CommittedDate: *commit.CommittedDate,
					Message:       commit.Message,
					SHA:           commit.ID,
				})
				continue
			}

			// Two GitLab APIs for reviews (reviews vs. approvals)
			// Use a map to consolidate results from both APIs by the user ID who performed the
			reviews := make(map[int]clients.Review)
			for _, reviewer := range mergeRequest.Reviewers {
				if reviewer.State == "unreviewed" {
					continue
				}
				reviews[reviewer.ID] = clients.Review{
					Author: &clients.User{Login: reviewer.Username, ID: int64(reviewer.ID)},
					State:  "COMMENTED",
				}
			}

			approvals, _, err := handler.glClient.MergeRequests.GetMergeRequestApprovals(
				handler.repourl.project, mergeRequest.IID,
			)
			if err != nil {
				handler.errSetup = fmt.Errorf("unable to find approvers for commit: %w", err)
				return
			}

			if approvals != nil {
				for _, approver := range approvals.ApprovedBy {
					reviews[approver.User.ID] = clients.Review{
						Author: &clients.User{Login: approver.User.Username, ID: int64(approver.User.ID)},
						State:  "APPROVED",
					}
				}
			}

			vals := make([]clients.Review, 0, len(reviews))
			for _, v := range reviews {
				vals = append(vals, v)
			}

			// Casting the Labels into []clients.Label.
			var labels []clients.Label
			for _, label := range mergeRequest.Labels {
				labels = append(labels, clients.Label{
					Name: label,
				})
			}

			// append the commits to the handler.
			handler.commits = append(handler.commits,
				clients.Commit{
					CommittedDate: *commit.CommittedDate,
					Message:       commit.Message,
					SHA:           commit.ID,
					AssociatedMergeRequest: clients.PullRequest{
						Number:   mergeRequest.ID,
						MergedAt: *mergeRequest.MergedAt,
						HeadSHA:  mergeRequest.SHA,
						Author:   clients.User{Login: mergeRequest.Author.Username, ID: int64(mergeRequest.Author.ID)},
						Labels:   labels,
						Reviews:  vals,
						MergedBy: clients.User{Login: mergeRequest.MergedBy.Username, ID: int64(mergeRequest.MergedBy.ID)},
					},
				})
		}
	})

	return handler.errSetup
}

func (handler *commitsHandler) listCommits() ([]clients.Commit, error) {
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during commitsHandler.setup: %w", err)
	}

	return handler.commits, nil
}

// Expected email form: <firstname>.<lastname>@<namespace>.com.
func parseEmailToName(email string) string {
	s := strings.Split(email, ".")
	firstName := s[0]
	lastName := strings.Split(s[1], "@")[0]
	return firstName + " " + lastName
}
