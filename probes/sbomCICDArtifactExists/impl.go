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
package sbomCICDArtifactExists

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/internal/utils/uerror"
)

//go:embed *.yml
var fs embed.FS

const Probe = "sbomCICDArtifactExists"

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	var findings []finding.Finding
	var outcome finding.Outcome
	var msg string

	sbomFiles := raw.SbomResults.SbomFiles

	if len(sbomFiles) == 0 {
		outcome = finding.OutcomeNegative
		msg = "Project does not produce an sbom file artifact as part of CICD"
		f, err := finding.NewWith(fs, Probe,
			msg, nil,
			outcome)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
		return findings, Probe, nil
	}

	for i := range sbomFiles {
		sbomFile := sbomFiles[i]

		if sbomFile.SbomInformation.Origin != checker.SbomOriginationTypeCICD {
			continue
		}

		loc := &finding.Location{
			Type:      sbomFile.File.Type,
			Path:      sbomFile.File.Path,
			LineStart: &sbomFile.File.Offset,
			LineEnd:   &sbomFile.File.EndOffset,
			Snippet:   &sbomFile.File.Snippet,
		}
		msg = "Project produces an sbom file artifact as part of CICD"
		outcome = finding.OutcomePositive
		f, err := finding.NewWith(fs, Probe,
			msg, loc,
			outcome)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	}

	if len(findings) == 0 {
		outcome = finding.OutcomeNegative
		msg = "Project does not produce an sbom file artifact as part of CICD"
		f, err := finding.NewWith(fs, Probe,
			msg, nil,
			outcome)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	}

	return findings, Probe, nil
}
