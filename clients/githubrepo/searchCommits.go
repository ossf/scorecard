package githubrepo

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v38/github"

	"github.com/ossf/scorecard/v4/clients"
)

type searchCommitsHandler struct {
	ghClient *github.Client
	ctx      context.Context
	repourl  *repoURL
}

func (handler *searchCommitsHandler) init(ctx context.Context, repourl *repoURL) {
	handler.ctx = ctx
	handler.repourl = repourl
}

func (handler *searchCommitsHandler) search(request clients.SearchCommitsOptions) (clients.SearchCommitsResponse, error) {
	if !strings.EqualFold(handler.repourl.commitSHA, clients.HeadSHA) {
		return clients.SearchCommitsResponse{}, fmt.Errorf(
			"%w: Search only supported for HEAD queries", clients.ErrUnsupportedFeature)
	}
	query, err := handler.buildQuery(request)
	if err != nil {
		return clients.SearchCommitsResponse{}, fmt.Errorf("handler.buildQuery: %w", err)
	}

	resp, _, err := handler.ghClient.Search.Commits(handler.ctx, query, &github.SearchOptions{ListOptions: github.ListOptions{PerPage: 100}})
	if err != nil {
		return clients.SearchCommitsResponse{}, fmt.Errorf("Search.Code: %w", err)
	}

	return searchCommitsResponseFrom(resp), nil
}

func (handler *searchCommitsHandler) buildQuery(request clients.SearchCommitsOptions) (string, error) {
	if request.Author == "" {
		return "", fmt.Errorf("%w", errEmptyQuery)
	}
	var queryBuilder strings.Builder
	if _, err := queryBuilder.WriteString(
		fmt.Sprintf("repo:%s/%s author:%s",
			handler.repourl.owner, handler.repourl.repo,
			request.Author)); err != nil {
		return "", fmt.Errorf("WriteString: %w", err)
	}

	return queryBuilder.String(), nil
}

func searchCommitsResponseFrom(resp *github.CommitsSearchResult) clients.SearchCommitsResponse {
	var ret clients.SearchCommitsResponse
	ret.Hits = resp.GetTotal()
	for _, result := range resp.Commits {
		ret.Results = append(ret.Results, clients.SearchCommitResult{
			ID: *result.Author.ID,
		})
	}
	return ret
}
