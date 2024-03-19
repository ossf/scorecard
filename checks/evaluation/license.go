// Copyright 2021 OpenSSF Scorecard Authors
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
	"github.com/ossf/scorecard/v4/probes/hasFSFOrOSIApprovedLicense"
	"github.com/ossf/scorecard/v4/probes/hasLicenseFile"
)

// License applies the score policy for the License check.
func License(name string,
	findings []finding.Finding,
	dl checker.DetailLogger,
) checker.CheckResult {
	// We have 3 unique probes, each should have a finding.
	expectedProbes := []string{
		hasLicenseFile.Probe,
		hasFSFOrOSIApprovedLicense.Probe,
	}

	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	// Compute the score.
	score := 0
	m := make(map[string]bool)
	for i := range findings {
		f := &findings[i]
		if f.Outcome == finding.OutcomePositive {
			switch f.Probe {
			case hasFSFOrOSIApprovedLicense.Probe:
				score += scoreProbeOnce(f.Probe, m, 1)
			case hasLicenseFile.Probe:
				score += scoreProbeOnce(f.Probe, m, 9)
			default:
				e := sce.WithMessage(sce.ErrScorecardInternal, "unknown probe results")
				return checker.CreateRuntimeErrorResult(name, e)
			}
		}
	}
	checker.LogFindings(findings, dl)

	_, defined := m[hasLicenseFile.Probe]
	if !defined {
		if score > 0 {
			e := sce.WithMessage(sce.ErrScorecardInternal, "score calculation problem")
			return checker.CreateRuntimeErrorResult(name, e)
		}
		return checker.CreateMinScoreResult(name, "license file not detected")
	}
	return checker.CreateResultWithScore(name, "license file detected", score)
}
