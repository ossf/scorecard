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

	"github.com/ossf/scorecard/v4/checker"
	scut "github.com/ossf/scorecard/v4/utests"
)

func TestGithubDangerousWorkflow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		expected scut.TestReturn
	}{
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
		{
			name:     "run untrusted code checkout test",
			filename: "./testdata/github-workflow-dangerous-pattern-untrusted-checkout.yml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "run trusted code checkout test",
			filename: "./testdata/github-workflow-dangerous-pattern-trusted-checkout.yml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "run default code checkout test",
			filename: "./testdata/github-workflow-dangerous-pattern-default-checkout.yml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "run safe trigger with code checkout test",
			filename: "./testdata/github-workflow-dangerous-pattern-safe-trigger.yml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "run script injection",
			filename: "./testdata/github-workflow-dangerous-pattern-untrusted-script-injection.yml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "run safe script injection",
			filename: "./testdata/github-workflow-dangerous-pattern-trusted-script-injection.yml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "run multiple script injection",
			filename: "./testdata/github-workflow-dangerous-pattern-untrusted-multiple-script-injection.yml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultConfidence,
				NumberOfWarn:  2,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "run inline script injection",
			filename: "./testdata/github-workflow-dangerous-pattern-untrusted-inline-script-injection.yml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultConfidence,
				NumberOfWarn:  1,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name:     "run wildcard script injection",
			filename: "./testdata/github-workflow-dangerous-pattern-untrusted-script-injection-wildcard.yml",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultConfidence,
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
			r := testValidateGitHubActionDangerousWorkflow(tt.filename, content, &dl)
			if !scut.ValidateTestReturn(t, tt.name, &tt.expected, &r, &dl) {
				t.Fail()
			}
		})
	}
}

func TestUntrustedContextVariables(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		variable string
		expected bool
	}{
		{
			name:     "trusted",
			variable: "github.action",
			expected: false,
		},
		{
			name:     "untrusted",
			variable: "github.head_ref",
			expected: true,
		},
		{
			name:     "untrusted event",
			variable: "github.event.issue.title",
			expected: true,
		},
		{
			name:     "untrusted pull request",
			variable: "github.event.pull_request.body",
			expected: true,
		},
		{
			name:     "trusted pull request",
			variable: "github.event.pull_request.number",
			expected: false,
		},
		{
			name:     "untrusted wildcard",
			variable: "github.event.commits[0].message",
			expected: true,
		},
		{
			name:     "trusted wildcard",
			variable: "github.event.commits[0].id",
			expected: false,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if r := containsUntrustedContextPattern(tt.variable); !r == tt.expected {
				t.Fail()
			}
		})
	}
}
