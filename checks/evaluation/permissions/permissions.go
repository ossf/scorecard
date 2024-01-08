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
var probes embed.FS

type permissions struct {
	topLevelWritePermissions map[string]bool
	jobLevelWritePermissions map[string]bool
}

var (
	stepsNoWriteID = "gitHubWorkflowPermissionsStepsNoWrite"
	topNoWriteID   = "gitHubWorkflowPermissionsTopNoWrite"
)

type permissionLevel int

const (
	// permissionLevelNone is a permission set to `none`.
	permissionLevelNone permissionLevel = iota
	// permissionLevelRead is a permission set to `read`.
	permissionLevelRead
	// permissionLevelUnknown is for other kinds of alerts, mostly to support debug messages.
	// TODO: remove it once we have implemented severity (#1874).
	permissionLevelUnknown
	// permissionLevelUndeclared is an undeclared permission.
	permissionLevelUndeclared
	// permissionLevelWrite is a permission set to `write` for a permission we consider potentially dangerous.
	permissionLevelWrite
)

// permissionLocation represents a declaration type.
type permissionLocationType int

const (
	// permissionLocationNil is in case the permission is nil.
	permissionLocationNil permissionLocationType = iota
	// permissionLocationNotDeclared is for undeclared permission.
	permissionLocationNotDeclared
	// permissionLocationTop is top-level workflow permission.
	permissionLocationTop
	// permissionLocationJob is job-level workflow permission.
	permissionLocationJob
)

// permissionType represents a permission type.
type permissionType int

const (
	// permissionTypeNone represents none permission type.
	permissionTypeNone permissionType = iota
	// permissionTypeNone is the "all" github permission type.
	permissionTypeAll
	// permissionTypeNone is the "statuses" github permission type.
	permissionTypeStatuses
	// permissionTypeNone is the "checks" github permission type.
	permissionTypeChecks
	// permissionTypeNone is the "security-events" github permission type.
	permissionTypeSecurityEvents
	// permissionTypeNone is the "deployments" github permission type.
	permissionTypeDeployments
	// permissionTypeNone is the "packages" github permission type.
	permissionTypePackages
	// permissionTypeNone is the "actions" github permission type.
	permissionTypeActions
)

// TokenPermissions applies the score policy for the Token-Permissions check.
func TokenPermissions(name string, c *checker.CheckRequest, r *checker.TokenPermissionsData) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	if r.NumTokens == 0 {
		return checker.CreateInconclusiveResult(name, "no tokens found")
	}

	// This is a temporary step that should be replaced by probes in ./probes
	findings, err := rawToFindings(r)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "could not convert raw data to findings")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	score, err := applyScorePolicy(findings, c)
	if err != nil {
		return checker.CreateRuntimeErrorResult(name, err)
	}

	if score != checker.MaxResultScore {
		return checker.CreateResultWithScore(name,
			"detected GitHub workflow tokens with excessive permissions", score)
	}

	return checker.CreateMaxScoreResult(name,
		"GitHub workflow tokens follow principle of least privilege")
}

// rawToFindings is a temporary step for converting the raw results
// to findings. This should be replaced by probes in ./probes.
func rawToFindings(results *checker.TokenPermissionsData) ([]finding.Finding, error) {
	var findings []finding.Finding

	for _, r := range results.TokenPermissions {
		var loc *finding.Location
		if r.File != nil {
			loc = &finding.Location{
				Type:      r.File.Type,
				Path:      r.File.Path,
				LineStart: newUint(r.File.Offset),
			}
			if r.File.Snippet != "" {
				loc.Snippet = newStr(r.File.Snippet)
			}
		}
		text, err := createText(r)
		if err != nil {
			return nil, err
		}

		f, err := createFinding(r.LocationType, text, loc)
		if err != nil {
			return nil, err
		}

		switch r.Type {
		case checker.PermissionLevelNone:
			f = f.WithOutcome(finding.OutcomePositive)
			f = f.WithValues(map[string]int{
				"PermissionLevel": int(permissionLevelNone),
			})
		case checker.PermissionLevelRead:
			f = f.WithOutcome(finding.OutcomePositive)
			f = f.WithValues(map[string]int{
				"PermissionLevel": int(permissionLevelRead),
			})

		case checker.PermissionLevelUnknown:
			f = f.WithValues(map[string]int{
				"PermissionLevel": int(permissionLevelUnknown),
			}).WithOutcome(finding.OutcomeError)
		case checker.PermissionLevelUndeclared:
			var locationType permissionLocationType
			//nolint:gocritic
			if r.LocationType == nil {
				locationType = permissionLocationNil
			} else if *r.LocationType == checker.PermissionLocationTop {
				locationType = permissionLocationTop
			} else {
				locationType = permissionLocationNotDeclared
			}
			permType := permTypeToEnum(r.Name)
			f = f.WithValues(map[string]int{
				"PermissionLevel": int(permissionLevelUndeclared),
				"LocationType":    int(locationType),
				"PermissionType":  int(permType),
			})
		case checker.PermissionLevelWrite:
			var locationType permissionLocationType
			switch *r.LocationType {
			case checker.PermissionLocationTop:
				locationType = permissionLocationTop
			case checker.PermissionLocationJob:
				locationType = permissionLocationJob
			default:
				locationType = permissionLocationNotDeclared
			}
			permType := permTypeToEnum(r.Name)
			f = f.WithValues(map[string]int{
				"PermissionLevel": int(permissionLevelWrite),
				"LocationType":    int(locationType),
				"PermissionType":  int(permType),
			})
			f = f.WithOutcome(finding.OutcomeNegative)
		}
		findings = append(findings, *f)
	}
	return findings, nil
}

