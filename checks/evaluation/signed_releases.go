// Copyright 2021 OpenSSF Scorecard Authors
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

package evaluation

import (
	"errors"
	"fmt"
	"math"

	"github.com/ossf/scorecard/v5/checker"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/releasesAreSigned"
	"github.com/ossf/scorecard/v5/probes/releasesHaveProvenance"
	"github.com/ossf/scorecard/v5/probes/releasesHaveVerifiedProvenance"
	"github.com/ossf/scorecard/v5/probes/releasesHaveVerifiedSignatures"
)

var errNoReleaseFound = errors.New("no release found")

type releaseState struct {
	verifiedSignatures        map[string]bool
	failedSignatures          map[string]bool
	verifiedProvenance        map[string]bool // GitHub release provenance (detected, not verified)
	verifiedPackageProvenance map[string]bool // Package provenance from deps.dev (actually verified)
	detectedSignatures        map[string]bool
	loggedReleases            []string
}

// validateProbes checks that all expected probes are present in findings.
// isPackageLevelProbe returns true if the probe operates on package manager releases
// rather than GitHub releases. Package-level probes don't have release names.
func isPackageLevelProbe(probeName string) bool {
	return probeName == releasesHaveVerifiedSignatures.Probe ||
		probeName == releasesHaveVerifiedProvenance.Probe
}

// isNewRelease returns true if this is the first time we've seen this release name.
func isNewRelease(releaseName string, loggedReleases []string) bool {
	return releaseName != "" && !contains(loggedReleases, releaseName)
}

func validateProbes(name string, findings []finding.Finding) error {
	expectedProbes := []string{
		releasesAreSigned.Probe,
		releasesHaveProvenance.Probe,
		releasesHaveVerifiedSignatures.Probe,
		releasesHaveVerifiedProvenance.Probe,
	}

	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		return sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
	}
	return nil
}

// processFindings extracts release state information from findings.
func processFindings(findings []finding.Finding, dl checker.DetailLogger) (*releaseState, error) {
	state := &releaseState{
		verifiedSignatures:        make(map[string]bool),
		failedSignatures:          make(map[string]bool),
		verifiedProvenance:        make(map[string]bool),
		verifiedPackageProvenance: make(map[string]bool),
		detectedSignatures:        make(map[string]bool),
		loggedReleases:            make([]string, 0),
	}

	for i := range findings {
		f := &findings[i]

		if f.Outcome == finding.OutcomeNotApplicable {
			continue
		}

		releaseName := getReleaseName(f)
		// GitHub release probes must have a release name; package probes may not
		if releaseName == "" && !isPackageLevelProbe(f.Probe) {
			return nil, errNoReleaseFound
		}

		// Log each GitHub release the first time we encounter it
		if isNewRelease(releaseName, state.loggedReleases) {
			dl.Debug(&checker.LogMessage{
				Text: fmt.Sprintf("GitHub release found: %s", releaseName),
			})
			state.loggedReleases = append(state.loggedReleases, releaseName)
		}

		// Process each probe's findings:
		// - OutcomeTrue means the check passed (signature verified, provenance found, etc.)
		// - OutcomeFalse means the check failed (verification failed, no signature, etc.)
		switch f.Probe {
		case releasesHaveVerifiedSignatures.Probe:
			// Track cryptographic verification of package signatures (Maven GPG, PyPI Sigstore)
			pkgKey := getPackageKey(f)
			if f.Outcome == finding.OutcomeTrue {
				// Signature was cryptographically verified
				state.verifiedSignatures[pkgKey] = true
			} else if f.Outcome == finding.OutcomeFalse {
				// Signature verification failed (wrong key, invalid sig, etc.)
				state.failedSignatures[pkgKey] = true
			}
		case releasesHaveVerifiedProvenance.Probe:
			// Track cryptographic verification of provenance from deps.dev
			pkgKey := getPackageKey(f)
			if f.Outcome == finding.OutcomeTrue {
				// Provenance was verified by deps.dev
				state.verifiedPackageProvenance[pkgKey] = true
			}
		case releasesHaveProvenance.Probe:
			// Track presence of provenance attestation files in GitHub releases
			if f.Outcome == finding.OutcomeTrue {
				// Release has a provenance file (.intoto.jsonl)
				state.verifiedProvenance[releaseName] = true
			}
		case releasesAreSigned.Probe:
			// Track detection of signature files in GitHub releases (not verified yet)
			if f.Outcome == finding.OutcomeTrue {
				// Release has a signature file (.asc, .sig, .minisig, etc.)
				state.detectedSignatures[releaseName] = true
			}
		}
	}

	return state, nil
}

