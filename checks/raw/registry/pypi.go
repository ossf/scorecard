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

package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ossf/scorecard/v5/checker"
)

const (
	pypiSimpleAPIURL = "https://pypi.org/simple"
)

var (
	errPyPIStatusNotOK     = errors.New("PyPI API returned non-OK status")
	errVersionNotFound     = errors.New("version not found for package")
	errDownloadStatusNotOK = errors.New("download failed with non-OK status")
)

// PyPISimpleResponse represents the JSON-based simple API response (PEP 740).
type PyPISimpleResponse struct {
	Files []PyPIFile `json:"files"`
}

// PyPIFile represents a file in the simple API with provenance attestations.
type PyPIFile struct {
	Hashes     map[string]string `json:"hashes"`
	Provenance string            `json:"provenance,omitempty"` // URL to attestation
	Filename   string            `json:"filename"`
	URL        string            `json:"url"`
}

// GetPyPIArtifacts fetches PyPI artifacts and their Sigstore attestations.
// Uses the JSON-based simple API (PEP 740) to get provenance attestations.
// If client is nil, a default HTTP client will be used.
func GetPyPIArtifacts(
	ctx context.Context,
	packageName, version string,
	client HTTPClient,
) ([]ArtifactSignaturePair, error) {
	if client == nil {
		client = &http.Client{}
	}

	// Query PyPI simple API with JSON format (PEP 740)
	// Accept header tells PyPI we want JSON instead of HTML
	url := fmt.Sprintf("%s/%s/", pypiSimpleAPIURL, packageName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.pypi.simple.v1+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query PyPI API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", errPyPIStatusNotOK, resp.StatusCode)
	}

	// Parse response
	var apiResp PyPISimpleResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse PyPI response: %w", err)
	}

	var pairs []ArtifactSignaturePair
	foundVersion := false

	// For each file, check if it matches the version and has attestation
	for _, file := range apiResp.Files {
		// Check if file matches the requested version
		// PyPI filenames include version, e.g., package-1.0.0-py3-none-any.whl
		if !strings.Contains(file.Filename, version) {
			continue
		}
		foundVersion = true

		// Skip if no attestation available
		if file.Provenance == "" {
			continue
		}

		// Download the artifact
		artifact, err := downloadFileContent(ctx, client, file.URL)
		if err != nil {
			continue // Skip this file on error
		}

		// Download the attestation from the provenance URL
		attestation, err := downloadFileContent(ctx, client, file.Provenance)
		if err != nil {
			continue // Skip this file on error
		}

		// Create pair with attestation bundle as signature
		pair := ArtifactSignaturePair{
			ArtifactURL:   file.URL,
			ArtifactData:  artifact,
			SignatureURL:  file.Provenance,
			SignatureData: attestation,
			SignatureType: checker.SignatureTypeSigstore,
		}
		pairs = append(pairs, pair)
	}

	// If no files matched the version, return error
	if !foundVersion {
		return nil, fmt.Errorf("%w: %s for %s", errVersionNotFound, version, packageName)
	}

	return pairs, nil
}

// downloadFileContent downloads a file from a URL.
func downloadFileContent(ctx context.Context, client HTTPClient, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", errDownloadStatusNotOK, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return data, nil
}
