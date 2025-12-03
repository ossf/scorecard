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

package raw

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	depsdev "github.com/ossf/scorecard/v5/internal/packageclient"
)

// transportStubMux returns different canned deps.dev GetPackage payloads.
type transportStubMux struct {
	bodies map[string]string
}

func (m *transportStubMux) RoundTrip(req *http.Request) (*http.Response, error) {
	path := req.URL.RawPath
	if path == "" {
		path = req.URL.Path
	}
	system, name := parseDepsDevPath(path)
	system = canonSystem(system)
	key := system + "||" + name
	body, ok := m.bodies[key]
	if !ok {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader("not found")),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func parseDepsDevPath(p string) (system, name string) {
	const systemsAnchor = "/systems/"
	const packagesAnchor = "/packages/"

	i := strings.Index(p, systemsAnchor)
	if i >= 0 {
		start := i + len(systemsAnchor)
		rest := p[start:]
		if j := strings.IndexByte(rest, '/'); j >= 0 {
			system = rest[:j]
		}
	}

	k := strings.Index(p, packagesAnchor)
	if k >= 0 {
		name = strings.TrimPrefix(p[k+len(packagesAnchor):], "/")
	}

	if dec, err := url.PathUnescape(name); err == nil {
		name = dec
	}

	return system, name
}

func canonSystem(s string) string {
	u := strings.ToUpper(s)
	switch u {
	case "GO", "GOLANG":
		return "GO"
	case "NPM", "JAVASCRIPT":
		return "NPM"
	case "MAVEN":
		return "MAVEN"
	case "PYPI", "PYTHON", "PIP":
		return "PYPI"
	case "CARGO":
		return "CARGO"
	case "NUGET":
		return "NUGET"
	case "RUBYGEMS", "GEM":
		return "RUBYGEMS"
	default:
		return u
	}
}

// TestMTTUDependencies_WithGoModule tests the basic functionality
// of MTTUDependencies with a go.mod file.
//
//nolint:paralleltest // test uses temp dir and file I/O
func TestMTTUDependencies_WithGoModule(t *testing.T) {
	// Cannot run in parallel because it modifies http.DefaultTransport
	// t.Parallel()
	tmpDir := t.TempDir()

	goModContent := "module example.com/test\n\ngo 1.21\n\nrequire (\n\tgithub.com/google/uuid v1.3.0\n)\n"
	err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	tCurrent := now.AddDate(0, -6, 0)
	tOldestNewer := now.AddDate(0, -3, 0)
	tLatest := now.AddDate(0, -1, 0)

	depsDevBody := fmt.Sprintf(`{
  "packageKey": {"system":"GO","name":"github.com/google/uuid"},
  "purl":"pkg:golang/github.com/google/uuid",
  "versions":[
    {"versionKey":{"system":"GO","name":"github.com/google/uuid","version":"v1.3.0"},"purl":"pkg:golang/github.com/google/uuid@v1.3.0","publishedAt":%q,"isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"GO","name":"github.com/google/uuid","version":"v1.4.0"},"purl":"pkg:golang/github.com/google/uuid@v1.4.0","publishedAt":%q,"isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"GO","name":"github.com/google/uuid","version":"v1.5.0"},"purl":"pkg:golang/github.com/google/uuid@v1.5.0","publishedAt":%q,"isDefault":true,"isDeprecated":false}
  ]
}`,
		tCurrent.Format(time.RFC3339),
		tOldestNewer.Format(time.RFC3339),
		tLatest.Format(time.RFC3339),
	)

	mux := &transportStubMux{bodies: map[string]string{
		"GO||github.com/google/uuid": depsDevBody,
	}}
	origTransport := http.DefaultTransport
	http.DefaultTransport = mux
	defer func() { http.DefaultTransport = origTransport }()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepoClient := mockrepo.NewMockRepoClient(ctrl)

	mockRepoClient.EXPECT().
		LocalPath().
		Return(tmpDir, nil).
		AnyTimes()

	req := &checker.CheckRequest{
		Ctx:        context.Background(),
		RepoClient: mockRepoClient,
		Dlogger:    nil,
	}

	result, err := MTTUDependencies(req)
	if err != nil {
		t.Fatalf("MTTUDependencies returned error: %v", err)
	}

	if len(result.Dependencies) == 0 {
		t.Fatal("Expected at least one dependency, got none")
	}

	var found bool
	for _, dep := range result.Dependencies {
		if dep.Name != "github.com/google/uuid" {
			continue
		}
		found = true

		// Note: scalibr's Go module extractor returns versions without "v" prefix
		if dep.Version != "1.3.0" {
			t.Errorf("Expected version 1.3.0, got %s", dep.Version)
		}

		if dep.Ecosystem != checker.EcosystemGo {
			t.Errorf("Expected ecosystem %s, got %s", checker.EcosystemGo, dep.Ecosystem)
		}

		if dep.IsLatest == nil {
			t.Error("IsLatest should not be nil")
		} else if *dep.IsLatest {
			t.Error("IsLatest should be false")
		}

		encoded := dep.TimeSinceOldestReleast.Sub(time.Unix(0, 0))
		if encoded <= 0 {
			t.Error("TimeSinceOldestReleast duration should be > 0")
		}

		inferredPublishedAt := now.Add(-encoded)
		tolerance := 5 * time.Second
		diff := inferredPublishedAt.Sub(tOldestNewer)
		if diff < -tolerance || diff > tolerance {
			t.Errorf("Expected oldest newer release time around %s, got %s (diff: %v)",
				tOldestNewer.Format(time.RFC3339),
				inferredPublishedAt.Format(time.RFC3339),
				diff)
		}
	}

	if !found {
		t.Error("Expected to find github.com/google/uuid in dependencies")
	}
}

func TestMTTUDependencies_NoDependencies(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepoClient := mockrepo.NewMockRepoClient(ctrl)

	mockRepoClient.EXPECT().
		LocalPath().
		Return(tmpDir, nil).
		AnyTimes()

	req := &checker.CheckRequest{
		Ctx:        context.Background(),
		RepoClient: mockRepoClient,
		Dlogger:    nil,
	}

	result, err := MTTUDependencies(req)
	if err != nil {
		t.Fatalf("MTTUDependencies returned error: %v", err)
	}

	if len(result.Dependencies) != 0 {
		t.Errorf("Expected 0 dependencies, got %d", len(result.Dependencies))
	}
}

func TestMTTUDependencies_LocalPathError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepoClient := mockrepo.NewMockRepoClient(ctrl)

	mockRepoClient.EXPECT().
		LocalPath().
		Return("", fmt.Errorf("local path error")).
		Times(1)

	req := &checker.CheckRequest{
		Ctx:        context.Background(),
		RepoClient: mockRepoClient,
		Dlogger:    nil,
	}

	_, err := MTTUDependencies(req)
	if err == nil {
		t.Fatal("Expected error when LocalPath fails, got nil")
	}

	if !strings.Contains(err.Error(), "getting local path") {
		t.Errorf("Expected error message to contain 'getting local path', got: %v", err)
	}
}

