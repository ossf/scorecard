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
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
)

// TestIsValidPath tests the path traversal protection.
func TestIsValidPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		path  string
		valid bool
	}{
		{
			name:  "normal relative path",
			path:  "normal/path/to/file.txt",
			valid: true,
		},
		{
			name:  "single level path",
			path:  "file.txt",
			valid: true,
		},
		{
			name:  "path with dot directory",
			path:  "./path/to/file.txt",
			valid: true,
		},
		{
			name:  "absolute path - should be rejected",
			path:  "/etc/passwd",
			valid: false,
		},
		{
			name:  "absolute path windows - should be rejected",
			path:  "C:\\Windows\\System32",
			valid: true, // On Linux, this is treated as a relative path
		},
		{
			name:  "path traversal with ..",
			path:  "../etc/passwd",
			valid: false,
		},
		{
			name:  "path traversal in middle",
			path:  "path/../../../etc/passwd",
			valid: false,
		},
		{
			name:  "path traversal hidden",
			path:  "path/to/../../../../../../etc/passwd",
			valid: false,
		},
		{
			name:  "double dot in filename",
			path:  "file..txt",
			valid: false, // Contains ".." so rejected
		},
		{
			name:  "multiple slashes",
			path:  "path//to///file.txt",
			valid: true,
		},
		{
			name:  "empty path",
			path:  "",
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := isValidPath(tt.path)
			if result != tt.valid {
				t.Errorf("isValidPath(%q) = %v, want %v", tt.path, result, tt.valid)
			}
		})
	}
}

