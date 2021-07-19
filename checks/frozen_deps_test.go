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
	"testing"

	"github.com/ossf/scorecard/checker"
	sce "github.com/ossf/scorecard/errors"
	scut "github.com/ossf/scorecard/utests"
)

func TestGithubWorkflowPinning(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		expected scut.TestReturn
	}{
		{
			name:     "Zero size content",
			filename: "",
			expected: scut.TestReturn{
				Errors:        []error{sce.ErrRunFailure},
				Score:         checker.InconclusiveResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "Pinned workflow",
			filename: "./testdata/workflow-pinned.yaml",
			expected: scut.TestReturn{
				Errors:        nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "Non-pinned workflow",
			filename: "./testdata/workflow-not-pinned.yaml",
			expected: scut.TestReturn{
				Errors:        nil,
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
			r := TestIsGitHubActionsWorkflowPinned(tt.filename, content, &dl)
			scut.ValidateTestReturn2(t, tt.name, &tt.expected, &r, &dl)
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
			name:     "Invalid dockerfile",
			filename: "./testdata/Dockerfile-invalid",
			expected: scut.TestReturn{
				Errors:        []error{sce.ErrRunFailure},
				Score:         checker.InconclusiveResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "Pinned dockerfile",
			filename: "./testdata/Dockerfile-pinned",
			expected: scut.TestReturn{
				Errors:        nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "Pinned dockerfile as",
			filename: "./testdata/Dockerfile-pinned-as",
			expected: scut.TestReturn{
				Errors:        nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "Non-pinned dockerfile as",
			filename: "./testdata/Dockerfile-not-pinned-as",
			expected: scut.TestReturn{
				Errors:        nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  3, // TODO:fix should be 2
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "Non-pinned dockerfile",
			filename: "./testdata/Dockerfile-not-pinned",
			expected: scut.TestReturn{
				Errors:        nil,
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
			r := TestValidateDockerfileIsPinned(tt.filename, content, &dl)
			scut.ValidateTestReturn2(t, tt.name, &tt.expected, &r, &dl)
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
				Errors:        nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  4,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "wget | /bin/sh",
			filename: "testdata/Dockerfile-wget-bin-sh",
			expected: scut.TestReturn{
				Errors:        nil,
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
				Errors:        nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "curl file sh",
			filename: "testdata/Dockerfile-curl-file-sh",
			expected: scut.TestReturn{
				Errors:        nil,
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
				Errors:        nil,
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
				Errors:        nil,
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
				Errors:        nil,
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
				Errors:        nil,
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
				Errors:        nil,
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
			if tt.filename == "" {
				content = make([]byte, 0)
			} else {
				content, err = ioutil.ReadFile(tt.filename)
				if err != nil {
					panic(fmt.Errorf("cannot read file: %w", err))
				}
			}
			dl := scut.TestDetailLogger{}
			r := TestValidateDockerfileIsFreeOfInsecureDownloads(tt.filename, content, &dl)
			scut.ValidateTestReturn2(t, tt.name, &tt.expected, &r, &dl)
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
				Errors:        nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  7,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "bash script",
			filename: "testdata/script-bash",
			expected: scut.TestReturn{
				Errors:        nil,
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
				Errors:        nil,
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
				Errors:        nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  24,
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
			r := TestValidateShellScriptIsFreeOfInsecureDownloads(tt.filename, content, &dl)
			scut.ValidateTestReturn2(t, tt.name, &tt.expected, &r, &dl)
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
			filename: "testdata/github-workflow-curl-default",
			expected: scut.TestReturn{
				Errors:        nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "workflow curl no default",
			filename: "testdata/github-workflow-curl-no-default",
			expected: scut.TestReturn{
				Errors:        nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "wget across steps",
			filename: "testdata/github-workflow-wget-across-steps",
			expected: scut.TestReturn{
				Errors:        nil,
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
			r := TestValidateGitHubWorkflowScriptFreeOfInsecureDownloads(tt.filename, content, &dl)
			scut.ValidateTestReturn2(t, tt.name, &tt.expected, &r, &dl)
		})
	}
}
