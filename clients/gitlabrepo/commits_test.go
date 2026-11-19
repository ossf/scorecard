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
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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

//nolint:govet
func Test_commitsHandler_zip(t *testing.T) {
	t.Parallel()

	commitDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	commitsRaw := []*gitlab.Commit{
		{
			ID:            "sha-1",
			CommittedDate: &commitDate,
			Message:       "commit 1",
		},
	}

	mrNode := graphqlMergeRequestNode{
		IID:      "10",
		MergedAt: commitDate,
	}
	mrNode.Author.Username = "alice"
	mrNode.Commits.Nodes = []struct {
		SHA string `graphql:"sha"`
	}{
		{SHA: "sha-1"},
	}
	mrNode.Reviewers.Nodes = []struct {
		Username                string    `graphql:"username"`
		ID                      GitlabGID `graphql:"id"`
		MergeRequestInteraction struct {
			ReviewState string `graphql:"reviewState"`
		} `graphql:"mergeRequestInteraction"`
	}{
		{
			Username: "reviewer-commented",
			MergeRequestInteraction: struct {
				ReviewState string `graphql:"reviewState"`
			}{ReviewState: "UNREVIEWED"},
		},
		{
			Username: "reviewer-approved-1",
			MergeRequestInteraction: struct {
				ReviewState string `graphql:"reviewState"`
			}{ReviewState: "REVIEWED"},
		},
		{
			Username: "reviewer-approved-2",
			MergeRequestInteraction: struct {
				ReviewState string `graphql:"reviewState"`
			}{ReviewState: "REVIEWED"},
		},
	}
	mrNode.Approvers.Nodes = []struct {
		Username string    `graphql:"username"`
		ID       GitlabGID `graphql:"id"`
	}{
		{Username: "approver-1"},
		{Username: "approver-2"},
	}

	var data graphqlData
	data.Project.MergeRequests.Nodes = []graphqlMergeRequestNode{mrNode}

	handler := &commitsHandler{}
	commits := handler.zip(commitsRaw, data)
	if len(commits) != 1 {
		t.Fatalf("expected 1 commit, got %d", len(commits))
	}

	gotReviews := commits[0].AssociatedMergeRequest.Reviews
	wantReviews := []clients.Review{
		{Author: &clients.User{Login: "approver-1"}, State: "APPROVED"},
		{Author: &clients.User{Login: "approver-2"}, State: "APPROVED"},
		{Author: &clients.User{Login: "reviewer-approved-1"}, State: "APPROVED"},
		{Author: &clients.User{Login: "reviewer-approved-2"}, State: "APPROVED"},
		{Author: &clients.User{Login: "reviewer-commented"}, State: "COMMENTED"},
	}

	sortReviews := func(a, b clients.Review) int {
		return strings.Compare(a.Author.Login, b.Author.Login)
	}
	slices.SortFunc(gotReviews, sortReviews)
	slices.SortFunc(wantReviews, sortReviews)

	if diff := cmp.Diff(wantReviews, gotReviews, cmpopts.IgnoreUnexported(clients.Review{})); diff != "" {
		t.Errorf("zip() reviews mismatch (-want +got):\n%s", diff)
	}
}
