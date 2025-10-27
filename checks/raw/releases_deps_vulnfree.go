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
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
)

var (
	errEmptyTarballURL   = errors.New("tarball URL is empty")
	errNoTarballSupport  = errors.New("repo client does not implement ReleaseTarballURL")
	errHTTPRequest       = errors.New("HTTP request failed")
	errPathTraversal     = errors.New("potential path traversal detected")
	errDecompressionBomb = errors.New("potential decompression bomb detected")
)

const (
	// Maximum file size to prevent decompression bombs (500MB).
	maxFileSize = 500 * 1024 * 1024
	// Maximum total decompressed size (1GB).
	maxTotalSize = 1024 * 1024 * 1024
)

// In-file defaults (used since the function no longer accepts options).
const (
	defaultMaxReleases  = 10
	defaultMaxIDsPerDep = 5
	httpTimeoutSec      = 60
)

// If RepoClient implements this, we'll use it to build the per-tag tarball URL.
type tarballURLer interface {
	ReleaseTarballURL(tag string) (string, error)
}

// CollectorClients holds injectable clients for testing.
type CollectorClients struct {
	OSV  clients.OSVAPIClient
	Deps clients.DepsClient
	HTTP *http.Client
}

// Collect per-release direct-deps + OSV vuln matches into checker.RawResults.
// This version materializes each release by downloading its tarball and scanning
// that snapshot (no clone).
func ReleasesDirectDepsVulnFree(c *checker.CheckRequest) (*checker.ReleaseDirectDepsVulnsData, error) {
	return ReleasesDirectDepsVulnFreeWithClients(c, nil)
}

