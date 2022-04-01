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
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ossf/scorecard/v4/checker"
	scut "github.com/ossf/scorecard/v4/utests"
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
			filename: "./testdata/.github/workflows/github-workflow-empty.yaml",
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
			filename: "./testdata/.github/workflows/github-workflow-comments.yaml",
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
			filename: "./testdata/.github/workflows/workflow-pinned.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  2,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "Local action workflow",
			filename: "./testdata/.github/workflows/workflow-local-action.yaml",
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
			filename: "./testdata/.github/workflows/workflow-not-pinned.yaml",
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
		{
			name:     "Matrix as expression",
			filename: "./testdata/.github/workflows/github-workflow-matrix-expression.yaml",
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

			content, err = os.ReadFile(tt.filename)
			if err != nil {
				t.Errorf("cannot read file: %v", err)
			}

			dl := scut.TestDetailLogger{}
			p := strings.Replace(tt.filename, "./testdata/", "", 1)

			s, e := testIsGitHubActionsWorkflowPinned(p, content, &dl)
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
			filename: "./testdata/.github/workflows/workflow-non-github-pinned.yaml",
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
			filename: "./testdata/.github/workflows/workflow-mix-github-and-non-github-not-pinned.yaml",
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
			filename: "./testdata/.github/workflows/workflow-mix-github-and-non-github-pinned.yaml",
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
			filename: "./testdata/.github/workflows/workflow-mix-pinned-and-non-pinned-github.yaml",
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
			filename: "./testdata/.github/workflows/workflow-mix-pinned-and-non-pinned-non-github.yaml",
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
				content, err = os.ReadFile(tt.filename)
				if err != nil {
					t.Errorf("cannot read file: %v", err)
				}
			}
			dl := scut.TestDetailLogger{}
			p := strings.Replace(tt.filename, "./testdata/", "", 1)

			s, e := testIsGitHubActionsWorkflowPinned(p, content, &dl)
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
			filename: "./testdata/.github/workflows/github-workflow-pkg-managers.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  27,
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

			content, err = os.ReadFile(tt.filename)
			if err != nil {
				t.Errorf("cannot read file: %v", err)
			}

			dl := scut.TestDetailLogger{}
			p := strings.Replace(tt.filename, "./testdata/", "", 1)

			s, e := testValidateGitHubWorkflowScriptFreeOfInsecureDownloads(p, content, &dl)
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
				NumberOfWarn:  2,
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
				content, err = os.ReadFile(tt.filename)
				if err != nil {
					t.Errorf("cannot read file: %v", err)
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

func TestDockerfilePinningFromLineNumber(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		filename string
		expected []struct {
			snippet   string
			startLine uint
			endLine   uint
		}
	}{
		{
			name:     "Non-pinned dockerfile as",
			filename: "./testdata/Dockerfile-not-pinned-as",
			expected: []struct {
				snippet   string
				startLine uint
				endLine   uint
			}{
				{
					snippet:   "FROM python:3.7 as build",
					startLine: 17,
					endLine:   17,
				},
				{
					snippet:   "FROM build",
					startLine: 23,
					endLine:   23,
				},
			},
		},
		{
			name:     "Non-pinned dockerfile",
			filename: "./testdata/Dockerfile-not-pinned",
			expected: []struct {
				snippet   string
				startLine uint
				endLine   uint
			}{
				{
					snippet:   "FROM python:3.7",
					startLine: 17,
					endLine:   17,
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
			dl := scut.TestDetailLogger{}
			var pinned pinnedResult
			_, err = validateDockerfileIsPinned(tt.filename, content, &dl, &pinned)
			if err != nil {
				t.Errorf("error during validateDockerfileIsPinned: %v", err)
			}
			for _, expectedLog := range tt.expected {
				isExpectedLog := func(logMessage checker.LogMessage, logType checker.DetailType) bool {
					return logMessage.Offset == expectedLog.startLine &&
						logMessage.EndOffset == expectedLog.endLine &&
						logMessage.Path == tt.filename &&
						logMessage.Snippet == expectedLog.snippet && logType == checker.DetailWarn &&
						strings.Contains(logMessage.Text, "image not pinned by hash")
				}
				if !scut.ValidateLogMessage(isExpectedLog, &dl) {
					t.Errorf("test failed: log message not present: %+v", tt.expected)
				}
			}
		})
	}
}

func TestDockerfileInvalidFiles(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{
			name:     "dockerfile go",
			filename: "./testdata/Dockerfile.go",
			expected: false,
		},
		{
			name:     "dockerfile c",
			filename: "./testdata/Dockerfile.c",
			expected: false,
		},
		{
			name:     "dockerfile cpp",
			filename: "./testdata/Dockerfile.cpp",
			expected: false,
		},
		{
			name:     "dockerfile rust",
			filename: "./testdata/Dockerfile.rs",
			expected: false,
		},
		{
			name:     "dockerfile js",
			filename: "./testdata/Dockerfile.js",
			expected: false,
		},
		{
			name:     "dockerfile sh",
			filename: "./testdata/Dockerfile.sh",
			expected: false,
		},
		{
			name:     "dockerfile py",
			filename: "./testdata/Dockerfile.py",
			expected: false,
		},
		{
			name:     "dockerfile pyc",
			filename: "./testdata/Dockerfile.pyc",
			expected: false,
		},
		{
			name:     "dockerfile java",
			filename: "./testdata/Dockerfile.java",
			expected: false,
		},
		{
			name:     "dockerfile ",
			filename: "./testdata/Dockerfile.any",
			expected: true,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var c []byte
			r := isDockerfile(tt.filename, c)
			if r != tt.expected {
				t.Errorf("test failed: %s. Expected %v. Got %v", tt.filename, r, tt.expected)
			}
		})
	}
}

