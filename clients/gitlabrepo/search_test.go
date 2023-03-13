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
	"testing"

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
				owner:   "testowner",
				project: "1234",
			},
			searchReq: clients.SearchRequest{
				Query: "testquery",
			},
			expectedQuery: "testquery project:testowner/1234",
		},
		{
			name: "EmptyQuery",
			repourl: &repoURL{
				owner:   "testowner",
				project: "1234",
			},
			searchReq:       clients.SearchRequest{},
			hasError:        true,
			expectedErrType: errEmptyQuery,
		},
		{
			name: "WithFilename",
			repourl: &repoURL{
				owner:   "testowner",
				project: "1234",
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
				owner:   "testowner",
				project: "1234",
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
				owner:   "testowner",
				project: "1234",
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
				owner:   "testowner",
				project: "1234",
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
