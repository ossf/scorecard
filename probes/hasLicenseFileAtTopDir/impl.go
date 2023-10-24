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
package hasLicenseFileAtTopDir

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/internal/utils/uerror"
)

//go:embed *.yml
var fs embed.FS

const Probe = "hasLicenseFileAtTopDir"

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	if raw.LicenseResults.LicenseFiles == nil || len(raw.LicenseResults.LicenseFiles) == 0 {
		f, err := finding.NewWith(fs, Probe,
			"project does not have a license file", nil,
			finding.OutcomeNotApplicable)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		return []finding.Finding{*f}, Probe, nil
	}

	for _, licenseFile := range raw.LicenseResults.LicenseFiles {
		switch licenseFile.LicenseInformation.Attribution {
		case checker.LicenseAttributionTypeAPI, checker.LicenseAttributionTypeHeuristics:
			// both repoAPI and scorecard (not using the API) follow checks.md
			// for a file to be found it must have been in the correct location
			// award location points.

			// Store the file path in the msg
			msg := licenseFile.File.Path
			f, err := finding.NewWith(fs, Probe,
				msg, nil,
				finding.OutcomePositive)
			if err != nil {
				return nil, Probe, fmt.Errorf("create finding: %w", err)
			}
			return []finding.Finding{*f}, Probe, nil

		case checker.LicenseAttributionTypeOther:
			msg := "License file found in unexpected location"
			f, err := finding.NewWith(fs, Probe,
				msg, nil,
				finding.OutcomeNegative)
			if err != nil {
				return nil, Probe, fmt.Errorf("create finding: %w", err)
			}
			return []finding.Finding{*f}, Probe, nil
		}
	}

	f, err := finding.NewWith(fs, Probe,
		"Did not find the license file at the expected location.", nil,
		finding.OutcomeNegative)
	if err != nil {
		return nil, Probe, fmt.Errorf("create finding: %w", err)
	}
	return []finding.Finding{*f}, Probe, nil
}
