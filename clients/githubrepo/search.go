// Copyright 2021 OpenSSF Scorecard Authors
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
	"errors"
	"fmt"
	"strings"

	"github.com/google/go-github/v38/github"

	"github.com/ossf/scorecard/v4/clients"
)

var errEmptyQuery = errors.New("search query is empty")

type searchHandler struct {
	ghClient *github.Client
	ctx      context.Context
	repourl  *repoURL
}

func (handler *searchHandler) init(ctx context.Context, repourl *repoURL) {
	handler.ctx = ctx
	handler.repourl = repourl
}

func (handler *searchHandler) search(request clients.SearchRequest) (clients.SearchResponse, error) {
	if !strings.EqualFold(handler.repourl.commitSHA, clients.HeadSHA) {
		return clients.SearchResponse{}, fmt.Errorf(
			"%w: Search only supported for HEAD queries", clients.ErrUnsupportedFeature)
	}
	query, err := handler.buildQuery(request)
	if err != nil {
		return clients.SearchResponse{}, fmt.Errorf("handler.buildQuery: %w", err)
	}

	resp, _, err := handler.ghClient.Search.Code(handler.ctx, query, &github.SearchOptions{})
	if err != nil {
		return clients.SearchResponse{}, fmt.Errorf("Search.Code: %w", err)
	}
	return searchResponseFrom(resp), nil
}

func (handler *searchHandler) buildQuery(request clients.SearchRequest) (string, error) {
	if request.Query == "" {
		return "", fmt.Errorf("%w", errEmptyQuery)
	}
	var queryBuilder strings.Builder
	if _, err := queryBuilder.WriteString(
		// The fuzzing check searches for GitHub URI, e.g. `github.com/org/repo`. The forward slash is one special character
		// that should be replaced with a space.
		// See https://docs.github.com/en/search-github/searching-on-github/searching-code#considerations-for-code-search
		// for reference.
		fmt.Sprintf("%s repo:%s/%s",
			strings.ReplaceAll(request.Query, "/", " "),
			handler.repourl.owner, handler.repourl.repo)); err != nil {
		return "", fmt.Errorf("WriteString: %w", err)
	}
	if request.Filename != "" {
		if _, err := queryBuilder.WriteString(
			fmt.Sprintf(" in:file filename:%s", request.Filename)); err != nil {
			return "", fmt.Errorf("WriteString: %w", err)
		}
	}
	if request.Path != "" {
		if _, err := queryBuilder.WriteString(fmt.Sprintf(" path:%s", request.Path)); err != nil {
			return "", fmt.Errorf("WriteString: %w", err)
		}
	}
	return queryBuilder.String(), nil
}

func searchResponseFrom(resp *github.CodeSearchResult) clients.SearchResponse {
	var ret clients.SearchResponse
	ret.Hits = resp.GetTotal()
	for _, result := range resp.CodeResults {
		ret.Results = append(ret.Results, clients.SearchResult{
			Path: result.GetPath(),
		})
	}
	return ret
}
