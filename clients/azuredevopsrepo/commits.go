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
	"regexp"
	"sync"
	"time"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"

	"github.com/ossf/scorecard/v5/clients"
)

var (
	errMultiplePullRequests = errors.New("expected 1 pull request for commit, got multiple")
	errRefNotFound          = errors.New("ref not found")
)

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
	getRefs              fnGetRefs
	getStatuses          fnGetStatuses
	commitDepth          int
}

func (c *commitsHandler) init(ctx context.Context, repourl *Repo, commitDepth int) {
	c.ctx = ctx
	c.repourl = repourl
	c.errSetup = nil
	c.once = new(sync.Once)
	c.commitDepth = commitDepth
	c.getCommits = c.gitClient.GetCommits
	c.getPullRequestQuery = c.gitClient.GetPullRequestQuery
	c.getFirstCommit = c.gitClient.GetCommits
	c.getRefs = c.gitClient.GetRefs
	c.getStatuses = c.gitClient.GetStatuses
}

type (
	fnGetCommits          func(ctx context.Context, args git.GetCommitsArgs) (*[]git.GitCommitRef, error)
	fnGetPullRequestQuery func(ctx context.Context, args git.GetPullRequestQueryArgs) (*git.GitPullRequestQuery, error)
	fnGetFirstCommit      func(ctx context.Context, args git.GetCommitsArgs) (*[]git.GitCommitRef, error)
	fnGetRefs             func(ctx context.Context, args git.GetRefsArgs) (*git.GetRefsResponseValue, error)
	fnGetStatuses         func(ctx context.Context, args git.GetStatusesArgs) (*[]git.GitStatus, error)
)

func (c *commitsHandler) setup() error {
	c.once.Do(func() {
		var itemVersion git.GitVersionDescriptor
		if c.repourl.commitSHA == HeadCommit {
			itemVersion = git.GitVersionDescriptor{
				VersionType: &git.GitVersionTypeValues.Branch,
				Version:     &c.repourl.defaultBranch,
			}
		} else {
			itemVersion = git.GitVersionDescriptor{
				VersionType: &git.GitVersionTypeValues.Commit,
				Version:     &c.repourl.commitSHA,
			}
		}

		opt := git.GetCommitsArgs{
			RepositoryId: &c.repourl.id,
			Top:          &c.commitDepth,
			SearchCriteria: &git.GitQueryCommitsCriteria{
				ItemVersion: &itemVersion,
			},
		}

		commits, err := c.getCommits(c.ctx, opt)
		if err != nil {
			c.errSetup = fmt.Errorf("request for commits failed with %w", err)
			return
		}

		commitIds := make([]string, len(*commits))
		for i := range *commits {
			commitIds[i] = *(*commits)[i].CommitId
		}

		pullRequestQuery := git.GetPullRequestQueryArgs{
			RepositoryId: &c.repourl.id,
			Queries: &git.GitPullRequestQuery{
				Queries: &[]git.GitPullRequestQueryInput{
					{
						Type:  &git.GitPullRequestQueryTypeValues.LastMergeCommit,
						Items: &commitIds,
					},
				},
			},
		}
		pullRequests, err := c.getPullRequestQuery(c.ctx, pullRequestQuery)
		if err != nil {
			c.errSetup = fmt.Errorf("request for pull requests failed with %w", err)
			return
		}

		switch {
		case len(*commits) == 0:
			c.firstCommitCreatedAt = time.Time{}
		case len(*commits) < c.commitDepth:
			c.firstCommitCreatedAt = (*commits)[len(*commits)-1].Committer.Date.Time
		default:
			firstCommit, err := c.getFirstCommit(c.ctx, git.GetCommitsArgs{
				RepositoryId: &c.repourl.id,
				SearchCriteria: &git.GitQueryCommitsCriteria{
					Top:                    &[]int{1}[0],
					ShowOldestCommitsFirst: &[]bool{true}[0],
					ItemVersion: &git.GitVersionDescriptor{
						VersionType: &git.GitVersionTypeValues.Branch,
						Version:     &c.repourl.defaultBranch,
					},
				},
			})
			if err != nil {
				c.errSetup = fmt.Errorf("request for first commit failed with %w", err)
				return
			}

			c.firstCommitCreatedAt = (*firstCommit)[0].Committer.Date.Time
		}

		c.commitsRaw = commits
		c.pullRequestsRaw = pullRequests

		c.errSetup = nil
	})
	return c.errSetup
}

