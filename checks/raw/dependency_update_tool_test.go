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

package raw

import (
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/ossf/scorecard/v4/checker"
	clients "github.com/ossf/scorecard/v4/clients"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
)

func Test_checkDependencyFileExists(t *testing.T) {
	t.Parallel()

	//nolint
	tests := []struct {
		name    string
		path    string
		want    bool
		wantErr bool
	}{
		{
			name:    ".github/dependabot.yml",
			path:    ".github/dependabot.yml",
			want:    true,
			wantErr: false,
		},
		{
			name:    ".github/dependabot.yaml",
			path:    ".github/dependabot.yaml",
			want:    true,
			wantErr: false,
		},
		{
			name:    ".other",
			path:    ".other",
			want:    false,
			wantErr: false,
		},
		{
			name:    ".github/renovate.json",
			path:    ".github/renovate.json",
			want:    true,
			wantErr: false,
		},
		{
			name:    ".github/renovate.json5",
			path:    ".github/renovate.json5",
			want:    true,
			wantErr: false,
		},
		{
			name:    ".renovaterc.json",
			path:    ".renovaterc.json",
			want:    true,
			wantErr: false,
		},
		{
			name:    "renovate.json",
			path:    "renovate.json",
			want:    true,
			wantErr: false,
		},
		{
			name:    "renovate.json5",
			path:    "renovate.json5",
			want:    true,
			wantErr: false,
		},
		{
			name:    ".renovaterc",
			path:    ".renovaterc",
			want:    true,
			wantErr: false,
		},
		{
			name:    ".pyup.yml",
			path:    ".pyup.yml",
			want:    true,
			wantErr: false,
		},
		{
			name:    ".lift.toml",
			path:    ".lift.toml",
			want:    true,
			wantErr: false,
		},
		{
			name:    ".lift/config.toml",
			path:    ".lift/config.toml",
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			results := []checker.Tool{}
			cont, err := checkDependencyFileExists(tt.path, &results)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkDependencyFileExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cont {
				t.Errorf("continue is false for %v", tt.name)
			}
			if tt.want != (len(results) == 1) {
				t.Errorf("checkDependencyFileExists() = %v, want %v for test %v", len(results), tt.want, tt.name)
			}
		})
	}
}

// TestDependencyUpdateTool tests the DependencyUpdateTool function.
func TestDependencyUpdateTool(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name              string
		SearchCommits     []clients.Commit
		files             []string
		want              int
		CallSearchCommits int
		wantErr           bool
	}{
		{
			name:              "dependency update tool",
			wantErr:           false,
			want:              1,
			CallSearchCommits: 0,
			files: []string{
				".github/dependabot.yml",
			},
		},
		{
			name:              "dependency update tool",
			wantErr:           false,
			want:              1,
			CallSearchCommits: 0,
			files: []string{
				".github/dependabot.yaml",
			},
		},
		{
			name:              "foo bar",
			wantErr:           false,
			want:              0,
			CallSearchCommits: 1,
			SearchCommits:     []clients.Commit{{Committer: clients.User{ID: 111111111}}},
			files: []string{
				".github/foobar.yml",
			},
		},
		{
			name:              "dependency update tool via commits",
			wantErr:           false,
			want:              1,
			CallSearchCommits: 1,
			files:             []string{},
			SearchCommits:     []clients.Commit{{Committer: clients.User{ID: dependabotID}}},
		},
		{
			name:              "dependency update tool via commits",
			wantErr:           false,
			want:              1,
			CallSearchCommits: 1,
			files:             []string{},
			SearchCommits: []clients.Commit{
				{Committer: clients.User{ID: 111111111}},
				{Committer: clients.User{ID: dependabotID}},
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
			mockRepo.EXPECT().SearchCommits(gomock.Any()).Return(tt.SearchCommits, nil).Times(tt.CallSearchCommits)

			got, err := DependencyUpdateTool(mockRepo)
			if (err != nil) != tt.wantErr {
				t.Errorf("DependencyUpdateTool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got.Tools) != tt.want {
					t.Errorf("DependencyUpdateTool() = %v, want %v", got.Tools, tt.want)
				}
			}
		})
	}
}
