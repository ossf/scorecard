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

	// check whether there are any negative findings
	if noPullRequestsFound(findings) {
		return checker.CreateInconclusiveResult(CheckCITests, "no pull request found")
	}

	totalMerged := findings[0].Values["totalMerged"]
	totalTested := findings[0].Values["totalTested"]

	if totalMerged < totalTested || len(findings) != totalTested {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	// Log all findings
	checker.LogFindings(nonNegativeFindings(findings), dl)

	// All findings (should) have "totalMerged" and "totalTested"
	reason := fmt.Sprintf("%d out of %d merged PRs checked by a CI test", totalTested, totalMerged)
	return checker.CreateProportionalScoreResult(CheckCITests, reason, totalTested, totalMerged)
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
