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
	"github.com/ossf/scorecard/v4/probes/packagedWithAutomatedWorkflow"
)

// Packaging applies the score policy for the Packaging check.
func Packaging(name string,
	findings []finding.Finding,
	dl checker.DetailLogger,
) checker.CheckResult {
	expectedProbes := []string{
		packagedWithAutomatedWorkflow.Probe,
	}

	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	// Currently there is only a single packaging probe that returns
	// a single positive or negative outcome. As such, in this evaluation,
	// we return max score if the outcome is positive and lowest score if
	// the outcome is negative.
	maxScore := false
	for _, f := range findings {
		f := f
		if f.Outcome == finding.OutcomePositive {
			maxScore = true
			// Log all findings except the negative ones.
			dl.Info(&checker.LogMessage{
				Finding: &f,
			})
		}
	}
	if maxScore {
		return checker.CreateMaxScoreResult(name, "packaging workflow detected")
	}

	checker.LogFindings(negativeFindings(findings), dl)
	return checker.CreateInconclusiveResult(name,
		"packaging workflow not detected")
}
