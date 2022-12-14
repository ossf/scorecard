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
	"strings"

	"github.com/google/osv-scanner/pkg/grouper"

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
)

// Vulnerabilities applies the score policy for the Vulnerabilities check.
func Vulnerabilities(name string, dl checker.DetailLogger,
	r *checker.VulnerabilitiesData,
) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	aliasVulnerabilities := []grouper.IDAliases{}
	for _, vuln := range r.Vulnerabilities {
		aliasVulnerabilities = append(aliasVulnerabilities, grouper.IDAliases(vuln))
	}

	IDs := grouper.Group(aliasVulnerabilities)
	score := checker.MaxResultScore - len(IDs)

	if score < checker.MinResultScore {
		score = checker.MinResultScore
	}

	if len(IDs) > 0 {
		for _, v := range IDs {
			dl.Warn(&checker.LogMessage{
				Text: fmt.Sprintf("Project is vulnerable to: %s", strings.Join(v.IDs, " / ")),
			})
		}

		return checker.CreateResultWithScore(name,
			fmt.Sprintf("%v existing vulnerabilities detected", len(IDs)), score)
	}

	return checker.CreateMaxScoreResult(name, "no vulnerabilities detected")
}
