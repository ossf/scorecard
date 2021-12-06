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

package checks

import (
	"fmt"
	"strings"

	"github.com/rhysd/actionlint"

	"github.com/ossf/scorecard/v3/checker"
	"github.com/ossf/scorecard/v3/checks/fileparser"
	sce "github.com/ossf/scorecard/v3/errors"
)

// CheckTokenPermissions is the exported name for Token-Permissions check.
const CheckTokenPermissions = "Token-Permissions"

//nolint:gochecknoinits
func init() {
	registerCheck(CheckTokenPermissions, TokenPermissions)
}

// Holds stateful data to pass thru callbacks.
// Each field correpsonds to a GitHub permission type, and
// will hold true if declared non-write, false otherwise.
type permissionCbData struct {
	topLevelWritePermissions map[string]bool
	runLevelWritePermissions map[string]bool
}

// TokenPermissions runs Token-Permissions check.
func TokenPermissions(c *checker.CheckRequest) checker.CheckResult {
	// data is shared across all GitHub workflows.
	data := permissionCbData{
		topLevelWritePermissions: make(map[string]bool),
		runLevelWritePermissions: make(map[string]bool),
	}
	err := fileparser.CheckFilesContent(".github/workflows/*", false,
		c, validateGitHubActionTokenPermissions, &data)
	return createResultForLeastPrivilegeTokens(data, err)
}

func validatePermission(permissionKey string, permissionValue *actionlint.PermissionScope, path string,
	dl checker.DetailLogger, pPermissions map[string]bool,
	ignoredPermissions map[string]bool) error {
	if permissionValue.Value == nil {
		return sce.WithMessage(sce.ErrScorecardInternal, errInvalidGitHubWorkflow.Error())
	}
	val := permissionValue.Value.Value
	lineNumber := checker.OffsetDefault
	if permissionValue.Value.Pos != nil {
		lineNumber = permissionValue.Value.Pos.Line
	}
	if strings.EqualFold(val, "write") {
		if isPermissionOfInterest(permissionKey, ignoredPermissions) {
			dl.Warn3(&checker.LogMessage{
				Path:   path,
				Type:   checker.FileTypeSource,
				Offset: lineNumber,
				Text:   fmt.Sprintf("'%v' permission set to '%v'", permissionKey, val),
				// TODO: set Snippet.
			})
			recordPermissionWrite(permissionKey, pPermissions)
		} else {
			// Only log for debugging, otherwise
			// it may confuse users.
			dl.Debug3(&checker.LogMessage{
				Path:   path,
				Type:   checker.FileTypeSource,
				Offset: lineNumber,
				Text:   fmt.Sprintf("'%v' permission set to '%v'", permissionKey, val),
				// TODO: set Snippet.
			})
		}
		return nil
	}

	dl.Info3(&checker.LogMessage{
		Path:   path,
		Type:   checker.FileTypeSource,
		Offset: lineNumber,
		Text:   fmt.Sprintf("'%v' permission set to '%v'", permissionKey, val),
		// TODO: set Snippet.
	})
	return nil
}

func validateMapPermissions(scopes map[string]*actionlint.PermissionScope, path string,
	dl checker.DetailLogger, pPermissions map[string]bool,
	ignoredPermissions map[string]bool) error {
	for key, v := range scopes {
		if err := validatePermission(key, v, path, dl, pPermissions, ignoredPermissions); err != nil {
			return err
		}
	}
	return nil
}

func recordPermissionWrite(name string, pPermissions map[string]bool) {
	pPermissions[name] = true
}

func recordAllPermissionsWrite(pPermissions map[string]bool) {
	// Special case: `all` does not correspond
	// to a GitHub permission.
	pPermissions["all"] = true
}