func (c *commitsHandler) listCommits() ([]clients.Commit, error) {
	err := c.setup()
	if err != nil {
		return nil, fmt.Errorf("error during commitsHandler.setup: %w", err)
	}

	commits := make([]clients.Commit, len(*c.commitsRaw))
	for i := range *c.commitsRaw {
		commit := &(*c.commitsRaw)[i]
		commits[i] = clients.Commit{
			SHA:           *commit.CommitId,
			Message:       *commit.Comment,
			CommittedDate: commit.Committer.Date.Time,
			Committer: clients.User{
				Login: *commit.Committer.Email,
			},
		}
	}

	// Associate pull requests with commits
	pullRequests, err := c.listPullRequests()
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

func (c *commitsHandler) listPullRequests() (map[string]clients.PullRequest, error) {
	err := c.setup()
	if err != nil {
		return nil, fmt.Errorf("error during commitsHandler.setup: %w", err)
	}

	pullRequests := make(map[string]clients.PullRequest)
	for commit, azdoPullRequests := range (*c.pullRequestsRaw.Results)[0] {
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
			HeadSHA:  *azdoPullRequest.LastMergeCommit.CommitId,
			MergedAt: azdoPullRequest.ClosedDate.Time,
		}
	}

	return pullRequests, nil
}

func (c *commitsHandler) getFirstCommitCreatedAt() (time.Time, error) {
	if err := c.setup(); err != nil {
		return time.Time{}, fmt.Errorf("error during commitsHandler.setup: %w", err)
	}

	return c.firstCommitCreatedAt, nil
}

func (c *commitsHandler) listStatuses(ref string) ([]clients.Status, error) {
	matched, err := regexp.MatchString("^[0-9a-f]{5,40}$", ref)
	if err != nil {
		return nil, fmt.Errorf("error matching ref: %w", err)
	}
	if matched {
		return c.statusFromCommit(ref)
	} else {
		return c.statusFromHead(ref)
	}
}

func (c *commitsHandler) statusFromHead(ref string) ([]clients.Status, error) {
	includeStatuses := true
	filter := fmt.Sprintf("heads/%s", ref)
	args := git.GetRefsArgs{
		RepositoryId:       &c.repourl.id,
		Filter:             &filter,
		IncludeStatuses:    &includeStatuses,
		LatestStatusesOnly: &[]bool{true}[0],
	}
	response, err := c.getRefs(c.ctx, args)
	if err != nil {
		return nil, fmt.Errorf("error getting refs: %w", err)
	}

	if len(response.Value) != 1 {
		return nil, errRefNotFound
	}
	statuses := response.Value[0].Statuses
	if statuses == nil {
		return []clients.Status{}, nil
	}

	result := make([]clients.Status, len(*statuses))
	for i, status := range *statuses {
		result[i] = clients.Status{
			State:     convertAzureDevOpsStatus(&status),
			Context:   *status.Context.Name,
			URL:       *status.TargetUrl,
			TargetURL: *status.TargetUrl,
		}
	}

	return result, nil
}

func (c *commitsHandler) statusFromCommit(ref string) ([]clients.Status, error) {
	args := git.GetStatusesArgs{
		RepositoryId: &c.repourl.id,
		CommitId:     &ref,
		LatestOnly:   &[]bool{true}[0],
	}
	response, err := c.getStatuses(c.ctx, args)
	if err != nil {
		return nil, fmt.Errorf("error getting statuses: %w", err)
	}

	result := make([]clients.Status, len(*response))
	for i, status := range *response {
		result[i] = clients.Status{
			Context:   *status.Context.Name,
			State:     convertAzureDevOpsStatus(&status),
			URL:       *status.TargetUrl,
			TargetURL: *status.TargetUrl,
		}
	}

	return result, nil
}

func convertAzureDevOpsStatus(s *git.GitStatus) string {
	switch *s.State {
	case "succeeded":
		return "success"
	case "failed", "error":
		return "failure"
	case "notApplicable", "notSet", "pending":
		return "neutral"
	default:
		return string(*s.State)
	}
}
