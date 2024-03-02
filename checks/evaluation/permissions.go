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
	"github.com/ossf/scorecard/v4/probes/hasGitHubWorkflowPermissionNone"
	"github.com/ossf/scorecard/v4/probes/hasGitHubWorkflowPermissionRead"
	"github.com/ossf/scorecard/v4/probes/hasGitHubWorkflowPermissionUndeclared"
	"github.com/ossf/scorecard/v4/probes/hasGitHubWorkflowPermissionUnknown"
	"github.com/ossf/scorecard/v4/probes/hasNoGitHubWorkflowPermissionWriteAllJob"
	"github.com/ossf/scorecard/v4/probes/hasNoGitHubWorkflowPermissionWriteAllTop"
	"github.com/ossf/scorecard/v4/probes/jobLevelPermissions"
	"github.com/ossf/scorecard/v4/probes/topLevelPermissions"
)

// TokenPermissions applies the score policy for the Token-Permissions check.
func TokenPermissions(name string,
	findings []finding.Finding,
	dl checker.DetailLogger,
) checker.CheckResult {
	expectedProbes := []string{
		hasNoGitHubWorkflowPermissionWriteAllTop.Probe,
		hasGitHubWorkflowPermissionUnknown.Probe,
		hasGitHubWorkflowPermissionNone.Probe,
		hasGitHubWorkflowPermissionRead.Probe,
		hasGitHubWorkflowPermissionUndeclared.Probe,
		hasNoGitHubWorkflowPermissionWriteAllJob.Probe,
		jobLevelPermissions.Probe,
		topLevelPermissions.Probe,
	}
	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	// Start with a perfect score.
	score := float32(checker.MaxResultScore)

	// hasWritePermissions is a map that holds information about the
	// workflows in the project that have write permissions. It holds
	// information about the write permissions of jobs and at the
	// top-level too. The inner map (map[string]bool) has the
	// workflow path as its key, and the value determines whether
	// that workflow has write permissions at either "job" or "top"
	// level.
	hasWritePermissions := make(map[string]map[string]bool)
	hasWritePermissions["jobLevel"] = make(map[string]bool)
	hasWritePermissions["topLevel"] = make(map[string]bool)

	// undeclaredPermissions is a map that holds information about the
	// workflows in the project that have undeclared permissions. It holds
	// information about the undeclared permissions of jobs and at the
	// top-level too. The inner map (map[string]bool) has the
	// workflow path as its key, and the value determines whether
	// that workflow has undeclared permissions at either "job" or "top"
	// level.
	undeclaredPermissions := make(map[string]map[string]bool)
	undeclaredPermissions["jobLevel"] = make(map[string]bool)
	undeclaredPermissions["topLevel"] = make(map[string]bool)

	for i := range findings {
		f := &findings[i]

		// Log workflows with "none" and "read" permissions.
		if f.Outcome == finding.OutcomePositive &&
			(f.Probe == hasGitHubWorkflowPermissionNone.Probe ||
				f.Probe == hasGitHubWorkflowPermissionRead.Probe) {
			dl.Info(&checker.LogMessage{
				Finding: f,
			})
		}

		if notAvailableOrNotApplicable(f, dl) {
			return checker.CreateInconclusiveResult(name, "Token permissions are not available")
		}

		// If there are no TokenPermissions
		if f.Outcome == finding.OutcomeNotAvailable {
			return checker.CreateInconclusiveResult(name, "No tokens found")
		}

		if f.Outcome != finding.OutcomeNegative {
			continue
		}
		fPath := f.Location.Path

		addProbeToMaps(fPath, undeclaredPermissions, hasWritePermissions)

		switch f.Probe {
		case hasGitHubWorkflowPermissionUndeclared.Probe:
			score = updateScoreAndMapFromUndeclared(undeclaredPermissions,
				hasWritePermissions, f, score, dl)
		case hasNoGitHubWorkflowPermissionWriteAllTop.Probe:
			// If no top level permissions are defined, all the permissions
			// are enabled by default.
			dl.Warn(&checker.LogMessage{
				Finding: f,
			})
			hasWritePermissions["topLevel"][fPath] = true

			if hasWritePermissions["jobLevel"][fPath] ||
				undeclaredPermissions["jobLevel"][fPath] {
				return checker.CreateMinScoreResult(name, "detected GitHub workflow tokens with excessive permissions")
			}
			score -= 0.5
		case hasNoGitHubWorkflowPermissionWriteAllJob.Probe:
			dl.Warn(&checker.LogMessage{
				Finding: f,
			})
			hasWritePermissions["jobLevel"][fPath] = true
			if hasWritePermissions["topLevel"][fPath] {
				score = checker.MinResultScore
				break
			}
			if undeclaredPermissions["topLevel"][fPath] {
				score -= 0.5
			}
		case jobLevelPermissions.Probe:
			if f.Values["level"] == "write" {
				dl.Warn(&checker.LogMessage{
					Finding: f,
				})
				hasWritePermissions["jobLevel"][fPath] = true
			}
		default:
			score = logAndReduceScore(f, dl, score)
		}
	}
	if score < checker.MinResultScore {
		score = checker.MinResultScore
	}

	logIfNoWritePermissionsFound(hasWritePermissions, dl)

	if score != checker.MaxResultScore {
		return checker.CreateResultWithScore(name,
			"detected GitHub workflow tokens with excessive permissions", int(score))
	}

	return checker.CreateMaxScoreResult(name,
		"GitHub workflow tokens follow principle of least privilege")
}

