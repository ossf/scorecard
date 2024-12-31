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

package azuredevopsrepo

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"

	"github.com/ossf/scorecard/v5/clients"
)

func Test_searchCommits(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                string
		commitSearchOptions clients.SearchCommitsOptions
		getCommits          fnGetCommits
		getPullRequestQuery fnGetPullRequestQuery
		want                []clients.Commit
		wantErr             bool
	}{
		{
			name: "empty response",
			commitSearchOptions: clients.SearchCommitsOptions{
				Author: "test",
			},
			getCommits: func(ctx context.Context, args git.GetCommitsArgs) (*[]git.GitCommitRef, error) {
				return &[]git.GitCommitRef{}, nil
			},
			getPullRequestQuery: func(ctx context.Context, args git.GetPullRequestQueryArgs) (*git.GitPullRequestQuery, error) {
				return &git.GitPullRequestQuery{}, nil
			},
			want:    []clients.Commit{},
			wantErr: false,
		},
		{
			name: "single commit",
			commitSearchOptions: clients.SearchCommitsOptions{
				Author: "test",
			},
			getCommits: func(ctx context.Context, args git.GetCommitsArgs) (*[]git.GitCommitRef, error) {
				return &[]git.GitCommitRef{
					{
						CommitId: toPtr("4b825dc642cb6eb9a060e54bf8d69288fbee4904"),
						Comment:  toPtr("Initial commit"),
						Committer: &git.GitUserDate{
							Email: toPtr("test@example.com"),
							Date: &azuredevops.Time{
								Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
							},
						},
					},
				}, nil
			},
			getPullRequestQuery: func(ctx context.Context, args git.GetPullRequestQueryArgs) (*git.GitPullRequestQuery, error) {
				return &git.GitPullRequestQuery{}, nil
			},
			want: []clients.Commit{
				{
					SHA:                    "4b825dc642cb6eb9a060e54bf8d69288fbee4904",
					Message:                "Initial commit",
					Committer:              clients.User{Login: "test@example.com"},
					CommittedDate:          time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					AssociatedMergeRequest: clients.PullRequest{},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := searchCommitsHandler{
				ctx:                 context.Background(),
				repourl:             &Repo{},
				getCommits:          tt.getCommits,
				getPullRequestQuery: tt.getPullRequestQuery,
			}

			got, err := s.searchCommits(tt.commitSearchOptions)
			if (err != nil) != tt.wantErr {
				t.Errorf("searchCommits() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("searchCommits() got = %v, want %v", got, tt.want)
			}
		})
	}
}
