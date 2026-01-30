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
	"time"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/checks/raw/signature"
	"github.com/ossf/scorecard/v5/clients"
)

// VerifyReleaseSignatures verifies cryptographic signatures on GitHub/GitLab
// release assets.
func VerifyReleaseSignatures(ctx context.Context, c *checker.CheckRequest,
	release *clients.Release,
) []checker.PackageSignature {
	var signatures []checker.PackageSignature

	logDebugf(c, "Attempting to match signature pairs from %d assets", len(release.Assets))

	// Match signature files to artifacts
	pairs := MatchAssetPairs(release.Assets)
	if len(pairs) == 0 {
		logDebugf(c, "No signature/artifact pairs found in release %s", release.TagName)
		return signatures
	}

	logDebugf(c, "Found %d signature/artifact pairs in release %s", len(pairs), release.TagName)

	// Discover GPG keys if needed
	var keyData string
	gpgKeyDiscovery := DiscoverGPGKeys(ctx, c, release)

	for _, pair := range pairs {
		sig := verifyAssetPair(ctx, c, release, pair, gpgKeyDiscovery, keyData)
		// Only include non-empty signatures (skip download failures)
		if sig.ArtifactURL != "" || sig.SignatureURL != "" {
			signatures = append(signatures, sig)
		}
	}

	return signatures
}

func verifyAssetPair(ctx context.Context, c *checker.CheckRequest, release *clients.Release,
	pair AssetPair, keyDiscovery []KeyDiscoveryResult, existingKeyData string,
) checker.PackageSignature {
	logDebugf(c, "Verifying %s signature for %s in release %s",
		pair.SignatureType, pair.ArtifactName, release.TagName)

	// Download the artifact
	artifactData, err := DownloadAsset(ctx, pair.ArtifactURL, nil)
	if err != nil {
		// Don't treat download failures as verification failures
		// This allows signature detection to work even when assets aren't downloadable
		logDebugf(c, "Skipping verification for %s - failed to download artifact: %v",
			pair.ArtifactName, err)
		return checker.PackageSignature{}
	}

	// Download the signature
	signatureData, err := DownloadAsset(ctx, pair.SignatureURL, nil)
	if err != nil {
		// Don't treat download failures as verification failures
		logDebugf(c, "Skipping verification for %s - failed to download signature: %v",
			pair.ArtifactName, err)
		return checker.PackageSignature{}
	}

	// Verify based on signature type
	switch pair.SignatureType {
	case "gpg":
		return verifyGPGSignature(ctx, c, artifactData, signatureData, pair, keyDiscovery, existingKeyData)
	case "sigstore":
		return verifySigstoreSignature(ctx, artifactData, signatureData, pair)
	case "minisign":
		return checker.PackageSignature{
			ArtifactURL:  pair.ArtifactURL,
			SignatureURL: pair.SignatureURL,
			Type:         checker.SignatureTypeMinisign,
			IsVerified:   false,
			ErrorMsg:     "minisign verification not yet implemented",
		}
	default:
		return checker.PackageSignature{
			ArtifactURL:  pair.ArtifactURL,
			SignatureURL: pair.SignatureURL,
			Type:         checker.SignatureTypeUnknown,
			IsVerified:   false,
			ErrorMsg:     fmt.Sprintf("unknown signature type: %s", pair.SignatureType),
		}
	}
}

