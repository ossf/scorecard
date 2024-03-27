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
	"github.com/ossf/scorecard/v4/probes/hasNoGitHubWorkflowPermissionUnknown"
	"github.com/ossf/scorecard/v4/probes/jobLevelPermissions"
	"github.com/ossf/scorecard/v4/probes/topLevelPermissions"
)

// TokenPermissions applies the score policy for the Token-Permissions check.
//
//nolint:gocognit
func TokenPermissions(name string,
	findings []finding.Finding,
	dl checker.DetailLogger,
) checker.CheckResult {
	expectedProbes := []string{
		hasNoGitHubWorkflowPermissionUnknown.Probe,
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

		// Log workflows with "none" permissions
		if permissionLevel(f) == checker.PermissionLevelNone {
			dl.Info(&checker.LogMessage{
				Finding: f,
			})
			continue
		}

		// Log workflows with "read" permissions
		if permissionLevel(f) == checker.PermissionLevelRead {
			dl.Info(&checker.LogMessage{
				Finding: f,
			})
		}

		if isBothUndeclaredAndNotAvailableOrNotApplicable(f, dl) {
			return checker.CreateInconclusiveResult(name, "Token permissions are not available")
		}

		// If there are no TokenPermissions
		if f.Outcome == finding.OutcomeNotApplicable {
			return checker.CreateInconclusiveResult(name, "No tokens found")
		}

		if f.Outcome != finding.OutcomeNegative {
			continue
		}
		if f.Location == nil {
			continue
		}
		fPath := f.Location.Path

		addProbeToMaps(fPath, undeclaredPermissions, hasWritePermissions)

		if permissionLevel(f) == checker.PermissionLevelUndeclared {
			score = updateScoreAndMapFromUndeclared(undeclaredPermissions,
				hasWritePermissions, f, score, dl)
			continue
		}

		switch f.Probe {
		case hasNoGitHubWorkflowPermissionUnknown.Probe:
			dl.Debug(&checker.LogMessage{
				Finding: f,
			})
		case topLevelPermissions.Probe:
			if permissionLevel(f) != checker.PermissionLevelWrite {
				continue
			}
			hasWritePermissions["topLevel"][fPath] = true

			if !isWriteAll(f) {
				score -= reduceBy(f, dl)
				continue
			}

			dl.Warn(&checker.LogMessage{
				Finding: f,
			})
			// "all" is evaluated separately. If the project also has write permissions
			// or undeclared permissions at the job level, this is particularly bad.
			if hasWritePermissions["jobLevel"][fPath] ||
				undeclaredPermissions["jobLevel"][fPath] {
				return checker.CreateMinScoreResult(name, "detected GitHub workflow tokens with excessive permissions")
			}
			score -= 0.5
		case jobLevelPermissions.Probe:
			if permissionLevel(f) != checker.PermissionLevelWrite {
				continue
			}

			dl.Warn(&checker.LogMessage{
				Finding: f,
			})
			hasWritePermissions["jobLevel"][fPath] = true

			// If project has "all" writepermissions too at top level, this is
			// particularly bad.
			if hasWritePermissions["topLevel"][fPath] {
				score = checker.MinResultScore
				break
			}
			// If project has not declared permissions at top level::
			if undeclaredPermissions["topLevel"][fPath] {
				score -= 0.5
			}
		default:
			continue
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

func isBothUndeclaredAndNotAvailableOrNotApplicable(f *finding.Finding, dl checker.DetailLogger) bool {
	if permissionLevel(f) == checker.PermissionLevelUndeclared {
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
	if f.Probe == jobLevelPermissions.Probe {
		dl.Debug(&checker.LogMessage{
			Finding: f,
		})
		undeclaredPermissions["jobLevel"][fPath] = true
		score = updateScoreFromUndeclaredJob(undeclaredPermissions,
			hasWritePermissions,
			fPath,
			score)
	} else if f.Probe == topLevelPermissions.Probe {
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

func reduceBy(f *finding.Finding, dl checker.DetailLogger) float32 {
	if permissionLevel(f) != checker.PermissionLevelWrite {
		return 0
	}
	switch tokenName(f) {
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

func isWriteAll(f *finding.Finding) bool {
	token := tokenName(f)
	return (token == "all" || token == "write-all")
}

func permissionLevel(f *finding.Finding) checker.PermissionLevel {
	var key string
	// these values should be the same, but better safe than sorry
	switch f.Probe {
	case jobLevelPermissions.Probe:
		key = jobLevelPermissions.PermissionLevelKey
	case topLevelPermissions.Probe:
		key = topLevelPermissions.PermissionLevelKey
	default:
	}
	return checker.PermissionLevel(f.Values[key])
}

func tokenName(f *finding.Finding) string {
	var key string
	// these values should be the same, but better safe than sorry
	switch f.Probe {
	case jobLevelPermissions.Probe:
		key = jobLevelPermissions.TokenNameKey
	case topLevelPermissions.Probe:
		key = topLevelPermissions.TokenNameKey
	default:
	}
	return f.Values[key]
}
