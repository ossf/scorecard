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
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/testsRunInCI"
)

const CheckCITests = "CI-Tests"

func CITests(name string,
	findings []finding.Finding,
	dl checker.DetailLogger,
) checker.CheckResult {
	expectedProbes := []string{
		testsRunInCI.Probe,
	}
	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	// Debug PRs that were merged without CI tests
	for i := range findings {
		f := &findings[i]
		if f.Outcome == finding.OutcomeFalse || f.Outcome == finding.OutcomeTrue {
			dl.Debug(&checker.LogMessage{
				Text: f.Message,
			})
		}
	}

	// check that the project has pull requests
	if noPullRequestsFound(findings) {
		return checker.CreateInconclusiveResult(CheckCITests, "no pull request found")
	}

	totalMerged, totalTested := getMergedAndTested(findings)

	if totalMerged < totalTested || len(findings) < totalTested {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid finding values")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	reason := fmt.Sprintf("%d out of %d merged PRs checked by a CI test", totalTested, totalMerged)
	return checker.CreateProportionalScoreResult(CheckCITests, reason, totalTested, totalMerged)
}

func getMergedAndTested(findings []finding.Finding) (int, int) {
	totalMerged := 0
	totalTested := 0

	for i := range findings {
		f := &findings[i]
		totalMerged++
		if f.Outcome == finding.OutcomeTrue {
			totalTested++
		}
	}

	return totalMerged, totalTested
}

func noPullRequestsFound(findings []finding.Finding) bool {
	for i := range findings {
		f := &findings[i]
		if f.Outcome == finding.OutcomeNotApplicable {
			return true
		}
	}
	return false
}
