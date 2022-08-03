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
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/ossf/scorecard/v4/checker"
	clients "github.com/ossf/scorecard/v4/clients"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
	scut "github.com/ossf/scorecard/v4/utests"
)

const (
	dependabotID = 49699333
)

// TestDependencyUpdateTool tests the DependencyUpdateTool checker.
func TestDependencyUpdateTool(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name              string
		wantErr           bool
		mergedPRs         []clients.PullRequest
		files             []string
		want              checker.CheckResult
		callListMergedPRs int
		expected          scut.TestReturn
	}{
		{
			name:    "dependency yml",
			wantErr: false,
			files: []string{
				".github/dependabot.yml",
			},
			callListMergedPRs: 0,
			expected: scut.TestReturn{
				NumberOfInfo: 1,
				Score:        10,
			},
		},
		{
			name:    "dependency yaml ",
			wantErr: false,
			files: []string{
				".github/dependabot.yaml",
			},
			callListMergedPRs: 0,
			expected: scut.TestReturn{
				NumberOfInfo: 1,
				Score:        10,
			},
		},
		{
			name:    "foo bar",
			wantErr: false,
			files: []string{
				".github/foobar.yml",
			},
			callListMergedPRs: 1,
			expected: scut.TestReturn{
				NumberOfWarn: 2,
			},
		},
		{
			name:    "dependabot PR",
			wantErr: false,
			mergedPRs: []clients.PullRequest{
				{Author: clients.User{ID: dependabotID}},
			},
			callListMergedPRs: 1,
			expected: scut.TestReturn{
				NumberOfInfo: 1,
				Score:        10,
			},
		},
		{
			name:    "both",
			wantErr: false,
			files: []string{
				".github/dependabot.yml",
			},
			mergedPRs: []clients.PullRequest{
				{Author: clients.User{ID: dependabotID}},
			},
			callListMergedPRs: 0,
			expected: scut.TestReturn{
				NumberOfInfo: 1,
				Score:        10,
			},
		},
		{
			name:    "foo PR",
			wantErr: false,
			mergedPRs: []clients.PullRequest{
				{Author: clients.User{ID: 111111111}},
			},
			callListMergedPRs: 1,
			expected: scut.TestReturn{
				NumberOfWarn: 2,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockRepo := mockrepo.NewMockRepoClient(ctrl)
			mockRepo.EXPECT().ListFiles(gomock.Any()).Return(tt.files, nil)
			mockRepo.EXPECT().ListMergedPRs().Return(tt.mergedPRs, nil).Times(tt.callListMergedPRs)
			dl := scut.TestDetailLogger{}
			c := &checker.CheckRequest{
				RepoClient: mockRepo,
				Dlogger:    &dl,
			}
			res := DependencyUpdateTool(c)

			if !scut.ValidateTestReturn(t, tt.name, &tt.expected, &res, &dl) {
				t.Fail()
			}
		})
	}
}
