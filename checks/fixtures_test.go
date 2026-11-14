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

package checks

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/clients/githubrepo"
	"github.com/ossf/scorecard/v5/clients/localdir"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/internal/packageclient"
	"github.com/ossf/scorecard/v5/log"
	scut "github.com/ossf/scorecard/v5/utests"
)

var errClient = errors.New("client failed")

const fixtureRepoURI = "github.com/ossf-tests/scorecard"

// checkFixture is what the RepoClient, Repo, and CII client return.
type checkFixture struct {
	createdAt       time.Time
	archivedErr     error
	commitsErr      error
	issuesErr       error
	createdErr      error
	runsErr         error
	statusesErr     error
	listFilesErr    error
	readErr         error
	defaultErr      error
	releasesErr     error
	branchErr       error
	contributorsErr error
	workflowRunsErr error
	badgeErr        error
	languagesErr    error
	searchErr       error
	licensesErr     error
	orgErr          error
	immutableErr    error
	defaultNameErr  error
	searchCommitErr error
	localPathErr    error
	versionsErr     error
	vulnsErr        error
	webhooksErr     error
	defaultBranch   *clients.BranchRef
	// versions is what deps.dev returns for the project.
	versions *packageclient.ProjectPackageVersions
	// repoClient, when set, replaces the RepoClient mock. Packaging picks
	// its collector by the client's concrete type.
	repoClient func(t *testing.T) clients.RepoClient
	// org is the organization's repository, for GetOrgRepoClient.
	org *checkFixture
	// ossFuzz is the OSS-Fuzz repository; nil leaves it unset.
	ossFuzz  *checkFixture
	runs     map[string][]clients.CheckRun
	statuses map[string][]clients.Status
	// files maps each path to its content. ListFiles applies the
	// predicate to the paths in sorted order.
	files        map[string]string
	branches     map[string]*clients.BranchRef
	workflowRuns map[string][]clients.WorkflowRun
	// immutable maps "owner/repo@tag" to whether its release is immutable.
	immutable    map[string]bool
	uri          string
	languages    []clients.Language
	licenses     []clients.License
	commits      []clients.Commit
	issues       []clients.Issue
	releases     []clients.Release
	contributors []clients.User
	// dependabotCommits is what SearchCommits returns.
	dependabotCommits []clients.Commit
	vulns             []clients.Vulnerability
	webhooks          []clients.Webhook
	badge             clients.BadgeLevel
	searchHits        int
	archived          bool
	// noCII leaves the request's CII client unset.
	noCII bool
}

// client returns the RepoClient mock.
func (f *checkFixture) client(ctrl *gomock.Controller) *mockrepo.MockRepoClient {
	m := mockrepo.NewMockRepoClient(ctrl)
	m.EXPECT().IsArchived().DoAndReturn(func() (bool, error) {
		return f.archived, f.archivedErr
	}).AnyTimes()
	m.EXPECT().ListCommits().DoAndReturn(func() ([]clients.Commit, error) {
		return f.commits, f.commitsErr
	}).AnyTimes()
	m.EXPECT().ListIssues().DoAndReturn(func() ([]clients.Issue, error) {
		return f.issues, f.issuesErr
	}).AnyTimes()
	m.EXPECT().GetCreatedAt().DoAndReturn(func() (time.Time, error) {
		return f.createdAt, f.createdErr
	}).AnyTimes()
	m.EXPECT().ListCheckRunsForRef(gomock.Any()).DoAndReturn(func(ref string) ([]clients.CheckRun, error) {
		return f.runs[ref], f.runsErr
	}).AnyTimes()
	m.EXPECT().ListStatuses(gomock.Any()).DoAndReturn(func(ref string) ([]clients.Status, error) {
		return f.statuses[ref], f.statusesErr
	}).AnyTimes()
	m.EXPECT().ListFiles(gomock.Any()).DoAndReturn(func(predicate func(string) (bool, error)) ([]string, error) {
		if f.listFilesErr != nil {
			return nil, f.listFilesErr
		}
		var matched []string
		for _, path := range slices.Sorted(maps.Keys(f.files)) {
			ok, err := predicate(path)
			if err != nil {
				return nil, err
			}
			if ok {
				matched = append(matched, path)
			}
		}
		return matched, nil
	}).AnyTimes()
	m.EXPECT().GetFileReader(gomock.Any()).DoAndReturn(func(path string) (io.ReadCloser, error) {
		if f.readErr != nil {
			return nil, f.readErr
		}
		return io.NopCloser(strings.NewReader(f.files[path])), nil
	}).AnyTimes()
	m.EXPECT().GetDefaultBranch().DoAndReturn(func() (*clients.BranchRef, error) {
		return f.defaultBranch, f.defaultErr
	}).AnyTimes()
	m.EXPECT().ListReleases().DoAndReturn(func() ([]clients.Release, error) {
		return f.releases, f.releasesErr
	}).AnyTimes()
	m.EXPECT().GetBranch(gomock.Any()).DoAndReturn(func(branch string) (*clients.BranchRef, error) {
		return f.branches[branch], f.branchErr
	}).AnyTimes()
	m.EXPECT().ListContributors().DoAndReturn(func() ([]clients.User, error) {
		return f.contributors, f.contributorsErr
	}).AnyTimes()
	m.EXPECT().ListSuccessfulWorkflowRuns(gomock.Any()).DoAndReturn(func(file string) ([]clients.WorkflowRun, error) {
		return f.workflowRuns[file], f.workflowRunsErr
	}).AnyTimes()
	m.EXPECT().URI().DoAndReturn(func() string {
		if f.uri == "" {
			return fixtureRepoURI
		}
		return f.uri
	}).AnyTimes()
	m.EXPECT().ListProgrammingLanguages().DoAndReturn(func() ([]clients.Language, error) {
		return f.languages, f.languagesErr
	}).AnyTimes()
	m.EXPECT().Search(gomock.Any()).DoAndReturn(func(req clients.SearchRequest) (clients.SearchResponse, error) {
		return clients.SearchResponse{Hits: f.searchHits}, f.searchErr
	}).AnyTimes()
	m.EXPECT().ListLicenses().DoAndReturn(func() ([]clients.License, error) {
		return f.licenses, f.licensesErr
	}).AnyTimes()
	m.EXPECT().GetOrgRepoClient(gomock.Any()).DoAndReturn(func(context.Context) (clients.RepoClient, error) {
		if f.orgErr != nil {
			return nil, f.orgErr
		}
		return f.org.client(ctrl), nil
	}).AnyTimes()
	m.EXPECT().IsReleaseImmutable(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(owner, repo, tag string) (bool, error) {
			release := owner + "/" + repo + "@" + tag
			return f.immutable[release], f.immutableErr
		}).AnyTimes()
	m.EXPECT().GetDefaultBranchName().DoAndReturn(func() (string, error) {
		return "main", f.defaultNameErr
	}).AnyTimes()
	m.EXPECT().Close().DoAndReturn(func() error {
		return nil
	}).AnyTimes()
	m.EXPECT().SearchCommits(gomock.Any()).DoAndReturn(func(opts clients.SearchCommitsOptions) ([]clients.Commit, error) {
		return f.dependabotCommits, f.searchCommitErr
	}).AnyTimes()
	m.EXPECT().LocalPath().DoAndReturn(func() (string, error) {
		return "/tmp/repo", f.localPathErr
	}).AnyTimes()
	m.EXPECT().ListWebhooks().DoAndReturn(func() ([]clients.Webhook, error) {
		return f.webhooks, f.webhooksErr
	}).AnyTimes()
	return m
}

