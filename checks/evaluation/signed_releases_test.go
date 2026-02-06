//  Copyright 2023 OpenSSF Scorecard Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package evaluation

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ossf/scorecard/v5/checker"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/releasesAreSigned"
	"github.com/ossf/scorecard/v5/probes/releasesHaveProvenance"
	"github.com/ossf/scorecard/v5/probes/releasesHaveVerifiedProvenance"
	"github.com/ossf/scorecard/v5/probes/releasesHaveVerifiedSignatures"
	scut "github.com/ossf/scorecard/v5/utests"
)

const (
	release0 = 0
	release1 = 1
	release2 = 2
	release3 = 3
	release4 = 4
	release5 = 5
)

const (
	asset0 = 0
	asset1 = 1
	asset2 = 2
	asset3 = 3
)

func signedProbe(release, asset int, outcome finding.Outcome) finding.Finding {
	return finding.Finding{
		Probe:   releasesAreSigned.Probe,
		Outcome: outcome,
		Values: map[string]string{
			releasesAreSigned.ReleaseNameKey: fmt.Sprintf("v%d", release),
			releasesAreSigned.AssetNameKey:   fmt.Sprintf("artifact-%d", asset),
		},
	}
}

func provenanceProbe(release, asset int, outcome finding.Outcome) finding.Finding {
	return finding.Finding{
		Probe:   releasesHaveProvenance.Probe,
		Outcome: outcome,
		Values: map[string]string{
			releasesHaveProvenance.ReleaseNameKey: fmt.Sprintf("v%d", release),
			releasesHaveProvenance.AssetNameKey:   fmt.Sprintf("artifact-%d", asset),
		},
	}
}

func verifiedSignaturesProbe(outcome finding.Outcome) finding.Finding {
	return finding.Finding{
		Probe:   releasesHaveVerifiedSignatures.Probe,
		Outcome: outcome,
		Values:  map[string]string{},
	}
}

func verifiedSignaturesProbeWithPackage(outcome finding.Outcome, system, name, version string) finding.Finding {
	return finding.Finding{
		Probe:   releasesHaveVerifiedSignatures.Probe,
		Outcome: outcome,
		Values: map[string]string{
			"packageSystem":  system,
			"packageName":    name,
			"packageVersion": version,
		},
	}
}

func verifiedProvenanceProbe(outcome finding.Outcome) finding.Finding {
	return finding.Finding{
		Probe:   releasesHaveVerifiedProvenance.Probe,
		Outcome: outcome,
		Values:  map[string]string{},
	}
}

func verifiedProvenanceProbeWithPackage(outcome finding.Outcome, system, name, version string) finding.Finding {
	return finding.Finding{
		Probe:   releasesHaveVerifiedProvenance.Probe,
		Outcome: outcome,
		Values: map[string]string{
			"packageSystem":  system,
			"packageName":    name,
			"packageVersion": version,
		},
	}
}

