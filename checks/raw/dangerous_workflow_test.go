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
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
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
		{
			name:     "commits author name",
			variable: "github.event.commits[2].author.name",
			expected: true,
		},
		{
			name:     "commits author email",
			variable: "github.event.commits[2].author.email",
			expected: true,
		},
		{
			name:     "PR label name",
			variable: "github.event.pull_request.labels.foo.name",
			expected: true,
		},
		{
			name:     "PR label wildcard name",
			variable: "github.event.pull_request.labels.*.name",
			expected: true,
		},
	}
	for _, tt := range tests {
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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
			mockRepoClient.EXPECT().ListFiles(gomock.Any()).Return([]string{tt.filename}, nil)
			mockRepoClient.EXPECT().GetFileReader(gomock.Any()).DoAndReturn(func(file string) (io.ReadCloser, error) {
				return os.Open("../testdata/" + file)
			})

			req := &checker.CheckRequest{
				Ctx:        context.Background(),
				RepoClient: mockRepoClient,
			}

			dw, err := DangerousWorkflow(req)

			if !errCmp(err, tt.expected.err) {
				t.Error(cmp.Diff(err, tt.expected.err, cmpopts.EquateErrors()))
			}
			if tt.expected.err != nil {
				return
			}

			nb := len(dw.Workflows)
			if nb != tt.expected.nb {
				t.Error(cmp.Diff(nb, tt.expected.nb))
			}
		})
	}
}
