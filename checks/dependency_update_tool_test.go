// Copyright 2020 OpenSSF Scorecard Authors
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

	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	clients "github.com/ossf/scorecard/v5/clients"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	scut "github.com/ossf/scorecard/v5/utests"
)

const (
	dependabotID = 49699333
)

// TestDependencyUpdateTool tests the DependencyUpdateTool checker.
func TestDependencyUpdateTool(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name              string
		want              checker.CheckResult
		SearchCommits     []clients.Commit
		files             []string
		expected          scut.TestReturn
		CallSearchCommits int
		wantErr           bool
	}{
		{
			name:    "dependabot config detected",
			wantErr: false,
			files: []string{
				".github/dependabot.yml",
			},
			CallSearchCommits: 0,
			expected: scut.TestReturn{
				NumberOfInfo: 1,
				NumberOfWarn: 0,
				Score:        10,
			},
		},
		{
			name:    "dependabot alternate yaml extension detected",
			wantErr: false,
			files: []string{
				".github/dependabot.yaml",
			},
			CallSearchCommits: 0,
			expected: scut.TestReturn{
				NumberOfInfo: 1,
				NumberOfWarn: 0,
				Score:        10,
			},
		},
		{
			name:    "renovatebot config detected",
			wantErr: false,
			files: []string{
				"renovate.json",
			},
			CallSearchCommits: 0,
			expected: scut.TestReturn{
				NumberOfInfo: 1,
				NumberOfWarn: 0,
				Score:        10,
			},
		},
		{
			name:    "alternate renovatebot config detected",
			wantErr: false,
			files: []string{
				".github/renovate.json5",
			},
			CallSearchCommits: 0,
			expected: scut.TestReturn{
				NumberOfInfo: 1,
				NumberOfWarn: 0,
				Score:        10,
			},
		},
		{
			name:    "pyup config detected",
			wantErr: false,
			files: []string{
				".pyup.yml",
			},
			CallSearchCommits: 0,
			expected: scut.TestReturn{
				NumberOfInfo: 1,
				NumberOfWarn: 0,
				Score:        10,
			},
		},
		{
			name:              "random committer ID not detected as dependecy tool bot",
			wantErr:           false,
			files:             []string{},
			SearchCommits:     []clients.Commit{{Committer: clients.User{ID: 111111111}}},
			CallSearchCommits: 1,
			expected: scut.TestReturn{
				NumberOfWarn: 1,
			},
		},
		{
			name:    "random yaml file not detected as update tool config",
			wantErr: false,
			files: []string{
				".github/foobar.yml",
			},
			SearchCommits:     []clients.Commit{},
			CallSearchCommits: 1,
			expected: scut.TestReturn{
				NumberOfWarn: 1,
			},
		},
		{
			name:    "dependabot found in recent commits",
			wantErr: false,
			files: []string{
				".github/foobar.yaml",
			},
			SearchCommits:     []clients.Commit{{Committer: clients.User{ID: dependabotID}}},
			CallSearchCommits: 1,
			expected: scut.TestReturn{
				NumberOfInfo: 1,
				NumberOfWarn: 0,
				Score:        10,
			},
		},
		{
			name:    "dependabot bot found in recent commits 2",
			wantErr: false,
			files:   []string{},
			SearchCommits: []clients.Commit{
				{Committer: clients.User{ID: 111111111}},
				{Committer: clients.User{ID: dependabotID}},
			},
			CallSearchCommits: 1,
			expected: scut.TestReturn{
				NumberOfInfo: 1,
				NumberOfWarn: 0,
				Score:        10,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockRepo := mockrepo.NewMockRepoClient(ctrl)
			mockRepo.EXPECT().ListFiles(gomock.Any()).Return(tt.files, nil)
			mockRepo.EXPECT().SearchCommits(gomock.Any()).Return(tt.SearchCommits, nil).Times(tt.CallSearchCommits)
			dl := scut.TestDetailLogger{}
			c := &checker.CheckRequest{
				RepoClient: mockRepo,
				Dlogger:    &dl,
			}
			res := DependencyUpdateTool(c)

			scut.ValidateTestReturn(t, tt.name, &tt.expected, &res, &dl)
		})
	}
}

func TestDependencyUpdateTool_noSearchCommits(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	mockRepo := mockrepo.NewMockRepoClient(ctrl)
	files := []string{"README.md"}
	mockRepo.EXPECT().ListFiles(gomock.Any()).Return(files, nil)
	mockRepo.EXPECT().SearchCommits(gomock.Any()).Return(nil, clients.ErrUnsupportedFeature)
	dl := scut.TestDetailLogger{}
	c := &checker.CheckRequest{
		RepoClient: mockRepo,
		Dlogger:    &dl,
	}
	got := DependencyUpdateTool(c)
	if got.Error != nil {
		t.Errorf("got: %v, wanted ErrUnsupportedFeature not to propagate", got.Error)
	}
}