func validatePermissions(permissions *actionlint.Permissions, path string,
	dl checker.DetailLogger, pPermissions map[string]bool,
	ignoredPermissions map[string]bool) error {
	allIsSet := permissions != nil && permissions.All != nil && permissions.All.Value != ""
	scopeIsSet := permissions != nil && len(permissions.Scopes) > 0
	if permissions == nil || (!allIsSet && !scopeIsSet) {
		dl.Info3(&checker.LogMessage{
			Path:   path,
			Type:   checker.FileTypeSource,
			Offset: checker.OffsetDefault,
			Text:   "permissions set to 'none'",
		})
	}
	if allIsSet {
		val := permissions.All.Value
		lineNumber := checker.OffsetDefault
		if permissions.All.Pos != nil {
			lineNumber = permissions.All.Pos.Line
		}
		if !strings.EqualFold(val, "read-all") && val != "" {
			dl.Warn3(&checker.LogMessage{
				Path:   path,
				Type:   checker.FileTypeSource,
				Offset: lineNumber,
				Text:   fmt.Sprintf("permissions set to '%v'", val),
				// TODO: set Snippet.
			})
			recordAllPermissionsWrite(pPermissions)
			return nil
		}

		dl.Info3(&checker.LogMessage{
			Path:   path,
			Type:   checker.FileTypeSource,
			Offset: lineNumber,
			Text:   fmt.Sprintf("permissions set to '%v'", val),
			// TODO: set Snippet.
		})
	} else /* scopeIsSet == true */ if err := validateMapPermissions(permissions.Scopes, path, dl, pPermissions,
		ignoredPermissions); err != nil {
		return err
	}
	return nil
}

func validateTopLevelPermissions(workflow *actionlint.Workflow, path string,
	dl checker.DetailLogger, pdata *permissionCbData) error {
	// Check if permissions are set explicitly.
	if workflow.Permissions == nil {
		dl.Warn3(&checker.LogMessage{
			Path:   path,
			Type:   checker.FileTypeSource,
			Offset: checker.OffsetDefault,
			Text:   "no permission defined",
		})
		recordAllPermissionsWrite(pdata.topLevelWritePermissions)
		return nil
	}

	return validatePermissions(workflow.Permissions, path, dl,
		pdata.topLevelWritePermissions, map[string]bool{})
}

func validateRunLevelPermissions(workflow *actionlint.Workflow, path string,
	dl checker.DetailLogger, pdata *permissionCbData,
	ignoredPermissions map[string]bool) error {
	for _, job := range workflow.Jobs {
		// Run-level permissions may be left undefined.
		// For most workflows, no write permissions are needed,
		// so only top-level read-only permissions need to be declared.
		if job.Permissions == nil {
			lineNumber := checker.OffsetDefault
			if job.Pos != nil {
				lineNumber = job.Pos.Line
			}
			dl.Debug3(&checker.LogMessage{
				Path:   path,
				Type:   checker.FileTypeSource,
				Offset: lineNumber,
				Text:   "no permission defined",
			})
			continue
		}
		err := validatePermissions(job.Permissions, path, dl, pdata.runLevelWritePermissions, ignoredPermissions)
		if err != nil {
			return err
		}
	}
	return nil
}

func isPermissionOfInterest(name string, ignoredPermissions map[string]bool) bool {
	permissions := []string{
		"statuses", "checks", "security-events",
		"deployments", "contents", "packages", "actions",
	}
	for _, p := range permissions {
		_, present := ignoredPermissions[p]
		if strings.EqualFold(name, p) && !present {
			return true
		}
	}
	return false
}

func permissionIsPresent(result permissionCbData, name string) bool {
	_, ok1 := result.topLevelWritePermissions[name]
	_, ok2 := result.runLevelWritePermissions[name]
	return ok1 || ok2
}

