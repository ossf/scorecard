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

	"github.com/ossf/scorecard/v5/checker"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/blocksDeleteOnTags"
	"github.com/ossf/scorecard/v5/probes/blocksForcePushOnTags"
	"github.com/ossf/scorecard/v5/probes/blocksUpdateOnTags"
	"github.com/ossf/scorecard/v5/probes/gitlabReleaseTagsAreProtected"
	"github.com/ossf/scorecard/v5/probes/requiresSignedTags"
	"github.com/ossf/scorecard/v5/probes/restrictsTagCreation"
	"github.com/ossf/scorecard/v5/probes/tagProtectionAppliesToAdmins"
	"github.com/ossf/scorecard/v5/probes/tagsAreProtected"
	"github.com/ossf/scorecard/v5/probes/tagsCannotDuplicateBranchNames"
)

type protectionMetrics struct {
	protected              int
	protectedTotal         int
	blocksDelete           int
	blocksDeleteTotal      int
	blocksForcePush        int
	blocksForcePushTotal   int
	blocksUpdate           int
	blocksUpdateTotal      int
	appliesToAdmins        int
	appliesToAdminsTotal   int
	restrictsCreation      int
	restrictsCreationTotal int
	requiresSigned         int
	requiresSignedTotal    int
}

func countFindings(findings []finding.Finding) *protectionMetrics {
	metrics := &protectionMetrics{}

	for i := range findings {
		f := &findings[i]
		// Skip NotApplicable outcomes (no tags to check)
		if f.Outcome == finding.OutcomeNotApplicable {
			continue
		}

		isTrue := f.Outcome == finding.OutcomeTrue

		switch f.Probe {
		case tagsAreProtected.Probe:
			if isTrue {
				metrics.protected++
			}
			metrics.protectedTotal++
		case blocksDeleteOnTags.Probe:
			if isTrue {
				metrics.blocksDelete++
			}
			metrics.blocksDeleteTotal++
		case blocksForcePushOnTags.Probe:
			if isTrue {
				metrics.blocksForcePush++
			}
			metrics.blocksForcePushTotal++
		case blocksUpdateOnTags.Probe:
			if isTrue {
				metrics.blocksUpdate++
			}
			metrics.blocksUpdateTotal++
		case tagProtectionAppliesToAdmins.Probe:
			if isTrue {
				metrics.appliesToAdmins++
			}
			metrics.appliesToAdminsTotal++
		case restrictsTagCreation.Probe:
			if isTrue {
				metrics.restrictsCreation++
			}
			metrics.restrictsCreationTotal++
		case requiresSignedTags.Probe:
			if isTrue {
				metrics.requiresSigned++
			}
			metrics.requiresSignedTotal++
		}
	}

	return metrics
}

// logUnprotectedTags logs debug messages for tags that lack
// specific protections.
func logUnprotectedTags(
	dl checker.DetailLogger,
	findings []finding.Finding,
	probeName string,
	featureName string,
) {
	for i := range findings {
		f := &findings[i]
		if f.Probe == probeName && f.Outcome == finding.OutcomeFalse {
			tagName := "unknown"
			// Tag name is stored in the Values map with key "tagName"
			if f.Values != nil {
				if name, ok := f.Values["tagName"]; ok && name != "" {
					tagName = name
				}
			}
			dl.Debug(&checker.LogMessage{
				Text: "Tag '" + tagName + "' lacks " + featureName,
			})
		}
	}
}

func isFullyProtected(count, total int) bool {
	return total > 0 && count == total
}

func logProtectionStatus(
	dl checker.DetailLogger,
	feature string,
	protected bool,
	total int,
) {
	if protected {
		dl.Info(&checker.LogMessage{
			Text: feature + " on all release tags",
		})
	} else if total > 0 {
		dl.Warn(&checker.LogMessage{
			Text: "Not " + feature + " on all release tags",
		})
	}
}

