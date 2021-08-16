// Copyright 2021 Security Scorecard Authors
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
	"errors"
	"testing"

	"github.com/ossf/scorecard/v2/clients"
)

func TestBuildQuery(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		searchReq       clients.SearchRequest
		expectedErrType error
		name            string
		owner           string
		repo            string
		expectedQuery   string
		hasError        bool
	}{
		{
			name:  "Basic",
			owner: "testowner",
			repo:  "testrepo",
			searchReq: clients.SearchRequest{
				Query: "testquery",
			},
			expectedQuery: "testquery repo:testowner/testrepo",
		},
		{
			name:            "EmptyQuery",
			owner:           "testowner",
			repo:            "testrepo",
			searchReq:       clients.SearchRequest{},
			hasError:        true,
			expectedErrType: errEmptyQuery,
		},
		{
			name:  "WithFilename",
			owner: "testowner",
			repo:  "testrepo",
			searchReq: clients.SearchRequest{
				Query:    "testquery",
				Filename: "filename1.txt",
			},
			expectedQuery: "testquery repo:testowner/testrepo in:file filename:filename1.txt",
		},
		{
			name:  "WithPath",
			owner: "testowner",
			repo:  "testrepo",
			searchReq: clients.SearchRequest{
				Query: "testquery",
				Path:  "dir1/file1.txt",
			},
			expectedQuery: "testquery repo:testowner/testrepo path:dir1/file1.txt",
		},
		{
			name:  "WithFilenameAndPath",
			owner: "testowner",
			repo:  "testrepo",
			searchReq: clients.SearchRequest{
				Query:    "testquery",
				Filename: "filename1.txt",
				Path:     "dir1/dir2",
			},
			expectedQuery: "testquery repo:testowner/testrepo in:file filename:filename1.txt path:dir1/dir2",
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			handler := searchHandler{
				owner: testcase.owner,
				repo:  testcase.repo,
			}

			query, err := handler.buildQuery(testcase.searchReq)
			if !testcase.hasError && err != nil {
				t.Fatalf("expected - no error, got: %v", err)
			}
			if testcase.hasError && !errors.Is(err, testcase.expectedErrType) {
				t.Fatalf("expectedErrType - %v, got - %v",
					testcase.expectedErrType, err)
			} else if query != testcase.expectedQuery {
				t.Fatalf("expectedQuery - %s, got - %s",
					testcase.expectedQuery, query)
			}
		})
	}
}
