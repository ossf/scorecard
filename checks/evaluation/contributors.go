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
	"math"
	"slices"
	"strings"

	"github.com/ossf/scorecard/v5/checker"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/contributorsFromCodeOwners"
	"github.com/ossf/scorecard/v5/probes/contributorsFromOrgOrCompany"
)

const (
	numberCompaniesForTopScore  = 3
	numberCodeOwnersForTopScore = 3
)

// Contributors applies the score policy for the Contributors check.
func Contributors(name string,
	findings []finding.Finding,
	dl checker.DetailLogger,
) checker.CheckResult {
	expectedProbes := []string{
		contributorsFromOrgOrCompany.Probe,
		contributorsFromCodeOwners.Probe,
	}

	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	// only allow max 3 for each
	numberOfTrueCompanies := int(
		math.Min(
			float64(getNumberOfTrue(findings, contributorsFromOrgOrCompany.Probe)),
			float64(numberCompaniesForTopScore),
		),
	)
	numberOfTrueOwners := int(
		math.Min(
			float64(getNumberOfTrue(findings, contributorsFromCodeOwners.Probe)),
			float64(numberCodeOwnersForTopScore),
		),
	)
	reason := fmt.Sprintf(
		"project has %d contributing companies or organizations and %d contributing code owners",
		numberOfTrueCompanies,
		numberOfTrueOwners,
	)

	if numberOfTrueCompanies+numberOfTrueOwners > 0 {
		logFindings(findings, dl)
	}
	if numberOfTrueCompanies >= numberCompaniesForTopScore && numberOfTrueOwners >= numberCodeOwnersForTopScore {
		return checker.CreateMaxScoreResult(name, reason)
	}

	return checker.CreateProportionalScoreResult(
		name,
		reason,
		numberOfTrueCompanies+numberOfTrueOwners,
		numberCompaniesForTopScore+numberCodeOwnersForTopScore,
	)
}

func getNumberOfTrue(findings []finding.Finding, probe string) int {
	var numberOfTrue int
	for i := range findings {
		f := &findings[i]
		if f.Outcome == finding.OutcomeTrue && f.Probe == probe {
			if f.Probe == contributorsFromOrgOrCompany.Probe || f.Probe == contributorsFromCodeOwners.Probe {
				numberOfTrue++
			}
		}
	}
	return numberOfTrue
}

func logFindings(findings []finding.Finding, dl checker.DetailLogger) {
	var entities []string
	for i := range findings {
		f := &findings[i]
		if f.Outcome == finding.OutcomeTrue {
			entity := strings.TrimSuffix(f.Message, " code owner contributor found")
			entity = strings.TrimSuffix(entity, " contributor org/company found")
			if entity != "" {
				entities = append(entities, entity)
			}
		}
	}
	slices.Sort(entities)
	dl.Info(&checker.LogMessage{
		Text: "found contributions from: " + strings.Join(entities, ", "),
	})
}
