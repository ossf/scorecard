// Copyright 2020 OpenSSF Scorecard Authors
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
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

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
			wantErr: true,
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
			if err := localDirClient.InitRepo(localRepo, clients.HeadSHA, 0); err != nil {
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

func TestRunScorecard(t *testing.T) {
	t.Parallel()
	type args struct {
		uri       string
		commitSHA string
	}
	tests := []struct {
		name    string
		args    args
		want    ScorecardResult
		wantErr bool
	}{
		{
			name: "empty commits repos should return repo details but no checks",
			args: args{
				uri:       "github.com/ossf/scorecard",
				commitSHA: "",
			},
			want: ScorecardResult{
				Repo: RepoInfo{
					Name: "github.com/ossf/scorecard",
				},
				Scorecard: ScorecardInfo{
					CommitSHA: "unknown",
				},
			},
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

			repo.EXPECT().URI().Return(tt.args.uri).AnyTimes()

			mockRepoClient.EXPECT().InitRepo(repo, tt.args.commitSHA, 0).Return(nil)

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
			got, err := RunScorecard(context.Background(), repo, tt.args.commitSHA, 0, nil, mockRepoClient, nil, nil, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunScorecard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			ignoreDate := cmpopts.IgnoreFields(ScorecardResult{}, "Date")
			if !cmp.Equal(got, tt.want, ignoreDate) {
				t.Errorf("expected %v, got %v", got, cmp.Diff(tt.want, got, ignoreDate))
			}
		})
	}
}