// TagProtection runs Tag-Protection check.
func TagProtection(
	name string,
	findings []finding.Finding,
	dl checker.DetailLogger,
) checker.CheckResult {
	if isGitLabRepository(findings) {
		return evaluateGitLabTagProtection(name, findings, dl)
	}
	return evaluateGitHubTagProtection(name, findings, dl)
}

// isGitLabRepository checks if findings indicate a GitLab repo.
func isGitLabRepository(findings []finding.Finding) bool {
	for i := range findings {
		f := &findings[i]
		if isGitLabProbe(f.Probe) && isApplicable(f.Outcome) {
			return true
		}
	}
	return false
}

// isGitLabProbe checks if a probe is GitLab-specific.
func isGitLabProbe(probeName string) bool {
	return probeName == tagsCannotDuplicateBranchNames.Probe ||
		probeName == gitlabReleaseTagsAreProtected.Probe
}

// isApplicable checks if an outcome is applicable.
func isApplicable(outcome finding.Outcome) bool {
	return outcome != finding.OutcomeNotApplicable
}

// evaluateGitLabTagProtection evaluates tag protection for GitLab.
//
// Scoring model (0-2-4-8-10):
//
// Branch shadowing protection (0/1/2 points):
// - 2 points: All branches protected with "No one" access
// - 1 point: All branches protected with "Maintainer+" access
// - 0 points: Otherwise.
//
// Release tag protection (0/4/8 points):
// - 8 points: All release tags protected with "No one" access
// - 4 points: All release tags protected with "Maintainer+" access
// - 0 points: Otherwise.
func evaluateGitLabTagProtection(
	name string,
	findings []finding.Finding,
	dl checker.DetailLogger,
) checker.CheckResult {
	if !hasReleaseTags(findings) {
		return checker.CreateInconclusiveResult(
			name,
			"no release tags found",
		)
	}

	score := 0
	score += evaluateBranchShadowingProtection(findings, dl)
	score += evaluateGitLabReleaseTagProtection(findings, dl)

	return checker.CreateProportionalScoreResult(name, "", score, 10)
}

// hasReleaseTags checks if there are release tags to evaluate.
func hasReleaseTags(findings []finding.Finding) bool {
	for i := range findings {
		f := &findings[i]
		if isGitLabProbe(f.Probe) && isApplicable(f.Outcome) {
			return true
		}
	}
	return false
}

// evaluateBranchShadowingProtection evaluates branch shadowing.
// Returns 0, 1, or 2 points.
func evaluateBranchShadowingProtection(
	findings []finding.Finding,
	dl checker.DetailLogger,
) int {
	counts := countProtectionLevels(
		findings,
		tagsCannotDuplicateBranchNames.Probe,
	)

	if counts.total == 0 {
		dl.Warn(&checker.LogMessage{
			Text: "No branches found for shadowing evaluation",
		})
		return 0
	}

	// All branches with "No one" protection = 2 points
	if counts.strongest == counts.total {
		dl.Info(&checker.LogMessage{
			Text: "All branches fully protected from tag shadowing",
		})
		return 2
	}

	// All branches with "Maintainer+" protection = 1 point
	if counts.strongest+counts.strong == counts.total {
		dl.Info(&checker.LogMessage{
			Text: "All branches protected from tag shadowing",
		})
		return 1
	}

	// Partial protection
	logPartialProtection(dl, counts, "branches", "tag shadowing")
	return 0
}

