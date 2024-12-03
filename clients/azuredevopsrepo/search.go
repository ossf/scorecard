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
	"strings"
	"sync"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/search"

	"github.com/ossf/scorecard/v5/clients"
)

var errEmptyQuery = errors.New("search query is empty")

type searchHandler struct {
	searchClient search.Client
	once         *sync.Once
	ctx          context.Context
	repourl      *Repo
	searchCode   fnSearchCode
}

func (s *searchHandler) init(ctx context.Context, repourl *Repo) {
	s.ctx = ctx
	s.once = new(sync.Once)
	s.repourl = repourl
	s.searchCode = s.searchClient.FetchCodeSearchResults
}

type (
	fnSearchCode func(ctx context.Context, args search.FetchCodeSearchResultsArgs) (*search.CodeSearchResponse, error)
)

func (s *searchHandler) search(request clients.SearchRequest) (clients.SearchResponse, error) {
	filters, query, err := s.buildFilters(request)
	if err != nil {
		return clients.SearchResponse{}, fmt.Errorf("handler.buildQuery: %w", err)
	}

	searchResultsPageSize := 1000
	args := search.FetchCodeSearchResultsArgs{
		Request: &search.CodeSearchRequest{
			Filters:    &filters,
			SearchText: &query,
			Top:        &searchResultsPageSize,
		},
	}
	searchResults, err := s.searchCode(s.ctx, args)
	if err != nil {
		return clients.SearchResponse{}, fmt.Errorf("FetchCodeSearchResults: %w", err)
	}

	return searchResponseFrom(searchResults), nil
}

func (s *searchHandler) buildFilters(request clients.SearchRequest) (map[string][]string, string, error) {
	filters := make(map[string][]string)
	query := strings.Builder{}
	if request.Query == "" {
		return filters, query.String(), fmt.Errorf("%w", errEmptyQuery)
	}
	query.WriteString(request.Query)
	query.WriteString(" ")

	filters["Project"] = []string{s.repourl.project}
	filters["Repository"] = []string{s.repourl.name}

	if request.Path != "" {
		filters["Path"] = []string{request.Path}
	}
	if request.Filename != "" {
		query.WriteString(fmt.Sprintf("file:%s", request.Filename))
	}

	return filters, query.String(), nil
}

func searchResponseFrom(searchResults *search.CodeSearchResponse) clients.SearchResponse {
	var results []clients.SearchResult
	for _, result := range *searchResults.Results {
		results = append(results, clients.SearchResult{
			Path: *result.Path,
		})
	}
	return clients.SearchResponse{
		Results: results,
		Hits:    *searchResults.Count,
	}
}