// ReleasesDirectDepsVulnFreeWithClients is the internal implementation that accepts injectable clients.
// If clients is nil, default clients are created. This enables testing without external dependencies.
//
//nolint:gocognit,gocyclo // Complex but necessary logic for downloading and scanning releases
func ReleasesDirectDepsVulnFreeWithClients(
	c *checker.CheckRequest,
	testClients *CollectorClients,
) (*checker.ReleaseDirectDepsVulnsData, error) {
	ctx := c.Ctx
	repo := c.RepoClient

	// 1) fetch all releases (RepoClient signature has no args), cap to defaultMaxReleases.
	all, err := repo.ListReleases()
	if err != nil {
		return nil, fmt.Errorf("ListReleases: %w", err)
	}
	rels := all
	if defaultMaxReleases > 0 && len(rels) > defaultMaxReleases {
		rels = rels[:defaultMaxReleases]
	}
	if len(rels) == 0 {
		return &checker.ReleaseDirectDepsVulnsData{}, nil
	}

	// 2) deps + OSV clients - use injected or create defaults
	var osv clients.OSVAPIClient
	var depClient clients.DepsClient
	var httpc *http.Client

	if testClients != nil {
		osv = testClients.OSV
		depClient = testClients.Deps
		httpc = testClients.HTTP
	}

	if osv == nil {
		osv = clients.NewOSVClient() // HTTP OSV API client (clients/osv.go)
	}
	if depClient == nil {
		depClient = clients.NewDirectDepsClient() // from clients/scalibr_client.go (package clients)
	}
	if httpc == nil {
		httpc = &http.Client{Timeout: httpTimeoutSec * time.Second}
	}

	// 3) cache (dedupe queries across releases)
	type key struct{ eco, name, version, purl string }
	dep2IDs := make(map[key][]string)

	out := &checker.ReleaseDirectDepsVulnsData{
		Releases: make([]checker.ReleaseDepsVulns, 0, len(rels)),
	}

	// 4) per-release processing (download tarball, extract, scan that snapshot)
	for _, r := range rels {
		tag := strings.TrimSpace(r.TagName)
		if tag == "" {
			// Skip degenerate entries.
			continue
		}
		ref := strings.TrimSpace(r.TargetCommitish) // Commit-ish available on current clients.Release

		// Build tarball URL strictly via repo helper (handlers must provide it).
		var tarURL string
		if t, ok := any(repo).(tarballURLer); ok {
			u, err := t.ReleaseTarballURL(tag)
			if err != nil {
				return nil, fmt.Errorf("ReleaseTarballURL(%q): %w", tag, err)
			}
			if strings.TrimSpace(u) == "" {
				return nil, fmt.Errorf("%w: tag=%q", errEmptyTarballURL, tag)
			}
			tarURL = u
		} else {
			return nil, fmt.Errorf("%w: please implement ReleaseTarballURL in the GitHub/GitLab handlers", errNoTarballSupport)
		}

		// Materialize tarball into a temporary directory.
		tmpBase := os.TempDir()
		baseDir, err := os.MkdirTemp(tmpBase, "scorecard_rel_")
		if err != nil {
			return nil, fmt.Errorf("mkdtemp: %w", err)
		}
		// Best-effort cleanup on function return; leave on disk during loop in
		// case a debug session wants to inspect intermediate state.
		defer os.RemoveAll(baseDir)

		if err := downloadAndExtractTarball(ctx, httpc, tarURL, baseDir); err != nil {
			return nil, fmt.Errorf("download/extract tarball (%s): %w", tag, err)
		}
		root, err := soleSubdir(baseDir)
		if err != nil {
			return nil, fmt.Errorf("soleSubdir: %w", err)
		}

		// Run OSV-Scalibr (offline, manifest-only) for *direct* deps at that snapshot.
		dr, err := depClient.GetDeps(ctx, root)
		if err != nil {
			return nil, fmt.Errorf("scalibr GetDeps(tag=%s, ref=%s): %w", tag, ref, err)
		}

		// Sort deps for stable output.
		sort.Slice(dr.Deps, func(i, j int) bool {
			a, b := dr.Deps[i], dr.Deps[j]
			if a.Ecosystem != b.Ecosystem {
				return a.Ecosystem < b.Ecosystem
			}
			if a.Name != b.Name {
				return a.Name < b.Name
			}
			return a.Version < b.Version
		})

		// Build the OSV batch for new deps.
		queries := make([]clients.OSVQuery, 0, len(dr.Deps))
		qIndex := make([]int, 0, len(dr.Deps))
		for i, d := range dr.Deps {
			if strings.TrimSpace(d.Name) == "" || strings.TrimSpace(d.Version) == "" {
				continue
			}
			// Skip stdlib (Go)
			if strings.EqualFold(strings.TrimSpace(d.Name), "stdlib") {
				continue
			}
			k := key{
				eco:     toOSVEcosystem(d.Ecosystem),
				name:    d.Name,
				version: d.Version,
				purl:    d.PURL,
			}
			if _, ok := dep2IDs[k]; ok {
				continue
			}
			queries = append(queries, toOSVQuery(k))
			qIndex = append(qIndex, i)
		}

		if len(queries) > 0 {
			batchIDs, err := osv.QueryBatch(ctx, queries)
			if err != nil {
				return nil, fmt.Errorf("osv.QueryBatch: %w", err)
			}
			for pos, ids := range batchIDs {
				i := qIndex[pos]
				d := dr.Deps[i]
				k := key{
					eco:     toOSVEcosystem(d.Ecosystem),
					name:    d.Name,
					version: d.Version,
					purl:    d.PURL,
				}
				dep2IDs[k] = ids
			}
		}

		// Build findings list, applying time-gating to only include vulnerabilities
		// that were published before or at the release date.
		var findings []checker.DepVuln
		for _, d := range dr.Deps {
			k := key{
				eco:     toOSVEcosystem(d.Ecosystem),
				name:    d.Name,
				version: d.Version,
				purl:    d.PURL,
			}
			ids := dep2IDs[k]
			if len(ids) == 0 {
				continue
			}

			keep := ids
			// Apply time-gating if we have a valid release timestamp
			if !r.PublishedAt.IsZero() {
				keep = filterVulnsByPublishTime(ctx, osv, ids, r.PublishedAt)
			}

			if defaultMaxIDsPerDep > 0 && len(keep) > defaultMaxIDsPerDep {
				keep = keep[:defaultMaxIDsPerDep]
			}

			if len(keep) == 0 {
				continue
			}

			findings = append(findings, checker.DepVuln{
				Ecosystem:    k.eco,
				Name:         k.name,
				Version:      k.version,
				PURL:         k.purl,
				OSVIDs:       append([]string(nil), keep...),
				ManifestPath: filepath.ToSlash(d.Location),
			})
		}

		out.Releases = append(out.Releases, checker.ReleaseDepsVulns{
			Tag:         tag,
			CommitSHA:   ref,
			PublishedAt: r.PublishedAt,
			DirectDeps:  toDirectDeps(dr.Deps),
			Findings:    findings,
		})
	}

	return out, nil
}

// filterVulnsByPublishTime filters vulnerability IDs to only those published before or at releaseTime.
// This implements the "known at release time" policy.
func filterVulnsByPublishTime(
	ctx context.Context,
	osv clients.OSVAPIClient,
	vulnIDs []string,
	releaseTime time.Time,
) []string {
	var filtered []string
	for _, id := range vulnIDs {
		vuln, err := osv.GetVuln(ctx, id)
		if err != nil {
			// If we can't fetch the vuln record, conservatively include it
			// (could also choose to skip it)
			filtered = append(filtered, id)
			continue
		}
		// Only include if vulnerability was published before/at release time
		if vuln.Published != nil && !vuln.Published.After(releaseTime) {
			filtered = append(filtered, id)
		}
	}
	return filtered
}

// --- helpers ---