// evaluateGitLabReleaseTagProtection evaluates release tags.
// Returns 0, 4, or 8 points.
func evaluateGitLabReleaseTagProtection(
	findings []finding.Finding,
	dl checker.DetailLogger,
) int {
	counts := countProtectionLevels(
		findings,
		gitlabReleaseTagsAreProtected.Probe,
	)

	if counts.total == 0 {
		dl.Warn(&checker.LogMessage{
			Text: "No release tags found to evaluate",
		})
		return 0
	}

	// All tags with "No one" protection = 8 points
	if counts.strongest == counts.total {
		dl.Info(&checker.LogMessage{
			Text: "All release tags fully protected",
		})
		return 8
	}

	// All tags with "Maintainer+" protection = 4 points
	if counts.strongest+counts.strong == counts.total {
		dl.Info(&checker.LogMessage{
			Text: "All release tags protected",
		})
		return 4
	}

	dl.Warn(&checker.LogMessage{
		Text: "Some release tags lack adequate restrictions",
	})
	return 0
}

// protectionCounts holds protection level statistics.
type protectionCounts struct {
	strongest int
	strong    int
	weak      int
	total     int
}

// countProtectionLevels counts findings by protection level.
func countProtectionLevels(
	findings []finding.Finding,
	probeName string,
) protectionCounts {
	var counts protectionCounts

	for i := range findings {
		f := &findings[i]
		if f.Probe != probeName {
			continue
		}
		if f.Outcome == finding.OutcomeNotApplicable {
			continue
		}

		counts.total++

		if f.Outcome == finding.OutcomeTrue {
			level := getProtectionLevel(f)
			switch level {
			case "strongest":
				counts.strongest++
			case "strong":
				counts.strong++
			}
		} else {
			counts.weak++
		}
	}

	return counts
}

// getProtectionLevel extracts protection level from finding.
func getProtectionLevel(f *finding.Finding) string {
	if f.Values == nil {
		return ""
	}
	return f.Values["protectionLevel"]
}

// logPartialProtection logs info about partial protection.
func logPartialProtection(
	dl checker.DetailLogger,
	counts protectionCounts,
	resourceType string,
	protectionType string,
) {
	if counts.strongest > 0 || counts.strong > 0 {
		protected := counts.strongest + counts.strong
		dl.Info(&checker.LogMessage{
			Text: fmt.Sprintf(
				"%d out of %d %s have some %s protection",
				protected,
				counts.total,
				resourceType,
				protectionType,
			),
		})
	}
	dl.Warn(&checker.LogMessage{
		Text: fmt.Sprintf(
			"Some %s lack adequate %s protection",
			resourceType,
			protectionType,
		),
	})
}

