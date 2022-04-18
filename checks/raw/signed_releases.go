// Copyright 2020 Security Scorecard Authors
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
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/sigstore/cosign/cmd/cosign/cli/rekor"
	"github.com/sigstore/cosign/pkg/cosign"

	"github.com/sigstore/rekor/pkg/generated/client"
	"github.com/sigstore/rekor/pkg/generated/client/index"
	"github.com/sigstore/rekor/pkg/generated/models"
)

// Default address of transparency log to search for signing events.
var defaultRekorAddr = "https://rekor.sigstore.dev"

var (
	errorRekorSearch = errors.New("error searching rekor entries")
)

func findTLogEntriesByPayload(ctx context.Context, rekorClient *client.Rekor, artifactSha string) (uuids []string, err error) {
	params := index.NewSearchIndexParamsWithContext(ctx)
	params.Query = &models.SearchIndex{}

	params.Query.Hash = fmt.Sprintf("sha256:%s", artifactSha)

	searchIndex, err := rekorClient.Index.SearchIndex(params)
	if err != nil {
		return nil, err
	}
	return searchIndex.GetPayload(), nil
}

func getRekorEntries(ctx context.Context, rClient *client.Rekor, url string) ([]models.LogEntryAnon, error) {
	// URL to hash reader
	hasher := sha256.New()
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching '%v': %w", url, err)
	}
	defer resp.Body.Close()

	if _, err := io.Copy(io.Writer(hasher), resp.Body); err != nil {
		return nil, err
	}

	artifactSha := strings.ToLower(hex.EncodeToString(hasher.Sum(nil)))

	// Use search index to find rekor entry UUIDs that match Subject Digest.
	uuids, err := findTLogEntriesByPayload(ctx, rClient, artifactSha)
	if err != nil {
		return nil, err
	}

	// Get and verify log entries
	res := []models.LogEntryAnon{}

	for _, uuid := range uuids {
		e, err := cosign.GetTlogEntry(ctx, rClient, uuid)
		if err != nil {
			continue
		}
		if err := cosign.VerifyTLogEntry(ctx, rClient, e); err != nil {
			continue
		}
		res = append(res, *e)
	}
	return res, nil
}

// SignedReleases checks for presence of signed release check.
func SignedReleases(c *checker.CheckRequest) (checker.SignedReleasesData, error) {
	releases, err := c.RepoClient.ListReleases()
	if err != nil {
		return checker.SignedReleasesData{}, fmt.Errorf("%w", err)
	}

	rClient, err := rekor.NewClient(defaultRekorAddr)
	if err != nil {
		// Maybe don't fail on error. Maybe some people use scorecard in weird environments.
		return checker.SignedReleasesData{}, fmt.Errorf("%w", err)
	}

	var results checker.SignedReleasesData
	for i, r := range releases {
		results.Releases = append(results.Releases,
			checker.Release{
				Tag: r.TagName,
				URL: r.URL,
			})

		for _, asset := range r.Assets {
			// Perform a Rekor search index query.
			// Only check the latest release, for sake of time.
			var entries []models.LogEntryAnon
			if rClient != nil && i == 0 {
				entries, err = getRekorEntries(c.Ctx, rClient, asset.BrowserDownloadURL)
				if err != nil {
					// TODO: Just skip if there's an error fetching Rekor entries?
					return checker.SignedReleasesData{}, fmt.Errorf("%w", err)
				}
			}

			a := checker.ReleaseAsset{
				URL:          asset.URL,
				Name:         asset.Name,
				RekorEntries: entries,
			}
			results.Releases[i].Assets = append(results.Releases[i].Assets, a)
		}
	}

	// Return raw results.
	return results, nil
}
