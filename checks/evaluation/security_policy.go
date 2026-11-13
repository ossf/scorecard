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

// privateReportingScore is the floor when GitHub private vulnerability
// reporting is enabled and the policy file alone would score lower.
const privateReportingScore = 8

// SecurityPolicy applies the score policy for the Security-Policy check.
func SecurityPolicy(name string, findings []finding.Finding, dl checker.DetailLogger) checker.CheckResult {
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

	score, policyPresent, privateReporting, hasLinks := tallySecurityPolicy(findings, dl)
	if !policyPresent {
		if score > 0 {
			e := sce.WithMessage(sce.ErrScorecardInternal, "score calculation problem")
			return checker.CreateRuntimeErrorResult(name, e)
		}
		if privateReporting {
			return checker.CreateResultWithScore(name, "private vulnerability reporting enabled", privateReportingScore)
		}
		return checker.CreateMinScoreResult(name, "security policy file not detected")
	}
	return checker.CreateResultWithScore(name, "security policy file detected",
		applyPrivateReporting(score, privateReporting, hasLinks))
}

func tallySecurityPolicy(
	findings []finding.Finding, dl checker.DetailLogger,
) (score int, policyPresent, privateReporting, hasLinks bool) {
	seen := make(map[string]bool)
	for i := range findings {
		f := &findings[i]
		logLevel := checker.DetailDebug
		switch f.Outcome {
		case finding.OutcomeTrue:
			logLevel = checker.DetailInfo
			switch f.Probe {
			case securityPolicyContainsVulnerabilityDisclosure.Probe:
				score += scoreProbeOnce(f.Probe, seen, 1)
			case securityPolicyContainsLinks.Probe:
				score += scoreProbeOnce(f.Probe, seen, 6)
				hasLinks = true
			case securityPolicyContainsText.Probe:
				score += scoreProbeOnce(f.Probe, seen, 3)
			case securityPolicyPresent.Probe:
				policyPresent = true
			case securityPolicyEnablesPrivateReporting.Probe:
				privateReporting = true
			}
		case finding.OutcomeFalse:
			logLevel = checker.DetailWarn
		default:
			logLevel = checker.DetailDebug
		}
		checker.LogFinding(dl, f, logLevel)
	}
	return score, policyPresent, privateReporting, hasLinks
}

func applyPrivateReporting(score int, enabled, hasLinks bool) int {
	if !enabled {
		return score
	}
	if hasLinks {
		return checker.MaxResultScore
	}
	if score < privateReportingScore {
		return privateReportingScore
	}
	return score
}

func scoreProbeOnce(probeID string, m map[string]bool, bump int) int {
	if _, exists := m[probeID]; !exists {
		m[probeID] = true
		return bump
	}
	return 0
}
