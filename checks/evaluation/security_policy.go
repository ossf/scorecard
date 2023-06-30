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
	"github.com/ossf/scorecard/v4/probes/securityPolicyContainsLinks"
	"github.com/ossf/scorecard/v4/probes/securityPolicyContainsText"
	"github.com/ossf/scorecard/v4/probes/securityPolicyContainsVulnerabilityDisclosure"
	"github.com/ossf/scorecard/v4/probes/securityPolicyPresent"
)

// SecurityPolicy applies the score policy for the Security-Policy check.
func SecurityPolicy(name string, findings []finding.Finding) checker.CheckResult {
	// We have 4 unique probes, each should have a finding.
	expectedProbes := []string{
		securityPolicyContainsVulnerabilityDisclosure.Probe,
		securityPolicyContainsLinks.Probe,
		securityPolicyContainsText.Probe,
		securityPolicyPresent.Probe,
	}
	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	// The probes should always contain at least on finding.
	if len(findings) == 0 {
		e := sce.WithMessage(sce.ErrScorecardInternal, "no findings")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	score := 0
	m := make(map[string]bool)
	for i := range findings {
		f := &findings[i]
		if f.Outcome == finding.OutcomePositive {
			switch f.Probe {
			case securityPolicyContainsVulnerabilityDisclosure.Probe:
				score += scoreProbeOnce(f.Probe, m, 1)
			case securityPolicyContainsLinks.Probe:
				score += scoreProbeOnce(f.Probe, m, 6)
			case securityPolicyContainsText.Probe:
				score += scoreProbeOnce(f.Probe, m, 3)
			case securityPolicyPresent.Probe:
				m[f.Probe] = true
			default:
				e := sce.WithMessage(sce.ErrScorecardInternal, "unknown probe results")
				return checker.CreateRuntimeErrorResult(name, e)
			}
		}
	}
	_, defined := m[securityPolicyPresent.Probe]
	if !defined {
		if score > 0 {
			e := sce.WithMessage(sce.ErrScorecardInternal, "score calculation problem")
			return checker.CreateRuntimeErrorResult(name, e)
		}
		return checker.CreateMinScoreResult(name, "security policy file not detected")
	}

	return checker.CreateResultWithScore(name, "security policy file detected", score)
}

func scoreProbeOnce(probeID string, m map[string]bool, bump int) int {
	if _, exists := m[probeID]; !exists {
		m[probeID] = true
		return bump
	}
	return 0
}
