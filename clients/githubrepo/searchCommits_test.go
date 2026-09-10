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

package githubrepo

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-github/v82/github"

	"github.com/ossf/scorecard/v5/clients"
)

func TestSearchCommitsBuildQuery(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		searchReq       clients.SearchCommitsOptions
		expectedErrType error
		name            string
		repourl         *Repo
		expectedQuery   string
		hasError        bool
	}{
		{
			name: "Basic",
			repourl: &Repo{
				owner: "testowner",
				repo:  "testrepo",
			},
			searchReq: clients.SearchCommitsOptions{
				Author: "testAuthor",
			},
			expectedQuery: "repo:testowner/testrepo author:testAuthor",
		},
		{
			name: "EmptyQuery",
			repourl: &Repo{
				owner: "testowner",
				repo:  "testrepo",
			},
			searchReq:       clients.SearchCommitsOptions{},
			hasError:        true,
			expectedErrType: errEmptyQuery,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			handler := searchCommitsHandler{
				repourl: testcase.repourl,
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

type unprocessableEntityRoundTripper struct{}

func (unprocessableEntityRoundTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	return &http.Response{
		Status:     "422 Unprocessable Entity",
		StatusCode: http.StatusUnprocessableEntity,
		Body:       io.NopCloser(strings.NewReader(`{"message": "Validation Failed"}`)),
		Header:     make(http.Header),
	}, nil
}

func TestSearchCommitsHandles422(t *testing.T) {
	t.Parallel()
	handler := searchCommitsHandler{
		ghClient: github.NewClient(&http.Client{Transport: unprocessableEntityRoundTripper{}}),
		ctx:      t.Context(),
		repourl: &Repo{
			commitSHA: clients.HeadSHA,
			owner:     "testowner",
			repo:      "testrepo",
		},
	}

	commits, err := handler.search(clients.SearchCommitsOptions{Author: "testAuthor"})
	if !errors.Is(err, clients.ErrCommitSearchUnprocessable) {
		t.Fatalf("expected ErrCommitSearchUnprocessable, got: %v", err)
	}
	if len(commits) != 0 {
		t.Fatalf("expected 0 commits, got: %d", len(commits))
	}
}
