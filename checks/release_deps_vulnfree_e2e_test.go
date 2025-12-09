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

//go:build integration

package checks_test

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v53/github"
	"golang.org/x/oauth2"

	"github.com/ossf/scorecard/v5/clients"
)

// Ground truth from the repo's README:
// https://github.com/AdamKorcz/repo1-with-vulnerable-releases
// “Vulnerability Matrix” and “CVE References” sections.
//
// --- Expected direct deps per tag (name -> version without the leading "v") ---
var expectedDeps = map[string]map[string]string{
	"v1.0.0": {
		"golang.org/x/text":            "0.3.8",
		"github.com/gorilla/websocket": "1.5.3",
		"golang.org/x/net":             "0.39.0",
		"golang.org/x/crypto":          "0.36.0",
		"github.com/gin-gonic/gin":     "1.9.1",
		"github.com/golang-jwt/jwt/v4": "4.5.2",
	},
	"v0.9.0": {
		"golang.org/x/text":            "0.3.8",
		"github.com/gorilla/websocket": "1.5.3",
		"golang.org/x/net":             "0.39.0",
		"golang.org/x/crypto":          "0.35.0",
		"github.com/gin-gonic/gin":     "1.9.1",
		"github.com/golang-jwt/jwt/v4": "4.5.2",
	},
	"v0.8.0": { // vulnerable: jwt v4.5.1
		"golang.org/x/text":            "0.3.8",
		"github.com/gorilla/websocket": "1.5.3",
		"golang.org/x/net":             "0.39.0",
		"golang.org/x/crypto":          "0.35.0",
		"github.com/gin-gonic/gin":     "1.9.1",
		"github.com/golang-jwt/jwt/v4": "4.5.1",
	},
	"v0.7.0": {
		"golang.org/x/text":            "0.3.8",
		"github.com/gorilla/websocket": "1.5.3",
		"golang.org/x/net":             "0.39.0",
		"golang.org/x/crypto":          "0.35.0",
		"github.com/gin-gonic/gin":     "1.9.1",
		"github.com/golang-jwt/jwt/v4": "4.5.2",
	},
	"v0.6.0": {
		"golang.org/x/text":            "0.3.8",
		"github.com/gorilla/websocket": "1.5.3",
		"golang.org/x/net":             "0.38.0",
		"golang.org/x/crypto":          "0.35.0",
		"github.com/gin-gonic/gin":     "1.9.1",
		"github.com/golang-jwt/jwt/v4": "4.5.2",
	},
	"v0.5.0": { // vulnerable: multiple
		"golang.org/x/text":            "0.3.5",
		"github.com/gorilla/websocket": "1.4.0",
		"golang.org/x/net":             "0.36.0",
		"golang.org/x/crypto":          "0.33.0",
		"github.com/gin-gonic/gin":     "1.6.3",
		"github.com/golang-jwt/jwt/v4": "4.5.2",
	},
	"v0.4.0": { // vulnerable: multiple
		"golang.org/x/text":            "0.3.5",
		"github.com/gorilla/websocket": "1.4.0",
		"golang.org/x/net":             "0.36.0",
		"golang.org/x/crypto":          "0.34.0",
		"github.com/gin-gonic/gin":     "1.6.3",
		"github.com/golang-jwt/jwt/v4": "4.5.2",
	},
	"v0.3.0": { // vulnerable: multiple
		"golang.org/x/text":            "0.3.5",
		"github.com/gorilla/websocket": "1.4.0",
		"golang.org/x/net":             "0.37.0",
		"golang.org/x/crypto":          "0.34.0",
		"github.com/gin-gonic/gin":     "1.6.3",
		"github.com/golang-jwt/jwt/v4": "4.5.2",
	},
	"v0.2.0": { // vulnerable: multiple
		"golang.org/x/text":            "0.3.5",
		"github.com/gorilla/websocket": "1.4.0",
		"golang.org/x/net":             "0.37.0",
		"golang.org/x/crypto":          "0.34.0",
		"github.com/gin-gonic/gin":     "1.8.1",
		"github.com/golang-jwt/jwt/v4": "4.5.2",
	},
	"v0.1.0": { // vulnerable: multiple
		"golang.org/x/text":            "0.3.6",
		"github.com/gorilla/websocket": "1.4.0",
		"golang.org/x/net":             "0.37.0",
		"golang.org/x/crypto":          "0.34.0",
		"github.com/gin-gonic/gin":     "1.8.1",
		"github.com/golang-jwt/jwt/v4": "4.5.2",
	},
}

// Which releases should (per README) have at least one vulnerable direct dep:
var expectedVulnerable = map[string]bool{
	"v0.1.0": true,
	"v0.2.0": true,
	"v0.3.0": true,
	"v0.4.0": true,
	"v0.5.0": true,
	"v0.8.0": true,
	"v0.6.0": false,
	"v0.7.0": false,
	"v0.9.0": false,
	"v1.0.0": false,
}

