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
	"strings"
	"sync"
	"time"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"

	"github.com/ossf/scorecard/v5/clients"
)

var (
	errMultiplePullRequests = errors.New("expected 1 pull request for commit, got multiple")
	errRefNotFound          = errors.New("ref not found")
	errInvalidPullRequest   = errors.New("invalid pull request data")
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
	getPullRequest       fnGetPullRequest
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
	c.getPullRequest = c.gitClient.GetPullRequest
	c.getPullRequestQuery = c.gitClient.GetPullRequestQuery
	c.getFirstCommit = c.gitClient.GetCommits
	c.getRefs = c.gitClient.GetRefs
	c.getStatuses = c.gitClient.GetStatuses
}

type (
	fnGetCommits          func(ctx context.Context, args git.GetCommitsArgs) (*[]git.GitCommitRef, error)
	fnGetPullRequest      func(ctx context.Context, args git.GetPullRequestArgs) (*git.GitPullRequest, error)
	fnGetPullRequestQuery func(ctx context.Context, args git.GetPullRequestQueryArgs) (*git.GitPullRequestQuery, error)
	fnGetFirstCommit      func(ctx context.Context, args git.GetCommitsArgs) (*[]git.GitCommitRef, error)
	fnGetRefs             func(ctx context.Context, args git.GetRefsArgs) (*git.GetRefsResponseValue, error)
	fnGetStatuses         func(ctx context.Context, args git.GetStatusesArgs) (*[]git.GitStatus, error)
)

func (c *commitsHandler) setup() error {
	c.once.Do(func() {
		var itemVersion git.GitVersionDescriptor
		if c.repourl.commitSHA == headCommit {
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
		if err := c.loadPullRequestDetails(pullRequests); err != nil {
			c.errSetup = err
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

func (c *commitsHandler) loadPullRequestDetails(query *git.GitPullRequestQuery) error {
	if query == nil || query.Results == nil || len(*query.Results) != 1 {
		return fmt.Errorf("%w: missing query results", errInvalidPullRequest)
	}
	detailsByID := make(map[int]*git.GitPullRequest)
	for commit, matches := range (*query.Results)[0] {
		if len(matches) > 1 {
			return errMultiplePullRequests
		}
		for i := range matches {
			if matches[i].PullRequestId == nil || *matches[i].PullRequestId <= 0 {
				return fmt.Errorf("%w: missing PR ID for commit %s", errInvalidPullRequest, commit)
			}
			id := *matches[i].PullRequestId
			details, ok := detailsByID[id]
			if !ok {
				var err error
				details, err = c.getPullRequest(c.ctx, git.GetPullRequestArgs{
					RepositoryId:  &c.repourl.id,
					Project:       &c.repourl.project,
					PullRequestId: &id,
				})
				if err != nil {
					return fmt.Errorf("request for pull request %d failed: %w", id, err)
				}
				if details == nil || details.PullRequestId == nil ||
					*details.PullRequestId != id || details.Status == nil {
					return fmt.Errorf("%w: incomplete details for PR %d", errInvalidPullRequest, id)
				}
				if *details.Status == git.PullRequestStatusValues.Completed &&
					(details.ClosedDate == nil || details.ClosedDate.Time.IsZero() ||
						details.LastMergeCommit == nil || details.LastMergeCommit.CommitId == nil ||
						*details.LastMergeCommit.CommitId == "") {
					return fmt.Errorf("%w: missing merge metadata for PR %d", errInvalidPullRequest, id)
				}
				detailsByID[id] = details
			}
			matches[i] = *details
		}
	}
	return nil
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

		pr := clients.PullRequest{
			Number: *azdoPullRequest.PullRequestId,
		}
		if azdoPullRequest.CreatedBy != nil {
			pr.Author.Login = identityLogin(azdoPullRequest.CreatedBy.Id)
		}
		if *azdoPullRequest.Status == git.PullRequestStatusValues.Completed {
			pr.HeadSHA = *azdoPullRequest.LastMergeCommit.CommitId
			pr.MergedAt = azdoPullRequest.ClosedDate.Time
			if azdoPullRequest.ClosedBy != nil &&
				(azdoPullRequest.ClosedBy.IsContainer == nil || !*azdoPullRequest.ClosedBy.IsContainer) {
				pr.MergedBy.Login = identityLogin(azdoPullRequest.ClosedBy.Id)
			}
		}
		if azdoPullRequest.Reviewers != nil {
			for i := range *azdoPullRequest.Reviewers {
				reviewer := &(*azdoPullRequest.Reviewers)[i]
				login := identityLogin(reviewer.Id)
				if login == "" || reviewer.Vote == nil ||
					(reviewer.IsContainer != nil && *reviewer.IsContainer) {
					continue
				}
				var state string
				switch *reviewer.Vote {
				case 10, 5:
					state = "APPROVED"
				case -5:
					state = "COMMENTED"
				case -10:
					state = "CHANGES_REQUESTED"
				default:
					continue
				}
				pr.Reviews = append(pr.Reviews, clients.Review{
					Author: &clients.User{Login: login},
					State:  state,
				})
			}
		}
		pullRequests[commit] = pr
	}

	return pullRequests, nil
}

func identityLogin(id *string) string {
	if id == nil {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(*id))
}

func (c *commitsHandler) getFirstCommitCreatedAt() (time.Time, error) {
	if err := c.setup(); err != nil {
		return time.Time{}, fmt.Errorf("error during commitsHandler.setup: %w", err)
	}

	return c.firstCommitCreatedAt, nil
}

func (c *commitsHandler) listStatuses(ref string) ([]clients.Status, error) {
	matched, err := regexp.MatchString("^[0-9a-fA-F]{40}$", ref)
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
