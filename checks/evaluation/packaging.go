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
	"github.com/ossf/scorecard/v5/checker"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/packagedWithAutomatedWorkflow"
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
	// a single true or false outcome. As such, in this evaluation,
	// we return max score if the outcome is true and lowest score if
	// the outcome is false.
	maxScore := false
	for i := range findings {
		f := &findings[i]
		var logLevel checker.DetailType
		switch f.Outcome {
		case finding.OutcomeFalse:
			logLevel = checker.DetailWarn
		case finding.OutcomeTrue:
			maxScore = true
			logLevel = checker.DetailInfo
		default:
			logLevel = checker.DetailDebug
		}
		checker.LogFinding(dl, f, logLevel)
	}
	if maxScore {
		return checker.CreateMaxScoreResult(name, "packaging workflow detected")
	}
	return checker.CreateInconclusiveResult(name, "packaging workflow not detected")
}
