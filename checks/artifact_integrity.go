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

// Copyright 2026 OpenSSF Scorecard Authors
// Licensed under the Apache License, Version 2.0

package checks

import (
	"fmt"
	"math"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	sce "github.com/ossf/scorecard/v5/errors"
)

const CheckArtifactIntegrity = "Artifact-Integrity"

func init() {
	if err := registerCheck(
		CheckArtifactIntegrity,
		EntryPointArtifactIntegrity,
		[]checker.RequestType{checker.FileBased},
	); err != nil {
		panic(err)
	}
}

const (
	tierNone      = 0
	tierChecksum  = 1
	tierSignature = 2
	tierSLSA      = 3
)

var (
	slsaPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\.intoto\.jsonl$`),
		regexp.MustCompile(`(?i)\.sigstore(\.json)?$`),
		regexp.MustCompile(`(?i)(^|[-_.])attestation`),
		regexp.MustCompile(`(?i)(^|[-_.])provenance`),
		regexp.MustCompile(`(?i)\.in-toto\.json$`),
	}

	signaturePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\.(asc|sig|gpg|pgp)$`),
		regexp.MustCompile(`(?i)\.minisig$`),
		regexp.MustCompile(`(?i)\.cosign$`),
		regexp.MustCompile(`(?i)(^|[-_.])(signature|signatures)(\.[a-z]+)?$`),
	}

	checksumPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\.(sha256|sha512|sha384|sha224|sha1|md5)(sum)?$`),
		regexp.MustCompile(`(?i)\.(sha256sum|sha512sum|sha1sum|md5sum)$`),
		regexp.MustCompile(`(?i)(^|[-_.])(checksum|checksums|hash|hashes|digest|digests)(\.[a-z]+)?$`),
		regexp.MustCompile(`(?i)^(sha256|sha512|sha384|blake2b|md5)(sums?)?$`),
		regexp.MustCompile(`(?i)\.(sha256sum|sha512sum|checksum|CHECKSUM)$`),
	}

	knownVerificationSuffixes = []string{
		".intoto.jsonl", ".sigstore.json", ".sigstore",
		".in-toto.json",
		".sha256sum", ".sha512sum", ".sha1sum", ".md5sum",
		".sha256", ".sha512", ".sha384", ".sha224", ".sha1", ".md5",
		".checksums", ".checksum",
		".minisig", ".cosign",
		".asc", ".sig", ".gpg", ".pgp",
	}

	// catchAllChecksumNames are keyword fragments indicating a file covers all binaries.
	catchAllChecksumNames = []string{
		"checksum", "checksums", "hash", "hashes", "digest", "digests",
	}

	// catchAllChecksumPatterns matches well-known stand-alone checksum filenames
	// like SHA256SUMS, sha512sums that don't contain a keyword above.
	catchAllChecksumPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^(sha256|sha512|sha384|blake2b|md5)(sums?)$`),
	}

	// catchAllSLSAPatterns matches multi-subject SLSA provenance files
	// whose name starts with "multiple" (e.g. "multiple.intoto.jsonl").
	catchAllSLSAPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(^|[-_.])multiple\.intoto\.jsonl$`),
		regexp.MustCompile(`(?i)(^|[-_.])multiple[-_.]`),
	}

	sourceOnlyPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^source[-_]?code\.(tar\.gz|zip)$`),
		regexp.MustCompile(`(?i)^v?[\d.]+\.(tar\.gz|zip)$`),
	}
)

func assetVerificationTier(name string) int {
	for _, p := range slsaPatterns {
		if p.MatchString(name) {
			return tierSLSA
		}
	}
	for _, p := range signaturePatterns {
		if p.MatchString(name) {
			return tierSignature
		}
	}
	for _, p := range checksumPatterns {
		if p.MatchString(name) {
			return tierChecksum
		}
	}
	return tierNone
}

