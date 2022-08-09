package githubrepo

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
			repourl: &repoURL{
				owner: "testowner",
				repo:  "testrepo",
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
