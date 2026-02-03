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

package hasInactiveMaintainers

import (
	"embed"
	"fmt"
	"sort"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/checknames"
	"github.com/ossf/scorecard/v5/internal/probes"
	"github.com/ossf/scorecard/v5/probes/internal/utils/uerror"
)

func init() {
	probes.MustRegister(
		Probe,
		Run,
		[]checknames.CheckName{checknames.Maintained},
	)
}

//go:embed *.yml
var fs embed.FS

const Probe = "hasInactiveMaintainers"

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	maintainerActivity := raw.MaintainedResults.MaintainerActivity
	var findings []finding.Finding

	// If no maintainers found, return a finding indicating this
	if len(maintainerActivity) == 0 {
		f, err := finding.NewWith(fs, Probe,
			"No maintainers with elevated permissions found in the repository.", nil,
			finding.OutcomeNotApplicable)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
		return findings, Probe, nil
	}

	// Create one finding per maintainer
	// Sort usernames to ensure deterministic order for testing
	var usernames []string
	for username := range maintainerActivity {
		usernames = append(usernames, username)
	}
	sort.Strings(usernames)

	for _, username := range usernames {
		isActive := maintainerActivity[username]
		var outcome finding.Outcome
		var msg string

		if isActive {
			outcome = finding.OutcomeFalse // Active maintainer (no issue)
			msg = fmt.Sprintf(
				"Maintainer %s has been active in the last 6 months",
				username,
			)
		} else {
			outcome = finding.OutcomeTrue // Inactive maintainer (potential issue)
			msg = fmt.Sprintf(
				"Maintainer %s has been inactive for the "+
					"last 6 months",
				username,
			)
		}

		f, err := finding.NewWith(fs, Probe, msg, nil, outcome)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	}

	return findings, Probe, nil
}
