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

package checks

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
	scut "github.com/ossf/scorecard/v4/utests"
)

// nolint
func TestGithubTokenPermissions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		filenames []string
		expected  scut.TestReturn
	}{
		{
			name:      "run workflow codeql write test",
			filenames: []string{"./testdata/.github/workflows/github-workflow-permissions-run-codeql-write.yaml"},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  2,
				NumberOfDebug: 5,
			},
		},
		{
			name:      "run workflow no codeql write test",
			filenames: []string{"./testdata/.github/workflows/github-workflow-permissions-run-no-codeql-write.yaml"},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  1,
				NumberOfDebug: 4,
			},
		},
		{
			name:      "run workflow write test",
			filenames: []string{"./testdata/.github/workflows/github-workflow-permissions-run-writes-2.yaml"},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  3,
				NumberOfInfo:  2,
				NumberOfDebug: 4,
			},
		},
		{
			name:      "run package workflow write test",
			filenames: []string{"./testdata/.github/workflows/github-workflow-permissions-run-package-workflow-write.yaml"},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  2,
				NumberOfDebug: 5,
			},
		},
		{
			name:      "run package write test",
			filenames: []string{"./testdata/.github/workflows/github-workflow-permissions-run-package-write.yaml"},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  1,
				NumberOfDebug: 4,
			},
		},
		{
			name:      "run writes test",
			filenames: []string{"./testdata/.github/workflows/github-workflow-permissions-run-writes.yaml"},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  2,
				NumberOfDebug: 5,
			},
		},
		{
			name:      "write all test",
			filenames: []string{"./testdata/.github/workflows/github-workflow-permissions-writeall.yaml"},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  1,
				NumberOfDebug: 5,
			},
		},
		{
			name:      "read all test",
			filenames: []string{"./testdata/.github/workflows/github-workflow-permissions-readall.yaml"},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  2,
				NumberOfDebug: 5,
			},
		},
		{
			name:      "no permission test",
			filenames: []string{"./testdata/.github/workflows/github-workflow-permissions-absent.yaml"},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  1,
				NumberOfDebug: 5,
			},
		},
		{
			name:      "writes test",
			filenames: []string{"./testdata/.github/workflows/github-workflow-permissions-writes.yaml"},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  2,
				NumberOfDebug: 6,
			},
		},
		{
			name:      "reads test",
			filenames: []string{"./testdata/.github/workflows/github-workflow-permissions-reads.yaml"},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  11,
				NumberOfDebug: 5,
			},
		},
		{
			name:      "nones test",
			filenames: []string{"./testdata/.github/workflows/github-workflow-permissions-nones.yaml"},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  11,
				NumberOfDebug: 5,
			},
		},
		{
			name:      "none test",
			filenames: []string{"./testdata/.github/workflows/github-workflow-permissions-none.yaml"},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  2,
				NumberOfDebug: 5,
			},
		},
		{
			name:      "status/checks write",
			filenames: []string{"./testdata/.github/workflows/github-workflow-permissions-status-checks.yaml"},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore - 1,
				NumberOfWarn:  2,
				NumberOfInfo:  3,
				NumberOfDebug: 6,
			},
		},
		{
			name:      "sec-events/deployments write",
			filenames: []string{"./testdata/.github/workflows/github-workflow-permissions-secevent-deployments.yaml"},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore - 2,
				NumberOfWarn:  2,
				NumberOfInfo:  4,
				NumberOfDebug: 5,
			},
		},
		{
			name:      "contents write",
			filenames: []string{"./testdata/.github/workflows/github-workflow-permissions-contents.yaml"},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  3,
				NumberOfDebug: 5,
			},
		},
		{
			name:      "actions write",
			filenames: []string{"./testdata/.github/workflows/github-workflow-permissions-actions.yaml"},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  3,
				NumberOfDebug: 5,
			},
		},
		{
			name:      "packages write",
			filenames: []string{"./testdata/.github/workflows/github-workflow-permissions-packages.yaml"},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  2,
				NumberOfDebug: 5,
			},
		},
		{
			name:      "Non-yaml file",
			filenames: []string{"./testdata/script.sh"},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.InconclusiveResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:      "package workflow contents write",
			filenames: []string{"./testdata/.github/workflows/github-workflow-permissions-contents-writes-no-release.yaml"},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  1,
				NumberOfDebug: 4,
			},
		},
		{
			name:      "release workflow contents write",
			filenames: []string{"./testdata/.github/workflows/github-workflow-permissions-contents-writes-release.yaml"},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  2,
				NumberOfDebug: 4,
			},
		},
		{
			name:      "release workflow contents write",
			filenames: []string{"./testdata/.github/workflows/github-workflow-permissions-contents-writes-release-mvn-release.yaml"},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  2,
				NumberOfDebug: 4,
			},
		},
		{
			name:      "release workflow contents write semantic-release",
			filenames: []string{"./testdata/.github/workflows/github-workflow-permissions-contents-writes-release-semantic-release.yaml"},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  2,
				NumberOfDebug: 4,
			},
		},
		{
			name:      "package workflow write",
			filenames: []string{"./testdata/.github/workflows/github-workflow-permissions-packages-writes.yaml"},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  2,
				NumberOfDebug: 5,
			},
		},
		{
			name:      "workflow jobs only",
			filenames: []string{"./testdata/.github/workflows/github-workflow-permissions-jobs-only.yaml"},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore - 1,
				NumberOfWarn:  1,
				NumberOfInfo:  4,
				NumberOfDebug: 4,
			},
		},
		{
			name:      "security-events write, codeql comment",
			filenames: []string{"./testdata/.github/workflows/github-workflow-permissions-run-write-codeql-comment.yaml"},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  1,
				NumberOfDebug: 4,
			},
		},
		{
			name:      "security-events write, known actions",
			filenames: []string{"./testdata/.github/workflows/github-workflow-permissions-secevent-known-actions.yaml"},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  2, // This is constant.
				NumberOfDebug: 8, // This is 4 + (number of actions)
			},
		},
		{
			name: "two files mix run-level and top-level",
			filenames: []string{
				"./testdata/.github/workflows/github-workflow-permissions-top-level-only.yaml",
				"./testdata/.github/workflows/github-workflow-permissions-run-level-only.yaml",
			},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore - 1,
				NumberOfWarn:  1,
				NumberOfInfo:  3,
				NumberOfDebug: 9,
			},
		},
		{
			name: "two files mix run-level and absent",
			filenames: []string{
				"./testdata/.github/workflows/github-workflow-permissions-run-level-only.yaml",
				"./testdata/.github/workflows/github-workflow-permissions-absent.yaml",
			},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  2,
				NumberOfInfo:  2,
				NumberOfDebug: 9,
			},
		},
		{
			name: "two files mix top-level and absent",
			filenames: []string{
				"./testdata/.github/workflows/github-workflow-permissions-top-level-only.yaml",
				"./testdata/.github/workflows/github-workflow-permissions-absent.yaml",
			},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  2,
				NumberOfDebug: 10,
			},
		},
		{
			name: "read permission with GitHub pages write",
			filenames: []string{
				"./testdata/.github/workflows/github-workflow-permissions-gh-pages.yaml",
			},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  2,
				NumberOfDebug: 5,
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockRepo := mockrepo.NewMockRepoClient(ctrl)
			mockRepo.EXPECT().GetDefaultBranchName().Return("main", nil).AnyTimes()

			main := "main"
			mockRepo.EXPECT().URI().Return("github.com/ossf/scorecard").AnyTimes()
			mockRepo.EXPECT().GetDefaultBranch().Return(&clients.BranchRef{Name: &main}, nil).AnyTimes()
			mockRepo.EXPECT().ListFiles(gomock.Any()).DoAndReturn(func(predicate func(string) (bool, error)) ([]string, error) {
				files := []string{}
				for _, fn := range tt.filenames {
					files = append(files, strings.TrimPrefix(fn, "./testdata/"))
				}
				return files, nil
			}).AnyTimes()
			mockRepo.EXPECT().GetFileContent(gomock.Any()).DoAndReturn(func(fn string) ([]byte, error) {
				content, err := os.ReadFile("./testdata/" + fn)
				if err != nil {
					return content, fmt.Errorf("%w", err)
				}
				return content, nil
			}).AnyTimes()
			dl := scut.TestDetailLogger{}
			c := checker.CheckRequest{
				RepoClient: mockRepo,
				Dlogger:    &dl,
			}

			res := TokenPermissions(&c)

			if !scut.ValidateTestReturn(t, tt.name, &tt.expected, &res, &dl) {
				t.Errorf("test failed: log message not present: %+v\n%+v", tt.expected, dl)
			}
		})
	}
}

