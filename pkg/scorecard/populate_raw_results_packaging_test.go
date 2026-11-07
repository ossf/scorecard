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
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net/http"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/clients/githubrepo"
	"github.com/ossf/scorecard/v5/clients/localdir"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/log"
	"github.com/ossf/scorecard/v5/probes/packagedWithAutomatedWorkflow"
)

// Packaging picks its collector by the concrete RepoClient type, so
// these tests use a real GitHub client over a fake transport instead
// of a mock. The GitLab client cannot take a test transport and is not
// covered here.

const (
	packagingWorkflowPath = ".github/workflows/publish.yml"
	packagingRunsPath     = "/repos/ossf-tests/scorecard/actions/workflows/publish.yml/runs"
	packagingRunURL       = "https://api.github.com/repos/ossf-tests/scorecard/actions/runs/1"
)

const npmPublishWorkflow = `on: push
jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/setup-node@v4
        with:
          registry-url: https://registry.npmjs.org
      - run: npm publish
`

type fakeResponse struct {
	body   []byte
	status int
}

// fakeGitHub serves canned responses by URL path and records every
// request it receives, so tests can pin which API calls a collector
// makes.
type fakeGitHub struct {
	responses map[string]fakeResponse
	requests  []string
	mu        sync.Mutex
}

func (f *fakeGitHub) RoundTrip(req *http.Request) (*http.Response, error) {
	f.mu.Lock()
	f.requests = append(f.requests, req.Method+" "+req.URL.RequestURI())
	f.mu.Unlock()
	r, ok := f.responses[req.URL.Path]
	if !ok {
		r = fakeResponse{status: http.StatusNotFound, body: []byte(`{}`)}
	}
	return &http.Response{
		StatusCode: r.status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader(r.body)),
		Request:    req,
	}, nil
}