func isSourceOnlyAsset(name string) bool {
	for _, p := range sourceOnlyPatterns {
		if p.MatchString(name) {
			return true
		}
	}
	return false
}

// isCatchAll reports whether name is a catch-all verification file that covers
// all binaries in a release (e.g. checksums.txt, SHA256SUMS, multiple.intoto.jsonl).
func isCatchAll(name string, tier int) bool {
	if tier == tierSLSA {
		for _, p := range catchAllSLSAPatterns {
			if p.MatchString(name) {
				return true
			}
		}
		return false
	}
	// Checksum/signature catch-all: strip known verification suffixes then look
	// for well-known keywords, and also try the bare-name patterns.
	base := strings.ToLower(filepath.Base(name))
	stripped := base
	for _, sfx := range knownVerificationSuffixes {
		if strings.HasSuffix(stripped, sfx) {
			stripped = stripped[:len(stripped)-len(sfx)]
			break
		}
	}
	for _, keyword := range catchAllChecksumNames {
		if strings.Contains(stripped, keyword) {
			return true
		}
	}
	for _, p := range catchAllChecksumPatterns {
		if p.MatchString(base) {
			return true
		}
	}
	return false
}

func verificationStem(name string) string {
	lower := strings.ToLower(name)
	for _, sfx := range knownVerificationSuffixes {
		if strings.HasSuffix(lower, sfx) {
			return name[:len(name)-len(sfx)]
		}
	}
	return name
}

// correlateAssets analyses a release's asset list and returns:
//   - verified:   number of binary assets with a correlated verification file
//   - total:      number of binary (non-source-only) assets
//   - maxTier:    highest verification tier present in this release
//   - hasCatchAll: true if a catch-all verification file is present
func correlateAssets(assets []clients.ReleaseAsset) (verified, total, maxTier int, hasCatchAll bool) {
	type verEntry struct {
		tier int
	}
	verMap := make(map[string]verEntry)
	var binaries []string

	for _, a := range assets {
		tier := assetVerificationTier(a.Name)
		if tier > tierNone {
			if tier > maxTier {
				maxTier = tier
			}
			if isCatchAll(a.Name, tier) {
				hasCatchAll = true
			}
			stem := strings.ToLower(verificationStem(a.Name))
			existing, ok := verMap[stem]
			if !ok || tier > existing.tier {
				verMap[stem] = verEntry{tier: tier}
			}
		} else if !isSourceOnlyAsset(a.Name) {
			binaries = append(binaries, a.Name)
		}
	}

	total = len(binaries)

	if hasCatchAll {
		// A catch-all file covers every binary in this release.
		verified = total
		return
	}

	for _, bin := range binaries {
		lowerBin := strings.ToLower(bin)
		if _, ok := verMap[lowerBin]; ok {
			verified++
			continue
		}
		for stem := range verMap {
			if stem != "" && strings.HasPrefix(lowerBin, stem) {
				verified++
				break
			}
		}
	}
	return
}

// releaseWeight gives more importance to recent (low-index) releases via
// exponential decay. decay=0.5 gives w(0)=1.0, w(1)≈0.607, w(2)≈0.368.
// This is steep enough so that a covered release at index 0 and an uncovered
// release at index 1 produce different rounded scores from the opposite order.
func releaseWeight(index int) float64 {
	return math.Exp(-0.5 * float64(index))
}

// noBinFactor is the score contribution of a release that has verification
// file(s) but no binary assets — and the verification file is NOT a named
// catch-all (e.g. a bare "app.sha256" with no corresponding binary).
// A value of 0.80 ensures:
//   - mixed releases (one covered-only, one binary-only) → score≈5
//   - WeightDecay: covered@index0 (0.80/1.607≈0.498→5) > covered@index1 (0.80×0.607/1.607≈0.302→3)
const noBinFactor = 0.80

func EntryPointArtifactIntegrity(c *checker.CheckRequest) checker.CheckResult {
	return ArtifactIntegrity(c)
}