func TestDockerfileInsecureDownloadsLineNumber(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		filename string
		expected []struct {
			snippet   string
			startLine uint
			endLine   uint
		}
	}{
		{
			name:     "dockerfile downloads",
			filename: "./testdata/Dockerfile-download-lines",
			expected: []struct {
				snippet   string
				startLine uint
				endLine   uint
			}{
				{
					snippet:   "curl bla | bash",
					startLine: 35,
					endLine:   36,
				},
				{
					snippet:   "pip install -r requirements.txt",
					startLine: 41,
					endLine:   42,
				},
			},
		},
		{
			name:     "dockerfile downloads multi-run",
			filename: "./testdata/Dockerfile-download-multi-runs",
			expected: []struct {
				snippet   string
				startLine uint
				endLine   uint
			}{
				{
					snippet:   "/tmp/file3",
					startLine: 28,
					endLine:   28,
				},
				{
					snippet:   "/tmp/file1",
					startLine: 30,
					endLine:   30,
				},
				{
					snippet:   "bash /tmp/file3",
					startLine: 32,
					endLine:   34,
				},
				{
					snippet:   "bash /tmp/file1",
					startLine: 37,
					endLine:   38,
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
			dl := scut.TestDetailLogger{}
			var pinned pinnedResult
			_, err = validateDockerfileIsFreeOfInsecureDownloads(tt.filename, content, &dl, &pinned)
			if err != nil {
				t.Errorf("error during validateDockerfileIsPinned: %v", err)
			}

			for _, expectedLog := range tt.expected {
				isExpectedLog := func(logMessage checker.LogMessage, logType checker.DetailType) bool {
					fmt.Println(logMessage)
					return logMessage.Offset == expectedLog.startLine &&
						logMessage.EndOffset == expectedLog.endLine &&
						logMessage.Path == tt.filename &&
						logMessage.Snippet == expectedLog.snippet && logType == checker.DetailWarn
				}
				if !scut.ValidateLogMessage(isExpectedLog, &dl) {
					t.Errorf("test failed: log message not present: %+v", tt.expected)
				}
			}
		})
	}
}

