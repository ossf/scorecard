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
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/checks"
	"github.com/ossf/scorecard/v5/clients"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/checknames"
	"github.com/ossf/scorecard/v5/internal/fuzzers"
	"github.com/ossf/scorecard/v5/internal/packageclient"
	proberegistration "github.com/ossf/scorecard/v5/internal/probes"
	"github.com/ossf/scorecard/v5/probes/archived"
	"github.com/ossf/scorecard/v5/probes/branchesAreProtected"
	"github.com/ossf/scorecard/v5/probes/codeApproved"
	"github.com/ossf/scorecard/v5/probes/contributorsFromOrgOrCompany"
	"github.com/ossf/scorecard/v5/probes/createdRecently"
	"github.com/ossf/scorecard/v5/probes/dependencyUpdateToolConfigured"
	"github.com/ossf/scorecard/v5/probes/fuzzed"
	"github.com/ossf/scorecard/v5/probes/hasBinaryArtifacts"
	"github.com/ossf/scorecard/v5/probes/hasDangerousWorkflowUntrustedCheckout"
	"github.com/ossf/scorecard/v5/probes/hasLicenseFile"
	"github.com/ossf/scorecard/v5/probes/hasOSVVulnerabilities"
	"github.com/ossf/scorecard/v5/probes/hasOpenSSFBadge"
	"github.com/ossf/scorecard/v5/probes/hasRecentCommits"
	"github.com/ossf/scorecard/v5/probes/hasSBOM"
	"github.com/ossf/scorecard/v5/probes/issueActivityByProjectMember"
	"github.com/ossf/scorecard/v5/probes/pinsDependencies"
	"github.com/ossf/scorecard/v5/probes/releasesAreSigned"
	"github.com/ossf/scorecard/v5/probes/sastToolConfigured"
	"github.com/ossf/scorecard/v5/probes/securityPolicyPresent"
	"github.com/ossf/scorecard/v5/probes/testsRunInCI"
	"github.com/ossf/scorecard/v5/probes/webhooksUseSecrets"
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

// testTokenPermissionsProbe requires Token-Permissions raw data. No
// registered probe does, so probe mode reaches that collector only
// through a probe like this one.
const testTokenPermissionsProbe = "testTokenPermissionsRawData"

func TestMain(m *testing.M) {
	// Register before any parallel test reads the probe registry.
	proberegistration.MustRegister(testTokenPermissionsProbe,
		func(*checker.RawResults) ([]finding.Finding, string, error) { return nil, "", nil },
		[]checknames.CheckName{checknames.TokenPermissions})
	os.Exit(m.Run())
}

// rawDataMocks holds every client a probe-mode collector can call.
type rawDataMocks struct {
	repoClient    *mockrepo.MockRepoClient
	ossFuzzRepo   *mockrepo.MockRepoClient
	repo          *mockrepo.MockRepo
	cii           *mockrepo.MockCIIBestPracticesClient
	vulns         *mockrepo.MockVulnerabilitiesClient
	projectClient *mockrepo.MockProjectPackageClient
}

// newRawDataRequest builds a request shaped like the one probe mode
// builds in Run, with every client mocked.
func newRawDataRequest(t *testing.T, ret *Result) (*rawDataMocks, *checker.CheckRequest) {
	t.Helper()
	ctrl := gomock.NewController(t)
	m := &rawDataMocks{
		repoClient:    mockrepo.NewMockRepoClient(ctrl),
		ossFuzzRepo:   mockrepo.NewMockRepoClient(ctrl),
		repo:          mockrepo.NewMockRepo(ctrl),
		cii:           mockrepo.NewMockCIIBestPracticesClient(ctrl),
		vulns:         mockrepo.NewMockVulnerabilitiesClient(ctrl),
		projectClient: mockrepo.NewMockProjectPackageClient(ctrl),
	}
	req := &checker.CheckRequest{
		Ctx:                   context.Background(),
		RepoClient:            m.repoClient,
		OssFuzzRepo:           m.ossFuzzRepo,
		CIIClient:             m.cii,
		VulnerabilitiesClient: m.vulns,
		ProjectClient:         m.projectClient,
		Repo:                  m.repo,
		RawResults:            &ret.RawResults,
	}
	return m, req
}

