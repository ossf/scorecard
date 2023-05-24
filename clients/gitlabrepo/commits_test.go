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
	"sync"
	"testing"
	"time"

	"github.com/xanzy/go-gitlab"
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

func TestParsingEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{
			name:     "Perfect Match Email Parser",
			email:    "john.doe@nowhere.com",
			expected: "john doe",
		},
		{
			name:     "Valid Email Not Formatted as expected",
			email:    "johndoe@nowhere.com",
			expected: "johndoe@nowhere com",
		},
		{
			name:     "Invalid email format",
			email:    "johndoe@nowherecom",
			expected: "johndoe@nowherecom",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := parseEmailToName(tt.email)

			if tt.expected != result {
				t.Errorf("Parser didn't work as expected: %s != %s", tt.expected, result)
			}
		})
	}
}

func TestQueryMergeRequestsByCommit(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		commits []*gitlab.Commit
		mrs     []*gitlab.MergeRequest
	}{
		{
			name: "Commit with no MR",
			commits: []*gitlab.Commit{
				{
					ID:            "10",
					CommittedDate: &time.Time{},
				},
			},
			mrs: []*gitlab.MergeRequest{},
		},
		{
			name: "Commit with a MR with Merged Date",
			commits: []*gitlab.Commit{
				{
					ID:            "10",
					CommittedDate: &time.Time{},
				},
			},
			mrs: []*gitlab.MergeRequest{
				{
					ID:       10,
					MergedAt: gitlab.Time(time.Now()),
					Author: &gitlab.BasicUser{
						ID:       100,
						Username: "no-one",
					},
					MergedBy: &gitlab.BasicUser{
						ID:       101,
						Username: "da-approver",
					},
					Reviewers: []*gitlab.BasicUser{
						{
							ID:       102,
							Username: "da-manager",
						},
					},
				},
			},
		},
		{
			name: "Commit with a MR Not Merged",
			commits: []*gitlab.Commit{
				{
					ID:            "10",
					CommittedDate: &time.Time{},
				},
			},
			mrs: []*gitlab.MergeRequest{
				{
					ID:       10,
					MergedAt: nil,
					Author: &gitlab.BasicUser{
						ID:       100,
						Username: "no-one",
					},
					Reviewers: []*gitlab.BasicUser{},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := commitsHandler{
				once: new(sync.Once),
				repourl: &repoURL{
					projectID: "5000",
				},
				listMergeRequestByCommit: func(pid interface{}, sha string,
					options ...gitlab.RequestOptionFunc) (
					[]*gitlab.MergeRequest, *gitlab.Response, error,
				) {
					return tt.mrs, nil, nil
				},
			}
			handler.once.Do(func() {})

			c := handler.queryMergeRequestsByCommits(tt.commits)

			if len(c) != len(tt.commits) {
				t.Errorf("Return commit count should equal what was sent in")
			}

			if len(tt.mrs) > 0 && tt.mrs[0].MergedAt != nil &&
				c[0].AssociatedMergeRequest.Author.ID != int64(tt.mrs[0].Author.ID) {
				t.Errorf("MR should have been associated to the commit")
			}

			if len(tt.mrs) > 0 && tt.mrs[0].MergedAt == nil &&
				c[0].AssociatedMergeRequest.Author.ID != 0 {
				t.Errorf("MR should NOT have been associated to the commit")
			}
		})
	}
}
