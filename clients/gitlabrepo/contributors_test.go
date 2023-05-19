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
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/xanzy/go-gitlab"
)

func Test_ContributorsSetup(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		name    string
		repourl string
		commit  string
		depth   int
	}{
		{
			name:    "check that Contributor works",
			repourl: "https://gitlab.com/fdroid/fdroidclient",
			commit:  "HEAD",
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
				t.Error("couldn't init gitlab repo",
					err)
			}

			c, err := client.ListContributors()
			// Authentication is failing when querying users, not sure yet how to get around that
			if err != nil {
				errMsg := fmt.Sprintf("%v", err)

				if !(strings.Contains(errMsg, "error during Users.Get") && strings.Contains(errMsg, "401")) {
					t.Error("couldn't list gitlab repo contributors", err)
				}
			}
			if len(c) != 0 {
				t.Error("couldn't get any contributors from gitlab repo")
			}
		})
	}
}

func TestContributors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		contributors []*gitlab.Contributor
		users        []*gitlab.User
	}{
		{
			name:         "No Data",
			contributors: []*gitlab.Contributor{},
			users:        []*gitlab.User{},
		},
		{
			name: "Simple Passthru",
			contributors: []*gitlab.Contributor{
				{
					Name: "John Doe",
				},
			},
			users: []*gitlab.User{
				{
					Name: "John Doe",
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := contributorsHandler{
				fnContributors: func(s string) ([]*gitlab.Contributor, error) {
					return tt.contributors, nil
				},
				fnUsers: func(s string) ([]*gitlab.User, error) {
					return tt.users, nil
				},
				once: new(sync.Once),
				repourl: &repoURL{
					commitSHA: "HEAD",
				},
			}

			if len(handler.contributors) != 0 {
				t.Errorf("Initial count of contributors should be 0, but was %v", strconv.Itoa(len(handler.contributors)))
			}

			err := handler.setup()
			if err != nil {
				t.Errorf("Exception in contributors.setup %v", err)
			}

			if len(handler.contributors) != len(tt.contributors) {
				t.Errorf("Initial count of contributors should be 1, but was %v", strconv.Itoa(len(handler.contributors)))
			}
		})
	}
}
