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

// Package winget implements a client that resolves Windows Package Manager (winget)
// packages to their source repository.
//
// Manifests are read from https://github.com/microsoft/winget-pkgs, the community
// repository that backs the default winget source.
package winget

import (
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"

	pmc "github.com/ossf/scorecard/v5/cmd/internal/packagemanager"
	sce "github.com/ossf/scorecard/v5/errors"
)

const (
	contentsURL = "https://api.github.com/repos/microsoft/winget-pkgs/contents/manifests/"
	rawURL      = "https://raw.githubusercontent.com/microsoft/winget-pkgs/master/manifests/"

	// maxResponseBytes bounds how much of a response is read. Manifests are a few KB and
	// the GitHub contents API returns at most 1000 entries.
	maxResponseBytes = 4 << 20
)

var (
	urlRegexp           = regexp.MustCompile(`https?://[^\s"'<>]+`)
	defaultLocaleRegexp = regexp.MustCompile(`(?m)^DefaultLocale:\s*["']?([\w-]+)`)
	fieldRegexp         = regexp.MustCompile(`^[\s-]*(\w+):`)
)

type Client interface {
	GitRepositoryByPackageName(packageName string) (string, error)
}

type WingetClient struct {
	Manager pmc.Client
}

type contentsEntry struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// GitRepositoryByPackageName returns the source repository of the winget package by
// reading the manifests of its newest version.
func (c *WingetClient) GitRepositoryByPackageName(packageID string) (string, error) {
	segments := strings.Split(packageID, ".")
	if len(segments) < 2 || slices.Contains(segments, "") {
		return "", sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("invalid winget package ID format (expected Publisher.Name): %s", packageID))
	}
	// Every dot in the ID is a directory, e.g. BurntSushi.ripgrep.MSVC is at b/BurntSushi/ripgrep/MSVC.
	for i := range segments {
		segments[i] = url.PathEscape(segments[i])
	}
	pkgPath := strings.ToLower(string([]rune(packageID)[0])) + "/" + strings.Join(segments, "/")

	version, err := c.latestVersion(packageID, pkgPath)
	if err != nil {
		return "", err
	}

	manifestPrefix := rawURL + pkgPath + "/" + url.PathEscape(version) + "/" + url.PathEscape(packageID)
	locale := "en-US"
	if m := defaultLocaleRegexp.FindSubmatch(c.fetchManifest(manifestPrefix + ".yaml")); m != nil {
		locale = string(m[1])
	}
	localeManifest := c.fetchManifest(manifestPrefix + ".locale." + locale + ".yaml")
	installerManifest := c.fetchManifest(manifestPrefix + ".installer.yaml")
	if localeManifest == nil && installerManifest == nil {
		return "", sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("failed to fetch manifests for winget package %s version %s", packageID, version))
	}

	if repo := mostCommonRepo(localeManifest, installerManifest); repo != "" {
		return repo, nil
	}
	return "", sce.WithMessage(sce.ErrScorecardInternal,
		fmt.Sprintf("winget package %s has no source repository link in its manifest, use --repo instead", packageID))
}

// latestVersion lists the package directory and returns its newest version.
func (c *WingetClient) latestVersion(packageID, pkgPath string) (string, error) {
	resp, err := c.Manager.GetURI(contentsURL + pkgPath)
	if err != nil {
		return "", sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("failed to fetch winget versions for %s: %v", packageID, err))
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return "", sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("could not find winget package %s (package identifiers are case-sensitive, e.g. Git.Git)",
				packageID))
	}
	if resp.StatusCode != http.StatusOK {
		return "", sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("failed to fetch winget versions for %s: HTTP %d", packageID, resp.StatusCode))
	}

	var entries []contentsEntry
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxResponseBytes)).Decode(&entries); err != nil {
		return "", sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("failed to parse winget versions response: %v", err))
	}
	// The package directory also holds sub-packages (e.g. "Preview", "EXE", locale codes),
	// so prefer directories that look like versions ("1.2.3" or "v1.2.3"). If there are none,
	// fall back to any directory; a sub-package has no manifests, so that fails safely.
	var latest, fallback string
	for _, entry := range entries {
		if entry.Type != "dir" || entry.Name == "" {
			continue
		}
		if !looksLikeVersion(entry.Name) {
			fallback = cmp.Or(fallback, entry.Name)
			continue
		}
		if latest == "" || compareVersions(entry.Name, latest) > 0 {
			latest = entry.Name
		}
	}
	latest = cmp.Or(latest, fallback)
	if latest == "" {
		return "", sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("no versions found for winget package: %s", packageID))
	}
	return latest, nil
}

// fetchManifest returns the manifest body, or nil if it can't be fetched.
func (c *WingetClient) fetchManifest(uri string) []byte {
	resp, err := c.Manager.GetURI(uri)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return nil
	}
	return body
}