func permTypeToEnum(tokenName *string) permissionType {
	if tokenName == nil {
		return permissionTypeNone
	}
	switch *tokenName {
	//nolint:goconst
	case "all":
		return permissionTypeAll
	case "statuses":
		return permissionTypeStatuses
	case "checks":
		return permissionTypeChecks
	case "security-events":
		return permissionTypeSecurityEvents
	case "deployments":
		return permissionTypeDeployments
	case "contents":
		return permissionTypePackages
	case "actions":
		return permissionTypeActions
	default:
		return permissionTypeNone
	}
}

func permTypeToName(permType int) *string {
	var permName string
	switch permissionType(permType) {
	case permissionTypeAll:
		permName = "all"
	case permissionTypeStatuses:
		permName = "statuses"
	case permissionTypeChecks:
		permName = "checks"
	case permissionTypeSecurityEvents:
		permName = "security-events"
	case permissionTypeDeployments:
		permName = "deployments"
	case permissionTypePackages:
		permName = "contents"
	case permissionTypeActions:
		permName = "actions"
	default:
		permName = ""
	}
	return &permName
}

func createFinding(loct *checker.PermissionLocation, text string, loc *finding.Location) (*finding.Finding, error) {
	probe := stepsNoWriteID
	if loct == nil || *loct == checker.PermissionLocationTop {
		probe = topNoWriteID
	}
	content, err := probes.ReadFile(probe + ".yml")
	if err != nil {
		return nil, fmt.Errorf("%w", err) // wrap the error to avoid linter issues
	}
	f, err := finding.FromBytes(content, probe)
	if err != nil {
		return nil,
			sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	}
	f = f.WithMessage(text)
	if loc != nil {
		f = f.WithLocation(loc)
	}
	return f, nil
}

// avoid memory aliasing by returning a new copy.
func newUint(u uint) *uint {
	return &u
}

// avoid memory aliasing by returning a new copy.
func newStr(s string) *string {
	return &s
}

func applyScorePolicy(findings []finding.Finding, c *checker.CheckRequest) (int, error) {
	// See list https://github.blog/changelog/2021-04-20-github-actions-control-permissions-for-github_token/.
	// Note: there are legitimate reasons to use some of the permissions like checks, deployments, etc.
	// in CI/CD systems https://docs.travis-ci.com/user/github-oauth-scopes/.

	hm := make(map[string]permissions)
	dl := c.Dlogger
	//nolint:errcheck
	remediationMetadata, _ := remediation.New(c)
	negativeProbeResults := map[string]bool{
		stepsNoWriteID: false,
		topNoWriteID:   false,
	}

	for i := range findings {
		f := &findings[i]
		pLevel := permissionLevel(f.Values["PermissionLevel"])
		switch pLevel {
		case permissionLevelNone, permissionLevelRead:
			dl.Info(&checker.LogMessage{
				Finding: f,
			})
		case permissionLevelUnknown:
			dl.Debug(&checker.LogMessage{
				Finding: f,
			})

		case permissionLevelUndeclared:
			switch permissionLocationType(f.Values["LocationType"]) {
			case permissionLocationNil:
				return checker.InconclusiveResultScore,
					sce.WithMessage(sce.ErrScorecardInternal, "locationType is nil")
			case permissionLocationTop:
				warnWithRemediation(dl, remediationMetadata, f, negativeProbeResults)
			default:
				// We warn only for top-level.
				dl.Debug(&checker.LogMessage{
					Finding: f,
				})
			}

			// Group results by workflow name for score computation.
			if err := updateWorkflowHashMap(hm, f); err != nil {
				return checker.InconclusiveResultScore, err
			}

		case permissionLevelWrite:
			warnWithRemediation(dl, remediationMetadata, f, negativeProbeResults)

			// Group results by workflow name for score computation.
			if err := updateWorkflowHashMap(hm, f); err != nil {
				return checker.InconclusiveResultScore, err
			}
		}
	}

	if err := reportDefaultFindings(findings, c.Dlogger, negativeProbeResults); err != nil {
		return checker.InconclusiveResultScore, err
	}
	return calculateScore(hm), nil
}

