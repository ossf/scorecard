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

// TODO(laurent): share this function across unit tests

func TestGithubWorkflowPinning(t *testing.T) {
	t.Parallel()

	tests := []scut.TestInfo{
		{
			Name: "Zero size content",
			Args: scut.TestArgs{
				Filename: "",
			},
			Expected: scut.TestReturn{
				Errors:        []error{sce.ErrRunFailure},
				Score:         checker.InconclusiveResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			Name: "Pinned workflow",
			Args: scut.TestArgs{
				Filename: "./testdata/workflow-pinned.yaml",
			},
			Expected: scut.TestReturn{
				Errors:        nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			Name: "Non-pinned workflow",
			Args: scut.TestArgs{
				Filename: "./testdata/workflow-not-pinned.yaml",
			},
			Expected: scut.TestReturn{
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
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error
			if tt.Args.Filename == "" {
				content = make([]byte, 0)
			} else {
				content, err = ioutil.ReadFile(tt.Args.Filename)
				if err != nil {
					panic(fmt.Errorf("cannot read file: %w", err))
				}
			}
			dl := scut.TestDetailLogger{}
			r := TestIsGitHubActionsWorkflowPinned(tt.Args.Filename, content, &dl)
			tt.Args.Dl = dl
			scut.ValidateTest(t, &tt, r)
		})
	}
}

func TestDockerfilePinning(t *testing.T) {
	t.Parallel()
	tests := []scut.TestInfo{
		{
			Name: "Invalid dockerfile",
			Args: scut.TestArgs{
				Filename: "./testdata/Dockerfile-invalid",
			},
			Expected: scut.TestReturn{
				Errors:        []error{sce.ErrRunFailure},
				Score:         checker.InconclusiveResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			Name: "Pinned dockerfile",
			Args: scut.TestArgs{
				Filename: "./testdata/Dockerfile-pinned",
			},
			Expected: scut.TestReturn{
				Errors:        nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			Name: "Pinned dockerfile as",
			Args: scut.TestArgs{
				Filename: "./testdata/Dockerfile-pinned-as",
			},
			Expected: scut.TestReturn{
				Errors:        nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			Name: "Non-pinned dockerfile as",
			Args: scut.TestArgs{
				Filename: "./testdata/Dockerfile-not-pinned-as",
			},
			Expected: scut.TestReturn{
				Errors:        nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  3, // TODO:fix should be 2
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			Name: "Non-pinned dockerfile",
			Args: scut.TestArgs{
				Filename: "./testdata/Dockerfile-not-pinned",
			},
			Expected: scut.TestReturn{
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
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error
			if tt.Args.Filename == "" {
				content = make([]byte, 0)
			} else {
				content, err = ioutil.ReadFile(tt.Args.Filename)
				if err != nil {
					panic(fmt.Errorf("cannot read file: %w", err))
				}
			}
			dl := scut.TestDetailLogger{}
			r := TestValidateDockerfileIsPinned(tt.Args.Filename, content, &dl)
			tt.Args.Dl = dl
			scut.ValidateTest(t, &tt, r)
		})
	}
}

func TestDockerfileScriptDownload(t *testing.T) {
	t.Parallel()
	tests := []scut.TestInfo{
		{
			Name: "curl | sh",
			Args: scut.TestArgs{
				Filename: "testdata/Dockerfile-curl-sh",
			},
			Expected: scut.TestReturn{
				Errors:        nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  4,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			Name: "wget | /bin/sh",
			Args: scut.TestArgs{
				Filename: "testdata/Dockerfile-wget-bin-sh",
			},
			Expected: scut.TestReturn{
				Errors:        nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  3,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			Name: "wget no exec",
			Args: scut.TestArgs{
				Filename: "testdata/Dockerfile-script-ok",
			},
			Expected: scut.TestReturn{
				Errors:        nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			Name: "curl file sh",
			Args: scut.TestArgs{
				Filename: "testdata/Dockerfile-curl-file-sh",
			},
			Expected: scut.TestReturn{
				Errors:        nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  12,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			Name: "proc substitution",
			Args: scut.TestArgs{
				Filename: "testdata/Dockerfile-proc-subs",
			},
			Expected: scut.TestReturn{
				Errors:        nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  6,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			Name: "wget file",
			Args: scut.TestArgs{
				Filename: "testdata/Dockerfile-wget-file",
			},
			Expected: scut.TestReturn{
				Errors:        nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  10,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			Name: "gsutil file",
			Args: scut.TestArgs{
				Filename: "testdata/Dockerfile-gsutil-file",
			},
			Expected: scut.TestReturn{
				Errors:        nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  17,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			Name: "aws file",
			Args: scut.TestArgs{
				Filename: "testdata/Dockerfile-aws-file",
			},
			Expected: scut.TestReturn{
				Errors:        nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  15,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			Name: "pkg managers",
			Args: scut.TestArgs{
				Filename: "testdata/Dockerfile-pkg-managers",
			},
			Expected: scut.TestReturn{
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
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error
			if tt.Args.Filename == "" {
				content = make([]byte, 0)
			} else {
				content, err = ioutil.ReadFile(tt.Args.Filename)
				if err != nil {
					panic(fmt.Errorf("cannot read file: %w", err))
				}
			}
			dl := scut.TestDetailLogger{}
			r := TestValidateDockerfileIsFreeOfInsecureDownloads(tt.Args.Filename, content, &dl)
			tt.Args.Dl = dl
			scut.ValidateTest(t, &tt, r)
		})
	}
}

func TestShellScriptDownload(t *testing.T) {
	t.Parallel()
	tests := []scut.TestInfo{
		{
			Name: "sh script",
			Args: scut.TestArgs{
				Filename: "testdata/script-sh",
			},
			Expected: scut.TestReturn{
				Errors:        nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  7,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			Name: "bash script",
			Args: scut.TestArgs{
				Filename: "testdata/script-bash",
			},
			Expected: scut.TestReturn{
				Errors:        nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  7,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			Name: "sh script 2",
			Args: scut.TestArgs{
				Filename: "testdata/script.sh",
			},
			Expected: scut.TestReturn{
				Errors:        nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  7,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			Name: "pkg managers",
			Args: scut.TestArgs{
				Filename: "testdata/script-pkg-managers",
			},
			Expected: scut.TestReturn{
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
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error
			if tt.Args.Filename == "" {
				content = make([]byte, 0)
			} else {
				content, err = ioutil.ReadFile(tt.Args.Filename)
				if err != nil {
					panic(fmt.Errorf("cannot read file: %w", err))
				}
			}
			dl := scut.TestDetailLogger{}
			r := TestValidateShellScriptIsFreeOfInsecureDownloads(tt.Args.Filename, content, &dl)
			tt.Args.Dl = dl
			scut.ValidateTest(t, &tt, r)
		})
	}
}

func TestGitHubWorflowRunDownload(t *testing.T) {
	t.Parallel()
	tests := []scut.TestInfo{
		{
			Name: "workflow curl default",
			Args: scut.TestArgs{
				Filename: "testdata/github-workflow-curl-default",
			},
			Expected: scut.TestReturn{
				Errors:        nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			Name: "workflow curl no default",
			Args: scut.TestArgs{
				Filename: "testdata/github-workflow-curl-no-default",
			},
			Expected: scut.TestReturn{
				Errors:        nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			Name: "wget across steps",
			Args: scut.TestArgs{
				Filename: "testdata/github-workflow-wget-across-steps",
			},
			Expected: scut.TestReturn{
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
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error
			if tt.Args.Filename == "" {
				content = make([]byte, 0)
			} else {
				content, err = ioutil.ReadFile(tt.Args.Filename)
				if err != nil {
					panic(fmt.Errorf("cannot read file: %w", err))
				}
			}
			dl := scut.TestDetailLogger{}
			r := TestValidateGitHubWorkflowScriptFreeOfInsecureDownloads(tt.Args.Filename, content, &dl)
			tt.Args.Dl = dl
			scut.ValidateTest(t, &tt, r)
		})
	}
}