func TestSignedReleases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		findings []finding.Finding
		result   scut.TestReturn
	}{
		{
			name: "Has one release that is signed but no provenance",
			findings: []finding.Finding{
				signedProbe(0, 0, finding.OutcomeTrue),
				provenanceProbe(0, 0, finding.OutcomeFalse),
				verifiedSignaturesProbe(finding.OutcomeNotApplicable),
				verifiedProvenanceProbe(finding.OutcomeNotApplicable),
			},
			result: scut.TestReturn{
				Score:         8,
				NumberOfInfo:  1,
				NumberOfWarn:  1,
				NumberOfDebug: 3,
			},
		},
		{
			name: "Has one release that is signed and has provenance",
			findings: []finding.Finding{
				signedProbe(0, 0, finding.OutcomeTrue),
				provenanceProbe(0, 0, finding.OutcomeTrue),
				verifiedSignaturesProbe(finding.OutcomeNotApplicable),
				verifiedProvenanceProbe(finding.OutcomeNotApplicable),
			},
			result: scut.TestReturn{
				Score:         10,
				NumberOfInfo:  2,
				NumberOfDebug: 3,
			},
		},
		{
			name: "Has one release that is not signed but has provenance",
			findings: []finding.Finding{
				signedProbe(0, 0, finding.OutcomeFalse),
				provenanceProbe(0, 0, finding.OutcomeTrue),
				verifiedSignaturesProbe(finding.OutcomeNotApplicable),
				verifiedProvenanceProbe(finding.OutcomeNotApplicable),
			},
			result: scut.TestReturn{
				Score:         checker.MaxResultScore,
				NumberOfInfo:  1,
				NumberOfWarn:  0,
				NumberOfDebug: 4,
			},
		},
		{
			name: "3 releases. One release has one signed, and one release has provenance.",
			findings: []finding.Finding{
				// Release 1:
				signedProbe(release0, asset1, finding.OutcomeTrue),
				provenanceProbe(release0, asset0, finding.OutcomeFalse),
				// Release 2
				signedProbe(release1, asset0, finding.OutcomeFalse),
				provenanceProbe(release1, asset0, finding.OutcomeFalse),
				// Release 3
				signedProbe(release2, asset0, finding.OutcomeFalse),
				provenanceProbe(release2, asset1, finding.OutcomeTrue),
				verifiedSignaturesProbe(finding.OutcomeNotApplicable),
				verifiedProvenanceProbe(finding.OutcomeNotApplicable),
			},
			result: scut.TestReturn{
				Score:         6,
				NumberOfInfo:  2,
				NumberOfWarn:  3,
				NumberOfDebug: 6,
			},
		},
		{
			name: "5 releases. Two releases have one signed each, and two releases have one provenance each.",
			findings: []finding.Finding{
				// Release 1:
				signedProbe(release0, asset1, finding.OutcomeTrue),
				provenanceProbe(release0, asset1, finding.OutcomeFalse),
				// Release 2:
				signedProbe(release1, asset0, finding.OutcomeTrue),
				provenanceProbe(release1, asset0, finding.OutcomeFalse),
				// Release 3:
				signedProbe(release2, asset0, finding.OutcomeFalse),
				provenanceProbe(release2, asset0, finding.OutcomeTrue),
				// Release 4, Asset 1:
				signedProbe(release3, asset0, finding.OutcomeFalse),
				provenanceProbe(release3, asset0, finding.OutcomeTrue),
				// Release 5, Asset 1:
				signedProbe(release4, asset0, finding.OutcomeFalse),
				provenanceProbe(release4, asset0, finding.OutcomeFalse),
				verifiedSignaturesProbe(finding.OutcomeNotApplicable),
				verifiedProvenanceProbe(finding.OutcomeNotApplicable),
			},
			result: scut.TestReturn{
				Score:         7,
				NumberOfInfo:  4,
				NumberOfWarn:  4,
				NumberOfDebug: 9,
			},
		},
		{
			name: "5 releases. All have one signed artifact.",
			findings: []finding.Finding{
				// Release 1:
				signedProbe(release0, asset1, finding.OutcomeTrue),
				provenanceProbe(release0, asset1, finding.OutcomeFalse),
				// Release 2:
				signedProbe(release1, asset0, finding.OutcomeTrue),
				provenanceProbe(release1, asset0, finding.OutcomeFalse),
				// Release 3:
				signedProbe(release2, asset0, finding.OutcomeTrue),
				provenanceProbe(release2, asset0, finding.OutcomeFalse),
				// Release 4:
				signedProbe(release3, asset0, finding.OutcomeTrue),
				provenanceProbe(release3, asset0, finding.OutcomeFalse),
				// Release 5:
				signedProbe(release4, asset0, finding.OutcomeTrue),
				provenanceProbe(release4, asset0, finding.OutcomeFalse),
				verifiedSignaturesProbe(finding.OutcomeNotApplicable),
				verifiedProvenanceProbe(finding.OutcomeNotApplicable),
			},
			result: scut.TestReturn{
				Score:         8,
				NumberOfInfo:  5,
				NumberOfWarn:  5,
				NumberOfDebug: 7,
			},
		},
		{
			name: "too many releases is an error (6 when lookback is 5)",
			findings: []finding.Finding{
				// Release 1:
				// Release 1, Asset 1:
				signedProbe(release0, asset0, finding.OutcomeTrue),
				provenanceProbe(release0, asset0, finding.OutcomeTrue),
				// Release 2:
				// Release 2, Asset 1:
				signedProbe(release1, asset0, finding.OutcomeTrue),
				provenanceProbe(release1, asset0, finding.OutcomeTrue),
				// Release 3, Asset 1:
				signedProbe(release2, asset0, finding.OutcomeTrue),
				provenanceProbe(release2, asset0, finding.OutcomeTrue),
				// Release 4, Asset 1:
				signedProbe(release3, asset0, finding.OutcomeTrue),
				provenanceProbe(release3, asset0, finding.OutcomeTrue),
				// Release 5, Asset 1:
				signedProbe(release4, asset0, finding.OutcomeTrue),
				provenanceProbe(release4, asset0, finding.OutcomeTrue),
				// Release 6, Asset 1:
				signedProbe(release5, asset0, finding.OutcomeTrue),
				provenanceProbe(release5, asset0, finding.OutcomeTrue),
				verifiedSignaturesProbe(finding.OutcomeNotApplicable),
				verifiedProvenanceProbe(finding.OutcomeNotApplicable),
			},
			result: scut.TestReturn{
				Error:         sce.ErrScorecardInternal,
				Score:         checker.InconclusiveResultScore,
				NumberOfInfo:  12,
				NumberOfDebug: 8,
			},
		},
		{
			name: "Maven package with verified GPG signature - max score",
			findings: []finding.Finding{
				signedProbe(release0, asset0, finding.OutcomeNotApplicable),
				provenanceProbe(release0, asset0, finding.OutcomeNotApplicable),
				verifiedSignaturesProbeWithPackage(finding.OutcomeTrue, "maven", "com.google.code.gson:gson", "2.10.1"),
				verifiedProvenanceProbe(finding.OutcomeNotApplicable),
			},
			result: scut.TestReturn{
				Score:         checker.MaxResultScore,
				NumberOfInfo:  1,
				NumberOfDebug: 3, // 3 OutcomeNotApplicable probes
			},
		},
		{
			name: "PyPI package with verified Sigstore attestation - max score",
			findings: []finding.Finding{
				signedProbe(release0, asset0, finding.OutcomeNotApplicable),
				provenanceProbe(release0, asset0, finding.OutcomeNotApplicable),
				verifiedSignaturesProbeWithPackage(finding.OutcomeTrue, "pypi", "numpy", "1.26.0"),
				verifiedProvenanceProbe(finding.OutcomeNotApplicable),
			},
			result: scut.TestReturn{
				Score:         checker.MaxResultScore,
				NumberOfInfo:  1,
				NumberOfDebug: 3,
			},
		},
		{
			name: "Package with failed signature verification - score 3",
			findings: []finding.Finding{
				signedProbe(release0, asset0, finding.OutcomeNotApplicable),
				provenanceProbe(release0, asset0, finding.OutcomeNotApplicable),
				verifiedSignaturesProbeWithPackage(finding.OutcomeFalse, "maven", "com.example:lib", "1.0.0"),
				verifiedProvenanceProbe(finding.OutcomeNotApplicable),
			},
			result: scut.TestReturn{
				Score:         3,
				NumberOfWarn:  2, // Failed verification + warning message
				NumberOfDebug: 3,
			},
		},
		{
			name: "Three packages: all verified - max score",
			findings: []finding.Finding{
				signedProbe(release0, asset0, finding.OutcomeNotApplicable),
				provenanceProbe(release0, asset0, finding.OutcomeNotApplicable),
				verifiedSignaturesProbeWithPackage(finding.OutcomeTrue, "maven", "com.example:lib1", "1.0.0"),
				verifiedSignaturesProbeWithPackage(finding.OutcomeTrue, "maven", "com.example:lib2", "2.0.0"),
				verifiedSignaturesProbeWithPackage(finding.OutcomeTrue, "pypi", "requests", "2.31.0"),
				verifiedProvenanceProbe(finding.OutcomeNotApplicable),
			},
			result: scut.TestReturn{
				Score:         checker.MaxResultScore,
				NumberOfInfo:  3,
				NumberOfDebug: 3,
			},
		},
		{
			name: "Three packages: two verified, one failed - score capped at 7",
			findings: []finding.Finding{
				signedProbe(release0, asset0, finding.OutcomeNotApplicable),
				provenanceProbe(release0, asset0, finding.OutcomeNotApplicable),
				verifiedSignaturesProbeWithPackage(finding.OutcomeTrue, "maven", "com.example:lib1", "1.0.0"),
				verifiedSignaturesProbeWithPackage(finding.OutcomeTrue, "pypi", "numpy", "1.26.0"),
				verifiedSignaturesProbeWithPackage(finding.OutcomeFalse, "maven", "com.example:bad", "1.0.0"),
				verifiedProvenanceProbe(finding.OutcomeNotApplicable),
			},
			result: scut.TestReturn{
				Score:         7, // Capped at 7 when there are failures
				NumberOfInfo:  2,
				NumberOfWarn:  1,
				NumberOfDebug: 3,
			},
		},
		{
			name: "Five packages: three verified (60%) - score 6",
			findings: []finding.Finding{
				signedProbe(release0, asset0, finding.OutcomeNotApplicable),
				provenanceProbe(release0, asset0, finding.OutcomeNotApplicable),
				verifiedSignaturesProbeWithPackage(finding.OutcomeTrue, "maven", "com.example:lib1", "1.0.0"),
				verifiedSignaturesProbeWithPackage(finding.OutcomeTrue, "maven", "com.example:lib2", "2.0.0"),
				verifiedSignaturesProbeWithPackage(finding.OutcomeTrue, "pypi", "numpy", "1.26.0"),
				verifiedSignaturesProbeWithPackage(finding.OutcomeFalse, "maven", "com.example:lib3", "3.0.0"),
				verifiedSignaturesProbeWithPackage(finding.OutcomeFalse, "pypi", "requests", "2.31.0"),
				verifiedProvenanceProbe(finding.OutcomeNotApplicable),
			},
			result: scut.TestReturn{
				Score:         6, // 60% verified = score 6
				NumberOfInfo:  3,
				NumberOfWarn:  2,
				NumberOfDebug: 3,
			},
		},
		{
			name: "Package with verified provenance from deps.dev",
			findings: []finding.Finding{
				signedProbe(release0, asset0, finding.OutcomeNotApplicable),
				provenanceProbe(release0, asset0, finding.OutcomeNotApplicable),
				verifiedSignaturesProbe(finding.OutcomeNotApplicable),
				verifiedProvenanceProbeWithPackage(finding.OutcomeTrue, "maven", "com.example:lib", "1.0.0"),
			},
			result: scut.TestReturn{
				Score:         checker.MaxResultScore,
				NumberOfInfo:  1,
				NumberOfDebug: 3,
			},
		},
		{
			name: "Package with verified signature and provenance - max score",
			findings: []finding.Finding{
				signedProbe(release0, asset0, finding.OutcomeNotApplicable),
				provenanceProbe(release0, asset0, finding.OutcomeNotApplicable),
				verifiedSignaturesProbeWithPackage(finding.OutcomeTrue, "maven", "com.example:lib", "1.0.0"),
				verifiedProvenanceProbeWithPackage(finding.OutcomeTrue, "maven", "com.example:lib", "1.0.0"),
			},
			result: scut.TestReturn{
				Score:         checker.MaxResultScore,
				NumberOfInfo:  2,
				NumberOfDebug: 2,
			},
		},
		{
			name: "Mix: GitHub releases and Maven packages",
			findings: []finding.Finding{
				// GitHub releases (detection only)
				signedProbe(release0, asset0, finding.OutcomeTrue),
				provenanceProbe(release0, asset0, finding.OutcomeFalse),
				signedProbe(release1, asset0, finding.OutcomeTrue),
				provenanceProbe(release1, asset0, finding.OutcomeFalse),
				// Maven packages (verified)
				verifiedSignaturesProbeWithPackage(finding.OutcomeTrue, "maven", "com.example:lib1", "1.0.0"),
				verifiedSignaturesProbeWithPackage(finding.OutcomeTrue, "maven", "com.example:lib2", "2.0.0"),
				verifiedProvenanceProbe(finding.OutcomeNotApplicable),
			},
			result: scut.TestReturn{
				Score:         8, // 2 GitHub releases with signatures (8 each), no deductions for verified packages
				NumberOfInfo:  4, // 2 GitHub signed + 2 Maven verified
				NumberOfWarn:  2, // 2 GitHub missing provenance
				NumberOfDebug: 3,
			},
		},
		{
			name: "No packages or releases found",
			findings: []finding.Finding{
				signedProbe(release0, asset0, finding.OutcomeNotApplicable),
				provenanceProbe(release0, asset0, finding.OutcomeNotApplicable),
				verifiedSignaturesProbe(finding.OutcomeNotApplicable),
				verifiedProvenanceProbe(finding.OutcomeNotApplicable),
			},
			result: scut.TestReturn{
				Score:         checker.InconclusiveResultScore,
				NumberOfDebug: 4,
			},
		},
		{
			name: "Three packages: one verified, two have no signatures at all",
			findings: []finding.Finding{
				signedProbe(release0, asset0, finding.OutcomeNotApplicable),
				provenanceProbe(release0, asset0, finding.OutcomeNotApplicable),
				// Only one package has a signature that was verified
				verifiedSignaturesProbeWithPackage(finding.OutcomeTrue, "maven", "com.example:lib1", "1.0.0"),
				// The other two packages (lib2, lib3) have no signatures, so no findings are generated for them
				verifiedProvenanceProbe(finding.OutcomeNotApplicable),
			},
			result: scut.TestReturn{
				Score:         checker.MaxResultScore, // 1 out of 1 verified (100%)
				NumberOfInfo:  1,
				NumberOfDebug: 3,
			},
		},
		{
			name: "Five packages: two verified, one failed, two have no signatures",
			findings: []finding.Finding{
				signedProbe(release0, asset0, finding.OutcomeNotApplicable),
				provenanceProbe(release0, asset0, finding.OutcomeNotApplicable),
				// Two packages verified successfully
				verifiedSignaturesProbeWithPackage(finding.OutcomeTrue, "maven", "com.example:lib1", "1.0.0"),
				verifiedSignaturesProbeWithPackage(finding.OutcomeTrue, "pypi", "requests", "2.31.0"),
				// One package failed verification
				verifiedSignaturesProbeWithPackage(finding.OutcomeFalse, "maven", "com.example:bad", "1.0.0"),
				// Two other packages (lib4, lib5) have no signatures at all - no findings
				verifiedProvenanceProbe(finding.OutcomeNotApplicable),
			},
			result: scut.TestReturn{
				Score:         7, // 2 verified, 1 failed out of 3 total = capped at 7
				NumberOfInfo:  2,
				NumberOfWarn:  1,
				NumberOfDebug: 3,
			},
		},
		{
			name: "Three packages: ALL verification attempts failed",
			findings: []finding.Finding{
				signedProbe(release0, asset0, finding.OutcomeNotApplicable),
				provenanceProbe(release0, asset0, finding.OutcomeNotApplicable),
				// All three packages have signatures but verification failed for all
				verifiedSignaturesProbeWithPackage(finding.OutcomeFalse, "maven", "com.example:bad1", "1.0.0"),
				verifiedSignaturesProbeWithPackage(finding.OutcomeFalse, "maven", "com.example:bad2", "2.0.0"),
				verifiedSignaturesProbeWithPackage(finding.OutcomeFalse, "pypi", "bad-package", "3.0.0"),
				verifiedProvenanceProbe(finding.OutcomeNotApplicable),
			},
			result: scut.TestReturn{
				Score:         3, // All failed = score 3 with warning
				NumberOfWarn:  4, // 3 failed verifications + 1 warning message
				NumberOfDebug: 3,
			},
		},
		{
			name: "Single package: signature present but verification failed",
			findings: []finding.Finding{
				signedProbe(release0, asset0, finding.OutcomeNotApplicable),
				provenanceProbe(release0, asset0, finding.OutcomeNotApplicable),
				verifiedSignaturesProbeWithPackage(finding.OutcomeFalse, "maven", "org.example:compromised", "1.0.0"),
				verifiedProvenanceProbe(finding.OutcomeNotApplicable),
			},
			result: scut.TestReturn{
				Score:         3, // Single package failed = score 3 with warning
				NumberOfWarn:  2, // 1 failed verification + 1 warning message
				NumberOfDebug: 3,
			},
		},
		{
			name: "Ten packages: all verified successfully",
			findings: []finding.Finding{
				signedProbe(release0, asset0, finding.OutcomeNotApplicable),
				provenanceProbe(release0, asset0, finding.OutcomeNotApplicable),
				verifiedSignaturesProbeWithPackage(finding.OutcomeTrue, "maven", "com.example:lib1", "1.0.0"),
				verifiedSignaturesProbeWithPackage(finding.OutcomeTrue, "maven", "com.example:lib2", "2.0.0"),
				verifiedSignaturesProbeWithPackage(finding.OutcomeTrue, "maven", "com.example:lib3", "3.0.0"),
				verifiedSignaturesProbeWithPackage(finding.OutcomeTrue, "pypi", "pkg1", "1.0.0"),
				verifiedSignaturesProbeWithPackage(finding.OutcomeTrue, "pypi", "pkg2", "2.0.0"),
				verifiedSignaturesProbeWithPackage(finding.OutcomeTrue, "pypi", "pkg3", "3.0.0"),
				verifiedSignaturesProbeWithPackage(finding.OutcomeTrue, "pypi", "pkg4", "4.0.0"),
				verifiedSignaturesProbeWithPackage(finding.OutcomeTrue, "maven", "com.example:lib4", "4.0.0"),
				verifiedSignaturesProbeWithPackage(finding.OutcomeTrue, "maven", "com.example:lib5", "5.0.0"),
				verifiedSignaturesProbeWithPackage(finding.OutcomeTrue, "pypi", "pkg5", "5.0.0"),
				verifiedProvenanceProbe(finding.OutcomeNotApplicable),
			},
			result: scut.TestReturn{
				Score:         checker.MaxResultScore, // All 10 verified = max score
				NumberOfInfo:  10,
				NumberOfDebug: 3,
			},
		},
		{
			name: "GitHub releases mixed with packages: some have no signatures",
			findings: []finding.Finding{
				// GitHub release with signature detected
				signedProbe(release0, asset0, finding.OutcomeTrue),
				provenanceProbe(release0, asset0, finding.OutcomeFalse),
				// One Maven package verified
				verifiedSignaturesProbeWithPackage(finding.OutcomeTrue, "maven", "com.example:lib1", "1.0.0"),
				// Other Maven packages (lib2, lib3) returned by deps.dev but have no signatures - no findings
				verifiedProvenanceProbe(finding.OutcomeNotApplicable),
			},
			result: scut.TestReturn{
				Score:         8, // 1 GitHub release with signature (8 points), package verification succeeded (no deductions)
				NumberOfInfo:  2,
				NumberOfWarn:  1, // GitHub release missing provenance
				NumberOfDebug: 2,
			},
		},
		{
			name: "Partial verification with no failures: only some packages found and verified",
			findings: []finding.Finding{
				signedProbe(release0, asset0, finding.OutcomeNotApplicable),
				provenanceProbe(release0, asset0, finding.OutcomeNotApplicable),
				// Only 2 out of potentially more packages were found and verified
				// The others simply don't exist in package managers (not failed, just not present)
				verifiedSignaturesProbeWithPackage(finding.OutcomeTrue, "maven", "com.example:lib1", "1.0.0"),
				verifiedSignaturesProbeWithPackage(finding.OutcomeTrue, "pypi", "pkg1", "1.0.0"),
				verifiedProvenanceProbe(finding.OutcomeNotApplicable),
			},
			result: scut.TestReturn{
				Score:         checker.MaxResultScore, // All packages that exist are verified (2/2 = 100%)
				NumberOfInfo:  2,
				NumberOfDebug: 3,
			},
		},
		{
			name: "Proportional scoring with only verified packages (no failures): 2 of 4 packages verified",
			findings: []finding.Finding{
				signedProbe(release0, asset0, finding.OutcomeNotApplicable),
				provenanceProbe(release0, asset0, finding.OutcomeNotApplicable),
				// 2 packages verified successfully
				verifiedSignaturesProbeWithPackage(finding.OutcomeTrue, "maven", "com.example:lib1", "1.0.0"),
				verifiedProvenanceProbeWithPackage(finding.OutcomeTrue, "npm", "lodash", "4.17.21"),
				// 2 other packages exist but have neither verified signatures nor verified provenance
				// They're in uniqueItems but not in verifiedItems (simulation of future probe behavior)
				// NOTE: Currently impossible with existing probes - packages are either verified (True),
				// failed (False), or not attempted (NotApplicable, not added to maps).
				// This test documents intended behavior if probes add "attempted but inconclusive" results.
				verifiedSignaturesProbeWithPackage(finding.OutcomeNotApplicable, "maven", "com.example:lib2", "2.0.0"),
				verifiedSignaturesProbeWithPackage(finding.OutcomeNotApplicable, "pypi", "requests", "2.31.0"),
			},
			result: scut.TestReturn{
				Score:         checker.MaxResultScore, // Currently: 2/2 = 100% (inconclusive packages not counted)
				NumberOfInfo:  2,
				NumberOfDebug: 4,
			},
		},
		{
			name: "GitHub releases with detected signatures but no package verification",
			findings: []finding.Finding{
				// GitHub releases have signature files detected
				signedProbe(release0, asset0, finding.OutcomeTrue),
				signedProbe(release1, asset1, finding.OutcomeTrue),
				provenanceProbe(release0, asset0, finding.OutcomeFalse),
				provenanceProbe(release1, asset1, finding.OutcomeFalse),
				// No package verification attempted (no packages found)
				verifiedSignaturesProbe(finding.OutcomeNotApplicable),
				verifiedProvenanceProbe(finding.OutcomeNotApplicable),
			},
			result: scut.TestReturn{
				Score:         8, // Signatures detected but not cryptographically verified
				NumberOfInfo:  2,
				NumberOfWarn:  2,
				NumberOfDebug: 4,
			},
		},
		{
			name: "No signatures found anywhere",
			findings: []finding.Finding{
				// No signatures in GitHub releases
				signedProbe(release0, asset0, finding.OutcomeFalse),
				provenanceProbe(release0, asset0, finding.OutcomeFalse),
				// No verification attempted (no packages or signatures found)
				verifiedSignaturesProbe(finding.OutcomeNotApplicable),
				verifiedProvenanceProbe(finding.OutcomeNotApplicable),
			},
			result: scut.TestReturn{
				Score:         checker.MinResultScore, // No signatures found
				NumberOfWarn:  2,
				NumberOfDebug: 3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			got := SignedReleases(tt.name, tt.findings, &dl)
			scut.ValidateTestReturn(t, tt.name, &tt.result, &got, &dl)
		})
	}
}

