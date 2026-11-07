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
	"context"
	"errors"
	"io"
	"maps"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/checks/evaluation"
	"github.com/ossf/scorecard/v5/checks/raw"
	"github.com/ossf/scorecard/v5/clients"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes"
	"github.com/ossf/scorecard/v5/probes/zrunner"
	scut "github.com/ossf/scorecard/v5/utests"
)

// The legacy* functions are the check bodies as they were before the
// checks moved to definitions. The differential tests below run each
// legacy body and its definition on identical inputs and require
// identical results, details, raw data, and RepoClient call sequences.

func legacyMaintained(c *checker.CheckRequest) checker.CheckResult {
	rawData, err := raw.Maintained(c)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckMaintained, e)
	}

	// Set the raw results.
	pRawResults := getRawResults(c)
	pRawResults.MaintainedResults = rawData

	// Evaluate the probes.
	findings, err := zrunner.Run(pRawResults, probes.Maintained)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckMaintained, e)
	}

	// Return the score evaluation.
	ret := evaluation.Maintained(CheckMaintained, findings, c.Dlogger)
	ret.Findings = findings
	return ret
}

func legacyCITests(c *checker.CheckRequest) checker.CheckResult {
	rawData, err := raw.CITests(c.RepoClient)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckCITests, e)
	}

	pRawResults := getRawResults(c)
	pRawResults.CITestResults = rawData

	// Evaluate the probes.
	findings, err := zrunner.Run(pRawResults, probes.CITests)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckCITests, e)
	}

	ret := evaluation.CITests(CheckCITests, findings, c.Dlogger)
	ret.Findings = findings
	return ret
}

func legacyBinaryArtifacts(c *checker.CheckRequest) checker.CheckResult {
	rawData, err := raw.BinaryArtifacts(c)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckBinaryArtifacts, e)
	}

	// Set the raw results.
	pRawResults := getRawResults(c)
	pRawResults.BinaryArtifactResults = rawData

	// Evaluate the probes.
	findings, err := zrunner.Run(pRawResults, probes.BinaryArtifacts)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckBinaryArtifacts, e)
	}

	ret := evaluation.BinaryArtifacts(CheckBinaryArtifacts, findings, c.Dlogger)
	ret.Findings = findings
	return ret
}

func legacyBranchProtection(c *checker.CheckRequest) checker.CheckResult {
	rawData, err := raw.BranchProtection(c)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckBranchProtection, e)
	}

	// Set the raw results.
	pRawResults := getRawResults(c)
	pRawResults.BranchProtectionResults = rawData

	// Evaluate the probes.
	findings, err := zrunner.Run(pRawResults, probes.BranchProtection)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckBranchProtection, e)
	}

	// Return the score evaluation.
	ret := evaluation.BranchProtection(CheckBranchProtection, findings, c.Dlogger)
	ret.Findings = findings
	return ret
}

func legacyCIIBestPractices(c *checker.CheckRequest) checker.CheckResult {
	rawData, err := raw.CIIBestPractices(c)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckCIIBestPractices, e)
	}

	// Set the raw results.
	pRawResults := getRawResults(c)
	pRawResults.CIIBestPracticesResults = rawData

	// Evaluate the probes.
	findings, err := zrunner.Run(pRawResults, probes.CIIBestPractices)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckCIIBestPractices, e)
	}

	// Return the score evaluation.
	ret := evaluation.CIIBestPractices(CheckCIIBestPractices, findings, c.Dlogger)
	ret.Findings = findings
	return ret
}

func legacyContributors(c *checker.CheckRequest) checker.CheckResult {
	rawData, err := raw.Contributors(c)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckContributors, e)
	}

	// Set the raw results.
	pRawResults := getRawResults(c)
	pRawResults.ContributorsResults = rawData

	// Evaluate the probes.
	findings, err := zrunner.Run(pRawResults, probes.Contributors)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckContributors, e)
	}

	// Return the score evaluation.
	ret := evaluation.Contributors(CheckContributors, findings, c.Dlogger)
	ret.Findings = findings
	return ret
}

func legacyDangerousWorkflow(c *checker.CheckRequest) checker.CheckResult {
	rawData, err := raw.DangerousWorkflow(c)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckDangerousWorkflow, e)
	}

	// Set the raw results.
	pRawResults := getRawResults(c)
	pRawResults.DangerousWorkflowResults = rawData

	// Evaluate the probes.
	findings, err := zrunner.Run(pRawResults, probes.DangerousWorkflows)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckDangerousWorkflow, e)
	}

	ret := evaluation.DangerousWorkflow(CheckDangerousWorkflow, findings, c.Dlogger)
	ret.Findings = findings
	return ret
}

