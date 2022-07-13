// Copyright 2022 Security Scorecard Authors
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

package evaluation

import (
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/raw"
	"github.com/ossf/scorecard/v4/clients"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
	scut "github.com/ossf/scorecard/v4/utests"
)

func strptr(s string) *string {
	return &s
}

// TestBinaryArtifacts tests the binary artifacts check.
func TestBinaryArtifacts(t *testing.T) {
	t.Run("GradleWrapperValidation", testGradleWrapperValidation)
	t.Parallel()
	//nolint
	type args struct {
		name string
		dl   checker.DetailLogger
		r    *checker.BinaryArtifactData
	}
	tests := []struct {
		name    string
		args    args
		want    checker.CheckResult
		wantErr bool
	}{
		{
			name: "r nil",
			args: args{
				name: "test_binary_artifacts_check_pass",
				dl:   &scut.TestDetailLogger{},
			},
			wantErr: true,
		},
		{
			name: "no binary artifacts",
			args: args{
				name: "no binary artifacts",
				dl:   &scut.TestDetailLogger{},
				r:    &checker.BinaryArtifactData{},
			},
			want: checker.CheckResult{
				Score: checker.MaxResultScore,
			},
		},
		{
			name: "1 binary artifact",
			args: args{
				name: "no binary artifacts",
				dl:   &scut.TestDetailLogger{},
				r: &checker.BinaryArtifactData{
					Files: []checker.File{
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
					},
				},
			},
			want: checker.CheckResult{
				Score: 9,
			},
		},
		{
			name: "many binary artifact",
			args: args{
				name: "no binary artifacts",
				dl:   &scut.TestDetailLogger{},
				r: &checker.BinaryArtifactData{
					Files: []checker.File{
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
					},
				},
			},
			want: checker.CheckResult{
				Score: 0,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := BinaryArtifacts(tt.args.name, tt.args.dl, tt.args.r, nil)
			if tt.wantErr {
				if got.Error == nil {
					t.Errorf("BinaryArtifacts() error = %v, wantErr %v", got.Error, tt.wantErr)
				}
			} else {
				if got.Score != tt.want.Score {
					t.Errorf("BinaryArtifacts() = %v, want %v", got.Score, tt.want.Score)
				}
			}
		})
	}
}

// testGradleWrapperValidation tests the BinaryArtifacts
// evaluator with repos containing gradle-wrapper.jar.
func testGradleWrapperValidation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                   string
		files                  [][]string
		successfulWorkflowRuns []clients.WorkflowRun
		commits                []clients.Commit
		getFileContentCount    int
		expect                 int
	}{
		{
			name: "Jar file",
			files: [][]string{
				{"../testdata/binaryartifacts/jars/aws-java-sdk-core-1.11.571.jar"},
			},
			getFileContentCount: 1,
			expect:              9,
		},
		{
			name: "gradle-wrapper.jar without verification action",
			files: [][]string{
				{"../testdata/binaryartifacts/jars/gradle-wrapper.jar"},
				{},
			},
			getFileContentCount: 1,
			expect:              9,
		},
		{
			name: "gradle-wrapper.jar with verification action",
			files: [][]string{
				{"../testdata/binaryartifacts/jars/gradle-wrapper.jar"},
				{
					"../testdata/binaryartifacts/workflows/nonverify.yml",
					"../testdata/binaryartifacts/workflows/verify.yml",
				},
			},
			successfulWorkflowRuns: []clients.WorkflowRun{
				{
					HeadSHA: strptr("sha-a"),
				},
			},
			commits: []clients.Commit{
				{
					SHA: "sha-a",
				},
				{
					SHA: "sha-old",
				},
			},
			getFileContentCount: 3,
			expect:              10,
		},
		{
			name: "gradle-wrapper.jar with non-verification action",
			files: [][]string{
				{"../testdata/binaryartifacts/jars/gradle-wrapper.jar"},
				{"../testdata/binaryartifacts/workflows/nonverify.yml"},
			},
			getFileContentCount: 2,
			expect:              9,
		},
		{
			name: "gradle-wrapper.jar with verification-failing commit",
			files: [][]string{
				{"../testdata/binaryartifacts/jars/gradle-wrapper.jar"},
				{"../testdata/binaryartifacts/workflows/verify.yml"},
			},
			successfulWorkflowRuns: []clients.WorkflowRun{
				{
					HeadSHA: strptr("sha-old"),
				},
			},
			commits: []clients.Commit{
				{
					SHA: "sha-a",
				},
				{
					SHA: "sha-old",
				},
			},
			getFileContentCount: 2,
			expect:              9,
		},
		{
			name: "gradle-wrapper.jar with outdated verification action",
			files: [][]string{
				{"../testdata/binaryartifacts/jars/gradle-wrapper.jar"},
				{
					"../testdata/binaryartifacts/workflows/nonverify.yml",
					"../testdata/binaryartifacts/workflows/verify-outdated-action.yml",
				},
			},
			getFileContentCount: 3,
			expect:              9,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
			for _, files := range tt.files {
				mockRepoClient.EXPECT().ListFiles(gomock.Any()).Return(files, nil)
			}
			for i := 0; i < tt.getFileContentCount; i++ {
				mockRepoClient.EXPECT().GetFileContent(gomock.Any()).DoAndReturn(func(file string) ([]byte, error) {
					// This will read the file and return the content
					content, err := os.ReadFile(file)
					if err != nil {
						return content, fmt.Errorf("%w", err)
					}
					return content, nil
				})
			}
			if tt.successfulWorkflowRuns != nil {
				mockRepoClient.EXPECT().ListSuccessfulWorkflowRuns(gomock.Any()).Return(tt.successfulWorkflowRuns, nil)
			}
			if tt.commits != nil {
				mockRepoClient.EXPECT().ListCommits().Return(tt.commits, nil)
			}

			dl := checker.NewLogger()
			data, err := raw.BinaryArtifacts(mockRepoClient)
			if err != nil {
				t.Errorf("error while running raw.BinaryArtifacts: %v", err)
			}

			result := BinaryArtifacts("BinaryArtifacts", dl, &data, mockRepoClient)

			t.Logf("Reason: %s", result.Reason)
			t.Logf("Details: %v", result.Details)

			if result.Error != nil {
				t.Errorf("error while evaluating: %v", result.Error)
			}
			// Check that the expected score is returned
			if result.Score != tt.expect {
				t.Errorf("expected score %d, got %d test %v", tt.expect, result.Score, tt.name)
			}
		})
	}
}
