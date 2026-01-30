// Copyright 2025 OpenSSF Scorecard Authors
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

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
	ghpp "github.com/ossf/scorecard/v5/probes/hasGitHubPushProtectionEnabled"
	ghe "github.com/ossf/scorecard/v5/probes/hasGitHubSecretScanningEnabled"
	glpd "github.com/ossf/scorecard/v5/probes/hasGitLabPipelineSecretDetection"
	glpr "github.com/ossf/scorecard/v5/probes/hasGitLabPushRulesPreventSecrets"
	glspp "github.com/ossf/scorecard/v5/probes/hasGitLabSecretPushProtection"
	tpds "github.com/ossf/scorecard/v5/probes/hasThirdPartyDetectSecrets"
	tpgg "github.com/ossf/scorecard/v5/probes/hasThirdPartyGGShield"
	tpgs "github.com/ossf/scorecard/v5/probes/hasThirdPartyGitSecrets"
	tpgl "github.com/ossf/scorecard/v5/probes/hasThirdPartyGitleaks"
	tprs "github.com/ossf/scorecard/v5/probes/hasThirdPartyRepoSupervisor"
	tpsh "github.com/ossf/scorecard/v5/probes/hasThirdPartyShhGit"
	tpth "github.com/ossf/scorecard/v5/probes/hasThirdPartyTruffleHog"
)

const thirdPartyScannerPresent = "; third-party scanner present"