var errDiff = errors.New("client failed")

const diffRepoURI = "github.com/ossf-tests/scorecard"

// diffFixture is what the RepoClient, Repo, and CII client return.
// Every call is recorded.
type diffFixture struct {
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
	defaultBranch   *clients.BranchRef
	runs            map[string][]clients.CheckRun
	statuses        map[string][]clients.Status
	// files maps each path to its content. ListFiles applies the
	// predicate to the paths in sorted order.
	files        map[string]string
	branches     map[string]*clients.BranchRef
	workflowRuns map[string][]clients.WorkflowRun
	commits      []clients.Commit
	issues       []clients.Issue
	releases     []clients.Release
	contributors []clients.User
	badge        clients.BadgeLevel
	archived     bool
	// noCII leaves the request's CII client unset.
	noCII bool
}

func (f *diffFixture) client(ctrl *gomock.Controller, calls *[]string) *mockrepo.MockRepoClient {
	m := mockrepo.NewMockRepoClient(ctrl)
	m.EXPECT().IsArchived().DoAndReturn(func() (bool, error) {
		*calls = append(*calls, "IsArchived")
		return f.archived, f.archivedErr
	}).AnyTimes()
	m.EXPECT().ListCommits().DoAndReturn(func() ([]clients.Commit, error) {
		*calls = append(*calls, "ListCommits")
		return f.commits, f.commitsErr
	}).AnyTimes()
	m.EXPECT().ListIssues().DoAndReturn(func() ([]clients.Issue, error) {
		*calls = append(*calls, "ListIssues")
		return f.issues, f.issuesErr
	}).AnyTimes()
	m.EXPECT().GetCreatedAt().DoAndReturn(func() (time.Time, error) {
		*calls = append(*calls, "GetCreatedAt")
		return f.createdAt, f.createdErr
	}).AnyTimes()
	m.EXPECT().ListCheckRunsForRef(gomock.Any()).DoAndReturn(func(ref string) ([]clients.CheckRun, error) {
		*calls = append(*calls, "ListCheckRunsForRef "+ref)
		return f.runs[ref], f.runsErr
	}).AnyTimes()
	m.EXPECT().ListStatuses(gomock.Any()).DoAndReturn(func(ref string) ([]clients.Status, error) {
		*calls = append(*calls, "ListStatuses "+ref)
		return f.statuses[ref], f.statusesErr
	}).AnyTimes()
	m.EXPECT().ListFiles(gomock.Any()).DoAndReturn(func(predicate func(string) (bool, error)) ([]string, error) {
		*calls = append(*calls, "ListFiles")
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
		*calls = append(*calls, "GetFileReader "+path)
		if f.readErr != nil {
			return nil, f.readErr
		}
		return io.NopCloser(strings.NewReader(f.files[path])), nil
	}).AnyTimes()
	m.EXPECT().GetDefaultBranch().DoAndReturn(func() (*clients.BranchRef, error) {
		*calls = append(*calls, "GetDefaultBranch")
		return f.defaultBranch, f.defaultErr
	}).AnyTimes()
	m.EXPECT().ListReleases().DoAndReturn(func() ([]clients.Release, error) {
		*calls = append(*calls, "ListReleases")
		return f.releases, f.releasesErr
	}).AnyTimes()
	m.EXPECT().GetBranch(gomock.Any()).DoAndReturn(func(branch string) (*clients.BranchRef, error) {
		*calls = append(*calls, "GetBranch "+branch)
		return f.branches[branch], f.branchErr
	}).AnyTimes()
	m.EXPECT().ListContributors().DoAndReturn(func() ([]clients.User, error) {
		*calls = append(*calls, "ListContributors")
		return f.contributors, f.contributorsErr
	}).AnyTimes()
	m.EXPECT().ListSuccessfulWorkflowRuns(gomock.Any()).DoAndReturn(func(file string) ([]clients.WorkflowRun, error) {
		*calls = append(*calls, "ListSuccessfulWorkflowRuns "+file)
		return f.workflowRuns[file], f.workflowRunsErr
	}).AnyTimes()
	return m
}

