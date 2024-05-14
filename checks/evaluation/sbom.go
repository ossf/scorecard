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

package evaluation

import (
	"github.com/ossf/scorecard/v5/checker"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/hasReleaseSBOM"
	"github.com/ossf/scorecard/v5/probes/hasSBOM"
)

// SBOM applies the score policy for the SBOM check.
func SBOM(name string,
	findings []finding.Finding,
	dl checker.DetailLogger,
) checker.CheckResult {
	// We have 4 unique probes, each should have a finding.
	expectedProbes := []string{
		hasSBOM.Probe,
		hasReleaseSBOM.Probe,
	}

	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	// Compute the score.
	score := 0
	var detailsMsg string
	m := make(map[string]bool)
	var logLevel checker.DetailType
	for i := range findings {
		f := &findings[i]
		switch f.Outcome {
		case finding.OutcomeTrue:
			logLevel = checker.DetailInfo
			switch f.Probe {
			case hasSBOM.Probe:
				detailsMsg = "SBOM file found in project"
				score += scoreProbeOnce(f.Probe, m, 5)
			case hasReleaseSBOM.Probe:
				detailsMsg = "SBOM file found in release artifacts"
				score += scoreProbeOnce(f.Probe, m, 5)
			}
		case finding.OutcomeFalse:
			logLevel = checker.DetailWarn
			switch f.Probe {
			case hasSBOM.Probe:
				detailsMsg = "SBOM file not found in project"
			case hasReleaseSBOM.Probe:
				detailsMsg = "SBOM file not found in release artifacts"
			}
		default:
			continue // for linting
		}
		checker.LogFinding(dl, f, logLevel)
	}

	_, defined := m[hasSBOM.Probe]
	if !defined {
		return checker.CreateMinScoreResult(name, "SBOM file not detected")
	}

	return checker.CreateResultWithScore(name, detailsMsg, score)
}
