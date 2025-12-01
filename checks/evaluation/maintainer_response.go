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
	"regexp"
	"strconv"
	"strings"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
)

const (
	thresholdDays     = 180
	maxIssuesInReason = 20
)

// MaintainerResponse evaluates issue-level findings for the check.
// Semantics for scoring:
//   - OutcomeTrue  => maintainers responded within 180 days (NOT a violation).
//   - OutcomeFalse => maintainers did NOT respond within 180 days (violation).
//   - Others       => ignored (not part of denominator).
func MaintainerResponse(
	name string,
	findings []finding.Finding,
	dl checker.DetailLogger,
) checker.CheckResult {
	evaluated := 0
	violations := 0
	worstViolation := 0
	worstNonViolation := 0
	hasTrackedIssues := false
	var violatingIssues []int

	for i := range findings {
		f := findings[i]

		// Try to infer lag from a number present in the message (best-effort).
		lag, hasLag := parseAnyInt(f.Message)

		switch f.Outcome {
		case finding.OutcomeFalse:
			violations++
			evaluated++
			hasTrackedIssues = true

			// Track worst violation lag
			if hasLag && lag > worstViolation {
				worstViolation = lag
			}

			// Warn per violating finding, append URL if available.
			msg := f.Message
			if u := urlOf(&f); u != "" && !strings.Contains(msg, u) {
				msg = msg + " (" + u + ")"
			}
			dl.Warn(&checker.LogMessage{Text: msg})

			// Capture issue number for lists.
			if n := issueNumberOf(&f); n > 0 {
				violatingIssues = append(violatingIssues, n)
			}

		case finding.OutcomeTrue:
			evaluated++
			hasTrackedIssues = true

			// Track worst non-violation lag (capped at threshold since these didn't violate)
			if hasLag && lag < thresholdDays && lag > worstNonViolation {
				worstNonViolation = lag
			}

		default:
			// Ignore neutrals/unknowns
			continue
		}
	}

	// Nothing to evaluate → max score with explanatory reason.
	if evaluated == 0 {
		reason := getReasonsForNoEvaluation(findings, hasTrackedIssues)
		ret := checker.CreateResultWithScore(name, reason, checker.MaxResultScore)
		ret.Findings = findings
		return ret
	}

	// All issues were addressed within timeframe (0 violations).
	if violations == 0 {
		reason := fmt.Sprintf(
			"Evaluated %d issues with bug/security labels. "+
				"All %d had timely maintainer activity (no label went ≥%d days without response)",
			evaluated,
			evaluated,
			thresholdDays,
		)
		ret := checker.CreateResultWithScore(name, reason, checker.MaxResultScore)
		ret.Findings = findings
		return ret
	}

	percent := float64(violations) / float64(evaluated) * 100.0

	var score int
	switch {
	case percent > 40.0:
		score = 0
	case percent > 20.0:
		score = 5
	default:
		score = 10
	}

	// Calculate issues with maintainer activity.
	issuesWithActivity := evaluated - violations

	// Base reason with more informative text.
	reason := fmt.Sprintf(
		"Evaluated %d issues with bug/security labels. %d had activity by a maintainer within %d days",
		evaluated,
		issuesWithActivity,
		thresholdDays,
	)

	// Add worst non-violation lag if available
	if worstNonViolation > 0 {
		reason = fmt.Sprintf("%s (worst %d days)", reason, worstNonViolation)
	}

	// Add violation percentage if there are violations.
	if violations > 0 {
		reason = fmt.Sprintf("%s. %.1f%% exceeded %d days without response",
			reason,
			percent,
			thresholdDays,
		)
	}

	// Append a compact list of violating issues directly into the reason (up to 20).
	if len(violatingIssues) > 0 {
		list := formatIssueList(violatingIssues, maxIssuesInReason)
		reason = fmt.Sprintf("%s; violating issues: %s", reason, list)
	}

	// Single summary debug line.
	dl.Debug(&checker.LogMessage{
		Text: fmt.Sprintf("evaluated issues: %d; violations: %d", evaluated, violations),
	})

	// Full list of violating issues (debug).
	if len(violatingIssues) > 0 {
		dl.Debug(&checker.LogMessage{
			Text: fmt.Sprintf("issues exceeding %d days without response: %v", thresholdDays, violatingIssues),
		})
	}

	ret := checker.CreateResultWithScore(name, reason, score)
	ret.Findings = findings
	return ret
}

// ---- helpers ----

func urlOf(f *finding.Finding) string {
	if f.Location == nil {
		return ""
	}
	if f.Location.Path != "" && looksLikeURL(f.Location.Path) {
		return f.Location.Path
	}
	return ""
}

func looksLikeURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// issueNumberOf tries to extract an issue number from the URL (preferred) or message.
func issueNumberOf(f *finding.Finding) int {
	if u := urlOf(f); u != "" {
		if n := parseIssueNumberFromURL(u); n > 0 {
			return n
		}
	}
	if n, ok := parseAnyInt(f.Message); ok {
		return n
	}
	return 0
}

func parseIssueNumberFromURL(u string) int {
	i := strings.LastIndex(u, "/issues/")
	if i == -1 {
		return 0
	}
	numStr := u[i+len("/issues/"):]
	if j := strings.IndexByte(numStr, '/'); j != -1 {
		numStr = numStr[:j]
	}
	n, err := strconv.Atoi(numStr)
	if err != nil {
		return 0
	}
	return n
}

func parseAnyInt(s string) (int, bool) {
	re := regexp.MustCompile(`\d+`)
	m := re.FindString(s)
	if m == "" {
		return 0, false
	}
	v, err := strconv.Atoi(m)
	if err != nil {
		return 0, false
	}
	return v, true
}

func formatIssueList(nums []int, maxCount int) string {
	if len(nums) == 0 {
		return ""
	}
	if len(nums) > maxCount {
		return fmt.Sprintf("%s, ... +%d more", joinIssueNums(nums[:maxCount]), len(nums)-maxCount)
	}
	return joinIssueNums(nums)
}

func joinIssueNums(nums []int) string {
	parts := make([]string, 0, len(nums))
	for _, n := range nums {
		parts = append(parts, fmt.Sprintf("#%d", n))
	}
	return strings.Join(parts, ", ")
}

// getReasonsForNoEvaluation returns appropriate reason text when no issues were evaluated.
func getReasonsForNoEvaluation(findings []finding.Finding, hasTrackedIssues bool) string {
	if hasTrackedIssues {
		// This case shouldn't happen, but handle it defensively.
		return "no issues with tracked labels to evaluate"
	}
	// Check if we have any findings at all.
	if len(findings) == 0 {
		return "no issues found in repository"
	}
	return "no issues with bug/security labels found"
}