// Calculate the score.
func calculateScore(result permissionCbData) int {
	// See list https://github.blog/changelog/2021-04-20-github-actions-control-permissions-for-github_token/.
	// Note: there are legitimate reasons to use some of the permissions like checks, deployments, etc.
	// in CI/CD systems https://docs.travis-ci.com/user/github-oauth-scopes/.

	if permissionIsPresent(result, "all") {
		return checker.MinResultScore
	}

	// Start with a perfect score.
	score := float32(checker.MaxResultScore)

	// status: https://docs.github.com/en/rest/reference/repos#statuses.
	// May allow an attacker to change the result of pre-submit and get a PR merged.
	// Low risk: -0.5.
	if permissionIsPresent(result, "statuses") {
		score -= 0.5
	}

	// checks.
	// May allow an attacker to edit checks to remove pre-submit and introduce a bug.
	// Low risk: -0.5.
	if permissionIsPresent(result, "checks") {
		score -= 0.5
	}

	// secEvents.
	// May allow attacker to read vuln reports before patch available.
	// Low risk: -1
	if permissionIsPresent(result, "security-events") {
		score--
	}

	// deployments: https://docs.github.com/en/rest/reference/repos#deployments.
	// May allow attacker to charge repo owner by triggering VM runs,
	// and tiny chance an attacker can trigger a remote
	// service with code they own if server accepts code/location var unsanitized.
	// Low risk: -1
	if permissionIsPresent(result, "deployments") {
		score--
	}

	// contents.
	// Allows attacker to commit unreviewed code.
	// High risk: -10
	if permissionIsPresent(result, "contents") {
		score -= checker.MaxResultScore
	}

	// packages: https://docs.github.com/en/packages/learn-github-packages/about-permissions-for-github-packages.
	// Allows attacker to publish packages.
	// High risk: -10
	if permissionIsPresent(result, "packages") {
		score -= checker.MaxResultScore
	}

	// actions.
	// May allow an attacker to steal GitHub secrets by adding a malicious workflow/action.
	// High risk: -10
	if permissionIsPresent(result, "actions") {
		score -= checker.MaxResultScore
	}

	// We're done, calculate the final score.
	if score < checker.MinResultScore {
		return checker.MinResultScore
	}

	return int(score)
}

// Create the result.
func createResultForLeastPrivilegeTokens(result permissionCbData, err error) checker.CheckResult {
	if err != nil {
		return checker.CreateRuntimeErrorResult(CheckTokenPermissions, err)
	}

	score := calculateScore(result)

	if score != checker.MaxResultScore {
		return checker.CreateResultWithScore(CheckTokenPermissions,
			"non read-only tokens detected in GitHub workflows", score)
	}

	return checker.CreateMaxScoreResult(CheckTokenPermissions,
		"tokens are read-only in GitHub workflows")
}

func testValidateGitHubActionTokenPermissions(pathfn string,
	content []byte, dl checker.DetailLogger) checker.CheckResult {
	data := permissionCbData{
		topLevelWritePermissions: make(map[string]bool),
		runLevelWritePermissions: make(map[string]bool),
	}
	_, err := validateGitHubActionTokenPermissions(pathfn, content, dl, &data)
	return createResultForLeastPrivilegeTokens(data, err)
}

// Check file content.
func validateGitHubActionTokenPermissions(path string, content []byte,
	dl checker.DetailLogger, data fileparser.FileCbData) (bool, error) {
	if !fileparser.IsWorkflowFile(path) {
		return true, nil
	}
	// Verify the type of the data.
	pdata, ok := data.(*permissionCbData)
	if !ok {
		// This never happens.
		panic("invalid type")
	}

	if !fileparser.CheckFileContainsCommands(content, "#") {
		return true, nil
	}

	workflow, errs := actionlint.Parse(content)
	if len(errs) > 0 && workflow == nil {
		return false, fileparser.FormatActionlintError(errs)
	}

	// 1. Top-level permission definitions.
	//nolint
	// https://docs.github.com/en/actions/reference/authentication-in-a-workflow#example-1-passing-the-github_token-as-an-input,
	// https://github.blog/changelog/2021-04-20-github-actions-control-permissions-for-github_token/,
	// https://docs.github.com/en/actions/reference/authentication-in-a-workflow#modifying-the-permissions-for-the-github_token.
	if err := validateTopLevelPermissions(workflow, path, dl, pdata); err != nil {
		return false, err
	}

	// 2. Run-level permission definitions,
	// see https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idpermissions.
	ignoredPermissions := createIgnoredPermissions(string(content), path, dl)
	if err := validateRunLevelPermissions(workflow, path, dl, pdata, ignoredPermissions); err != nil {
		return false, err
	}

	// TODO(laurent): 2. Identify github actions that require write and add checks.

	// TODO(laurent): 3. Read a few runs and ensures they have the same permissions.

	return true, nil
}