// TestToOSVEcosystem tests ecosystem name normalization.
func TestToOSVEcosystem(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Go - golang",
			input:    "golang",
			expected: "Go",
		},
		{
			name:     "Go - gomod",
			input:    "gomod",
			expected: "Go",
		},
		{
			name:     "Go - go",
			input:    "go",
			expected: "Go",
		},
		{
			name:     "Go - case insensitive",
			input:    "GOLANG",
			expected: "Go",
		},
		{
			name:     "npm",
			input:    "npm",
			expected: "npm",
		},
		{
			name:     "npm - node",
			input:    "node",
			expected: "npm",
		},
		{
			name:     "npm - packagejson",
			input:    "packagejson",
			expected: "npm",
		},
		{
			name:     "PyPI - pypi",
			input:    "pypi",
			expected: "PyPI",
		},
		{
			name:     "PyPI - python",
			input:    "python",
			expected: "PyPI",
		},
		{
			name:     "PyPI - requirements",
			input:    "requirements",
			expected: "PyPI",
		},
		{
			name:     "Maven",
			input:    "maven",
			expected: "Maven",
		},
		{
			name:     "Maven - pomxml",
			input:    "pomxml",
			expected: "Maven",
		},
		{
			name:     "Maven - gradle",
			input:    "gradle",
			expected: "Maven",
		},
		{
			name:     "Crates.io - cargo",
			input:    "cargo",
			expected: "Crates.io",
		},
		{
			name:     "Crates.io - rust",
			input:    "rust",
			expected: "Crates.io",
		},
		{
			name:     "NuGet",
			input:    "nuget",
			expected: "NuGet",
		},
		{
			name:     "NuGet - .net",
			input:    ".net",
			expected: "NuGet",
		},
		{
			name:     "RubyGems - gem",
			input:    "gem",
			expected: "RubyGems",
		},
		{
			name:     "RubyGems - ruby",
			input:    "ruby",
			expected: "RubyGems",
		},
		{
			name:     "unknown ecosystem",
			input:    "unknown",
			expected: "unknown",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace trimmed",
			input:    "  golang  ",
			expected: "Go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := toOSVEcosystem(tt.input)
			if result != tt.expected {
				t.Errorf("toOSVEcosystem(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestToOSVQuery tests OSV query construction.
func TestToOSVQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		key          struct{ eco, name, version, purl string }
		expectedPURL string
		expectedEco  string
		expectedName string
		expectedVer  string
		prefersPURL  bool
	}{
		{
			name: "with PURL - prefers PURL",
			key: struct{ eco, name, version, purl string }{
				eco:     "Go",
				name:    "github.com/pkg/errors",
				version: "0.9.1",
				purl:    "pkg:golang/github.com/pkg/errors@0.9.1",
			},
			expectedPURL: "pkg:golang/github.com/pkg/errors@0.9.1",
			prefersPURL:  true,
		},
		{
			name: "without PURL - uses ecosystem/name/version",
			key: struct{ eco, name, version, purl string }{
				eco:     "Go",
				name:    "github.com/pkg/errors",
				version: "0.9.1",
				purl:    "",
			},
			expectedEco:  "Go",
			expectedName: "github.com/pkg/errors",
			expectedVer:  "v0.9.1", // Go gets v prefix
			prefersPURL:  false,
		},
		{
			name: "Go without v prefix - adds it",
			key: struct{ eco, name, version, purl string }{
				eco:     "Go",
				name:    "golang.org/x/text",
				version: "0.3.6",
				purl:    "",
			},
			expectedEco:  "Go",
			expectedName: "golang.org/x/text",
			expectedVer:  "v0.3.6",
			prefersPURL:  false,
		},
		{
			name: "Go with v prefix - keeps it",
			key: struct{ eco, name, version, purl string }{
				eco:     "Go",
				name:    "golang.org/x/text",
				version: "v0.3.6",
				purl:    "",
			},
			expectedEco:  "Go",
			expectedName: "golang.org/x/text",
			expectedVer:  "v0.3.6",
			prefersPURL:  false,
		},
		{
			name: "npm without v prefix - no change",
			key: struct{ eco, name, version, purl string }{
				eco:     "npm",
				name:    "lodash",
				version: "4.17.21",
				purl:    "",
			},
			expectedEco:  "npm",
			expectedName: "lodash",
			expectedVer:  "4.17.21",
			prefersPURL:  false,
		},
		{
			name: "PyPI normalization",
			key: struct{ eco, name, version, purl string }{
				eco:     "python",
				name:    "requests",
				version: "2.28.0",
				purl:    "",
			},
			expectedEco:  "PyPI",
			expectedName: "requests",
			expectedVer:  "2.28.0",
			prefersPURL:  false,
		},
		{
			name: "empty version",
			key: struct{ eco, name, version, purl string }{
				eco:     "Go",
				name:    "github.com/pkg/errors",
				version: "",
				purl:    "",
			},
			expectedEco:  "Go",
			expectedName: "github.com/pkg/errors",
			expectedVer:  "",
			prefersPURL:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := toOSVQuery(tt.key)

			//nolint:nestif // Clear if/else structure for test validation
			if tt.prefersPURL {
				if result.Package.PURL != tt.expectedPURL {
					t.Errorf("toOSVQuery(%+v).Package.PURL = %q, want %q",
						tt.key, result.Package.PURL, tt.expectedPURL)
				}
			} else {
				if result.Package.Ecosystem != tt.expectedEco {
					t.Errorf("toOSVQuery(%+v).Package.Ecosystem = %q, want %q",
						tt.key, result.Package.Ecosystem, tt.expectedEco)
				}
				if result.Package.Name != tt.expectedName {
					t.Errorf("toOSVQuery(%+v).Package.Name = %q, want %q",
						tt.key, result.Package.Name, tt.expectedName)
				}
				if result.Version != tt.expectedVer {
					t.Errorf("toOSVQuery(%+v).Version = %q, want %q",
						tt.key, result.Version, tt.expectedVer)
				}
			}
		})
	}
}

// TestToDirectDeps tests conversion from clients.Dep to checker.DirectDep.
//
//nolint:gocognit // Test function has detailed validation for each dependency field
func TestToDirectDeps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    []clients.Dep
		expected int
	}{
		{
			name:     "empty slice",
			input:    []clients.Dep{},
			expected: 0,
		},
		{
			name: "single dependency",
			input: []clients.Dep{
				{
					Ecosystem: "Go",
					Name:      "github.com/pkg/errors",
					Version:   "0.9.1",
					PURL:      "pkg:golang/github.com/pkg/errors@0.9.1",
					Location:  "go.mod",
				},
			},
			expected: 1,
		},
		{
			name: "multiple dependencies",
			input: []clients.Dep{
				{
					Ecosystem: "Go",
					Name:      "github.com/pkg/errors",
					Version:   "0.9.1",
					Location:  "go.mod",
				},
				{
					Ecosystem: "Go",
					Name:      "golang.org/x/text",
					Version:   "0.3.6",
					Location:  "go.mod",
				},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := toDirectDeps(tt.input)

			if len(result) != tt.expected {
				t.Errorf("toDirectDeps() returned %d deps, want %d", len(result), tt.expected)
			}

			// Verify field mapping
			for i, dep := range result {
				if i >= len(tt.input) {
					break
				}
				orig := tt.input[i]
				if dep.Ecosystem != orig.Ecosystem {
					t.Errorf("dep[%d].Ecosystem = %q, want %q", i, dep.Ecosystem, orig.Ecosystem)
				}
				if dep.Name != orig.Name {
					t.Errorf("dep[%d].Name = %q, want %q", i, dep.Name, orig.Name)
				}
				if dep.Version != orig.Version {
					t.Errorf("dep[%d].Version = %q, want %q", i, dep.Version, orig.Version)
				}
				if dep.PURL != orig.PURL {
					t.Errorf("dep[%d].PURL = %q, want %q", i, dep.PURL, orig.PURL)
				}
				// Check path slash conversion
				if !strings.Contains(dep.Location, "\\") && strings.Contains(orig.Location, "\\") {
					t.Errorf("dep[%d].Location should have forward slashes, got %q", i, dep.Location)
				}
			}
		})
	}
}

