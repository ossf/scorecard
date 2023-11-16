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

//nolint:stylecheck
package contributorsFromOrgOrCompany

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/internal/utils/uerror"
)

const (
	minContributionsPerUser = 5
)

//go:embed *.yml
var fs embed.FS

const Probe = "contributorsFromOrgOrCompany"

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}
	var findings []finding.Finding

	users := raw.ContributorsResults.Users

	if len(users) == 0 {
		f, err := finding.NewWith(fs, Probe,
			"Project does not have contributors.", nil,
			finding.OutcomeNegative)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}

		findings = append(findings, *f)
		return findings, Probe, nil
	}

	entities := make(map[string]bool)

	for _, user := range users {
		if user.NumContributions < minContributionsPerUser {
			continue
		}

		for _, org := range user.Organizations {
			entities[org.Login] = true
		}

		for _, comp := range user.Companies {
			entities[comp] = true
		}
	}

	if len(entities) == 0 {
		f, err := finding.NewWith(fs, Probe,
			"No companies/organizations have contributed to the project.", nil,
			finding.OutcomeNegative)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}

		findings = append(findings, *f)
		return findings, Probe, nil
	}

	// Convert entities map to findings slice
	for e := range entities {
		f, err := finding.NewWith(fs, Probe,
			fmt.Sprintf("%s contributor org/company found", e), nil,
			finding.OutcomePositive)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}

		findings = append(findings, *f)
	}

	return findings, Probe, nil
}
