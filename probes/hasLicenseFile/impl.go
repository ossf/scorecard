// Copyright 2023 OpenSSF Scorecard Authors
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

// nolint:stylecheck
package hasLicenseFile

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/internal/utils/uerror"
)

//go:embed *.yml
var fs embed.FS

const Probe = "hasLicenseFile"

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	var findings []finding.Finding
	var outcome finding.Outcome
	var msg string

	licenseFiles := raw.LicenseResults.LicenseFiles

	if len(licenseFiles) == 0 {
		outcome = finding.OutcomeNegative
		msg = "project does not have a license file"
		f, err := finding.NewWith(fs, Probe,
			msg, nil,
			outcome)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
		return findings, Probe, nil
	} else {
		for _, licenseFile := range licenseFiles {
			licenseFile := licenseFile
			loc := &finding.Location{
				Type:      licenseFile.File.Type,
				Path:      licenseFile.File.Path,
				LineStart: &licenseFile.File.Offset,
				LineEnd:   &licenseFile.File.EndOffset,
				Snippet:   &licenseFile.File.Snippet,
			}
			msg = "project has a license file"
			outcome = finding.OutcomePositive
			f, err := finding.NewWith(fs, Probe,
				msg, loc,
				outcome)
			if err != nil {
				return nil, Probe, fmt.Errorf("create finding: %w", err)
			}
			findings = append(findings, *f)
		}
	}
	return findings, Probe, nil
}