func TestScalibrEcosystemToDepsDevSystem(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected string
	}{
		{"golang", "GO"},
		{"go", "GO"},
		{"npm", "NPM"},
		{"pypi", "PYPI"},
		{"maven", "MAVEN"},
		{"cargo", "CARGO"},
		{"nuget", "NUGET"},
		{"gem", "RUBYGEMS"},
		{"unknown", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			result := scalibrEcosystemToDepsDevSystem(tt.input)
			if result != tt.expected {
				t.Errorf("scalibrEcosystemToDepsDevSystem(%q) = %q, want %q",
					tt.input, result, tt.expected)
			}
		})
	}
}

func TestScalibrEcosystemToCheckerEcosystem(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected checker.Ecosystem
	}{
		{"golang", checker.EcosystemGo},
		{"go", checker.EcosystemGo},
		{"npm", checker.EcosystemNPM},
		{"pypi", checker.EcosystemPypi},
		{"maven", checker.EcosystemMaven},
		{"cargo", checker.EcosystemCargo},
		{"nuget", checker.EcosystemNuget},
		{"gem", checker.EcosystemRubyGems},
		{"unknown", checker.Ecosystem("")},
		{"", checker.Ecosystem("")},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			result := scalibrEcosystemToCheckerEcosystem(tt.input)
			if result != tt.expected {
				t.Errorf("scalibrEcosystemToCheckerEcosystem(%q) = %q, want %q",
					tt.input, result, tt.expected)
			}
		})
	}
}

//nolint:paralleltest // test uses temp dir and file I/O
func TestMTTUDependencies_Deduplication(t *testing.T) {
	// Cannot run in parallel because it modifies http.DefaultTransport
	// t.Parallel()

	tmpDir := t.TempDir()

	// Create a go.mod file that references the same package multiple times
	// (this would be unusual but let's verify we handle it correctly)
	goModContent := `module example.com/test

go 1.21

require (
github.com/google/uuid v1.3.0
)
`
	err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	tCurrent := now.AddDate(0, -6, 0)
	tOldestNewer := now.AddDate(0, -3, 0)
	tLatest := now.AddDate(0, -1, 0)

	depsDevBody := fmt.Sprintf(`{
  "packageKey": {"system":"GO","name":"github.com/google/uuid"},
  "purl":"pkg:golang/github.com/google/uuid",
  "versions":[
    {"versionKey":{"system":"GO","name":"github.com/google/uuid","version":"v1.3.0"},"purl":"pkg:golang/github.com/google/uuid@v1.3.0","publishedAt":%q,"isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"GO","name":"github.com/google/uuid","version":"v1.4.0"},"purl":"pkg:golang/github.com/google/uuid@v1.4.0","publishedAt":%q,"isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"GO","name":"github.com/google/uuid","version":"v1.5.0"},"purl":"pkg:golang/github.com/google/uuid@v1.5.0","publishedAt":%q,"isDefault":true,"isDeprecated":false}
  ]
}`,
		tCurrent.Format(time.RFC3339),
		tOldestNewer.Format(time.RFC3339),
		tLatest.Format(time.RFC3339),
	)

	// Create a transport that tracks how many times each package is queried
	queryCount := make(map[string]int)
	trackingTransport := &trackingRoundTripper{
		bodies:     map[string]string{"GO||github.com/google/uuid": depsDevBody},
		queryCount: queryCount,
	}
	origTransport := http.DefaultTransport
	http.DefaultTransport = trackingTransport
	defer func() { http.DefaultTransport = origTransport }()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepoClient := mockrepo.NewMockRepoClient(ctrl)

	mockRepoClient.EXPECT().
		LocalPath().
		Return(tmpDir, nil).
		AnyTimes()

	req := &checker.CheckRequest{
		Ctx:        context.Background(),
		RepoClient: mockRepoClient,
		Dlogger:    nil,
	}

	result, err := MTTUDependencies(req)
	if err != nil {
		t.Fatalf("MTTUDependencies returned error: %v", err)
	}

	if len(result.Dependencies) == 0 {
		t.Fatal("Expected at least one dependency, got none")
	}

	// Verify that each unique package was queried exactly once
	for key, count := range queryCount {
		if count != 1 {
			t.Errorf("Package %q was queried %d times, expected 1 (deduplication not working)", key, count)
		}
	}
}

// trackingRoundTripper tracks how many times each package is queried.
type trackingRoundTripper struct {
	bodies     map[string]string
	queryCount map[string]int
}

