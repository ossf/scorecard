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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/search"

	"github.com/ossf/scorecard/v5/clients"
)

func Test_buildFilters(t *testing.T) {
	t.Parallel()
	tests := []struct {
		expectedFilters map[string][]string
		repourl         *Repo
		request         clients.SearchRequest
		expectedErr     error
		expectedQuery   string
		name            string
	}{
		{
			name: "empty query",
			request: clients.SearchRequest{
				Query: "",
			},
			repourl: &Repo{
				project: "project",
				name:    "repo",
			},
			expectedFilters: make(map[string][]string),
			expectedQuery:   "",
			expectedErr:     errEmptyQuery,
		},
		{
			name: "query only",
			request: clients.SearchRequest{
				Query: "query",
			},
			repourl: &Repo{
				project: "project",
				name:    "repo",
			},
			expectedFilters: map[string][]string{
				"Project":    {"project"},
				"Repository": {"repo"},
			},
			expectedQuery: "query ",
			expectedErr:   nil,
		},
		{
			name: "query and path",
			request: clients.SearchRequest{
				Query: "query",
				Path:  "path",
			},
			repourl: &Repo{
				project: "project",
				name:    "repo",
			},
			expectedFilters: map[string][]string{
				"Project":    {"project"},
				"Repository": {"repo"},
				"Path":       {"path"},
			},
			expectedQuery: "query ",
			expectedErr:   nil,
		},
		{
			name: "query and filename",
			request: clients.SearchRequest{
				Query:    "query",
				Filename: "filename",
			},
			repourl: &Repo{
				project: "project",
				name:    "repo",
			},
			expectedFilters: map[string][]string{
				"Project":    {"project"},
				"Repository": {"repo"},
			},
			expectedQuery: "query file:filename",
			expectedErr:   nil,
		},
		{
			name: "query, path, and filename",
			request: clients.SearchRequest{
				Query:    "query",
				Path:     "path",
				Filename: "filename",
			},
			repourl: &Repo{
				project: "project",
				name:    "repo",
			},
			expectedFilters: map[string][]string{
				"Project":    {"project"},
				"Repository": {"repo"},
				"Path":       {"path"},
			},
			expectedQuery: "query file:filename",
			expectedErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := searchHandler{
				repourl: tt.repourl,
			}
			filters, query, err := s.buildFilters(tt.request)
			if tt.expectedErr != nil && !errors.Is(err, tt.expectedErr) {
				t.Fatalf("buildFilters() error = %v, ExpectedErr %v", err, tt.expectedErr)
			}
			if query != tt.expectedQuery {
				t.Errorf("buildFilters() query = %v, ExpectedQuery %v", query, tt.expectedQuery)
			}
			if diff := cmp.Diff(filters, tt.expectedFilters); diff != "" {
				t.Errorf("buildFilters() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_search(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		request     clients.SearchRequest
		searchCode  func(ctx context.Context, args search.FetchCodeSearchResultsArgs) (*search.CodeSearchResponse, error)
		wantResults []clients.SearchResult
		wantHits    int
		wantErr     bool
	}{
		{
			name: "empty query",
			request: clients.SearchRequest{
				Query: "",
			},
			searchCode: func(ctx context.Context, args search.FetchCodeSearchResultsArgs) (*search.CodeSearchResponse, error) {
				return &search.CodeSearchResponse{}, nil
			},
			wantErr: true,
		},
		{
			name: "valid query",
			request: clients.SearchRequest{
				Query: "query",
			},
			searchCode: func(ctx context.Context, args search.FetchCodeSearchResultsArgs) (*search.CodeSearchResponse, error) {
				return &search.CodeSearchResponse{
					Count:   toPtr(1),
					Results: &[]search.CodeResult{{Path: strptr("path")}},
				}, nil
			},
			wantResults: []clients.SearchResult{{Path: "path"}},
			wantHits:    1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := searchHandler{
				searchCode: tt.searchCode,
				repourl: &Repo{
					project: "project",
					name:    "repo",
				},
			}

			got, err := s.search(tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("search() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got.Results, tt.wantResults); diff != "" {
				t.Errorf("search() mismatch (-want +got):\n%s", diff)
			}
			if got.Hits != tt.wantHits {
				t.Errorf("search() gotHits = %v, want %v", got.Hits, tt.wantHits)
			}
		})
	}
}