func (f *checkFixture) repo(ctrl *gomock.Controller) *mockrepo.MockRepo {
	m := mockrepo.NewMockRepo(ctrl)
	m.EXPECT().URI().DoAndReturn(func() string {
		return fixtureRepoURI
	}).AnyTimes()
	m.EXPECT().Host().DoAndReturn(func() string {
		return "github.com"
	}).AnyTimes()
	m.EXPECT().Path().DoAndReturn(func() string {
		return "ossf-tests/scorecard"
	}).AnyTimes()
	return m
}

func (f *checkFixture) project(ctrl *gomock.Controller) packageclient.ProjectPackageClient {
	m := mockrepo.NewMockProjectPackageClient(ctrl)
	m.EXPECT().GetProjectPackageVersions(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, host, project string) (*packageclient.ProjectPackageVersions, error) {
			return f.versions, f.versionsErr
		}).AnyTimes()
	return m
}

func (f *checkFixture) vulnerabilities(ctrl *gomock.Controller) clients.VulnerabilitiesClient {
	m := mockrepo.NewMockVulnerabilitiesClient(ctrl)
	m.EXPECT().ListUnfixedVulnerabilities(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, commit, localPath string) (clients.VulnerabilitiesResponse, error) {
			return clients.VulnerabilitiesResponse{Vulnerabilities: f.vulns}, f.vulnsErr
		}).AnyTimes()
	return m
}

func (f *checkFixture) cii(ctrl *gomock.Controller) clients.CIIBestPracticesClient {
	if f.noCII {
		return nil
	}
	m := mockrepo.NewMockCIIBestPracticesClient(ctrl)
	m.EXPECT().GetBadgeLevel(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, uri string) (clients.BadgeLevel, error) {
			return f.badge, f.badgeErr
		}).AnyTimes()
	return m
}

// The *Fixtures tests run each exported check on fixtures that reach its
// collector's success and error paths, and pin the score or error text.

// runWant is a check's expected result: its score, or the text of its
// runtime error.
type runWant struct {
	err   string
	score int
}

// assertRun runs check on f, with and without a RawResults shared across
// checks, and compares the score and error text with want.
func assertRun(t *testing.T, check checker.CheckFn, f *checkFixture, want runWant) {
	t.Helper()
	for _, sharedRaw := range []bool{false, true} {
		ctrl := gomock.NewController(t)
		req := &checker.CheckRequest{
			Repo:                  f.repo(ctrl),
			CIIClient:             f.cii(ctrl),
			Dlogger:               &scut.TestDetailLogger{},
			ProjectClient:         f.project(ctrl),
			VulnerabilitiesClient: f.vulnerabilities(ctrl),
		}
		if f.repoClient != nil {
			req.RepoClient = f.repoClient(t)
		} else {
			req.RepoClient = f.client(ctrl)
		}
		if f.ossFuzz != nil {
			req.OssFuzzRepo = f.ossFuzz.client(ctrl)
		}
		if sharedRaw {
			// The scorecard runner shares one RawResults across checks.
			req.RawResults = &checker.RawResults{}
		}
		got := check(req)
		gotErr := ""
		if got.Error != nil {
			gotErr = got.Error.Error()
		}
		wantScore := want.score
		if want.err != "" {
			wantScore = checker.InconclusiveResultScore
		}
		if got.Score != wantScore || gotErr != want.err {
			t.Errorf("sharedRaw=%v: score %d, error %q; want score %d, error %q",
				sharedRaw, got.Score, gotErr, wantScore, want.err)
		}
	}
}

func TestMaintainedFixtures(t *testing.T) {
	t.Parallel()
	now := time.Now()
	recent := now.AddDate(0, 0, -5)
	old := now.AddDate(-2, 0, 0)
	member := clients.RepoAssociationMember
	none := clients.RepoAssociationNone
	commits := []clients.Commit{
		{SHA: "a", CommittedDate: now.AddDate(0, 0, -1)},
		{SHA: "b", CommittedDate: now.AddDate(0, 0, -10)},
		{SHA: "c", CommittedDate: now.AddDate(0, -6, 0)},
	}
	issues := []clients.Issue{
		{
			CreatedAt:         &recent,
			AuthorAssociation: &member,
			Author:            &clients.User{Login: "maintainer"},
		},
		{
			CreatedAt:         &recent,
			AuthorAssociation: &none,
			Author:            &clients.User{Login: "someone"},
			Comments: []clients.IssueComment{
				{CreatedAt: &recent, AuthorAssociation: &member, Author: &clients.User{Login: "maintainer"}},
			},
		},
	}
	tests := []struct {
		name    string
		want    runWant
		fixture checkFixture
	}{
		{name: "archived", want: runWant{score: 0}, fixture: checkFixture{archived: true, createdAt: old}},
		{name: "active", want: runWant{score: 3}, fixture: checkFixture{commits: commits, issues: issues, createdAt: old}},
		{name: "created recently", want: runWant{score: 0}, fixture: checkFixture{commits: commits, issues: issues, createdAt: recent}},
		{name: "no activity", want: runWant{score: 0}, fixture: checkFixture{createdAt: old}},
		{name: "IsArchived error", want: runWant{err: "internal error: client failed"}, fixture: checkFixture{archivedErr: errClient}},
		{name: "ListCommits error", want: runWant{err: "internal error: client failed"}, fixture: checkFixture{commitsErr: errClient}},
		{name: "ListIssues error", want: runWant{err: "internal error: client failed"}, fixture: checkFixture{commits: commits, issuesErr: errClient}},
		{name: "GetCreatedAt error", want: runWant{err: "internal error: client failed"}, fixture: checkFixture{commits: commits, issues: issues, createdErr: errClient}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertRun(t, Maintained, &tt.fixture, tt.want)
		})
	}
}

func TestCITestsFixtures(t *testing.T) {
	t.Parallel()
	merged := time.Date(2024, time.May, 6, 7, 8, 9, 0, time.UTC)
	pr := func(sha, head string, n int) clients.Commit {
		return clients.Commit{
			SHA:                    sha,
			AssociatedMergeRequest: clients.PullRequest{Number: n, HeadSHA: head, MergedAt: merged},
		}
	}
	runs := map[string][]clients.CheckRun{
		"head-a": {{Status: "completed", Conclusion: "success", App: clients.CheckRunApp{Slug: "e2e"}}},
		"head-b": {{Status: "completed", Conclusion: "failure", App: clients.CheckRunApp{Slug: "e2e"}}},
	}
	statuses := map[string][]clients.Status{
		"head-a": {{State: "success", Context: "ci/test"}},
		"head-c": {{State: "success", Context: "lint"}},
	}
	tests := []struct {
		name    string
		want    runWant
		fixture checkFixture
	}{
		{name: "no merged PRs", want: runWant{score: -1}, fixture: checkFixture{commits: []clients.Commit{{SHA: "x"}}}},
		{
			name: "one PR with a CI run", want: runWant{score: 10},
			fixture: checkFixture{commits: []clients.Commit{pr("a1", "head-a", 1)}, runs: runs, statuses: statuses},
		},
		{
			name: "squashed, failing, and status-only PRs", want: runWant{score: 3},
			fixture: checkFixture{
				commits: []clients.Commit{
					pr("a1", "head-a", 1),
					pr("b1", "head-b", 2),
					pr("b2", "head-b", 2),
					pr("c1", "head-c", 3),
					{SHA: "d1"},
				},
				runs:     runs,
				statuses: statuses,
			},
		},
		{name: "ListCommits error", want: runWant{err: "internal error: internal error: RepoClient.ListCommits: client failed"}, fixture: checkFixture{commitsErr: errClient}},
		{
			name: "ListCheckRunsForRef error", want: runWant{err: "internal error: internal error: Client.Repositories.ListCheckRunsForRef: client failed"},
			fixture: checkFixture{commits: []clients.Commit{pr("a1", "head-a", 1)}, runsErr: errClient},
		},
		{
			name: "ListStatuses error", want: runWant{err: "internal error: internal error: Client.Repositories.ListStatuses: client failed"},
			fixture: checkFixture{commits: []clients.Commit{pr("a1", "head-a", 1)}, runs: runs, statusesErr: errClient},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertRun(t, CITests, &tt.fixture, tt.want)
		})
	}
}

