// Copyright 2024 OpenSSF Scorecard Authors
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
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/finding"
)

// reChangelogFile matches common changelog filenames at the repo root.
//   - CHANGELOG, CHANGES: standard changelog names
//   - NEWS: common in GNU projects
//   - HISTORY: used by some projects for version history
//   - RELEASE-NOTES, RELEASE_NOTES: common in Apache/Java projects
//
// Optional extensions: .md, .txt, .rst, .adoc.
var reChangelogFile = regexp.MustCompile(
	`(?i)^(changelog|changes|news|history|release[-_]notes)(\.(md|txt|rst|adoc))?$`,
)

// reVersion extracts semver-like version strings from changelog lines.
// Matches patterns like:
//   - ## [1.0.0] - 2024-01-15  (Keep a Changelog)
//   - ## 1.0.0                  (simple markdown)
//   - # 1.0.0 / 2024-01-15     (Ruby style)
//   - Version 3.0 (2024-01-15) (GNU NEWS)
//   - 3.0.0 (2024-01-15)       (Python CHANGES.rst)
//   - Release 3.0.0            (Apache style)
//
// The core pattern captures a semver-like string: major.minor with optional .patch
// and optional pre-release suffix (e.g., -rc.1, -beta.2).
var reVersion = regexp.MustCompile(
	`(?:^|\s|\[|v)(\d+\.\d+(?:\.\d+)?(?:-[A-Za-z0-9.]+)?)(?:\]|\s|$)`,
)

const changelogReleaseLookBack = 5

// Changelog retrieves the raw data for the Changelog check.
func Changelog(c *checker.CheckRequest) (checker.ChangelogData, error) {
	var results checker.ChangelogData

	// Look for changelog files at the top level of the repo.
	repoFiles, err := c.RepoClient.ListFiles(func(file string) (bool, error) {
		return reChangelogFile.MatchString(file), nil
	})
	if err != nil {
		return results, fmt.Errorf("RepoClient.ListFiles: %w", err)
	}
	for _, f := range repoFiles {
		results.ChangelogFiles = append(results.ChangelogFiles, checker.File{
			Path: f,
			Type: finding.FileTypeSource,
		})
	}

	// Read the first changelog file found and extract version strings
	// that have substantive content.
	if len(results.ChangelogFiles) > 0 {
		versions, err := extractVersionsFromFile(c, results.ChangelogFiles[0].Path)
		if err != nil {
			// Non-fatal: we found the file but couldn't parse it.
			if c.Dlogger != nil {
				c.Dlogger.Debug(&checker.LogMessage{
					Text: fmt.Sprintf("could not parse changelog: %v", err),
				})
			}
		} else {
			results.ChangelogVersions = versions
		}
	}

	// Get releases and check which ones are covered by the changelog.
	releases, err := c.RepoClient.ListReleases()
	if err != nil && !errors.Is(err, clients.ErrUnsupportedFeature) {
		return results, fmt.Errorf("RepoClient.ListReleases: %w", err)
	}

	versionSet := makeVersionSet(results.ChangelogVersions)

	for i, r := range releases {
		if i >= changelogReleaseLookBack {
			break
		}
		results.TotalReleases++
		tag := normalizeVersion(r.TagName)
		if versionSet[tag] || hasSubstantiveBody(r.Body) {
			results.ReleasesWithChangelog++
		}
	}

	return results, nil
}

// hasSubstantiveBody checks whether a release body has meaningful content
// beyond just whitespace or boilerplate like "Full Changelog: ...".
func hasSubstantiveBody(body string) bool {
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// Skip GitHub auto-generated "Full Changelog" links.
		if strings.HasPrefix(trimmed, "**Full Changelog**") {
			continue
		}
		return true
	}
	return false
}

// extractVersionsFromFile reads a changelog file and returns all version
// strings that have substantive content below them.
func extractVersionsFromFile(c *checker.CheckRequest, path string) ([]string, error) {
	reader, err := c.RepoClient.GetFileReader(path)
	if err != nil {
		return nil, fmt.Errorf("GetFileReader: %w", err)
	}
	defer reader.Close()

	return extractVersions(reader), nil
}

// extractVersions scans a changelog and returns versions that have
// substantive content. A version entry is considered substantive if
// there is at least one non-empty, non-header line between it and
// the next version header (or end of file).
func extractVersions(r io.Reader) []string {
	var versions []string
	seen := make(map[string]bool)

	var currentVersion string
	hasContent := false

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if v := extractVersionFromLine(line); v != "" {
			// We hit a new version header. Finalize the previous one.
			if currentVersion != "" && hasContent && !seen[currentVersion] {
				seen[currentVersion] = true
				versions = append(versions, currentVersion)
			}
			currentVersion = v
			hasContent = false
			continue
		}

		// Check if this line is substantive content (not empty, not a
		// sub-header like "### Added", not an rst underline like "====").
		if currentVersion != "" && !hasContent && isContentLine(trimmed) {
			hasContent = true
		}
	}

	// Finalize the last version entry.
	if currentVersion != "" && hasContent && !seen[currentVersion] {
		versions = append(versions, currentVersion)
	}

	return versions
}

// extractVersionFromLine returns a normalized version string if the line
// contains a version header, or empty string if not.
func extractVersionFromLine(line string) string {
	matches := reVersion.FindStringSubmatch(line)
	if len(matches) < 2 {
		return ""
	}
	// Only treat lines that look like headers — they should start with
	// a markdown header (#), contain "version"/"release", start with
	// the version number, or have brackets around the version.
	trimmed := strings.TrimSpace(line)
	v := normalizeVersion(matches[1])

	switch {
	case strings.HasPrefix(trimmed, "#"):
		return v
	case strings.HasPrefix(trimmed, v), strings.HasPrefix(trimmed, "v"+v):
		return v
	case strings.HasPrefix(strings.ToLower(trimmed), "version"):
		return v
	case strings.HasPrefix(strings.ToLower(trimmed), "release"):
		return v
	case strings.Contains(trimmed, "["+matches[1]+"]"):
		return v
	}
	return ""
}

// isContentLine returns true if the line is substantive content —
// not empty, not a markdown sub-header, not an rst underline.
func isContentLine(trimmed string) bool {
	if trimmed == "" {
		return false
	}
	// Markdown sub-headers (### Added, ### Fixed, etc.)
	if strings.HasPrefix(trimmed, "#") {
		return false
	}
	// RST underlines (====, ----, ~~~~)
	if isRSTUnderline(trimmed) {
		return false
	}
	return true
}

// isRSTUnderline checks for reStructuredText section underlines.
func isRSTUnderline(s string) bool {
	if len(s) < 3 {
		return false
	}
	c := s[0]
	if c != '=' && c != '-' && c != '~' && c != '^' && c != '+' {
		return false
	}
	for i := 1; i < len(s); i++ {
		if s[i] != c {
			return false
		}
	}
	return true
}

// normalizeVersion strips a leading "v" prefix for comparison.
func normalizeVersion(tag string) string {
	return strings.TrimPrefix(strings.TrimSpace(tag), "v")
}

// makeVersionSet creates a lookup set from a slice of version strings.
func makeVersionSet(versions []string) map[string]bool {
	m := make(map[string]bool, len(versions))
	for _, v := range versions {
		m[v] = true
	}
	return m
}
