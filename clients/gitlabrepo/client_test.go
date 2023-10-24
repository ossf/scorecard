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
	"testing"

	"github.com/xanzy/go-gitlab"
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
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := checkRepoInaccessible(tt.repo)
			if !errors.Is(got, tt.want) {
				t.Errorf("checkRepoInaccessible() got %v, want %v", got, tt.want)
			}
		})
	}
}