func fixturePtr[T any](v T) *T {
	return &v
}

const (
	fixtureSafeWorkflow = `on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: make test
`
	fixtureDangerousWorkflow = `on: pull_request_target
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}
      - run: echo "${{ github.event.issue.title }}"
`
	fixtureGradleWorkflow = `on: push
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: gradle/actions/wrapper-validation@v4
`
	// Not valid YAML, so the workflow parser returns no workflow.
	fixtureInvalidWorkflow = `on: push
jobs: {build
`
)

func TestBinaryArtifactsFixtures(t *testing.T) {
	t.Parallel()
	const jar = "gradle/wrapper/gradle-wrapper.jar"
	gradleFiles := map[string]string{
		jar:                            "\x00\x00\x00\x00\x00",
		".github/workflows/gradle.yml": fixtureGradleWorkflow,
	}
	latest := []clients.Commit{{SHA: "head"}}
	runsOnLatest := map[string][]clients.WorkflowRun{"gradle.yml": {{HeadSHA: fixturePtr("head")}}}
	tests := []struct {
		name    string
		want    runWant
		fixture checkFixture
	}{
		{name: "no binaries", want: runWant{score: 10}, fixture: checkFixture{files: map[string]string{"README.md": "hello"}}},
		{
			name: "binary file", want: runWant{score: 9},
			fixture: checkFixture{files: map[string]string{
				"README.md":     "hello",
				"lib/libfoo.so": "\x00\x00\x00\x00\x00",
			}},
		},
		{
			name: "gradle wrapper validated on the latest commit", want: runWant{score: 10},
			fixture: checkFixture{files: gradleFiles, workflowRuns: runsOnLatest, commits: latest},
		},
		{
			name: "gradle wrapper validated on an older commit", want: runWant{score: 9},
			fixture: checkFixture{
				files:        gradleFiles,
				workflowRuns: runsOnLatest,
				commits:      []clients.Commit{{SHA: "newer"}},
			},
		},
		{
			name: "workflow runs unsupported", want: runWant{score: 9},
			fixture: checkFixture{files: gradleFiles, workflowRunsErr: clients.ErrUnsupportedFeature},
		},
		{name: "ListFiles error", want: runWant{err: "internal error: error during ListFiles: client failed"}, fixture: checkFixture{listFilesErr: errClient}},
		{name: "GetFileReader error", want: runWant{err: "internal error: error during GetFileReader: client failed"}, fixture: checkFixture{files: gradleFiles, readErr: errClient}},
		{
			name: "ListSuccessfulWorkflowRuns error", want: runWant{err: "internal error: failure checking for Gradle wrapper validating Action: failure listing workflow runs: client failed"},
			fixture: checkFixture{files: gradleFiles, workflowRunsErr: errClient},
		},
		{
			name: "ListCommits error", want: runWant{err: "internal error: failure checking for Gradle wrapper validating Action: failure listing commits: client failed"},
			fixture: checkFixture{files: gradleFiles, workflowRuns: runsOnLatest, commitsErr: errClient},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertRun(t, BinaryArtifacts, &tt.fixture, tt.want)
		})
	}
}

func TestBranchProtectionFixtures(t *testing.T) {
	t.Parallel()
	protected := &clients.BranchRef{
		Name:      fixturePtr("main"),
		Protected: fixturePtr(true),
		BranchProtectionRule: clients.BranchProtectionRule{
			AllowDeletions:   fixturePtr(false),
			AllowForcePushes: fixturePtr(false),
			EnforceAdmins:    fixturePtr(true),
			PullRequestRule: clients.PullRequestRule{
				Required:                     fixturePtr(true),
				RequiredApprovingReviewCount: fixturePtr(int32(2)),
				RequireCodeOwnerReviews:      fixturePtr(true),
			},
			CheckRules: clients.StatusChecksRule{
				RequiresStatusChecks: fixturePtr(true),
				Contexts:             []string{"ci/test"},
			},
		},
	}
	develop := &clients.BranchRef{Name: fixturePtr("develop"), Protected: fixturePtr(false)}
	release := &clients.BranchRef{Name: fixturePtr("release-1"), Protected: fixturePtr(false)}
	codeowners := map[string]string{"CODEOWNERS": "* @maintainer"}
	tests := []struct {
		name    string
		want    runWant
		fixture checkFixture
	}{
		{
			name: "protected default branch with CODEOWNERS", want: runWant{score: 10},
			fixture: checkFixture{defaultBranch: protected, files: codeowners},
		},
		{
			name: "release branches", want: runWant{score: 1},
			fixture: checkFixture{
				defaultBranch: protected,
				releases: []clients.Release{
					{TagName: "v1", TargetCommitish: "release-1"},
					{TagName: "v2", TargetCommitish: "0123456789abcdef0123456789abcdef01234567"},
					{TagName: "v3", TargetCommitish: "main"},
				},
				branches: map[string]*clients.BranchRef{"release-1": release},
			},
		},
		{
			name: "master release redirects to main", want: runWant{score: 1},
			fixture: checkFixture{
				defaultBranch: develop,
				releases:      []clients.Release{{TagName: "v1", TargetCommitish: "master"}},
				branches:      map[string]*clients.BranchRef{"main": protected},
			},
		},
		{
			name: "releases unsupported", want: runWant{score: 0},
			fixture: checkFixture{defaultBranch: develop, releasesErr: clients.ErrUnsupportedFeature},
		},
		{name: "GetDefaultBranch error", want: runWant{err: "internal error: client failed"}, fixture: checkFixture{defaultErr: errClient}},
		{name: "ListReleases error", want: runWant{err: "internal error: client failed"}, fixture: checkFixture{defaultBranch: protected, releasesErr: errClient}},
		{
			name: "release without a target", want: runWant{err: "internal error: commitish is nil"},
			fixture: checkFixture{
				defaultBranch: protected,
				releases:      []clients.Release{{TagName: "v1"}},
			},
		},
		{
			name: "GetBranch error", want: runWant{err: "internal error: error during GetBranch(release-1): client failed"},
			fixture: checkFixture{
				defaultBranch: protected,
				releases:      []clients.Release{{TagName: "v1", TargetCommitish: "release-1"}},
				branchErr:     errClient,
			},
		},
		{name: "ListFiles error", want: runWant{err: "internal error: error during ListFiles: client failed"}, fixture: checkFixture{defaultBranch: protected, listFilesErr: errClient}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertRun(t, BranchProtection, &tt.fixture, tt.want)
		})
	}
}

