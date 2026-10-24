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
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/shurcooL/graphql"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/ossf/scorecard/v5/clients"
)

func TestLatestCommittedDate(t *testing.T) {
	t.Parallel()

	oldest := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	newest := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)
	middle := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		commits []*gitlab.Commit
		want    *time.Time
	}{
		{
			name: "finds newest commit in unordered response",
			commits: []*gitlab.Commit{
				{CommittedDate: &oldest},
				{CommittedDate: &newest},
				{CommittedDate: &middle},
			},
			want: &newest,
		},
		{
			name: "ignores missing commit dates",
			commits: []*gitlab.Commit{
				{},
				{CommittedDate: &middle},
			},
			want: &middle,
		},
		{
			name:    "returns nil without a commit date",
			commits: []*gitlab.Commit{{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := latestCommittedDate(tt.commits)
			if !cmp.Equal(got, tt.want) {
				t.Errorf("latestCommittedDate() mismatch (-got +want):\n%s", cmp.Diff(got, tt.want))
			}
		})
	}
}

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
		{
			name:         "Commit without committed date",
			responsePath: "./testdata/commit-without-date",
			commits: []clients.Commit{{
				Message: "missing date",
				SHA:     "missing-date",
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			httpClient := &http.Client{
				Transport: suffixStubTripper{
					responsePaths: map[string]string{
						"commits": tt.responsePath, // corresponds to projects/<id>/repository/commits
						"graphql": "./testdata/empty-graphql-response",
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
				host:      "gitlab.example.com",
				commitSHA: clients.HeadSHA,
			}

			commitshandler.init(&repoURL, 30)

			gqlhandler := graphqlHandler{
				client: httpClient,
			}
			gqlhandler.init(t.Context(), &repoURL)
			gqlhandler.graphClient = graphql.NewClient("https://gitlab.example.com/api/graphql", httpClient)

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