func toOSVEcosystem(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "gomod", "go", "golang":
		return "Go"
	case "npm", "node", "packagejson":
		return "npm"
	case "pypi", "python", "pyproject", "requirements":
		return "PyPI"
	case "maven", "pomxml", "gradle":
		return "Maven"
	case "cargo", "rust", "cargotoml", "crates.io":
		return "Crates.io"
	case "nuget", ".net", "nugetproj":
		return "NuGet"
	case "gem", "ruby", "rubygems", "gemfile":
		return "RubyGems"
	default:
		return s
	}
}

func toOSVQuery(k struct{ eco, name, version, purl string }) clients.OSVQuery {
	eco := toOSVEcosystem(k.eco)

	// Prefer PURL if provided.
	if strings.TrimSpace(k.purl) != "" {
		return clients.OSVQuery{Package: clients.OSVPackage{PURL: k.purl}}
	}

	// For Go, OSV expects v-prefixed semver when not using purl.
	v := strings.TrimSpace(k.version)
	if strings.EqualFold(eco, "Go") {
		if v != "" && !strings.HasPrefix(v, "v") && !strings.HasPrefix(v, "V") {
			v = "v" + v
		}
	}

	return clients.OSVQuery{
		Package: clients.OSVPackage{
			Name:      k.name,
			Ecosystem: eco,
		},
		Version: v,
	}
}

func toDirectDeps(in []clients.Dep) []checker.DirectDep {
	out := make([]checker.DirectDep, 0, len(in))
	for _, d := range in {
		out = append(out, checker.DirectDep{
			Ecosystem: d.Ecosystem,
			Name:      d.Name,
			Version:   d.Version,
			PURL:      d.PURL,
			Location:  filepath.ToSlash(d.Location),
		})
	}
	return out
}

// downloadAndExtractTarball streams a .tar.gz into dst/.
//
//nolint:gocognit // Complex but necessary logic for secure tar extraction
func downloadAndExtractTarball(ctx context.Context, hc *http.Client, url, dst string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("User-Agent", "ossf-scorecard-release-tarball")

	resp, err := hc.Do(req)
	if err != nil {
		return fmt.Errorf("GET tarball: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		b, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("%w: %s (failed to read body: %w)", errHTTPRequest, resp.Status, readErr)
		}
		return fmt.Errorf("%w: %s: %s", errHTTPRequest, resp.Status, string(b))
	}

	gr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("gzip: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	var totalSize int64
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar next: %w", err)
		}

		// Prevent path traversal.
		if !isValidPath(hdr.Name) {
			return fmt.Errorf("%w: %s", errPathTraversal, hdr.Name)
		}

		//nolint:gosec // G305: False positive - we validate path above with isValidPath
		target := filepath.Join(dst, hdr.Name)
		// Ensure target is within dst.
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(dst)+string(os.PathSeparator)) &&
			filepath.Clean(target) != filepath.Clean(dst) {
			return fmt.Errorf("%w: %s", errPathTraversal, hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return fmt.Errorf("mkdir: %w", err)
			}
		case tar.TypeReg:
			// Check file size to prevent decompression bomb.
			if hdr.Size > maxFileSize {
				return fmt.Errorf("%w: file %s size %d exceeds limit %d", errDecompressionBomb, hdr.Name, hdr.Size, maxFileSize)
			}
			totalSize += hdr.Size
			if totalSize > maxTotalSize {
				return fmt.Errorf("%w: total size %d exceeds limit %d", errDecompressionBomb, totalSize, maxTotalSize)
			}

			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return fmt.Errorf("mkdir: %w", err)
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return fmt.Errorf("open file: %w", err)
			}
			// Use LimitReader to enforce size limits.
			limited := io.LimitReader(tr, hdr.Size+1) // +1 to detect overflow
			written, copyErr := io.Copy(f, limited)
			closeErr := f.Close()
			if copyErr != nil {
				return fmt.Errorf("copy file: %w", copyErr)
			}
			if closeErr != nil {
				return fmt.Errorf("close file: %w", closeErr)
			}
			if written > hdr.Size {
				return fmt.Errorf("%w: file %s actual size exceeds header size", errDecompressionBomb, hdr.Name)
			}
		default:
			// ignore other types (symlinks, etc.)
		}
	}
	return nil
}

// isValidPath checks for path traversal attempts.
func isValidPath(p string) bool {
	clean := filepath.Clean(p)
	// Reject absolute paths and paths with .. components.
	if filepath.IsAbs(clean) || strings.Contains(clean, "..") {
		return false
	}
	return true
}

func soleSubdir(dir string) (string, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("read dir: %w", err)
	}
	for _, e := range ents {
		if e.IsDir() {
			return filepath.Join(dir, e.Name()), nil
		}
	}
	return dir, nil
}
