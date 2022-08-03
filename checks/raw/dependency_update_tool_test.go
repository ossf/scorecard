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
	type args struct {
		name string
		data *[]checker.Tool
	}
	//nolint
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "check dependency file exists",
			args: args{
				name: ".github/dependabot.yml",
				data: &[]checker.Tool{},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: ".other",
			args: args{
				name: ".other",
				data: &[]checker.Tool{},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: ".github/renovate.json",
			args: args{
				name: ".github/renovate.json",
				data: &[]checker.Tool{},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: ".github/renovate.json5",
			args: args{
				name: ".github/renovate.json5",
				data: &[]checker.Tool{},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: ".renovaterc.json",
			args: args{
				name: ".renovaterc.json",
				data: &[]checker.Tool{},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "renovate.json",
			args: args{
				name: "renovate.json",
				data: &[]checker.Tool{},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "renovate.json5",
			args: args{
				name: "renovate.json5",
				data: &[]checker.Tool{},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: ".renovaterc",
			args: args{
				name: ".renovaterc",
				data: &[]checker.Tool{},
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := checkDependencyFileExists(tt.args.name, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkDependencyFileExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("checkDependencyFileExists() = %v, want %v for test %v", got, tt.want, tt.name)
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
		wantErr           bool
		want              int
		mergedPRs         []clients.PullRequest
		callListMergedPRs int
		files             []string
	}{
		{
			name:              "dependency update tool",
			wantErr:           false,
			want:              1,
			mergedPRs:         []clients.PullRequest{},
			callListMergedPRs: 0,
			files: []string{
				".github/dependabot.yml",
			},
		},
		{
			name:              "dependency update tool",
			wantErr:           false,
			want:              1,
			mergedPRs:         []clients.PullRequest{},
			callListMergedPRs: 0,
			files: []string{
				".github/dependabot.yaml",
			},
		},
		{
			name:              "foo bar",
			wantErr:           false,
			want:              0,
			mergedPRs:         []clients.PullRequest{},
			callListMergedPRs: 1,
			files: []string{
				".github/foobar.yml",
			},
		},
		{
			name:              "dependency update tool PRs",
			wantErr:           false,
			want:              1,
			mergedPRs:         []clients.PullRequest{{Author: clients.User{ID: dependabotID}}},
			callListMergedPRs: 1,
			files:             []string{},
		},
		{
			name:              "dependency update tool mix",
			wantErr:           false,
			want:              1,
			mergedPRs:         []clients.PullRequest{{Author: clients.User{ID: dependabotID}}},
			callListMergedPRs: 0,
			files: []string{
				".github/dependabot.yaml",
			},
		},
		{
			name:              "dependency update tool none on both",
			wantErr:           false,
			want:              0,
			mergedPRs:         []clients.PullRequest{{Author: clients.User{ID: 1111111}}},
			callListMergedPRs: 1,
			files: []string{
				".github/foobar.yml",
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
