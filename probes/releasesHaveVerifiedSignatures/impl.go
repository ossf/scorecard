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

//nolint:stylecheck
package releasesHaveVerifiedSignatures

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/checknames"
	"github.com/ossf/scorecard/v5/internal/probes"
	"github.com/ossf/scorecard/v5/probes/internal/utils/uerror"
)

func init() {
	probes.MustRegister(Probe, Run, []checknames.CheckName{checknames.SignedReleases})
}

//go:embed *.yml
var fs embed.FS

const (
	Probe = "releasesHaveVerifiedSignatures"
)

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	var findings []finding.Finding

	// Check if there are any packages to verify
	if len(raw.SignedReleasesResults.Packages) == 0 {
		f, err := finding.NewWith(fs, Probe,
			"no packages found to verify signatures",
			nil,
			finding.OutcomeNotApplicable)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
		return findings, Probe, nil
	}

	// Track if we found any signatures at all
	foundAnySignatures := false

	// Check each package for verified signatures
	for _, pkg := range raw.SignedReleasesResults.Packages {
		if len(pkg.Signatures) == 0 {
			continue // No signatures to verify for this package
		}

		foundAnySignatures = true

		for _, sig := range pkg.Signatures {
			if !sig.IsVerified {
				// Signature present but verification failed
				message := fmt.Sprintf("package %s:%s signature verification failed", pkg.System, pkg.Name)
				if sig.ErrorMsg != "" {
					message += ": " + sig.ErrorMsg
				}
				f, err := finding.NewWith(fs, Probe,
					message,
					nil,
					finding.OutcomeFalse)
				if err != nil {
					return nil, Probe, fmt.Errorf("create finding: %w", err)
				}
				f = f.WithValue("packageSystem", pkg.System)
				f = f.WithValue("packageName", pkg.Name)
				f = f.WithValue("packageVersion", pkg.Version)
				f = f.WithValue("signatureType", string(sig.Type))
				f = f.WithValue("errorMsg", sig.ErrorMsg)
				findings = append(findings, *f)
				continue
			}

			// Signature successfully verified
			message := fmt.Sprintf(
				"package %s:%s version %s has verified %s signature for %s",
				pkg.System, pkg.Name, pkg.Version, sig.Type, sig.ArtifactURL,
			)
			f, err := finding.NewWith(fs, Probe,
				message,
				nil,
				finding.OutcomeTrue)
			if err != nil {
				return nil, Probe, fmt.Errorf("create finding: %w", err)
			}
			f = f.WithValue("packageSystem", pkg.System)
			f = f.WithValue("packageName", pkg.Name)
			f = f.WithValue("packageVersion", pkg.Version)
			f = f.WithValue("signatureType", string(sig.Type))
			f = f.WithValue("artifactURL", sig.ArtifactURL)
			if sig.KeyID != "" {
				f = f.WithValue("keyID", sig.KeyID)
			}
			findings = append(findings, *f)
		}
	} // If we didn't find any signatures to verify
	if !foundAnySignatures {
		f, err := finding.NewWith(fs, Probe,
			"no signatures found to verify",
			nil,
			finding.OutcomeNotApplicable)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	}

	return findings, Probe, nil
}
