// Copyright 2025 OpenSSF Scorecard Authors
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
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/checks/raw"
	"github.com/ossf/scorecard/v5/clients"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	scut "github.com/ossf/scorecard/v5/utests"
)

// TestReleasesDirectDepsVulnFree_E2E_Mock provides comprehensive end-to-end testing
// with all external dependencies mocked (HTTP tarballs, OSV API).
func TestReleasesDirectDepsVulnFree_E2E_Mock(t *testing.T) {
	t.Parallel()

	// Test scenario: Repository with 3 releases
	// - v1.0.0: Clean (no vulnerabilities)
	// - v0.9.0: Vulnerable (has known CVEs)
	// - v0.8.0: Clean (no vulnerabilities)
	// Expected score: 2/3 clean = 6.67/10 = score 6

	tests := []struct {
		name           string
		expectedReason string
		releases       []releaseScenario
		expectedScore  int
		wantErr        bool
	}{
		{
			name: "all releases clean",
			releases: []releaseScenario{
				{
					tag:    "v1.0.0",
					commit: "abc123",
					deps: []dependency{
						{ecosystem: "npm", name: "lodash", version: "4.17.21", hasVuln: false},
						{ecosystem: "pypi", name: "requests", version: "2.31.0", hasVuln: false},
					},
				},
				{
					tag:    "v0.9.0",
					commit: "def456",
					deps: []dependency{
						{ecosystem: "npm", name: "axios", version: "1.6.0", hasVuln: false},
					},
				},
			},
			expectedScore:  10,
			expectedReason: "2/2 recent releases had no known vulnerable direct dependencies",
		},
		{
			name: "all releases vulnerable",
			releases: []releaseScenario{
				{
					tag:    "v1.0.0",
					commit: "abc123",
					deps: []dependency{
						{ecosystem: "npm", name: "lodash", version: "4.17.0", hasVuln: true, vulnIDs: []string{"CVE-2020-8203"}},
					},
				},
				{
					tag:    "v0.9.0",
					commit: "def456",
					deps: []dependency{
						{ecosystem: "pypi", name: "requests", version: "2.6.0", hasVuln: true, vulnIDs: []string{"CVE-2018-18074"}},
					},
				},
			},
			expectedScore:  0,
			expectedReason: "0/2 recent releases had no known vulnerable direct dependencies",
		},
		{
			name: "mixed clean and vulnerable releases",
			releases: []releaseScenario{
				{
					tag:    "v1.0.0",
					commit: "abc123",
					deps: []dependency{
						{ecosystem: "npm", name: "express", version: "4.18.0", hasVuln: false},
					},
				},
				{
					tag:    "v0.9.0",
					commit: "def456",
					deps: []dependency{
						{ecosystem: "npm", name: "lodash", version: "4.17.0", hasVuln: true, vulnIDs: []string{"CVE-2020-8203"}},
					},
				},
				{
					tag:    "v0.8.0",
					commit: "ghi789",
					deps: []dependency{
						{ecosystem: "pypi", name: "requests", version: "2.31.0", hasVuln: false},
					},
				},
			},
			expectedScore:  6, // 2/3 clean = 6.67 -> rounds to 6
			expectedReason: "2/3 recent releases had no known vulnerable direct dependencies",
		},
		{
			name: "single release with multiple dependencies",
			releases: []releaseScenario{
				{
					tag:    "v1.0.0",
					commit: "abc123",
					deps: []dependency{
						{ecosystem: "npm", name: "lodash", version: "4.17.21", hasVuln: false},
						{ecosystem: "npm", name: "axios", version: "1.6.0", hasVuln: false},
						{ecosystem: "pypi", name: "requests", version: "2.31.0", hasVuln: false},
						{ecosystem: "golang", name: "github.com/pkg/errors", version: "0.9.1", hasVuln: false},
					},
				},
			},
			expectedScore:  10,
			expectedReason: "1/1 recent releases had no known vulnerable direct dependencies",
		},
		{
			name: "release with mixed vulnerable and clean deps",
			releases: []releaseScenario{
				{
					tag:    "v1.0.0",
					commit: "abc123",
					deps: []dependency{
						{ecosystem: "npm", name: "lodash", version: "4.17.21", hasVuln: false},
						{ecosystem: "npm", name: "express", version: "4.0.0", hasVuln: true, vulnIDs: []string{"CVE-2014-6393"}},
					},
				},
			},
			expectedScore:  0, // Any vulnerability means the release is not clean
			expectedReason: "0/1 recent releases had no known vulnerable direct dependencies",
		},
		{
			name: "golang module dependencies",
			releases: []releaseScenario{
				{
					tag:    "v1.0.0",
					commit: "abc123",
					deps: []dependency{
						{ecosystem: "golang", name: "golang.org/x/crypto", version: "0.17.0", hasVuln: false},
						{ecosystem: "golang", name: "github.com/gin-gonic/gin", version: "1.9.1", hasVuln: false},
					},
				},
			},
			expectedScore:  10,
			expectedReason: "1/1 recent releases had no known vulnerable direct dependencies",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Set up mock HTTP servers for tarballs and OSV API
			tarballServer, osvServer := setupMockServers(t, tt.releases)
			defer tarballServer.Close()
			defer osvServer.Close()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepoBase := mockrepo.NewMockRepoClient(ctrl)
			mockRepo := &mockRepoClientWithTarball{
				MockRepoClient: mockRepoBase,
				tarballServer:  tarballServer,
			}

			// Convert test scenarios to clients.Release
			var releases []clients.Release
			for _, rs := range tt.releases {
				releases = append(releases, clients.Release{
					TagName:         rs.tag,
					TargetCommitish: rs.commit,
					URL:             fmt.Sprintf("https://github.com/test/repo/releases/tag/%s", rs.tag),
				})
			}

			mockRepoBase.EXPECT().ListReleases().Return(releases, nil).AnyTimes()
			mockRepoBase.EXPECT().URI().Return("github.com/test/repo").AnyTimes()

			// Create mock HTTP client
			mockHTTPClient := &http.Client{
				Transport: &http.Transport{},
			}

			// Create mock Scalibr client with access to test scenarios
			mockScalibrClient := &mockDepsClient{
				tarballServer: tarballServer,
				releases:      tt.releases,
			}

			// Create mock OSV client with access to test scenarios
			mockOSVClient := &mockOSVClient{
				osvServer: osvServer,
				releases:  tt.releases,
			}

			dl := scut.TestDetailLogger{}
			req := checker.CheckRequest{
				RepoClient: mockRepo,
				Ctx:        context.Background(),
				Dlogger:    &dl,
			}

			// Create injectable clients
			testClients := &raw.CollectorClients{
				OSV:  mockOSVClient,
				Deps: mockScalibrClient,
				HTTP: mockHTTPClient,
			}

			// Execute check with injectable clients
			rawData, err := raw.ReleasesDirectDepsVulnFreeWithClients(&req, testClients)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify raw data is structured correctly
			if rawData == nil {
				t.Errorf("expected non-nil raw data")
				return
			}

			// Count clean releases (those with no vulnerable dependencies)
			cleanReleases := 0
			for _, rel := range rawData.Releases {
				if len(rel.Findings) == 0 {
					cleanReleases++
				}
			}

			totalReleases := len(rawData.Releases)
			var actualScore int
			if totalReleases > 0 {
				actualScore = (cleanReleases * checker.MaxResultScore) / totalReleases
			}

			if actualScore != tt.expectedScore {
				t.Errorf("expected score %d, got %d (clean=%d, total=%d)",
					tt.expectedScore, actualScore, cleanReleases, totalReleases)
			}
		})
	}
}