func (f *diffFixture) repo(ctrl *gomock.Controller, calls *[]string) *mockrepo.MockRepo {
	m := mockrepo.NewMockRepo(ctrl)
	m.EXPECT().URI().DoAndReturn(func() string {
		*calls = append(*calls, "URI")
		return diffRepoURI
	}).AnyTimes()
	return m
}

func (f *diffFixture) cii(ctrl *gomock.Controller, calls *[]string) clients.CIIBestPracticesClient {
	if f.noCII {
		return nil
	}
	m := mockrepo.NewMockCIIBestPracticesClient(ctrl)
	m.EXPECT().GetBadgeLevel(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, uri string) (clients.BadgeLevel, error) {
			*calls = append(*calls, "GetBadgeLevel "+uri)
			return f.badge, f.badgeErr
		}).AnyTimes()
	return m
}

type runOutput struct {
	raw     *checker.RawResults
	result  checker.CheckResult
	details []checker.CheckDetail
	calls   []string
}

func runWith(
	t *testing.T,
	check checker.CheckFn,
	f *diffFixture,
	sharedRaw bool,
) runOutput {
	t.Helper()
	ctrl := gomock.NewController(t)
	var out runOutput
	dl := scut.TestDetailLogger{}
	req := &checker.CheckRequest{
		RepoClient: f.client(ctrl, &out.calls),
		Repo:       f.repo(ctrl, &out.calls),
		CIIClient:  f.cii(ctrl, &out.calls),
		Dlogger:    &dl,
	}
	if sharedRaw {
		// The scorecard runner shares one RawResults across checks.
		req.RawResults = &checker.RawResults{}
	}
	out.result = check(req)
	out.details = dl.Flush()
	out.raw = req.RawResults
	return out
}

// errorsByText compares errors by message and sentinel, since
// sce.WithMessage returns a new value on every call.
var errorsByText = cmp.Comparer(func(a, b error) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.Error() == b.Error() &&
		errors.Is(a, sce.ErrScorecardInternal) == errors.Is(b, sce.ErrScorecardInternal)
})

func assertSameRun(t *testing.T, legacy, current checker.CheckFn, f *diffFixture, opts ...cmp.Option) {
	t.Helper()
	opts = append(opts, errorsByText, cmpopts.IgnoreUnexported(finding.Finding{}))
	for _, sharedRaw := range []bool{false, true} {
		want := runWith(t, legacy, f, sharedRaw)
		got := runWith(t, current, f, sharedRaw)
		if diff := cmp.Diff(want, got, append(opts, cmp.AllowUnexported(runOutput{}))...); diff != "" {
			t.Errorf("sharedRaw=%v: definition differs from legacy body (-legacy +definition):\n%s", sharedRaw, diff)
		}
	}
}

func TestMaintainedDefinitionMatchesLegacy(t *testing.T) {
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
		fixture diffFixture
	}{
		{name: "archived", fixture: diffFixture{archived: true, createdAt: old}},
		{name: "active", fixture: diffFixture{commits: commits, issues: issues, createdAt: old}},
		{name: "created recently", fixture: diffFixture{commits: commits, issues: issues, createdAt: recent}},
		{name: "no activity", fixture: diffFixture{createdAt: old}},
		{name: "IsArchived error", fixture: diffFixture{archivedErr: errDiff}},
		{name: "ListCommits error", fixture: diffFixture{commitsErr: errDiff}},
		{name: "ListIssues error", fixture: diffFixture{commits: commits, issuesErr: errDiff}},
		{name: "GetCreatedAt error", fixture: diffFixture{commits: commits, issues: issues, createdErr: errDiff}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertSameRun(t, legacyMaintained, Maintained, &tt.fixture)
		})
	}
}

func TestCITestsDefinitionMatchesLegacy(t *testing.T) {
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
	// raw.CITests collates through a map, so with more than one PR the
	// order of CIInfo, findings, and details is not deterministic.
	unordered := []cmp.Option{
		cmpopts.SortSlices(func(a, b checker.RevisionCIInfo) bool { return a.HeadSHA < b.HeadSHA }),
		cmpopts.SortSlices(func(a, b finding.Finding) bool { return a.Message < b.Message }),
		cmpopts.SortSlices(func(a, b checker.CheckDetail) bool { return a.Msg.Text < b.Msg.Text }),
	}
	tests := []struct {
		opts    []cmp.Option
		name    string
		fixture diffFixture
	}{
		{name: "no merged PRs", fixture: diffFixture{commits: []clients.Commit{{SHA: "x"}}}},
		{
			name:    "one PR with a CI run",
			fixture: diffFixture{commits: []clients.Commit{pr("a1", "head-a", 1)}, runs: runs, statuses: statuses},
		},
		{
			name: "squashed, failing, and status-only PRs",
			fixture: diffFixture{
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
			opts: unordered,
		},
		{name: "ListCommits error", fixture: diffFixture{commitsErr: errDiff}},
		{
			name:    "ListCheckRunsForRef error",
			fixture: diffFixture{commits: []clients.Commit{pr("a1", "head-a", 1)}, runsErr: errDiff},
		},
		{
			name:    "ListStatuses error",
			fixture: diffFixture{commits: []clients.Commit{pr("a1", "head-a", 1)}, runs: runs, statusesErr: errDiff},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertSameRun(t, legacyCITests, CITests, &tt.fixture, tt.opts...)
		})
	}
}

