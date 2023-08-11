// Copyright 2022 OpenSSF Scorecard Authors
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

package gitlabrepo

import (
	"errors"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/xanzy/go-gitlab"

	"github.com/ossf/scorecard/v4/clients"
)

func TestBuildQuery(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		searchReq       clients.SearchRequest
		expectedErrType error
		name            string
		repourl         *repoURL
		expectedQuery   string
		hasError        bool
	}{
		{
			name: "Basic",
			repourl: &repoURL{
				owner:     "testowner",
				projectID: "1234",
			},
			searchReq: clients.SearchRequest{
				Query: "testquery",
			},
			expectedQuery: "testquery project:testowner/1234",
		},
		{
			name: "EmptyQuery",
			repourl: &repoURL{
				owner:     "testowner",
				projectID: "1234",
			},
			searchReq:       clients.SearchRequest{},
			hasError:        true,
			expectedErrType: errEmptyQuery,
		},
		{
			name: "WithFilename",
			repourl: &repoURL{
				owner:     "testowner",
				projectID: "1234",
			},
			searchReq: clients.SearchRequest{
				Query:    "testquery",
				Filename: "filename1.txt",
			},
			expectedQuery: "testquery project:testowner/1234 in:file filename:filename1.txt",
		},
		{
			name: "WithPath",
			repourl: &repoURL{
				owner:     "testowner",
				projectID: "1234",
			},
			searchReq: clients.SearchRequest{
				Query: "testquery",
				Path:  "dir1/file1.txt",
			},
			expectedQuery: "testquery project:testowner/1234 path:dir1/file1.txt",
		},
		{
			name: "WithFilenameAndPath",
			repourl: &repoURL{
				owner:     "testowner",
				projectID: "1234",
			},
			searchReq: clients.SearchRequest{
				Query:    "testquery",
				Filename: "filename1.txt",
				Path:     "dir1/dir2",
			},
			expectedQuery: "testquery project:testowner/1234 in:file filename:filename1.txt path:dir1/dir2",
		},
		{
			name: "WithFilenameAndPathWithSeperator",
			repourl: &repoURL{
				owner:     "testowner",
				projectID: "1234",
			},
			searchReq: clients.SearchRequest{
				Query:    "testquery/query",
				Filename: "filename1.txt",
				Path:     "dir1/dir2",
			},
			expectedQuery: "testquery query project:testowner/1234 in:file filename:filename1.txt path:dir1/dir2",
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			handler := searchHandler{
				repourl: testcase.repourl,
			}

			query, err := handler.buildQuery(testcase.searchReq)
			if !testcase.hasError && err != nil {
				t.Fatalf("expected - no error, got: %v", err)
			}
			if testcase.hasError && !errors.Is(err, testcase.expectedErrType) {
				t.Fatalf("expectedErrType - %v, got -%v", testcase.expectedErrType, err)
			} else if query != testcase.expectedQuery {
				t.Fatalf("expectedQuery - %s, got - %s", testcase.expectedQuery, query)
			}
		})
	}
}

func Test_search(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		responsePath string
		want         clients.SearchResponse
		wantErr      bool
	}{
		{
			name:         "valid search",
			responsePath: "./testdata/valid-search-result",
			want: clients.SearchResponse{
				Results: []clients.SearchResult{
					{
						Path: "README.md",
					},
				},
				Hits: 1,
			},

			wantErr: false,
		},
		{
			name:         "valid search with zero results",
			responsePath: "./testdata/valid-search-result-1",
			want: clients.SearchResponse{
				Hits: 0,
			},

			wantErr: false,
		},
		{
			name:         "failure fetching the search",
			responsePath: "./testdata/invalid-search-result",
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			httpClient := &http.Client{
				Transport: stubTripper{
					responsePath: tt.responsePath,
				},
			}
			client, err := gitlab.NewClient("", gitlab.WithHTTPClient(httpClient))
			if err != nil {
				t.Fatalf("gitlab.NewClient error: %v", err)
			}
			handler := &searchHandler{
				glClient: client,
			}

			repoURL := repoURL{
				owner:     "ossf-tests",
				commitSHA: clients.HeadSHA,
			}
			handler.init(&repoURL)
			got, err := handler.search(clients.SearchRequest{
				Query:    "testquery",
				Filename: "filename1.txt",
				Path:     "dir1/dir2",
			})
			if (err != nil) != tt.wantErr {
				t.Fatalf("search error: %v, wantedErr: %t", err, tt.wantErr)
			}

			if !cmp.Equal(got, tt.want) {
				t.Errorf("search() = %v, want %v", got, cmp.Diff(got, tt.want))
			}
		})
	}
}
