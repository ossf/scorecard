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
	"sync"
	"time"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"

	"github.com/ossf/scorecard/v5/clients"
)

var errMultiplePullRequests = errors.New("expected 1 pull request for commit, got multiple")

type commitsHandler struct {
	gitClient            git.Client
	ctx                  context.Context
	errSetup             error
	once                 *sync.Once
	repourl              *Repo
	commitsRaw           *[]git.GitCommitRef
	pullRequestsRaw      *git.GitPullRequestQuery
	firstCommitCreatedAt time.Time
	getCommits           fnGetCommits
	getPullRequestQuery  fnGetPullRequestQuery
	getFirstCommit       fnGetFirstCommit
	commitDepth          int
}

func (handler *commitsHandler) init(ctx context.Context, repourl *Repo, commitDepth int) {
	handler.ctx = ctx
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
	handler.commitDepth = commitDepth
	handler.getCommits = handler.gitClient.GetCommits
	handler.getPullRequestQuery = handler.gitClient.GetPullRequestQuery
	handler.getFirstCommit = handler.gitClient.GetCommits
}

type (
	fnGetCommits          func(ctx context.Context, args git.GetCommitsArgs) (*[]git.GitCommitRef, error)
	fnGetPullRequestQuery func(ctx context.Context, args git.GetPullRequestQueryArgs) (*git.GitPullRequestQuery, error)
	fnGetFirstCommit      func(ctx context.Context, args git.GetCommitsArgs) (*[]git.GitCommitRef, error)
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
		for i := range *commits {
			commitIds[i] = *(*commits)[i].CommitId
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

		switch {
		case len(*commits) == 0:
			handler.firstCommitCreatedAt = time.Time{}
		case len(*commits) < handler.commitDepth:
			handler.firstCommitCreatedAt = (*commits)[len(*commits)-1].Committer.Date.Time
		default:
			firstCommit, err := handler.getFirstCommit(handler.ctx, git.GetCommitsArgs{
				RepositoryId: &handler.repourl.id,
				SearchCriteria: &git.GitQueryCommitsCriteria{
					Top:                    &[]int{1}[0],
					ShowOldestCommitsFirst: &[]bool{true}[0],
					ItemVersion: &git.GitVersionDescriptor{
						VersionType: &git.GitVersionTypeValues.Branch,
						Version:     &handler.repourl.defaultBranch,
					},
				},
			})
			if err != nil {
				handler.errSetup = fmt.Errorf("request for first commit failed with %w", err)
				return
			}

			handler.firstCommitCreatedAt = (*firstCommit)[0].Committer.Date.Time
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
	for i := range *handler.commitsRaw {
		commit := &(*handler.commitsRaw)[i]
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

	for i := range commits {
		commit := &commits[i]
		associatedPullRequest, ok := pullRequests[commit.SHA]
		if !ok {
			continue
		}

		commit.AssociatedMergeRequest = associatedPullRequest
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
			return nil, errMultiplePullRequests
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

func (handler *commitsHandler) getFirstCommitCreatedAt() (time.Time, error) {
	if err := handler.setup(); err != nil {
		return time.Time{}, fmt.Errorf("error during commitsHandler.setup: %w", err)
	}

	return handler.firstCommitCreatedAt, nil
}