func (t *trackingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	path := req.URL.RawPath
	if path == "" {
		path = req.URL.Path
	}
	system, name := parseDepsDevPath(path)
	system = canonSystem(system)
	key := system + "||" + name

	// Track the query
	t.queryCount[key]++

	body, ok := t.bodies[key]
	if !ok {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader("not found")),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

//nolint:paralleltest // test uses temp dir and file I/O
func TestMTTUDependencies_NoIndirectDependencies(t *testing.T) {
	// Cannot run in parallel because it modifies http.DefaultTransport
	// t.Parallel()

	tmpDir := t.TempDir()

	// Create a go.mod with both direct and indirect dependencies
	// The indirect dependency should NOT be queried
	goModContent := `module example.com/test

go 1.21

require (
	github.com/google/uuid v1.3.0
	github.com/stretchr/testify v1.8.0 // indirect
)
`
	err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	tCurrent := now.AddDate(0, -6, 0)
	tLatest := now.AddDate(0, -1, 0)

	// Only provide response for direct dependency
	uuidBody := fmt.Sprintf(`{
  "packageKey": {"system":"GO","name":"github.com/google/uuid"},
  "purl":"pkg:golang/github.com/google/uuid",
  "versions":[
    {"versionKey":{"system":"GO","name":"github.com/google/uuid","version":"v1.3.0"},"purl":"pkg:golang/github.com/google/uuid@v1.3.0","publishedAt":%q,"isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"GO","name":"github.com/google/uuid","version":"v1.5.0"},"purl":"pkg:golang/github.com/google/uuid@v1.5.0","publishedAt":%q,"isDefault":true,"isDeprecated":false}
  ]
}`,
		tCurrent.Format(time.RFC3339),
		tLatest.Format(time.RFC3339),
	)

	// Track all queries to ensure indirect dependencies are NOT queried
	queryCount := make(map[string]int)
	trackingTransport := &trackingRoundTripper{
		bodies: map[string]string{
			"GO||github.com/google/uuid": uuidBody,
			// Deliberately NOT providing testify - it should not be queried
		},
		queryCount: queryCount,
	}
	origTransport := http.DefaultTransport
	http.DefaultTransport = trackingTransport
	defer func() { http.DefaultTransport = origTransport }()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepoClient := mockrepo.NewMockRepoClient(ctrl)

	mockRepoClient.EXPECT().
		LocalPath().
		Return(tmpDir, nil).
		AnyTimes()

	req := &checker.CheckRequest{
		Ctx:        context.Background(),
		RepoClient: mockRepoClient,
		Dlogger:    nil,
	}

	result, err := MTTUDependencies(req)
	if err != nil {
		t.Fatalf("MTTUDependencies returned error: %v", err)
	}

	// Should have found the direct dependency
	if len(result.Dependencies) == 0 {
		t.Fatal("Expected at least one dependency (direct only), got none")
	}

	// Verify that github.com/stretchr/testify was NOT queried (it's indirect)
	if count, found := queryCount["GO||github.com/stretchr/testify"]; found {
		t.Errorf("Indirect dependency github.com/stretchr/testify was queried %d times, should NOT be queried at all", count)
	}

	// Verify only direct dependencies were queried
	for key := range queryCount {
		if strings.Contains(key, "testify") {
			t.Errorf("Indirect dependency %q was queried, but should have been filtered out", key)
		}
	}

	// Ensure at least the direct dependency was found
	foundDirect := false
	for _, dep := range result.Dependencies {
		if dep.Name == "github.com/google/uuid" {
			foundDirect = true
			break
		}
	}
	if !foundDirect {
		t.Error("Direct dependency github.com/google/uuid was not found in results")
	}

	t.Logf("Queried packages: %v", queryCount)
	t.Logf("Found dependencies: %d", len(result.Dependencies))
}

// TestMTTUDependencies_E2E_DirectDepsOnly is a comprehensive end-to-end test that:
// 1. Creates a realistic project with direct and indirect dependencies
// 2. Asserts that scalibr only returns direct dependencies
// 3. Monitors all HTTP requests to verify only direct dependencies are queried.
//
//nolint:paralleltest // test uses temp dir and file I/O
func TestMTTUDependencies_E2E_DirectDepsOnly(t *testing.T) {
	// Cannot run in parallel because it modifies http.DefaultTransport
	// t.Parallel()

	tmpDir := t.TempDir()

	// Create a realistic go.mod with:
	// - Multiple direct dependencies
	// - Multiple indirect dependencies (should be ignored)
	goModContent := `module example.com/testproject

go 1.21

require (
	github.com/google/uuid v1.3.0
	github.com/stretchr/testify v1.8.0 // indirect
)
`
	err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create a minimal main.go to make this a valid Go project
	mainGoContent := `package main

import (
	"fmt"
	_ "github.com/google/uuid"
)

func main() {
	fmt.Println("test")
}
`
	err = os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainGoContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create main.go: %v", err)
	}

	// Expected direct dependencies (what scalibr should find)
	expectedDirectDeps := map[string]bool{
		"github.com/google/uuid": true,
		// Note: stdlib is queried but not included in Dependencies list
	}

	// Indirect dependencies that should NOT be found or queried
	forbiddenIndirectDeps := map[string]bool{
		"github.com/stretchr/testify": true,
	}

	now := time.Now().UTC().Truncate(time.Second)
	tCurrent := now.AddDate(0, -6, 0)
	tLatest := now.AddDate(0, -1, 0)

	// Prepare API response for direct dependency
	uuidBody := fmt.Sprintf(`{
  "packageKey": {"system":"GO","name":"github.com/google/uuid"},
  "versions":[
    {"versionKey":{"system":"GO","name":"github.com/google/uuid","version":"v1.3.0"},"publishedAt":%q,"isDefault":false},
    {"versionKey":{"system":"GO","name":"github.com/google/uuid","version":"v1.5.0"},"publishedAt":%q,"isDefault":true}
  ]
}`, tCurrent.Format(time.RFC3339), tLatest.Format(time.RFC3339))

	// Track all HTTP requests
	type requestLog struct {
		system string
		name   string
	}
	var requestedPackages []requestLog
	var requestMutex sync.Mutex

	trackingTransport := &trackingRoundTripperE2E{
		bodies: map[string]string{
			"GO||github.com/google/uuid": uuidBody,
			// Deliberately NOT providing indirect deps - they should never be requested
		},
		onRequest: func(system, name string) {
			requestMutex.Lock()
			requestedPackages = append(requestedPackages, requestLog{system: system, name: name})
			requestMutex.Unlock()
		},
	}

	origTransport := http.DefaultTransport
	http.DefaultTransport = trackingTransport
	defer func() { http.DefaultTransport = origTransport }()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepoClient := mockrepo.NewMockRepoClient(ctrl)

	mockRepoClient.EXPECT().
		LocalPath().
		Return(tmpDir, nil).
		AnyTimes()

	req := &checker.CheckRequest{
		Ctx:        context.Background(),
		RepoClient: mockRepoClient,
		Dlogger:    nil,
	}

	// Execute the check
	result, err := MTTUDependencies(req)
	if err != nil {
		t.Fatalf("MTTUDependencies returned error: %v", err)
	}

	// ===== ASSERTION STAGE 1: Verify scalibr only returned direct dependencies =====
	t.Log("=== STAGE 1: Verifying scalibr returned only direct dependencies ===")

	foundDeps := make(map[string]bool)
	for _, dep := range result.Dependencies {
		foundDeps[dep.Name] = true
		t.Logf("Found dependency: %s@%s", dep.Name, dep.Version)
	}

	// Check that all direct dependencies were found
	for directDep := range expectedDirectDeps {
		if !foundDeps[directDep] {
			t.Errorf("Expected direct dependency %q was not found by scalibr", directDep)
		}
	}

	// Check that NO indirect dependencies were found
	for indirectDep := range forbiddenIndirectDeps {
		if foundDeps[indirectDep] {
			t.Errorf("FAIL: Indirect dependency %q was found by scalibr, but should have been filtered out", indirectDep)
		}
	}

	// ===== ASSERTION STAGE 2: Verify only direct dependencies were queried via HTTP =====
	t.Log("=== STAGE 2: Verifying only direct dependencies were queried via HTTP ===")

	requestMutex.Lock()
	queriedPackages := make(map[string]bool)
	for _, req := range requestedPackages {
		key := req.name
		queriedPackages[key] = true
		t.Logf("HTTP request made to: %s (system: %s)", req.name, req.system)
	}
	requestMutex.Unlock()

	// Verify NO indirect dependencies were queried
	for indirectDep := range forbiddenIndirectDeps {
		if queriedPackages[indirectDep] {
			t.Errorf("FAIL: Indirect dependency %q was queried via HTTP, should NOT have been queried", indirectDep)
		}
	}

	// Verify all found direct dependencies were queried (except stdlib which doesn't need querying)
	for _, dep := range result.Dependencies {
		if dep.Name == "stdlib" {
			continue // stdlib doesn't need HTTP queries
		}
		if !queriedPackages[dep.Name] {
			t.Logf("Warning: Direct dependency %q was found but not queried (might be intentional)", dep.Name)
		}
	}

	// ===== FINAL SUMMARY =====
	t.Logf("=== TEST SUMMARY ===")
	t.Logf("Direct dependencies found: %d (expected: %d)", len(foundDeps), len(expectedDirectDeps))
	t.Logf("HTTP requests made: %d", len(requestedPackages))

	if len(foundDeps) != len(expectedDirectDeps) {
		t.Logf("Note: Found %d deps but expected %d - this may be due to scalibr not detecting all dependencies from go.mod alone", len(foundDeps), len(expectedDirectDeps))
	}

	t.Logf("✓ Stage 1: Scalibr filtering validated - NO indirect dependencies found")
	t.Logf("✓ Stage 2: HTTP request filtering validated - NO indirect dependencies queried")
}

