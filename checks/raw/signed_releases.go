// Copyright 2020 OpenSSF Scorecard Authors
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
	"fmt"
	"strings"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/checks/raw/registry"
	"github.com/ossf/scorecard/v5/checks/raw/releaseassets"
	"github.com/ossf/scorecard/v5/checks/raw/signature"
)

// Helper functions for logging to reduce boilerplate.

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

// SignedReleases checks for presence of signed release check.
func SignedReleases(c *checker.CheckRequest) (checker.SignedReleasesData, error) {
	releases, err := c.RepoClient.ListReleases()
	if err != nil {
		return checker.SignedReleasesData{}, fmt.Errorf("%w", err)
	}
	logInfof(c, "Found %d GitHub releases from ListReleases()", len(releases))

	pkgs := []checker.ProjectPackage{}

	// Get package versions from deps.dev for Maven/PyPI verification
	versions, err := c.ProjectClient.GetProjectPackageVersions(
		c.Ctx,
		c.Repo.Host(),
		c.Repo.Path(),
	)
	if err != nil {
		logDebugf(c, "GetProjectPackageVersions: %v", err)
	} else {
		// Process package versions for Maven/PyPI
		for _, v := range versions.Versions {
			prov := checker.PackageProvenance{}

			if len(v.SLSAProvenances) > 0 {
				prov = checker.PackageProvenance{
					Commit:     v.SLSAProvenances[0].Commit,
					IsVerified: v.SLSAProvenances[0].Verified,
				}
			}

			pkg := checker.ProjectPackage{
				System:     v.VersionKey.System,
				Name:       v.VersionKey.Name,
				Version:    v.VersionKey.Version,
				Provenance: prov,
				Signatures: []checker.PackageSignature{},
			}

			// Attempt to verify signatures for this package
			signatures := verifyPackageSignatures(c, &pkg)
			pkg.Signatures = signatures

			pkgs = append(pkgs, pkg)
		}
	}

	// Verify GitHub/GitLab release signatures
	if len(releases) > 0 {
		logInfof(c, "Starting GitHub/GitLab release verification for %d releases", len(releases))
		for i := range releases {
			logDebugf(c, "Checking release %s with %d assets", releases[i].TagName, len(releases[i].Assets))
			releaseSignatures := releaseassets.VerifyReleaseSignatures(c.Ctx, c, &releases[i])
			if len(releaseSignatures) > 0 {
				logInfof(c, "Found %d verified signatures for release %s",
					len(releaseSignatures), releases[i].TagName)
				// Store signatures in a pseudo-package for the release
				releasePkg := checker.ProjectPackage{
					System:     "github_release",
					Name:       releases[i].TagName,
					Version:    releases[i].TagName,
					Signatures: releaseSignatures,
				}
				pkgs = append(pkgs, releasePkg)
			}
		}
	} else {
		logDebugf(c, "No GitHub/GitLab releases found to verify")
	}

	return checker.SignedReleasesData{
		Releases: releases,
		Packages: pkgs,
	}, nil
}

// verifyPackageSignatures attempts to verify signatures for a package.
func verifyPackageSignatures(c *checker.CheckRequest, pkg *checker.ProjectPackage) []checker.PackageSignature {
	var signatures []checker.PackageSignature

	switch strings.ToLower(pkg.System) {
	case "maven":
		signatures = verifyMavenSignatures(c, pkg)
	case "pypi":
		signatures = verifyPyPISignatures(c, pkg)
	default:
		// Package system not supported for verification yet
		logDebugf(c,
			"Signature verification not supported for package system: %s (package: %s:%s)",
			pkg.System, pkg.Name, pkg.Version)
	}

	return signatures
}

// verifyMavenSignatures verifies GPG signatures for Maven Central packages.
func verifyMavenSignatures(c *checker.CheckRequest, pkg *checker.ProjectPackage) []checker.PackageSignature {
	var signatures []checker.PackageSignature

	logDebugf(c, "Attempting Maven GPG verification for %s:%s", pkg.Name, pkg.Version)

	// Get Maven artifacts with GPG signatures
	pairs, err := registry.GetMavenArtifacts(c.Ctx, pkg.Name, pkg.Version, nil)
	if err != nil {
		logDebugf(c, "Failed to fetch Maven artifact: %v", err)
		return signatures
	}

	if len(pairs) == 0 {
		logDebugf(c, "No Maven artifacts found for %s:%s", pkg.Name, pkg.Version)
		return signatures
	}

	// Verify GPG signature
	verifier := &signature.GPGVerifier{}
	for _, pair := range pairs {
		result, err := verifier.Verify(c.Ctx, pair.ArtifactData, pair.SignatureData, &signature.VerifyOptions{})
		if err != nil {
			logWarnf(c, "Maven GPG verification error for %s:%s: %v", pkg.Name, pkg.Version, err)
			continue
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

		signatures = append(signatures, sig)

		if result.Verified {
			logInfof(c, "Maven GPG signature verified for %s:%s (Key: %s)", pkg.Name, pkg.Version, result.KeyID)
		} else {
			errorDetail := ""
			if result.Error != nil {
				errorDetail = fmt.Sprintf(": %v", result.Error)
			}
			logWarnf(c, "Maven GPG signature verification failed for %s:%s%s", pkg.Name, pkg.Version, errorDetail)
		}
	}

	return signatures
}

// verifyPyPISignatures verifies Sigstore attestations for PyPI packages.
func verifyPyPISignatures(c *checker.CheckRequest, pkg *checker.ProjectPackage) []checker.PackageSignature {
	var signatures []checker.PackageSignature

	logDebugf(c, "Attempting PyPI Sigstore verification for %s:%s", pkg.Name, pkg.Version)

	// Get PyPI artifacts with attestations
	pairs, err := registry.GetPyPIArtifacts(c.Ctx, pkg.Name, pkg.Version, nil)
	if err != nil {
		logDebugf(c, "Failed to fetch PyPI artifacts: %v", err)
		return signatures
	}

	if len(pairs) == 0 {
		logDebugf(c, "No PyPI attestations found for %s:%s", pkg.Name, pkg.Version)
		return signatures
	}

	// Verify each artifact's attestation
	verifier := &signature.SigstoreVerifier{}
	for _, pair := range pairs {
		result, err := verifier.Verify(c.Ctx, pair.ArtifactData, pair.SignatureData, &signature.VerifyOptions{})
		if err != nil {
			logDebugf(c, "Sigstore verification error for %s: %v", pair.ArtifactURL, err)
			// Still record the signature even if verification failed
			sig := checker.PackageSignature{
				ArtifactURL:  pair.ArtifactURL,
				SignatureURL: pair.SignatureURL,
				Type:         checker.SignatureTypeSigstore,
				IsVerified:   false,
				ErrorMsg:     err.Error(),
			}
			signatures = append(signatures, sig)
			continue
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

		signatures = append(signatures, sig)

		if result.Verified {
			logInfof(c,
				"PyPI Sigstore attestation verified for %s:%s - %s",
				pkg.Name, pkg.Version, pair.ArtifactURL)
		} else {
			logWarnf(c,
				"PyPI Sigstore attestation verification failed for %s:%s - %s",
				pkg.Name, pkg.Version, pair.ArtifactURL)
		}
	}

	return signatures
}
