// Copyright 2020 Security Scorecard Authors
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
	"io/ioutil"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"

	"github.com/ossf/scorecard/v3/checker"
	scut "github.com/ossf/scorecard/v3/utests"
)

func TestGithubWorkflowPinning(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		expected scut.TestReturn
	}{
		{
			name:     "empty file",
			filename: "./testdata/github-workflow-empty.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  2,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "comments only",
			filename: "./testdata/github-workflow-comments.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  2,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "Pinned workflow",
			filename: "./testdata/workflow-pinned.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  2,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "Non-pinned workflow",
			filename: "./testdata/workflow-not-pinned.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore - 2,
				NumberOfWarn:  1,
				NumberOfInfo:  1,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "Non-yaml file",
			filename: "./testdata/script.sh",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  2,
				NumberOfDebug: 0,
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error

			content, err = ioutil.ReadFile(tt.filename)
			if err != nil {
				panic(fmt.Errorf("cannot read file: %w", err))
			}

			dl := scut.TestDetailLogger{}
			s, e := testIsGitHubActionsWorkflowPinned(tt.filename, content, &dl)
			actual := checker.CheckResult{
				Score:  s,
				Error2: e,
			}
			if !scut.ValidateTestReturn(t, tt.name, &tt.expected, &actual, &dl) {
				t.Fail()
			}
		})
	}
}

func TestNonGithubWorkflowPinning(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		expected scut.TestReturn
	}{
		{
			name:     "Pinned non-github workflow",
			filename: "./testdata/workflow-non-github-pinned.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  2,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "Pinned github workflow",
			filename: "./testdata/workflow-mix-github-and-non-github-not-pinned.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  2,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "Pinned github workflow",
			filename: "./testdata/workflow-mix-github-and-non-github-pinned.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  2,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "Mix of pinned and non-pinned GitHub actions",
			filename: "./testdata/workflow-mix-pinned-and-non-pinned-github.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore - 2,
				NumberOfWarn:  1,
				NumberOfInfo:  1,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "Mix of pinned and non-pinned non-GitHub actions",
			filename: "./testdata/workflow-mix-pinned-and-non-pinned-non-github.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore - 8,
				NumberOfWarn:  1,
				NumberOfInfo:  1,
				NumberOfDebug: 0,
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error
			if tt.filename == "" {
				content = make([]byte, 0)
			} else {
				content, err = ioutil.ReadFile(tt.filename)
				if err != nil {
					panic(fmt.Errorf("cannot read file: %w", err))
				}
			}
			dl := scut.TestDetailLogger{}
			s, e := testIsGitHubActionsWorkflowPinned(tt.filename, content, &dl)
			actual := checker.CheckResult{
				Score:  s,
				Error2: e,
			}
			if !scut.ValidateTestReturn(t, tt.name, &tt.expected, &actual, &dl) {
				t.Fail()
			}
		})
	}
}

func TestGithubWorkflowPkgManagerPinning(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		expected scut.TestReturn
	}{
		{
			name:     "npm packages without verification",
			filename: "./testdata/github-workflow-pkg-managers.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  22,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error

			content, err = ioutil.ReadFile(tt.filename)
			if err != nil {
				panic(fmt.Errorf("cannot read file: %w", err))
			}

			dl := scut.TestDetailLogger{}
			s, e := testValidateGitHubWorkflowScriptFreeOfInsecureDownloads(tt.filename, content, &dl)
			actual := checker.CheckResult{
				Score:  s,
				Error2: e,
			}

			if !scut.ValidateTestReturn(t, tt.name, &tt.expected, &actual, &dl) {
				t.Fail()
			}
		})
	}
}

func TestDockerfilePinning(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		filename string
		expected scut.TestReturn
	}{
		{
			name:     "invalid dockerfile",
			filename: "./testdata/Dockerfile-invalid",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  1,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "invalid dockerfile sh",
			filename: "./testdata/script-sh",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  1,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "empty file",
			filename: "./testdata/Dockerfile-empty",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  1,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "comments only",
			filename: "./testdata/Dockerfile-comments",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  1,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "Pinned dockerfile",
			filename: "./testdata/Dockerfile-pinned",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  1,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "Pinned dockerfile as",
			filename: "./testdata/Dockerfile-pinned-as",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  1,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "Non-pinned dockerfile as",
			filename: "./testdata/Dockerfile-not-pinned-as",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  3, // TODO: should be 2, https://github.com/ossf/scorecard/issues/701.
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "Non-pinned dockerfile",
			filename: "./testdata/Dockerfile-not-pinned",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error
			if tt.filename == "" {
				content = make([]byte, 0)
			} else {
				content, err = ioutil.ReadFile(tt.filename)
				if err != nil {
					panic(fmt.Errorf("cannot read file: %w", err))
				}
			}
			dl := scut.TestDetailLogger{}
			s, e := testValidateDockerfileIsPinned(tt.filename, content, &dl)
			actual := checker.CheckResult{
				Score:  s,
				Error2: e,
			}
			if !scut.ValidateTestReturn(t, tt.name, &tt.expected, &actual, &dl) {
				t.Fail()
			}
		})
	}
}

