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
	"embed"
	"fmt"
	"strings"

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/remediation"
)

//go:embed *.yml
var rules embed.FS

type permissions struct {
	topLevelWritePermissions map[string]bool
	jobLevelWritePermissions map[string]bool
}

// TokenPermissions applies the score policy for the Token-Permissions check.
func TokenPermissions(name string, c *checker.CheckRequest, r *checker.TokenPermissionsData) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	score, err := applyScorePolicy(r, c)
	if err != nil {
		return checker.CreateRuntimeErrorResult(name, err)
	}

	if score != checker.MaxResultScore {
		return checker.CreateResultWithScore(name,
			"non read-only tokens detected in GitHub workflows", score)
	}

	return checker.CreateMaxScoreResult(name,
		"tokens are read-only in GitHub workflows")
}

func applyScorePolicy(results *checker.TokenPermissionsData, c *checker.CheckRequest) (int, error) {
	// See list https://github.blog/changelog/2021-04-20-github-actions-control-permissions-for-github_token/.
	// Note: there are legitimate reasons to use some of the permissions like checks, deployments, etc.
	// in CI/CD systems https://docs.travis-ci.com/user/github-oauth-scopes/.

	hm := make(map[string]permissions)
	dl := c.Dlogger
	//nolint:errcheck
	remediationMetadata, _ := remediation.New(c)

	for _, r := range results.TokenPermissions {
		var loc *finding.Location
		if r.File != nil {
			loc = &finding.Location{
				Type:      r.File.Type,
				Value:     r.File.Path,
				LineStart: &r.File.Offset,
			}
			if r.File.Snippet != "" {
				loc.Snippet = &r.File.Snippet
			}

			loc.Value = r.File.Path
		}

		text, err := createText(r)
		if err != nil {
			return checker.MinResultScore, err
		}

		msg, err := createLogMsg(r.LocationType)
		if err != nil {
			return checker.InconclusiveResultScore, err
		}
		msg.Finding = msg.Finding.WithMessage(text).WithLocation(loc)
		switch r.Type {
		case checker.PermissionLevelNone, checker.PermissionLevelRead:
			msg.Finding = msg.Finding.WithOutcome(finding.OutcomePositive)
			dl.Info(msg)
		case checker.PermissionLevelUnknown:
			dl.Debug(msg)

		case checker.PermissionLevelUndeclared:
			if r.LocationType == nil {
				return checker.InconclusiveResultScore,
					sce.WithMessage(sce.ErrScorecardInternal, "locationType is nil")
			}

			// We warn only for top-level.
			if *r.LocationType == checker.PermissionLocationTop {
				warnWithRemediation(dl, msg, remediationMetadata, loc)
			} else {
				dl.Debug(msg)
			}

			// Group results by workflow name for score computation.
			if err := updateWorkflowHashMap(hm, r); err != nil {
				return checker.InconclusiveResultScore, err
			}

		case checker.PermissionLevelWrite:
			warnWithRemediation(dl, msg, remediationMetadata, loc)

			// Group results by workflow name for score computation.
			if err := updateWorkflowHashMap(hm, r); err != nil {
				return checker.InconclusiveResultScore, err
			}
		}
	}

	return calculateScore(hm), nil
}

func createLogMsg(loct *checker.PermissionLocation) (*checker.LogMessage, error) {
	ruleName := "GitHubWorkflowPermissionsStepsNoWrite"
	if loct == nil || *loct == checker.PermissionLocationTop {
		ruleName = "GitHubWorkflowPermissionsTopNoWrite"
	}
	f, err := finding.New(rules, ruleName)
	if err != nil {
		return nil,
			sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	}
	return &checker.LogMessage{
		Finding: f,
	}, nil
}

func warnWithRemediation(logger checker.DetailLogger, msg *checker.LogMessage,
	rem *remediation.RemediationMetadata, loc *finding.Location,
) {
	if loc != nil && loc.Value != "" {
		msg.Finding = msg.Finding.WithRemediationMetadata(map[string]string{
			"repo":     rem.Repo,
			"branch":   rem.Branch,
			"workflow": strings.TrimPrefix(loc.Value, ".github/workflows/"),
		})
	}
	logger.Warn(msg)
}

func recordPermissionWrite(hm map[string]permissions, path string,
	locType checker.PermissionLocation, permName *string,
) {
	if _, exists := hm[path]; !exists {
		hm[path] = permissions{
			topLevelWritePermissions: make(map[string]bool),
			jobLevelWritePermissions: make(map[string]bool),
		}
	}

	// Select the hash map to update.
	m := hm[path].jobLevelWritePermissions
	if locType == checker.PermissionLocationTop {
		m = hm[path].topLevelWritePermissions
	}

	// Set the permission name to record.
	name := "all"
	if permName != nil && *permName != "" {
		name = *permName
	}
	m[name] = true
}

