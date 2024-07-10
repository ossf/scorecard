// Copyright 2024 OpenSSF Scorecard Authors
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
package releasesHaveVerifiedProvenance

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/checknames"
	"github.com/ossf/scorecard/v5/internal/probes"
)

func init() {
	probes.MustRegister(Probe, Run, []checknames.CheckName{checknames.SignedReleases})
}

//go:embed *.yml
var fs embed.FS

const (
	Probe = "releasesHaveVerifiedProvenance"
)

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	var findings []finding.Finding

	if len(raw.SignedReleasesResults.Packages) == 0 {
		f, err := finding.NewNotApplicable(fs, Probe, "no package manager releases found", nil)
		if err != nil {
			return []finding.Finding{}, Probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
		return findings, Probe, nil
	}

	for i := range raw.SignedReleasesResults.Packages {
		p := raw.SignedReleasesResults.Packages[i]

		if !p.Provenance.IsVerified {
			f, err := finding.NewFalse(fs, Probe, "release without verified provenance", nil)
			if err != nil {
				return []finding.Finding{}, Probe, fmt.Errorf("create finding: %w", err)
			}
			findings = append(findings, *f)
			continue
		}

		f, err := finding.NewTrue(fs, Probe, "release with verified provenance", nil)
		if err != nil {
			return []finding.Finding{}, Probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	}

	return findings, Probe, nil
}