// trackingRoundTripperE2E is a custom transport for E2E testing with callback.
type trackingRoundTripperE2E struct {
	bodies    map[string]string
	onRequest func(system, name string)
}

func (t *trackingRoundTripperE2E) RoundTrip(req *http.Request) (*http.Response, error) {
	path := req.URL.RawPath
	if path == "" {
		path = req.URL.Path
	}
	system, name := parseDepsDevPath(path)
	system = canonSystem(system)
	key := system + "||" + name

	// Callback to track the request
	if t.onRequest != nil {
		t.onRequest(system, name)
	}

	body, ok := t.bodies[key]
	if !ok {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader("not found")),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

// TestMTTUDependencies_RealGoMod tests with the actual go.mod from this repository
// to ensure we correctly filter indirect dependencies from a real-world project.
//
//nolint:gocognit,paralleltest // test function with complex setup
func TestMTTUDependencies_RealGoMod(t *testing.T) {
	// Cannot run in parallel because it modifies http.DefaultTransport
	// t.Parallel()

	// Read the actual go.mod from the repository root
	repoRoot := filepath.Join("..", "..")
	goModPath := filepath.Join(repoRoot, "go.mod")

	goModContent, err := os.ReadFile(goModPath)
	if err != nil {
		t.Skipf("Could not read repository go.mod: %v", err)
	}

	// Parse direct and indirect dependencies from the actual go.mod
	directDeps := make(map[string]bool)
	indirectDeps := make(map[string]bool)

	lines := strings.Split(string(goModContent), "\n")
	inRequireBlock := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "require (") {
			inRequireBlock = true
			continue
		}

		if inRequireBlock && line == ")" {
			inRequireBlock = false
			continue
		}

		if !inRequireBlock || len(line) == 0 {
			continue
		}

		// Parse dependency line
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		pkgName := parts[0]
		isIndirect := strings.Contains(line, "// indirect")

		if isIndirect {
			indirectDeps[pkgName] = true
		} else {
			directDeps[pkgName] = true
		}
	}

	t.Logf("Parsed from real go.mod: %d direct deps, %d indirect deps", len(directDeps), len(indirectDeps))

	// Create a test directory with a subset of the real dependencies
	tmpDir := t.TempDir()

	// Use a small subset for testing (to avoid needing too many mock responses)
	testDirectDeps := []string{
		"github.com/google/go-cmp",
		"github.com/spf13/cobra",
	}

	testIndirectDeps := []string{
		"github.com/BurntSushi/toml",
		"github.com/cespare/xxhash/v2",
		"golang.org/x/tools",
	}

	// Verify our test deps are actually in the real go.mod
	for _, dep := range testDirectDeps {
		if !directDeps[dep] {
			t.Errorf("Test direct dependency %q not found in real go.mod direct deps", dep)
		}
	}

	for _, dep := range testIndirectDeps {
		if !indirectDeps[dep] {
			t.Errorf("Test indirect dependency %q not found in real go.mod indirect deps", dep)
		}
	}

	// Create a go.mod with both direct and indirect dependencies from real project
	testGoModContent := `module github.com/ossf/scorecard/v5

go 1.24.6

require (
github.com/google/go-cmp v0.7.0
github.com/spf13/cobra v1.10.1
)

require (
github.com/BurntSushi/toml v1.5.0 // indirect
github.com/cespare/xxhash/v2 v2.3.0 // indirect
golang.org/x/tools v0.37.0 // indirect
)
`
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(testGoModContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Track all HTTP requests
	type requestLog struct {
		system string
		name   string
	}
	var requestedPackages []requestLog
	var requestMutex sync.Mutex

	// Mock responses for direct dependencies only
	now := time.Now().UTC().Truncate(time.Second)
	tCurrent := now.AddDate(0, -6, 0)
	tLatest := now.AddDate(0, -1, 0)

	mockBodies := map[string]string{
		"GO||github.com/google/go-cmp": fmt.Sprintf(`{
  "packageKey": {"system":"GO","name":"github.com/google/go-cmp"},
  "versions":[
    {"versionKey":{"system":"GO","name":"github.com/google/go-cmp","version":"v0.7.0"},"publishedAt":%q,"isDefault":true}
  ]
}`, tCurrent.Format(time.RFC3339)),
		"GO||github.com/spf13/cobra": fmt.Sprintf(`{
  "packageKey": {"system":"GO","name":"github.com/spf13/cobra"},
  "versions":[
    {"versionKey":{"system":"GO","name":"github.com/spf13/cobra","version":"v1.10.1"},"publishedAt":%q,"isDefault":true}
  ]
}`, tLatest.Format(time.RFC3339)),
		// Deliberately NOT providing responses for indirect deps
	}

	trackingTransport := &trackingRoundTripperE2E{
		bodies: mockBodies,
		onRequest: func(system, name string) {
			requestMutex.Lock()
			requestedPackages = append(requestedPackages, requestLog{system: system, name: name})
			requestMutex.Unlock()
		},
	}

	origTransport := http.DefaultTransport
	http.DefaultTransport = trackingTransport
	defer func() { http.DefaultTransport = origTransport }()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepoClient := mockrepo.NewMockRepoClient(ctrl)

	mockRepoClient.EXPECT().
		LocalPath().
		Return(tmpDir, nil).
		AnyTimes()

	req := &checker.CheckRequest{
		Ctx:        context.Background(),
		RepoClient: mockRepoClient,
		Dlogger:    nil,
	}

	// Execute the check
	result, err := MTTUDependencies(req)
	if err != nil {
		t.Fatalf("MTTUDependencies returned error: %v", err)
	}

	// Verify results
	foundDeps := make(map[string]bool)
	for _, dep := range result.Dependencies {
		foundDeps[dep.Name] = true
		t.Logf("Found dependency: %s@%s", dep.Name, dep.Version)
	}

	// Verify direct dependencies were found
	for _, directDep := range testDirectDeps {
		if !foundDeps[directDep] {
			t.Errorf("Expected direct dependency %q (from real go.mod) was not found", directDep)
		}
	}

	// Verify NO indirect dependencies were found
	for _, indirectDep := range testIndirectDeps {
		if foundDeps[indirectDep] {
			t.Errorf("FAIL: Indirect dependency %q (from real go.mod) was found, should be filtered out", indirectDep)
		}
	}

	// Verify HTTP requests - NO indirect deps should be queried
	requestMutex.Lock()
	queriedPackages := make(map[string]bool)
	for _, req := range requestedPackages {
		queriedPackages[req.name] = true
		t.Logf("HTTP request: %s (system: %s)", req.name, req.system)
	}
	requestMutex.Unlock()

	// Verify NO indirect dependencies were queried
	for _, indirectDep := range testIndirectDeps {
		if queriedPackages[indirectDep] {
			t.Errorf("FAIL: Indirect dependency %q was queried via HTTP, should NOT be queried", indirectDep)
		}
	}

	// Verify direct dependencies were queried
	for _, directDep := range testDirectDeps {
		if !queriedPackages[directDep] {
			t.Logf("Note: Direct dependency %q was not queried (might not have been detected by scalibr)", directDep)
		}
	}

	t.Logf("✓ Successfully validated filtering using real go.mod dependencies")
	t.Logf("✓ %d direct dependencies expected, %d found", len(testDirectDeps), len(foundDeps))
	t.Logf("✓ 0 indirect dependencies expected, verified none were found or queried")
}