func updateWorkflowHashMap(hm map[string]permissions, t checker.TokenPermission) error {
	if t.LocationType == nil {
		return sce.WithMessage(sce.ErrScorecardInternal, "locationType is nil")
	}

	if t.File == nil || t.File.Path == "" {
		return sce.WithMessage(sce.ErrScorecardInternal, "path is not set")
	}

	if t.Type != checker.PermissionLevelWrite &&
		t.Type != checker.PermissionLevelUndeclared {
		return nil
	}

	recordPermissionWrite(hm, t.File.Path, *t.LocationType, t.Name)

	return nil
}

func createText(t checker.TokenPermission) (string, error) {
	// By default, use the message already present.
	if t.Msg != nil {
		return *t.Msg, nil
	}

	// Ensure there's no implementation bug.
	if t.LocationType == nil {
		return "", sce.WithMessage(sce.ErrScorecardInternal, "locationType is nil")
	}

	// Use a different text depending on the type.
	if t.Type == checker.PermissionLevelUndeclared {
		return fmt.Sprintf("no %s permission defined", *t.LocationType), nil
	}

	if t.Value == nil {
		return "", sce.WithMessage(sce.ErrScorecardInternal, "Value fields is nil")
	}

	if t.Name == nil {
		return fmt.Sprintf("%s permissions set to '%v'", *t.LocationType,
			*t.Value), nil
	}

	return fmt.Sprintf("%s '%v' permission set to '%v'", *t.LocationType,
		*t.Name, *t.Value), nil
}

// Calculate the score.
func calculateScore(result map[string]permissions) int {
	// See list https://github.blog/changelog/2021-04-20-github-actions-control-permissions-for-github_token/.
	// Note: there are legitimate reasons to use some of the permissions like checks, deployments, etc.
	// in CI/CD systems https://docs.travis-ci.com/user/github-oauth-scopes/.

	// Start with a perfect score.
	score := float32(checker.MaxResultScore)

	// Retrieve the overall results.
	for _, perms := range result {
		// If no top level permissions are defined, all the permissions
		// are enabled by default. In this case,
		if permissionIsPresentInTopLevel(perms, "all") {
			if permissionIsPresentInRunLevel(perms, "all") {
				// ... give lowest score if no run level permissions are defined either.
				return checker.MinResultScore
			}
			// ... reduce score if run level permissions are defined.
			score -= 0.5
		}

		// status: https://docs.github.com/en/rest/reference/repos#statuses.
		// May allow an attacker to change the result of pre-submit and get a PR merged.
		// Low risk: -0.5.
		if permissionIsPresent(perms, "statuses") {
			score -= 0.5
		}

		// checks.
		// May allow an attacker to edit checks to remove pre-submit and introduce a bug.
		// Low risk: -0.5.
		if permissionIsPresent(perms, "checks") {
			score -= 0.5
		}

		// secEvents.
		// May allow attacker to read vuln reports before patch available.
		// Low risk: -1
		if permissionIsPresent(perms, "security-events") {
			score--
		}

		// deployments: https://docs.github.com/en/rest/reference/repos#deployments.
		// May allow attacker to charge repo owner by triggering VM runs,
		// and tiny chance an attacker can trigger a remote
		// service with code they own if server accepts code/location var unsanitized.
		// Low risk: -1
		if permissionIsPresent(perms, "deployments") {
			score--
		}

		// contents.
		// Allows attacker to commit unreviewed code.
		// High risk: -10
		if permissionIsPresentInTopLevel(perms, "contents") {
			score -= checker.MaxResultScore
		}

		// packages: https://docs.github.com/en/packages/learn-github-packages/about-permissions-for-github-packages.
		// Allows attacker to publish packages.
		// High risk: -10
		if permissionIsPresentInTopLevel(perms, "packages") {
			score -= checker.MaxResultScore
		}

		// actions.
		// May allow an attacker to steal GitHub secrets by approving to run an action that needs approval.
		// High risk: -10
		if permissionIsPresentInTopLevel(perms, "actions") {
			score -= checker.MaxResultScore
		}

		if score < checker.MinResultScore {
			break
		}
	}

	// We're done, calculate the final score.
	if score < checker.MinResultScore {
		return checker.MinResultScore
	}

	return int(score)
}

func permissionIsPresent(perms permissions, name string) bool {
	return permissionIsPresentInTopLevel(perms, name) ||
		permissionIsPresentInRunLevel(perms, name)
}

func permissionIsPresentInTopLevel(perms permissions, name string) bool {
	_, ok := perms.topLevelWritePermissions[name]
	return ok
}

func permissionIsPresentInRunLevel(perms permissions, name string) bool {
	_, ok := perms.jobLevelWritePermissions[name]
	return ok
}
