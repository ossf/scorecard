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

	"github.com/ossf/scorecard/checker"
	sce "github.com/ossf/scorecard/errors"
)

const CheckPermissions = "Token-Permissions"

//nolint:gochecknoinits
func init() {
	registerCheck(CheckPermissions, leastPrivilegedTokens)
}

func leastPrivilegedTokens(c *checker.CheckRequest) checker.CheckResult {
	r, err := CheckFilesContent2(".github/workflows/*", false, c, validateGitHubActionTokenPermissions)
	return createResultForLeastPrivilegeTokens(r, err)
}

func validatePermission(key string, value interface{}, path string,
	dl checker.DetailLogger) (bool, error) {
	val, ok := value.(string)
	if !ok {
		return false, sce.Create(sce.ErrScorecardInternal, errInvalidGitHubWorkflowFile.Error())
	}

	if strings.EqualFold(val, "write") {
		dl.Warn("'%v' permission set to '%v' in %v", key, val, path)
		return false, nil
	}

	dl.Info("'%v' permission set to '%v' in %v", key, val, path)
	return true, nil
}

func validateMapPermissions(values map[interface{}]interface{}, path string,
	dl checker.DetailLogger) (bool, error) {
	permissionRead := true
	var r bool
	var err error

	// Iterate over the permission, verify keys and values are strings.
	for k, v := range values {
		key, ok := k.(string)
		if !ok {
			return false, sce.Create(sce.ErrScorecardInternal, errInvalidGitHubWorkflowFile.Error())
		}

		if r, err = validatePermission(key, v, path, dl); err != nil {
			return false, err
		}

		if !r {
			permissionRead = false
		}
	}
	return permissionRead, nil
}

func validateReadPermissions(config map[interface{}]interface{}, path string,
	dl checker.DetailLogger) (bool, error) {
	var permissions interface{}

	// Check if permissions are set explicitly.
	permissions, ok := config["permissions"]
	if !ok {
		dl.Warn("no permission defined in %v", path)
		return false, nil
	}

	// Check the type of our values.
	switch val := permissions.(type) {
	// Empty string is nil type.
	// It defaults to 'none'
	case nil:
		dl.Info("permission set to 'none' in %v", path)
	// String type.
	case string:
		if !strings.EqualFold(val, "read-all") && val != "" {
			dl.Warn("permission set to '%v' in %v", val, path)
			return false, nil
		}
		dl.Info("permission set to '%v' in %v", val, path)

	// Map type.
	case map[interface{}]interface{}:
		if res, err := validateMapPermissions(val, path, dl); err != nil {
			return false, err
		} else if !res {
			return false, nil
		}

	// Invalid type.
	default:
		return false, sce.Create(sce.ErrScorecardInternal, errInvalidGitHubWorkflowFile.Error())
	}

	return true, nil
}

// Create the result.
func createResultForLeastPrivilegeTokens(r bool, err error) checker.CheckResult {
	if err != nil {
		return checker.CreateRuntimeErrorResult(CheckPermissions, err)
	}
	if !r {
		return checker.CreateMinScoreResult(CheckPermissions,
			"non read-only tokens detected in GitHub workflows")
	}

	return checker.CreateMaxScoreResult(CheckPermissions,
		"tokens are read-only in GitHub workflows")
}

func testValidateGitHubActionTokenPermissions(pathfn string,
	content []byte, dl checker.DetailLogger) checker.CheckResult {
	r, err := validateGitHubActionTokenPermissions(pathfn, content, dl)
	return createResultForLeastPrivilegeTokens(r, err)
}

// Check file content.
func validateGitHubActionTokenPermissions(path string, content []byte,
	dl checker.DetailLogger) (bool, error) {
	if len(content) == 0 {
		return false, sce.Create(sce.ErrScorecardInternal, errInternalEmptyFile.Error())
	}

	var workflow map[interface{}]interface{}
	var r bool
	var err error
	err = yaml.Unmarshal(content, &workflow)
	if err != nil {
		return false, sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("yaml.Unmarshal: %v", err))
	}

	// 1. Check that each file uses 'content: read' only or 'none'.
	//nolint
	// https://docs.github.com/en/actions/reference/authentication-in-a-workflow#example-1-passing-the-github_token-as-an-input,
	// https://github.blog/changelog/2021-04-20-github-actions-control-permissions-for-github_token/,
	// https://docs.github.com/en/actions/reference/authentication-in-a-workflow#modifying-the-permissions-for-the-github_token.
	if r, err = validateReadPermissions(workflow, path, dl); err != nil {
		return false, nil
	}
	if !r {
		return r, nil
	}

	// TODO(laurent): 2. Identify github actions that require write and add checks.

	// TODO(laurent): 3. Read a few runs and ensures they have the same permissions.

	return true, nil
}