// mostCommonRepo returns the repository linked from the most manifest fields.
// A field counts once per repository, so e.g. one InstallerUrl per architecture
// doesn't outvote PackageUrl, and a related repository that is only linked from
// one field (e.g. a LicenseUrl) loses. Ties go to the repository seen first.
func mostCommonRepo(manifests ...[]byte) string {
	type vote struct{ field, repo string }
	seen := map[vote]bool{}
	counts := map[string]int{}
	var best string
	for _, manifest := range manifests {
		for _, line := range strings.Split(string(manifest), "\n") {
			// URLs outside a "Field: value" line (e.g. in a multi-line description) share one field.
			field := ""
			if m := fieldRegexp.FindStringSubmatch(line); m != nil {
				field = m[1]
			}
			for _, rawURL := range urlRegexp.FindAllString(line, -1) {
				repo := repoFromURL(rawURL)
				if repo == "" || seen[vote{field, repo}] {
					continue
				}
				seen[vote{field, repo}] = true
				counts[repo]++
				if counts[repo] > counts[best] {
					best = repo
				}
			}
		}
	}
	return best
}

// repoFromURL maps a URL on a forge Scorecard supports to its repository root.
func repoFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	host := strings.ToLower(u.Hostname())
	// Split the escaped path so an encoded "/" (e.g. GitLab's group%2Fproject) stays in its segment.
	parts := strings.Split(strings.Trim(u.EscapedPath(), "/"), "/")
	for i := range parts {
		if unescaped, err := url.PathUnescape(parts[i]); err == nil {
			parts[i] = unescaped
		}
	}
	switch {
	case host == "github.com" || host == "www.github.com":
		if len(parts) < 2 {
			return ""
		}
		return githubRepo(parts[0], parts[1])
	case strings.HasSuffix(host, ".github.io") && parts[0] != "":
		// GitHub Pages: https://<owner>.github.io/<repo>
		return githubRepo(strings.TrimSuffix(host, ".github.io"), parts[0])
	case host == "dev.azure.com":
		// https://dev.azure.com/<org>/<project>/_git/<repo>
		if len(parts) >= 4 && parts[2] == "_git" {
			return "https://dev.azure.com/" + strings.Join(parts[:4], "/")
		}
		return ""
	case len(parts) >= 4 && parts[0] == "api" && parts[1] == "v4" && parts[2] == "projects":
		// GitLab API (e.g. the generic package registry used for release files). The project is
		// either a URL-encoded path or a numeric ID, which can't be mapped without another request.
		if strings.Contains(parts[3], "/") {
			return strings.ToLower("https://" + host + "/" + parts[3])
		}
		return ""
	case parts[0] == "api" || parts[0] == "-":
		// GitLab system paths such as /-/project/<id>/uploads/... never start with a project.
		return ""
	}
	// GitLab (gitlab.com or self-hosted) separates the project path from the resource with "/-/".
	if i := slices.Index(parts, "-"); i >= 2 {
		return strings.ToLower("https://" + host + "/" + strings.Join(parts[:i], "/"))
	}
	if host == "gitlab.com" && len(parts) >= 2 {
		return strings.ToLower("https://gitlab.com/" + parts[0] + "/" + strings.TrimSuffix(parts[1], ".git"))
	}
	return ""
}

// githubRepo returns the lowercased GitHub repository URL, or "" for links that
// aren't repositories (e.g. GitHub Sponsors).
func githubRepo(owner, repo string) string {
	owner = strings.ToLower(owner)
	repo = strings.TrimSuffix(strings.ToLower(repo), ".git")
	if owner == "" || repo == "" || owner == "sponsors" {
		return ""
	}
	return fmt.Sprintf("https://github.com/%s/%s", owner, repo)
}

// looksLikeVersion reports whether a directory name starts like a version, e.g. "1.2" or "v1.2".
func looksLikeVersion(name string) bool {
	name = strings.TrimPrefix(strings.TrimPrefix(name, "v"), "V")
	return name != "" && name[0] >= '0' && name[0] <= '9'
}

// compareVersions compares dotted versions segment by segment, numerically
// where both segments are integers. A leading "v" is ignored.
// It returns >0 if a is newer than b, <0 if older.
func compareVersions(a, b string) int {
	a = strings.TrimPrefix(strings.TrimPrefix(a, "v"), "V")
	b = strings.TrimPrefix(strings.TrimPrefix(b, "v"), "V")
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(as) && i < len(bs); i++ {
		an, aErr := strconv.Atoi(as[i])
		bn, bErr := strconv.Atoi(bs[i])
		switch {
		case aErr == nil && bErr == nil && an != bn:
			return an - bn
		case (aErr != nil || bErr != nil) && as[i] != bs[i]:
			return strings.Compare(as[i], bs[i])
		}
	}
	return len(as) - len(bs)
}
