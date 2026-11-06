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

package scorecard

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/probes/archived"
	"github.com/ossf/scorecard/v5/probes/createdRecently"
	"github.com/ossf/scorecard/v5/probes/hasRecentCommits"
	"github.com/ossf/scorecard/v5/probes/issueActivityByProjectMember"
	"github.com/ossf/scorecard/v5/probes/testsRunInCI"
)

// These tests characterize probe-mode raw data collection
// (populateRawResults and assignRawData) so that moving the
// check-to-collector wiring cannot change which collectors run,
// how often they call the RepoClient, what they store, or how
// errors surface.

var errCollect = errors.New("collect failed")

var maintainedProbes = []string{
	archived.Probe,
	hasRecentCommits.Probe,
	issueActivityByProjectMember.Probe,
	createdRecently.Probe,
}

type maintainedFixture struct {
	createdAt time.Time
	commits   []clients.Commit
	issues    []clients.Issue
	archived  bool
}

func newMaintainedFixture() maintainedFixture {
	created := time.Date(2020, time.January, 2, 3, 4, 5, 0, time.UTC)
	return maintainedFixture{
		createdAt: created,
		commits: []clients.Commit{
			{SHA: "c1", CommittedDate: created.AddDate(1, 0, 0)},
			{SHA: "c2", CommittedDate: created.AddDate(2, 0, 0)},
		},
		issues: []clients.Issue{
			{URI: asPointer("https://example.com/issues/1"), CreatedAt: &created},
		},
	}
}

func (f *maintainedFixture) expect(m *mockrepo.MockRepoClient) {
	m.EXPECT().IsArchived().Return(f.archived, nil).Times(1)
	m.EXPECT().ListCommits().Return(f.commits, nil).Times(1)
	m.EXPECT().ListIssues().Return(f.issues, nil).Times(1)
	m.EXPECT().GetCreatedAt().Return(f.createdAt, nil).Times(1)
}

func (f *maintainedFixture) want() checker.MaintainedData {
	return checker.MaintainedData{
		CreatedAt:            f.createdAt,
		Issues:               f.issues,
		DefaultBranchCommits: f.commits,
		ArchivedStatus:       checker.ArchivedStatus{Status: f.archived},
	}
}

func mergedCommit(sha, headSHA string, pr int) clients.Commit {
	return clients.Commit{
		SHA: sha,
		AssociatedMergeRequest: clients.PullRequest{
			Number:   pr,
			HeadSHA:  headSHA,
			MergedAt: time.Date(2024, time.May, 6, 7, 8, 9, 0, time.UTC),
		},
	}
}

// sortCIInfo makes CITestData comparable: raw.CITests builds CIInfo by
// iterating a map, so its order is not deterministic.
var sortCIInfo = cmpopts.SortSlices(func(a, b checker.RevisionCIInfo) bool {
	return a.HeadSHA < b.HeadSHA
})

func TestPopulateRawResults_Maintained(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
	fixture := newMaintainedFixture()
	// All four Maintained probes need the same raw data, so the
	// collector must run exactly once (Times(1) in expect).
	fixture.expect(mockRepoClient)

	req := &checker.CheckRequest{RepoClient: mockRepoClient}
	var ret Result
	if err := populateRawResults(req, maintainedProbes, &ret); err != nil {
		t.Fatalf("populateRawResults: %v", err)
	}

	want := checker.RawResults{MaintainedResults: fixture.want()}
	if diff := cmp.Diff(want, ret.RawResults); diff != "" {
		t.Errorf("RawResults mismatch (-want +got):\n%s", diff)
	}
}

func TestPopulateRawResults_CITests(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	mockRepoClient := mockrepo.NewMockRepoClient(ctrl)

	commits := []clients.Commit{
		mergedCommit("c1", "head-a", 1),
		// Squash merges map several commits to one head SHA: check runs
		// are fetched once per head, statuses once per commit.
		mergedCommit("c2", "head-b", 2),
		mergedCommit("c3", "head-b", 2),
		{SHA: "c4"}, // not merged: skipped
	}
	runsA := []clients.CheckRun{{Status: "completed", Conclusion: "success", App: clients.CheckRunApp{Slug: "a"}}}
	runsB := []clients.CheckRun{{Status: "completed", Conclusion: "failure", App: clients.CheckRunApp{Slug: "b"}}}
	statusA := []clients.Status{{State: "success", Context: "ci/a"}}
	statusB := []clients.Status{{State: "pending", Context: "ci/b"}}

	mockRepoClient.EXPECT().ListCommits().Return(commits, nil).Times(1)
	mockRepoClient.EXPECT().ListCheckRunsForRef("head-a").Return(runsA, nil).Times(1)
	mockRepoClient.EXPECT().ListCheckRunsForRef("head-b").Return(runsB, nil).Times(1)
	mockRepoClient.EXPECT().ListStatuses("head-a").Return(statusA, nil).Times(1)
	mockRepoClient.EXPECT().ListStatuses("head-b").Return(statusB, nil).Times(2)

	req := &checker.CheckRequest{RepoClient: mockRepoClient}
	var ret Result
	if err := populateRawResults(req, []string{testsRunInCI.Probe}, &ret); err != nil {
		t.Fatalf("populateRawResults: %v", err)
	}

	want := checker.RawResults{
		CITestResults: checker.CITestData{
			CIInfo: []checker.RevisionCIInfo{
				{HeadSHA: "head-a", CheckRuns: runsA, Statuses: statusA, PullRequestNumber: 1},
				{HeadSHA: "head-b", CheckRuns: runsB, Statuses: append(statusB, statusB...), PullRequestNumber: 2},
			},
		},
	}
	if diff := cmp.Diff(want, ret.RawResults, sortCIInfo); diff != "" {
		t.Errorf("RawResults mismatch (-want +got):\n%s", diff)
	}
}

