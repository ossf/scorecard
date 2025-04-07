// Copyright 2023 OpenSSF Scorecard Authors
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

package gitlabrepo

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/ossf/scorecard/v5/clients"
)

func TestCheckRepoInaccessible(t *testing.T) {
	t.Parallel()

	tests := []struct {
		want error
		repo *gitlab.Project
		name string
	}{
		{
			name: "if repo is enabled then it is accessible",
			repo: &gitlab.Project{
				RepositoryAccessLevel: gitlab.EnabledAccessControl,
			},
		},
		{
			name: "repo should not have public access in this case, but if it does it is accessible",
			repo: &gitlab.Project{
				RepositoryAccessLevel: gitlab.PublicAccessControl,
			},
		},
		{
			name: "if repo is disabled then is inaccessible",
			repo: &gitlab.Project{
				RepositoryAccessLevel: gitlab.DisabledAccessControl,
			},
			want: errRepoAccess,
		},
		{
			name: "if repo is private then it is accessible",
			repo: &gitlab.Project{
				RepositoryAccessLevel: gitlab.PrivateAccessControl,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := checkRepoInaccessible(tt.repo)
			if !errors.Is(got, tt.want) {
				t.Errorf("checkRepoInaccessible() got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListCommits(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		responsePath string
		commits      []clients.Commit
		wantErr      bool
	}{
		{
			name:         "Error in ListRawCommits",
			responsePath: "./testdata/invalid-commits",
			commits:      []clients.Commit{},
			wantErr:      true,
		},
		{
			name:         "No commits in repo",
			responsePath: "./testdata/empty-response",
			commits:      []clients.Commit{},
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			httpClient := &http.Client{
				Transport: suffixStubTripper{
					responsePaths: map[string]string{
						"commits": tt.responsePath, // corresponds to projects/<id>/repository/commits
					},
				},
			}
			glclient, err := gitlab.NewClient("", gitlab.WithHTTPClient(httpClient))
			if err != nil {
				t.Fatalf("gitlab.NewClient error: %v", err)
			}
			commitshandler := &commitsHandler{
				glClient: glclient,
			}

			repoURL := Repo{
				owner:     "ossf-tests",
				commitSHA: clients.HeadSHA,
			}

			commitshandler.init(&repoURL, 30)

			gqlhandler := graphqlHandler{
				client: httpClient,
			}
			gqlhandler.init(context.Background(), &repoURL)

			client := &Client{glClient: glclient, commits: commitshandler, graphql: &gqlhandler}

			got, Err := client.ListCommits()

			if (Err != nil) != tt.wantErr {
				t.Fatalf("ListCommits, wanted Error: %v, got Error: %v", tt.wantErr, Err)
			}
			if !cmp.Equal(got, tt.commits) {
				t.Errorf("ListCommits() got %v, want %v", got, tt.commits)
			}
		})
	}
}
