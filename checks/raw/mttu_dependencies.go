// Copyright 2025 OpenSSF Scorecard Authors
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

// Package raw implements raw data collectors for Scorecard checks.
package raw

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/mod/semver"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	depsdev "github.com/ossf/scorecard/v5/internal/packageclient"
)

const (
	ecosystemGo       = "GO"
	ecosystemNPM      = "NPM"
	ecosystemPyPI     = "PYPI"
	ecosystemMaven    = "MAVEN"
	ecosystemCargo    = "CARGO"
	ecosystemNuGet    = "NUGET"
	ecosystemRubyGems = "RUBYGEMS"
)

// MTTUDependencies uses osv-scalibr to discover direct dependencies in the repository,
// queries deps.dev to check for newer versions, and annotates dependencies with
// IsLatest and TimeSinceOldestReleast (mean time to update).
//
//nolint:gocognit // Complex business logic with multiple dependency handling paths
func MTTUDependencies(c *checker.CheckRequest) (checker.MTTUDependenciesData, error) {
	var data checker.MTTUDependenciesData

	// Get the local path to the repository.
	localPath, err := c.RepoClient.LocalPath()
	if err != nil {
		return data, fmt.Errorf("getting local path: %w", err)
	}

	// Create context if not provided.
	ctx := c.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

	// Use osv-scalibr to scan for direct dependencies.
	// Note: The scalibr client is configured with DirectFS capability and
	// filters out transitive/indirect dependencies based on metadata.
	depsClient := clients.NewDirectDepsClient()
	depsResp, err := depsClient.GetDeps(ctx, localPath)
	if err != nil {
		return data, fmt.Errorf("scanning dependencies with scalibr: %w", err)
	}

	// Nothing to do if no dependencies found.
	if len(depsResp.Deps) == 0 {
		if c.Dlogger != nil {
			c.Dlogger.Info(&checker.LogMessage{
				Text: "No dependencies found by scalibr",
			})
		}
		return data, nil
	}

	// Log summary of what was found
	ecosystemCounts := make(map[string]int)
	for _, dep := range depsResp.Deps {
		ecosystemCounts[dep.Ecosystem]++
	}

	if c.Dlogger != nil {
		c.Dlogger.Info(&checker.LogMessage{
			Text: fmt.Sprintf("Detected %d total packages from repository", len(depsResp.Deps)),
		})

		for eco, count := range ecosystemCounts {
			c.Dlogger.Debug(&checker.LogMessage{
				Text: fmt.Sprintf("  - %s: %d packages", eco, count),
			})
		}
	}

	// Build deps.dev client.
	httpClient := &http.Client{Timeout: 15 * time.Second}
	dd := depsdev.NewClient(httpClient)

	// Annotate each dependency with IsLatest and TimeSinceOldestReleast.
	now := time.Now().UTC()

	// Cache to avoid duplicate queries to deps.dev for the same package.
	// Key format: "SYSTEM||NAME" (e.g., "NPM||express")
	packageCache := make(map[string]*depsdev.Package)

	for _, dep := range depsResp.Deps {
		if dep.Name == "" || dep.Version == "" {
			continue
		}

		// Convert scalibr ecosystem to deps.dev system name.
		systemName := scalibrEcosystemToDepsDevSystem(dep.Ecosystem)
		if systemName == "" {
			// Unknown ecosystem, skip.
			continue
		}

		// Build cache key.
		cacheKey := systemName + "||" + dep.Name

		// Check cache first to avoid duplicate queries.
		pkg, cached := packageCache[cacheKey]
		if !cached {
			// Query deps.dev for all versions of this package.
			var err error
			pkg, err = dd.GetPackage(ctx, systemName, dep.Name)
			if err != nil || pkg == nil || len(pkg.Versions) == 0 {
				// Cache nil result to avoid retrying failed queries.
				packageCache[cacheKey] = nil
				continue
			}
			// Cache successful result.
			packageCache[cacheKey] = pkg
		}

		// Skip if package data is unavailable (cached failure).
		if pkg == nil || len(pkg.Versions) == 0 {
			continue
		}

		isLatest, oldestNewerPublishedAt := newestInfoFor(dep.Version, pkg.Versions)

		// Create LockDependency from scalibr Dep.
		lockDep := checker.LockDependency{
			Name:      dep.Name,
			Version:   dep.Version,
			Ecosystem: scalibrEcosystemToCheckerEcosystem(dep.Ecosystem),
			IsLatest:  ptrBool(isLatest),
		}

		// If not latest, encode the "time since oldest newer release".
		if oldestNewerPublishedAt != nil {
			delta := now.Sub(*oldestNewerPublishedAt)
			if delta < 0 {
				delta = 0
			}
			lockDep.TimeSinceOldestReleast = time.Unix(0, 0).UTC().Add(delta)
		}

		data.Dependencies = append(data.Dependencies, lockDep)
	}

	// Count how many are up-to-date vs outdated
	upToDate := 0
	outdated := 0
	for _, dep := range data.Dependencies {
		if dep.IsLatest != nil && *dep.IsLatest {
			upToDate++
		} else {
			outdated++
		}
	}

	if len(data.Dependencies) > 0 && c.Dlogger != nil {
		c.Dlogger.Info(&checker.LogMessage{
			Text: fmt.Sprintf("Summary: %d dependencies up-to-date, %d outdated", upToDate, outdated),
		})
	}

	return data, nil
}