func reportDefaultFindings(results []finding.Finding,
	dl checker.DetailLogger, negativeProbeResults map[string]bool,
) error {
	// Workflow files found, report positive findings if no
	// negative findings were found.
	// NOTE: we don't consider probe `topNoWriteID`
	// because positive results are already reported.
	found := negativeProbeResults[stepsNoWriteID]
	if !found {
		text := fmt.Sprintf("no %s write permissions found", checker.PermissionLocationJob)
		if err := reportFinding(stepsNoWriteID,
			text, finding.OutcomePositive, dl); err != nil {
			return err
		}
	}

	return nil
}

func reportFinding(probe, text string, o finding.Outcome, dl checker.DetailLogger) error {
	content, err := probes.ReadFile(probe + ".yml")
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	f, err := finding.FromBytes(content, probe)
	if err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	}
	f = f.WithMessage(text).WithOutcome(o)
	dl.Info(&checker.LogMessage{
		Finding: f,
	})
	return nil
}

func warnWithRemediation(logger checker.DetailLogger,
	rem *remediation.RemediationMetadata,
	f *finding.Finding,
	negativeProbeResults map[string]bool,
) {
	if f.Location != nil && f.Location.Path != "" {
		f = f.WithRemediationMetadata(map[string]string{
			"repo":     rem.Repo,
			"branch":   rem.Branch,
			"workflow": strings.TrimPrefix(f.Location.Path, ".github/workflows/"),
		})
	}
	logger.Warn(&checker.LogMessage{
		Finding: f,
	})

	// Record that we found a negative result.
	negativeProbeResults[f.Probe] = true
}

func recordPermissionWrite(hm map[string]permissions, path string,
	locType permissionLocationType, permType int,
) {
	if _, exists := hm[path]; !exists {
		hm[path] = permissions{
			topLevelWritePermissions: make(map[string]bool),
			jobLevelWritePermissions: make(map[string]bool),
		}
	}

	// Select the hash map to update.
	m := hm[path].jobLevelWritePermissions
	if locType == permissionLocationTop {
		m = hm[path].topLevelWritePermissions
	}

	// Set the permission name to record.
	permName := permTypeToName(permType)
	name := "all"
	if permName != nil && *permName != "" {
		name = *permName
	}
	m[name] = true
}

func updateWorkflowHashMap(hm map[string]permissions, f *finding.Finding) error {
	if _, ok := f.Values["LocationType"]; !ok {
		return sce.WithMessage(sce.ErrScorecardInternal, "locationType is nil")
	}

	if f.Location == nil || f.Location.Path == "" {
		return sce.WithMessage(sce.ErrScorecardInternal, "path is not set")
	}

	if permissionLevel(f.Values["PermissionLevel"]) != permissionLevelWrite &&
		permissionLevel(f.Values["PermissionLevel"]) != permissionLevelUndeclared {
		return nil
	}
	plt := permissionLocationType(f.Values["LocationType"])
	recordPermissionWrite(hm, f.Location.Path, plt, f.Values["PermissionType"])

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
		if permissionIsPresentInTopLevel(perms, "statuses") {
			score -= 0.5
		}

		// checks.
		// May allow an attacker to edit checks to remove pre-submit and introduce a bug.
		// Low risk: -0.5.
		if permissionIsPresentInTopLevel(perms, "checks") {
			score -= 0.5
		}

		// secEvents.
		// May allow attacker to read vuln reports before patch available.
		// Low risk: -1
		if permissionIsPresentInTopLevel(perms, "security-events") {
			score--
		}

		// deployments: https://docs.github.com/en/rest/reference/repos#deployments.
		// May allow attacker to charge repo owner by triggering VM runs,
		// and tiny chance an attacker can trigger a remote
		// service with code they own if server accepts code/location var unsanitized.
		// Low risk: -1
		if permissionIsPresentInTopLevel(perms, "deployments") {
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

func permissionIsPresentInTopLevel(perms permissions, name string) bool {
	_, ok := perms.topLevelWritePermissions[name]
	return ok
}

func permissionIsPresentInRunLevel(perms permissions, name string) bool {
	_, ok := perms.jobLevelWritePermissions[name]
	return ok
}
