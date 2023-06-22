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
	"fmt"
	"testing"

	"github.com/xanzy/go-gitlab"
)

func Test_checkRepoInaccessible(t *testing.T) {
	tcs := []struct {
		desc    string
		repourl string
	}{
		{
			desc:    "private project",
			repourl: "https://gitlab.com/ossf-test/private-project",
		},
	}

	for _, tt := range tcs {
		t.Run(tt.desc, func(t *testing.T) {
			repo, err := MakeGitlabRepo(tt.repourl)
			if err != nil {
				t.Errorf(tt.repourl, err)
			}

			client, err := CreateGitlabClient(context.Background(), repo.Host())
			if err != nil {
				t.Errorf(tt.repourl, err)
			}

			glRepo, ok := repo.(*repoURL)
			if !ok {
				t.Errorf(tt.repourl, errInputRepoType)
			}

			// Sanity check.
			glcl, ok := client.(*Client)
			if !ok {
				t.Error(tt.repourl, errInputRepoType)
			}

			path := fmt.Sprintf("%s/%s", glRepo.owner, glRepo.project)
			proj, _, err := glcl.glClient.Projects.GetProject(path, &gitlab.GetProjectOptions{})
			if err != nil {
				t.Error(tt.repourl, err)
			}

			if err = checkRepoInaccessible(proj); err == nil {
				t.Errorf(fmt.Sprintf("repo %v was supposed to be unreachable but was reachable", repo))
			}
		})
	}
}

func Test_InitRepo(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		name    string
		repourl string
		commit  string
		depth   int
	}{
		{
			repourl: "https://gitlab.com/fdroid/fdroidclient",
			commit:  "a4bbef5c70fd2ac7c15437a22ef0f9ef0b447d08",
			depth:   20,
		},
	}

	for _, tt := range tcs {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := MakeGitlabRepo(tt.repourl)
			if err != nil {
				t.Error("couldn't make gitlab repo", err)
			}

			client, err := CreateGitlabClient(context.Background(), repo.Host())
			if err != nil {
				t.Error("couldn't make gitlab client", err)
			}

			err = client.InitRepo(repo, tt.commit, tt.depth)
			if err != nil {
				t.Error("couldn't init gitlab repo", err)
			}
		})
	}
}
