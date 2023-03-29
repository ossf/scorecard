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
	"strconv"
	"strings"
	"sync"

	"github.com/xanzy/go-gitlab"

	"github.com/ossf/scorecard/v4/clients"
)

type commitsHandler struct {
	glClient         *gitlab.Client
	once             *sync.Once
	errSetup         error
	repourl          *repoURL
	commitsRaw       []*gitlab.Commit
	mergeRequestsRaw []*gitlab.MergeRequest
}

func (handler *commitsHandler) init(repourl *repoURL) {
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
}

func (handler *commitsHandler) setup() error {
	handler.once.Do(func() {
		commits, _, err := handler.glClient.Commits.ListCommits(handler.repourl.projectID, &gitlab.ListCommitsOptions{})
		if err != nil {
			handler.errSetup = fmt.Errorf("request for commits failed with %w", err)
			return
		}
		handler.commitsRaw = commits

		state := "merged"
		scope := "all"
		lmo := &gitlab.ListProjectMergeRequestsOptions{
			State: &state,
			Scope: &scope,
		}

		mergeRequests, _, err := handler.glClient.MergeRequests.ListProjectMergeRequests(handler.repourl.projectID, lmo)
		if err != nil {
			handler.errSetup = fmt.Errorf("request for merge requests failed with %w", err)
			return
		}
		handler.mergeRequestsRaw = mergeRequests[:12]
	})

	return handler.errSetup
}

func (handler *commitsHandler) listRawCommits() ([]*gitlab.Commit, error) {
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during commitsHandler.setup: %w", err)
	}

	return handler.commitsRaw, nil
}

func (handler *commitsHandler) listRawMergeRequests() ([]*gitlab.MergeRequest, error) {
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during commitsHandler.setup: %w", err)
	}

	return handler.mergeRequestsRaw, nil
}

// zip combines Commit and MergeRequest information from the GitLab REST API with
// information from the GitLab GraphQL API. The REST API doesn't provide any way to
// get from Commits -> MRs that they were part of or vice-versa (MRs -> commits they
// contain), except through a separate API call. Instead of calling the REST API
// len(commits) times to get the associated MR, we make 3 calls (2 REST, 1 GraphQL).
func (handler *commitsHandler) zip(
	mrsRaw []*gitlab.MergeRequest, commitsRaw []*gitlab.Commit, graphMRs []graphQlMergeRequest,
) []clients.Commit {
	commitToMRIID := make(map[string]string)
	for _, mr := range graphMRs {
		for _, commit := range mr.Commits.Nodes {
			commitToMRIID[commit.SHA] = mr.IID
		}
	}

	iidToMr := make(map[string]clients.PullRequest)
	for i := range mrsRaw {
		mr := mrsRaw[i]
		// Two GitLab APIs for reviews (reviews vs. approvals)
		// Use a map to consolidate results from both APIs by the user ID who performed the
		reviews := make(map[int]clients.Review)
		for _, reviewer := range mr.Reviewers {
			reviews[reviewer.ID] = clients.Review{
				Author: &clients.User{Login: reviewer.Username, ID: int64(reviewer.ID)},
				State:  "COMMENTED",
			}
		}

		for _, graphMr := range graphMRs {
			if fmt.Sprintf("%d", mr.IID) != graphMr.IID {
				continue
			}

			// Check approvers
			for _, approver := range graphMr.Approvers.Nodes {
				var approverID int64
				var err error
				approverID, err = strconv.ParseInt(approver.ID, 10, 64)
				if err != nil {
					continue
				}
				reviews[int(approverID)] = clients.Review{
					Author: &clients.User{Login: approver.Username, ID: approverID},
					State:  "APPROVED",
				}
				break
			}

			// Check reviewers (sometimes unofficial approvals end up here)
			for _, reviewer := range graphMr.Reviewers.Nodes {
				var reviewerID int64
				var err error
				reviewerID, err = strconv.ParseInt(reviewer.ID, 10, 64)
				if err != nil {
					continue
				}
				if reviewer.MergeRequestInteraction.ReviewState != "REVIEWED" {
					continue
				}
				reviews[int(reviewerID)] = clients.Review{
					Author: &clients.User{Login: reviewer.Username, ID: reviewerID},
					State:  "APPROVED",
				}
				break
			}

		}

		vals := []clients.Review{}
		for _, v := range reviews {
			vals = append(vals, v)
		}

		// Casting the Labels into []clients.Label.
		labels := []clients.Label{}
		for _, label := range mr.Labels {
			labels = append(labels, clients.Label{
				Name: label,
			})
		}

		iidToMr[fmt.Sprintf("%d", mr.IID)] = clients.PullRequest{
			Number:   mr.ID,
			MergedAt: *mr.MergedAt,
			HeadSHA:  mr.SHA,
			Author:   clients.User{Login: mr.Author.Username, ID: int64(mr.Author.ID)},
			Labels:   labels,
			Reviews:  vals,
			MergedBy: clients.User{Login: mr.MergedBy.Username, ID: int64(mr.MergedBy.ID)},
		}
	}

	// Associate Merge Requests with Commits based on the GitLab Merge Request IID
	commits := []clients.Commit{}
	for _, cRaw := range commitsRaw {
		// Get IID of Merge Request that this commit was merged as part of
		mrIID := commitToMRIID[cRaw.ID]
		associatedMr := iidToMr[mrIID]

		commits = append(commits,
			clients.Commit{
				CommittedDate:          *cRaw.CommittedDate,
				Message:                cRaw.Message,
				SHA:                    cRaw.ID,
				AssociatedMergeRequest: associatedMr,
			})
	}

	return commits
}

// Expected email form: <firstname>.<lastname>@<namespace>.com.
func parseEmailToName(email string) string {
	s := strings.Split(email, ".")
	firstName := s[0]
	lastName := strings.Split(s[1], "@")[0]
	return firstName + " " + lastName
}