func TestSignedReleases_ErrorHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		errorContains string
		findings      []finding.Finding
		expectError   bool
	}{
		{
			name: "Missing required probe: releasesAreSigned",
			findings: []finding.Finding{
				// Missing releasesAreSigned probe
				provenanceProbe(release0, asset0, finding.OutcomeNotApplicable),
				verifiedSignaturesProbe(finding.OutcomeNotApplicable),
				verifiedProvenanceProbe(finding.OutcomeNotApplicable),
			},
			expectError:   true,
			errorContains: "invalid probe results",
		},
		{
			name: "Missing required probe: releasesHaveProvenance",
			findings: []finding.Finding{
				signedProbe(release0, asset0, finding.OutcomeNotApplicable),
				// Missing releasesHaveProvenance probe
				verifiedSignaturesProbe(finding.OutcomeNotApplicable),
				verifiedProvenanceProbe(finding.OutcomeNotApplicable),
			},
			expectError:   true,
			errorContains: "invalid probe results",
		},
		{
			name: "Missing required probe: releasesHaveVerifiedSignatures",
			findings: []finding.Finding{
				signedProbe(release0, asset0, finding.OutcomeNotApplicable),
				provenanceProbe(release0, asset0, finding.OutcomeNotApplicable),
				// Missing releasesHaveVerifiedSignatures probe
				verifiedProvenanceProbe(finding.OutcomeNotApplicable),
			},
			expectError:   true,
			errorContains: "invalid probe results",
		},
		{
			name: "Missing required probe: releasesHaveVerifiedProvenance",
			findings: []finding.Finding{
				signedProbe(release0, asset0, finding.OutcomeNotApplicable),
				provenanceProbe(release0, asset0, finding.OutcomeNotApplicable),
				verifiedSignaturesProbe(finding.OutcomeNotApplicable),
				// Missing releasesHaveVerifiedProvenance probe
			},
			expectError:   true,
			errorContains: "invalid probe results",
		},
		{
			name: "Extra unexpected probe",
			findings: []finding.Finding{
				signedProbe(release0, asset0, finding.OutcomeNotApplicable),
				provenanceProbe(release0, asset0, finding.OutcomeNotApplicable),
				verifiedSignaturesProbe(finding.OutcomeNotApplicable),
				verifiedProvenanceProbe(finding.OutcomeNotApplicable),
				// Extra probe that shouldn't be here
				{
					Probe:   "someOtherProbe",
					Outcome: finding.OutcomeNotApplicable,
				},
			},
			expectError:   true,
			errorContains: "invalid probe results",
		},
		{
			name:          "Empty findings list",
			findings:      []finding.Finding{},
			expectError:   true,
			errorContains: "invalid probe results",
		},
		{
			name: "Duplicate probe",
			findings: []finding.Finding{
				signedProbe(release0, asset0, finding.OutcomeNotApplicable),
				signedProbe(release1, asset1, finding.OutcomeNotApplicable), // Duplicate probe (different values but same probe name)
				provenanceProbe(release0, asset0, finding.OutcomeNotApplicable),
				verifiedSignaturesProbe(finding.OutcomeNotApplicable),
				verifiedProvenanceProbe(finding.OutcomeNotApplicable),
			},
			expectError:   false, // Multiple findings from same probe are valid (different releases)
			errorContains: "",
		},
		{
			name: "Finding with missing release information",
			findings: []finding.Finding{
				signedProbe(release0, asset0, finding.OutcomeNotApplicable),
				provenanceProbe(release0, asset0, finding.OutcomeNotApplicable),
				verifiedSignaturesProbe(finding.OutcomeNotApplicable),
				verifiedProvenanceProbe(finding.OutcomeNotApplicable),
				// Finding without Values map (causes error)
				{
					Probe:   releasesAreSigned.Probe,
					Outcome: finding.OutcomeTrue,
					Values:  nil, // Missing Values
				},
			},
			expectError:   true,
			errorContains: "no release found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			got := SignedReleases("Signed-Releases", tt.findings, &dl)

			if tt.expectError {
				if got.Error == nil {
					t.Errorf("Expected error containing %q, but got no error", tt.errorContains)
				} else if !strings.Contains(got.Error.Error(), tt.errorContains) {
					t.Errorf("Expected error containing %q, but got: %v", tt.errorContains, got.Error)
				}
			} else {
				if got.Error != nil {
					t.Errorf("Expected no error, but got: %v", got.Error)
				}
			}
		})
	}
}

func Test_getReleaseName(t *testing.T) {
	t.Parallel()
	type args struct {
		f *finding.Finding
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no release",
			args: args{
				f: &finding.Finding{
					Values: map[string]string{},
				},
			},
			want: "",
		},
		{
			name: "release",
			args: args{
				f: &finding.Finding{
					Values: map[string]string{
						releasesAreSigned.ReleaseNameKey: "v1",
					},
					Probe: releasesAreSigned.Probe,
				},
			},
			want: "v1",
		},
		{
			name: "release and asset",
			args: args{
				f: &finding.Finding{
					Values: map[string]string{
						releasesAreSigned.ReleaseNameKey: "v1",
						releasesAreSigned.AssetNameKey:   "artifact-1",
					},
					Probe: releasesAreSigned.Probe,
				},
			},
			want: "v1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := getReleaseName(tt.args.f); got != tt.want {
				t.Errorf("getReleaseName() = %v, want %v", got, tt.want)
			}
		})
	}
}