func logIfNoWritePermissionsFound(hasWritePermissions map[string]map[string]bool,
	dl checker.DetailLogger,
) {
	foundWritePermissions := false
	for _, isWritePermission := range hasWritePermissions["jobLevel"] {
		if isWritePermission {
			foundWritePermissions = true
		}
	}
	if !foundWritePermissions {
		text := fmt.Sprintf("no %s write permissions found", checker.PermissionLocationJob)
		dl.Info(&checker.LogMessage{
			Text: text,
		})
	}
}

func updateScoreFromUndeclaredJob(undeclaredPermissions map[string]map[string]bool,
	hasWritePermissions map[string]map[string]bool,
	fPath string,
	score float32,
) float32 {
	if hasWritePermissions["topLevel"][fPath] ||
		undeclaredPermissions["topLevel"][fPath] {
		score = checker.MinResultScore
	}
	return score
}

func updateScoreFromUndeclaredTop(undeclaredPermissions map[string]map[string]bool,
	fPath string,
	score float32,
) float32 {
	if undeclaredPermissions["jobLevel"][fPath] {
		score = checker.MinResultScore
	} else {
		score -= 0.5
	}
	return score
}

func notAvailableOrNotApplicable(f *finding.Finding, dl checker.DetailLogger) bool {
	if f.Probe == hasGitHubWorkflowPermissionUndeclared.Probe {
		if f.Outcome == finding.OutcomeNotAvailable {
			return true
		} else if f.Outcome == finding.OutcomeNotApplicable {
			dl.Debug(&checker.LogMessage{
				Finding: f,
			})
			return false
		}
	}
	return false
}

func updateScoreAndMapFromUndeclared(undeclaredPermissions map[string]map[string]bool,
	hasWritePermissions map[string]map[string]bool,
	f *finding.Finding,
	score float32, dl checker.DetailLogger,
) float32 {
	fPath := f.Location.Path
	if f.Values["permissionLocation"] == string(checker.PermissionLocationJob) {
		dl.Debug(&checker.LogMessage{
			Finding: f,
		})
		undeclaredPermissions["jobLevel"][fPath] = true
		score = updateScoreFromUndeclaredJob(undeclaredPermissions,
			hasWritePermissions,
			fPath,
			score)
	} else if f.Values["permissionLocation"] == string(checker.PermissionLocationTop) {
		dl.Warn(&checker.LogMessage{
			Finding: f,
		})
		undeclaredPermissions["topLevel"][fPath] = true
		score = updateScoreFromUndeclaredTop(undeclaredPermissions,
			fPath,
			score)
	}
	return score
}

func addProbeToMaps(fPath string, hasWritePermissions, undeclaredPermissions map[string]map[string]bool) {
	if _, ok := undeclaredPermissions["jobLevel"][fPath]; !ok {
		undeclaredPermissions["jobLevel"][fPath] = false
	}
	if _, ok := undeclaredPermissions["topLevel"][fPath]; !ok {
		undeclaredPermissions["topLevel"][fPath] = false
	}
	if _, ok := hasWritePermissions["jobLevel"][fPath]; !ok {
		hasWritePermissions["jobLevel"][fPath] = false
	}
	if _, ok := hasWritePermissions["topLevel"][fPath]; !ok {
		hasWritePermissions["topLevel"][fPath] = false
	}
}

// Some probes with negative outcomes triggers logging and simple reduce in score.
// logAndReduceScore represents these cases.
func logAndReduceScore(f *finding.Finding, dl checker.DetailLogger, score float32) float32 {
	switch f.Probe {
	case hasGitHubWorkflowPermissionUnknown.Probe:
		dl.Debug(&checker.LogMessage{
			Finding: f,
		})
	case topLevelPermissions.Probe:
		if f.Values["level"] == "write" {
			score -= reduceBy(f, dl)
		}
	}
	return score
}

func reduceBy(f *finding.Finding, dl checker.DetailLogger) float32 {
	if f.Values["permissionLevel"] != "write" {
		return 0
	}
	tokenName := f.Values["tokenName"]
	switch tokenName {
	case "checks", "statuses":
		dl.Warn(&checker.LogMessage{
			Finding: f,
		})
		return 0.5
	case "contents", "packages", "actions":
		dl.Warn(&checker.LogMessage{
			Finding: f,
		})
		return checker.MaxResultScore
	case "deployments", "security-events":
		dl.Warn(&checker.LogMessage{
			Finding: f,
		})
		return 1.0
	}
	return 0
}