func TestDockerfilePinningWihoutHash(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		filename string
		expected scut.TestReturn
	}{
		{
			name:     "Pinned dockerfile as",
			filename: "./testdata/Dockerfile-pinned-as-without-hash",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  4,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error

			content, err = ioutil.ReadFile(tt.filename)
			if err != nil {
				panic(fmt.Errorf("cannot read file: %w", err))
			}
			dl := scut.TestDetailLogger{}
			s, e := testValidateDockerfileIsPinned(tt.filename, content, &dl)
			actual := checker.CheckResult{
				Score:  s,
				Error2: e,
			}
			if !scut.ValidateTestReturn(t, tt.name, &tt.expected, &actual, &dl) {
				t.Fail()
			}

			isExpectedLog := func(logMessage checker.LogMessage, logType checker.DetailType) bool {
				return strings.Contains(logMessage.Text, "dependency not pinned by hash")
			}

			if !scut.ValidateLogMessage(isExpectedLog, &dl) {
				t.Fail()
			}
		})
	}
}

func TestDockerfileScriptDownload(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		filename string
		expected scut.TestReturn
	}{
		{
			name:     "curl | sh",
			filename: "testdata/Dockerfile-curl-sh",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  4,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "empty file",
			filename: "./testdata/Dockerfile-empty",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  1,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "invalid file sh",
			filename: "./testdata/script.sh",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  1,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "comments only",
			filename: "./testdata/Dockerfile-comments",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  1,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "wget | /bin/sh",
			filename: "testdata/Dockerfile-wget-bin-sh",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  3,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "wget no exec",
			filename: "testdata/Dockerfile-script-ok",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  1,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "curl file sh",
			filename: "testdata/Dockerfile-curl-file-sh",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  12,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "proc substitution",
			filename: "testdata/Dockerfile-proc-subs",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  6,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "wget file",
			filename: "testdata/Dockerfile-wget-file",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  10,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "gsutil file",
			filename: "testdata/Dockerfile-gsutil-file",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  17,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "aws file",
			filename: "testdata/Dockerfile-aws-file",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  15,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "pkg managers",
			filename: "testdata/Dockerfile-pkg-managers",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  31,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "download with some python",
			filename: "testdata/Dockerfile-some-python",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error
			if tt.filename == "" {
				content = make([]byte, 0)
			} else {
				content, err = ioutil.ReadFile(tt.filename)
				if err != nil {
					panic(fmt.Errorf("cannot read file: %w", err))
				}
			}
			dl := scut.TestDetailLogger{}
			s, e := testValidateDockerfileIsFreeOfInsecureDownloads(tt.filename, content, &dl)
			actual := checker.CheckResult{
				Score:  s,
				Error2: e,
			}
			if !scut.ValidateTestReturn(t, tt.name, &tt.expected, &actual, &dl) {
				t.Fail()
			}
		})
	}
}

func TestDockerfileScriptDownloadInfo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		filename string
		expected scut.TestReturn
	}{
		{
			name:     "curl | sh",
			filename: "testdata/Dockerfile-no-curl-sh",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  1,
				NumberOfDebug: 0,
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error

			content, err = ioutil.ReadFile(tt.filename)
			if err != nil {
				panic(fmt.Errorf("cannot read file: %w", err))
			}
			dl := scut.TestDetailLogger{}
			s, e := testValidateDockerfileIsFreeOfInsecureDownloads(tt.filename, content, &dl)
			actual := checker.CheckResult{
				Score:  s,
				Error2: e,
			}
			if !scut.ValidateTestReturn(t, tt.name, &tt.expected, &actual, &dl) {
				t.Fail()
			}

			isExpectedLog := func(logMessage checker.LogMessage, logType checker.DetailType) bool {
				return strings.Contains(logMessage.Text,
					"no insecure (not pinned by hash) dependency downloads found in Dockerfiles")
			}

			if !scut.ValidateLogMessage(isExpectedLog, &dl) {
				t.Fail()
			}
		})
	}
}

