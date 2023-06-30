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
)

// SecurityPolicy applies the score policy for the Security-Policy check.
func SecurityPolicy(name string, findings []finding.Finding) checker.CheckResult {
	// We have 5 unique probes, each should have a finding.
	expectedProbes := []string{
		"securityPolicyContainsDisclosure", "securityPolicyContainsLinks",
		"securityPolicyContainsText", "securityPolicyPresentInOrg",
		"securityPolicyPresentInRepo",
	}
	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	score := 0
	filePresent := false
	for i := range findings {
		f := &findings[i]
		if f.Outcome == finding.OutcomePositive {
			switch f.Probe {
			case "securityPolicyContainsDisclosure":
				score += 1
			case "securityPolicyContainsLinks":
				score += 6
			case "securityPolicyContainsText":
				score += 3
			case "securityPolicyPresentInOrg":
				filePresent = true
			case "securityPolicyPresentInRepo":
				filePresent = true
			default:
				e := sce.WithMessage(sce.ErrScorecardInternal, "unknown probe results")
				return checker.CreateRuntimeErrorResult(name, e)
			}
		}
	}
	if !filePresent {
		if score > 0 {
			e := sce.WithMessage(sce.ErrScorecardInternal, "score calculation problem")
			return checker.CreateRuntimeErrorResult(name, e)
		}
		return checker.CreateMinScoreResult(name, "no security file found")
	}
	// Presence of the file gives 1 point.
	score += 1

	return checker.CreateResultWithScore(name, "security policy file found", score)
}