// Exact CVE IDs from README “CVE References” mapped per tag.
var expectedCVEs = map[string][]string{
	"v0.1.0": {"CVE-2021-38561", "CVE-2020-27813", "CVE-2025-22872", "CVE-2025-22869", "CVE-2023-26125"},
	"v0.2.0": {"CVE-2021-38561", "CVE-2020-27813", "CVE-2025-22872", "CVE-2025-22869", "CVE-2023-26125"},
	"v0.3.0": {"CVE-2021-38561", "CVE-2020-27813", "CVE-2025-22872", "CVE-2025-22869", "CVE-2023-26125"},
	"v0.4.0": {"CVE-2021-38561", "CVE-2020-27813", "CVE-2025-22872", "CVE-2025-22869", "CVE-2023-26125"},
	"v0.5.0": {"CVE-2021-38561", "CVE-2020-27813", "CVE-2025-22872", "CVE-2025-22869", "CVE-2023-26125"},
	"v0.6.0": {},                 // clean
	"v0.7.0": {},                 // clean
	"v0.8.0": {"CVE-2025-30204"}, // jwt 4.5.1
	"v0.9.0": {},                 // clean
	"v1.0.0": {},                 // clean
}

// aliasCache caches OSV ID -> set of aliases (including the ID itself).
type aliasCache struct {
	cache map[string]map[string]struct{}
}

func newAliasCache() *aliasCache {
	return &aliasCache{cache: make(map[string]map[string]struct{})}
}

// expandAliases fetches a vuln by the exact ID string (case-sensitive as returned by batch)
// and returns a set of all IDs/aliases (uppercased) for comparisons.
func (ac *aliasCache) expandAliases(ctx context.Context, osv clients.OSVAPIClient, exactID string) (map[string]struct{}, error) {
	exact := strings.TrimSpace(exactID)
	if exact == "" {
		return map[string]struct{}{}, nil
	}
	key := strings.ToUpper(exact) // cache key normalized, but fetch with exact

	if s, ok := ac.cache[key]; ok {
		return s, nil
	}

	rec, err := osv.GetVuln(ctx, exact) // IMPORTANT: use exact case as returned by batch
	if err != nil {
		// Do not fail the whole test; return a minimal set containing the ID itself.
		// Caller can still succeed if another ID (e.g., GO-...) carries the CVE alias.
		min := map[string]struct{}{key: {}}
		ac.cache[key] = min
		return min, nil
	}

	set := make(map[string]struct{}, 1+len(rec.Aliases))
	set[strings.ToUpper(strings.TrimSpace(rec.ID))] = struct{}{}
	for _, a := range rec.Aliases {
		set[strings.ToUpper(strings.TrimSpace(a))] = struct{}{}
	}
	ac.cache[key] = set
	return set, nil
}

