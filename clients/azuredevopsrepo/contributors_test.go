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
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"

	"github.com/ossf/scorecard/v5/clients"
)

func Test_listContributors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		getCommits   fnGetCommits
		wantContribs []clients.User
		wantErr      bool
	}{
		{
			name: "no commits",
			getCommits: func(ctx context.Context, args git.GetCommitsArgs) (*[]git.GitCommitRef, error) {
				return &[]git.GitCommitRef{}, nil
			},
			wantContribs: nil,
			wantErr:      false,
		},
		{
			name: "single contributor",
			getCommits: func(ctx context.Context, args git.GetCommitsArgs) (*[]git.GitCommitRef, error) {
				if *args.SearchCriteria.Skip == 0 {
					return &[]git.GitCommitRef{
						{
							Author: &git.GitUserDate{
								Email: toPtr("test@example.com"),
							},
						},
					}, nil
				} else {
					return &[]git.GitCommitRef{}, nil
				}
			},
			wantContribs: []clients.User{
				{
					Login:            "test@example.com",
					Companies:        []string{"testOrg"},
					NumContributions: 1,
				},
			},
			wantErr: false,
		},
		{
			name: "multiple contributors",
			getCommits: func(ctx context.Context, args git.GetCommitsArgs) (*[]git.GitCommitRef, error) {
				if *args.SearchCriteria.Skip == 0 {
					return &[]git.GitCommitRef{
						{
							Author: &git.GitUserDate{
								Email: toPtr("test@example.com"),
							},
						},
						{
							Author: &git.GitUserDate{
								Email: toPtr("test2@example.com"),
							},
						},
						{
							Author: &git.GitUserDate{
								Email: toPtr("test2@example.com"),
							},
						},
					}, nil
				} else {
					return &[]git.GitCommitRef{}, nil
				}
			},
			wantContribs: []clients.User{
				{
					Login:            "test@example.com",
					Companies:        []string{"testOrg"},
					NumContributions: 1,
				},
				{
					Login:            "test2@example.com",
					Companies:        []string{"testOrg"},
					NumContributions: 2,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := contributorsHandler{
				ctx:  context.Background(),
				once: new(sync.Once),
				repourl: &Repo{
					organization: "testOrg",
				},
				getCommits: tt.getCommits,
			}
			err := c.setup()
			if (err != nil) != tt.wantErr {
				t.Errorf("contributorsHandler.setup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.wantContribs, c.contributors, cmpopts.SortSlices(func(a, b string) bool { return a < b })); diff != "" {
				t.Errorf("contributorsHandler.setup() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