// scalibrEcosystemToDepsDevSystem maps scalibr ecosystem names to deps.dev system names.
func scalibrEcosystemToDepsDevSystem(eco string) string {
	switch strings.ToLower(strings.TrimSpace(eco)) {
	case "golang", "go":
		return ecosystemGo
	case "npm":
		return ecosystemNPM
	case "pypi":
		return ecosystemPyPI
	case "maven":
		return ecosystemMaven
	case "cargo":
		return ecosystemCargo
	case "nuget":
		return ecosystemNuGet
	case "gem":
		return ecosystemRubyGems
	default:
		return ""
	}
}

// scalibrEcosystemToCheckerEcosystem maps scalibr ecosystem names to checker.Ecosystem.
func scalibrEcosystemToCheckerEcosystem(eco string) checker.Ecosystem {
	switch strings.ToLower(strings.TrimSpace(eco)) {
	case "golang", "go":
		return checker.EcosystemGo
	case "npm":
		return checker.EcosystemNPM
	case "pypi":
		return checker.EcosystemPypi
	case "maven":
		return checker.EcosystemMaven
	case "cargo":
		return checker.EcosystemCargo
	case "nuget":
		return checker.EcosystemNuget
	case "gem":
		return checker.EcosystemRubyGems
	default:
		return checker.Ecosystem("")
	}
}

func ptrBool(b bool) *bool { return &b }

// isPseudoVersion checks if a version is a Go pseudo-version.
// Pseudo-versions represent unreleased commits and follow the pattern:
// vX.Y.Z-yyyymmddhhmmss-abcdefabcdef or vX.Y.Z-0.yyyymmddhhmmss-abcdefabcdef
// or vX.Y.Z-pre.0.yyyymmddhhmmss-abcdefabcdef
// Examples: v1.2.1-0.20250916002408-014fb9c9e8f7
// We want to exclude these and only consider actual tagged releases.
func isPseudoVersion(version string) bool {
	// Pseudo-versions contain a timestamp in the format: -yyyymmddhhmmss-
	// The canonical format includes: -0.yyyymmddhhmmss- or -pre.0.yyyymmddhhmmss-
	// Look for a sequence of exactly 14 digits preceded and followed by non-digit characters
	if len(version) < 25 { // Minimum length for a pseudo-version
		return false
	}

	// Look for the timestamp pattern: 14 consecutive digits
	// In practice, these appear as .(14 digits)- in the version string
	//nolint:nestif // Timestamp validation requires multiple nested checks
	for i := 0; i < len(version)-15; i++ {
		// Check if we have 14 consecutive digits
		if version[i] >= '0' && version[i] <= '9' {
			allDigits := true
			digitCount := 0
			for j := 0; j < 14 && i+j < len(version); j++ {
				if version[i+j] >= '0' && version[i+j] <= '9' {
					digitCount++
				} else {
					allDigits = false
					break
				}
			}
			// If we found exactly 14 consecutive digits, and there's a dash after them,
			// and there was a dot or dash before them, it's likely a pseudo-version timestamp
			if allDigits && digitCount == 14 && i+14 < len(version) && version[i+14] == '-' {
				// Check if there's a dot before the digits (common pattern: .20250916002408-)
				if i > 0 && (version[i-1] == '.' || version[i-1] == '-') {
					return true
				}
			}
		}
	}
	return false
}

