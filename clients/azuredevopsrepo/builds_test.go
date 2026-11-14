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
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/build"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"

	"github.com/ossf/scorecard/v5/clients"
)

func Test_listSuccessfulBuilds(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                string
		getBuildDefinitions fnListBuildDefinitions
		getBuilds           fnGetBuilds
		want                []clients.WorkflowRun
		wantErr             bool
	}{
		{
			name: "no build definitions",
			getBuildDefinitions: func(ctx context.Context, args build.GetDefinitionsArgs) (*build.GetDefinitionsResponseValue, error) {
				return &build.GetDefinitionsResponseValue{
					Value: []build.BuildDefinitionReference{},
				}, nil
			},
			getBuilds: func(ctx context.Context, args build.GetBuildsArgs) (*build.GetBuildsResponseValue, error) {
				return nil, errors.New("getBuilds should not be called")
			},
			want:    []clients.WorkflowRun{},
			wantErr: false,
		},
		{
			name: "no builds",
			getBuildDefinitions: func(ctx context.Context, args build.GetDefinitionsArgs) (*build.GetDefinitionsResponseValue, error) {
				return &build.GetDefinitionsResponseValue{
					Value: []build.BuildDefinitionReference{
						{Id: toPtr(123)},
					},
				}, nil
			},
			getBuilds: func(ctx context.Context, args build.GetBuildsArgs) (*build.GetBuildsResponseValue, error) {
				return &build.GetBuildsResponseValue{
					Value: []build.Build{},
				}, nil
			},
			want:    []clients.WorkflowRun{},
			wantErr: false,
		},
		{
			name: "single definition and build",
			getBuildDefinitions: func(ctx context.Context, args build.GetDefinitionsArgs) (*build.GetDefinitionsResponseValue, error) {
				return &build.GetDefinitionsResponseValue{
					Value: []build.BuildDefinitionReference{
						{Id: toPtr(123)},
					},
				}, nil
			},
			getBuilds: func(ctx context.Context, args build.GetBuildsArgs) (*build.GetBuildsResponseValue, error) {
				return &build.GetBuildsResponseValue{
					Value: []build.Build{
						{
							Url:           toPtr("https://example.com"),
							SourceVersion: toPtr("abc123"),
						},
					},
				}, nil
			},
			want: []clients.WorkflowRun{
				{
					URL:     "https://example.com",
					HeadSHA: toPtr("abc123"),
				},
			},
			wantErr: false,
		},
		{
			name: "multiple definitions and builds",
			getBuildDefinitions: func(ctx context.Context, args build.GetDefinitionsArgs) (*build.GetDefinitionsResponseValue, error) {
				return &build.GetDefinitionsResponseValue{
					Value: []build.BuildDefinitionReference{
						{Id: toPtr(123)},
						{Id: toPtr(456)},
					},
				}, nil
			},
			getBuilds: func(ctx context.Context, args build.GetBuildsArgs) (*build.GetBuildsResponseValue, error) {
				return &build.GetBuildsResponseValue{
					Value: []build.Build{
						{
							Url:           toPtr("https://example.com"),
							SourceVersion: toPtr("abc123"),
						},
						{
							Url:           toPtr("https://test.com"),
							SourceVersion: toPtr("def456"),
						},
					},
				}, nil
			},
			want: []clients.WorkflowRun{
				{
					URL:     "https://example.com",
					HeadSHA: toPtr("abc123"),
				},
				{
					URL:     "https://test.com",
					HeadSHA: toPtr("def456"),
				},
			},
			wantErr: false,
		},
		{
			name: "multiple definitions and builds with continuation token",
			getBuildDefinitions: func(ctx context.Context, args build.GetDefinitionsArgs) (*build.GetDefinitionsResponseValue, error) {
				if args.ContinuationToken == nil {
					return &build.GetDefinitionsResponseValue{
						Value: []build.BuildDefinitionReference{
							{Id: toPtr(123)},
						},
						ContinuationToken: "abc123",
					}, nil
				}
				return &build.GetDefinitionsResponseValue{
					Value: []build.BuildDefinitionReference{
						{Id: toPtr(789)},
					},
				}, nil
			},
			getBuilds: func(ctx context.Context, args build.GetBuildsArgs) (*build.GetBuildsResponseValue, error) {
				return &build.GetBuildsResponseValue{
					Value: []build.Build{
						{
							Url:           toPtr("https://example.com"),
							SourceVersion: toPtr("abc123"),
						},
					},
				}, nil
			},
			want: []clients.WorkflowRun{
				{
					URL:     "https://example.com",
					HeadSHA: toPtr("abc123"),
				},
			},
			wantErr: false,
		},
		{
			name: "build definitions error",
			getBuildDefinitions: func(ctx context.Context, args build.GetDefinitionsArgs) (*build.GetDefinitionsResponseValue, error) {
				return nil, errors.New("error")
			},
			getBuilds: func(ctx context.Context, args build.GetBuildsArgs) (*build.GetBuildsResponseValue, error) {
				return &build.GetBuildsResponseValue{}, nil
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "builds error",
			getBuildDefinitions: func(ctx context.Context, args build.GetDefinitionsArgs) (*build.GetDefinitionsResponseValue, error) {
				return &build.GetDefinitionsResponseValue{
					Value: []build.BuildDefinitionReference{
						{Id: toPtr(123)},
					},
				}, nil
			},
			getBuilds: func(ctx context.Context, args build.GetBuildsArgs) (*build.GetBuildsResponseValue, error) {
				return nil, errors.New("error")
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			b := &buildsHandler{
				repourl: &Repo{
					project: "test",
					id:      "123",
				},
				getBuildDefinitions: tt.getBuildDefinitions,
				getBuilds:           tt.getBuilds,
			}
			got, err := b.listSuccessfulBuilds("test.yaml")

			if (err != nil) != tt.wantErr {
				t.Errorf("listSuccessfulBuilds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("listSuccessfulBuilds() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

type ciBuildTest struct {
	wantErr           error
	name              string
	branch            string
	version           string
	result            build.BuildResult
	parents           []string
	paged             bool
	want              bool
	missingCollection bool
}

func TestListCheckRunsBuildEvidence(t *testing.T) {
	t.Parallel()
	const (
		source = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		target = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
		tested = "cccccccccccccccccccccccccccccccccccccccc"
	)
	for _, tt := range []ciBuildTest{
		{name: "source build", branch: "refs/heads/feature", version: source, want: true},
		{
			name: "tested merge differs from completed merge", branch: "refs/pull/7/merge", version: tested,
			parents: []string{target, source}, want: true,
		},
		{
			name: "tested merge matches completed merge", branch: "refs/pull/7/merge",
			version: "dddddddddddddddddddddddddddddddddddddddd", want: true,
		},
		{
			name: "tested commit is not a merge", branch: "refs/pull/7/merge", version: tested,
			parents: []string{source},
		},
		{
			name: "stale merge target", branch: "refs/pull/7/merge", version: tested,
			parents: []string{source, "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"},
		},
		{name: "different source revision", branch: "refs/heads/feature", version: tested},
		{name: "failed build", branch: "refs/heads/feature", version: source, result: build.BuildResultValues.Failed},
		{name: "matching build after empty page", branch: "refs/heads/feature", version: source, paged: true, want: true},
		{name: "build request denied", wantErr: errors.New("build access denied")},
		{name: "malformed build page", missingCollection: true, wantErr: errBuildData},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := newCIBuildTestClient(t, &tt)
			got, err := client.ListCheckRunsForRef(source)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ListCheckRunsForRef() error = %v, want %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			want := []clients.CheckRun{}
			if tt.want {
				want = append(want, clients.CheckRun{
					Status: "completed", Conclusion: "success", App: clients.CheckRunApp{Slug: "azure-pipelines"},
					URL: "https://dev.azure.com/org/project/_build/results?buildId=41",
				})
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("build evidence mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func newCIBuildTestClient(t *testing.T, tt *ciBuildTest) *Client {
	t.Helper()
	closed := &azuredevops.Time{Time: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	client, _ := newCommitsTestClient(t, &git.GitPullRequest{
		PullRequestId: toPtr(7), Status: toPtr(git.PullRequestStatusValues.Completed), ClosedDate: closed,
		SourceRefName:         toPtr("refs/heads/feature"),
		LastMergeSourceCommit: &git.GitCommitRef{CommitId: toPtr("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")},
		LastMergeTargetCommit: &git.GitCommitRef{CommitId: toPtr("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")},
		LastMergeCommit:       &git.GitCommitRef{CommitId: toPtr("dddddddddddddddddddddddddddddddddddddddd")},
	}, nil)
	client.azdoClient = &buildCommitTestClient{commit: &git.GitCommit{
		CommitId: toPtr("cccccccccccccccccccccccccccccccccccccccc"), Parents: &tt.parents,
	}}
	client.builds = &buildsHandler{
		ctx: t.Context(), repourl: client.commits.repourl,
		getBuilds: func(_ context.Context, args build.GetBuildsArgs) (*build.GetBuildsResponseValue, error) {
			if tt.missingCollection {
				return &build.GetBuildsResponseValue{}, nil
			}
			if tt.wantErr != nil {
				return nil, tt.wantErr
			}
			if args.BranchName == nil || *args.BranchName != tt.branch {
				return &build.GetBuildsResponseValue{Value: []build.Build{}}, nil
			}
			if tt.paged && args.ContinuationToken == nil {
				return &build.GetBuildsResponseValue{Value: []build.Build{}, ContinuationToken: "page-two"}, nil
			}
			if tt.paged && *args.ContinuationToken != "page-two" {
				t.Errorf("unexpected builds continuation token: %q", *args.ContinuationToken)
			}
			result := tt.result
			if result == "" {
				result = build.BuildResultValues.Succeeded
			}
			return &build.GetBuildsResponseValue{Value: []build.Build{{
				Id: toPtr(41), Repository: &build.BuildRepository{Id: toPtr("repo-id"), Type: toPtr("TfsGit")},
				SourceBranch: &tt.branch, SourceVersion: &tt.version, Reason: toPtr(build.BuildReasonValues.PullRequest),
				Status: toPtr(build.BuildStatusValues.Completed), Result: &result,
				FinishTime: closed, Url: toPtr("https://dev.azure.com/org/project/_apis/build/builds/41"),
				Links: map[string]any{"web": map[string]any{"href": "https://dev.azure.com/org/project/_build/results?buildId=41"}},
			}}}, nil
		},
	}
	return client
}

type buildCommitTestClient struct {
	git.Client
	commit *git.GitCommit
}

func (c *buildCommitTestClient) GetCommit(context.Context, git.GetCommitArgs) (*git.GitCommit, error) {
	return c.commit, nil
}
