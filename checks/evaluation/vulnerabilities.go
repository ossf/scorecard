// Copyright 2022 OpenSSF Scorecard Authors
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

package evaluation

import (
	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/hasKnownVulnerabilities"
)

// Vulnerabilities applies the score policy for the Vulnerabilities check.
func Vulnerabilities(name string,
	findings []finding.Finding,
	dl checker.DetailLogger,
) checker.CheckResult {
	expectedProbes := []string{
		hasKnownVulnerabilities.Probe,
	}

	err := validateFindings(findings, expectedProbes)
	if err != nil {
		return checker.CreateRuntimeErrorResult(name, err)
	}

	vulnsFound := 0
	for _, f := range findings {
		if f.Outcome == finding.OutcomeNegative {
			// Log all the negative findings.
			checker.LogFindings(negativeFindings(findings), dl)
			vulnsFound++
		}
	}

	score := checker.MaxResultScore - vulnsFound

	if score < checker.MinResultScore {
		score = checker.MinResultScore
	}

	return checker.CreateResultWithScore(name, "vulnerabilities detected", score)
}

func validateFindings(findings []finding.Finding, expectedProbes []string) error {
	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		return sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
	}

	if len(findings) == 0 {
		return sce.WithMessage(sce.ErrScorecardInternal, "found 0 findings. Should not happen")
	}
	return nil
}

func negativeFindings(findings []finding.Finding) []finding.Finding {
	var ff []finding.Finding
	for i := range findings {
		f := &findings[i]
		if f.Outcome == finding.OutcomePositive {
			continue
		}
		ff = append(ff, *f)
	}
	return ff
}