// TestSoleSubdir tests finding the sole subdirectory.
func TestSoleSubdir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		setup     func(t *testing.T) string // returns temp dir path
		checkFunc func(t *testing.T, result string, baseDir string)
		name      string
		wantError bool
	}{
		{
			name: "single subdirectory",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				subdir := filepath.Join(dir, "repo-v1.0.0")
				if err := os.Mkdir(subdir, 0o755); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			wantError: false,
			checkFunc: func(t *testing.T, result string, baseDir string) {
				t.Helper()
				if !strings.HasSuffix(result, "repo-v1.0.0") {
					t.Errorf("soleSubdir() = %q, want path ending in 'repo-v1.0.0'", result)
				}
			},
		},
		{
			name: "no subdirectories - returns base dir",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				// Create a file but no directories
				if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("test"), 0o600); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			wantError: false,
			checkFunc: func(t *testing.T, result string, baseDir string) {
				t.Helper()
				if result != baseDir {
					t.Errorf("soleSubdir() = %q, want %q", result, baseDir)
				}
			},
		},
		{
			name: "multiple subdirectories - returns first",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				if err := os.Mkdir(filepath.Join(dir, "subdir1"), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(filepath.Join(dir, "subdir2"), 0o755); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			wantError: false,
			checkFunc: func(t *testing.T, result string, baseDir string) {
				t.Helper()
				// Should return one of the subdirs (implementation returns first found)
				if !strings.Contains(result, "subdir") {
					t.Errorf("soleSubdir() = %q, should contain 'subdir'", result)
				}
			},
		},
		{
			name: "non-existent directory",
			setup: func(t *testing.T) string {
				t.Helper()
				return filepath.Join(t.TempDir(), "nonexistent")
			},
			wantError: true,
			checkFunc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			baseDir := tt.setup(t)

			result, err := soleSubdir(baseDir)

			if tt.wantError {
				if err == nil {
					t.Error("soleSubdir() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("soleSubdir() unexpected error: %v", err)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, result, baseDir)
			}
		})
	}
}

// TestDownloadAndExtractTarball_PathTraversal tests path traversal protection.
func TestDownloadAndExtractTarball_PathTraversal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		errorMsg   string
		tarEntries []struct {
			name string
			body string
		}
		wantError bool
	}{
		{
			name: "normal files",
			tarEntries: []struct {
				name string
				body string
			}{
				{"repo/file1.txt", "content1"},
				{"repo/subdir/file2.txt", "content2"},
			},
			wantError: false,
		},
		{
			name: "path traversal with ..",
			tarEntries: []struct {
				name string
				body string
			}{
				{"../etc/passwd", "malicious"},
			},
			wantError: true,
			errorMsg:  "path traversal",
		},
		{
			name: "absolute path",
			tarEntries: []struct {
				name string
				body string
			}{
				{"/etc/passwd", "malicious"},
			},
			wantError: true,
			errorMsg:  "path traversal",
		},
		{
			name: "hidden traversal",
			tarEntries: []struct {
				name string
				body string
			}{
				{"repo/../../../etc/passwd", "malicious"},
			},
			wantError: true,
			errorMsg:  "path traversal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create test server with malicious tarball
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var buf bytes.Buffer
				gw := gzip.NewWriter(&buf)
				tw := tar.NewWriter(gw)

				for _, entry := range tt.tarEntries {
					hdr := &tar.Header{
						Name: entry.name,
						Mode: 0o600,
						Size: int64(len(entry.body)),
					}
					if err := tw.WriteHeader(hdr); err != nil {
						t.Fatal(err)
					}
					if _, err := tw.Write([]byte(entry.body)); err != nil {
						t.Fatal(err)
					}
				}

				tw.Close()
				gw.Close()
				//nolint:errcheck // Test fixture write
				w.Write(buf.Bytes())
			}))
			defer server.Close()

			dst := t.TempDir()
			ctx := context.Background()
			hc := &http.Client{}

			err := downloadAndExtractTarball(ctx, hc, server.URL, dst)

			if tt.wantError {
				if err == nil {
					t.Error("downloadAndExtractTarball() expected error, got nil")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("downloadAndExtractTarball() error = %v, should contain %q", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("downloadAndExtractTarball() unexpected error: %v", err)
				}
			}
		})
	}
}

