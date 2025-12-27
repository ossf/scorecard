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
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	sce "github.com/ossf/scorecard/v5/errors"
	scut "github.com/ossf/scorecard/v5/utests"
)

func Test_SAST(t *testing.T) {
	t.Parallel()

	tests := []struct {
		err           error
		searchRequest clients.SearchRequest
		name          string
		path          string
		commits       []clients.Commit
		checkRuns     []clients.CheckRun
		searchresult  clients.SearchResponse
		expected      scut.TestReturn
	}{
		{
			name:         "SAST checker should return min score when no PRs are found",
			commits:      []clients.Commit{},
			searchresult: clients.SearchResponse{},
			checkRuns:    []clients.CheckRun{},
			expected: scut.TestReturn{
				Score:         checker.MinResultScore,
				NumberOfWarn:  1,
				NumberOfDebug: 8,
			},
		},
		{
			name:         "SAST checker should return failed status when an error occurs",
			err:          errors.New("error"),
			commits:      []clients.Commit{},
			searchresult: clients.SearchResponse{},
			checkRuns:    []clients.CheckRun{},
			expected: scut.TestReturn{
				Score:         checker.InconclusiveResultScore,
				Error:         sce.ErrScorecardInternal,
				NumberOfDebug: 1,
			},
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
			expected: scut.TestReturn{
				Score:         checker.MaxResultScore,
				NumberOfInfo:  1,
				NumberOfDebug: 9,
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
			expected: scut.TestReturn{
				Score:         checker.MaxResultScore,
				NumberOfInfo:  1,
				NumberOfDebug: 9,
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
			expected: scut.TestReturn{
				Score:         checker.MaxResultScore,
				NumberOfInfo:  1,
				NumberOfDebug: 9,
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
			expected: scut.TestReturn{
				Score:         checker.MaxResultScore,
				NumberOfInfo:  1,
				NumberOfDebug: 9,
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
			expected: scut.TestReturn{
				Score:         7,
				NumberOfWarn:  1,
				NumberOfInfo:  1,
				NumberOfDebug: 8,
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
			expected: scut.TestReturn{
				Score:         checker.MaxResultScore,
				NumberOfInfo:  2,
				NumberOfDebug: 9,
			},
		},
		{
			name: `Airflow Workflow has CodeQL and two check runs one of 
			which has wrong type of conclusion. The other is 'success'`,
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
			expected: scut.TestReturn{
				Score:         checker.MaxResultScore,
				NumberOfInfo:  2,
				NumberOfDebug: 9,
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
			expected: scut.TestReturn{
				Score:         7,
				NumberOfWarn:  1,
				NumberOfInfo:  1,
				NumberOfDebug: 8,
			},
		},
		{
			name: `Hadolint Workflow has CodeQL and two commits none of which 
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
			path:         ".github/workflows/github-hadolint-workflow.yaml",
			expected: scut.TestReturn{
				Score:         checker.MaxResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  1,
				NumberOfDebug: 8,
			},
		},
	}
	for _, tt := range tests {
		searchRequest := clients.SearchRequest{
			Query: "github/hadolint/hadolint-action",
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
			mockRepoClient.EXPECT().ListStatuses("").Return([]clients.Status{}, nil).AnyTimes()
			mockRepoClient.EXPECT().Search(searchRequest).Return(tt.searchresult, nil).AnyTimes()
			mockRepoClient.EXPECT().ListFiles(gomock.Any()).DoAndReturn(
				func(predicate func(string) (bool, error)) ([]string, error) {
					if strings.Contains(tt.path, "pom") {
						return []string{"pom.xml"}, nil
					}
					return []string{tt.path}, nil
				}).AnyTimes()
			mockRepoClient.EXPECT().GetFileReader(gomock.Any()).DoAndReturn(func(fn string) (io.ReadCloser, error) {
				if tt.path == "" {
					return io.NopCloser(strings.NewReader("")), nil
				}
				return os.Open("./testdata/" + tt.path)
			}).AnyTimes()

			dl := scut.TestDetailLogger{}
			req := checker.CheckRequest{
				RepoClient: mockRepoClient,
				Ctx:        t.Context(),
				Dlogger:    &dl,
			}
			res := SAST(&req)

			scut.ValidateTestReturn(t, tt.name, &tt.expected, &res, &dl)
		})
	}
}