const testRepoURI = "github.com/ossf-tests/scorecard"

const safeWorkflow = `on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo hi
`

// The digest pins the image, so no remediation (a registry lookup) is
// computed for it.
const (
	pinnedDockerfile        = "Dockerfile"
	pinnedDockerfileAt      = "3.7@sha256:5f957b2a08d223e48133e1a914ed046bea12e578fe2f6ae4de47fdbe691a2468"
	pinnedDockerfileContent = "FROM python:" + pinnedDockerfileAt
)

func asBoolPointer(b bool) *bool {
	return &b
}

// topLevelPermission is addressable so want can point at it, as
// TokenPermission.LocationType does.
var topLevelPermission = checker.PermissionLocationTop

func tokenPermissionDebug(msg string) checker.TokenPermission {
	return checker.TokenPermission{
		File: &checker.File{
			Path:   ".github/workflows/ci.yml",
			Type:   finding.FileTypeSource,
			Offset: checker.OffsetDefault,
		},
		Msg:  asPointer(msg),
		Type: checker.PermissionLevelUnknown,
	}
}

// rawDataCases pin, for each check, which client calls its collector
// makes and which RawResults field receives the data. Each fixture
// yields non-empty data, so a result stored in the wrong field fails.
var rawDataCases = []struct {
	expect func(m *rawDataMocks)
	check  string
	probe  string
	want   checker.RawResults
}{
	{
		check: checks.CheckCIIBestPractices,
		probe: hasOpenSSFBadge.Probe,
		expect: func(m *rawDataMocks) {
			m.repo.EXPECT().URI().Return(testRepoURI).Times(1)
			m.cii.EXPECT().GetBadgeLevel(gomock.Any(), testRepoURI).Return(clients.Silver, nil).Times(1)
		},
		want: checker.RawResults{
			CIIBestPracticesResults: checker.CIIBestPracticesData{Badge: clients.Silver},
		},
	},
	{
		check: checks.CheckBinaryArtifacts,
		probe: hasBinaryArtifacts.Probe,
		expect: func(m *rawDataMocks) {
			m.repoClient.EXPECT().ListFiles(gomock.Any()).Return([]string{"lib/libfoo.so"}, nil).Times(1)
			// No magic number needed: non-text bytes in a file with a
			// binary extension are flagged.
			m.repoClient.EXPECT().GetFileReader("lib/libfoo.so").
				Return(io.NopCloser(strings.NewReader("\x00\x00\x00\x00\x00")), nil).Times(1)
		},
		want: checker.RawResults{
			BinaryArtifactResults: checker.BinaryArtifactData{
				Files: []checker.File{{Path: "lib/libfoo.so", Type: finding.FileTypeBinary, Offset: checker.OffsetDefault}},
			},
		},
	},
	{
		check: checks.CheckBranchProtection,
		probe: branchesAreProtected.Probe,
		expect: func(m *rawDataMocks) {
			m.repoClient.EXPECT().GetDefaultBranch().Return(&clients.BranchRef{Name: asPointer("main")}, nil).Times(1)
			m.repoClient.EXPECT().ListReleases().Return(nil, nil).Times(1)
			// CODEOWNERS lookup; no files, so no GetFileReader.
			m.repoClient.EXPECT().ListFiles(gomock.Any()).Return(nil, nil).Times(1)
		},
		want: checker.RawResults{
			BranchProtectionResults: checker.BranchProtectionsData{
				Branches:        []clients.BranchRef{{Name: asPointer("main")}},
				CodeownersFiles: []string{},
			},
		},
	},
	{
		check: checks.CheckCodeReview,
		probe: codeApproved.Probe,
		expect: func(m *rawDataMocks) {
			m.repoClient.EXPECT().ListCommits().Return([]clients.Commit{{SHA: "c1"}}, nil).Times(1)
		},
		want: checker.RawResults{
			CodeReviewResults: checker.CodeReviewData{
				// One changeset, so the collector's map order cannot matter.
				DefaultBranchChangesets: []checker.Changeset{{
					ReviewPlatform: checker.ReviewPlatformUnknown,
					RevisionID:     "c1",
					Commits:        []clients.Commit{{SHA: "c1"}},
				}},
			},
		},
	},
	{
		check: checks.CheckContributors,
		probe: contributorsFromOrgOrCompany.Probe,
		expect: func(m *rawDataMocks) {
			m.repoClient.EXPECT().ListContributors().
				Return([]clients.User{{Login: "alice", NumContributions: 3}}, nil).Times(1)
		},
		want: checker.RawResults{
			ContributorsResults: checker.ContributorsData{
				Users: []clients.User{{Login: "alice", NumContributions: 3}},
			},
		},
	},
	{
		check: checks.CheckDangerousWorkflow,
		probe: hasDangerousWorkflowUntrustedCheckout.Probe,
		expect: func(m *rawDataMocks) {
			m.repoClient.EXPECT().ListFiles(gomock.Any()).Return([]string{".github/workflows/ci.yml"}, nil).Times(1)
			m.repoClient.EXPECT().GetFileReader(".github/workflows/ci.yml").
				Return(io.NopCloser(strings.NewReader(safeWorkflow)), nil).Times(1)
		},
		// A safe workflow still counts, which makes the data non-empty.
		want: checker.RawResults{
			DangerousWorkflowResults: checker.DangerousWorkflowData{NumWorkflows: 1},
		},
	},
	{
		check: checks.CheckDependencyUpdateTool,
		probe: dependencyUpdateToolConfigured.Probe,
		expect: func(m *rawDataMocks) {
			// A config file is found, so the dependabot commit search
			// (SearchCommits) is skipped.
			m.repoClient.EXPECT().ListFiles(gomock.Any()).Return([]string{".github/dependabot.yml"}, nil).Times(1)
		},
		want: checker.RawResults{
			DependencyUpdateToolResults: checker.DependencyUpdateToolData{
				Tools: []checker.Tool{{
					Name: "Dependabot",
					URL:  asPointer("https://github.com/dependabot"),
					Desc: asPointer("Automated dependency updates built into GitHub"),
					Files: []checker.File{{
						Path:   ".github/dependabot.yml",
						Type:   finding.FileTypeSource,
						Offset: checker.OffsetDefault,
					}},
				}},
			},
		},
	},
	{
		check: checks.CheckFuzzing,
		probe: fuzzed.Probe,
		expect: func(m *rawDataMocks) {
			// No ClusterFuzzLite Dockerfile and no languages, so only the
			// OSS-Fuzz lookup, served by the mocked OssFuzzRepo, finds a fuzzer.
			m.repoClient.EXPECT().ListFiles(gomock.Any()).Return(nil, nil).Times(1)
			m.repoClient.EXPECT().URI().Return(testRepoURI).Times(1)
			m.ossFuzzRepo.EXPECT().
				Search(clients.SearchRequest{Query: testRepoURI, Filename: "project.yaml"}).
				Return(clients.SearchResponse{Hits: 1}, nil).Times(1)
			m.repoClient.EXPECT().ListProgrammingLanguages().Return(nil, nil).Times(1)
		},
		want: checker.RawResults{
			FuzzingResults: checker.FuzzingData{
				Fuzzers: []checker.Tool{{
					Name: fuzzers.OSSFuzz,
					URL:  asPointer("https://github.com/google/oss-fuzz"),
					Desc: asPointer("Continuous Fuzzing for Open Source Software"),
				}},
			},
		},
	},
	{
		check: checks.CheckLicense,
		probe: hasLicenseFile.Probe,
		expect: func(m *rawDataMocks) {
			// The licenses API answers, so no file search follows.
			m.repoClient.EXPECT().ListLicenses().Return([]clients.License{
				{Path: "LICENSE", Name: "Apache License 2.0", SPDXId: "Apache-2.0"},
			}, nil).Times(1)
		},
		want: checker.RawResults{
			LicenseResults: checker.LicenseData{
				LicenseFiles: []checker.LicenseFile{{
					File: checker.File{Path: "LICENSE", Type: finding.FileTypeSource},
					LicenseInformation: checker.License{
						Name:        "Apache License 2.0",
						SpdxID:      "Apache-2.0",
						Attribution: checker.LicenseAttributionTypeAPI,
						Approved:    true,
					},
				}},
			},
		},
	},
	{
		check: checks.CheckPinnedDependencies,
		probe: pinsDependencies.Probe,
		expect: func(m *rawDataMocks) {
			// Five passes list files: workflow pinning, Dockerfile
			// pinning, Dockerfile downloads, shell script downloads and
			// workflow script downloads. Applying the predicate, as a real
			// client does, keeps the Dockerfile out of the workflow passes.
			m.repoClient.EXPECT().ListFiles(gomock.Any()).DoAndReturn(
				func(predicate func(string) (bool, error)) ([]string, error) {
					match, err := predicate(pinnedDockerfile)
					if err != nil || !match {
						return nil, err
					}
					return []string{pinnedDockerfile}, nil
				}).Times(5)
			// The three non-workflow passes each read the Dockerfile.
			m.repoClient.EXPECT().GetFileReader(pinnedDockerfile).DoAndReturn(
				func(string) (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader(pinnedDockerfileContent)), nil
				}).Times(3)
			// Workflow remediation metadata is built even with no workflows.
			m.repoClient.EXPECT().GetDefaultBranchName().Return("main", nil).Times(1)
			m.repoClient.EXPECT().URI().Return(testRepoURI).Times(1)
		},
		want: checker.RawResults{
			PinningDependenciesResults: checker.PinningDependenciesData{
				Dependencies: []checker.Dependency{{
					Location: &checker.File{
						Path:      pinnedDockerfile,
						Type:      finding.FileTypeSource,
						Offset:    1,
						EndOffset: 1,
						Snippet:   pinnedDockerfileContent,
					},
					Name:     asPointer("python"),
					PinnedAt: asPointer(pinnedDockerfileAt),
					Pinned:   asBoolPointer(true),
					Type:     checker.DependencyUseTypeDockerfileContainerImage,
				}},
			},
		},
	},
	{
		check: checks.CheckSAST,
		probe: sastToolConfigured.Probe,
		expect: func(m *rawDataMocks) {
			m.repoClient.EXPECT().ListCommits().Return([]clients.Commit{mergedCommit("c1", "head-a", 1)}, nil).Times(1)
			m.repoClient.EXPECT().ListCheckRunsForRef("head-a").Return(nil, nil).Times(1)
			// One workflow pass per SAST tool: CodeQL, Sonar, Snyk, Pysa,
			// Qodana and Hadolint.
			m.repoClient.EXPECT().ListFiles(gomock.Any()).Return(nil, nil).Times(6)
		},
		want: checker.RawResults{
			SASTResults: checker.SASTData{
				Commits: []checker.SASTCommit{{
					SHA:                    "c1",
					AssociatedMergeRequest: mergedCommit("c1", "head-a", 1).AssociatedMergeRequest,
				}},
			},
		},
	},
	{
		// SBOM is experimental, but probe mode collects it without
		// SCORECARD_EXPERIMENTAL; only the check's Run gates on it.
		check: checks.CheckSBOM,
		probe: hasSBOM.Probe,
		expect: func(m *rawDataMocks) {
			m.repoClient.EXPECT().ListReleases().Return(nil, nil).Times(1)
			m.repoClient.EXPECT().ListFiles(gomock.Any()).Return([]string{"app.spdx.json"}, nil).Times(1)
		},
		want: checker.RawResults{
			SBOMResults: checker.SBOMData{SBOMFiles: []checker.SBOM{{
				Name: "app.spdx.json",
				File: checker.File{Path: "app.spdx.json", Type: finding.FileTypeSource},
			}}},
		},
	},
	{
		check: checks.CheckSecurityPolicy,
		probe: securityPolicyPresent.Probe,
		expect: func(m *rawDataMocks) {
			// One listing finds the policy, a second matches it for reading.
			m.repoClient.EXPECT().ListFiles(gomock.Any()).Return([]string{"SECURITY.md"}, nil).Times(2)
			m.repoClient.EXPECT().GetFileReader("SECURITY.md").
				Return(io.NopCloser(strings.NewReader("x")), nil).Times(1)
		},
		want: checker.RawResults{
			SecurityPolicyResults: checker.SecurityPolicyData{PolicyFiles: []checker.SecurityPolicyFile{{
				File: checker.File{
					Path:     "SECURITY.md",
					Type:     finding.FileTypeText,
					Offset:   checker.OffsetDefault,
					FileSize: 1,
				},
				Information: []checker.SecurityPolicyInformation{},
			}}},
		},
	},
	{
		check: checks.CheckSignedReleases,
		probe: releasesAreSigned.Probe,
		expect: func(m *rawDataMocks) {
			m.repoClient.EXPECT().ListReleases().Return([]clients.Release{{TagName: "v1.0.0"}}, nil).Times(1)
			m.repo.EXPECT().Host().Return("github.com").Times(1)
			m.repo.EXPECT().Path().Return("ossf-tests/scorecard").Times(1)
			m.projectClient.EXPECT().
				GetProjectPackageVersions(gomock.Any(), "github.com", "ossf-tests/scorecard").
				Return(&packageclient.ProjectPackageVersions{}, nil).Times(1)
		},
		want: checker.RawResults{
			SignedReleasesResults: checker.SignedReleasesData{
				Releases: []clients.Release{{TagName: "v1.0.0"}},
				Packages: []checker.ProjectPackage{},
			},
		},
	},
	{
		check: checks.CheckTokenPermissions,
		probe: testTokenPermissionsProbe,
		expect: func(m *rawDataMocks) {
			m.repoClient.EXPECT().ListFiles(gomock.Any()).Return([]string{".github/workflows/ci.yml"}, nil).Times(1)
			m.repoClient.EXPECT().GetFileReader(".github/workflows/ci.yml").
				Return(io.NopCloser(strings.NewReader("on: push\n")), nil).Times(1)
		},
		want: checker.RawResults{
			TokenPermissionsResults: checker.TokenPermissionsData{
				NumTokens: 1,
				TokenPermissions: []checker.TokenPermission{
					{
						File: &checker.File{
							Path:   ".github/workflows/ci.yml",
							Type:   finding.FileTypeSource,
							Offset: checker.OffsetDefault,
						},
						LocationType: &topLevelPermission,
						Type:         checker.PermissionLevelUndeclared,
					},
					// The collector records these debug entries for every
					// workflow, even a minimal one.
					tokenPermissionDebug("not a publishing workflow: .github/workflows/ci.yml"),
					tokenPermissionDebug("not a releasing workflow: .github/workflows/ci.yml"),
					tokenPermissionDebug("not a GitHub Pages deployment workflow: .github/workflows/ci.yml"),
					tokenPermissionDebug("not a SARIF workflow, or not an allowed one"),
				},
			},
		},
	},
	{
		check: checks.CheckVulnerabilities,
		probe: hasOSVVulnerabilities.Probe,
		expect: func(m *rawDataMocks) {
			m.repoClient.EXPECT().ListCommits().Return([]clients.Commit{{SHA: "c1"}}, nil).Times(1)
			m.repoClient.EXPECT().LocalPath().Return("/repo", nil).Times(1)
			m.vulns.EXPECT().ListUnfixedVulnerabilities(gomock.Any(), "c1", "/repo").
				Return(clients.VulnerabilitiesResponse{
					Vulnerabilities: []clients.Vulnerability{{ID: "OSV-TEST-1"}},
				}, nil).Times(1)
		},
		want: checker.RawResults{
			VulnerabilitiesResults: checker.VulnerabilitiesData{
				Vulnerabilities: []clients.Vulnerability{{ID: "OSV-TEST-1"}},
			},
		},
	},
	{
		// Experimental like SBOM, and likewise not gated in probe mode.
		check: checks.CheckWebHooks,
		probe: webhooksUseSecrets.Probe,
		expect: func(m *rawDataMocks) {
			m.repoClient.EXPECT().ListWebhooks().Return([]clients.Webhook{{ID: 1}}, nil).Times(1)
		},
		want: checker.RawResults{
			WebhookResults: checker.WebhooksData{Webhooks: []clients.Webhook{{ID: 1}}},
		},
	},
}

