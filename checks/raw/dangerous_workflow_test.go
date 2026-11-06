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
			name:     "blocked_user name",
			variable: "github.event.pull_request.organization.blocked_user.name",
			expected: true,
		},
		{
			name:     "blocked_user email",
			variable: "github.event.pull_request.organization.blocked_user.email",
			expected: true,
		},
		{
			name:     "discussion body",
			variable: "github.event.discussion.body",
			expected: true,
		},
		{
			name:     "discussion title",
			variable: "github.event.discussion.title",
			expected: true,
		},
		{
			name:     "issue_comment body",
			variable: "github.event.issue_comment.comment.body",
			expected: true,
		},
		{
			name:     "commit_comment body",
			variable: "github.event.commit_comment.comment.body",
			expected: true,
		},
		{
			name:     "fork forkee name",
			variable: "github.event.fork.forkee.name",
			expected: true,
		},
		{
			name:     "fork forkee full_name",
			variable: "github.event.fork.forkee.full_name",
			expected: true,
		},
		{
			name:     "fork forkee description",
			variable: "github.event.fork.forkee.description",
			expected: true,
		},
		{
			name:     "fork forkee homepage",
			variable: "github.event.fork.forkee.homepage",
			expected: true,
		},
		{
			name:     "fork forkee default_branch",
			variable: "github.event.fork.forkee.default_branch",
			expected: true,
		},
		{
			name:     "trusted fork forkee id",
			variable: "github.event.fork.forkee.id",
			expected: false,
		},
		{
			name:     "workflow_run head_branch",
			variable: "github.event.workflow_run.head_branch",
			expected: true,
		},
		{
			name:     "workflow_run display_title",
			variable: "github.event.workflow_run.display_title",
			expected: true,
		},
		{
			name:     "workflow_run head_repository description",
			variable: "github.event.workflow_run.head_repository.description",
			expected: true,
		},
		{
			name:     "workflow_run pull_requests head ref",
			variable: "github.event.workflow_run.pull_requests[0].head.ref",
			expected: true,
		},
		{
			name:     "trusted workflow_run id",
			variable: "github.event.workflow_run.id",
			expected: false,
		},
		{
			name:     "toJSON github.event",
			variable: "toJSON(github.event)",
			expected: true,
		},
		{
			name:     "toJSON github context",
			variable: "toJSON(github)",
			expected: true,
		},
		{
			name:     "toJSON github.event subfield ignored",
			variable: "toJSON(github.event.pull_request)",
			expected: false,
		},
		{
			name:     "toJSON case insensitive",
			variable: "TOJSON(github.event)",
			expected: true,
		},
		{
			name:     "toJSON with spaces",
			variable: "toJSON( github.event )",
			expected: true,
		},
		{
			name:     "toJSON safe property",
			variable: "toJSON(github.repository)",
			expected: false,
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
		{
			name:     "run toJSON script injection",
			filename: ".github/workflows/github-workflow-dangerous-pattern-untrusted-script-injection-tojson.yml",
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
				Ctx:        t.Context(),
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