func verifyGPGSignature(ctx context.Context, c *checker.CheckRequest,
	artifactData, signatureData []byte, pair AssetPair,
	keyDiscovery []KeyDiscoveryResult, existingKeyData string,
) checker.PackageSignature {
	verifier := &signature.GPGVerifier{}

	// First try configured key URLs if available (per-release or global)
	for _, discovery := range keyDiscovery {
		isConfigSource := discovery.Source == keySourceConfigRelease ||
			discovery.Source == keySourceConfigGlobal
		if isConfigSource && len(discovery.KeyURLs) > 0 {
			for _, keyURL := range discovery.KeyURLs {
				logInfof(c, "Trying GPG key from config: %s", keyURL)

				// Download the key with timeout
				httpClient := &http.Client{Timeout: 30 * time.Second}
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, keyURL, nil)
				if err != nil {
					logWarnf(c, "Failed to create request for %s: %v", keyURL, err)
					continue
				}

				resp, err := httpClient.Do(req)
				if err != nil {
					logWarnf(c, "Failed to download key from %s: %v", keyURL, err)
					continue
				}

				if resp.StatusCode != http.StatusOK {
					resp.Body.Close()
					logWarnf(c, "Key download from %s returned status: %d", keyURL, resp.StatusCode)
					continue
				}

				// Read key data with size limit (5MB)
				const maxKeySize = 5 * 1024 * 1024
				keyData, err := io.ReadAll(io.LimitReader(resp.Body, maxKeySize))
				resp.Body.Close()
				if err != nil {
					logWarnf(c, "Failed to read key from %s: %v", keyURL, err)
					continue
				}

				// Try importing and verifying with this key
				// Note: The GPG verifier will need the key in its keyring
				// For now, we'll just log success of downloading
				logInfof(c, "Successfully downloaded GPG key from %s (%d bytes)", keyURL, len(keyData))
				// TODO: Actually use the key for verification once signature package supports it
			}
		}
	}

	// Try verification with keyserver (which will extract key ID from signature)
	opts := &signature.VerifyOptions{
		KeyserverURL: "https://keys.openpgp.org",
	}

	result, err := verifier.Verify(ctx, artifactData, signatureData, opts)
	if err != nil {
		return checker.PackageSignature{
			ArtifactURL:  pair.ArtifactURL,
			SignatureURL: pair.SignatureURL,
			Type:         checker.SignatureTypeGPG,
			IsVerified:   false,
			ErrorMsg:     fmt.Sprintf("GPG verification failed: %v", err),
		}
	}

	sig := checker.PackageSignature{
		ArtifactURL:  pair.ArtifactURL,
		SignatureURL: pair.SignatureURL,
		Type:         checker.SignatureTypeGPG,
		IsVerified:   result.Verified,
		KeyID:        result.KeyID,
	}

	if !result.Verified && result.Error != nil {
		sig.ErrorMsg = result.Error.Error()
	}

	if result.Verified {
		logInfof(c, "GPG signature verified for %s in release %s (Key: %s)",
			pair.ArtifactName, pair.SignatureURL, result.KeyID)
	} else {
		// Don't penalize projects for key discovery issues - log as info
		errorDetail := ""
		if result.Error != nil {
			errorDetail = fmt.Sprintf(" - error: %v", result.Error)
		}
		logInfof(c, "GPG signature verification could not be completed for %s in release %s%s (this does not affect score)",
			pair.ArtifactName, pair.SignatureURL, errorDetail)
	}

	return sig
}

func verifySigstoreSignature(ctx context.Context, artifactData, signatureData []byte,
	pair AssetPair,
) checker.PackageSignature {
	verifier := &signature.SigstoreVerifier{}
	result, err := verifier.Verify(ctx, artifactData, signatureData, &signature.VerifyOptions{})
	if err != nil {
		return checker.PackageSignature{
			ArtifactURL:  pair.ArtifactURL,
			SignatureURL: pair.SignatureURL,
			Type:         checker.SignatureTypeSigstore,
			IsVerified:   false,
			ErrorMsg:     fmt.Sprintf("Sigstore verification failed: %v", err),
		}
	}

	sig := checker.PackageSignature{
		ArtifactURL:  pair.ArtifactURL,
		SignatureURL: pair.SignatureURL,
		Type:         checker.SignatureTypeSigstore,
		IsVerified:   result.Verified,
	}

	if !result.Verified && result.Error != nil {
		sig.ErrorMsg = result.Error.Error()
	}

	return sig
}

func logDebugf(c *checker.CheckRequest, format string, args ...interface{}) {
	if c.Dlogger != nil {
		c.Dlogger.Debug(&checker.LogMessage{
			Text: fmt.Sprintf(format, args...),
		})
	}
}

func logInfof(c *checker.CheckRequest, format string, args ...interface{}) {
	if c.Dlogger != nil {
		c.Dlogger.Info(&checker.LogMessage{
			Text: fmt.Sprintf(format, args...),
		})
	}
}

func logWarnf(c *checker.CheckRequest, format string, args ...interface{}) {
	if c.Dlogger != nil {
		c.Dlogger.Warn(&checker.LogMessage{
			Text: fmt.Sprintf(format, args...),
		})
	}
}