// TestDownloadAndExtractTarball_DecompressionBomb tests size limit protection.
//
//nolint:errcheck // Test fixture HTTP writes don't need error checking
func TestDownloadAndExtractTarball_DecompressionBomb(t *testing.T) {
	t.Parallel()

	t.Run("file exceeds max size", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var buf bytes.Buffer
			gw := gzip.NewWriter(&buf)
			tw := tar.NewWriter(gw)

			// Create a tar entry claiming to be larger than maxFileSize
			hdr := &tar.Header{
				Name: "huge.txt",
				Mode: 0o600,
				Size: maxFileSize + 1,
			}
			_ = tw.WriteHeader(hdr)
			// Don't actually write that much data
			_, _ = tw.Write([]byte("small content"))

			tw.Close()
			gw.Close()
			_, _ = w.Write(buf.Bytes())
		}))
		defer server.Close()

		dst := t.TempDir()
		ctx := context.Background()
		hc := &http.Client{}

		err := downloadAndExtractTarball(ctx, hc, server.URL, dst)
		if err == nil {
			t.Error("expected decompression bomb error")
		} else if !strings.Contains(err.Error(), "decompression bomb") {
			t.Errorf("error should mention decompression bomb, got: %v", err)
		}
	})

	t.Run("HTTP error", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("not found"))
		}))
		defer server.Close()

		dst := t.TempDir()
		ctx := context.Background()
		hc := &http.Client{}

		err := downloadAndExtractTarball(ctx, hc, server.URL, dst)
		if err == nil {
			t.Error("expected HTTP error")
		} else if !strings.Contains(err.Error(), "HTTP request failed") {
			t.Errorf("error should mention HTTP request failed, got: %v", err)
		}
	})

	t.Run("invalid gzip", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("not a gzip file"))
		}))
		defer server.Close()

		dst := t.TempDir()
		ctx := context.Background()
		hc := &http.Client{}

		err := downloadAndExtractTarball(ctx, hc, server.URL, dst)
		if err == nil {
			t.Error("expected gzip error")
		} else if !strings.Contains(err.Error(), "gzip") {
			t.Errorf("error should mention gzip, got: %v", err)
		}
	})

	t.Run("successful extraction", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var buf bytes.Buffer
			gw := gzip.NewWriter(&buf)
			tw := tar.NewWriter(gw)

			// Create normal tar entries
			entries := []struct {
				name string
				body string
			}{
				{"repo/file1.txt", "content1"},
				{"repo/file2.txt", "content2"},
			}

			for _, entry := range entries {
				hdr := &tar.Header{
					Name: entry.name,
					Mode: 0o600,
					Size: int64(len(entry.body)),
				}
				_ = tw.WriteHeader(hdr)
				_, _ = tw.Write([]byte(entry.body))
			}

			tw.Close()
			gw.Close()
			_, _ = w.Write(buf.Bytes())
		}))
		defer server.Close()

		dst := t.TempDir()
		ctx := context.Background()
		hc := &http.Client{}

		err := downloadAndExtractTarball(ctx, hc, server.URL, dst)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Verify files were created
		file1 := filepath.Join(dst, "repo", "file1.txt")
		if content, err := os.ReadFile(file1); err != nil {
			t.Errorf("failed to read extracted file: %v", err)
		} else if string(content) != "content1" {
			t.Errorf("file content = %q, want %q", string(content), "content1")
		}
	})
}

