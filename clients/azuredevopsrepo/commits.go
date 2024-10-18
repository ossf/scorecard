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
	"fmt"
	"sync"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/ossf/scorecard/v5/clients"
)

type commitsHandler struct {
	gitClient           git.Client
	ctx                 context.Context
	once                *sync.Once
	errSetup            error
	repourl             *Repo
	commitsRaw          *[]git.GitCommitRef
	pullRequestsRaw     *git.GitPullRequestQuery
	commitDepth         int
	getCommits          fnGetCommits
	getPullRequestQuery fnGetPullRequestQuery
}

func (handler *commitsHandler) init(ctx context.Context, repourl *Repo, commitDepth int) {
	handler.ctx = ctx
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
	handler.commitDepth = commitDepth
	handler.getCommits = handler.gitClient.GetCommits
	handler.getPullRequestQuery = handler.gitClient.GetPullRequestQuery
}

type (
	fnGetCommits          func(ctx context.Context, args git.GetCommitsArgs) (*[]git.GitCommitRef, error)
	fnGetPullRequestQuery func(ctx context.Context, args git.GetPullRequestQueryArgs) (*git.GitPullRequestQuery, error)
)

func (handler *commitsHandler) setup() error {
	handler.once.Do(func() {
		var itemVersion git.GitVersionDescriptor
		if handler.repourl.commitSHA == "HEAD" {
			itemVersion = git.GitVersionDescriptor{
				VersionType: &git.GitVersionTypeValues.Branch,
				Version:     &handler.repourl.defaultBranch,
			}
		} else {
			itemVersion = git.GitVersionDescriptor{
				VersionType: &git.GitVersionTypeValues.Commit,
				Version:     &handler.repourl.commitSHA,
			}
		}

		opt := git.GetCommitsArgs{
			RepositoryId: &handler.repourl.id,
			Top:          &handler.commitDepth,
			SearchCriteria: &git.GitQueryCommitsCriteria{
				ItemVersion: &itemVersion,
			},
		}

		commits, err := handler.getCommits(handler.ctx, opt)
		if err != nil {
			handler.errSetup = fmt.Errorf("request for commits failed with %w", err)
			return
		}

		commitIds := make([]string, len(*commits))
		for i, commit := range *commits {
			commitIds[i] = *commit.CommitId
		}

		pullRequestQuery := git.GetPullRequestQueryArgs{
			RepositoryId: &handler.repourl.id,
			Queries: &git.GitPullRequestQuery{
				Queries: &[]git.GitPullRequestQueryInput{
					{
						Type:  &git.GitPullRequestQueryTypeValues.Commit,
						Items: &commitIds,
					},
				},
			},
		}
		pullRequests, err := handler.getPullRequestQuery(handler.ctx, pullRequestQuery)
		if err != nil {
			handler.errSetup = fmt.Errorf("request for pull requests failed with %w", err)
			return
		}

		handler.commitsRaw = commits
		handler.pullRequestsRaw = pullRequests

		handler.errSetup = nil
	})
	return handler.errSetup
}

func (handler *commitsHandler) listCommits() ([]clients.Commit, error) {
	err := handler.setup()
	if err != nil {
		return nil, fmt.Errorf("error during commitsHandler.setup: %w", err)
	}

	commits := make([]clients.Commit, len(*handler.commitsRaw))
	for i, commit := range *handler.commitsRaw {
		commits[i] = clients.Commit{
			SHA:           *commit.CommitId,
			Message:       *commit.Comment,
			CommittedDate: commit.Committer.Date.Time,
			Committer: clients.User{
				Login: *commit.Committer.Name,
			},
		}
	}

	// Associate pull requests with commits
	pullRequests, err := handler.listPullRequests()
	if err != nil {
		return nil, fmt.Errorf("error during commitsHandler.listPullRequests: %w", err)
	}

	for i, commit := range commits {
		associatedPullRequest, ok := pullRequests[commit.SHA]
		if !ok {
			continue
		}

		commits[i].AssociatedMergeRequest = associatedPullRequest
	}

	return commits, nil
}

func (handler *commitsHandler) listPullRequests() (map[string]clients.PullRequest, error) {
	err := handler.setup()
	if err != nil {
		return nil, fmt.Errorf("error during commitsHandler.setup: %w", err)
	}

	pullRequests := make(map[string]clients.PullRequest)
	for commit, azdoPullRequests := range (*handler.pullRequestsRaw.Results)[0] {
		if len(azdoPullRequests) == 0 {
			continue
		}

		if len(azdoPullRequests) > 1 {
			return nil, fmt.Errorf("expected 1 pull request for commit %s, got %d", commit, len(azdoPullRequests))
		}

		azdoPullRequest := azdoPullRequests[0]

		pullRequests[commit] = clients.PullRequest{
			Number: *azdoPullRequest.PullRequestId,
			Author: clients.User{
				Login: *azdoPullRequest.CreatedBy.DisplayName,
			},
			HeadSHA:  *azdoPullRequest.LastMergeSourceCommit.CommitId,
			MergedAt: azdoPullRequest.ClosedDate.Time,
		}
	}

	return pullRequests, nil
}
