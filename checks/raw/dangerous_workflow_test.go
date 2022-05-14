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

package raw

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
)

func errCmp(e1, e2 error) bool {
	return errors.Is(e1, e2) || errors.Is(e2, e1)
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

func TestGithubDangerousWorkflow(t *testing.T) {
	t.Parallel()

	type ret struct {
		err error
		nb  int
	}
	tests := []struct {
		name     string
		filename string
		expected ret
	}{
		{
			name:     "Non-yaml file",
			filename: "script.sh",
			expected: ret{nb: 0},
		},
		{
			name:     "run untrusted code checkout test - workflow_run",
			filename: ".github/workflows/github-workflow-dangerous-pattern-untrusted-checkout-workflow_run.yml",
			expected: ret{nb: 1},
		},
		{
			name:     "run untrusted code checkout test",
			filename: ".github/workflows/github-workflow-dangerous-pattern-untrusted-checkout.yml",
			expected: ret{nb: 1},
		},
		{
			name:     "run trusted code checkout test",
			filename: ".github/workflows/github-workflow-dangerous-pattern-trusted-checkout.yml",
			expected: ret{nb: 0},
		},
		{
			name:     "run default code checkout test",
			filename: ".github/workflows/github-workflow-dangerous-pattern-default-checkout.yml",
			expected: ret{nb: 0},
		},
		{
			name:     "run script injection",
			filename: ".github/workflows/github-workflow-dangerous-pattern-untrusted-script-injection.yml",
			expected: ret{nb: 1},
		},
		{
			name:     "run safe script injection",
			filename: ".github/workflows/github-workflow-dangerous-pattern-trusted-script-injection.yml",
			expected: ret{nb: 0},
		},
		{
			name:     "run multiple script injection",
			filename: ".github/workflows/github-workflow-dangerous-pattern-untrusted-multiple-script-injection.yml",
			expected: ret{nb: 2},
		},
		{
			name:     "run inline script injection",
			filename: ".github/workflows/github-workflow-dangerous-pattern-untrusted-inline-script-injection.yml",
			expected: ret{nb: 1},
		},
		{
			name:     "run wildcard script injection",
			filename: ".github/workflows/github-workflow-dangerous-pattern-untrusted-script-injection-wildcard.yml",
			expected: ret{nb: 1},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
			mockRepoClient.EXPECT().ListFiles(gomock.Any()).Return([]string{tt.filename}, nil)
			mockRepoClient.EXPECT().GetFileContent(gomock.Any()).DoAndReturn(func(file string) ([]byte, error) {
				// This will read the file and return the content
				content, err := os.ReadFile("../testdata/" + file)
				if err != nil {
					return content, fmt.Errorf("%w", err)
				}
				return content, nil
			})

			dw, err := DangerousWorkflow(mockRepoClient)

			if !errCmp(err, tt.expected.err) {
				t.Errorf(cmp.Diff(err, tt.expected.err, cmpopts.EquateErrors()))
			}
			if tt.expected.err != nil {
				return
			}

			nb := len(dw.Workflows)
			if nb != tt.expected.nb {
				t.Errorf(cmp.Diff(nb, tt.expected.nb))
			}
		})
	}
}
