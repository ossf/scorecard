// Copyright 2024 OpenSSF Scorecard Authors
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
	"io"
	"os"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	scut "github.com/ossf/scorecard/v5/utests"
)

func TestDangerousWorkflow(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		workflowPaths []string
		err           error
		expected      scut.TestReturn
	}{
		{
			name:          "no workflows is an inconclusive score",
			workflowPaths: nil,
			err:           nil,
			expected: scut.TestReturn{
				Score: checker.InconclusiveResultScore,
			},
		},
		{
			name:          "untrusted checkout is a failing score",
			workflowPaths: []string{".github/workflows/github-workflow-dangerous-pattern-untrusted-checkout.yml"},
			err:           nil,
			expected: scut.TestReturn{
				Score:        checker.MinResultScore,
				NumberOfWarn: 1,
			},
		},
		{
			name:          "script injection is a failing score",
			workflowPaths: []string{".github/workflows/github-workflow-dangerous-pattern-untrusted-script-injection.yml"},
			err:           nil,
			expected: scut.TestReturn{
				Score:        checker.MinResultScore,
				NumberOfWarn: 1,
			},
		},
		{
			name: "only safe workflows is passing score",
			workflowPaths: []string{
				".github/workflows/github-workflow-dangerous-pattern-safe-trigger.yml",
				".github/workflows/github-workflow-dangerous-pattern-trusted-checkout.yml",
			},
			err: nil,
			expected: scut.TestReturn{
				Score: checker.MaxResultScore,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			ctrl := gomock.NewController(t)
			mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
			mockRepoClient.EXPECT().ListFiles(gomock.Any()).Return(tt.workflowPaths, nil)
			mockRepoClient.EXPECT().GetFileReader(gomock.Any()).DoAndReturn(func(file string) (io.ReadCloser, error) {
				return os.Open("./testdata/" + file)
			}).AnyTimes()

			req := &checker.CheckRequest{
				Ctx:        t.Context(),
				RepoClient: mockRepoClient,
				Dlogger:    &dl,
			}

			result := DangerousWorkflow(req)
			scut.ValidateTestReturn(t, tt.name, &tt.expected, &result, &dl)
		})
	}
}
