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
	"errors"
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

var errDiff = errors.New("client failed")

// diffFixture is what the RepoClient returns. Every call is recorded.
type diffFixture struct {
	createdAt   time.Time
	archivedErr error
	commitsErr  error
	issuesErr   error
	createdErr  error
	runsErr     error
	statusesErr error
	runs        map[string][]clients.CheckRun
	statuses    map[string][]clients.Status
	commits     []clients.Commit
	issues      []clients.Issue
	archived    bool
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
