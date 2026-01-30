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

package releaseassets

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
)

const (
	keySourceConfigRelease = "config-release"
	keySourceConfigGlobal  = "config-global"
)

var (
	// keyFingerprintRegex matches GPG key fingerprints in text.
	keyFingerprintRegex = regexp.MustCompile(`(?i)(?:key|fingerprint|gpg).*?([0-9A-F]{40})`)

	// Common locations for KEYS files in repositories.
	keysFilePaths = []string{
		"KEYS",
		"KEYS.txt",
		".github/KEYS",
		"keys/KEYS",
		"gpg/KEYS",
		"GPG-KEYS",
	}
)

// matchesPattern checks if a release tag matches a glob pattern.
// Supports wildcards like "v1.*", "v2.0.*", or exact matches like "v1.2.3".
func matchesPattern(tag, pattern string) bool {
	if pattern == "" || pattern == "*" {
		return true
	}

	// Try exact match first
	if tag == pattern {
		return true
	}

	// Use filepath.Match for glob pattern matching
	matched, err := filepath.Match(pattern, tag)
	if err != nil {
		// If pattern is invalid, fall back to prefix match
		return strings.HasPrefix(tag, strings.TrimSuffix(pattern, "*"))
	}
	return matched
}

// KeyDiscoveryResult contains information about discovered GPG keys.
type KeyDiscoveryResult struct {
	// Source indicates where the keys were found:
	// "config-release", "config-global", "KEYS_file", "release_notes", "github_profile", "keyserver"
	Source          string
	KeyFingerprints []string
	KeyURLs         []string // URLs to fetch GPG keys from
}

// DiscoverGPGKeys attempts to find GPG keys for verifying release signatures.
// It tries multiple strategies in order of reliability:
// 0. User-configured key URLs from scorecard.yml (per-release or global)
// 1. KEYS file in repository.
// 2. Key fingerprints in release notes.
// 3. GitHub user's GPG keys (if available).
//
//nolint:gocognit,nestif // Configuration parsing requires nested type assertions
func DiscoverGPGKeys(ctx context.Context, c *checker.CheckRequest, release *clients.Release) []KeyDiscoveryResult {
	var results []KeyDiscoveryResult

	// Strategy 0: Check user-configured GPG key URLs from scorecard.yml
	if c.Config != nil {
		if cfg, ok := c.Config.(interface {
			GetGPGKeys() interface{}
		}); ok {
			if gpgKeys := cfg.GetGPGKeys(); gpgKeys != nil {
				// First try to match per-release keys
				if keyConfig, ok := gpgKeys.(interface {
					GetReleases() interface{}
				}); ok {
					if releasesRaw := keyConfig.GetReleases(); releasesRaw != nil {
						if releases, ok := releasesRaw.([]interface{}); ok {
							for _, relRaw := range releases {
								if rel, ok := relRaw.(interface {
									GetTag() string
									GetURLs() []string
								}); ok {
									if matchesPattern(release.TagName, rel.GetTag()) {
										urls := rel.GetURLs()
										if len(urls) > 0 {
											results = append(results, KeyDiscoveryResult{
												KeyURLs: urls,
												Source:  "config-release",
											})
											break // Use first matching pattern
										}
									}
								}
							}
						}
					}
				}

				// If no per-release match, try global keys
				if len(results) == 0 {
					if keyConfig, ok := gpgKeys.(interface {
						GetURLs() []string
					}); ok {
						urls := keyConfig.GetURLs()
						if len(urls) > 0 {
							results = append(results, KeyDiscoveryResult{
								KeyURLs: urls,
								Source:  "config-global",
							})
						}
					}
				}
			}
		}
	}

	// Strategy 1: Check for KEYS file in repository
	if keys := findKeysFile(ctx, c); len(keys) > 0 {
		results = append(results, KeyDiscoveryResult{
			KeyFingerprints: keys,
			Source:          "KEYS_file",
		})
	}

	// Strategy 2: Extract fingerprints from release notes
	if keys := extractKeyFingerprints(release.Body); len(keys) > 0 {
		results = append(results, KeyDiscoveryResult{
			KeyFingerprints: keys,
			Source:          "release_notes",
		})
	}

	// Strategy 3: Try to get keys from release tag name (sometimes contains commit)
	// This is a fallback and less reliable

	return results
}

// findKeysFile looks for a KEYS file in common locations in the repository.
func findKeysFile(ctx context.Context, c *checker.CheckRequest) []string {
	for _, path := range keysFilePaths {
		reader, err := c.RepoClient.GetFileReader(path)
		if err != nil {
			continue
		}

		// Read the file content
		content, err := io.ReadAll(reader)
		_ = reader.Close()
		if err != nil {
			continue
		}

		// Parse GPG keys from the KEYS file
		keys := parseGPGKeysFile(string(content))
		if len(keys) > 0 {
			return keys
		}
	}
	return nil
}

// parseGPGKeysFile extracts GPG key fingerprints from a KEYS file.
func parseGPGKeysFile(content string) []string {
	var fingerprints []string

	// Look for fingerprint lines in GPG key blocks
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check for "Key fingerprint = " format
		if strings.Contains(line, "fingerprint") && strings.Contains(line, "=") {
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				fp := strings.ReplaceAll(strings.TrimSpace(parts[1]), " ", "")
				if isValidFingerprint(fp) {
					fingerprints = append(fingerprints, fp)
				}
			}
		}

		// Also check for standalone fingerprints
		if len(line) == 40 && isValidFingerprint(line) {
			fingerprints = append(fingerprints, line)
		}
	}

	return fingerprints
}

// extractKeyFingerprints finds GPG key fingerprints in release notes or other text.
func extractKeyFingerprints(text string) []string {
	matches := keyFingerprintRegex.FindAllStringSubmatch(text, -1)

	var fingerprints []string
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			fp := strings.ToUpper(strings.ReplaceAll(match[1], " ", ""))
			if isValidFingerprint(fp) && !seen[fp] {
				fingerprints = append(fingerprints, fp)
				seen[fp] = true
			}
		}
	}

	return fingerprints
}

// isValidFingerprint checks if a string is a valid GPG fingerprint format.
func isValidFingerprint(fp string) bool {
	if len(fp) != 40 {
		return false
	}

	for _, c := range fp {
		if (c < '0' || c > '9') && (c < 'A' || c > 'F') && (c < 'a' || c > 'f') {
			return false
		}
	}

	return true
}

// FetchKeyFromKeyserver retrieves a GPG public key from a keyserver given a fingerprint.
func FetchKeyFromKeyserver(ctx context.Context, fingerprint, keyserverURL string, client HTTPClient) (string, error) {
	if keyserverURL == "" {
		keyserverURL = "https://keyserver.ubuntu.com"
	}

	// Use HKP protocol to fetch the key
	url := fmt.Sprintf("%s/pks/lookup?op=get&search=0x%s", keyserverURL, fingerprint)

	data, err := DownloadAsset(ctx, url, client)
	if err != nil {
		return "", fmt.Errorf("fetch key from keyserver: %w", err)
	}

	return string(data), nil
}
