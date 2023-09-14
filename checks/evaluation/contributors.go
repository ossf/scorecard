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
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/contributorsFromOrgOrCompany"
)

const (
	numberCompaniesForTopScore = 3
)

// Contributors applies the score policy for the Contributors check.
func Contributors(name string,
	findings []finding.Finding,
	dl checker.DetailLogger,
) checker.CheckResult {
	expectedProbes := []string{
		contributorsFromOrgOrCompany.Probe,
	}

	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	reason := fmt.Sprintf("project has %d contributing companies or organizations", len(findings))

	if len(findings) >= numberCompaniesForTopScore {
		// Return max score. This may need changing if other probes
		// are added for other contributors metrics. Right now, the
		// scoring is designed for a single probe that returns true
		// or false.
		checker.LogFindings(nonNegativeFindings(findings), dl)
		return checker.CreateMaxScoreResult(name, reason)
	}

	checker.LogFindings(negativeFindings(findings), dl)
	return checker.CreateProportionalScoreResult(name, reason, len(findings), numberCompaniesForTopScore)
}
