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
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go"

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
			repourl: &Repo{
				owner:     "testowner",
				projectID: "1234",
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

func Test_searchCommits_search(t *testing.T) {
	t.Parallel()

	commitsJSON := `[
  {
    "id": "commit1",
    "committer_email": "unregistered@example.com"
  },
  {
    "id": "commit2",
    "committer_email": "registered@example.com"
  },
  {
    "id": "commit3",
    "committer_email": "unregistered@example.com"
  }
]`

	registeredUsersJSON := `[
  {
    "id": 99999,
    "username": "registered_user"
  }
]`
	emptyUsersJSON := `[]`

	userSearchCalls := 0
	httpClient := &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			body := ""
			switch {
			case strings.Contains(r.URL.RawQuery, "scope=commits") || strings.Contains(r.URL.Path, "commits"):
				body = commitsJSON
			case r.URL.Query().Get("search") == "registered@example.com":
				userSearchCalls++
				body = registeredUsersJSON
			default:
				userSearchCalls++
				body = emptyUsersJSON
			}
			return &http.Response{
				Status:     "200 OK",
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
			}, nil
		}),
	}

	glClient, err := gitlab.NewClient("", gitlab.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("gitlab.NewClient error: %v", err)
	}

	handler := searchCommitsHandler{
		glClient: glClient,
		repourl: &Repo{
			owner:     "testowner",
			projectID: "1234",
			commitSHA: clients.HeadSHA,
		},
	}

	commits, err := handler.search(clients.SearchCommitsOptions{Author: "testAuthor"})
	if err != nil {
		t.Fatalf("handler.search unexpected error: %v", err)
	}

	want := []clients.Commit{
		{
			Committer: clients.User{ID: 0},
		},
		{
			Committer: clients.User{ID: 99999},
		},
		{
			Committer: clients.User{ID: 0},
		},
	}

	if !cmp.Equal(commits, want) {
		t.Errorf("handler.search diff (-got +want):\n%s", cmp.Diff(commits, want))
	}

	if userSearchCalls != 2 {
		t.Errorf("expected 2 user search API calls due to email caching, got %d", userSearchCalls)
	}
}

func Test_searchCommits_search_NonHeadSHA(t *testing.T) {
	t.Parallel()

	handler := searchCommitsHandler{
		repourl: &Repo{
			commitSHA: "abcdef123456",
		},
	}

	_, err := handler.search(clients.SearchCommitsOptions{Author: "testAuthor"})
	if !errors.Is(err, clients.ErrUnsupportedFeature) {
		t.Errorf("expected ErrUnsupportedFeature, got: %v", err)
	}
}