func TestShellScriptDownload(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		filename string
		expected scut.TestReturn
	}{
		{
			name:     "sh script",
			filename: "testdata/script-sh",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  7,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "empty file",
			filename: "./testdata/script-empty.sh",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  1,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "comments",
			filename: "./testdata/script-comments.sh",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  1,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "bash script",
			filename: "testdata/script-bash",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  7,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "sh script 2",
			filename: "testdata/script.sh",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  7,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "pkg managers",
			filename: "testdata/script-pkg-managers",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  28,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error
			if tt.filename == "" {
				content = make([]byte, 0)
			} else {
				content, err = ioutil.ReadFile(tt.filename)
				if err != nil {
					panic(fmt.Errorf("cannot read file: %w", err))
				}
			}
			dl := scut.TestDetailLogger{}
			s, e := testValidateShellScriptIsFreeOfInsecureDownloads(tt.filename, content, &dl)
			actual := checker.CheckResult{
				Score:  s,
				Error2: e,
			}
			if !scut.ValidateTestReturn(t, tt.name, &tt.expected, &actual, &dl) {
				t.Fail()
			}
		})
	}
}

func TestShellScriptDownloadPinned(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		filename string
		expected scut.TestReturn
	}{
		{
			name:     "sh script",
			filename: "testdata/script-comments.sh",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  1,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "script free of download",
			filename: "testdata/script-free-from-download.sh",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  1,
				NumberOfDebug: 0,
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error

			content, err = ioutil.ReadFile(tt.filename)
			if err != nil {
				panic(fmt.Errorf("cannot read file: %w", err))
			}

			dl := scut.TestDetailLogger{}
			s, e := testValidateShellScriptIsFreeOfInsecureDownloads(tt.filename, content, &dl)
			actual := checker.CheckResult{
				Score:  s,
				Error2: e,
			}
			if !scut.ValidateTestReturn(t, tt.name, &tt.expected, &actual, &dl) {
				t.Fail()
			}

			isExpectedLog := func(logMessage checker.LogMessage, logType checker.DetailType) bool {
				return strings.Contains(logMessage.Text,
					"no insecure (not pinned by hash) dependency downloads found in shell scripts")
			}

			if !scut.ValidateLogMessage(isExpectedLog, &dl) {
				t.Fail()
			}
		})
	}
}

func TestGitHubWorflowRunDownload(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		filename string
		expected scut.TestReturn
	}{
		{
			name:     "workflow curl default",
			filename: "testdata/github-workflow-curl-default.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "workflow curl no default",
			filename: "testdata/github-workflow-curl-no-default.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "wget across steps",
			filename: "testdata/github-workflow-wget-across-steps.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  2,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error
			if tt.filename == "" {
				content = make([]byte, 0)
			} else {
				content, err = ioutil.ReadFile(tt.filename)
				if err != nil {
					panic(fmt.Errorf("cannot read file: %w", err))
				}
			}
			dl := scut.TestDetailLogger{}
			s, e := testValidateGitHubWorkflowScriptFreeOfInsecureDownloads(tt.filename, content, &dl)
			actual := checker.CheckResult{
				Score:  s,
				Error2: e,
			}
			if !scut.ValidateTestReturn(t, tt.name, &tt.expected, &actual, &dl) {
				t.Fail()
			}
		})
	}
}

func TestGitHubWorkflowUsesLineNumber(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		filename string
		expected []struct {
			dependency string
			lineNumber int
		}
	}{
		{
			name:     "unpinned dependency in uses",
			filename: "testdata/github-workflow-permissions-run-codeql-write.yaml",
			expected: []struct {
				dependency string
				lineNumber int
			}{
				{
					dependency: "github/codeql-action/analyze@v1",
					lineNumber: 25,
				},
			},
		},
		{
			name:     "multiple unpinned dependency in uses",
			filename: "testdata/github-workflow-multiple-unpinned-uses.yaml",
			expected: []struct {
				dependency string
				lineNumber int
			}{
				{
					dependency: "github/codeql-action/analyze@v1",
					lineNumber: 22,
				},
				{
					dependency: "docker/build-push-action@1.2.3",
					lineNumber: 26,
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			content, err := ioutil.ReadFile(tt.filename)
			if err != nil {
				t.Errorf("cannot read file: %w", err)
			}
			dl := scut.TestDetailLogger{}
			var pinned worklowPinningResult
			_, err = validateGitHubActionWorkflow(tt.filename, content, &dl, &pinned)
			if err != nil {
				t.Errorf("error during validateGitHubActionWorkflow: %w", err)
			}
			for _, expectedLog := range tt.expected {
				isExpectedLog := func(logMessage checker.LogMessage, logType checker.DetailType) bool {
					return logMessage.Offset == expectedLog.lineNumber && logMessage.Path == tt.filename &&
						logMessage.Snippet == expectedLog.dependency && logType == checker.DetailWarn &&
						strings.Contains(logMessage.Text, "dependency not pinned by hash")
				}
				if !scut.ValidateLogMessage(isExpectedLog, &dl) {
					t.Errorf("test failed: log message not present: %+v", tt.expected)
				}
			}
		})
	}
}

