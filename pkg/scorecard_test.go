// Copyright 2020 Security Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package pkg

import (
	"context"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/clients/localdir"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
	"github.com/ossf/scorecard/v4/log"
)

func Test_getRepoCommitHash(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{
			name:    "success",
			want:    "abcdef",
			wantErr: false,
		},
		{
			name:    "empty commit",
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
			defer ctrl.Finish()
			mockRepoClient.EXPECT().ListCommits().DoAndReturn(func() ([]clients.Commit, error) {
				if tt.want == "" {
					return []clients.Commit{}, nil
				}
				return []clients.Commit{
					{
						SHA: tt.want,
					},
				}, nil
			})

			got, err := getRepoCommitHash(mockRepoClient)
			if (err != nil) != tt.wantErr {
				t.Errorf("getRepoCommitHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getRepoCommitHash() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getRepoCommitHashLocal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{
			name:    "local directory",
			path:    "testdata",
			want:    "unknown",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := log.NewLogger(log.DebugLevel)
			localDirClient := localdir.CreateLocalDirClient(context.Background(), logger)
			localRepo, err := localdir.MakeLocalDirRepo("testdata")
			if err != nil {
				t.Errorf("MakeLocalDirRepo: %v", err)
				return
			}
			if err := localDirClient.InitRepo(localRepo, clients.HeadSHA, 30); err != nil {
				t.Errorf("InitRepo: %v", err)
				return
			}

			got, err := getRepoCommitHash(localDirClient)
			if (err != nil) != tt.wantErr {
				t.Errorf("getRepoCommitHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getRepoCommitHash() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRunScorecards(t *testing.T) {
	t.Parallel()
	type args struct {
		commitSHA string
	}
	tests := []struct {
		name    string
		args    args
		want    ScorecardResult
		wantErr bool
	}{
		{
			name: "empty commits repos should return empty result",
			args: args{
				commitSHA: "",
			},
			want:    ScorecardResult{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
			repo := mockrepo.NewMockRepo(ctrl)

			mockRepoClient.EXPECT().InitRepo(repo, tt.args.commitSHA, 30).Return(nil)

			mockRepoClient.EXPECT().Close().DoAndReturn(func() error {
				return nil
			})

			mockRepoClient.EXPECT().ListCommits().DoAndReturn(func() ([]clients.Commit, error) {
				if tt.args.commitSHA == "" {
					return []clients.Commit{}, nil
				}
				return []clients.Commit{
					{
						SHA: tt.args.commitSHA,
					},
				}, nil
			})
			defer ctrl.Finish()
			got, err := RunScorecards(context.Background(), repo, tt.args.commitSHA, nil, mockRepoClient, nil, nil, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunScorecards() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RunScorecards() got = %v, want %v", got, tt.want)
			}
		})
	}
}