func createIgnoredPermissions(s, fp string, dl checker.DetailLogger) map[string]bool {
	ignoredPermissions := make(map[string]bool)
	if requiresPackagesPermissions(s, fp, dl) {
		ignoredPermissions["packages"] = true
	}
	if requiresContentsPermissions(s, fp, dl) {
		ignoredPermissions["contents"] = true
	}
	if isSARIFUploadWorkflow(s, fp, dl) {
		ignoredPermissions["security-events"] = true
	}

	return ignoredPermissions
}

// Scanning tool run externally and SARIF file uploaded.
func isSARIFUploadWorkflow(s, fp string, dl checker.DetailLogger) bool {
	//nolint
	// CodeQl analysis workflow automatically sends sarif file to GitHub.
	// https://docs.github.com/en/code-security/secure-coding/integrating-with-code-scanning/uploading-a-sarif-file-to-github#about-sarif-file-uploads-for-code-scanning.
	// `The CodeQL action uploads the SARIF file automatically when it completes analysis`.
	if isCodeQlAnalysisWorkflow(s, fp, dl) {
		return true
	}

	//nolint
	// Third-party scanning tools use the SARIF-upload action from code-ql.
	// https://docs.github.com/en/code-security/secure-coding/integrating-with-code-scanning/uploading-a-sarif-file-to-github#uploading-a-code-scanning-analysis-with-github-actions
	// We only support CodeQl today.
	if isSARIFUploadAction(s, fp, dl) {
		return true
	}

	// TODO: some third party tools may upload directly thru their actions.
	// Very unlikely.
	// See https://github.com/marketplace for tools.

	return false
}

// CodeQl run externally and SARIF file uploaded.
func isSARIFUploadAction(s, fp string, dl checker.DetailLogger) bool {
	if strings.Contains(s, "github/codeql-action/upload-sarif@") {
		dl.Debug3(&checker.LogMessage{
			Path: fp,
			Type: checker.FileTypeSource,
			// TODO: set line.
			Offset: 1,
			Text:   "codeql SARIF upload workflow detected",
			// TODO: set Snippet.
		})
		return true
	}
	dl.Debug3(&checker.LogMessage{
		Path:   fp,
		Type:   checker.FileTypeSource,
		Offset: 1,
		Text:   "not a codeql upload SARIF workflow",
	})
	return false
}

//nolint
// CodeQl run within GitHub worklow automatically bubbled up to
// security events, see
// https://docs.github.com/en/code-security/secure-coding/automatically-scanning-your-code-for-vulnerabilities-and-errors/configuring-code-scanning.
func isCodeQlAnalysisWorkflow(s, fp string, dl checker.DetailLogger) bool {
	if strings.Contains(s, "github/codeql-action/analyze@") {
		dl.Debug3(&checker.LogMessage{
			Path: fp,
			Type: checker.FileTypeSource,
			// TODO: set line.
			Offset: 1,
			Text:   "codeql workflow detected",
			// TODO: set Snippet.
		})
		return true
	}
	dl.Debug3(&checker.LogMessage{
		Path:   fp,
		Type:   checker.FileTypeSource,
		Offset: 1,
		Text:   "not a codeql workflow",
	})
	return false
}

// A packaging workflow using GitHub's supported packages:
// https://docs.github.com/en/packages.
func requiresPackagesPermissions(s, fp string, dl checker.DetailLogger) bool {
	// TODO: add support for GitHub registries.
	// Example: https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-npm-registry.
	// This feature requires parsing actions properly.
	// For now, we just re-use the Packaging check to verify that the
	// workflow is a packaging workflow.
	return isPackagingWorkflow(s, fp, dl)
}

// Note: this needs to be improved.
// Currently we don't differentiate between publishing on GitHub vs
// pubishing on registries. In terms of risk, both are similar, as
// an attacker would gain the ability to push a package.
func requiresContentsPermissions(s, fp string, dl checker.DetailLogger) bool {
	return requiresPackagesPermissions(s, fp, dl)
}
