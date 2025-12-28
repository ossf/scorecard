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
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ossf/scorecard/v5/checker"
)

const (
	mavenCentralBaseURL = "https://repo1.maven.org/maven2"
)

var (
	errInvalidMavenPackageName = errors.New("invalid Maven package name (expected groupId:artifactId)")
	errHTTPStatusNotOK         = errors.New("HTTP request failed with non-OK status")
)

// ArtifactSignaturePair represents an artifact and its signature(s).
type ArtifactSignaturePair struct {
	SignatureType checker.SignatureType
	ArtifactURL   string
	SignatureURL  string
	ArtifactData  []byte
	SignatureData []byte
}

// HTTPClient defines the interface for making HTTP requests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// GetMavenArtifacts fetches Maven Central artifacts and their signatures.
// Package name format: groupId:artifactId (e.g., "com.example:myapp").
// If client is nil, a default HTTP client will be used.
func GetMavenArtifacts(
	ctx context.Context,
	packageName, version string,
	client HTTPClient,
) ([]ArtifactSignaturePair, error) {
	groupID, artifactID, err := parseMavenCoordinates(packageName)
	if err != nil {
		return nil, err
	}

	if client == nil {
		client = &http.Client{}
	}
	extensions := []string{".jar", ".pom"}

	var pairs []ArtifactSignaturePair
	for _, ext := range extensions {
		baseURL := constructMavenURL(groupID, artifactID, version, ext)
		pair := findSignatureForArtifact(ctx, client, baseURL)
		if pair != nil {
			pairs = append(pairs, *pair)
		}
	}

	return pairs, nil
}

// parseMavenCoordinates extracts groupId and artifactId from package name.
func parseMavenCoordinates(packageName string) (string, string, error) {
	parts := strings.Split(packageName, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("%w: %s", errInvalidMavenPackageName, packageName)
	}
	return parts[0], parts[1], nil
}

// findSignatureForArtifact tries Sigstore first, falls back to GPG.
func findSignatureForArtifact(
	ctx context.Context,
	client HTTPClient,
	baseURL string,
) *ArtifactSignaturePair {
	// Try Sigstore first (preferred)
	if pair := trySigstoreSignature(ctx, client, baseURL); pair != nil {
		return pair
	}

	// Fall back to GPG (mandatory on Maven Central)
	return tryGPGSignature(ctx, client, baseURL)
}

// trySigstoreSignature attempts to find and download Sigstore signature.
func trySigstoreSignature(
	ctx context.Context,
	client HTTPClient,
	baseURL string,
) *ArtifactSignaturePair {
	sigstoreURLs := []string{
		baseURL + ".sigstore.json",
		baseURL + ".sigstore",
	}

	for _, sigURL := range sigstoreURLs {
		if !checkURLExists(ctx, client, sigURL) {
			continue
		}

		pair, err := downloadArtifactPair(
			ctx,
			client,
			baseURL,
			sigURL,
			checker.SignatureTypeSigstore,
		)
		if err == nil {
			return &pair
		}
	}

	return nil
}

// tryGPGSignature attempts to find and download GPG signature.
func tryGPGSignature(
	ctx context.Context,
	client HTTPClient,
	baseURL string,
) *ArtifactSignaturePair {
	gpgURL := baseURL + ".asc"
	if !checkURLExists(ctx, client, gpgURL) {
		return nil
	}

	pair, err := downloadArtifactPair(
		ctx,
		client,
		baseURL,
		gpgURL,
		checker.SignatureTypeGPG,
	)
	if err != nil {
		return nil
	}

	return &pair
}

// constructMavenURL builds a Maven Central URL from coordinates.
// Example: com.example:myapp:1.0.0.jar ->
// https://repo1.maven.org/maven2/com/example/myapp/1.0.0/myapp-1.0.0.jar
func constructMavenURL(
	groupID, artifactID, version, extension string,
) string {
	groupPath := strings.ReplaceAll(groupID, ".", "/")
	filename := fmt.Sprintf("%s-%s%s", artifactID, version, extension)
	return fmt.Sprintf(
		"%s/%s/%s/%s/%s",
		mavenCentralBaseURL,
		groupPath,
		artifactID,
		version,
		filename,
	)
}

// checkURLExists performs a HEAD request to check if a URL exists.
func checkURLExists(
	ctx context.Context,
	client HTTPClient,
	url string,
) bool {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodHead,
		url,
		nil,
	)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// downloadArtifactPair downloads both artifact and signature.
func downloadArtifactPair(
	ctx context.Context,
	client HTTPClient,
	artifactURL, signatureURL string,
	signatureType checker.SignatureType,
) (ArtifactSignaturePair, error) {
	// Download artifact
	artifactData, err := downloadFile(ctx, client, artifactURL)
	if err != nil {
		return ArtifactSignaturePair{}, fmt.Errorf(
			"artifact download failed: %w",
			err,
		)
	}

	// Download signature
	signatureData, err := downloadFile(ctx, client, signatureURL)
	if err != nil {
		return ArtifactSignaturePair{}, fmt.Errorf(
			"signature download failed: %w",
			err,
		)
	}

	return ArtifactSignaturePair{
		ArtifactURL:   artifactURL,
		ArtifactData:  artifactData,
		SignatureURL:  signatureURL,
		SignatureData: signatureData,
		SignatureType: signatureType,
	}, nil
}

// downloadFile downloads a file from a URL.
func downloadFile(ctx context.Context, client HTTPClient, url string) ([]byte, error) {
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
		return nil, fmt.Errorf("%w: %d", errHTTPStatusNotOK, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return data, nil
}
