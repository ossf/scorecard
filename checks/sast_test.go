// Copyright 2020 OpenSSF Scorecard Authors
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

package checks

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
	scut "github.com/ossf/scorecard/v4/utests"
)

func Test_SAST(t *testing.T) {
	t.Parallel()

	//nolint:govet,goerr113
	tests := []struct {
		name          string
		err           error
		commits       []clients.Commit
		searchresult  clients.SearchResponse
		checkRuns     []clients.CheckRun
		searchRequest clients.SearchRequest
		path          string
		expected      checker.CheckResult
	}{
		{
			name:         "SAST checker should return failed status when no PRs are found",
			commits:      []clients.Commit{},
			searchresult: clients.SearchResponse{},
			checkRuns:    []clients.CheckRun{},
		},
		{
			name:         "SAST checker should return failed status when no PRs are found",
			err:          errors.New("error"),
			commits:      []clients.Commit{},
			searchresult: clients.SearchResponse{},
			checkRuns:    []clients.CheckRun{},
			expected:     checker.CheckResult{Score: -1},
		},
		{
			name: "Successful SAST checker should return success status for github-advanced-security",
			commits: []clients.Commit{
				{
					AssociatedMergeRequest: clients.PullRequest{
						MergedAt: time.Now().Add(time.Hour - 1),
					},
				},
			},
			searchresult: clients.SearchResponse{},
			checkRuns: []clients.CheckRun{
				{
					Status:     "completed",
					Conclusion: "success",
					App: clients.CheckRunApp{
						Slug: "github-advanced-security",
					},
				},
			},
			expected: checker.CheckResult{
				Score: 10,
			},
		},
		{
			name: "Successful SAST checker should return success status for github-code-scanning",
			commits: []clients.Commit{
				{
					AssociatedMergeRequest: clients.PullRequest{
						MergedAt: time.Now().Add(time.Hour - 1),
					},
				},
			},
			searchresult: clients.SearchResponse{},
			checkRuns: []clients.CheckRun{
				{
					Status:     "completed",
					Conclusion: "success",
					App: clients.CheckRunApp{
						Slug: "github-code-scanning",
					},
				},
			},
			expected: checker.CheckResult{
				Score: 10,
			},
		},
		{
			name: "Successful SAST checker should return success status for lgtm",
			commits: []clients.Commit{
				{
					AssociatedMergeRequest: clients.PullRequest{
						MergedAt: time.Now().Add(time.Hour - 1),
					},
				},
			},
			searchresult: clients.SearchResponse{},
			checkRuns: []clients.CheckRun{
				{
					Status:     "completed",
					Conclusion: "success",
					App: clients.CheckRunApp{
						Slug: "lgtm-com",
					},
				},
			},
			path: "",
			expected: checker.CheckResult{
				Score: 10,
			},
		},
		{
			name: "Successful SAST checker should return success status for sonarcloud",
			commits: []clients.Commit{
				{
					AssociatedMergeRequest: clients.PullRequest{
						MergedAt: time.Now().Add(time.Hour - 1),
					},
				},
			},
			searchresult: clients.SearchResponse{},
			checkRuns: []clients.CheckRun{
				{
					Status:     "completed",
					Conclusion: "success",
					App: clients.CheckRunApp{
						Slug: "sonarcloud",
					},
				},
			},
			expected: checker.CheckResult{
				Score: 10,
			},
		},
		{
			name: "Airflow Workflow has CodeQL but has no check runs.",
			err:  nil,
			commits: []clients.Commit{
				{
					AssociatedMergeRequest: clients.PullRequest{
						MergedAt: time.Now().Add(time.Hour - 1),
					},
				},
			},
			searchresult: clients.SearchResponse{},
			path:         ".github/workflows/airflow-codeql-workflow.yaml",
			expected: checker.CheckResult{
				Score: 7,
			},
		},
		{
			name: "Airflow Workflow has CodeQL and two check runs.",
			err:  nil,
			commits: []clients.Commit{
				{
					AssociatedMergeRequest: clients.PullRequest{
						MergedAt: time.Now().Add(time.Hour - 1),
					},
				},
			},
			searchresult: clients.SearchResponse{},
			checkRuns: []clients.CheckRun{
				{
					Status:     "completed",
					Conclusion: "success",
					App: clients.CheckRunApp{
						Slug: "lgtm-com",
					},
				},
				{
					Status:     "completed",
					Conclusion: "success",
					App: clients.CheckRunApp{
						Slug: "lgtm-com",
					},
				},
			},
			path: ".github/workflows/airflow-codeql-workflow.yaml",
			expected: checker.CheckResult{
				Score: 10,
			},
		},
		{
			name: `Airflow Workflow has CodeQL and two check runs one of 
			which has wrong type of conlusion. The other is 'success'`,
			err: nil,
			commits: []clients.Commit{
				{
					AssociatedMergeRequest: clients.PullRequest{
						MergedAt: time.Now().Add(time.Hour - 1),
					},
				},
			},
			searchresult: clients.SearchResponse{},
			checkRuns: []clients.CheckRun{
				{
					Status:     "completed",
					Conclusion: "wrongConclusionValue",
					App: clients.CheckRunApp{
						Slug: "lgtm-com",
					},
				},
				{
					Status:     "completed",
					Conclusion: "success",
					App: clients.CheckRunApp{
						Slug: "lgtm-com",
					},
				},
			},
			path: ".github/workflows/airflow-codeql-workflow.yaml",
			expected: checker.CheckResult{
				Score: 10,
			},
		},
		{
			name: `Airflow Workflow has CodeQL and two commits none of which 
			ran the workflow.`,
			err: nil,
			commits: []clients.Commit{
				{
					AssociatedMergeRequest: clients.PullRequest{
						MergedAt: time.Now().Add(time.Hour - 1),
					},
				},
				{
					AssociatedMergeRequest: clients.PullRequest{
						MergedAt: time.Now().Add(time.Hour - 2),
					},
				},
			},
			searchresult: clients.SearchResponse{},
			checkRuns: []clients.CheckRun{
				{
					Status:     "notCompletedForTestingOnly",
					Conclusion: "notSuccessForTestingOnly",
					App: clients.CheckRunApp{
						Slug: "lgtm-com",
					},
				},
				{
					Status:     "notCompletedForTestingOnly",
					Conclusion: "notSuccessForTestingOnly",
					App: clients.CheckRunApp{
						Slug: "lgtm-com",
					},
				},
			},
			path: ".github/workflows/airflow-codeql-workflow.yaml",
			expected: checker.CheckResult{
				Score: 7,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		searchRequest := clients.SearchRequest{
			Query: "github/codeql-action/analyze",
			Path:  "/.github/workflows",
		}
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
			mockRepoClient.EXPECT().ListCommits().DoAndReturn(func() ([]clients.Commit, error) {
				if tt.err != nil {
					return nil, tt.err
				}
				return tt.commits, tt.err
			})
			mockRepoClient.EXPECT().ListCheckRunsForRef("").Return(tt.checkRuns, nil).AnyTimes()
			mockRepoClient.EXPECT().Search(searchRequest).Return(tt.searchresult, nil).AnyTimes()
			mockRepoClient.EXPECT().ListFiles(gomock.Any()).DoAndReturn(
				func(predicate func(string) (bool, error)) ([]string, error) {
					if strings.Contains(tt.path, "pom") {
						return []string{"pom.xml"}, nil
					}
					return []string{tt.path}, nil
				}).AnyTimes()
			mockRepoClient.EXPECT().GetFileContent(gomock.Any()).DoAndReturn(func(fn string) ([]byte, error) {
				if tt.path == "" {
					return nil, nil
				}
				content, err := os.ReadFile("./testdata/" + tt.path)
				if err != nil {
					return content, fmt.Errorf("%w", err)
				}
				return content, nil
			}).AnyTimes()

			dl := scut.TestDetailLogger{}
			req := checker.CheckRequest{
				RepoClient: mockRepoClient,
				Ctx:        context.TODO(),
				Dlogger:    &dl,
			}
			res := SAST(&req)

			if res.Score != tt.expected.Score {
				t.Errorf("Expected score %d, got %d for %v", tt.expected.Score, res.Score, tt.name)
			}
			ctrl.Finish()
		})
	}
}