func diffPtr[T any](v T) *T {
	return &v
}

const (
	diffSafeWorkflow = `on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: make test
`
	diffDangerousWorkflow = `on: pull_request_target
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}
      - run: echo "${{ github.event.issue.title }}"
`
	diffGradleWorkflow = `on: push
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: gradle/actions/wrapper-validation@v4
`
	// Not valid YAML, so the workflow parser returns no workflow.
	diffInvalidWorkflow = `on: push
jobs: {build
`
)

func TestBinaryArtifactsDefinitionMatchesLegacy(t *testing.T) {
	t.Parallel()
	const jar = "gradle/wrapper/gradle-wrapper.jar"
	gradleFiles := map[string]string{
		jar:                            "\x00\x00\x00\x00\x00",
		".github/workflows/gradle.yml": diffGradleWorkflow,
	}
	latest := []clients.Commit{{SHA: "head"}}
	runsOnLatest := map[string][]clients.WorkflowRun{"gradle.yml": {{HeadSHA: diffPtr("head")}}}
	tests := []struct {
		name    string
		fixture diffFixture
	}{
		{name: "no binaries", fixture: diffFixture{files: map[string]string{"README.md": "hello"}}},
		{
			name: "binary file",
			fixture: diffFixture{files: map[string]string{
				"README.md":     "hello",
				"lib/libfoo.so": "\x00\x00\x00\x00\x00",
			}},
		},
		{
			name:    "gradle wrapper validated on the latest commit",
			fixture: diffFixture{files: gradleFiles, workflowRuns: runsOnLatest, commits: latest},
		},
		{
			name: "gradle wrapper validated on an older commit",
			fixture: diffFixture{
				files:        gradleFiles,
				workflowRuns: runsOnLatest,
				commits:      []clients.Commit{{SHA: "newer"}},
			},
		},
		{
			name:    "workflow runs unsupported",
			fixture: diffFixture{files: gradleFiles, workflowRunsErr: clients.ErrUnsupportedFeature},
		},
		{name: "ListFiles error", fixture: diffFixture{listFilesErr: errDiff}},
		{name: "GetFileReader error", fixture: diffFixture{files: gradleFiles, readErr: errDiff}},
		{
			name:    "ListSuccessfulWorkflowRuns error",
			fixture: diffFixture{files: gradleFiles, workflowRunsErr: errDiff},
		},
		{
			name:    "ListCommits error",
			fixture: diffFixture{files: gradleFiles, workflowRuns: runsOnLatest, commitsErr: errDiff},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertSameRun(t, legacyBinaryArtifacts, BinaryArtifacts, &tt.fixture)
		})
	}
}

