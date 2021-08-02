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

	"gopkg.in/yaml.v2"

	"github.com/ossf/scorecard/v2/checker"
	sce "github.com/ossf/scorecard/v2/errors"
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
	data := permissionCbData{
		topLevelWritePermissions: make(map[string]bool),
		runLevelWritePermissions: make(map[string]bool),
	}
	err := CheckFilesContent(".github/workflows/*", false,
		c, validateGitHubActionTokenPermissions, &data)
	return createResultForLeastPrivilegeTokens(data, err)
}

func validatePermission(key string, value interface{}, path string,
	dl checker.DetailLogger, pPermissions map[string]bool) error {
	val, ok := value.(string)
	if !ok {
		//nolint
		return sce.Create(sce.ErrScorecardInternal, errInvalidGitHubWorkflow.Error())
	}

	if strings.EqualFold(val, "write") {
		if isPermissionOfInterest(key) {
			dl.Warn("'%v' permission set to '%v' in %v", key, val, path)
			recordPermissionWrite(key, pPermissions)
		} else {
			// Only log for debugging, otherwise
			// it may confuse users.
			dl.Debug("'%v' permission set to '%v' in %v", key, val, path)
		}
		return nil
	}

	dl.Info("'%v' permission set to '%v' in %v", key, val, path)
	return nil
}

func validateMapPermissions(values map[interface{}]interface{}, path string,
	dl checker.DetailLogger, pPermissions map[string]bool) error {
	// Iterate over the permission, verify keys and values are strings.
	for k, v := range values {
		key, ok := k.(string)
		if !ok {
			//nolint
			return sce.Create(sce.ErrScorecardInternal, errInvalidGitHubWorkflow.Error())
		}

		if err := validatePermission(key, v, path, dl, pPermissions); err != nil {
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

func validatePermissions(config map[interface{}]interface{}, path string,
	dl checker.DetailLogger, pPermissions map[string]bool) error {
	var permissions interface{}

	// Check if permissions are set explicitly.
	permissions, ok := config["permissions"]
	if !ok {
		dl.Warn("no permission defined in %v", path)
		recordAllPermissionsWrite(pPermissions)
		return nil
	}

	// Check the type of our values.
	switch val := permissions.(type) {
	// Empty string is nil type.
	// It defaults to 'none'
	case nil:
		dl.Info("permissions set to 'none' in %v", path)
	// String type.
	case string:
		if !strings.EqualFold(val, "read-all") && val != "" {
			dl.Warn("permissions set to '%v' in %v", val, path)
			recordAllPermissionsWrite(pPermissions)
			return nil
		}
		dl.Info("permission set to '%v' in %v", val, path)

	// Map type.
	case map[interface{}]interface{}:
		if err := validateMapPermissions(val, path, dl, pPermissions); err != nil {
			return err
		}

	// Invalid type.
	default:
		//nolint
		return sce.Create(sce.ErrScorecardInternal, errInvalidGitHubWorkflow.Error())
	}
	return nil
}

func validateTopLevelPermissions(config map[interface{}]interface{}, path string,
	dl checker.DetailLogger, pdata *permissionCbData) error {
	return validatePermissions(config, path, dl, pdata.topLevelWritePermissions)
}

func validateRunLevelPermissions(config map[interface{}]interface{}, path string,
	dl checker.DetailLogger, pdata *permissionCbData) error {
	var jobs interface{}

	// Check if permissions are set explicitly.
	jobs, ok := config["jobs"]
	if !ok {
		return nil
	}
	fmt.Printf("jobs:%v\n", jobs)
	fmt.Printf("type:%T\n", jobs)
	mjobs, ok := jobs.(map[interface{}]interface{})
	if !ok {
		//nolint:wrapcheck
		return sce.Create(sce.ErrScorecardInternal, errInvalidGitHubWorkflow.Error())
	}
	fmt.Printf("jobsList:%v\n", mjobs)
	fmt.Printf("list:\n")
	for key, value := range mjobs {
		fmt.Printf("%v: %v\n", key, value)
	}
	return validatePermissions(config, path, dl, pdata.runLevelWritePermissions)
}

func isPermissionOfInterest(name string) bool {
	return strings.EqualFold(name, "statuses") ||
		strings.EqualFold(name, "checks") ||
		strings.EqualFold(name, "security-events") ||
		strings.EqualFold(name, "deployments") ||
		strings.EqualFold(name, "contents") ||
		strings.EqualFold(name, "packages") ||
		strings.EqualFold(name, "options")
}

// Calculate the score.
func calculateScore(result permissionCbData) int {
	// See list https://github.blog/changelog/2021-04-20-github-actions-control-permissions-for-github_token/.
	// Note: there are legitimate reasons to use some of the permissions like checks, deployments, etc.
	// in CI/CD systems https://docs.travis-ci.com/user/github-oauth-scopes/.
	if _, ok := result.topLevelWritePermissions["all"]; ok {
		return checker.MinResultScore
	}

	score := float32(checker.MaxResultScore)
	// status: https://docs.github.com/en/rest/reference/repos#statuses.
	// May allow an attacker to change the result of pre-submit and get a PR merged.
	// Low risk: -0.5.
	if _, ok := result.topLevelWritePermissions["statuses"]; ok {
		score -= 0.5
	}

	// checks.
	// May allow an attacker to edit checks to remove pre-submit and introduce a bug.
	// Low risk: -0.5.
	if _, ok := result.topLevelWritePermissions["checks"]; ok {
		score -= 0.5
	}

	// secEvents.
	// May allow attacker to read vuln reports before patch available.
	// Low risk: -1
	if _, ok := result.topLevelWritePermissions["security-events"]; ok {
		score--
	}

	// deployments: https://docs.github.com/en/rest/reference/repos#deployments.
	// May allow attacker to charge repo owner by triggering VM runs,
	// and tiny chance an attacker can trigger a remote
	// service with code they own if server accepts code/location var unsanitized.
	// Low risk: -1
	if _, ok := result.topLevelWritePermissions["deployments"]; ok {
		score--
	}

	// contents.
	// Allows attacker to commit unreviewed code.
	// High risk: -10
	if _, ok := result.topLevelWritePermissions["contents"]; ok {
		score -= checker.MaxResultScore
	}

	// packages: https://docs.github.com/en/packages/learn-github-packages/about-permissions-for-github-packages.
	// Allows attacker to publish packages.
	// High risk: -10
	if _, ok := result.topLevelWritePermissions["packages"]; ok {
		score -= checker.MaxResultScore
	}

	// actions.
	// May allow an attacker to steal GitHub secrets by adding a malicious workflow/action.
	// High risk: -10
	if _, ok := result.topLevelWritePermissions["actions"]; ok {
		score -= checker.MaxResultScore
	}

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
	dl checker.DetailLogger, data FileCbData) (bool, error) {
	// Verify the type of the data.
	pdata, ok := data.(*permissionCbData)
	if !ok {
		// This never happens.
		panic("invalid type")
	}

	if !CheckFileContainsCommands(content, "#") {
		return true, nil
	}

	var workflow map[interface{}]interface{}
	err := yaml.Unmarshal(content, &workflow)
	if err != nil {
		//nolint
		return false,
			sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("yaml.Unmarshal: %v", err))
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
	if err := validateRunLevelPermissions(workflow, path, dl, pdata); err != nil {
		return false, err
	}

	// TODO(laurent): 2. Identify github actions that require write and add checks.

	// TODO(laurent): 3. Read a few runs and ensures they have the same permissions.

	return true, nil
}
