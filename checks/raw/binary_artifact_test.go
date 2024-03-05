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

	"github.com/golang/mock/gomock"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
	scut "github.com/ossf/scorecard/v4/utests"
)

func strptr(s string) *string {
	return &s
}

// TestBinaryArtifact tests the BinaryArtifact checker.
func TestBinaryArtifacts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                   string
		err                    error
		files                  [][]string
		successfulWorkflowRuns []clients.WorkflowRun
		commits                []clients.Commit
		getFileContentCount    int
		expect                 int
	}{
		{
			name: "Wasm file",
			err:  nil,
			files: [][]string{
				{"../testdata/binaryartifacts/wasms/simple.wasm"},
			},
			getFileContentCount: 1,
			expect:              1,
		},
		{
			name: "Jar file",
			err:  nil,
			files: [][]string{
				{"../testdata/binaryartifacts/jars/aws-java-sdk-core-1.11.571.jar"},
			},
			getFileContentCount: 1,
			expect:              1,
		},
		{
			name: "Mach-O ARM64 executable",
			err:  nil,
			files: [][]string{
				{"../testdata/binaryartifacts/executables/darwin-arm64-bt"},
			},
			getFileContentCount: 1,
			expect:              1,
		},
		{
			name: "non binary file",
			err:  nil,
			files: [][]string{
				{"../testdata/licensedir/withlicense/LICENSE"},
			},
			getFileContentCount: 1,
		},
		{
			name: "non binary file",
			err:  nil,
			files: [][]string{
				{"../nonexistent"},
			},
			getFileContentCount: 1,
		},
		{
			name: "printable character .lib",
			err:  nil,
			files: [][]string{
				{"../testdata/binaryartifacts/printable.lib"},
			},
			getFileContentCount: 1,
		},
		{
			name: "gradle-wrapper.jar without verification action",
			err:  nil,
			files: [][]string{
				{"../testdata/binaryartifacts/jars/gradle-wrapper.jar"},
				{},
			},
			getFileContentCount: 1,
			expect:              1,
		},
		{
			name: "gradle-wrapper.jar with verification action",
			err:  nil,
			files: [][]string{
				{"../testdata/binaryartifacts/jars/gradle-wrapper.jar"},
				{
					"../testdata/binaryartifacts/workflows/nonverify.yaml",
					"../testdata/binaryartifacts/workflows/verify.yaml",
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
			expect:              1,
		},
		{
			name: "gradle-wrapper.jar with non-verification action",
			err:  nil,
			files: [][]string{
				{"../testdata/binaryartifacts/jars/gradle-wrapper.jar"},
				{"../testdata/binaryartifacts/workflows/nonverify.yaml"},
			},
			getFileContentCount: 2,
			expect:              1,
		},
		{
			name: "gradle-wrapper.jar with verification-failing commit",
			err:  nil,
			files: [][]string{
				{"../testdata/binaryartifacts/jars/gradle-wrapper.jar"},
				{"../testdata/binaryartifacts/workflows/verify.yaml"},
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
			expect:              1,
		},
		{
			name: "gradle-wrapper.jar with outdated verification action",
			err:  nil,
			files: [][]string{
				{"../testdata/binaryartifacts/jars/gradle-wrapper.jar"},
				{
					"../testdata/binaryartifacts/workflows/nonverify.yaml",
					"../testdata/binaryartifacts/workflows/verify-outdated-action.yaml",
				},
			},
			getFileContentCount: 3,
			expect:              1,
		},
		{
			name: "gradle-wrapper.jar with no verification action due to bad workflow",
			err:  nil,
			files: [][]string{
				{"../testdata/binaryartifacts/jars/gradle-wrapper.jar"},
				{
					"../testdata/binaryartifacts/workflows/invalid.yaml",
				},
			},
			getFileContentCount: 2,
			expect:              1,
		},
		{
			name: "gradle-wrapper.jar with verification and bad workflow",
			err:  nil,
			files: [][]string{
				{"../testdata/binaryartifacts/jars/gradle-wrapper.jar"},
				{
					"../testdata/binaryartifacts/workflows/invalid.yaml",
					"../testdata/binaryartifacts/workflows/verify.yaml",
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
			expect:              1,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
			mockRepo := mockrepo.NewMockRepo(ctrl)
			for _, files := range tt.files {
				mockRepoClient.EXPECT().ListFiles(gomock.Any()).Return(files, nil)
			}
			for i := 0; i < tt.getFileContentCount; i++ {
				mockRepoClient.EXPECT().GetFileReader(gomock.Any()).DoAndReturn(func(file string) (io.ReadCloser, error) {
					return os.Open(file)
				})
			}
			if tt.successfulWorkflowRuns != nil {
				mockRepoClient.EXPECT().ListSuccessfulWorkflowRuns(gomock.Any()).Return(tt.successfulWorkflowRuns, nil)
			}
			if tt.commits != nil {
				mockRepoClient.EXPECT().ListCommits().Return(tt.commits, nil)
			}

			dl := scut.TestDetailLogger{}
			c := &checker.CheckRequest{
				RepoClient: mockRepoClient,
				Repo:       mockRepo,
				Dlogger:    &dl,
			}

			f, err := BinaryArtifacts(c)

			if tt.err != nil {
				// If we expect an error, make sure it is the same
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				// Check that the expected number of files are returned
				if len(f.Files) != tt.expect {
					t.Errorf("expected %d files, got %d test %v", tt.expect, len(f.Files), tt.name)
				}
			}
		})
	}
}

func TestBinaryArtifacts_workflow_runs_unsupported(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
	mockRepo := mockrepo.NewMockRepo(ctrl)
	const jarFile = "gradle-wrapper.jar"
	const verifyWorkflow = ".github/workflows/verify.yaml"
	files := []string{jarFile, verifyWorkflow}
	mockRepoClient.EXPECT().ListFiles(gomock.Any()).Return(files, nil).AnyTimes()
	mockRepoClient.EXPECT().GetFileReader(jarFile).DoAndReturn(func(file string) (io.ReadCloser, error) {
		return os.Open("../testdata/binaryartifacts/jars/gradle-wrapper.jar")
	}).AnyTimes()
	mockRepoClient.EXPECT().GetFileReader(verifyWorkflow).DoAndReturn(func(file string) (io.ReadCloser, error) {
		return os.Open("../testdata/binaryartifacts/workflows/verify.yaml")
	}).AnyTimes()

	mockRepoClient.EXPECT().ListSuccessfulWorkflowRuns(gomock.Any()).Return(nil, clients.ErrUnsupportedFeature).AnyTimes()
	dl := scut.TestDetailLogger{}
	c := &checker.CheckRequest{
		RepoClient: mockRepoClient,
		Repo:       mockRepo,
		Dlogger:    &dl,
	}
	got, err := BinaryArtifacts(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Files) != 1 {
		t.Errorf("expected 1 file, got %d", len(got.Files))
	}
}