func TestShellscriptInsecureDownloadsLineNumber(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		filename string
		expected []struct {
			snippet   string
			startLine uint
			endLine   uint
		}
	}{
		{
			name:     "shell downloads",
			filename: "./testdata/shell-download-lines.sh",
			expected: []struct {
				snippet   string
				startLine uint
				endLine   uint
			}{
				{
					snippet:   "bash /tmp/file",
					startLine: 6,
					endLine:   6,
				},
				{
					snippet:   "curl bla | bash",
					startLine: 11,
					endLine:   11,
				},
				{
					snippet:   "bash <(wget -qO- http://website.com/my-script.sh)",
					startLine: 18,
					endLine:   18,
				},
				{
					snippet:   "bash <(wget -qO- http://website.com/my-script.sh)",
					startLine: 20,
					endLine:   20,
				},
				{
					snippet:   "pip install -r requirements.txt",
					startLine: 26,
					endLine:   26,
				},
				{
					snippet:   "curl bla | bash",
					startLine: 28,
					endLine:   28,
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
			dl := scut.TestDetailLogger{}
			var pinned pinnedResult
			_, err = validateShellScriptIsFreeOfInsecureDownloads(tt.filename, content, &dl, &pinned)
			if err != nil {
				t.Errorf("error during validateDockerfileIsPinned: %v", err)
			}

			for _, expectedLog := range tt.expected {
				isExpectedLog := func(logMessage checker.LogMessage, logType checker.DetailType) bool {
					return logMessage.Offset == expectedLog.startLine &&
						logMessage.EndOffset == expectedLog.endLine &&
						logMessage.Path == tt.filename &&
						logMessage.Snippet == expectedLog.snippet && logType == checker.DetailWarn
				}
				if !scut.ValidateLogMessage(isExpectedLog, &dl) {
					t.Errorf("test failed: log message not present: %+v", tt.expected)
				}
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
			name:     "Pinned dockerfile as no hash",
			filename: "./testdata/Dockerfile-pinned-as-without-hash",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  4,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "Dockerfile with args",
			filename: "./testdata/Dockerfile-args",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  2,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "Dockerfile with base",
			filename: "./testdata/Dockerfile-base",
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

			content, err = os.ReadFile(tt.filename)
			if err != nil {
				t.Errorf("cannot read file: %v", err)
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
				if tt.expected.NumberOfWarn > 0 {
					return strings.Contains(logMessage.Text, "image not pinned by hash")
				}
				return true
			}

			if !scut.ValidateLogMessage(isExpectedLog, &dl) {
				t.Errorf("test failed: log message not present: %+v", tt.expected)
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
			filename: "./testdata/Dockerfile-curl-sh",
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
			filename: "./testdata/Dockerfile-wget-bin-sh",
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
			filename: "./testdata/Dockerfile-script-ok",
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
			filename: "./testdata/Dockerfile-curl-file-sh",
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
			filename: "./testdata/Dockerfile-proc-subs",
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
			filename: "./testdata/Dockerfile-wget-file",
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
			filename: "./testdata/Dockerfile-gsutil-file",
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
			filename: "./testdata/Dockerfile-aws-file",
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
			filename: "./testdata/Dockerfile-pkg-managers",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  37,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "download with some python",
			filename: "./testdata/Dockerfile-some-python",
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
				content, err = os.ReadFile(tt.filename)
				if err != nil {
					t.Errorf("cannot read file: %v", err)
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
			filename: "./testdata/Dockerfile-no-curl-sh",
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

			content, err = os.ReadFile(tt.filename)
			if err != nil {
				t.Errorf("cannot read file: %v", err)
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
			filename: "./testdata/script-sh",
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
			filename: "./testdata/script-bash",
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
			filename: "./testdata/script.sh",
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
			filename: "./testdata/script-pkg-managers",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  34,
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
				content, err = os.ReadFile(tt.filename)
				if err != nil {
					t.Errorf("cannot read file: %v", err)
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
			name:     "script free of download",
			filename: "./testdata/script-free-from-download.sh",
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

			content, err = os.ReadFile(tt.filename)
			if err != nil {
				t.Errorf("cannot read file: %v", err)
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
			filename: "./testdata/.github/workflows/github-workflow-curl-default.yaml",
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
			filename: "./testdata/.github/workflows/github-workflow-curl-no-default.yaml",
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
			filename: "./testdata/.github/workflows/github-workflow-wget-across-steps.yaml",
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
				content, err = os.ReadFile(tt.filename)
				if err != nil {
					t.Errorf("cannot read file: %v", err)
				}
			}
			dl := scut.TestDetailLogger{}
			p := strings.Replace(tt.filename, "./testdata/", "", 1)

			s, e := testValidateGitHubWorkflowScriptFreeOfInsecureDownloads(p, content, &dl)
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
			startLine  uint
			endLine    uint
		}
	}{
		{
			name:     "unpinned dependency in uses",
			filename: "./testdata/.github/workflows/github-workflow-permissions-run-codeql-write.yaml",
			expected: []struct {
				dependency string
				startLine  uint
				endLine    uint
			}{
				{
					dependency: "github/codeql-action/analyze@v1",
					startLine:  25,
					endLine:    25,
				},
			},
		},
		{
			name:     "multiple unpinned dependency in uses",
			filename: "./testdata/.github/workflows/github-workflow-multiple-unpinned-uses.yaml",
			expected: []struct {
				dependency string
				startLine  uint
				endLine    uint
			}{
				{
					dependency: "github/codeql-action/analyze@v1",
					startLine:  22,
					endLine:    22,
				},
				{
					dependency: "docker/build-push-action@1.2.3",
					startLine:  24,
					endLine:    24,
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
			dl := scut.TestDetailLogger{}
			var pinned worklowPinningResult
			p := strings.Replace(tt.filename, "./testdata/", "", 1)

			_, err = validateGitHubActionWorkflow(p, content, &dl, &pinned)
			if err != nil {
				t.Errorf("error during validateGitHubActionWorkflow: %v", err)
			}
			for _, expectedLog := range tt.expected {
				isExpectedLog := func(logMessage checker.LogMessage, logType checker.DetailType) bool {
					return logMessage.Offset == expectedLog.startLine &&
						logMessage.EndOffset == expectedLog.endLine &&
						logMessage.Path == p &&
						logMessage.Snippet == expectedLog.dependency && logType == checker.DetailWarn &&
						strings.Contains(logMessage.Text, "action not pinned by hash")
				}

				if !scut.ValidateLogMessage(isExpectedLog, &dl) {
					t.Errorf("test failed: log message not present: %+v", tt.expected)
				}
			}
		})
	}
}

func TestGitHubWorkInsecureDownloadsLineNumber(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		filename string
		expected []struct {
			snippet   string
			startLine uint
			endLine   uint
		}
	}{
		{
			name:     "downloads",
			filename: "./testdata/.github/workflows/github-workflow-download-lines.yaml",
			expected: []struct {
				snippet   string
				startLine uint
				endLine   uint
			}{
				{
					snippet:   "bash /tmp/file",
					startLine: 27,
					endLine:   27,
				},
				{
					snippet:   "/tmp/file2",
					startLine: 29,
					endLine:   29,
				},
				{
					snippet:   "curl bla | bash",
					startLine: 32,
					endLine:   32,
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
			dl := scut.TestDetailLogger{}
			var pinned pinnedResult
			p := strings.Replace(tt.filename, "./testdata/", "", 1)

			_, err = validateGitHubWorkflowIsFreeOfInsecureDownloads(p, content, &dl, &pinned)
			if err != nil {
				t.Errorf("error during validateGitHubWorkflowIsFreeOfInsecureDownloads: %v", err)
			}
			for _, expectedLog := range tt.expected {
				isExpectedLog := func(logMessage checker.LogMessage, logType checker.DetailType) bool {
					return logMessage.Offset == expectedLog.startLine &&
						logMessage.EndOffset == expectedLog.endLine &&
						logMessage.Path == p &&
						logMessage.Snippet == expectedLog.snippet && logType == checker.DetailWarn
				}

				if !scut.ValidateLogMessage(isExpectedLog, &dl) {
					t.Errorf("test failed: log message not present: %+v", tt.expected)
				}
			}
		})
	}
}

func Test_createReturnValuesForGitHubActionsWorkflowPinned(t *testing.T) {
	t.Parallel()
	//nolint
	type args struct {
		r       worklowPinningResult
		infoMsg string
		dl      checker.DetailLogger
		err     error
	}
	//nolint
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "both actions workflow pinned",
			args: args{
				r: worklowPinningResult{
					thirdParties: 1,
					gitHubOwned:  1,
				},
				infoMsg: "",
				dl:      &scut.TestDetailLogger{},
				err:     nil,
			},
			want:    10,
			wantErr: false,
		},
		{
			name: "github actions workflow pinned",
			args: args{
				r: worklowPinningResult{
					thirdParties: 2,
					gitHubOwned:  2,
				},
				infoMsg: "",
				dl:      &scut.TestDetailLogger{},
				err:     nil,
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "error in github actions workflow pinned",
			args: args{
				r: worklowPinningResult{
					thirdParties: 2,
					gitHubOwned:  2,
				},
				infoMsg: "",
				dl:      &scut.TestDetailLogger{},
				err:     errors.New("error"),
			},
			want:    -1,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := createReturnValuesForGitHubActionsWorkflowPinned(tt.args.r, tt.args.infoMsg, tt.args.dl, tt.args.err)
			if (err != nil) != tt.wantErr {
				t.Errorf("createReturnValuesForGitHubActionsWorkflowPinned() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("createReturnValuesForGitHubActionsWorkflowPinned() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createReturnValues(t *testing.T) {
	t.Parallel()
	//nolint
	type args struct {
		r       pinnedResult
		infoMsg string
		dl      checker.DetailLogger
		err     error
	}
	//nolint
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "returns 10 if no error",
			args: args{
				r:       1,
				infoMsg: "",
				dl:      &scut.TestDetailLogger{},
				err:     nil,
			},
			want:    10,
			wantErr: false,
		},
		{
			name: "returns 0 if unpinned ",
			args: args{
				r:       2,
				infoMsg: "",
				dl:      &scut.TestDetailLogger{},
				err:     nil,
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "if err is not nil, returns 0",
			args: args{
				r:       2,
				infoMsg: "",
				dl:      &scut.TestDetailLogger{},
				//nolint
				err: errors.New("error"),
			},
			want:    -1,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := createReturnValues(tt.args.r, tt.args.infoMsg, tt.args.dl, tt.args.err)
			if (err != nil) != tt.wantErr {
				t.Errorf("createReturnValues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("createReturnValues() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_maxScore(t *testing.T) {
	t.Parallel()
	type args struct {
		s1 int
		s2 int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "returns s1 if s1 is greater than s2",
			args: args{
				s1: 10,
				s2: 5,
			},
			want: 10,
		},
		{
			name: "returns s2 if s2 is greater than s1",
			args: args{
				s1: 5,
				s2: 10,
			},
			want: 10,
		},
		{
			name: "returns s1 if s1 is equal to s2",
			args: args{
				s1: 10,
				s2: 10,
			},
			want: 10,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := maxScore(tt.args.s1, tt.args.s2); got != tt.want {
				t.Errorf("maxScore() = %v, want %v", got, tt.want)
			}
		})
	}
}