// scoreWithVerificationDeductions uses the original detection-based scoring
// for GitHub releases, but deducts 1 point for each package release that failed verification.
// This way, verification failures only affect the signature score, not the provenance score.
func scoreWithVerificationDeductions(
	name string,
	state *releaseState,
	findings []finding.Finding,
	dl checker.DetailLogger,
) checker.CheckResult {
	// Build a release map using the original logic for GitHub releases
	releaseMap := make(map[string]int)
	hasProvenance := make(map[string]bool)
	hasGitHubReleases := len(state.loggedReleases) > 0

	// Process GitHub release findings to build the base score
	for i := range findings {
		f := &findings[i]

		// Skip package-level probes for the base score calculation
		if isPackageLevelProbe(f.Probe) {
			continue
		}

		releaseName := getReleaseName(f)
		if releaseName == "" {
			continue
		}

		// Track provenance
		if f.Probe == releasesHaveProvenance.Probe && f.Outcome == finding.OutcomeTrue {
			hasProvenance[releaseName] = true
		}

		// Score each release using original logic
		if f.Outcome == finding.OutcomeTrue {
			switch f.Probe {
			case releasesAreSigned.Probe:
				// Only set to 8 if not already set (provenance takes precedence)
				if _, ok := releaseMap[releaseName]; !ok {
					releaseMap[releaseName] = 8
				}
			case releasesHaveProvenance.Probe:
				// Provenance gets max score and overrides signature score
				releaseMap[releaseName] = 10
			}
		}
	}

	// Log all findings
	for i := range findings {
		f := &findings[i]
		relName := getReleaseName(f)

		var logLevel checker.DetailType
		switch f.Outcome {
		case finding.OutcomeTrue:
			logLevel = checker.DetailInfo
		case finding.OutcomeFalse:
			// For GitHub releases: Don't warn about missing signature if provenance exists
			if f.Probe == releasesAreSigned.Probe && hasProvenance[relName] {
				logLevel = checker.DetailDebug
			} else {
				logLevel = checker.DetailWarn
			}
		default:
			logLevel = checker.DetailDebug
		}
		checker.LogFinding(dl, f, logLevel)
	}

	// If we only have package verification (no GitHub releases), use pure package scoring
	if !hasGitHubReleases {
		return scorePackageVerificationOnly(name, state, dl)
	}

	totalReleases := len(state.loggedReleases)
	if totalReleases > 5 {
		err := sce.CreateInternal(sce.ErrScorecardInternal, "too many releases, please report this")
		return checker.CreateRuntimeErrorResult(name, err)
	}
	if totalReleases == 0 {
		return checker.CreateInconclusiveResult(name, "no releases found")
	}

	// Calculate base score from GitHub releases
	score := calculateBaseScore(releaseMap)

	// Apply deductions for failed package verifications
	failedCount := len(state.failedSignatures)
	if failedCount > 0 {
		score = applyVerificationDeductions(releaseMap, score, failedCount)
		dl.Warn(&checker.LogMessage{
			Text: fmt.Sprintf("%d package release(s) failed cryptographic verification", failedCount),
		})
	}

	// Calculate final average score
	if totalReleases > 0 {
		score = int(math.Floor(float64(score) / float64(totalReleases)))
	}

	// Build reason string and return result
	return buildVerificationResult(name, releaseMap, totalReleases, failedCount, score)
}

// calculateBaseScore sums up scores from the release map.
func calculateBaseScore(releaseMap map[string]int) int {
	score := 0
	for _, s := range releaseMap {
		score += s
	}
	return score
}

