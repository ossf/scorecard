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
	"github.com/ossf/scorecard/v5/probes/securityPolicyContainsLinks"
	"github.com/ossf/scorecard/v5/probes/securityPolicyContainsText"
	"github.com/ossf/scorecard/v5/probes/securityPolicyContainsVulnerabilityDisclosure"
	"github.com/ossf/scorecard/v5/probes/securityPolicyEnablesPrivateReporting"
	"github.com/ossf/scorecard/v5/probes/securityPolicyPresent"
)

// SecurityPolicy applies the score policy for the Security-Policy check.
func SecurityPolicy(name string, findings []finding.Finding, dl checker.DetailLogger) checker.CheckResult {
	// We have 5 unique probes, each should have a finding.
	expectedProbes := []string{
		securityPolicyContainsVulnerabilityDisclosure.Probe,
		securityPolicyContainsLinks.Probe,
		securityPolicyContainsText.Probe,
		securityPolicyPresent.Probe,
		securityPolicyEnablesPrivateReporting.Probe,
	}
	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	pvrOutcome := finding.OutcomeNotApplicable
	for i := range findings {
		if findings[i].Probe == securityPolicyEnablesPrivateReporting.Probe {
			pvrOutcome = findings[i].Outcome
		}
	}

	score := 0
	m := make(map[string]bool)
	var logLevel checker.DetailType
	for i := range findings {
		f := &findings[i]
		// all of the security policy probes are good things if true and bad if false
		switch f.Outcome {
		case finding.OutcomeTrue:
			logLevel = checker.DetailInfo
			switch f.Probe {
			case securityPolicyContainsVulnerabilityDisclosure.Probe:
				score += scoreProbeOnce(f.Probe, m, 1)
			case securityPolicyContainsLinks.Probe:
				if pvrOutcome == finding.OutcomeNotApplicable {
					score += scoreProbeOnce(f.Probe, m, 6)
				} else {
					score += scoreProbeOnce(f.Probe, m, 5)
				}
			case securityPolicyContainsText.Probe:
				if pvrOutcome == finding.OutcomeNotApplicable {
					score += scoreProbeOnce(f.Probe, m, 3)
				} else {
					score += scoreProbeOnce(f.Probe, m, 2)
				}
			case securityPolicyPresent.Probe:
				m[f.Probe] = true
			case securityPolicyEnablesPrivateReporting.Probe:
				score += scoreProbeOnce(f.Probe, m, 2)
			default:
				e := sce.WithMessage(sce.ErrScorecardInternal, "unknown probe results")
				return checker.CreateRuntimeErrorResult(name, e)
			}
		case finding.OutcomeFalse:
			logLevel = checker.DetailWarn
		case finding.OutcomeNotApplicable:
			logLevel = checker.DetailInfo
		default:
			logLevel = checker.DetailDebug
		}
		checker.LogFinding(dl, f, logLevel)
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