// SecretScanning applies scoring policy for the Secret-Scanning check.
//
// This function evaluates whether a repository has secret scanning enabled to detect
// accidentally committed credentials, API keys, tokens, and other sensitive data.
// The scoring differs between GitHub and GitLab platforms and considers both native
// platform features and third-party scanning tools.
//
// Policy summary:
//
// GitHub repositories:
//   - Native secret scanning enabled → score 10 (maximum)
//     GitHub's native secret scanning provides comprehensive coverage and is the gold standard.
//   - Native disabled + 3rd-party tool present + actively running in CI → score based on CI coverage
//     Third-party tools are evaluated on how frequently they run (see third-party scoring below).
//   - Native disabled + 3rd-party tool present + no CI run data → score 1
//     Tool is configured but we cannot verify it's actually running.
//   - Native disabled + no 3rd-party tools detected → score 0
//     No secret scanning protection whatsoever.
//   - Native status inconclusive → fall through to GitLab-style additive scoring
//     When we can't determine GitHub native status, use component-based scoring.
//
// GitLab repositories (additive scoring model):
//   - Secret Push Protection enabled → +4 points
//   - Pipeline Secret Detection configured → +4 points
//   - Push Rules prevent secrets enabled → +1 point
//   - Third-party scanner present and running → +1 to +10 points (see third-party scoring below)
//
// Third-party CI scoring (applies to both platforms):
//
//	Periodic scanners (shhgit, repo-supervisor):
//	  These tools run on a schedule (daily/weekly) rather than per-commit.
//	  - Ran in last 30 days → +10 points
//	  - Present but no recent runs → +1 point
//
//	Commit-based scanners (Gitleaks, TruffleHog, detect-secrets, git-secrets, GGShield):
//	  These tools typically run on every push or pull request.
//	  Scored based on coverage of last 100 commits:
//	  - 100% coverage → +10 points (runs on every commit)
//	  - 70-99% coverage → +7 points (runs on most commits)
//	  - 50-69% coverage → +5 points (runs on half of commits)
//	  - <50% coverage → +3 points (runs on some commits)
//	  - Tool present but not running → +1 point (configured but inactive)
//
//nolint:gocognit,nestif // Complex scoring logic requires nested conditions
func SecretScanning(
	name string,
	findings []finding.Finding,
	dl checker.DetailLogger,
	rawData *checker.SecretScanningData,
) checker.CheckResult {
	// Map probeID -> outcome (each probe returns exactly one finding).
	po := map[string]finding.Outcome{}
	for _, f := range findings {
		po[f.Probe] = f.Outcome
	}

	// -----------------------
	// GitHub evaluation path
	// -----------------------
	// Take the GitHub path only when the native probe is definitively True or False.
	if ghOutcome, ok := po[ghe.Probe]; ok &&
		(ghOutcome == finding.OutcomeTrue || ghOutcome == finding.OutcomeFalse) {
		switch ghOutcome {
		case finding.OutcomeTrue:
			reason := "GitHub native secret scanning is enabled"
			if po[ghpp.Probe] == finding.OutcomeTrue {
				reason += " (push protection enabled)"
			}
			if tpPresent(po) {
				details := tpDetails(findings)
				if len(details) > 0 {
					reason += "; " + strings.Join(details, "; ")
				} else {
					reason += thirdPartyScannerPresent
				}
			}
			return checker.CreateMaxScoreResult(name, reason)

		case finding.OutcomeFalse:
			reason := "GitHub native secret scanning is disabled"
			if po[ghpp.Probe] == finding.OutcomeTrue {
				reason += " (push protection enabled)"
			}
			if tpPresent(po) {
				// Calculate score based on CI run frequency
				tpScore := calculateThirdPartyScore(rawData)
				details := tpDetails(findings)
				if len(details) > 0 {
					reason += "; " + strings.Join(details, "; ")
				} else {
					reason += thirdPartyScannerPresent
				}
				// Add per-tool CI coverage details to reason
				if rawData != nil && len(rawData.ThirdPartyCIInfo) > 0 {
					reason += formatCICoverageDetails(rawData.ThirdPartyCIInfo)
				}
				return checker.CreateResultWithScore(name, reason, tpScore)
			}
			return checker.CreateResultWithScore(name, reason, 0)
		default:
			// Handle other outcomes (NotAvailable, Error, NotSupported, NotApplicable)
			// Fall through to GitLab path or return minimal score
		}
	}

	// Check if we're on GitHub but couldn't determine native status due to permissions
	if rawData != nil && rawData.Platform == "github" {
		for _, evidence := range rawData.Evidence {
			if strings.Contains(evidence, "permission_denied") {
				// Token doesn't have permissions to check GitHub native secret scanning
				reason := "Token has insufficient permissions to get information about native GitHub secret scanning"
				if tpPresent(po) {
					details := tpDetails(findings)
					if len(details) > 0 {
						reason += "; " + strings.Join(details, "; ")
					} else {
						reason += thirdPartyScannerPresent
					}
					if rawData != nil && len(rawData.ThirdPartyCIInfo) > 0 {
						reason += formatCICoverageDetails(rawData.ThirdPartyCIInfo)
					}
				}
				return checker.CreateInconclusiveResult(name, reason)
			}
		}
	}

	// -----------------------
	// GitLab evaluation path
	// -----------------------
	score := 0
	var bits []string

	add := func(probe string, on string, off string, pts int) {
		if po[probe] == finding.OutcomeTrue {
			score += pts
			bits = append(bits, on)
		} else {
			bits = append(bits, off)
		}
	}

	// Native GitLab knobs
	add(glspp.Probe, "Secret Push Protection: on", "Secret Push Protection: off", 4)
	add(glpd.Probe, "Pipeline Secret Detection: on", "Pipeline Secret Detection: off", 4)
	add(glpr.Probe, "Push rules prevent_secrets: on", "Push rules prevent_secrets: off", 1)

	// Third-party scanner with CI-aware scoring
	if tpPresent(po) {
		tpScore := calculateThirdPartyScore(rawData)
		score += tpScore
		details := tpDetails(findings)
		if len(details) > 0 {
			bits = append(bits, "3rd-party scanner: "+strings.Join(details, "; "))
		} else {
			bits = append(bits, "3rd-party scanner: present")
		}
		// Add per-tool CI coverage details
		if rawData != nil && len(rawData.ThirdPartyCIInfo) > 0 {
			coverageDetails := formatCICoverageDetails(rawData.ThirdPartyCIInfo)
			if coverageDetails != "" {
				bits = append(bits, "CI stats:"+coverageDetails)
			}
		}
	} else {
		bits = append(bits, "3rd-party scanner: not found")
	}

	if score > checker.MaxResultScore {
		score = checker.MaxResultScore
	}
	reason := "GitLab secret scanning posture — " + strings.Join(bits, "; ")
	return checker.CreateResultWithScore(name, reason, score)
}

// tpPresent returns true if any third-party secret scanner
// probe is OutcomeTrue.
func tpPresent(po map[string]finding.Outcome) bool {
	return po[tpgl.Probe] == finding.OutcomeTrue ||
		po[tpth.Probe] == finding.OutcomeTrue ||
		po[tpds.Probe] == finding.OutcomeTrue ||
		po[tpgs.Probe] == finding.OutcomeTrue ||
		po[tpgg.Probe] == finding.OutcomeTrue ||
		po[tpsh.Probe] == finding.OutcomeTrue ||
		po[tprs.Probe] == finding.OutcomeTrue
}

// tpDetails gathers human-readable messages from the third-party probes,
// which are already formatted as "<tool> found at <path>" when a path is known.
func tpDetails(findings []finding.Finding) []string {
	var out []string
	want := map[string]struct{}{
		tpgl.Probe: {},
		tpth.Probe: {},
		tpds.Probe: {},
		tpgs.Probe: {},
		tpgg.Probe: {},
		tpsh.Probe: {},
		tprs.Probe: {},
	}
	for _, f := range findings {
		if _, ok := want[f.Probe]; !ok {
			continue
		}
		if f.Outcome != finding.OutcomeTrue {
			continue
		}
		if f.Message != "" {
			out = append(out, f.Message)
		}
	}
	return out
}