func TestCIIBestPracticesFixtures(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		want    runWant
		fixture checkFixture
	}{
		{name: "unknown", want: runWant{err: "internal error: unsupported badge detected"}, fixture: checkFixture{badge: clients.Unknown}},
		{name: "not found", want: runWant{score: 0}, fixture: checkFixture{badge: clients.NotFound}},
		{name: "in progress", want: runWant{score: 2}, fixture: checkFixture{badge: clients.InProgress}},
		{name: "passing", want: runWant{score: 5}, fixture: checkFixture{badge: clients.Passing}},
		{name: "silver", want: runWant{score: 7}, fixture: checkFixture{badge: clients.Silver}},
		{name: "gold", want: runWant{score: 10}, fixture: checkFixture{badge: clients.Gold}},
		{name: "GetBadgeLevel error", want: runWant{err: "internal error: client failed"}, fixture: checkFixture{badgeErr: errClient}},
		{name: "no CII client", want: runWant{err: "internal error: CII client is nil"}, fixture: checkFixture{noCII: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertRun(t, CIIBestPractices, &tt.fixture, tt.want)
		})
	}
}

func TestContributorsFixtures(t *testing.T) {
	t.Parallel()
	contributors := []clients.User{
		{
			Login:            "alice",
			NumContributions: 20,
			IsCodeOwner:      true,
			Organizations:    []clients.User{{Login: "ossf"}, {Login: "ossf"}, {Login: ""}},
			Companies:        []string{"@Example Inc.", "example"},
		},
		{Login: "bob", NumContributions: 15, Organizations: []clients.User{{Login: "cncf"}}},
		{Login: "carol", NumContributions: 12, Companies: []string{"Widgets, LLC", ""}},
		{Login: "dave", NumContributions: 1, Organizations: []clients.User{{Login: "k8s"}}},
	}
	tests := []struct {
		name    string
		want    runWant
		fixture checkFixture
	}{
		{
			name: "contributors from several organizations", want: runWant{score: 10},
			fixture: checkFixture{contributors: contributors},
		},
		{name: "no contributors", want: runWant{score: 0}, fixture: checkFixture{}},
		{name: "ListContributors error", want: runWant{err: "internal error: Client.Repositories.ListContributors: client failed"}, fixture: checkFixture{contributorsErr: errClient}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertRun(t, Contributors, &tt.fixture, tt.want)
		})
	}
}

func TestDangerousWorkflowFixtures(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		want    runWant
		fixture checkFixture
	}{
		{name: "no workflows", want: runWant{score: -1}, fixture: checkFixture{files: map[string]string{"README.md": "hello"}}},
		{
			name: "safe workflow", want: runWant{score: 10},
			fixture: checkFixture{files: map[string]string{".github/workflows/ci.yml": fixtureSafeWorkflow}},
		},
		{
			name: "untrusted checkout and script injection", want: runWant{score: 0},
			fixture: checkFixture{files: map[string]string{
				".github/workflows/ci.yml":     fixtureSafeWorkflow,
				".github/workflows/triage.yml": fixtureDangerousWorkflow,
			}},
		},
		{
			name: "workflow that does not parse", want: runWant{err: "internal error: internal error: invalid GitHub workflow (.github/workflows/bad.yml):\n:3:1: could not parse as YAML: did not find expected ',' or '}' [syntax-check]"},
			fixture: checkFixture{files: map[string]string{".github/workflows/bad.yml": fixtureInvalidWorkflow}},
		},
		{name: "ListFiles error", want: runWant{err: "internal error: error during ListFiles: client failed"}, fixture: checkFixture{listFilesErr: errClient}},
		{
			name: "GetFileReader error", want: runWant{err: "internal error: error during GetFileReader: client failed"},
			fixture: checkFixture{
				files:   map[string]string{".github/workflows/ci.yml": fixtureSafeWorkflow},
				readErr: errClient,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertRun(t, DangerousWorkflow, &tt.fixture, tt.want)
		})
	}
}

func TestFuzzingFixtures(t *testing.T) {
	t.Parallel()
	goFuzz := map[string]string{
		"parse_test.go": "package parse\n\nfunc FuzzParse(f *testing.F) {}\n",
		"README.md":     "hello",
	}
	goLang := []clients.Language{{Name: clients.Go, NumLines: 1000}}
	tests := []struct {
		name    string
		want    runWant
		fixture checkFixture
	}{
		{name: "no fuzzing", want: runWant{score: 0}, fixture: checkFixture{files: map[string]string{"README.md": "hello"}}},
		{
			name: "ClusterFuzzLite", want: runWant{score: 10},
			fixture: checkFixture{files: map[string]string{
				".clusterfuzzlite/Dockerfile": "FROM gcr.io/oss-fuzz-base/base-builder\nRUN make fuzz\n",
			}},
		},
		{name: "OSS-Fuzz project", want: runWant{score: 10}, fixture: checkFixture{ossFuzz: &checkFixture{searchHits: 1}}},
		{name: "not an OSS-Fuzz project", want: runWant{score: 0}, fixture: checkFixture{ossFuzz: &checkFixture{}}},
		{name: "Go fuzz function", want: runWant{score: 10}, fixture: checkFixture{files: goFuzz, languages: goLang}},
		{
			name: "several prominent languages", want: runWant{score: 10},
			fixture: checkFixture{
				files: map[string]string{
					"parse_test.go": goFuzz["parse_test.go"],
					"prop.erl":      "-include_lib(\"proper/include/proper.hrl\").\n",
				},
				languages: []clients.Language{
					{Name: clients.Go, NumLines: 1000},
					{Name: clients.Erlang, NumLines: 800},
					{Name: clients.Python, NumLines: 10},
				},
			},
		},
		{name: "ListFiles error", want: runWant{err: "internal error: error during ListFiles: client failed"}, fixture: checkFixture{listFilesErr: errClient}},
		{name: "Search error", want: runWant{err: "internal error: internal error: Client.Search.Code: client failed"}, fixture: checkFixture{ossFuzz: &checkFixture{searchErr: errClient}}},
		{name: "ListProgrammingLanguages error", want: runWant{err: "internal error: cannot get langs of repo: client failed"}, fixture: checkFixture{languagesErr: errClient}},
		{name: "GetFileReader error", want: runWant{err: "internal error: error when OnMatchingFileContentDo: error during GetFileReader: client failed"}, fixture: checkFixture{files: goFuzz, languages: goLang, readErr: errClient}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertRun(t, Fuzzing, &tt.fixture, tt.want)
		})
	}
}

func TestLicenseFixtures(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		want    runWant
		fixture checkFixture
	}{
		{
			name: "approved license from the API", want: runWant{score: 10},
			fixture: checkFixture{licenses: []clients.License{
				{Path: "LICENSE", Name: "Apache License 2.0", SPDXId: "Apache-2.0"},
			}},
		},
		{
			name: "unapproved license from the API", want: runWant{score: 9},
			fixture: checkFixture{licenses: []clients.License{
				{Path: "LICENSE", Name: "Custom", SPDXId: "NOASSERTION"},
			}},
		},
		{
			name: "license file, API unsupported", want: runWant{score: 10},
			fixture: checkFixture{
				licensesErr: clients.ErrUnsupportedFeature,
				files:       map[string]string{"LICENSE-MIT.txt": "MIT License", "README.md": "hello"},
			},
		},
		{
			name: "API finds nothing, license file found", want: runWant{score: 9},
			fixture: checkFixture{files: map[string]string{"COPYING": "GNU GENERAL PUBLIC LICENSE"}},
		},
		{name: "no license", want: runWant{score: 0}, fixture: checkFixture{files: map[string]string{"README.md": "hello"}}},
		{name: "ListLicenses error", want: runWant{err: "internal error: RepoClient.ListLicenses: client failed"}, fixture: checkFixture{licensesErr: errClient}},
		{
			name: "ListFiles error", want: runWant{err: "internal error: fileparser.OnAllFilesDo: error during ListFiles: client failed"},
			fixture: checkFixture{licensesErr: clients.ErrUnsupportedFeature, listFilesErr: errClient},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertRun(t, License, &tt.fixture, tt.want)
		})
	}
}