func TestGitHubWorkflowShell(t *testing.T) {
	t.Parallel()

	repeatItem := func(item string, count int) []string {
		ret := make([]string, 0, count)
		for i := 0; i < count; i++ {
			ret = append(ret, item)
		}
		return ret
	}

	tests := []struct {
		name     string
		filename string
		// The shells used in each step, listed in order that the steps are listed in the file
		expectedShells []string
	}{
		{
			name:           "all windows, shell specified in step",
			filename:       "testdata/github-workflow-shells-all-windows-bash.yaml",
			expectedShells: []string{"bash"},
		},
		{
			name:           "all windows, OSes listed in matrix.os",
			filename:       "testdata/github-workflow-shells-all-windows-matrix.yaml",
			expectedShells: []string{"pwsh"},
		},
		{
			name:           "all windows",
			filename:       "testdata/github-workflow-shells-all-windows.yaml",
			expectedShells: []string{"pwsh"},
		},
		{
			name:           "macOS defaults to bash",
			filename:       "testdata/github-workflow-shells-default-macos.yaml",
			expectedShells: []string{"bash"},
		},
		{
			name:           "ubuntu defaults to bash",
			filename:       "testdata/github-workflow-shells-default-ubuntu.yaml",
			expectedShells: []string{"bash"},
		},
		{
			name:           "windows defaults to pwsh",
			filename:       "testdata/github-workflow-shells-default-windows.yaml",
			expectedShells: []string{"pwsh"},
		},
		{
			name:           "windows specified in 'if'",
			filename:       "testdata/github-workflow-shells-runner-windows-ubuntu.yaml",
			expectedShells: append(repeatItem("pwsh", 7), repeatItem("bash", 4)...),
		},
		{
			name:           "shell specified in job and step",
			filename:       "testdata/github-workflow-shells-specified-job-step.yaml",
			expectedShells: []string{"bash"},
		},
		{
			name:           "windows, shell specified in job",
			filename:       "testdata/github-workflow-shells-specified-job-windows.yaml",
			expectedShells: []string{"bash"},
		},
		{
			name:           "shell specified in job",
			filename:       "testdata/github-workflow-shells-specified-job.yaml",
			expectedShells: []string{"pwsh"},
		},
		{
			name:           "shell specified in step",
			filename:       "testdata/github-workflow-shells-speficied-step.yaml",
			expectedShells: []string{"pwsh"},
		},
		{
			name:           "different shells in each step",
			filename:       "testdata/github-workflow-shells-two-shells.yaml",
			expectedShells: []string{"bash", "pwsh"},
		},
		{
			name:           "windows step, bash specified",
			filename:       "testdata/github-workflow-shells-windows-bash.yaml",
			expectedShells: []string{"bash", "bash"},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			content, err := ioutil.ReadFile(tt.filename)
			if err != nil {
				t.Errorf("cannot read file: %w", err)
			}
			var workflow gitHubActionWorkflowConfig
			err = yaml.Unmarshal(content, &workflow)
			if err != nil {
				t.Errorf("cannot unmarshal file: %w", err)
			}
			actualShells := make([]string, 0)
			for _, job := range workflow.Jobs {
				job := job
				for _, step := range job.Steps {
					step := step
					shell, err := getShellForStep(&step, &job)
					if err != nil {
						t.Errorf("error getting shell: %w", err)
					}
					actualShells = append(actualShells, shell)
				}
			}
			if !cmp.Equal(tt.expectedShells, actualShells) {
				t.Errorf("%v: Got (%v) expected (%v)", tt.name, actualShells, tt.expectedShells)
			}
		})
	}
}
