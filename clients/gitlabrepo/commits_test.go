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
	"testing"
)

func Test_Setup(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		name    string
		repourl string
		commit  string
		depth   int
	}{
		{
			name:    "check that listcommits works",
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

			client, err := CreateGitlabClientWithToken(context.Background(), "", repo)
			if err != nil {
				t.Error("couldn't make gitlab client", err)
			}

			err = client.InitRepo(repo, tt.commit, tt.depth)
			if err != nil {
				t.Error("couldn't init gitlab repo", err)
			}

			c, err := client.ListCommits()
			if err != nil {
				t.Error("couldn't list gitlab repo commits", err)
			}
			if len(c) == 0 {
				t.Error("couldn't get any commits from gitlab repo")
			}
		})
	}
}
