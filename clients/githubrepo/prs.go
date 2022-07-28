package githubrepo

import (
	"context"
	"fmt"
	"sync"

	sce "github.com/ossf/scorecard/v4/errors"

	"github.com/google/go-github/v38/github"
	"github.com/ossf/scorecard/v4/clients"
)

const (
	mergedPRsThreshold = 30
)

type PRsHandler struct {
	client  *github.Client
	ctx     context.Context
	repourl *repoURL
	once    *sync.Once
}

func (handler *PRsHandler) init(ctx context.Context, repourl *repoURL) {
	handler.ctx = ctx
	handler.repourl = repourl
	handler.once = new(sync.Once)

}

func (handler *PRsHandler) listPullRequests() ([]clients.PullRequest, error) {
	opts := &github.PullRequestListOptions{
		State:       "closed",
		ListOptions: github.ListOptions{Page: 1, PerPage: 30},
	}
	var pullRequests []clients.PullRequest

	mergedCount := 0
	for mergedCount <= mergedPRsThreshold {
		PRs, _, err := handler.client.PullRequests.List(handler.ctx, handler.repourl.owner, handler.repourl.repo, opts)
		if err != nil {
			return nil, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("ListMergedPRs: %v", err))
		}

		if len(PRs) == 0 {
			// ran out of PRs
			break
		}

		for _, pr := range PRs {
			if !pr.GetMergedAt().IsZero() {
				// PR is merged
				pullRequests = append(pullRequests, clients.PullRequest{
					Author: clients.User{Login: pr.User.GetLogin(), ID: *pr.User.ID},
				})
				mergedCount += 1
			}
		}
		opts.ListOptions.Page += 1
	}

	return pullRequests, nil
}
