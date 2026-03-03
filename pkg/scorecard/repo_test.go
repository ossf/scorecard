// Copyright 2026 OpenSSF Scorecard Authors
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

package scorecard

import (
	"testing"

	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/clients/azuredevopsrepo"
	"github.com/ossf/scorecard/v5/clients/githubrepo"
	"github.com/ossf/scorecard/v5/clients/gitlabrepo"
	"github.com/ossf/scorecard/v5/clients/localdir"
)

func TestGetRepoType(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		repo clients.Repo
		want RepoType
	}{
		{
			name: "local directory repo",
			repo: &localdir.Repo{},
			want: RepoLocal,
		},
		{
			name: "github repo",
			repo: &githubrepo.Repo{},
			want: RepoGitHub,
		},
		{
			name: "gitlab repo",
			repo: &gitlabrepo.Repo{},
			want: RepoGitLab,
		},
		{
			name: "azure devops repo",
			repo: &azuredevopsrepo.Repo{},
			want: RepoAzureDevOPs,
		},
		{
			name: "unknown repo type",
			repo: nil,
			want: RepoUnknown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := GetRepoType(tt.repo)
			if got != tt.want {
				t.Errorf("getRepoType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepoTypeFromString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		repo string
		want RepoType
	}{
		{
			name: "local lowercase",
			repo: "local",
			want: RepoLocal,
		},
		{
			name: "local uppercase",
			repo: "LOCAL",
			want: RepoLocal,
		},
		{
			name: "local mixed case",
			repo: "LoCaL",
			want: RepoLocal,
		},
		{
			name: "local with whitespace",
			repo: "  local  ",
			want: RepoLocal,
		},
		{
			name: "git-local",
			repo: "git-local",
			want: RepoGitLocal,
		},
		{
			name: "github lowercase",
			repo: "github",
			want: RepoGitHub,
		},
		{
			name: "github uppercase",
			repo: "GITHUB",
			want: RepoGitHub,
		},
		{
			name: "github mixedcase",
			repo: "GitHub",
			want: RepoGitHub,
		},
		{
			name: "gitlab lowercase",
			repo: "gitlab",
			want: RepoGitLab,
		},
		{
			name: "gitlab uppercase",
			repo: "GITLAB",
			want: RepoGitLab,
		},
		{
			name: "gitlab mixedcase",
			repo: "GitLab",
			want: RepoGitLab,
		},
		{
			name: "azuredevops lowercase",
			repo: "azuredevops",
			want: RepoAzureDevOPs,
		},
		{
			name: "azuredevops uppercase",
			repo: "AZUREDEVOPS",
			want: RepoAzureDevOPs,
		},
		{
			name: "unknown type",
			repo: "unknown",
			want: RepoUnknown,
		},
		{
			name: "empty string",
			repo: "",
			want: RepoUnknown,
		},
		{
			name: "invalid type",
			repo: "bitbucket",
			want: RepoUnknown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := RepoTypeFromString(tt.repo)
			if got != tt.want {
				t.Errorf("repoTypeFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}