// TestReleasesDirectDepsVulnFree_Integration tests the main function with various scenarios.
// Note: These are integration tests that test error paths; full E2E tests exist in
// checks/release_deps_vulnfree_e2e_test.go (475 lines) which tests actual GitHub repos.
func TestReleasesDirectDepsVulnFree_Integration(t *testing.T) {
	t.Parallel()

	t.Run("empty releases list", func(t *testing.T) {
		t.Parallel()

		// Create a mock repo client that returns no releases
		mockRepo := &mockRepoClient{
			releases: []clients.Release{},
		}

		req := &checker.CheckRequest{
			Ctx:        context.Background(),
			RepoClient: mockRepo,
		}

		result, err := ReleasesDirectDepsVulnFree(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.Releases) != 0 {
			t.Errorf("expected 0 releases, got %d", len(result.Releases))
		}
	})

	t.Run("release with empty tag", func(t *testing.T) {
		t.Parallel()

		// Create a mock repo client with a release that has empty tag
		mockRepo := &mockRepoClient{
			releases: []clients.Release{
				{TagName: "  ", TargetCommitish: "abc123"},
			},
		}

		req := &checker.CheckRequest{
			Ctx:        context.Background(),
			RepoClient: mockRepo,
		}

		result, err := ReleasesDirectDepsVulnFree(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should skip releases with empty tags
		if len(result.Releases) != 0 {
			t.Errorf("expected 0 releases (empty tag should be skipped), got %d", len(result.Releases))
		}
	})
}

// mockRepoClient is a minimal mock implementation for testing.
//
//nolint:govet // Field alignment is a minor optimization
type mockRepoClient struct {
	releases []clients.Release
	tarURL   string
	tarErr   error
}

func (m *mockRepoClient) ListReleases() ([]clients.Release, error) {
	return m.releases, nil
}

func (m *mockRepoClient) ReleaseTarballURL(tag string) (string, error) {
	if m.tarErr != nil {
		return "", m.tarErr
	}
	return m.tarURL, nil
}

// Implement remaining RepoClient interface methods as no-ops.
func (m *mockRepoClient) InitRepo(repo clients.Repo, commitSHA string, commitDepth int) error {
	return nil
}

func (m *mockRepoClient) URI() string {
	return "github.com/test/repo"
}

func (m *mockRepoClient) IsArchived() (bool, error) {
	return false, nil
}

func (m *mockRepoClient) LocalPath() (string, error) {
	return "", nil
}

func (m *mockRepoClient) ListFiles(predicate func(string) (bool, error)) ([]string, error) {
	return nil, nil
}

func (m *mockRepoClient) GetFileReader(filename string) (io.ReadCloser, error) {
	return nil, nil
}

func (m *mockRepoClient) GetBranch(branch string) (*clients.BranchRef, error) {
	return nil, nil
}

func (m *mockRepoClient) GetCreatedAt() (time.Time, error) {
	return time.Time{}, nil
}

func (m *mockRepoClient) GetDefaultBranchName() (string, error) {
	return "main", nil
}

func (m *mockRepoClient) GetDefaultBranch() (*clients.BranchRef, error) {
	return nil, nil
}

func (m *mockRepoClient) GetOrgRepoClient(ctx context.Context) (clients.RepoClient, error) {
	return nil, nil
}

func (m *mockRepoClient) ListCommits() ([]clients.Commit, error) {
	return nil, nil
}

func (m *mockRepoClient) ListIssues() ([]clients.Issue, error) {
	return nil, nil
}

func (m *mockRepoClient) ListLicenses() ([]clients.License, error) {
	return nil, nil
}

func (m *mockRepoClient) ListProgrammingLanguages() ([]clients.Language, error) {
	return nil, nil
}

func (m *mockRepoClient) ListSuccessfulWorkflowRuns(filename string) ([]clients.WorkflowRun, error) {
	return nil, nil
}

func (m *mockRepoClient) ListCheckRunsForRef(ref string) ([]clients.CheckRun, error) {
	return nil, nil
}

func (m *mockRepoClient) ListStatuses(ref string) ([]clients.Status, error) {
	return nil, nil
}

func (m *mockRepoClient) ListWebhooks() ([]clients.Webhook, error) {
	return nil, nil
}

func (m *mockRepoClient) Search(request clients.SearchRequest) (clients.SearchResponse, error) {
	return clients.SearchResponse{}, nil
}

func (m *mockRepoClient) SearchCommits(request clients.SearchCommitsOptions) ([]clients.Commit, error) {
	return nil, nil
}

func (m *mockRepoClient) Close() error {
	return nil
}

func (m *mockRepoClient) GetFileContent(filename string) ([]byte, error) {
	return nil, nil
}

func (m *mockRepoClient) ListContributors() ([]clients.User, error) {
	return nil, nil
}
