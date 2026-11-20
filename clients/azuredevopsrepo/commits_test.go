// Copyright 2024 OpenSSF Scorecard Authors
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

package azuredevopsrepo

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"

	"github.com/ossf/scorecard/v5/clients"
)

func Test_listStatuses(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		ref         string
		getRefs     fnGetRefs
		getStatuses fnGetStatuses
		want        []clients.Status
		wantErr     bool
	}{
		{
			name: "sha - no statuses",
			ref:  "4b825dc642cb6eb9a060e54bf8d69288fbee4904",
			getStatuses: func(ctx context.Context, args git.GetStatusesArgs) (*[]git.GitStatus, error) {
				return &[]git.GitStatus{}, nil
			},
			want:    []clients.Status{},
			wantErr: false,
		},
		{
			name: "sha - single status",
			ref:  "4b825dc642cb6eb9a060e54bf8d69288fbee4904",
			getStatuses: func(ctx context.Context, args git.GetStatusesArgs) (*[]git.GitStatus, error) {
				return &[]git.GitStatus{
					{
						Context: &git.GitStatusContext{
							Name: toPtr("test"),
						},
						State:     toPtr(git.GitStatusStateValues.Succeeded),
						TargetUrl: toPtr("https://example.com"),
					},
				}, nil
			},
			want: []clients.Status{
				{
					Context:   "test",
					State:     "success",
					TargetURL: "https://example.com",
					URL:       "https://example.com",
				},
			},
			wantErr: false,
		},
		{
			name: "main - no statuses",
			ref:  "main",
			getRefs: func(ctx context.Context, args git.GetRefsArgs) (*git.GetRefsResponseValue, error) {
				return &git.GetRefsResponseValue{
					Value: []git.GitRef{
						{
							Statuses: &[]git.GitStatus{},
						},
					},
				}, nil
			},
			want:    []clients.Status{},
			wantErr: false,
		},
		{
			name: "main - single status",
			ref:  "main",
			getRefs: func(ctx context.Context, args git.GetRefsArgs) (*git.GetRefsResponseValue, error) {
				return &git.GetRefsResponseValue{
					Value: []git.GitRef{
						{
							Statuses: &[]git.GitStatus{
								{
									Context: &git.GitStatusContext{
										Name: toPtr("test"),
									},
									State:     toPtr(git.GitStatusStateValues.Succeeded),
									TargetUrl: toPtr("https://example.com"),
								},
							},
						},
					},
				}, nil
			},
			want: []clients.Status{
				{
					Context:   "test",
					State:     "success",
					TargetURL: "https://example.com",
					URL:       "https://example.com",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &commitsHandler{
				ctx: t.Context(),
				repourl: &Repo{
					id: "id",
				},
				getRefs:     tt.getRefs,
				getStatuses: tt.getStatuses,
			}
			got, err := c.listStatuses(tt.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("listStatuses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("listStatuses() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestListCommitsPullRequestDetails(t *testing.T) {
	t.Parallel()
	date := &azuredevops.Time{Time: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	details := &git.GitPullRequest{
		PullRequestId: toPtr(7),
		Status:        toPtr(git.PullRequestStatusValues.Completed),
		CreatedBy:     &webapi.IdentityRef{Id: toPtr(" AUTHOR-ID "), DisplayName: toPtr("Same name")},
		ClosedBy:      &webapi.IdentityRef{Id: toPtr("MERGER-ID"), DisplayName: toPtr("Same name")},
		ClosedDate:    date,
		LastMergeCommit: &git.GitCommitRef{
			CommitId: toPtr("merge-sha"),
		},
		LastMergeSourceCommit: &git.GitCommitRef{
			CommitId: toPtr("source-sha"),
		},
		Reviewers: &[]git.IdentityRefWithVote{
			{Id: toPtr("REVIEWER-10"), Vote: toPtr(10)},
			{Id: toPtr("reviewer-5"), Vote: toPtr(5)},
			{Id: toPtr("waiting"), Vote: toPtr(-5)},
			{Id: toPtr("rejected"), Vote: toPtr(-10)},
			{Id: toPtr("no-vote"), Vote: toPtr(0)},
			{Id: toPtr("missing-vote")},
			{Id: toPtr("unknown-vote"), Vote: toPtr(99)},
			{Id: toPtr("group"), Vote: toPtr(10), IsContainer: toPtr(true)},
			{Vote: toPtr(10)},
		},
	}
	client, requests := newCommitsTestClient(t, details, nil)
	want := clients.PullRequest{
		Number:   7,
		Author:   clients.User{Login: "author-id"},
		MergedBy: clients.User{Login: "merger-id"},
		MergedAt: date.Time,
		HeadSHA:  "source-sha",
		Reviews: []clients.Review{
			{Author: &clients.User{Login: "reviewer-10"}, State: "APPROVED"},
			{Author: &clients.User{Login: "reviewer-5"}, State: "APPROVED"},
			{Author: &clients.User{Login: "waiting"}, State: "COMMENTED"},
			{Author: &clients.User{Login: "rejected"}, State: "CHANGES_REQUESTED"},
		},
	}
	var readers sync.WaitGroup
	for range 3 {
		readers.Go(func() {
			commits, err := client.ListCommits()
			if err != nil {
				t.Error(err)
				return
			}
			if len(commits) != 2 {
				t.Errorf("ListCommits() returned %d commits, want 2", len(commits))
				return
			}
			for _, commit := range commits {
				if diff := cmp.Diff(want, commit.AssociatedMergeRequest); diff != "" {
					t.Errorf("PR details mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
	readers.Wait()
	if got := requests.Load(); got != 1 {
		t.Errorf("PR detail requests = %d, want 1", got)
	}
}

func newCommitsTestClient(t *testing.T, details *git.GitPullRequest, detailErr error) (*Client, *atomic.Int32) {
	t.Helper()
	date := &azuredevops.Time{Time: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	queryPR := git.GitPullRequest{
		PullRequestId:   toPtr(7),
		CreatedBy:       &webapi.IdentityRef{DisplayName: toPtr("Query author")},
		ClosedDate:      date,
		LastMergeCommit: &git.GitCommitRef{CommitId: toPtr("query-merge")},
	}
	requests := new(atomic.Int32)
	handler := &commitsHandler{
		ctx:         t.Context(),
		once:        new(sync.Once),
		commitDepth: 30,
		repourl:     &Repo{id: "repo-id", project: "project", defaultBranch: "main", commitSHA: clients.HeadSHA},
		getCommits: func(context.Context, git.GetCommitsArgs) (*[]git.GitCommitRef, error) {
			return &[]git.GitCommitRef{
				{CommitId: toPtr("first"), Comment: toPtr("first change"), Committer: &git.GitUserDate{Email: toPtr("author@example.invalid"), Date: date}},
				{CommitId: toPtr("second"), Comment: toPtr("second change"), Committer: &git.GitUserDate{Email: toPtr("author@example.invalid"), Date: date}},
			}, nil
		},
		getPullRequestQuery: func(context.Context, git.GetPullRequestQueryArgs) (*git.GitPullRequestQuery, error) {
			return &git.GitPullRequestQuery{Results: &[]map[string][]git.GitPullRequest{{
				"first": {queryPR}, "second": {queryPR},
			}}}, nil
		},
		getPullRequest: func(_ context.Context, args git.GetPullRequestArgs) (*git.GitPullRequest, error) {
			requests.Add(1)
			if args.RepositoryId == nil || *args.RepositoryId != "repo-id" ||
				args.Project == nil || *args.Project != "project" ||
				args.PullRequestId == nil || *args.PullRequestId != 7 {
				t.Errorf("unexpected PR detail request: %+v", args)
			}
			return details, detailErr
		},
	}
	return &Client{commits: handler}, requests
}

func TestListCommitsPullRequestFailures(t *testing.T) {
	t.Parallel()
	requestErr := errors.New("PR access denied")
	for _, tt := range []struct {
		details *git.GitPullRequest
		err     error
		name    string
	}{
		{name: "request denied", err: requestErr},
		{name: "nil response"},
		{name: "missing status", details: &git.GitPullRequest{PullRequestId: toPtr(7)}},
		{name: "missing merge metadata", details: &git.GitPullRequest{
			PullRequestId: toPtr(7), Status: toPtr(git.PullRequestStatusValues.Completed),
		}},
		{name: "missing source commit", details: &git.GitPullRequest{
			PullRequestId: toPtr(7), Status: toPtr(git.PullRequestStatusValues.Completed),
			ClosedDate:      &azuredevops.Time{Time: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
			LastMergeCommit: &git.GitCommitRef{CommitId: toPtr("merge-sha")},
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client, requests := newCommitsTestClient(t, tt.details, tt.err)
			for range 2 {
				commits, err := client.ListCommits()
				if err == nil || commits != nil {
					t.Fatalf("ListCommits() = %v, %v; want PR failure", commits, err)
				}
				if !strings.Contains(err.Error(), "7") {
					t.Errorf("PR context missing from error: %v", err)
				}
				if tt.err != nil && !errors.Is(err, tt.err) {
					t.Errorf("request cause missing from error: %v", err)
				}
			}
			if got := requests.Load(); got != 1 {
				t.Errorf("PR requests after repeated failure = %d, want 1", got)
			}
		})
	}
}
