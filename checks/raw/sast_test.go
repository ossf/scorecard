// Copyright 2021 OpenSSF Scorecard Authors
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

package raw

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	"github.com/ossf/scorecard/v5/finding"
)

var mergedOneHourAgo = time.Now().Add(time.Hour - 1)

func TestSAST(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		files     []string
		commits   []clients.Commit
		checkRuns []clients.CheckRun
		expected  checker.SASTData
	}{
		{
			name: "has codeql 1",
			files: []string{
				".github/workflows/workflow-not-pinned.yaml",
				".github/workflows/pom.xml",
			},
			commits: []clients.Commit{
				{
					AssociatedMergeRequest: clients.PullRequest{
						Number: 1,
					},
				},
			},
			expected: checker.SASTData{
				Workflows: []checker.SASTWorkflow{
					{
						Type: checker.CodeQLWorkflow,
						File: checker.File{
							Path:   ".github/workflows/workflow-not-pinned.yaml",
							Offset: checker.OffsetDefault,
							Type:   finding.FileTypeSource,
						},
					},
					{
						Type: checker.SonarWorkflow,
						File: checker.File{
							Path:      ".github/workflows/pom.xml",
							Type:      finding.FileTypeSource,
							Snippet:   "https://sonarqube.private.domain",
							Offset:    2,
							EndOffset: 2,
						},
					},
				},
			},
		},
		{
			name:  "has codeql 2",
			files: []string{".github/workflows/github-workflow-multiple-unpinned-uses.yaml"},
			commits: []clients.Commit{
				{
					AssociatedMergeRequest: clients.PullRequest{
						Number: 1,
					},
				},
			},
			expected: checker.SASTData{
				Workflows: []checker.SASTWorkflow{
					{
						Type: checker.CodeQLWorkflow,
						File: checker.File{
							Path:   ".github/workflows/github-workflow-multiple-unpinned-uses.yaml",
							Offset: checker.OffsetDefault,
							Type:   finding.FileTypeSource,
						},
					},
				},
			},
		},
		{
			name:  "Does not use CodeQL",
			files: []string{".github/workflows/github-workflow-download-lines.yaml"},
			expected: checker.SASTData{
				Workflows: nil,
			},
		},
		{
			name:  "Airflows CodeQL workflow - Has CodeQL",
			files: []string{".github/workflows/airflows-codeql.yaml"},
			commits: []clients.Commit{
				{
					AssociatedMergeRequest: clients.PullRequest{
						Number: 1,
					},
				},
			},
			expected: checker.SASTData{
				Workflows: []checker.SASTWorkflow{
					{
						Type: checker.CodeQLWorkflow,
						File: checker.File{
							Path:   ".github/workflows/airflows-codeql.yaml",
							Offset: checker.OffsetDefault,
							Type:   finding.FileTypeSource,
						},
					},
				},
			},
		},
		{
			name:  "Airflows CodeQL workflow - Has CodeQL with MergedAt",
			files: []string{".github/workflows/airflows-codeql.yaml"},
			commits: []clients.Commit{
				{
					AssociatedMergeRequest: clients.PullRequest{
						Number:   1,
						MergedAt: mergedOneHourAgo,
					},
				},
			},
			expected: checker.SASTData{
				Workflows: []checker.SASTWorkflow{
					{
						Type: checker.CodeQLWorkflow,
						File: checker.File{
							Path:   ".github/workflows/airflows-codeql.yaml",
							Offset: checker.OffsetDefault,
							Type:   finding.FileTypeSource,
						},
					},
				},
				Commits: []checker.SASTCommit{
					{
						AssociatedMergeRequest: clients.PullRequest{
							Number:   1,
							MergedAt: mergedOneHourAgo,
						},
						Compliant: true,
					},
				},
			},
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
		},
		{
			name:  "Has Snyk",
			files: []string{".github/workflows/github-workflow-snyk.yaml"},
			commits: []clients.Commit{
				{
					AssociatedMergeRequest: clients.PullRequest{
						Number: 1,
					},
				},
			},
			expected: checker.SASTData{
				Workflows: []checker.SASTWorkflow{
					{
						Type: checker.SnykWorkflow,
						File: checker.File{
							Path:   ".github/workflows/github-workflow-snyk.yaml",
							Offset: checker.OffsetDefault,
							Type:   finding.FileTypeSource,
						},
					},
				},
			},
		},
		{
			name:  "Has Pysa",
			files: []string{".github/workflows/github-pysa-workflow.yaml"},
			commits: []clients.Commit{
				{
					AssociatedMergeRequest: clients.PullRequest{
						Number: 1,
					},
				},
			},
			expected: checker.SASTData{
				Workflows: []checker.SASTWorkflow{
					{
						Type: checker.PysaWorkflow,
						File: checker.File{
							Path:   ".github/workflows/github-pysa-workflow.yaml",
							Offset: checker.OffsetDefault,
							Type:   finding.FileTypeSource,
						},
					},
				},
			},
		},
		{
			name:  "Has Qodana",
			files: []string{".github/workflows/github-qodana-workflow.yaml"},
			commits: []clients.Commit{
				{
					AssociatedMergeRequest: clients.PullRequest{
						Number: 1,
					},
				},
			},
			expected: checker.SASTData{
				Workflows: []checker.SASTWorkflow{
					{
						Type: checker.QodanaWorkflow,
						File: checker.File{
							Path:   ".github/workflows/github-qodana-workflow.yaml",
							Offset: checker.OffsetDefault,
							Type:   finding.FileTypeSource,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
			mockRepoClient.EXPECT().ListFiles(gomock.Any()).Return(tt.files, nil).AnyTimes()
			mockRepoClient.EXPECT().ListCommits().DoAndReturn(func() ([]clients.Commit, error) {
				return tt.commits, nil
			})
			mockRepoClient.EXPECT().ListCheckRunsForRef("").Return(tt.checkRuns, nil).AnyTimes()
			mockRepoClient.EXPECT().GetFileReader(gomock.Any()).DoAndReturn(func(file string) (io.ReadCloser, error) {
				return os.Open("./testdata/" + file)
			}).AnyTimes()
			req := checker.CheckRequest{
				RepoClient: mockRepoClient,
				Dlogger:    nil,
			}
			sastWorkflowsGot, err := SAST(&req)
			if err != nil {
				t.Error(err)
			}
			if diff := cmp.Diff(tt.expected, sastWorkflowsGot); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