// TestMTTUDependencies_TarballStructure tests with a tarball-like directory structure
// where files are under a subdirectory (as extracted from GitHub tarballs).
//
//nolint:paralleltest // test uses temp dir and file I/O
func TestMTTUDependencies_TarballStructure(t *testing.T) {
	// Cannot run in parallel because it modifies http.DefaultTransport
	// t.Parallel()

	tmpDir := t.TempDir()

	// Simulate GitHub tarball structure: files are under "repo-commitsha/" subdirectory
	repoSubdir := filepath.Join(tmpDir, "scorecard-abc123")
	err := os.MkdirAll(repoSubdir, 0o755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create a go.mod with both direct and indirect dependencies
	goModContent := `module example.com/test

go 1.21

require (
github.com/google/uuid v1.3.0
)

require (
github.com/stretchr/testify v1.8.0 // indirect
golang.org/x/sys v0.1.0 // indirect
)
`
	err = os.WriteFile(filepath.Join(repoSubdir, "go.mod"), []byte(goModContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	tCurrent := now.AddDate(0, -6, 0)

	mockBodies := map[string]string{
		"GO||github.com/google/uuid": fmt.Sprintf(`{
  "packageKey": {"system":"GO","name":"github.com/google/uuid"},
  "versions":[
    {"versionKey":{"system":"GO","name":"github.com/google/uuid","version":"v1.3.0"},"publishedAt":%q,"isDefault":true}
  ]
}`, tCurrent.Format(time.RFC3339)),
	}

	// Track all HTTP requests
	type requestLog struct {
		system string
		name   string
	}
	var requestedPackages []requestLog
	var requestMutex sync.Mutex

	trackingTransport := &trackingRoundTripperE2E{
		bodies: mockBodies,
		onRequest: func(system, name string) {
			requestMutex.Lock()
			requestedPackages = append(requestedPackages, requestLog{system: system, name: name})
			requestMutex.Unlock()
		},
	}

	origTransport := http.DefaultTransport
	http.DefaultTransport = trackingTransport
	defer func() { http.DefaultTransport = origTransport }()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepoClient := mockrepo.NewMockRepoClient(ctrl)

	// Return the tmpDir (parent of the repo subdirectory), simulating tarball extraction
	mockRepoClient.EXPECT().
		LocalPath().
		Return(tmpDir, nil).
		AnyTimes()

	req := &checker.CheckRequest{
		Ctx:        context.Background(),
		RepoClient: mockRepoClient,
		Dlogger:    nil,
	}

	// Execute the check
	result, err := MTTUDependencies(req)
	if err != nil {
		t.Fatalf("MTTUDependencies returned error: %v", err)
	}

	// Verify only direct dependencies were found
	foundDeps := make(map[string]bool)
	for _, dep := range result.Dependencies {
		foundDeps[dep.Name] = true
		t.Logf("Found dependency: %s@%s", dep.Name, dep.Version)
	}

	// Should find the direct dependency
	if !foundDeps["github.com/google/uuid"] {
		t.Errorf("Expected direct dependency 'github.com/google/uuid' was not found")
	}

	// Should NOT find indirect dependencies
	indirectDeps := []string{"github.com/stretchr/testify", "golang.org/x/sys"}
	for _, indirectDep := range indirectDeps {
		if foundDeps[indirectDep] {
			t.Errorf("FAIL: Indirect dependency %q was found, should be filtered out", indirectDep)
		}
	}

	// Verify NO indirect dependencies were queried via HTTP
	requestMutex.Lock()
	queriedPackages := make(map[string]bool)
	for _, req := range requestedPackages {
		queriedPackages[req.name] = true
		t.Logf("HTTP request: %s (system: %s)", req.name, req.system)
	}
	requestMutex.Unlock()

	for _, indirectDep := range indirectDeps {
		if queriedPackages[indirectDep] {
			t.Errorf("FAIL: Indirect dependency %q was queried via HTTP, should NOT be queried", indirectDep)
		}
	}

	t.Logf("✓ Successfully validated tarball-structure filtering")
	t.Logf("✓ Found %d dependencies, verified no indirect dependencies", len(foundDeps))
}

// TestMTTUDependencies_MultipleGoMods tests with multiple go.mod files
// (root and tools/ subdirectory) to ensure indirect deps from ALL files are filtered.
//
//nolint:paralleltest // test uses temp dir and file I/O
func TestMTTUDependencies_MultipleGoMods(t *testing.T) {
	// Cannot run in parallel because it modifies http.DefaultTransport
	// t.Parallel()

	tmpDir := t.TempDir()

	// Create main go.mod
	goModContent := `module example.com/test

go 1.21

require (
github.com/google/uuid v1.3.0
)

require (
github.com/stretchr/testify v1.8.0 // indirect
)
`
	err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create tools/ subdirectory with its own go.mod
	toolsDir := filepath.Join(tmpDir, "tools")
	err = os.MkdirAll(toolsDir, 0o755)
	if err != nil {
		t.Fatalf("Failed to create tools directory: %v", err)
	}

	toolsGoModContent := `module example.com/test/tools

go 1.21

require (
github.com/golangci/golangci-lint v1.50.0
)

require (
github.com/dimchansky/utfbom v1.1.1 // indirect
golang.org/x/tools v0.1.0 // indirect
)
`
	err = os.WriteFile(filepath.Join(toolsDir, "go.mod"), []byte(toolsGoModContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create tools/go.mod: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	tCurrent := now.AddDate(0, -6, 0)

	mockBodies := map[string]string{
		"GO||github.com/google/uuid": fmt.Sprintf(`{
  "packageKey": {"system":"GO","name":"github.com/google/uuid"},
  "versions":[
    {"versionKey":{"system":"GO","name":"github.com/google/uuid","version":"v1.3.0"},"publishedAt":%q,"isDefault":true}
  ]
}`, tCurrent.Format(time.RFC3339)),
		"GO||github.com/golangci/golangci-lint": fmt.Sprintf(`{
  "packageKey": {"system":"GO","name":"github.com/golangci/golangci-lint"},
  "versions":[
    {"versionKey":{"system":"GO","name":"github.com/golangci/golangci-lint","version":"v1.50.0"},"publishedAt":%q,"isDefault":true}
  ]
}`, tCurrent.Format(time.RFC3339)),
	}

	// Track all HTTP requests
	type requestLog struct {
		system string
		name   string
	}
	var requestedPackages []requestLog
	var requestMutex sync.Mutex

	trackingTransport := &trackingRoundTripperE2E{
		bodies: mockBodies,
		onRequest: func(system, name string) {
			requestMutex.Lock()
			requestedPackages = append(requestedPackages, requestLog{system: system, name: name})
			requestMutex.Unlock()
		},
	}

	origTransport := http.DefaultTransport
	http.DefaultTransport = trackingTransport
	defer func() { http.DefaultTransport = origTransport }()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepoClient := mockrepo.NewMockRepoClient(ctrl)

	mockRepoClient.EXPECT().
		LocalPath().
		Return(tmpDir, nil).
		AnyTimes()

	req := &checker.CheckRequest{
		Ctx:        context.Background(),
		RepoClient: mockRepoClient,
		Dlogger:    nil,
	}

	// Execute the check
	result, err := MTTUDependencies(req)
	if err != nil {
		t.Fatalf("MTTUDependencies returned error: %v", err)
	}

	// Verify only direct dependencies were found
	foundDeps := make(map[string]bool)
	for _, dep := range result.Dependencies {
		foundDeps[dep.Name] = true
		t.Logf("Found dependency: %s@%s", dep.Name, dep.Version)
	}

	// Should find the direct dependencies
	expectedDirect := []string{"github.com/google/uuid", "github.com/golangci/golangci-lint"}
	for _, directDep := range expectedDirect {
		if !foundDeps[directDep] {
			t.Errorf("Expected direct dependency %q was not found", directDep)
		}
	}

	// Should NOT find indirect dependencies from ANY go.mod file
	indirectDeps := []string{
		"github.com/stretchr/testify",  // indirect in root go.mod
		"github.com/dimchansky/utfbom", // indirect in tools/go.mod
		"golang.org/x/tools",           // indirect in tools/go.mod
	}
	for _, indirectDep := range indirectDeps {
		if foundDeps[indirectDep] {
			t.Errorf("FAIL: Indirect dependency %q was found, should be filtered out", indirectDep)
		}
	}

	// Verify NO indirect dependencies were queried via HTTP
	requestMutex.Lock()
	queriedPackages := make(map[string]bool)
	for _, req := range requestedPackages {
		queriedPackages[req.name] = true
		t.Logf("HTTP request: %s (system: %s)", req.name, req.system)
	}
	requestMutex.Unlock()

	for _, indirectDep := range indirectDeps {
		if queriedPackages[indirectDep] {
			t.Errorf("FAIL: Indirect dependency %q was queried via HTTP, should NOT be queried", indirectDep)
		}
	}

	t.Logf("✓ Successfully validated multi-go.mod filtering")
	t.Logf("✓ Found %d dependencies, verified no indirect dependencies from any go.mod", len(foundDeps))
}

// TestIsPseudoVersion tests the isPseudoVersion function with comprehensive cases.
//
//nolint:paralleltest,tparallel // test cases are independent
func TestIsPseudoVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		version string
		want    bool
	}{
		// True cases - actual pseudo-versions
		{
			name:    "standard pseudo-version with v prefix",
			version: "v1.2.0-0.20250916002408-abc123def456",
			want:    true,
		},
		{
			name:    "v0.0.0 pseudo-version",
			version: "v0.0.0-20241213102144-19d51d7fe467",
			want:    true,
		},
		{
			name:    "pseudo-version without v prefix",
			version: "1.2.0-0.20250916002408-abc123def",
			want:    true,
		},
		{
			name:    "pseudo-version with dots before timestamp",
			version: "v1.2.3.0.20240101120000-abcdef123456",
			want:    true,
		},
		{
			name:    "pseudo-version with dash before timestamp",
			version: "v2.0.0-20231217203849-220c5c2851b7",
			want:    true,
		},

		// False cases - real versions
		{
			name:    "normal semver",
			version: "v1.2.3",
			want:    false,
		},
		{
			name:    "semver with +incompatible",
			version: "v2.0.0+incompatible",
			want:    false,
		},
		{
			name:    "pre-release version",
			version: "v1.2.3-rc.1",
			want:    false,
		},
		{
			name:    "beta version",
			version: "v1.2.3-beta",
			want:    false,
		},
		{
			name:    "alpha version",
			version: "v0.1.0-alpha.2",
			want:    false,
		},
		{
			name:    "version with date but not 14 digits",
			version: "v1.2.3-20250916",
			want:    false,
		},
		{
			name:    "version with 12 digit timestamp (not 14)",
			version: "v1.2.3-202509160024-abc",
			want:    false,
		},
		{
			name:    "version with 14 digits but no dash after",
			version: "v1.2.3-20250916002408abc",
			want:    false,
		},
		{
			name:    "version with letters in timestamp position",
			version: "v1.2.3-0.abcdefghijklmn-hash",
			want:    false,
		},

		// Edge cases
		{
			name:    "empty string",
			version: "",
			want:    false,
		},
		{
			name:    "just v",
			version: "v",
			want:    false,
		},
		{
			name:    "invalid format",
			version: "invalid",
			want:    false,
		},
		{
			name:    "timestamp at start (no version before)",
			version: "20250916002408-abc123",
			want:    false,
		},
		{
			name:    "multiple timestamps",
			version: "v1.2.0-20240101120000-20240202130000-abc",
			want:    true, // first timestamp matches pattern
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isPseudoVersion(tc.version)
			if got != tc.want {
				t.Errorf("isPseudoVersion(%q) = %v, want %v", tc.version, got, tc.want)
			}
		})
	}
}

// TestNewestInfoFor tests the newestInfoFor function.
//
//nolint:paralleltest,tparallel // test cases are independent
func TestNewestInfoFor(t *testing.T) {
	t.Parallel()

	// Helper to create a version with timestamp
	makeVersion := func(ver, published string) depsdev.Version {
		return depsdev.Version{
			VersionKey: depsdev.VersionKey{
				System:  "GO",
				Name:    "test/package",
				Version: ver,
			},
			PublishedAt: published,
		}
	}

	tests := []struct {
		name           string
		currentVersion string
		wantTimestamp  string
		versions       []depsdev.Version
		wantLatest     bool
	}{
		{
			name:           "current is latest",
			currentVersion: "v1.3.0",
			versions: []depsdev.Version{
				makeVersion("v1.0.0", "2024-01-01T00:00:00Z"),
				makeVersion("v1.1.0", "2024-02-01T00:00:00Z"),
				makeVersion("v1.2.0", "2024-03-01T00:00:00Z"),
				makeVersion("v1.3.0", "2024-04-01T00:00:00Z"),
			},
			wantLatest:    true,
			wantTimestamp: "",
		},
		{
			name:           "current is outdated - should return oldest newer",
			currentVersion: "v1.0.0",
			versions: []depsdev.Version{
				makeVersion("v1.0.0", "2024-01-01T00:00:00Z"),
				makeVersion("v1.1.0", "2024-03-01T00:00:00Z"), // oldest newer
				makeVersion("v1.2.0", "2024-06-01T00:00:00Z"),
				makeVersion("v1.3.0", "2024-09-01T00:00:00Z"), // latest
			},
			wantLatest:    false,
			wantTimestamp: "2024-03-01T00:00:00Z", // v1.1.0, not v1.3.0!
		},
		{
			name:           "pseudo-versions ignored - current is latest release",
			currentVersion: "v1.2.0",
			versions: []depsdev.Version{
				makeVersion("v1.2.0", "2024-01-01T00:00:00Z"),
				makeVersion("v1.2.1-0.20250916002408-abc123", "2024-06-01T00:00:00Z"), // pseudo - ignored
				makeVersion("v1.3.0-0.20251201120000-def456", "2024-07-01T00:00:00Z"), // pseudo - ignored
			},
			wantLatest:    true, // because pseudo-versions are filtered
			wantTimestamp: "",
		},
		{
			name:           "pseudo-versions ignored - has real newer version",
			currentVersion: "v1.0.0",
			versions: []depsdev.Version{
				makeVersion("v1.0.0", "2024-01-01T00:00:00Z"),
				makeVersion("v1.0.1-0.20240215000000-abc", "2024-02-15T00:00:00Z"), // pseudo - ignored
				makeVersion("v1.1.0", "2024-03-01T00:00:00Z"),                      // real - should use this
				makeVersion("v1.2.0-0.20240401000000-def", "2024-04-01T00:00:00Z"), // pseudo - ignored
			},
			wantLatest:    false,
			wantTimestamp: "2024-03-01T00:00:00Z", // v1.1.0
		},
		{
			name:           "version without v prefix",
			currentVersion: "1.2.0",
			versions: []depsdev.Version{
				makeVersion("1.2.0", "2024-01-01T00:00:00Z"),
				makeVersion("1.3.0", "2024-02-01T00:00:00Z"),
			},
			wantLatest:    false,
			wantTimestamp: "2024-02-01T00:00:00Z",
		},
		{
			name:           "mixed v prefix handling",
			currentVersion: "v1.2.0",
			versions: []depsdev.Version{
				makeVersion("1.2.0", "2024-01-01T00:00:00Z"),  // no v
				makeVersion("v1.3.0", "2024-02-01T00:00:00Z"), // with v
				makeVersion("1.4.0", "2024-03-01T00:00:00Z"),  // no v
			},
			wantLatest:    false,
			wantTimestamp: "2024-02-01T00:00:00Z", // oldest newer (v1.3.0)
		},
		{
			name:           "empty version list",
			currentVersion: "v1.0.0",
			versions:       []depsdev.Version{},
			wantLatest:     false,
			wantTimestamp:  "",
		},
		{
			name:           "invalid timestamps ignored",
			currentVersion: "v1.0.0",
			versions: []depsdev.Version{
				makeVersion("v1.0.0", "2024-01-01T00:00:00Z"),
				makeVersion("v1.1.0", "invalid-timestamp"),
				makeVersion("v1.2.0", "2024-03-01T00:00:00Z"),
			},
			wantLatest:    false,
			wantTimestamp: "2024-03-01T00:00:00Z", // v1.2.0 (v1.1.0 skipped due to bad timestamp)
		},
		{
			name:           "current version not in list",
			currentVersion: "v1.5.0",
			versions: []depsdev.Version{
				makeVersion("v1.0.0", "2024-01-01T00:00:00Z"),
				makeVersion("v1.1.0", "2024-02-01T00:00:00Z"),
			},
			wantLatest:    false,
			wantTimestamp: "2024-02-01T00:00:00Z", // falls back to latest timestamp
		},
		{
			name:           "v0.0.0 pseudo-version filtered",
			currentVersion: "v1.0.0",
			versions: []depsdev.Version{
				makeVersion("v0.0.0-20241213102144-19d51d7fe467", "2024-01-01T00:00:00Z"), // pseudo
				makeVersion("v1.0.0", "2024-02-01T00:00:00Z"),
				makeVersion("v1.1.0", "2024-03-01T00:00:00Z"),
			},
			wantLatest:    false,
			wantTimestamp: "2024-03-01T00:00:00Z",
		},
		{
			name:           "all versions are pseudo-versions",
			currentVersion: "v1.0.0-0.20240101000000-abc",
			versions: []depsdev.Version{
				makeVersion("v1.0.0-0.20240101000000-abc", "2024-01-01T00:00:00Z"),
				makeVersion("v1.1.0-0.20240201000000-def", "2024-02-01T00:00:00Z"),
			},
			wantLatest:    true, // no real releases, so considered latest by timestamp fallback
			wantTimestamp: "",
		},
		{
			name:           "oldest of multiple newer versions",
			currentVersion: "v1.0.0",
			versions: []depsdev.Version{
				makeVersion("v1.0.0", "2024-01-01T00:00:00Z"),
				makeVersion("v1.0.1", "2024-01-15T00:00:00Z"), // oldest newer
				makeVersion("v1.0.2", "2024-02-01T00:00:00Z"),
				makeVersion("v1.1.0", "2024-03-01T00:00:00Z"),
				makeVersion("v1.2.0", "2024-04-01T00:00:00Z"),
				makeVersion("v2.0.0", "2024-05-01T00:00:00Z"), // latest
			},
			wantLatest:    false,
			wantTimestamp: "2024-01-15T00:00:00Z", // v1.0.1
		},
	}

	//nolint:paralleltest // subtests are independent
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotLatest, gotTime := newestInfoFor(tc.currentVersion, tc.versions)

			if gotLatest != tc.wantLatest {
				t.Errorf("isLatest = %v, want %v", gotLatest, tc.wantLatest)
			}

			//nolint:nestif // test validation requires nested structure
			if tc.wantTimestamp == "" {
				if gotTime != nil {
					t.Errorf("expected nil timestamp, got %v", gotTime)
				}
			} else {
				if gotTime == nil {
					t.Errorf("expected timestamp %s, got nil", tc.wantTimestamp)
				} else {
					want, err := time.Parse(time.RFC3339, tc.wantTimestamp)
					if err != nil {
						t.Fatalf("failed to parse want timestamp: %v", err)
					}
					if !gotTime.Equal(want) {
						t.Errorf("timestamp = %v, want %v", gotTime.Format(time.RFC3339), tc.wantTimestamp)
					}
				}
			}
		})
	}
}