// newestInfoFor determines:
//  1. Whether currentVersion is the latest release among versions.
//  2. If not, returns the PublishedAt time of the OLDEST version that is strictly newer.
//
// Primary ordering is semantic versioning using golang.org/x/mod/semver (after normalizing to
// a "v" prefix). Only considers actual tagged releases, excluding Go pseudo-versions which
// represent unreleased commits.
//
//nolint:gocognit // Complex version comparison logic with multiple edge cases and fallbacks
func newestInfoFor(currentVersion string, versions []depsdev.Version) (bool, *time.Time) {
	normalize := func(v string) string {
		if v == "" {
			return v
		}
		if v[0] != 'v' {
			return "v" + v
		}
		return v
	}

	curSem := normalize(currentVersion)
	curIsSem := semver.IsValid(curSem)

	var (
		latestSemver      string
		oldestNewerSemver string
		oldestNewerPub    time.Time
		latestPub         time.Time
	)

	// First pass: try to use semver ordering wherever possible.
	// IMPORTANT: Skip pseudo-versions - only consider actual tagged releases.
	for i := range versions {
		v := &versions[i]
		pub, err := time.Parse(time.RFC3339, v.PublishedAt)
		if err != nil {
			continue
		}
		vs := normalize(v.VersionKey.Version)

		// Skip pseudo-versions - they represent unreleased commits, not official releases
		if isPseudoVersion(vs) {
			continue
		}

		//nolint:nestif // Version comparison logic requires nested checks for accuracy
		if semver.IsValid(vs) {
			// Track the overall latest by semver
			if latestSemver == "" || semver.Compare(vs, latestSemver) > 0 {
				latestSemver = vs
				latestPub = pub
			}

			// Track the OLDEST version that is newer than current
			if curIsSem && semver.Compare(vs, curSem) > 0 {
				// This version is newer than current
				if oldestNewerSemver == "" || semver.Compare(vs, oldestNewerSemver) < 0 {
					oldestNewerSemver = vs
					oldestNewerPub = pub
				}
			}
		} else if latestPub.IsZero() || pub.After(latestPub) {
			// Track latest by published time as a fallback signal.
			latestPub = pub
			latestSemver = "" // signal: relying on time-based latest
		}
	}

	// If we have valid semver for both current and at least one release, use semver result.
	if latestSemver != "" && curIsSem {
		isLatest := semver.Compare(curSem, latestSemver) == 0
		if isLatest {
			return true, nil
		}
		// Return the oldest newer version's timestamp if we found one
		if !oldestNewerPub.IsZero() {
			return false, &oldestNewerPub
		}
		// Fallback to latest if no specific older-newer found
		if !latestPub.IsZero() {
			return false, &latestPub
		}
		// No latest found (odd), but not equal to latest -> not latest, no timestamp.
		return false, nil
	}
	// Pure timestamp fallback: determine if current is latest by publish time.
	var curPub time.Time
	for i := range versions {
		if versions[i].VersionKey.Version == currentVersion {
			var err error
			curPub, err = time.Parse(time.RFC3339, versions[i].PublishedAt)
			if err != nil {
				continue
			}
			break
		}
	}
	isLatest := !curPub.IsZero() && (curPub.Equal(latestPub) || curPub.After(latestPub))
	if isLatest {
		return true, nil
	}

	if !latestPub.IsZero() {
		return false, &latestPub
	}
	return false, nil
}