func TestPopulateRawResults_EachCheck(t *testing.T) {
	t.Parallel()
	for _, tt := range rawDataCases {
		t.Run(tt.check, func(t *testing.T) {
			t.Parallel()
			var ret Result
			m, req := newRawDataRequest(t, &ret)
			tt.expect(m)

			if err := populateRawResults(req, []string{tt.probe}, &ret); err != nil {
				t.Fatalf("populateRawResults: %v", err)
			}
			if diff := cmp.Diff(tt.want, ret.RawResults); diff != "" {
				t.Errorf("RawResults mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// rawDataErrorCases pin each check's error text when its collector's
// first client call fails, and that collection stops there.
var rawDataErrorCases = []struct {
	expect  func(m *rawDataMocks)
	check   string
	probe   string
	wantErr string
}{
	{
		check: checks.CheckCIIBestPractices,
		probe: hasOpenSSFBadge.Probe,
		expect: func(m *rawDataMocks) {
			m.repo.EXPECT().URI().Return(testRepoURI)
			m.cii.EXPECT().GetBadgeLevel(gomock.Any(), testRepoURI).Return(clients.Unknown, errCollect)
		},
		wantErr: "internal error: collect failed",
	},
	{
		check: checks.CheckBinaryArtifacts,
		probe: hasBinaryArtifacts.Probe,
		expect: func(m *rawDataMocks) {
			m.repoClient.EXPECT().ListFiles(gomock.Any()).Return(nil, errCollect)
		},
		wantErr: "internal error: error during ListFiles: collect failed",
	},
	{
		check: checks.CheckBranchProtection,
		probe: branchesAreProtected.Probe,
		expect: func(m *rawDataMocks) {
			m.repoClient.EXPECT().GetDefaultBranch().Return(nil, errCollect)
		},
		wantErr: "internal error: collect failed",
	},
	{
		check: checks.CheckCodeReview,
		probe: codeApproved.Probe,
		expect: func(m *rawDataMocks) {
			m.repoClient.EXPECT().ListCommits().Return(nil, errCollect)
		},
		wantErr: "internal error: collect failed",
	},
	{
		check: checks.CheckContributors,
		probe: contributorsFromOrgOrCompany.Probe,
		expect: func(m *rawDataMocks) {
			m.repoClient.EXPECT().ListContributors().Return(nil, errCollect)
		},
		wantErr: "internal error: Client.Repositories.ListContributors: collect failed",
	},
	{
		check: checks.CheckDangerousWorkflow,
		probe: hasDangerousWorkflowUntrustedCheckout.Probe,
		expect: func(m *rawDataMocks) {
			m.repoClient.EXPECT().ListFiles(gomock.Any()).Return(nil, errCollect)
		},
		wantErr: "internal error: error during ListFiles: collect failed",
	},
	{
		check: checks.CheckDependencyUpdateTool,
		probe: dependencyUpdateToolConfigured.Probe,
		expect: func(m *rawDataMocks) {
			m.repoClient.EXPECT().ListFiles(gomock.Any()).Return(nil, errCollect)
		},
		wantErr: "internal error: error during ListFiles: collect failed",
	},
	{
		check: checks.CheckFuzzing,
		probe: fuzzed.Probe,
		expect: func(m *rawDataMocks) {
			m.repoClient.EXPECT().ListFiles(gomock.Any()).Return(nil, errCollect)
		},
		wantErr: "internal error: error during ListFiles: collect failed",
	},
	{
		check: checks.CheckLicense,
		probe: hasLicenseFile.Probe,
		expect: func(m *rawDataMocks) {
			m.repoClient.EXPECT().ListLicenses().Return(nil, errCollect)
		},
		wantErr: "internal error: RepoClient.ListLicenses: collect failed",
	},
	{
		check: checks.CheckPinnedDependencies,
		probe: pinsDependencies.Probe,
		expect: func(m *rawDataMocks) {
			m.repoClient.EXPECT().ListFiles(gomock.Any()).Return(nil, errCollect)
		},
		wantErr: "internal error: error during ListFiles: collect failed",
	},
	{
		check: checks.CheckSAST,
		probe: sastToolConfigured.Probe,
		expect: func(m *rawDataMocks) {
			m.repoClient.EXPECT().ListCommits().Return(nil, errCollect)
		},
		// raw.SAST already wraps in ErrScorecardInternal, and
		// assignRawData wraps again.
		wantErr: "internal error: internal error: RepoClient.ListCommits: collect failed",
	},
	{
		check: checks.CheckSBOM,
		probe: hasSBOM.Probe,
		expect: func(m *rawDataMocks) {
			m.repoClient.EXPECT().ListReleases().Return(nil, errCollect)
		},
		wantErr: "internal error: RepoClient.ListReleases: collect failed",
	},
	{
		check: checks.CheckSecurityPolicy,
		probe: securityPolicyPresent.Probe,
		expect: func(m *rawDataMocks) {
			m.repoClient.EXPECT().ListFiles(gomock.Any()).Return(nil, errCollect)
		},
		wantErr: "internal error: error during ListFiles: collect failed",
	},
	{
		check: checks.CheckSignedReleases,
		probe: releasesAreSigned.Probe,
		expect: func(m *rawDataMocks) {
			m.repoClient.EXPECT().ListReleases().Return(nil, errCollect)
		},
		wantErr: "internal error: collect failed",
	},
	{
		check: checks.CheckTokenPermissions,
		probe: testTokenPermissionsProbe,
		expect: func(m *rawDataMocks) {
			m.repoClient.EXPECT().ListFiles(gomock.Any()).Return(nil, errCollect)
		},
		wantErr: "internal error: error during ListFiles: collect failed",
	},
	{
		// The collector ignores a ListCommits error and scans with an
		// empty commit, so the first error it returns is from LocalPath.
		check: checks.CheckVulnerabilities,
		probe: hasOSVVulnerabilities.Probe,
		expect: func(m *rawDataMocks) {
			m.repoClient.EXPECT().ListCommits().Return(nil, errCollect)
			m.repoClient.EXPECT().LocalPath().Return("", errCollect)
		},
		wantErr: "internal error: RepoClient.LocalPath: collect failed",
	},
	{
		// raw.WebHook already wraps in ErrScorecardInternal and drops
		// the client error, so its cause never reaches the caller.
		check: checks.CheckWebHooks,
		probe: webhooksUseSecrets.Probe,
		expect: func(m *rawDataMocks) {
			m.repoClient.EXPECT().ListWebhooks().Return(nil, errCollect)
		},
		wantErr: "internal error: internal error: Client.Repositories.ListWebhooks",
	},
}

func TestPopulateRawResults_EachCheckError(t *testing.T) {
	t.Parallel()
	for _, tt := range rawDataErrorCases {
		t.Run(tt.check, func(t *testing.T) {
			t.Parallel()
			var ret Result
			m, req := newRawDataRequest(t, &ret)
			tt.expect(m)

			err := populateRawResults(req, []string{tt.probe}, &ret)
			if err == nil {
				t.Fatalf("populateRawResults: want error %q, got nil", tt.wantErr)
			}
			if err.Error() != tt.wantErr {
				t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
			}
			if !errors.Is(err, sce.ErrScorecardInternal) {
				t.Errorf("errors.Is(%v, sce.ErrScorecardInternal) = false", err)
			}
		})
	}
}
