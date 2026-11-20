// Copyright 2022 OpenSSF Scorecard Authors
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
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/internal/ciiowner"
)

var (
	errEmptyClient = errors.New("CII client is nil")

	// badgeURLRe matches bestpractices.dev (and legacy coreinfrastructure.org) badge URLs.
	// The host is anchored to the scheme so that only the exact registrable domains
	// (optionally with subdomains) match; look-alike hosts such as
	// "evil-bestpractices.dev" or "bestpractices.dev.attacker.com" do not.
	badgeURLRe = regexp.MustCompile(
		`https?://(?:[a-zA-Z0-9-]+\.)*(?:bestpractices\.dev|coreinfrastructure\.org)/(?:en/)?projects/(\d+)`,
	)
)

const maxReadmeBytes = 512 * 1024 // 512 KB

// CIIBestPractices retrieves the raw data for the CIIBestPractices check.
func CIIBestPractices(c *checker.CheckRequest) (checker.CIIBestPracticesData, error) {
	var results checker.CIIBestPracticesData
	if c.CIIClient == nil {
		return results, fmt.Errorf("%w", errEmptyClient)
	}

	badge, err := c.CIIClient.GetBadgeLevel(c.Ctx, c.Repo.URI())
	if err != nil {
		return results, fmt.Errorf("%w", err)
	}
	if badge != clients.NotFound {
		results.Badge = badge
		return results, nil
	}

	// Fallback: check whether a sibling repo in the same org has a registered badge.
	// Only the HTTP client implements CIIProjectClient; the blob client (scorecard.dev)
	// does not, so the fallback is skipped there and today's behaviour is preserved.
	projectClient, ok := c.CIIClient.(clients.CIIProjectClient)
	if !ok {
		results.Badge = badge
		return results, nil
	}

	for _, id := range extractBadgeIDs(c) {
		resp, lookupErr := projectClient.GetBadgeByProjectID(c.Ctx, id)
		if lookupErr != nil || resp.RepoURL == "" {
			continue
		}
		if !ciiowner.SameProject(c.Repo.URI(), resp.RepoURL) {
			continue
		}
		level, levelErr := resp.Level()
		if levelErr != nil {
			continue
		}
		results.Badge = level
		results.InheritedFrom = resp.RepoURL
		return results, nil
	}

	results.Badge = clients.NotFound
	return results, nil
}

// extractBadgeIDs scans README files at the repo root for bestpractices.dev
// project IDs, skipping fenced code blocks. Returns at most 2 unique IDs.
func extractBadgeIDs(c *checker.CheckRequest) []int {
	readmes, err := c.RepoClient.ListFiles(func(path string) (bool, error) {
		if strings.ContainsRune(path, '/') {
			return false, nil
		}
		return strings.HasPrefix(strings.ToUpper(path), "README"), nil
	})
	if err != nil || len(readmes) == 0 {
		return nil
	}

	seen := map[int]bool{}
	var ids []int

	for _, file := range readmes {
		if len(ids) >= 2 {
			break
		}
		reader, openErr := c.RepoClient.GetFileReader(file)
		if openErr != nil {
			continue
		}
		data, readErr := io.ReadAll(io.LimitReader(reader, maxReadmeBytes))
		reader.Close()
		if readErr != nil {
			continue
		}
		ids = appendIDsFromContent(data, seen, ids)
	}
	return ids
}

// appendIDsFromContent scans readme content line by line, skipping fenced code
// blocks, and appends any new badge project IDs to ids (capped at 2 total).
func appendIDsFromContent(content []byte, seen map[int]bool, ids []int) []int {
	inFence := false
	for _, line := range strings.Split(string(content), "\n") {
		if len(ids) >= 2 {
			break
		}
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		ids = appendIDsFromLine(line, seen, ids)
	}
	return ids
}

// appendIDsFromLine extracts badge project IDs from a single line of text.
func appendIDsFromLine(line string, seen map[int]bool, ids []int) []int {
	for _, m := range badgeURLRe.FindAllStringSubmatch(line, -1) {
		if len(ids) >= 2 {
			break
		}
		id, err := strconv.Atoi(m[1])
		if err != nil || seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
	}
	return ids
}