func TestReleasesDirectDepsVulnFree_GitHubRepo(t *testing.T) {
	const (
		owner        = "AdamKorcz"
		repo         = "repo1-with-vulnerable-releases"
		wantReleases = 10
		httpTimeout  = 60 * time.Second
	)

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()

	// Optional:
	// t.Setenv("SCALIBR_DEBUG", "1")
	// t.Setenv("OSV_DEBUG", "1")

	// GitHub client (token optional, recommended to avoid rate limits)
	var gh *github.Client
	if tok := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); tok != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: tok})
		tc := oauth2.NewClient(ctx, ts)
		gh = github.NewClient(tc)
	} else {
		gh = github.NewClient(nil)
	}

	// 1) Fetch releases
	rels, _, err := gh.Repositories.ListReleases(ctx, owner, repo, &github.ListOptions{PerPage: wantReleases})
	if err != nil {
		t.Fatalf("ListReleases: %v", err)
	}
	if len(rels) == 0 {
		t.Fatalf("no releases found")
	}
	if len(rels) > wantReleases {
		rels = rels[:wantReleases]
	}

	// Ensure we got all tags we have ground truth for.
	var tags []string
	for _, r := range rels {
		if tag := strings.TrimSpace(r.GetTagName()); tag != "" {
			tags = append(tags, tag)
		}
	}
	slices.Sort(tags)
	for tag := range expectedDeps {
		if !slices.Contains(tags, tag) {
			t.Fatalf("expected tag %s to be among latest %d releases; got %v", tag, wantReleases, tags)
		}
	}

	httpClient := &http.Client{Timeout: httpTimeout}
	authHeader := ""
	if tok := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); tok != "" {
		authHeader = "Bearer " + tok
	}

	depClient := clients.NewDirectDepsClient()
	osvClient := clients.NewOSVClient()
	aliasC := newAliasCache()

	type perRelease struct {
		Tag    string
		Deps   map[string]string   // name->version (normalized)
		OSVIDs []string            // raw OSV IDs returned by batch
		AllIDs map[string]struct{} // expanded IDs+aliases (uppercased)
	}
	var got []perRelease

	// 2) For each release: tarball -> deps -> OSV IDs (+ aliases)
	for _, r := range rels {
		tag := strings.TrimSpace(r.GetTagName())
		if tag == "" {
			continue
		}
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/tarball/%s", owner, repo, tag)
		tmp := t.TempDir()
		if err := downloadAndExtractTarball(ctx, httpClient, authHeader, url, tmp); err != nil {
			t.Fatalf("tag %s: download/extract: %v", tag, err)
		}
		root, err := soleSubdir(tmp)
		if err != nil {
			t.Fatalf("tag %s: root: %v", tag, err)
		}

		// Direct deps from Scalibr client.
		dr, err := depClient.GetDeps(ctx, root)
		if err != nil {
			t.Fatalf("tag %s: GetDeps: %v", tag, err)
		}
		if len(dr.Deps) == 0 {
			t.Fatalf("tag %s: expected deps, got none", tag)
		}

		// Build actual direct deps map (skip stdlib), normalize versions (strip 'v').
		actual := make(map[string]string)
		for _, d := range dr.Deps {
			if strings.EqualFold(d.Name, "stdlib") {
				continue
			}
			ver := strings.TrimSpace(d.Version)
			if strings.HasPrefix(ver, "v") || strings.HasPrefix(ver, "V") {
				ver = ver[1:]
			}
			actual[d.Name] = ver
		}

		// 2a) ASSERT: actual deps == expected deps for that tag.
		exp, ok := expectedDeps[tag]
		if !ok {
			t.Fatalf("no expectedDeps for tag %s", tag)
		}
		if len(actual) != len(exp) {
			t.Fatalf("tag %s: dep count mismatch: got %d, want %d\nactual=%v\nwant=%v", tag, len(actual), len(exp), actual, exp)
		}
		for name, wantVer := range exp {
			gotVer, ok := actual[name]
			if !ok {
				t.Fatalf("tag %s: missing expected dep %q", tag, name)
			}
			if gotVer != wantVer {
				t.Fatalf("tag %s: dep %q version mismatch: got %q, want %q", tag, name, gotVer, wantVer)
			}
		}
		for name := range actual {
			if _, ok := exp[name]; !ok {
				t.Fatalf("tag %s: unexpected extra dep %q", tag, name)
			}
		}

		// 2b) Build OSV batch queries (ecosystem "Go", v-prefixed versions).
		var queries []clients.OSVQuery
		for name, ver := range actual {
			v := ver
			if v != "" && !strings.HasPrefix(v, "v") && !strings.HasPrefix(v, "V") {
				v = "v" + v
			}
			queries = append(queries, clients.OSVQuery{
				Package: clients.OSVPackage{Ecosystem: "Go", Name: name},
				Version: v,
			})
		}

		res, err := osvClient.QueryBatch(ctx, queries)
		if err != nil {
			body, _ := json.Marshal(map[string]any{"queries": queries})
			t.Fatalf("tag %s: OSV QueryBatch error: %v\npayload=%s", tag, err, string(body))
		}
		var ids []string
		for _, row := range res {
			ids = append(ids, row...)
		}

		// Expand to include aliases (CVE/GHSA/GO-) for assertion.
		expanded := make(map[string]struct{})
		for _, id := range ids {
			set, err := aliasC.expandAliases(ctx, osvClient, id)
			if err != nil {
				// Non-fatal: continue, another ID may carry the aliases (e.g., GO-…).
				t.Logf("tag %s: expand aliases for %s: %v", tag, id, err)
				continue
			}
			for a := range set {
				expanded[a] = struct{}{}
			}
		}

		got = append(got, perRelease{
			Tag:    tag,
			Deps:   actual,
			OSVIDs: ids,
			AllIDs: expanded,
		})
		t.Logf("tag %s: deps OK (%d); OSV IDs=%v", tag, len(actual), ids)
	}

	// 3) Assertions vs. README:
	//    a) vuln presence/absence
	//    b) specific CVE IDs for that tag must appear among expanded aliases
	for _, r := range got {
		wantVuln := expectedVulnerable[r.Tag]
		if wantVuln && len(r.OSVIDs) == 0 {
			t.Fatalf("tag %s: expected >=1 OSV vuln but got none", r.Tag)
		}
		if !wantVuln && len(r.OSVIDs) > 0 {
			t.Fatalf("tag %s: expected 0 OSV vulns but got %d (%v)", r.Tag, len(r.OSVIDs), r.OSVIDs)
		}

		expectIDs := expectedCVEs[r.Tag]
		if len(expectIDs) == 0 {
			continue // clean tag
		}
		for _, cve := range expectIDs {
			if _, ok := r.AllIDs[strings.ToUpper(cve)]; !ok {
				t.Fatalf("tag %s: expected OSV to include %s via IDs/aliases; got raw IDs=%v", r.Tag, cve, r.OSVIDs)
			}
		}
	}
}

// ---- helpers ----

func soleSubdir(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if e.IsDir() {
			return filepath.Join(dir, e.Name()), nil
		}
	}
	return dir, nil
}

func downloadAndExtractTarball(ctx context.Context, hc *http.Client, authHeader, url, dst string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "ossf-scorecard-integration-test")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := hc.Do(req)
	if err != nil {
		return fmt.Errorf("GET tarball: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GET %s: %s: %s", url, resp.Status, string(b))
	}

	gr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("gzip: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar next: %w", err)
		}
		target := filepath.Join(dst, hdr.Name)
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(f, tr)
			closeErr := f.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
		default:
			// Ignore other entry types for the test.
		}
	}
	return nil
}