// calculateThirdPartyScore calculates the score contribution from third-party
// secret scanning tools based on their CI execution frequency and patterns.
//
// This function analyzes how actively third-party tools are used, not just whether
// they're configured. The scoring differs by tool type because different tools have
// different intended execution patterns.
//
// Scoring policy:
//
// Periodic scanners (shhgit, repo-supervisor):
//
//	These tools are designed to run on a regular schedule (e.g., nightly or weekly)
//	rather than on every commit. We check the last 30 days of activity:
//	- Tool ran within last 30 days → 10 points
//	  This indicates active, regular scanning on a reasonable schedule.
//	- Tool configured but no runs in last 30 days → 1 point
//	  Tool exists but appears inactive or misconfigured.
//
// Commit-based scanners (Gitleaks, TruffleHog, detect-secrets, git-secrets, GGShield):
//
//	These tools are typically configured to run on every push or pull request.
//	We analyze the last 100 commits to calculate coverage percentage:
//	- 100% coverage (ran on every commit) → 10 points
//	  Excellent: every code change is scanned before merge.
//	- 70-99% coverage → 7 points
//	  Good: most commits are scanned, minor gaps acceptable.
//	- 50-69% coverage → 5 points
//	  Fair: tool runs regularly but with significant gaps.
//	- <50% coverage → 3 points
//	  Poor: tool runs occasionally but not consistently.
//	- Tool configured but no CI run data available → 1 point
//	  We cannot verify the tool is actually running.
//
// Returns the highest score from all detected third-party tools.
func calculateThirdPartyScore(rawData *checker.SecretScanningData) int {
	if rawData == nil || len(rawData.ThirdPartyCIInfo) == 0 {
		return 1 // Tool present, but no data
	}

	// Calculate the highest score from all detected tools
	maxScore := 1
	for toolName, stats := range rawData.ThirdPartyCIInfo {
		score := scoreForTool(toolName, stats)
		if score > maxScore {
			maxScore = score
		}
	}

	return maxScore
}

// scoreForTool calculates the score for a specific tool
// based on its execution pattern.
func scoreForTool(toolName string, stats *checker.ToolCIStats) int {
	if stats == nil || stats.TotalCommitsAnalyzed == 0 {
		return 1 // Tool present but no CI data
	}

	if stats.ExecutionPattern == "periodic" {
		// Periodic tools: check if they ran in last 30 days
		if stats.HasRecentRuns {
			return 10 // Ran recently
		}
		return 1 // Present but not running recently
	}

	// Commit-based tools: score based on coverage percentage
	var coverage float64
	if stats.CommitsWithToolRun > 0 {
		coverage = float64(stats.CommitsWithToolRun) /
			float64(stats.TotalCommitsAnalyzed)
	}

	return scoreFromCoverage(coverage)
}

// scoreFromCoverage returns a score based on commit coverage
// percentage for commit-based tools.
func scoreFromCoverage(coverage float64) int {
	switch {
	case coverage >= 1.0:
		return 10 // 100% coverage
	case coverage >= 0.70:
		return 7 // 70-99% coverage
	case coverage >= 0.50:
		return 5 // 50-69% coverage
	case coverage > 0:
		return 3 // Some coverage but less than 50%
	default:
		return 1 // Tool present but not running
	}
}

// formatCICoverageDetails formats CI coverage information
// for display in the reason string.
func formatCICoverageDetails(ciInfo map[string]*checker.ToolCIStats) string {
	if len(ciInfo) == 0 {
		return ""
	}

	var details []string
	for toolName, stats := range ciInfo {
		if stats == nil || stats.TotalCommitsAnalyzed == 0 {
			continue
		}

		if stats.ExecutionPattern == "periodic" {
			if stats.HasRecentRuns {
				details = append(
					details,
					fmt.Sprintf("%s: ran recently", toolName),
				)
			} else {
				details = append(
					details,
					fmt.Sprintf("%s: no recent runs", toolName),
				)
			}
		} else {
			// Commit-based tool
			coverage := float64(stats.CommitsWithToolRun) /
				float64(stats.TotalCommitsAnalyzed) * 100
			details = append(
				details,
				fmt.Sprintf("%s: %.0f%% coverage", toolName, coverage),
			)
		}
	}

	if len(details) == 0 {
		return ""
	}

	return " (" + strings.Join(details, ", ") + ")"
}