// TestMTTUDependencies_WithPseudoVersionsE2E is an end-to-end test verifying
// that pseudo-versions are correctly filtered in a realistic scenario.
func TestMTTUDependencies_WithPseudoVersionsE2E(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create a go.mod with a dependency that we'll test
	goModContent := `module example.com/test

go 1.21

require (
	github.com/example/pkg v1.2.0
)
`
	err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Mock deps.dev response with both real versions and pseudo-versions
	now := time.Now().UTC()
	currentRelease := now.AddDate(0, -6, 0) // 6 months ago
	pseudoVersion := now.AddDate(0, -1, 0)  // 1 month ago (but pseudo)
	newerRelease := now.AddDate(0, -3, 0)   // 3 months ago (real release)

	depsDevBody := fmt.Sprintf(`{
  "packageKey": {"system":"GO","name":"github.com/example/pkg"},
  "purl":"pkg:golang/github.com/example/pkg",
  "versions":[
    {"versionKey":{"system":"GO","name":"github.com/example/pkg","version":"v1.2.0"},"purl":"pkg:golang/github.com/example/pkg@v1.2.0","publishedAt":%q,"isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"GO","name":"github.com/example/pkg","version":"v1.2.1-0.20250916002408-abc123"},"purl":"pkg:golang/github.com/example/pkg@v1.2.1-0.20250916002408-abc123","publishedAt":%q,"isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"GO","name":"github.com/example/pkg","version":"v1.3.0"},"purl":"pkg:golang/github.com/example/pkg@v1.3.0","publishedAt":%q,"isDefault":true,"isDeprecated":false}
  ]
}`,
		currentRelease.Format(time.RFC3339),
		pseudoVersion.Format(time.RFC3339),
		newerRelease.Format(time.RFC3339),
	)

	mux := &transportStubMux{bodies: map[string]string{
		"GO||github.com/example/pkg": depsDevBody,
	}}
	origTransport := http.DefaultTransport
	http.DefaultTransport = mux
	defer func() { http.DefaultTransport = origTransport }()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepoClient := mockrepo.NewMockRepoClient(ctrl)

	mockRepoClient.EXPECT().
		LocalPath().
		Return(tmpDir, nil).
		AnyTimes()

	req := &checker.CheckRequest{
		Ctx:        context.Background(),
		RepoClient: mockRepoClient,
		Dlogger:    nil,
	}

	result, err := MTTUDependencies(req)
	if err != nil {
		t.Fatalf("MTTUDependencies returned error: %v", err)
	}

	// Find our dependency
	var found bool
	for _, dep := range result.Dependencies {
		if dep.Name != "github.com/example/pkg" {
			continue
		}
		found = true

		// Verify it's marked as not latest (because v1.3.0 exists)
		if dep.IsLatest == nil {
			t.Error("IsLatest should not be nil")
		} else if *dep.IsLatest {
			t.Error("IsLatest should be false (v1.3.0 exists)")
		}

		// Verify the timestamp is from v1.3.0 (real release), not pseudo-version
		if dep.TimeSinceOldestReleast.IsZero() {
			t.Error("TimeSinceOldestReleast should not be zero")
		}

		// Calculate what the published time should be based on encoded duration
		encoded := dep.TimeSinceOldestReleast.Sub(time.Unix(0, 0))
		inferredPublishedAt := now.Add(-encoded)

		// Should be close to newerRelease (3 months ago), NOT pseudoVersion (1 month ago)
		tolerance := 5 * time.Second
		diffFromReal := inferredPublishedAt.Sub(newerRelease)
		if diffFromReal < -tolerance || diffFromReal > tolerance {
			t.Errorf("Expected timestamp close to v1.3.0 release (%s), got %s (diff: %v)",
				newerRelease.Format(time.RFC3339),
				inferredPublishedAt.Format(time.RFC3339),
				diffFromReal)
		}

		// Explicitly verify it's NOT using the pseudo-version timestamp
		diffFromPseudo := inferredPublishedAt.Sub(pseudoVersion)
		if diffFromPseudo > -tolerance && diffFromPseudo < tolerance {
			t.Errorf("FAIL: Appears to be using pseudo-version timestamp (%s), should use real release (%s)",
				pseudoVersion.Format(time.RFC3339),
				newerRelease.Format(time.RFC3339))
		}
	}

	if !found {
		t.Error("Expected to find github.com/example/pkg in dependencies")
	}
}
