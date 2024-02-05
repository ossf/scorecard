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
package hasPermissiveLicense

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/internal/utils/uerror"
)

//go:embed *.yml
var fs embed.FS

const Probe = "hasPermissiveLicense"

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
		spdxID := licenseFile.LicenseInformation.SpdxID
		switch spdxID {
		case
			"Unlicense",
			"Beerware",
			"Apache-2.0",
			"MIT",
			"0BSD",
			"BSD-1-Clause",
			"BSD-2-Clause",
			"BSD-3-Clause",
			"BSD-4-Clause",
			"APSL-1.0",
			"APSL-1.1",
			"APSL-1.2",
			"APSL-2.0",
			"ECL-1.0",
			"ECL-2.0",
			"EFL-1.0",
			"EFL-2.0",
			"Fair",
			"FSFAP",
			"WTFPL",
			"Zlib",
			"CNRI-Python",
			"ISC",
			"Intel":
			// Store the license name in the msg
			msg := licenseFile.LicenseInformation.Name

			f, err := finding.NewWith(fs, Probe,
				msg, nil,
				finding.OutcomePositive)
			if err != nil {
				return nil, Probe, fmt.Errorf("create finding: %w", err)
			}
			return []finding.Finding{*f}, Probe, nil
		}
	}

	f, err := finding.NewWith(fs, Probe,
		"", nil,
		finding.OutcomeNegative)
	if err != nil {
		return nil, Probe, fmt.Errorf("create finding: %w", err)
	}
	return []finding.Finding{*f}, Probe, nil
}