// releaseScenario describes a release with its dependencies for testing.
type releaseScenario struct {
	tag    string
	commit string
	deps   []dependency
}

// dependency describes a package dependency for testing.
type dependency struct {
	ecosystem string
	name      string
	version   string
	vulnIDs   []string
	hasVuln   bool
}

// setupMockServers creates HTTP test servers for tarballs and OSV API.
func setupMockServers(t *testing.T, releases []releaseScenario) (*httptest.Server, *httptest.Server) {
	t.Helper()
	// Tarball server: serves .tar.gz files with manifest files
	tarballServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract tag from path: /v1.0.0.tar.gz
		tag := r.URL.Path[1 : len(r.URL.Path)-7] // Remove leading / and .tar.gz

		// Find the release scenario
		var rs *releaseScenario
		for i := range releases {
			if releases[i].tag == tag {
				rs = &releases[i]
				break
			}
		}

		if rs == nil {
			http.NotFound(w, r)
			return
		}

		// Create tarball with manifest file
		tarball := createTarballWithManifest(t, rs)
		w.Header().Set("Content-Type", "application/gzip")
		if _, err := w.Write(tarball); err != nil {
			t.Errorf("failed to write tarball: %v", err)
		}
	}))

	// OSV API server: responds to /v1/querybatch
	osvServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/querybatch" {
			http.NotFound(w, r)
			return
		}

		// Parse the batch query
		var queries struct {
			Queries []map[string]interface{} `json:"queries"`
		}
		if err := json.NewDecoder(r.Body).Decode(&queries); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Build response based on test data
		response := make([]map[string]interface{}, len(queries.Queries))
		for i := range queries.Queries {
			// For this mock, we'll return vulnerabilities based on hardcoded rules
			// In reality, you'd match against the test scenario
			response[i] = map[string]interface{}{
				"vulns": []map[string]interface{}{}, // Empty array = no vulns
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"results": response,
		}); err != nil {
			t.Errorf("failed to encode OSV response: %v", err)
		}
	}))

	return tarballServer, osvServer
}

