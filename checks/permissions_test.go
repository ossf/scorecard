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

	"github.com/ossf/scorecard/v2/checker"
	scut "github.com/ossf/scorecard/v2/utests"
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
			name:     "Write all test",
			filename: "./testdata/github-workflow-permissions-writeall.yaml",
			expected: scut.TestReturn{
				Errors:        nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "Read all test",
			filename: "./testdata/github-workflow-permissions-readall.yaml",
			expected: scut.TestReturn{
				Errors:        nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  1,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "No permission test",
			filename: "./testdata/github-workflow-permissions-absent.yaml",
			expected: scut.TestReturn{
				Errors:        nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "Writes test",
			filename: "./testdata/github-workflow-permissions-writes.yaml",
			expected: scut.TestReturn{
				Errors:        nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  1,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "Reads test",
			filename: "./testdata/github-workflow-permissions-reads.yaml",
			expected: scut.TestReturn{
				Errors:        nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  10,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "Nones test",
			filename: "./testdata/github-workflow-permissions-nones.yaml",
			expected: scut.TestReturn{
				Errors:        nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  10,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "None test",
			filename: "./testdata/github-workflow-permissions-none.yaml",
			expected: scut.TestReturn{
				Errors:        nil,
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
			scut.ValidateTestReturn(t, tt.name, &tt.expected, &r, &dl)
		})
	}
}