// workflowTarball returns a gzipped tarball, laid out like GitHub's,
// holding one workflow file.
func workflowTarball(t *testing.T, workflow string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	const root = "ossf-tests-scorecard-0000000/"
	for _, dir := range []string{root, root + ".github/", root + ".github/workflows/"} {
		if err := tw.WriteHeader(&tar.Header{Name: dir, Typeflag: tar.TypeDir, Mode: 0o755}); err != nil {
			t.Fatalf("tar.WriteHeader: %v", err)
		}
	}
	hdr := &tar.Header{
		Name:     root + packagingWorkflowPath,
		Typeflag: tar.TypeReg,
		Mode:     0o644,
		Size:     int64(len(workflow)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("tar.WriteHeader: %v", err)
	}
	if _, err := tw.Write([]byte(workflow)); err != nil {
		t.Fatalf("tar.Write: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("tar.Close: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("gzip.Close: %v", err)
	}
	return buf.Bytes()
}

// newFakeGitHubClient returns a GitHub client initialized at HEAD
// against fake, with the setup requests cleared.
func newFakeGitHubClient(t *testing.T, fake *fakeGitHub) clients.RepoClient {
	t.Helper()
	fake.responses["/repos/ossf-tests/scorecard"] = fakeResponse{status: http.StatusOK, body: []byte(`{
		"name": "scorecard",
		"owner": {"login": "ossf-tests"},
		"default_branch": "main",
		"archive_url": "https://api.github.com/repos/ossf-tests/scorecard/{archive_format}{/ref}"
	}`)}
	client := githubrepo.CreateGithubRepoClientWithTransport(context.Background(), fake)
	repo, err := githubrepo.MakeGithubRepo(testRepoURI)
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
	fake.requests = nil
	return client
}

func packagingFile(offset uint) *checker.File {
	return &checker.File{Path: packagingWorkflowPath, Type: finding.FileTypeSource, Offset: offset}
}

func packagingMsg(msg string) *string {
	return &msg
}

const (
	requestTarball = "GET /repos/ossf-tests/scorecard/tarball/"
	requestRuns    = "GET " + packagingRunsPath + "?status=success"
)

func TestPopulateRawResults_PackagingGitHub(t *testing.T) {
	t.Parallel()
	tests := []struct {
		runs     *fakeResponse
		name     string
		workflow string
		requests []string
		want     checker.PackagingData
	}{
		{
			name:     "publishing workflow with a successful run",
			workflow: npmPublishWorkflow,
			runs: &fakeResponse{status: http.StatusOK, body: []byte(`{"total_count": 1, "workflow_runs": [` +
				`{"url": "` + packagingRunURL + `"}]}`)},
			requests: []string{requestTarball, requestRuns},
			want: checker.PackagingData{Packages: []checker.Package{
				{
					Msg:  packagingMsg("candidate node publishing workflow using npm: " + packagingWorkflowPath),
					File: packagingFile(checker.OffsetDefault),
				},
				{
					File: packagingFile(3),
					// The first run is stored twice.
					Runs: []checker.Run{{URL: packagingRunURL}, {URL: packagingRunURL}},
				},
			}},
		},
		{
			name:     "publishing workflow with no successful runs",
			workflow: npmPublishWorkflow,
			runs:     &fakeResponse{status: http.StatusOK, body: []byte(`{"total_count": 0, "workflow_runs": []}`)},
			requests: []string{requestTarball, requestRuns},
			want: checker.PackagingData{Packages: []checker.Package{
				{
					Msg:  packagingMsg("candidate node publishing workflow using npm: " + packagingWorkflowPath),
					File: packagingFile(checker.OffsetDefault),
				},
				{
					Msg:  packagingMsg("GitHub publishing workflow not used in runs: " + packagingWorkflowPath),
					File: packagingFile(checker.OffsetDefault),
				},
			}},
		},
		{
			name:     "workflow that does not publish",
			workflow: safeWorkflow,
			requests: []string{requestTarball},
			want: checker.PackagingData{Packages: []checker.Package{
				{
					Msg:  packagingMsg("not a publishing workflow: " + packagingWorkflowPath),
					File: packagingFile(checker.OffsetDefault),
				},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fake := &fakeGitHub{responses: map[string]fakeResponse{
				"/repos/ossf-tests/scorecard/tarball/": {status: http.StatusOK, body: workflowTarball(t, tt.workflow)},
			}}
			if tt.runs != nil {
				fake.responses[packagingRunsPath] = *tt.runs
			}
			var ret Result
			req := &checker.CheckRequest{
				Ctx:        context.Background(),
				RepoClient: newFakeGitHubClient(t, fake),
				RawResults: &ret.RawResults,
			}

			if err := populateRawResults(req, []string{packagedWithAutomatedWorkflow.Probe}, &ret); err != nil {
				t.Fatalf("populateRawResults: %v", err)
			}
			if diff := cmp.Diff(checker.RawResults{PackagingResults: tt.want}, ret.RawResults); diff != "" {
				t.Errorf("RawResults mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.requests, fake.requests); diff != "" {
				t.Errorf("requests mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPopulateRawResults_PackagingError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		client   func(t *testing.T) (clients.RepoClient, func() []string)
		name     string
		wantErr  string
		requests []string
	}{
		{
			name: "github workflow runs request fails",
			client: func(t *testing.T) (clients.RepoClient, func() []string) {
				t.Helper()
				fake := &fakeGitHub{responses: map[string]fakeResponse{
					"/repos/ossf-tests/scorecard/tarball/": {
						status: http.StatusOK, body: workflowTarball(t, npmPublishWorkflow),
					},
					packagingRunsPath: {status: http.StatusInternalServerError, body: []byte(`{}`)},
				}}
				return newFakeGitHubClient(t, fake), func() []string { return fake.requests }
			},
			wantErr: "internal error: Client.Actions.ListWorkflowRunsByFileName: internal error: " +
				"ListWorkflowRunsByFileName: GET https://api.github.com" + packagingRunsPath + "?status=success: 500  []",
			requests: []string{requestTarball, requestRuns},
		},
		{
			name: "local directory",
			client: func(t *testing.T) (clients.RepoClient, func() []string) {
				t.Helper()
				return localdir.CreateLocalDirClient(context.Background(), log.NewLogger(log.DefaultLevel)), nil
			},
			wantErr: "internal error: Only github and gitlab are supported",
		},
		{
			name: "mock client",
			client: func(t *testing.T) (clients.RepoClient, func() []string) {
				t.Helper()
				// The mock fails the test on any call.
				return mockrepo.NewMockRepoClient(gomock.NewController(t)), nil
			},
			wantErr: "internal error: Only github and gitlab are supported",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client, requests := tt.client(t)
			var ret Result
			req := &checker.CheckRequest{
				Ctx:        context.Background(),
				RepoClient: client,
				RawResults: &ret.RawResults,
			}

			err := populateRawResults(req, []string{packagedWithAutomatedWorkflow.Probe}, &ret)
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("error = %v, want %q", err, tt.wantErr)
			}
			if !errors.Is(err, sce.ErrScorecardInternal) {
				t.Errorf("error %v is not ErrScorecardInternal", err)
			}
			if diff := cmp.Diff(checker.RawResults{}, ret.RawResults); diff != "" {
				t.Errorf("RawResults mismatch (-want +got):\n%s", diff)
			}
			if requests != nil {
				if diff := cmp.Diff(tt.requests, requests()); diff != "" {
					t.Errorf("requests mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