// applyVerificationDeductions deducts points for failed verifications.
func applyVerificationDeductions(releaseMap map[string]int, score, failedCount int) int {
	// For releases with signatures (8 points), deduct 1 per failed verification
	deduction := failedCount
	for relName, releaseScore := range releaseMap {
		if releaseScore == 8 && deduction > 0 {
			// Deduct 1 point from this release's signature score
			releaseMap[relName] = 7
			deduction--
			if deduction == 0 {
				break
			}
		}
	}

	// If there are more failures than releases with signatures,
	// apply remaining deductions to overall score
	if deduction > 0 {
		score -= deduction
		if score < 0 {
			score = 0
		}
	} else {
		// Recalculate score after per-release deductions
		score = calculateBaseScore(releaseMap)
	}

	return score
}

// buildVerificationResult creates the final check result.
func buildVerificationResult(
	name string,
	releaseMap map[string]int,
	totalReleases, failedCount, score int,
) checker.CheckResult {
	// Count provenance releases
	provenanceCount := 0
	for _, s := range releaseMap {
		if s == 10 {
			provenanceCount++
		}
	}

	if len(releaseMap) == 0 {
		return checker.CreateMinScoreResult(name, "no signatures or provenance found")
	}

	if provenanceCount == totalReleases && failedCount == 0 {
		return checker.CreateMaxScoreResult(name, "all releases have provenance")
	}

	// Build reason string
	var reason string
	if failedCount > 0 {
		reason = fmt.Sprintf("%d out of %d releases have signatures/provenance (%d verification failures)",
			len(releaseMap), totalReleases, failedCount)
	} else {
		reason = fmt.Sprintf("%d out of %d releases have signatures or provenance",
			len(releaseMap), totalReleases)
	}

	return checker.CreateResultWithScore(name, reason, score)
}

// scorePackageVerificationOnly handles scoring when there are only package-level
// verifications (Maven/PyPI) and no GitHub releases.
func scorePackageVerificationOnly(
	name string,
	state *releaseState,
	dl checker.DetailLogger,
) checker.CheckResult {
	// Count unique packages
	uniquePackages := make(map[string]bool)
	for k := range state.verifiedSignatures {
		uniquePackages[k] = true
	}
	for k := range state.verifiedPackageProvenance {
		uniquePackages[k] = true
	}
	for k := range state.failedSignatures {
		uniquePackages[k] = true
	}

	totalPackages := len(uniquePackages)
	if totalPackages == 0 {
		return checker.CreateInconclusiveResult(name, "no releases found")
	}

	// Count verified packages (either signature or provenance)
	verifiedPackages := make(map[string]bool)
	for k := range state.verifiedSignatures {
		verifiedPackages[k] = true
	}
	for k := range state.verifiedPackageProvenance {
		verifiedPackages[k] = true
	}
	verifiedCount := len(verifiedPackages)
	failedCount := len(state.failedSignatures)

	// All packages verified successfully
	if verifiedCount == totalPackages && failedCount == 0 {
		return checker.CreateMaxScoreResult(name,
			"all package releases have verified signatures or provenance")
	}

	// All verification attempts failed
	if failedCount > 0 && verifiedCount == 0 {
		dl.Warn(&checker.LogMessage{
			Text: "signatures found but verification failed - possible security issue",
		})
		return checker.CreateResultWithScore(name,
			"signatures present but verification failed for all releases",
			3)
	}

	// Partial verification - proportional scoring with cap at 7 if there are failures
	if verifiedCount > 0 {
		score := int(math.Round(10.0 * float64(verifiedCount) / float64(totalPackages)))
		if failedCount > 0 && score > 7 {
			score = 7
		}
		return checker.CreateResultWithScore(name,
			fmt.Sprintf("%d verified, %d failed out of %d package releases",
				verifiedCount, failedCount, totalPackages),
			score)
	}

	return checker.CreateMinScoreResult(name, "no verified signatures found")
}

// SignedReleases applies the score policy for the Signed-Releases check.
func SignedReleases(name string,
	findings []finding.Finding, dl checker.DetailLogger,
) checker.CheckResult {
	if err := validateProbes(name, findings); err != nil {
		return checker.CreateRuntimeErrorResult(name, err)
	}

	state, err := processFindings(findings, dl)
	if err != nil {
		return checker.CreateRuntimeErrorResult(name, err)
	}

	hasVerificationAttempts := len(state.verifiedSignatures) > 0 ||
		len(state.failedSignatures) > 0 ||
		len(state.verifiedPackageProvenance) > 0

	if !hasVerificationAttempts {
		return scoreBasedOnDetection(
			name,
			findings,
			dl,
			state.detectedSignatures,
			state.verifiedProvenance,
			state.loggedReleases,
		)
	}

	return scoreWithVerificationDeductions(name, state, findings, dl)
}

