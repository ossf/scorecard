// Copyright 2022 Security Scorecard Authors
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

func (handler *searchCommitsHandler) search(request clients.SearchCommitsOptions) ([]clients.Commit, error) {
	if !strings.EqualFold(handler.repourl.commitSHA, clients.HeadSHA) {
		return nil, fmt.Errorf(
			"%w: Search only supported for HEAD queries", clients.ErrUnsupportedFeature)
	}
	query, err := handler.buildQuery(request)
	if err != nil {
		return nil, fmt.Errorf("handler.buildQuery: %w", err)
	}

	resp, _, err := handler.ghClient.Search.Commits(handler.ctx, query, &github.SearchOptions{ListOptions: github.ListOptions{PerPage: 100}})
	if err != nil {
		return nil, fmt.Errorf("Search.Code: %w", err)
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

func searchCommitsResponseFrom(resp *github.CommitsSearchResult) []clients.Commit {
	var ret []clients.Commit
	for _, result := range resp.Commits {
		ret = append(ret, clients.Commit{
			Committer: clients.User{ID: *result.Author.ID},
		})
	}
	return ret
}