func TestBranchProtectionDefinitionMatchesLegacy(t *testing.T) {
	t.Parallel()
	protected := &clients.BranchRef{
		Name:      diffPtr("main"),
		Protected: diffPtr(true),
		BranchProtectionRule: clients.BranchProtectionRule{
			AllowDeletions:   diffPtr(false),
			AllowForcePushes: diffPtr(false),
			EnforceAdmins:    diffPtr(true),
			PullRequestRule: clients.PullRequestRule{
				Required:                     diffPtr(true),
				RequiredApprovingReviewCount: diffPtr(int32(2)),
				RequireCodeOwnerReviews:      diffPtr(true),
			},
			CheckRules: clients.StatusChecksRule{
				RequiresStatusChecks: diffPtr(true),
				Contexts:             []string{"ci/test"},
			},
		},
	}
	develop := &clients.BranchRef{Name: diffPtr("develop"), Protected: diffPtr(false)}
	release := &clients.BranchRef{Name: diffPtr("release-1"), Protected: diffPtr(false)}
	codeowners := map[string]string{"CODEOWNERS": "* @maintainer"}
	tests := []struct {
		name    string
		fixture diffFixture
	}{
		{
			name:    "protected default branch with CODEOWNERS",
			fixture: diffFixture{defaultBranch: protected, files: codeowners},
		},
		{
			name: "release branches",
			fixture: diffFixture{
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
			name: "master release redirects to main",
			fixture: diffFixture{
				defaultBranch: develop,
				releases:      []clients.Release{{TagName: "v1", TargetCommitish: "master"}},
				branches:      map[string]*clients.BranchRef{"main": protected},
			},
		},
		{
			name:    "releases unsupported",
			fixture: diffFixture{defaultBranch: develop, releasesErr: clients.ErrUnsupportedFeature},
		},
		{name: "GetDefaultBranch error", fixture: diffFixture{defaultErr: errDiff}},
		{name: "ListReleases error", fixture: diffFixture{defaultBranch: protected, releasesErr: errDiff}},
		{
			name: "release without a target",
			fixture: diffFixture{
				defaultBranch: protected,
				releases:      []clients.Release{{TagName: "v1"}},
			},
		},
		{
			name: "GetBranch error",
			fixture: diffFixture{
				defaultBranch: protected,
				releases:      []clients.Release{{TagName: "v1", TargetCommitish: "release-1"}},
				branchErr:     errDiff,
			},
		},
		{name: "ListFiles error", fixture: diffFixture{defaultBranch: protected, listFilesErr: errDiff}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertSameRun(t, legacyBranchProtection, BranchProtection, &tt.fixture)
		})
	}
}

func TestCIIBestPracticesDefinitionMatchesLegacy(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		fixture diffFixture
	}{
		{name: "unknown", fixture: diffFixture{badge: clients.Unknown}},
		{name: "not found", fixture: diffFixture{badge: clients.NotFound}},
		{name: "in progress", fixture: diffFixture{badge: clients.InProgress}},
		{name: "passing", fixture: diffFixture{badge: clients.Passing}},
		{name: "silver", fixture: diffFixture{badge: clients.Silver}},
		{name: "gold", fixture: diffFixture{badge: clients.Gold}},
		{name: "GetBadgeLevel error", fixture: diffFixture{badgeErr: errDiff}},
		{name: "no CII client", fixture: diffFixture{noCII: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertSameRun(t, legacyCIIBestPractices, CIIBestPractices, &tt.fixture)
		})
	}
}

func TestContributorsDefinitionMatchesLegacy(t *testing.T) {
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
	// The probe collects entities through a map, so with more than one
	// entity the order of findings is not deterministic.
	unordered := cmpopts.SortSlices(func(a, b finding.Finding) bool { return a.Message < b.Message })
	tests := []struct {
		opts    []cmp.Option
		name    string
		fixture diffFixture
	}{
		{
			name:    "contributors from several organizations",
			fixture: diffFixture{contributors: contributors},
			opts:    []cmp.Option{unordered},
		},
		{name: "no contributors", fixture: diffFixture{}},
		{name: "ListContributors error", fixture: diffFixture{contributorsErr: errDiff}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertSameRun(t, legacyContributors, Contributors, &tt.fixture, tt.opts...)
		})
	}
}

func TestDangerousWorkflowDefinitionMatchesLegacy(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		fixture diffFixture
	}{
		{name: "no workflows", fixture: diffFixture{files: map[string]string{"README.md": "hello"}}},
		{
			name:    "safe workflow",
			fixture: diffFixture{files: map[string]string{".github/workflows/ci.yml": diffSafeWorkflow}},
		},
		{
			name: "untrusted checkout and script injection",
			fixture: diffFixture{files: map[string]string{
				".github/workflows/ci.yml":     diffSafeWorkflow,
				".github/workflows/triage.yml": diffDangerousWorkflow,
			}},
		},
		{
			name:    "workflow that does not parse",
			fixture: diffFixture{files: map[string]string{".github/workflows/bad.yml": diffInvalidWorkflow}},
		},
		{name: "ListFiles error", fixture: diffFixture{listFilesErr: errDiff}},
		{
			name: "GetFileReader error",
			fixture: diffFixture{
				files:   map[string]string{".github/workflows/ci.yml": diffSafeWorkflow},
				readErr: errDiff,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertSameRun(t, legacyDangerousWorkflow, DangerousWorkflow, &tt.fixture)
		})
	}
}
