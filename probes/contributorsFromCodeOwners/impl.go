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
package contributorsFromCodeOwners

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
	probes.MustRegister(Probe, Run, []checknames.CheckName{checknames.Contributors})
}

//go:embed *.yml
var fs embed.FS

const Probe = "contributorsFromCodeOwners"

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}
	var findings []finding.Finding

	contributors := raw.ContributorsResults.Contributors
	owners := raw.ContributorsResults.CodeOwners

	if len(contributors) == 0 {
		f, err := finding.NewWith(fs, Probe,
			"Project does not have contributors.", nil,
			finding.OutcomeFalse)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}

		findings = append(findings, *f)
		return findings, Probe, nil
	}

	if len(owners) == 0 {
		f, err := finding.NewWith(fs, Probe,
			"Project does not have code owners.", nil,
			finding.OutcomeFalse)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}

		findings = append(findings, *f)
		return findings, Probe, nil
	}

	ownerContributors := make(map[string]bool)

	for _, owner := range owners {
		for _, contributor := range contributors {
			if owner.Login == contributor.Login {
				ownerContributors[owner.Login] = true
				break
			}
		}
	}

	if len(ownerContributors) == 0 {
		f, err := finding.NewWith(fs, Probe,
			"Project does not have any code owners who are contributors.", nil,
			finding.OutcomeFalse)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}

		findings = append(findings, *f)
		return findings, Probe, nil
	}

	for e := range ownerContributors {
		f, err := finding.NewWith(fs, Probe,
			fmt.Sprintf("%s code owner contributor found", e), nil,
			finding.OutcomeTrue)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}

		findings = append(findings, *f)
	}

	return findings, Probe, nil
}