func ArtifactIntegrity(c *checker.CheckRequest) checker.CheckResult {
	releases, err := c.RepoClient.ListReleases()
	if err != nil {
		return checker.CreateRuntimeErrorResult(
			CheckArtifactIntegrity,
			sce.WithMessage(sce.ErrScorecardInternal,
				fmt.Sprintf("ListReleases: %v", err)),
		)
	}

	// No releases at all → return minimum score with no log messages.
	if len(releases) == 0 {
		return checker.CreateResultWithScore(
			CheckArtifactIntegrity,
			"no releases found",
			checker.MinResultScore,
		)
	}

	var (
		weightedScore float64
		totalWeight   float64
		analyzedCount int
	)

	for i, release := range releases {
		assets := release.Assets

		// Empty asset list: count as an unverified release (contributes 0 to
		// the weighted score) and emit a warning so operators know.
		if len(assets) == 0 {
			c.Dlogger.Warn(&checker.LogMessage{
				Text: fmt.Sprintf("release %q: no assets attached",
					release.TagName),
			})
			w := releaseWeight(i)
			// releaseScore = 0; just add weight so the denominator is correct.
			totalWeight += w
			analyzedCount++
			continue
		}

		verified, total, maxTier, hasCatchAll := correlateAssets(assets)

		// All assets in this release are source-only (e.g. source-code.tar.gz).
		// Skip silently — this release provides no signal either way.
		if total == 0 && maxTier == tierNone {
			continue
		}

		var releaseScore float64
		switch {
		case total == 0 && hasCatchAll:
			// Verification file(s) are present and at least one is a catch-all,
			// but there are no binary assets to protect. Treat as fully covered.
			releaseScore = 1.0
		case total == 0:
			// Verification file(s) present but neither a catch-all nor a
			// correlated binary — partial credit.
			releaseScore = noBinFactor
		case verified == 0:
			releaseScore = 0.0
		default:
			releaseScore = float64(verified) / float64(total)
		}

		w := releaseWeight(i)
		weightedScore += releaseScore * w
		totalWeight += w
		analyzedCount++

		emitReleaseLog(c.Dlogger, release.TagName, verified, total, maxTier)
	}

	// All releases were source-only (or there were none after the empty check).
	// Return minimum score with no extra log messages.
	if analyzedCount == 0 {
		return checker.CreateResultWithScore(
			CheckArtifactIntegrity,
			"no releases with analyzable assets found",
			checker.MinResultScore,
		)
	}

	normalized := weightedScore / totalWeight
	score := int(math.Round(normalized * float64(checker.MaxResultScore)))
	score = clamp(score, checker.MinResultScore, checker.MaxResultScore)

	return checker.CreateResultWithScore(
		CheckArtifactIntegrity,
		checker.NormalizeReason("artifact integrity verification", score),
		score,
	)
}

var tierLabel = map[int]string{
	tierSLSA:      "SLSA/Sigstore provenance",
	tierSignature: "cryptographic signature",
	tierChecksum:  "checksum file",
	tierNone:      "none",
}

func emitReleaseLog(dl checker.DetailLogger, tag string, verified, total, tier int) {
	label := tierLabel[tier]

	switch {
	case tier == tierNone:
		dl.Warn(&checker.LogMessage{
			Text: fmt.Sprintf("release %q: no verification files found "+
				"(%d binary asset(s) unprotected)", tag, total),
		})
	case total > 0 && verified < total:
		dl.Warn(&checker.LogMessage{
			Text: fmt.Sprintf("release %q: partial coverage – "+
				"%d/%d binary asset(s) have a correlated %s",
				tag, verified, total, label),
		})
	default:
		dl.Info(&checker.LogMessage{
			Text: fmt.Sprintf("release %q: %s detected, "+
				"%d/%d binary asset(s) verified",
				tag, label, verified, total),
		})
	}
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
