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

func TestSearchCommitsBuildQuery(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		searchReq       clients.SearchCommitsOptions
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
			searchReq: clients.SearchCommitsOptions{
				Author: "testAuthor",
			},
			expectedQuery: "project:testowner/1234 author:testAuthor",
		},
		{
			name: "EmptyQuery:",
			repourl: &repoURL{
				owner:     "testowner",
				projectID: "1234",
			},
			searchReq:       clients.SearchCommitsOptions{},
			hasError:        true,
			expectedErrType: errEmptyQuery,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			handler := searchCommitsHandler{
				repourl: testcase.repourl,
			}

			query, err := handler.buildQuery(testcase.searchReq)
			if !testcase.hasError && err != nil {
				t.Fatalf("expected - no error, get: %v", err)
			}
			if testcase.hasError && !errors.Is(err, testcase.expectedErrType) {
				t.Fatalf("expectedErrType - %v, got - %v", testcase.expectedErrType, err)
			} else if query != testcase.expectedQuery {
				t.Fatalf("expectedQuery - %s, got - %s", testcase.expectedQuery, query)
			}
		})
	}
}
