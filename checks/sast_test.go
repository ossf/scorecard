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

	//nolint: govet, goerr113
	tests := []struct {
		name          string
		commits       []clients.Commit
		err           error
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
			name: "Successful SAST checker should return success status",
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
			expected: checker.CheckResult{
				Score: 10,
			},
		},
		{
			name: "Failed SAST checker should return success status",
			commits: []clients.Commit{
				{
					AssociatedMergeRequest: clients.PullRequest{
						MergedAt: time.Now().Add(time.Hour - 1),
					},
				},
				{
					AssociatedMergeRequest: clients.PullRequest{
						MergedAt: time.Now().Add(time.Hour - 10),
					},
				},
				{
					AssociatedMergeRequest: clients.PullRequest{
						MergedAt: time.Now().Add(time.Hour - 20),
					},
				},
				{
					AssociatedMergeRequest: clients.PullRequest{
						MergedAt: time.Now().Add(time.Hour - 30),
					},
				},
			},
			path: ".github/workflows/github-workflow-sast-codeql.yaml",
			checkRuns: []clients.CheckRun{
				{
					Status: "completed",
					App: clients.CheckRunApp{
						Slug: "lgtm-com",
					},
				},
			},
			expected: checker.CheckResult{
				Score: 7,
			},
		},
		{
			name: "Failed SAST checker with checkRuns not completed",
			commits: []clients.Commit{
				{
					AssociatedMergeRequest: clients.PullRequest{
						MergedAt: time.Now().Add(time.Hour - 1),
					},
				},
				{
					AssociatedMergeRequest: clients.PullRequest{
						MergedAt: time.Now().Add(time.Hour - 10),
					},
				},
				{
					AssociatedMergeRequest: clients.PullRequest{
						MergedAt: time.Now().Add(time.Hour - 20),
					},
				},
				{
					AssociatedMergeRequest: clients.PullRequest{
						MergedAt: time.Now().Add(time.Hour - 30),
					},
				},
			},
			path: ".github/workflows/github-workflow-sast-no-codeql.yaml",
			checkRuns: []clients.CheckRun{
				{
					App: clients.CheckRunApp{
						Slug: "lgtm-com",
					},
				},
			},
			expected: checker.CheckResult{
				Score: 0,
			},
		},
		{
			name: "Failed SAST with PullRequest not merged",
			commits: []clients.Commit{
				{
					AssociatedMergeRequest: clients.PullRequest{
						Number: 1,
					},
				},
			},
			searchresult: clients.SearchResponse{},
			checkRuns: []clients.CheckRun{
				{
					App: clients.CheckRunApp{
						Slug: "lgtm-com",
					},
				},
			},
			expected: checker.CheckResult{
				Score: 0,
			},
		},
		{
			name: "Merged PullRequest in a different repo",
			commits: []clients.Commit{
				{
					AssociatedMergeRequest: clients.PullRequest{
						MergedAt: time.Now(),
						Number:   1,
					},
				},
			},
			searchresult: clients.SearchResponse{},
			checkRuns: []clients.CheckRun{
				{
					App: clients.CheckRunApp{
						Slug: "lgtm-com",
					},
				},
			},
			expected: checker.CheckResult{
				Score: 0,
			},
		},
		{
			name: "sonartype config 1 line",
			path: "pom-1line.xml",
			expected: checker.CheckResult{
				Score: 10,
			},
		},
		{
			name: "sonartype config 2 lines",
			path: "pom-2lines.xml",
			expected: checker.CheckResult{
				Score: 10,
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

func Test_validateSonarConfig(t *testing.T) {
	t.Parallel()

	//nolint: govet
	tests := []struct {
		name      string
		path      string
		offset    uint
		endOffset uint
		url       string
		score     int
	}{
		{
			name:      "sonartype config 1 line",
			path:      "./testdata/pom-1line.xml",
			offset:    2,
			endOffset: 2,
			url:       "https://sonarqube.private.domain",
		},
		{
			name:      "sonartype config 2 lines",
			path:      "./testdata/pom-2lines.xml",
			offset:    2,
			endOffset: 4,
			url:       "https://sonarqube.private.domain",
		},
		{
			name: "wrong filename",
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var config []sonarConfig
			var content []byte
			var err error
			var path string
			if tt.path != "" {
				content, err = os.ReadFile(tt.path)
				if err != nil {
					t.Errorf("ReadFile: %v", err)
				}
				path = "pom.xml"
			}
			_, err = validateSonarConfig(path, content, &config)
			if err != nil {
				t.Errorf("Caught error: %v", err)
			}

			if path == "" {
				if len(config) != 0 {
					t.Errorf("Expected no result, got %d for %v", len(config), tt.name)
				}
				return
			}
			if len(config) != 1 {
				t.Errorf("Expected 1 result, got %d for %v", len(config), tt.name)
			}

			if config[0].file.Offset != tt.offset {
				t.Errorf("Expected offset %d, got %d for %v", tt.offset,
					config[0].file.Offset, tt.name)
			}

			if config[0].file.EndOffset != tt.endOffset {
				t.Errorf("Expected offset %d, got %d for %v", tt.endOffset,
					config[0].file.EndOffset, tt.name)
			}

			if config[0].url != tt.url {
				t.Errorf("Expected offset %v, got %v for %v", tt.url,
					config[0].url, tt.name)
			}
		})
	}
}
