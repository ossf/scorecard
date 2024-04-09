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
	"github.com/ossf/scorecard/v4/probes/hasOSVVulnerabilities"
)

// Vulnerabilities applies the score policy for the Vulnerabilities check.
func Vulnerabilities(name string,
	findings []finding.Finding,
	dl checker.DetailLogger,
) checker.CheckResult {
	expectedProbes := []string{
		hasOSVVulnerabilities.Probe,
	}

	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	var numVulnsFound int
	for i := range findings {
		f := &findings[i]
		// TODO(#3654), this needs to be swapped. But it's a complicated swap so doing it not in here.
		if f.Outcome == finding.OutcomeFalse {
			numVulnsFound++
			checker.LogFinding(dl, f, checker.DetailWarn)
		}
	}

	score := checker.MaxResultScore - numVulnsFound

	if score < checker.MinResultScore {
		score = checker.MinResultScore
	}

	return checker.CreateResultWithScore(name,
		fmt.Sprintf("%v existing vulnerabilities detected", numVulnsFound), score)
}