const (
	// Container images are pinned by digest: an unpinned image makes the
	// remediation query a registry for its digest.
	fixtureDockerfile = `FROM golang:1.22@sha256:0000000000000000000000000000000000000000000000000000000000000000
RUN curl -sSL https://example.com/install.sh | bash
RUN pip install requests
`
	fixturePinningWorkflow = `on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@8e5e7e5ab8b370d6c329ec480221332ada57f0ab
      - uses: actions/setup-go@v5
      - uses: example/immutable-action@v1
      - run: curl -sSL https://example.com/install.sh | sh
      - run: npm install left-pad
      - run: nuget install Example.Package
`
)

func TestPinningDependenciesFixtures(t *testing.T) {
	t.Parallel()
	files := map[string]string{
		".github/workflows/ci.yml": fixturePinningWorkflow,
		"Dockerfile":               fixtureDockerfile,
		"scripts/install.sh":       "#!/bin/bash\nwget https://example.com/tool.sh -O - | sh\n",
	}
	immutable := map[string]bool{"example/immutable-action@v1": true}
	tests := []struct {
		name    string
		want    runWant
		fixture checkFixture
	}{
		{name: "no dependencies", want: runWant{score: -1}, fixture: checkFixture{files: map[string]string{"README.md": "hello"}}},
		{name: "pinned and unpinned dependencies", want: runWant{score: 2}, fixture: checkFixture{files: files, immutable: immutable}},
		{
			name: "NuGet with central package management", want: runWant{score: 2},
			fixture: checkFixture{files: map[string]string{
				".github/workflows/ci.yml": fixturePinningWorkflow,
				"Directory.Packages.props": `<Project>
  <PropertyGroup>
    <ManagePackageVersionsCentrally>true</ManagePackageVersionsCentrally>
  </PropertyGroup>
  <ItemGroup>
    <PackageVersion Include="Example.Package" Version="[1.2.3]" />
  </ItemGroup>
</Project>
`,
			}},
		},
		{name: "IsReleaseImmutable error", want: runWant{score: 1}, fixture: checkFixture{files: files, immutableErr: errClient}},
		{
			name: "workflow that does not parse", want: runWant{err: "internal error: internal error: invalid GitHub workflow (.github/workflows/bad.yml):\n:3:1: could not parse as YAML: did not find expected ',' or '}' [syntax-check]"},
			fixture: checkFixture{files: map[string]string{".github/workflows/bad.yml": fixtureInvalidWorkflow}},
		},
		{name: "ListFiles error", want: runWant{err: "internal error: error during ListFiles: client failed"}, fixture: checkFixture{listFilesErr: errClient}},
		{name: "GetFileReader error", want: runWant{err: "internal error: error during GetFileReader: client failed"}, fixture: checkFixture{files: files, readErr: errClient}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertRun(t, PinningDependencies, &tt.fixture, tt.want)
		})
	}
}

func TestSASTFixtures(t *testing.T) {
	t.Parallel()
	merged := time.Date(2024, time.May, 6, 7, 8, 9, 0, time.UTC)
	pr := func(sha, head string, n int) clients.Commit {
		return clients.Commit{
			SHA:                    sha,
			AssociatedMergeRequest: clients.PullRequest{Number: n, HeadSHA: head, MergedAt: merged},
		}
	}
	commits := []clients.Commit{pr("a1", "head-a", 1), pr("b1", "head-b", 2), {SHA: "c1"}}
	runs := map[string][]clients.CheckRun{
		"head-a": {{
			Status:     "completed",
			Conclusion: "success",
			URL:        "https://example.com/run/1",
			App:        clients.CheckRunApp{Slug: "github-code-scanning"},
		}},
		"head-b": {
			{Status: "in_progress", App: clients.CheckRunApp{Slug: "github-code-scanning"}},
			{Status: "completed", Conclusion: "failure", App: clients.CheckRunApp{Slug: "sonarcloud"}},
		},
	}
	workflows := map[string]string{
		".github/workflows/codeql.yml": `on: push
jobs:
  analyze:
    runs-on: ubuntu-latest
    steps:
      - uses: github/codeql-action/analyze@v3
`,
		".github/workflows/snyk.yml": `on: push
jobs:
  snyk:
    runs-on: ubuntu-latest
    steps:
      - uses: snyk/actions/golang@master
`,
		"pom.xml": `<project>
  <properties>
    <sonar.host.url>https://sonarcloud.io</sonar.host.url>
  </properties>
</project>
`,
	}
	tests := []struct {
		name    string
		want    runWant
		fixture checkFixture
	}{
		{name: "no SAST", want: runWant{score: 0}, fixture: checkFixture{commits: []clients.Commit{pr("a1", "head-a", 1)}}},
		{name: "SAST check runs", want: runWant{score: 5}, fixture: checkFixture{commits: commits, runs: runs}},
		{name: "SAST workflows and Sonar", want: runWant{score: 10}, fixture: checkFixture{files: workflows}},
		{
			name: "commits unsupported", want: runWant{score: 10},
			fixture: checkFixture{commitsErr: clients.ErrUnsupportedFeature, files: workflows},
		},
		{name: "ListCommits error", want: runWant{err: "internal error: internal error: RepoClient.ListCommits: client failed"}, fixture: checkFixture{commitsErr: errClient}},
		{name: "ListCheckRunsForRef error", want: runWant{err: "internal error: internal error: Client.Checks.ListCheckRunsForRef: client failed"}, fixture: checkFixture{commits: commits, runsErr: errClient}},
		{name: "ListFiles error", want: runWant{err: "internal error: error during ListFiles: client failed"}, fixture: checkFixture{commits: commits, runs: runs, listFilesErr: errClient}},
		{name: "GetFileReader error", want: runWant{err: "internal error: error during GetFileReader: client failed"}, fixture: checkFixture{files: workflows, readErr: errClient}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertRun(t, SAST, &tt.fixture, tt.want)
		})
	}
}

func TestSecurityPolicyFixtures(t *testing.T) {
	t.Parallel()
	const policy = `# Security Policy

Report problems to security@example.com or through
https://example.com/security. We respond within 7 days and disclose
fixes after a 90-day embargo.
`
	org := func(files map[string]string) *checkFixture {
		return &checkFixture{uri: "github.com/ossf-tests/.github", files: files}
	}
	noFiles := map[string]string{"README.md": "hello"}
	tests := []struct {
		name    string
		want    runWant
		fixture checkFixture
	}{
		{name: "policy in the repository", want: runWant{score: 10}, fixture: checkFixture{files: map[string]string{"SECURITY.md": policy}}},
		{
			name: "policy in the organization", want: runWant{score: 10},
			fixture: checkFixture{files: noFiles, org: org(map[string]string{"SECURITY.md": policy})},
		},
		{name: "no policy", want: runWant{score: 0}, fixture: checkFixture{files: noFiles, org: org(noFiles)}},
		{name: "organization unreachable", want: runWant{score: 0}, fixture: checkFixture{files: noFiles, orgErr: sce.ErrRepoUnreachable}},
		{
			name: "organization unsupported", want: runWant{score: 0},
			fixture: checkFixture{files: noFiles, orgErr: clients.ErrUnsupportedFeature},
		},
		{name: "GetOrgRepoClient error", want: runWant{err: "internal error: client failed"}, fixture: checkFixture{files: noFiles, orgErr: errClient}},
		{name: "ListFiles error", want: runWant{err: "internal error: error during ListFiles: client failed"}, fixture: checkFixture{listFilesErr: errClient}},
		{
			name: "organization ListFiles error", want: runWant{err: "internal error: unable to create github client: error during ListFiles: client failed"},
			fixture: checkFixture{files: noFiles, org: &checkFixture{listFilesErr: errClient}},
		},
		{
			name: "GetFileReader error", want: runWant{err: "internal error: error during GetFileReader: client failed"},
			fixture: checkFixture{files: map[string]string{"SECURITY.md": policy}, readErr: errClient},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertRun(t, SecurityPolicy, &tt.fixture, tt.want)
		})
	}
}

