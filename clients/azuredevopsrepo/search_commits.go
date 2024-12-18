// Copyright 2024 OpenSSF Scorecard Authors
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

package azuredevopsrepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"

	"github.com/ossf/scorecard/v5/clients"
)

var errMoreThanOnePullRequest = errors.New("more than one pull request found for a commit")

type searchCommitsHandler struct {
	ctx                 context.Context
	repourl             *Repo
	gitClient           git.Client
	getCommits          fnGetCommits
	getPullRequestQuery fnGetPullRequestQuery
}

func (s *searchCommitsHandler) init(ctx context.Context, repourl *Repo) {
	s.ctx = ctx
	s.repourl = repourl
	s.getCommits = s.gitClient.GetCommits
	s.getPullRequestQuery = s.gitClient.GetPullRequestQuery
}

func (s *searchCommitsHandler) searchCommits(searchOptions clients.SearchCommitsOptions) ([]clients.Commit, error) {
	commits := make([]clients.Commit, 0)
	commitsPageSize := 1000
	skip := 0

	var itemVersion git.GitVersionDescriptor
	if s.repourl.commitSHA == headCommit {
		itemVersion = git.GitVersionDescriptor{
			VersionType: &git.GitVersionTypeValues.Branch,
			Version:     &s.repourl.defaultBranch,
		}
	} else {
		itemVersion = git.GitVersionDescriptor{
			VersionType: &git.GitVersionTypeValues.Commit,
			Version:     &s.repourl.commitSHA,
		}
	}

	for {
		args := git.GetCommitsArgs{
			RepositoryId: &s.repourl.id,
			SearchCriteria: &git.GitQueryCommitsCriteria{
				ItemVersion: &itemVersion,
				Author:      &searchOptions.Author,
				Top:         &commitsPageSize,
				Skip:        &skip,
			},
		}
		response, err := s.getCommits(s.ctx, args)
		if err != nil {
			return nil, fmt.Errorf("failed to get commits: %w", err)
		}

		if response == nil || len(*response) == 0 {
			break
		}

		for i := range *response {
			commit := &(*response)[i]
			pullRequest, err := s.getAssociatedPullRequest(commit)
			if err != nil {
				return nil, fmt.Errorf("failed to get associated pull request: %w", err)
			}

			commits = append(commits, clients.Commit{
				SHA:           *commit.CommitId,
				Message:       *commit.Comment,
				CommittedDate: commit.Committer.Date.Time,
				Committer: clients.User{
					Login: *commit.Committer.Email,
				},
				AssociatedMergeRequest: pullRequest,
			})
		}

		if len(*response) < commitsPageSize {
			break
		}

		skip += commitsPageSize
	}

	return commits, nil
}

func (s *searchCommitsHandler) getAssociatedPullRequest(commit *git.GitCommitRef) (clients.PullRequest, error) {
	query, err := s.getPullRequestQuery(s.ctx, git.GetPullRequestQueryArgs{
		RepositoryId: &s.repourl.id,
		Queries: &git.GitPullRequestQuery{
			Queries: &[]git.GitPullRequestQueryInput{
				{
					Items: &[]string{*commit.CommitId},
					Type:  &git.GitPullRequestQueryTypeValues.Commit,
				},
			},
		},
	})
	if err != nil {
		return clients.PullRequest{}, err
	}

	if query == nil || query.Results == nil {
		return clients.PullRequest{}, nil
	}

	results := *query.Results
	if len(results) == 0 {
		return clients.PullRequest{}, nil
	}

	if len(results) > 1 {
		return clients.PullRequest{}, errMoreThanOnePullRequest
	}

	// TODO: Azure DevOps API returns a list of pull requests for a commit.
	// Scorecard currently only supports one pull request per commit.
	result := results[0]
	pullRequests, ok := result[*commit.CommitId]
	if !ok || len(pullRequests) == 0 {
		return clients.PullRequest{}, nil
	}

	return clients.PullRequest{
		Number: *pullRequests[0].PullRequestId,
	}, nil
}
