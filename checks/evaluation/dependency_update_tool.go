// Copyright 2020 OpenSSF Scorecard Authors
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
	"github.com/ossf/scorecard/v4/probes/toolDependabotInstalled"
	"github.com/ossf/scorecard/v4/probes/toolPyUpInstalled"
	"github.com/ossf/scorecard/v4/probes/toolRenovateInstalled"
)

// DependencyUpdateTool applies the score policy and logs the details
// for the Dependency-Update-Tool check.
func DependencyUpdateTool(name string,
	findings []finding.Finding, dl checker.DetailLogger,
) checker.CheckResult {
	expectedProbes := []string{
		toolDependabotInstalled.Probe,
		toolPyUpInstalled.Probe,
		toolRenovateInstalled.Probe,
	}
	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	for i := range findings {
		f := &findings[i]
		if f.Outcome == finding.OutcomePositive {
			// Log all findings except the negative ones.
			checker.LogFindings(nonNegativeFindings(findings), dl)
			return checker.CreateMaxScoreResult(name, "update tool detected")
		}
	}

	// Log all findings.
	checker.LogFindings(findings, dl)
	return checker.CreateMinScoreResult(name, "no update tool detected")
}
