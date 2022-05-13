// Copyright 2021 Security Scorecard Authors
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
)

// TokenPermissions applies the score policy for the Token-Permissions check.
func TokenPermissions(name string, dl checker.DetailLogger, r *checker.TokenPermissionsData) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	score, details := calculateScore(r)

	if score != checker.MaxResultScore {
		return checker.CreateResultWithScore(name,
			"non read-only tokens detected in GitHub workflows", score)
	}

	return checker.CreateMaxScoreResult(name,
		"tokens are read-only in GitHub workflows")
}

func calculateScore(results *checker.TokenPermissionsData, dl checker.DetailLogger) int {
	// See list https://github.blog/changelog/2021-04-20-github-actions-control-permissions-for-github_token/.
	// Note: there are legitimate reasons to use some of the permissions like checks, deployments, etc.
	// in CI/CD systems https://docs.travis-ci.com/user/github-oauth-scopes/.

	// Start with a perfect score.
	score := float32(checker.MaxResultScore)

	for _, r := range results.TokenPermissions {
		msg := checker.LogMessage{Remediation: r.Remediation}
		if r.File != nil {
			msg.Path = r.File.Path
			msg.Offset = r.File.Offset
			msg.Type = r.File.Type
			msg.Snippet = r.File.Snippet
		}
		if r.Log.Msg != "" {
			msg.Text = r.Log.Msg
		}

		switch r.Log.Level {
		case checker.LogLevelDebug:
			dl.Debug(&msg)

		case checker.LogLevelInfo:
			dl.Info(&msg)

		case checker.LogLevelWarn:

			// TODO: construct a hash map indexed by workflow file.
		}
	}

	// TODO: use the hash map to compute the score.
	return 10
}

/*
// Calculate the score.
func calculateScore(result permissionCbData) int {
	// See list https://github.blog/changelog/2021-04-20-github-actions-control-permissions-for-github_token/.
	// Note: there are legitimate reasons to use some of the permissions like checks, deployments, etc.
	// in CI/CD systems https://docs.travis-ci.com/user/github-oauth-scopes/.

	// Start with a perfect score.
	score := float32(checker.MaxResultScore)

	// Retrieve the overall results.
	for _, perms := range result.workflows {
		// If no top level permissions are defined, all the permissions
		// are enabled by default, hence permissionAll. In this case,
		if permissionIsPresentInTopLevel(perms, permissionAll) {
			if permissionIsPresentInRunLevel(perms, permissionAll) {
				// ... give lowest score if no run level permissions are defined either.
				return checker.MinResultScore
			}
			// ... reduce score if run level permissions are defined.
			score -= 0.5
		}

		// status: https://docs.github.com/en/rest/reference/repos#statuses.
		// May allow an attacker to change the result of pre-submit and get a PR merged.
		// Low risk: -0.5.
		if permissionIsPresent(perms, permissionStatuses) {
			score -= 0.5
		}

		// checks.
		// May allow an attacker to edit checks to remove pre-submit and introduce a bug.
		// Low risk: -0.5.
		if permissionIsPresent(perms, permissionChecks) {
			score -= 0.5
		}

		// secEvents.
		// May allow attacker to read vuln reports before patch available.
		// Low risk: -1
		if permissionIsPresent(perms, permissionSecurityEvents) {
			score--
		}

		// deployments: https://docs.github.com/en/rest/reference/repos#deployments.
		// May allow attacker to charge repo owner by triggering VM runs,
		// and tiny chance an attacker can trigger a remote
		// service with code they own if server accepts code/location var unsanitized.
		// Low risk: -1
		if permissionIsPresent(perms, permissionDeployments) {
			score--
		}

		// contents.
		// Allows attacker to commit unreviewed code.
		// High risk: -10
		if permissionIsPresent(perms, permissionContents) {
			score -= checker.MaxResultScore
		}

		// packages: https://docs.github.com/en/packages/learn-github-packages/about-permissions-for-github-packages.
		// Allows attacker to publish packages.
		// High risk: -10
		if permissionIsPresent(perms, permissionPackages) {
			score -= checker.MaxResultScore
		}

		// actions.
		// May allow an attacker to steal GitHub secrets by approving to run an action that needs approval.
		// High risk: -10
		if permissionIsPresent(perms, permissionActions) {
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

func permissionIsPresent(perms permissions, name permission) bool {
	return permissionIsPresentInTopLevel(perms, name) ||
		permissionIsPresentInRunLevel(perms, name)
}

func permissionIsPresentInTopLevel(perms permissions, name permission) bool {
	_, ok := perms.topLevelWritePermissions[name]
	return ok
}

func permissionIsPresentInRunLevel(perms permissions, name permission) bool {
	_, ok := perms.jobLevelWritePermissions[name]
	return ok
}
*/
