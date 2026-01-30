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
	"net/http"
	"strings"
	"time"

	"github.com/ossf/scorecard/v5/clients"
)

// HTTPClient interface for dependency injection and testing.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// AssetPair represents a release artifact and its corresponding signature file.
type AssetPair struct {
	ArtifactName  string
	ArtifactURL   string
	SignatureURL  string
	SignatureType string // "gpg", "sigstore", "minisig"
}

var signatureExtensions = map[string]string{
	".asc":           "gpg",
	".sig":           "gpg", // Often GPG, but could be other
	".sign":          "gpg", // Often GPG
	".minisig":       "minisig",
	".sigstore":      "sigstore",
	".sigstore.json": "sigstore",
}

// MatchAssetPairs finds signature files and matches them to their corresponding artifacts.
func MatchAssetPairs(assets []clients.ReleaseAsset) []AssetPair {
	var pairs []AssetPair

	// Create a map of signature files
	signatures := make(map[string]clients.ReleaseAsset)
	for _, asset := range assets {
		for ext, sigType := range signatureExtensions {
			if strings.HasSuffix(asset.Name, ext) {
				signatures[asset.Name] = asset
				// Store the signature type for later
				asset.Name = asset.Name + ":" + sigType
				break
			}
		}
	}

	// Match signatures to artifacts
	for _, asset := range assets {
		// Skip if this is a signature file itself
		isSignature := false
		for ext := range signatureExtensions {
			if strings.HasSuffix(asset.Name, ext) {
				isSignature = true
				break
			}
		}
		if isSignature {
			continue
		}

		// Look for signature files for this artifact
		for ext, sigType := range signatureExtensions {
			sigName := asset.Name + ext
			if sig, found := signatures[sigName]; found {
				pairs = append(pairs, AssetPair{
					ArtifactName:  asset.Name,
					ArtifactURL:   asset.URL,
					SignatureURL:  sig.URL,
					SignatureType: sigType,
				})
			}
		}
	}

	return pairs
}

// DownloadAsset downloads a release asset from the given URL.
func DownloadAsset(ctx context.Context, url string, client HTTPClient) ([]byte, error) {
	if client == nil {
		client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode) //nolint:err113
	}

	// Limit download size to 100MB to prevent abuse
	limitedReader := io.LimitReader(resp.Body, 100*1024*1024)

	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return data, nil
}
