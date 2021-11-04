// Copyright 2021 Security Scorecard Authors
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

	"github.com/ossf/scorecard/v3/checker"
	scut "github.com/ossf/scorecard/v3/utests"
)

//nolint
func TestGithubTokenPermissions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		expected scut.TestReturn
	}{
		{
			name:     "run workflow codeql write test",
			filename: "./testdata/github-workflow-permissions-run-codeql-write.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  1,
				NumberOfDebug: 3,
			},
		},
		{
			name:     "run workflow no codeql write test",
			filename: "./testdata/github-workflow-permissions-run-no-codeql-write.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore - 1,
				NumberOfWarn:  1,
				NumberOfInfo:  1,
				NumberOfDebug: 3,
			},
		},
		{
			name:     "run workflow write test",
			filename: "./testdata/github-workflow-permissions-run-writes-2.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  3,
				NumberOfInfo:  2,
				NumberOfDebug: 3,
			},
		},
		{
			name:     "run package workflow write test",
			filename: "./testdata/github-workflow-permissions-run-package-workflow-write.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  2,
				NumberOfDebug: 3,
			},
		},
		{
			name:     "run package write test",
			filename: "./testdata/github-workflow-permissions-run-package-write.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  1,
				NumberOfDebug: 3,
			},
		},
		{
			name:     "run writes test",
			filename: "./testdata/github-workflow-permissions-run-writes.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  1,
				NumberOfDebug: 4,
			},
		},
		{
			name:     "write all test",
			filename: "./testdata/github-workflow-permissions-writeall.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  0,
				NumberOfDebug: 4,
			},
		},
		{
			name:     "read all test",
			filename: "./testdata/github-workflow-permissions-readall.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  1,
				NumberOfDebug: 4,
			},
		},
		{
			name:     "no permission test",
			filename: "./testdata/github-workflow-permissions-absent.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  0,
				NumberOfDebug: 4,
			},
		},
		{
			name:     "writes test",
			filename: "./testdata/github-workflow-permissions-writes.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  1,
				NumberOfDebug: 5,
			},
		},
		{
			name:     "reads test",
			filename: "./testdata/github-workflow-permissions-reads.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  10,
				NumberOfDebug: 4,
			},
		},
		{
			name:     "nones test",
			filename: "./testdata/github-workflow-permissions-nones.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  10,
				NumberOfDebug: 4,
			},
		},
		{
			name:     "none test",
			filename: "./testdata/github-workflow-permissions-none.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  1,
				NumberOfDebug: 4,
			},
		},
		{
			name:     "status/checks write",
			filename: "./testdata/github-workflow-permissions-status-checks.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore - 1,
				NumberOfWarn:  2,
				NumberOfInfo:  2,
				NumberOfDebug: 5,
			},
		},
		{
			name:     "sec-events/deployments write",
			filename: "./testdata/github-workflow-permissions-secevent-deployments.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore - 2,
				NumberOfWarn:  2,
				NumberOfInfo:  3,
				NumberOfDebug: 4,
			},
		},
		{
			name:     "contents write",
			filename: "./testdata/github-workflow-permissions-contents.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  2,
				NumberOfDebug: 4,
			},
		},
		{
			name:     "actions write",
			filename: "./testdata/github-workflow-permissions-actions.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  2,
				NumberOfDebug: 4,
			},
		},
		{
			name:     "packages write",
			filename: "./testdata/github-workflow-permissions-packages.yaml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  1,
				NumberOfDebug: 4,
			},
		},
		{
			name:     "Non-yaml file",
			filename: "./testdata/script.sh",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
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
			r := testValidateGitHubActionTokenPermissions(tt.filename, content, &dl)
			if !scut.ValidateTestReturn(t, tt.name, &tt.expected, &r, &dl) {
				t.Fail()
			}
		})
	}
}
