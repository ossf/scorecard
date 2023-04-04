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
	glClient   *gitlab.Client
	once       *sync.Once
	errSetup   error
	repourl    *repoURL
	commitsRaw []*gitlab.Commit
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
	})

	return handler.errSetup
}

func (handler *commitsHandler) listRawCommits() ([]*gitlab.Commit, error) {
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during commitsHandler.setup: %w", err)
	}

	return handler.commitsRaw, nil
}

// zip combines Commit and MergeRequest information from the GitLab REST API with
// information from the GitLab GraphQL API. The REST API doesn't provide any way to
// get from Commits -> MRs that they were part of or vice-versa (MRs -> commits they
// contain), except through a separate API call. Instead of calling the REST API
// len(commits) times to get the associated MR, we make 3 calls (2 REST, 1 GraphQL).
func (handler *commitsHandler) zip(commitsRaw []*gitlab.Commit, data graphqlData) []clients.Commit {
	commitToMRIID := make(map[string]string) // which mr does a commit belong to?
	for i := range data.Project.MergeRequests.Nodes {
		mr := data.Project.MergeRequests.Nodes[i]
		for _, commit := range mr.Commits.Nodes {
			commitToMRIID[commit.SHA] = mr.IID
		}
		commitToMRIID[mr.MergeCommitSHA] = mr.IID
	}

	iidToMr := make(map[string]clients.PullRequest)
	for i := range data.Project.MergeRequests.Nodes {
		mr := data.Project.MergeRequests.Nodes[i]
		// Two GitLab APIs for reviews (reviews vs. approvals)
		// Use a map to consolidate results from both APIs by the user ID who performed review
		reviews := make(map[string]clients.Review)
		for _, reviewer := range mr.Reviewers.Nodes {
			reviews[reviewer.Username] = clients.Review{
				Author: &clients.User{Login: reviewer.Username},
				State:  "COMMENTED",
			}
		}

		if fmt.Sprintf("%v", mr.IID) != mr.IID {
			continue
		}

		// Check approvers
		for _, approver := range mr.Approvers.Nodes {
			reviews[approver.Username] = clients.Review{
				Author: &clients.User{Login: approver.Username},
				State:  "APPROVED",
			}
			break
		}

		// Check reviewers (sometimes unofficial approvals end up here)
		for _, reviewer := range mr.Reviewers.Nodes {
			if reviewer.MergeRequestInteraction.ReviewState != "REVIEWED" {
				continue
			}
			reviews[reviewer.Username] = clients.Review{
				Author: &clients.User{Login: reviewer.Username},
				State:  "APPROVED",
			}
			break
		}

		vals := []clients.Review{}
		for _, v := range reviews {
			vals = append(vals, v)
		}

		var mrno int
		mrno, err := strconv.Atoi(mr.IID)
		if err != nil {
			mrno = mr.ID.ID
		}

		iidToMr[mr.IID] = clients.PullRequest{
			Number:   mrno,
			MergedAt: mr.MergedAt,
			HeadSHA:  mr.MergeCommitSHA,
			Author:   clients.User{Login: mr.Author.Username, ID: int64(mr.Author.ID.ID)},
			// Labels:   labels,
			Reviews:  vals,
			MergedBy: clients.User{Login: mr.MergedBy.Username, ID: int64(mr.MergedBy.ID.ID)},
		}
	}

	fmt.Println("from commitsRaw==")
	for _, craw := range commitsRaw {
		// print mr iids needed for raw commits
		fmt.Printf("%s ", craw.ID)
	}

	fmt.Println("\nfrom graphql==")
	for commit := range commitToMRIID {
		// print mr iids that we got from graphql
		fmt.Printf("%s ", commit)
	}
	fmt.Println("")
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