func TestPopulateRawResults_CollectsEachCheckOnce(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
	fixture := newMaintainedFixture()

	// Deduplication is per check, not per RepoClient method:
	// Maintained and CI-Tests each call ListCommits once.
	gomock.InOrder(
		mockRepoClient.EXPECT().IsArchived().Return(false, nil),
		mockRepoClient.EXPECT().ListCommits().Return(fixture.commits, nil),
		mockRepoClient.EXPECT().ListIssues().Return(fixture.issues, nil),
		mockRepoClient.EXPECT().GetCreatedAt().Return(fixture.createdAt, nil),
		mockRepoClient.EXPECT().ListCommits().Return(fixture.commits, nil),
	)

	probesToRun := []string{
		archived.Probe,
		testsRunInCI.Probe,
		hasRecentCommits.Probe,
		testsRunInCI.Probe,
		createdRecently.Probe,
	}
	req := &checker.CheckRequest{RepoClient: mockRepoClient}
	var ret Result
	if err := populateRawResults(req, probesToRun, &ret); err != nil {
		t.Fatalf("populateRawResults: %v", err)
	}

	want := checker.RawResults{
		MaintainedResults: fixture.want(),
		// The fixture commits have no merged PRs.
		CITestResults: checker.CITestData{CIInfo: []checker.RevisionCIInfo{}},
	}
	if diff := cmp.Diff(want, ret.RawResults); diff != "" {
		t.Errorf("RawResults mismatch (-want +got):\n%s", diff)
	}
}

func TestPopulateRawResults_Errors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		wantSCEType error
		expect      func(m *mockrepo.MockRepoClient)
		name        string
		wantErr     string
		probes      []string
	}{
		{
			name:        "unknown probe",
			probes:      []string{"nonExistentProbe"},
			expect:      func(m *mockrepo.MockRepoClient) {},
			wantErr:     `getting probe "nonExistentProbe": internal error: probe "nonExistentProbe" not found`,
			wantSCEType: sce.ErrScorecardInternal,
		},
		{
			name:   "Maintained collector error stops collection",
			probes: append([]string{archived.Probe}, testsRunInCI.Probe),
			expect: func(m *mockrepo.MockRepoClient) {
				m.EXPECT().IsArchived().Return(false, nil)
				m.EXPECT().ListCommits().Return(nil, errCollect)
				// No ListIssues, GetCreatedAt, or CI-Tests calls follow.
			},
			wantErr:     "internal error: collect failed",
			wantSCEType: sce.ErrScorecardInternal,
		},
		{
			name:   "CI-Tests collector error is wrapped twice",
			probes: []string{testsRunInCI.Probe},
			expect: func(m *mockrepo.MockRepoClient) {
				m.EXPECT().ListCommits().Return(nil, errCollect)
			},
			wantErr:     "internal error: internal error: RepoClient.ListCommits: collect failed",
			wantSCEType: sce.ErrScorecardInternal,
		},
		{
			name:   "CI-Tests per-ref error",
			probes: []string{testsRunInCI.Probe},
			expect: func(m *mockrepo.MockRepoClient) {
				m.EXPECT().ListCommits().Return([]clients.Commit{mergedCommit("c1", "head-a", 1)}, nil)
				m.EXPECT().ListCheckRunsForRef("head-a").Return(nil, errCollect)
			},
			wantErr:     "internal error: internal error: Client.Repositories.ListCheckRunsForRef: collect failed",
			wantSCEType: sce.ErrScorecardInternal,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
			tt.expect(mockRepoClient)

			req := &checker.CheckRequest{RepoClient: mockRepoClient}
			var ret Result
			err := populateRawResults(req, tt.probes, &ret)
			if err == nil {
				t.Fatalf("populateRawResults: want error %q, got nil", tt.wantErr)
			}
			if err.Error() != tt.wantErr {
				t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
			}
			if tt.wantSCEType != nil && !errors.Is(err, tt.wantSCEType) {
				t.Errorf("errors.Is(%v, %v) = false", err, tt.wantSCEType)
			}
		})
	}
}