func TestGithubTokenPermissionsLineNumber(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		filename string
		expected []struct {
			lineNumber uint
		}
	}{
		{
			name:     "Job level write permission",
			filename: "./testdata/.github/workflows/github-workflow-permissions-run-no-codeql-write.yaml",
			expected: []struct {
				lineNumber uint
			}{
				{
					lineNumber: 22,
				},
			},
		},
		{
			name:     "Workflow level write permission",
			filename: "./testdata/.github/workflows/github-workflow-permissions-writeall.yaml",
			expected: []struct {
				lineNumber uint
			}{
				{
					lineNumber: 16,
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			content, err := os.ReadFile(tt.filename)
			if err != nil {
				t.Errorf("cannot read file: %v", err)
			}

			p := strings.Replace(tt.filename, "./testdata/", "", 1)
			ctrl := gomock.NewController(t)
			mockRepo := mockrepo.NewMockRepoClient(ctrl)

			main := "main"
			mockRepo.EXPECT().URI().Return("github.com/ossf/scorecard").AnyTimes()
			mockRepo.EXPECT().GetDefaultBranchName().Return(main, nil).AnyTimes()
			mockRepo.EXPECT().ListFiles(gomock.Any()).DoAndReturn(func(predicate func(string) (bool, error)) ([]string, error) {
				return []string{p}, nil
			}).AnyTimes()
			mockRepo.EXPECT().GetFileContent(gomock.Any()).DoAndReturn(func(fn string) ([]byte, error) {
				return content, nil
			}).AnyTimes()
			dl := scut.TestDetailLogger{}
			c := checker.CheckRequest{
				RepoClient: mockRepo,
				Dlogger:    &dl,
			}

			_ = TokenPermissions(&c)

			for _, expectedLog := range tt.expected {
				isExpectedLog := func(logMessage checker.LogMessage, logType checker.DetailType) bool {
					return logMessage.Finding != nil &&
						logMessage.Finding.Location != nil &&
						logMessage.Finding.Location.LineStart != nil &&
						*logMessage.Finding.Location.LineStart == expectedLog.lineNumber &&
						logMessage.Finding.Location.Path == p &&
						logType == checker.DetailWarn
				}
				if !scut.ValidateLogMessage(isExpectedLog, &dl) {
					t.Errorf("test failed: log message not present: %+v", tt.expected)
				}
			}
		})
	}
}