// evaluateGitHubTagProtection evaluates tag protection for GitHub.
func evaluateGitHubTagProtection(
	name string,
	findings []finding.Finding,
	dl checker.DetailLogger,
) checker.CheckResult {
	// Filter out GitLab-specific probes
	// (they return NotApplicable for GitHub repos)
	var githubFindings []finding.Finding
	for i := range findings {
		f := &findings[i]
		// Skip GitLab-specific probes
		if f.Probe == tagsCannotDuplicateBranchNames.Probe ||
			f.Probe == gitlabReleaseTagsAreProtected.Probe {
			continue
		}
		githubFindings = append(githubFindings, *f)
	}

	expectedProbes := []string{
		tagsAreProtected.Probe,
		blocksDeleteOnTags.Probe,
		blocksForcePushOnTags.Probe,
		blocksUpdateOnTags.Probe,
		tagProtectionAppliesToAdmins.Probe,
		restrictsTagCreation.Probe,
		requiresSignedTags.Probe,
	}

	if !finding.UniqueProbesEqual(githubFindings, expectedProbes) {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	m := countFindings(githubFindings)

	// If no tags found, return inconclusive score (-1)
	if m.protectedTotal == 0 {
		return checker.CreateInconclusiveResult(name, "no release tags found")
	}

	// Calculate score using tiered approach (similar to Branch Protection)
	// Each tier must be fully satisfied before points are awarded for
	// the next tier.
	score := 0
	maxScore := 10

	// Tier 1: Base requirement - if not all tags are protected, score is 0
	if !isFullyProtected(m.protected, m.protectedTotal) {
		dl.Warn(&checker.LogMessage{
			Text: "Not all release tags are protected",
		})
		// Debug: show which tags are unprotected
		logUnprotectedTags(dl, githubFindings, tagsAreProtected.Probe, "protection")
		return checker.CreateMinScoreResult(name, "not all release tags are protected")
	}

	// Tier 1 complete: All tags are protected - award base 3 points
	score = 3
	dl.Info(&checker.LogMessage{
		Text: "All release tags are protected",
	})

	// Tier 2: Delete and force push protection (must have BOTH to proceed)
	// 6 points total
	deleteProtected := isFullyProtected(m.blocksDelete, m.blocksDeleteTotal)
	forcePushProtected := isFullyProtected(m.blocksForcePush, m.blocksForcePushTotal)

	logProtectionStatus(
		dl,
		"Tag deletion is blocked",
		deleteProtected,
		m.blocksDeleteTotal,
	)
	logProtectionStatus(
		dl,
		"Force push is blocked",
		forcePushProtected,
		m.blocksForcePushTotal,
	)

	// Debug: show which tags lack delete/force-push protection
	if !deleteProtected {
		logUnprotectedTags(
			dl,
			githubFindings,
			blocksDeleteOnTags.Probe,
			"delete protection",
		)
	}
	if !forcePushProtected {
		logUnprotectedTags(
			dl,
			githubFindings,
			blocksForcePushOnTags.Probe,
			"force-push protection",
		)
	}

	// Must have both delete and force push protection to advance
	if !deleteProtected || !forcePushProtected {
		return checker.CreateProportionalScoreResult(name, "", score, maxScore)
	}
	score = 6

	// Tier 3: Update protection - 8 points total
	updateProtected := isFullyProtected(m.blocksUpdate, m.blocksUpdateTotal)
	logProtectionStatus(
		dl,
		"Tag updates are blocked",
		updateProtected,
		m.blocksUpdateTotal,
	)

	// Debug: show which tags lack update protection
	if !updateProtected {
		logUnprotectedTags(
			dl,
			githubFindings,
			blocksUpdateOnTags.Probe,
			"update protection",
		)
	}

	if updateProtected {
		score = 8
	} else {
		return checker.CreateProportionalScoreResult(name, "", score, maxScore)
	}

	// Tier 4: Admin enforcement and creation restriction (must have BOTH)
	// 10 points total
	adminProtected := isFullyProtected(m.appliesToAdmins, m.appliesToAdminsTotal)
	creationRestricted := isFullyProtected(m.restrictsCreation, m.restrictsCreationTotal)

	logProtectionStatus(
		dl,
		"Tag protection applies to administrators",
		adminProtected,
		m.appliesToAdminsTotal,
	)
	logProtectionStatus(
		dl,
		"Tag creation is restricted",
		creationRestricted,
		m.restrictsCreationTotal,
	)

	// Debug: show which tags lack admin enforcement or creation restriction
	if !adminProtected {
		logUnprotectedTags(
			dl,
			findings,
			tagProtectionAppliesToAdmins.Probe,
			"admin enforcement",
		)
	}
	if !creationRestricted {
		logUnprotectedTags(
			dl,
			findings,
			restrictsTagCreation.Probe,
			"creation restriction",
		)
	}

	// Must have both admin enforcement and creation restriction for max score
	if adminProtected && creationRestricted {
		score = 10
	}

	// Requires signed tags: bonus points (doesn't reduce score if not present)
	// This is checked but doesn't affect the base score
	if isFullyProtected(m.requiresSigned, m.requiresSignedTotal) {
		dl.Info(&checker.LogMessage{
			Text: "Signed tags are required on all release tags",
		})
	} else if m.requiresSignedTotal > 0 {
		dl.Debug(&checker.LogMessage{
			Text: "Signed tags are not required on all release tags (optional security enhancement)",
		})
	}

	return checker.CreateProportionalScoreResult(name, "", score, maxScore)
}
