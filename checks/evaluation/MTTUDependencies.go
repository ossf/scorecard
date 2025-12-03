// Copyright 2025 OpenSSF Scorecard Authors
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
	"fmt"

	"github.com/ossf/scorecard/v5/checker"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/MTTUDependenciesIsHigh"
	"github.com/ossf/scorecard/v5/probes/MTTUDependenciesIsLow"
	"github.com/ossf/scorecard/v5/probes/MTTUDependenciesIsVeryLow"
)

const CheckMTTUDependencies = "MTTUDependencies"

func MTTUDependencies(name string, findings []finding.Finding, dl checker.DetailLogger) checker.CheckResult {
	expectedProbes := []string{
		MTTUDependenciesIsHigh.Probe,
		MTTUDependenciesIsLow.Probe,
		MTTUDependenciesIsVeryLow.Probe,
	}
	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	score := 0
	reason := ""
	for i := range findings {
		f := &findings[i]
		switch f.Probe {
		case MTTUDependenciesIsHigh.Probe:
			if f.Outcome == finding.OutcomeTrue {
				score = 0
				reason = fmt.Sprintf("Mean time to update dependencies is high (>= 6 months): %s", f.Message)
			}
		case MTTUDependenciesIsLow.Probe:
			if f.Outcome == finding.OutcomeTrue {
				score = 5
				reason = fmt.Sprintf("Mean time to update dependencies is moderate (< 6 months, >= 3 months): %s", f.Message)
			}
		case MTTUDependenciesIsVeryLow.Probe:
			if f.Outcome == finding.OutcomeTrue {
				score = 10
				reason = fmt.Sprintf("Mean time to update dependencies is low (< 3 months): %s", f.Message)
			}
		}
	}

	return checker.CreateResultWithScore(CheckMTTUDependencies, reason, score)
}