// scoreBasedOnDetection implements the scoring logic for cases
// where no verification was attempted (GitHub releases only, no Maven/PyPI).
func scoreBasedOnDetection(
	name string,
	findings []finding.Finding,
	dl checker.DetailLogger,
	detectedSignatures map[string]bool,
	verifiedProvenance map[string]bool,
	loggedReleases []string,
) checker.CheckResult {
	// Log all findings:
	// - True outcomes are Info
	// - False outcomes:
	//   * Missing signature but has provenance: Debug (provenance is sufficient)
	//   * Missing provenance but has signature: Warn (provenance is better)
	//   * Missing both: Warn
	for i := range findings {
		f := &findings[i]
		relName := getReleaseName(f)

		var logLevel checker.DetailType
		switch f.Outcome {
		case finding.OutcomeTrue:
			logLevel = checker.DetailInfo
		case finding.OutcomeFalse:
			// Missing signature but has provenance: OK
			if f.Probe == releasesAreSigned.Probe && verifiedProvenance[relName] {
				logLevel = checker.DetailDebug
			} else {
				// Missing provenance or missing both: Warn
				logLevel = checker.DetailWarn
			}
		default:
			logLevel = checker.DetailDebug
		}
		checker.LogFinding(dl, f, logLevel)
	}

	// Count unique releases
	uniqueReleases := make(map[string]bool)
	for k := range detectedSignatures {
		uniqueReleases[k] = true
	}
	for k := range verifiedProvenance {
		uniqueReleases[k] = true
	}

	// Total releases analyzed = all releases we logged
	totalReleases := len(loggedReleases)

	// TODO, the evaluation code should be the one limiting to 5, not assuming the probes have done it already
	// however there are some ordering issues to consider, so going with the easy way for now
	if totalReleases > 5 {
		err := sce.CreateInternal(sce.ErrScorecardInternal, "too many releases, please report this")
		return checker.CreateRuntimeErrorResult(name, err)
	}
	if totalReleases == 0 {
		// This should not happen in production, but it is useful to have
		// for testing.
		return checker.CreateInconclusiveResult(name, "no releases found")
	}

	// Count how many releases have signature or provenance
	signedOrProvenance := len(uniqueReleases)

	// All releases have provenance - best score
	if len(verifiedProvenance) == totalReleases {
		return checker.CreateMaxScoreResult(name,
			"all releases have provenance")
	}

	// All releases have either signatures or provenance
	if signedOrProvenance == totalReleases {
		return checker.CreateResultWithScore(name,
			"all releases are signed",
			8)
	}

	// Partial coverage - use proportional scoring (max 9 for detection-based scoring)
	if signedOrProvenance > 0 {
		score := 9 * signedOrProvenance / totalReleases
		return checker.CreateResultWithScore(name,
			fmt.Sprintf("%d out of %d releases have signatures or provenance",
				signedOrProvenance, totalReleases),
			score)
	}

	// No signatures or provenance
	return checker.CreateMinScoreResult(name, "no signatures or provenance found")
}

func getPackageKey(f *finding.Finding) string {
	system := f.Values["packageSystem"]
	name := f.Values["packageName"]
	version := f.Values["packageVersion"]
	return fmt.Sprintf("%s:%s:%s", system, name, version)
}

func getReleaseName(f *finding.Finding) string {
	var key string
	// these keys should be the same, but might as handle situations when they're not
	switch f.Probe {
	case releasesAreSigned.Probe:
		key = releasesAreSigned.ReleaseNameKey
	case releasesHaveProvenance.Probe:
		key = releasesHaveProvenance.ReleaseNameKey
	}
	return f.Values[key]
}

func contains(releases []string, release string) bool {
	for _, r := range releases {
		if r == release {
			return true
		}
	}
	return false
}