// createTarballWithManifest creates a .tar.gz file with appropriate manifest files.
func createTarballWithManifest(t *testing.T, rs *releaseScenario) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	// Create a directory structure with manifest files based on dependencies
	hasNPM := false
	hasPython := false
	hasGo := false

	for _, dep := range rs.deps {
		switch dep.ecosystem {
		case "npm":
			hasNPM = true
		case "pypi":
			hasPython = true
		case "golang":
			hasGo = true
		}
	}

	// Add package.json for npm dependencies
	if hasNPM {
		var npmDeps []string
		for _, dep := range rs.deps {
			if dep.ecosystem == "npm" {
				npmDeps = append(npmDeps, fmt.Sprintf(`    "%s": "%s"`, dep.name, dep.version))
			}
		}
		packageJSON := fmt.Sprintf(`{
  "name": "test-repo",
  "version": "1.0.0",
  "dependencies": {
%s
  }
}`, joinStrings(npmDeps, ",\n"))

		addFileToTar(t, tw, fmt.Sprintf("repo-%s/package.json", rs.tag), packageJSON)
	}

	// Add requirements.txt for Python dependencies
	if hasPython {
		var lines []string
		for _, dep := range rs.deps {
			if dep.ecosystem == "pypi" {
				lines = append(lines, fmt.Sprintf("%s==%s", dep.name, dep.version))
			}
		}
		addFileToTar(t, tw, fmt.Sprintf("repo-%s/requirements.txt", rs.tag), joinStrings(lines, "\n"))
	}

	// Add go.mod for Go dependencies
	if hasGo {
		var requires []string
		for _, dep := range rs.deps {
			if dep.ecosystem == "golang" {
				requires = append(requires, fmt.Sprintf("\t%s v%s", dep.name, dep.version))
			}
		}
		goMod := fmt.Sprintf(`module github.com/test/repo

go 1.21

require (
%s
)`, joinStrings(requires, "\n"))

		addFileToTar(t, tw, fmt.Sprintf("repo-%s/go.mod", rs.tag), goMod)
	}

	tw.Close()
	gw.Close()

	return buf.Bytes()
}