func TestCodeReviewFixtures(t *testing.T) {
	t.Parallel()
	merged := time.Date(2024, time.May, 6, 7, 8, 9, 0, time.UTC)
	alice := clients.User{Login: "alice"}
	bob := clients.User{Login: "bob"}
	pr := func(sha string, n int, reviews ...clients.Review) clients.Commit {
		return clients.Commit{
			SHA:       sha,
			Committer: bob,
			AssociatedMergeRequest: clients.PullRequest{
				Number:   n,
				MergedAt: merged,
				Author:   alice,
				MergedBy: alice,
				Reviews:  reviews,
			},
		}
	}
	approved := clients.Review{State: "APPROVED", Author: &bob}
	commits := []clients.Commit{
		pr("a1", 1, approved),
		pr("a2", 1, approved),
		pr("b1", 2),
		{SHA: "c1", Committer: alice, Message: "Fix\n\nDifferential Revision: https://example.com/D123"},
		{SHA: "d1", Committer: alice, Message: "Fix\n\nReviewed-on: https://example.com/c/1\nReviewed-by: bob"},
		{SHA: "e1", Committer: alice, Message: "Fix\n\nPiperOrigin-RevId: 123456"},
		{SHA: "f1", Committer: alice, Message: "Push to the default branch"},
	}
	prow := pr("g1", 3)
	prow.AssociatedMergeRequest.Labels = []clients.Label{{Name: "lgtm"}}
	tests := []struct {
		name    string
		want    runWant
		fixture checkFixture
	}{
		{name: "no commits", want: runWant{score: -1}},
		{name: "one approved PR", want: runWant{score: 10}, fixture: checkFixture{commits: []clients.Commit{pr("a1", 1, approved)}}},
		{name: "prow", want: runWant{score: 10}, fixture: checkFixture{commits: []clients.Commit{prow}}},
		{name: "every review platform", want: runWant{score: 6}, fixture: checkFixture{commits: commits}},
		{name: "ListCommits error", want: runWant{err: "internal error: client failed"}, fixture: checkFixture{commitsErr: errClient}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertRun(t, CodeReview, &tt.fixture, tt.want)
		})
	}
}

func TestDependencyUpdateToolFixtures(t *testing.T) {
	t.Parallel()
	dependabot := []clients.Commit{
		{SHA: "a1", Committer: clients.User{ID: 1}},
		{SHA: "b1", Committer: clients.User{ID: 49699333}},
	}
	noTools := map[string]string{"README.md": "hello"}
	tests := []struct {
		name    string
		want    runWant
		fixture checkFixture
	}{
		{name: "Dependabot config", want: runWant{score: 10}, fixture: checkFixture{files: map[string]string{".github/dependabot.yml": ""}}},
		{
			name: "Renovate and Scala Steward configs", want: runWant{score: 10},
			fixture: checkFixture{files: map[string]string{
				"renovate.json":       "{}",
				".scala-steward.conf": "",
			}},
		},
		{name: "Dependabot commits", want: runWant{score: 10}, fixture: checkFixture{files: noTools, dependabotCommits: dependabot}},
		{name: "no tool", want: runWant{score: 0}, fixture: checkFixture{files: noTools, dependabotCommits: dependabot[:1]}},
		{
			name: "commit search unsupported", want: runWant{score: 0},
			fixture: checkFixture{files: noTools, searchCommitErr: clients.ErrUnsupportedFeature},
		},
		{name: "ListFiles error", want: runWant{err: "internal error: error during ListFiles: client failed"}, fixture: checkFixture{listFilesErr: errClient}},
		{name: "SearchCommits error", want: runWant{err: "internal error: dependabot commit search: client failed"}, fixture: checkFixture{files: noTools, searchCommitErr: errClient}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertRun(t, DependencyUpdateTool, &tt.fixture, tt.want)
		})
	}
}

// fixtureVersions parses a deps.dev response, since ProjectPackageVersions
// is built from anonymous structs.
func fixtureVersions(t *testing.T, body string) *packageclient.ProjectPackageVersions {
	t.Helper()
	var v packageclient.ProjectPackageVersions
	if err := json.Unmarshal([]byte(body), &v); err != nil {
		t.Fatal(err)
	}
	return &v
}

func TestSignedReleasesFixtures(t *testing.T) {
	t.Parallel()
	asset := func(name string) clients.ReleaseAsset {
		return clients.ReleaseAsset{Name: name, URL: "https://example.com/" + name}
	}
	releases := []clients.Release{
		{TagName: "v3", URL: "https://example.com/v3", Assets: []clients.ReleaseAsset{asset("app"), asset("app.sig")}},
		{TagName: "v2", URL: "https://example.com/v2", Assets: []clients.ReleaseAsset{asset("app.intoto.jsonl")}},
		{TagName: "v1", URL: "https://example.com/v1", Assets: []clients.ReleaseAsset{asset("app")}, IsImmutable: true},
	}
	versions := fixtureVersions(t, `{"versions": [
		{"versionKey": {"system": "NPM", "name": "app", "version": "3.0.0"},
		 "slsaProvenances": [{"commit": "abc", "verified": true}]},
		{"versionKey": {"system": "NPM", "name": "app", "version": "2.0.0"}}
	]}`)
	tests := []struct {
		name    string
		want    runWant
		fixture checkFixture
	}{
		{name: "no releases", want: runWant{score: -1}, fixture: checkFixture{versions: &packageclient.ProjectPackageVersions{}}},
		{name: "signed, attested, and immutable releases", want: runWant{score: 9}, fixture: checkFixture{releases: releases, versions: versions}},
		{name: "package provenance only", want: runWant{score: -1}, fixture: checkFixture{versions: versions}},
		{name: "GetProjectPackageVersions error", want: runWant{score: 9}, fixture: checkFixture{releases: releases, versionsErr: errClient}},
		{name: "ListReleases error", want: runWant{err: "internal error: client failed"}, fixture: checkFixture{releasesErr: errClient}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertRun(t, SignedReleases, &tt.fixture, tt.want)
		})
	}
}

