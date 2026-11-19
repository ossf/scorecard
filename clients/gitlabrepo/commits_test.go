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
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/ossf/scorecard/v5/clients"
)

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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := parseEmailToName(tt.email)

			if tt.expected != result {
				t.Errorf("Parser didn't work as expected: %s != %s", tt.expected, result)
			}
		})
	}
}

func TestListRawCommits(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		commitsPath string
		commitDepth int
		want        int
		wantErr     bool
	}{
		{
			name:        "commits to non-default depth",
			commitsPath: "./testdata/valid-commits",
			commitDepth: 17,
			want:        17,
			wantErr:     false,
		},
		{
			name:        "commits to default depth",
			commitsPath: "./testdata/valid-commits",
			commitDepth: 30,
			want:        30,
			wantErr:     false,
		},
		{
			name:        "request more commits than exist in the repo",
			commitsPath: "./testdata/valid-commits",
			commitDepth: 60,
			want:        32,
			wantErr:     false,
		},
		{
			name:        "failure fetching commits",
			commitsPath: "./testdata/invalid-commits",
			commitDepth: 30,
			want:        0,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			httpClient := &http.Client{
				Transport: suffixStubTripper{
					responsePaths: map[string]string{
						"commits": tt.commitsPath, // corresponds to projects/<id>/repository/commits
					},
				},
			}
			client, err := gitlab.NewClient("", gitlab.WithHTTPClient(httpClient))
			if err != nil {
				t.Fatalf("gitlab.NewClient error: %v", err)
			}
			handler := &commitsHandler{
				glClient: client,
			}

			repoURL := Repo{
				owner:     "ossf-tests",
				commitSHA: clients.HeadSHA,
			}

			handler.init(&repoURL, tt.commitDepth)
			commits, err := handler.listRawCommits()
			if (err != nil) != tt.wantErr {
				t.Fatalf("listIssues error: %v, wantedErr: %t", err, tt.wantErr)
			}
			if !cmp.Equal(len(commits), tt.want) {
				t.Errorf("listCommits() = %v, want %v", len(commits), cmp.Diff(len(commits), tt.want))
			}
		})
	}
}

func mergeRequestNode(iid, mergeCommitSHA string, commitSHAs ...string) graphqlMergeRequestNode {
	var mr graphqlMergeRequestNode
	mr.IID = iid
	mr.MergeCommitSHA = mergeCommitSHA
	for _, sha := range commitSHAs {
		mr.Commits.Nodes = append(mr.Commits.Nodes, struct {
			SHA string `graphql:"sha"`
		}{SHA: sha})
	}
	return mr
}

func TestZipLinksSquashCommits(t *testing.T) {
	t.Parallel()

	var data graphqlData
	data.Project.MergeRequests.Nodes = []graphqlMergeRequestNode{
		// Squash merge: the squash commit isn't in the MR's commit list.
		mergeRequestNode("1", "merge1", "source1"),
		// Ordinary merge: the MR's commit is on the branch.
		mergeRequestNode("2", "merge2", "source2"),
		// Fast-forward merge without a merge commit.
		mergeRequestNode("3", "", "source3"),
	}

	now := time.Now()
	commit := func(sha string, parents ...string) *gitlab.Commit {
		return &gitlab.Commit{ID: sha, ParentIDs: parents, CommittedDate: &now}
	}
	commitsRaw := []*gitlab.Commit{
		commit("direct", "other"),
		commit("other", "merge2", "unrelated"),
		commit("merge2", "merge1", "source2"),
		commit("source2", "merge1"),
		commit("merge1", "source3", "squash1"),
		commit("squash1", "source3"),
		commit("source3", "base"),
	}

	handler := &commitsHandler{}
	got := map[string]int{}
	for _, c := range handler.zip(commitsRaw, data) {
		got[c.SHA] = c.AssociatedMergeRequest.Number
	}

	want := map[string]int{
		"direct":  0,
		"other":   0,
		"merge2":  2,
		"source2": 2,
		"merge1":  1,
		"squash1": 1,
		"source3": 3,
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("zip() merge request numbers mismatch (-want +got):\n%s", diff)
	}
}