// addFileToTar adds a file to a tar archive.
func addFileToTar(t *testing.T, tw *tar.Writer, name, content string) {
	t.Helper()
	hdr := &tar.Header{
		Name: name,
		Mode: 0o644,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("failed to write tar header: %v", err)
	}
	if _, err := tw.Write([]byte(content)); err != nil {
		t.Fatalf("failed to write tar content: %v", err)
	}
}

// joinStrings joins strings with a separator.
func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

// mockDepsClient implements clients.DepsClient for testing.
type mockDepsClient struct {
	tarballServer *httptest.Server
	releases      []releaseScenario
	callIndex     int
}

func (m *mockDepsClient) GetDeps(ctx context.Context, localDir string) (clients.DepsResponse, error) {
	// Return dependencies for the current release being processed
	// The raw collector processes releases in order, so we track with callIndex
	if m.callIndex >= len(m.releases) {
		return clients.DepsResponse{Deps: []clients.Dep{}}, nil
	}

	matchedRelease := &m.releases[m.callIndex]
	m.callIndex++

	// Convert test dependencies to clients.Dep format
	var deps []clients.Dep
	for _, d := range matchedRelease.deps {
		deps = append(deps, clients.Dep{
			Ecosystem: d.ecosystem,
			Name:      d.name,
			Version:   d.version,
			PURL:      fmt.Sprintf("pkg:%s/%s@%s", d.ecosystem, d.name, d.version),
		})
	}

	return clients.DepsResponse{
		Deps: deps,
	}, nil
}

// mockOSVClient implements clients.OSVAPIClient for testing.
type mockOSVClient struct {
	osvServer *httptest.Server
	releases  []releaseScenario
}

func (m *mockOSVClient) QueryBatch(ctx context.Context, queries []clients.OSVQuery) ([][]string, error) {
	// Return vulnerabilities based on our test data
	results := make([][]string, len(queries))

	for i, q := range queries {
		// Extract package name from either direct name or PURL
		var queryName, queryEco string
		if q.Package.PURL != "" {
			// Parse PURL: pkg:ecosystem/name@version
			// Example: pkg:npm/lodash@4.17.0
			parts := strings.SplitN(q.Package.PURL, "/", 2)
			if len(parts) == 2 {
				ecosystemPart := strings.TrimPrefix(parts[0], "pkg:")
				namePart := strings.Split(parts[1], "@")[0]
				queryName = namePart
				queryEco = ecosystemPart
			}
		} else {
			queryName = q.Package.Name
			queryEco = strings.ToLower(q.Package.Ecosystem)
		}

		// Check if this dependency has vulnerabilities in our test data
		for _, rel := range m.releases {
			for _, dep := range rel.deps {
				// Match by name (with ecosystem for safety)
				nameMatches := queryName == dep.name
				ecoMatches := queryEco == "" || queryEco == dep.ecosystem

				if nameMatches && ecoMatches && dep.hasVuln {
					// Return the vulnerability IDs
					if len(dep.vulnIDs) > 0 {
						results[i] = dep.vulnIDs
					} else {
						// Default vulnerability ID for testing
						results[i] = []string{fmt.Sprintf("VULN-%s-%s", dep.name, dep.version)}
					}
					break
				}
			}
			if len(results[i]) > 0 {
				break
			}
		}
		// If no match, leave as empty slice (no vulnerabilities)
	}

	return results, nil
}

func (m *mockOSVClient) GetVuln(ctx context.Context, id string) (*clients.OSVVuln, error) {
	return nil, fmt.Errorf("not implemented in mock")
}

// mockRepoClientWithTarball wraps gomock RepoClient and adds ReleaseTarballURL support.
type mockRepoClientWithTarball struct {
	*mockrepo.MockRepoClient
	tarballServer *httptest.Server
}

func (m *mockRepoClientWithTarball) ReleaseTarballURL(tag string) (string, error) {
	// Return URL pointing to our mock tarball server
	return fmt.Sprintf("%s/%s.tar.gz", m.tarballServer.URL, tag), nil
}