func TestTokenPermissionsFixtures(t *testing.T) {
	t.Parallel()
	const readAll = `on: push
permissions: read-all
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: make
`
	const jobWrite = `on: push
permissions:
  contents: read
jobs:
  release:
    runs-on: ubuntu-latest
    permissions:
      contents: write
      packages: write
    steps:
      - run: make release
`
	const undeclared = `on: pull_request
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: make test
`
	const writeAll = `on: push
permissions: write-all
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - run: make deploy
`
	tests := []struct {
		name    string
		want    runWant
		fixture checkFixture
	}{
		{name: "no workflows", want: runWant{score: -1}, fixture: checkFixture{files: map[string]string{"README.md": "hello"}}},
		{name: "read-all", want: runWant{score: 10}, fixture: checkFixture{files: map[string]string{".github/workflows/build.yml": readAll}}},
		{
			name: "job writes, undeclared, and write-all", want: runWant{score: 0},
			fixture: checkFixture{files: map[string]string{
				".github/workflows/build.yml":   readAll,
				".github/workflows/release.yml": jobWrite,
				".github/workflows/test.yml":    undeclared,
				".github/workflows/deploy.yml":  writeAll,
			}},
		},
		{
			name: "unparsable workflow", want: runWant{err: "internal error: internal error: invalid GitHub workflow (.github/workflows/bad.yml):\n:2:5: could not parse as YAML: did not find expected ',' or ']' [syntax-check]"},
			fixture: checkFixture{files: map[string]string{".github/workflows/bad.yml": "on: [push\njobs: {"}},
		},
		{name: "ListFiles error", want: runWant{err: "internal error: error during ListFiles: client failed"}, fixture: checkFixture{listFilesErr: errClient}},
		{
			name: "GetFileReader error", want: runWant{err: "internal error: error during GetFileReader: client failed"},
			fixture: checkFixture{
				files:   map[string]string{".github/workflows/build.yml": readAll},
				readErr: errClient,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertRun(t, TokenPermissions, &tt.fixture, tt.want)
		})
	}
}

func TestVulnerabilitiesFixtures(t *testing.T) {
	t.Parallel()
	commits := []clients.Commit{{SHA: "a1"}, {SHA: "b1"}}
	vulns := []clients.Vulnerability{
		{ID: "GHSA-aaaa-bbbb-cccc", Aliases: []string{"CVE-2024-0001"}},
		{ID: "GO-2024-0002"},
	}
	tests := []struct {
		name    string
		want    runWant
		fixture checkFixture
	}{
		{name: "no vulnerabilities", want: runWant{score: 10}, fixture: checkFixture{commits: commits}},
		{name: "vulnerabilities", want: runWant{score: 8}, fixture: checkFixture{commits: commits, vulns: vulns}},
		{name: "commits without SHAs", want: runWant{score: 8}, fixture: checkFixture{commits: []clients.Commit{{}}, vulns: vulns}},
		{name: "ListCommits error", want: runWant{score: 8}, fixture: checkFixture{commitsErr: errClient, vulns: vulns}},
		{name: "LocalPath error", want: runWant{err: "internal error: RepoClient.LocalPath: client failed"}, fixture: checkFixture{commits: commits, localPathErr: errClient}},
		{name: "ListUnfixedVulnerabilities error", want: runWant{err: "internal error: vulnerabilitiesClient.ListUnfixedVulnerabilities: client failed"}, fixture: checkFixture{commits: commits, vulnsErr: errClient}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertRun(t, Vulnerabilities, &tt.fixture, tt.want)
		})
	}
}

// unsetExperimental unsets SCORECARD_EXPERIMENTAL. Call it after
// t.Setenv, which restores the original value when the test ends.
func unsetExperimental(t *testing.T) {
	t.Helper()
	if err := os.Unsetenv("SCORECARD_EXPERIMENTAL"); err != nil {
		t.Fatalf("Unsetenv: %v", err)
	}
}

func TestSBOMFixtures(t *testing.T) {
	release := func(assets ...string) clients.Release {
		r := clients.Release{TagName: "v1"}
		for _, name := range assets {
			r.Assets = append(r.Assets, clients.ReleaseAsset{Name: name, URL: "https://example.com/" + name})
		}
		return r
	}
	sources := map[string]string{"sbom.spdx.json": "", "docs/app.cdx.json": "", "README.md": ""}
	tests := []struct {
		name     string
		want     runWant
		fixture  checkFixture
		disabled bool
	}{
		{name: "SCORECARD_EXPERIMENTAL unset", want: runWant{err: "check is not supported for this request: SCORECARD_EXPERIMENTAL is not set, not running the SBOM check"}, fixture: checkFixture{files: sources}, disabled: true},
		{name: "no SBOM", want: runWant{score: 0}},
		{
			name: "SBOM in releases and in source", want: runWant{score: 10},
			fixture: checkFixture{
				releases: []clients.Release{release("app.tar.gz", "app.spdx.json"), release("app.cdx.json")},
				files:    sources,
			},
		},
		{
			name: "ListReleases unsupported", want: runWant{score: 5},
			fixture: checkFixture{releasesErr: clients.ErrUnsupportedFeature, files: sources},
		},
		{name: "ListReleases error", want: runWant{err: "internal error: RepoClient.ListReleases: client failed"}, fixture: checkFixture{releasesErr: errClient}},
		{name: "ListFiles error", want: runWant{err: "internal error: error during ListFiles: client failed"}, fixture: checkFixture{listFilesErr: errClient}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SCORECARD_EXPERIMENTAL", "true")
			if tt.disabled {
				unsetExperimental(t)
			}
			assertRun(t, SBOM, &tt.fixture, tt.want)
		})
	}
}

func TestWebhooksFixtures(t *testing.T) {
	webhooks := []clients.Webhook{
		{ID: 1, Path: "https://example.com/hook", UsesAuthSecret: true},
		{ID: 2, Path: "https://example.com/open"},
	}
	tests := []struct {
		name     string
		want     runWant
		fixture  checkFixture
		disabled bool
	}{
		{name: "SCORECARD_EXPERIMENTAL unset", want: runWant{err: "check is not supported for this request: SCORECARD_EXPERIMENTAL is not set, not running the Webhook check"}, fixture: checkFixture{webhooks: webhooks}, disabled: true},
		{name: "no webhooks", want: runWant{score: 10}},
		{name: "webhooks with and without a secret", want: runWant{score: 5}, fixture: checkFixture{webhooks: webhooks}},
		{name: "ListWebhooks error", want: runWant{err: "internal error: internal error: Client.Repositories.ListWebhooks"}, fixture: checkFixture{webhooksErr: errClient}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SCORECARD_EXPERIMENTAL", "true")
			if tt.disabled {
				unsetExperimental(t)
			}
			assertRun(t, WebHooks, &tt.fixture, tt.want)
		})
	}
}

const (
	fixturePublishWorkflow = `on: push
jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/setup-node@v4
        with:
          registry-url: https://registry.npmjs.org
      - run: npm publish
`
	fixtureGitlabPublishJob = `publish:
  script:
    - twine upload dist/*
`
	fixtureGitlabFile   = "gitlabscorecard_flattened_ci.yaml"
	fixtureWorkflowPath = ".github/workflows/publish.yml"
	fixtureRunsPath     = "/repos/ossf-tests/scorecard/actions/workflows/publish.yml/runs"
	fixtureTarballPath  = "/repos/ossf-tests/scorecard/tarball/"
)

// localDirClient returns a local directory client over files. The
// unreadable paths are listed but cannot be opened, except by root.
func localDirClient(files map[string]string, unreadable ...string) func(*testing.T) clients.RepoClient {
	return func(t *testing.T) clients.RepoClient {
		t.Helper()
		dir := t.TempDir()
		for path, content := range files {
			full := filepath.Join(dir, path)
			if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
				t.Fatalf("MkdirAll: %v", err)
			}
			if err := os.WriteFile(full, []byte(content), 0o600); err != nil {
				t.Fatalf("WriteFile: %v", err)
			}
		}
		for _, path := range unreadable {
			if err := os.Chmod(filepath.Join(dir, path), 0); err != nil {
				t.Fatalf("Chmod: %v", err)
			}
		}
		repo, err := localdir.MakeLocalDirRepo(dir)
		if err != nil {
			t.Fatalf("MakeLocalDirRepo: %v", err)
		}
		client := localdir.CreateLocalDirClient(context.Background(), log.NewLogger(log.DefaultLevel))
		if err := client.InitRepo(repo, clients.HeadSHA, 0); err != nil {
			t.Fatalf("InitRepo: %v", err)
		}
		return client
	}
}

type fakeGitHubResponse struct {
	body   []byte
	status int
}

// fakeGitHub serves canned GitHub API responses by path.
type fakeGitHub struct {
	responses map[string]fakeGitHubResponse
}

func (f *fakeGitHub) RoundTrip(req *http.Request) (*http.Response, error) {
	r, ok := f.responses[req.URL.Path]
	if !ok {
		r = fakeGitHubResponse{status: http.StatusNotFound, body: []byte(`{}`)}
	}
	return &http.Response{
		StatusCode: r.status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader(r.body)),
		Request:    req,
	}, nil
}

// fixtureTarball returns a gzipped tarball of files, laid out like GitHub's.
func fixtureTarball(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	const root = "ossf-tests-scorecard-0000000/"
	// The tarball handler creates each directory from its own entry.
	dirs := map[string]bool{root: true}
	for path := range files {
		for dir := filepath.Dir(path); dir != "."; dir = filepath.Dir(dir) {
			dirs[root+dir+"/"] = true
		}
	}
	for _, dir := range slices.Sorted(maps.Keys(dirs)) {
		if err := tw.WriteHeader(&tar.Header{Name: dir, Typeflag: tar.TypeDir, Mode: 0o755}); err != nil {
			t.Fatalf("tar.WriteHeader: %v", err)
		}
	}
	for _, path := range slices.Sorted(maps.Keys(files)) {
		hdr := &tar.Header{Name: root + path, Typeflag: tar.TypeReg, Mode: 0o644, Size: int64(len(files[path]))}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("tar.WriteHeader: %v", err)
		}
		if _, err := tw.Write([]byte(files[path])); err != nil {
			t.Fatalf("tar.Write: %v", err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("tar.Close: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("gzip.Close: %v", err)
	}
	return buf.Bytes()
}

// gitHubClient returns a GitHub client initialized at HEAD over a fake
// API serving files and the workflow runs response.
func gitHubClient(files map[string]string, runs fakeGitHubResponse) func(*testing.T) clients.RepoClient {
	return func(t *testing.T) clients.RepoClient {
		t.Helper()
		fake := &fakeGitHub{responses: map[string]fakeGitHubResponse{
			"/repos/ossf-tests/scorecard": {status: http.StatusOK, body: []byte(`{
				"name": "scorecard",
				"owner": {"login": "ossf-tests"},
				"default_branch": "main",
				"archive_url": "https://api.github.com/repos/ossf-tests/scorecard/{archive_format}{/ref}"
			}`)},
			fixtureTarballPath: {status: http.StatusOK, body: fixtureTarball(t, files)},
			fixtureRunsPath:    runs,
		}}
		client := githubrepo.CreateGithubRepoClientWithTransport(context.Background(), fake)
		repo, err := githubrepo.MakeGithubRepo(fixtureRepoURI)
		if err != nil {
			t.Fatalf("MakeGithubRepo: %v", err)
		}
		if err := client.InitRepo(repo, clients.HeadSHA, 0); err != nil {
			t.Fatalf("InitRepo: %v", err)
		}
		t.Cleanup(func() {
			if err := client.Close(); err != nil {
				t.Errorf("Close: %v", err)
			}
		})
		return client
	}
}

func TestPackagingFixtures(t *testing.T) {
	t.Parallel()
	publishing := map[string]string{fixtureWorkflowPath: fixturePublishWorkflow}
	runs := fakeGitHubResponse{status: http.StatusOK, body: []byte(`{"total_count": 1, "workflow_runs": [` +
		`{"url": "https://api.github.com/repos/ossf-tests/scorecard/actions/runs/1"}]}`)}
	tests := []struct {
		name    string
		want    runWant
		fixture checkFixture
	}{
		{name: "unsupported client", want: runWant{score: -1}, fixture: checkFixture{files: publishing}},
		{name: "local directory without workflows", want: runWant{score: -1}, fixture: checkFixture{repoClient: localDirClient(nil)}},
		{name: "local directory, GitHub workflow", want: runWant{score: 10}, fixture: checkFixture{repoClient: localDirClient(publishing)}},
		{
			name: "local directory, GitLab job", want: runWant{score: 10},
			fixture: checkFixture{repoClient: localDirClient(map[string]string{
				fixtureGitlabFile: fixtureGitlabPublishJob,
			})},
		},
		{
			name: "local directory, GitHub workflow and GitLab job", want: runWant{score: 10},
			fixture: checkFixture{repoClient: localDirClient(map[string]string{
				fixtureWorkflowPath: fixturePublishWorkflow,
				fixtureGitlabFile:   fixtureGitlabPublishJob,
			})},
		},
		{
			name: "local directory, invalid GitHub workflow", want: runWant{err: "internal error: internal error: invalid GitHub workflow (.github/workflows/publish.yml):\n:3:1: could not parse as YAML: did not find expected ',' or '}' [syntax-check]"},
			fixture: checkFixture{repoClient: localDirClient(map[string]string{
				fixtureWorkflowPath: fixtureInvalidWorkflow,
				fixtureGitlabFile:   fixtureGitlabPublishJob,
			})},
		},
		{
			name: "local directory, unreadable GitLab file", want: runWant{err: "internal error: RepoClient.GetFileReader: open file: openat gitlabscorecard_flattened_ci.yaml: permission denied"},
			fixture: checkFixture{repoClient: localDirClient(map[string]string{
				fixtureWorkflowPath: fixturePublishWorkflow,
				fixtureGitlabFile:   fixtureGitlabPublishJob,
			}, fixtureGitlabFile)},
		},
		{
			name: "local directory, both collectors fail", want: runWant{err: "internal error: internal error: invalid GitHub workflow (.github/workflows/publish.yml):\n:3:1: could not parse as YAML: did not find expected ',' or '}' [syntax-check]"},
			fixture: checkFixture{repoClient: localDirClient(map[string]string{
				fixtureWorkflowPath: fixtureInvalidWorkflow,
				fixtureGitlabFile:   fixtureGitlabPublishJob,
			}, fixtureGitlabFile)},
		},
		{name: "GitHub, publishing workflow with a run", want: runWant{score: 10}, fixture: checkFixture{repoClient: gitHubClient(publishing, runs)}},
		{
			name: "GitHub, workflow that does not publish", want: runWant{score: -1},
			fixture: checkFixture{repoClient: gitHubClient(map[string]string{
				fixtureWorkflowPath: fixtureSafeWorkflow,
			}, runs)},
		},
		{
			name: "GitHub, workflow runs request fails", want: runWant{err: "internal error: Client.Actions.ListWorkflowRunsByFileName: internal error: ListWorkflowRunsByFileName: GET https://api.github.com/repos/ossf-tests/scorecard/actions/workflows/publish.yml/runs?status=success: 500  []"},
			fixture: checkFixture{repoClient: gitHubClient(publishing, fakeGitHubResponse{
				status: http.StatusInternalServerError, body: []byte(`{}`),
			})},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertRun(t, Packaging, &tt.fixture, tt.want)
		})
	}
}
