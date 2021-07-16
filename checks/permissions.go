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

//nolint:gci
import (
	"errors"
	"fmt"
	"strings"

	"github.com/ossf/scorecard/checker"
	sce "github.com/ossf/scorecard/errors"
	"gopkg.in/yaml.v2"
)

const CheckPermissions = "Token-Permissions"

// ErrInvalidGitHubWorkflowFile : Invalid GitHub workflow file.
var ErrInvalidGitHubWorkflowFile = errors.New("invalid GitHub workflow file")

//nolint:gochecknoinits
func init() {
	registerCheck(CheckPermissions, leastPrivilegedTokens)
}

func leastPrivilegedTokens(c *checker.CheckRequest) checker.CheckResult {
	return CheckFilesContent(CheckPermissions, ".github/workflows/*", true, c, validateGitHubActionTokenPermissions)
}

func validatePermission(key string, value interface{}, path string,
	logf func(s string, f ...interface{})) (bool, error) {
	val, ok := value.(string)
	if !ok {
		return false, ErrInvalidGitHubWorkflowFile
	}

	if strings.EqualFold(val, "write") {
		logf("!! token-permissions/github-token - %v permission set to '%v' in %v", key, val, path)
		return false, nil
	}

	return true, nil
}

func validateMapPermissions(values map[interface{}]interface{}, path string,
	logf func(s string, f ...interface{})) (bool, error) {
	permissionRead := true
	var r bool
	var err error

	// Iterate over the permission, verify keys and values are strings.
	for k, v := range values {
		key, ok := k.(string)
		if !ok {
			return false, ErrInvalidGitHubWorkflowFile
		}

		if r, err = validatePermission(key, v, path, logf); err != nil {
			return false, err
		}

		if !r {
			permissionRead = false
		}
	}
	return permissionRead, nil
}

func validateReadPermissions(config map[interface{}]interface{}, path string,
	logf func(s string, f ...interface{})) (bool, error) {
	var permissions interface{}

	// Check if permissions are set explicitly.
	permissions, ok := config["permissions"]
	if !ok {
		logf("!! token-permissions/github-token - no permission defined in %v", path)
		return false, nil
	}

	// Check the type of our values.
	switch val := permissions.(type) {
	// Empty string is nil type.
	// It defaults to 'none'
	case nil:

	// String type.
	case string:
		if !strings.EqualFold(val, "read-all") && val != "" {
			logf("!! token-permissions/github-token - permission set to '%v' in %v", val, path)
			return false, nil
		}

	// Map type.
	case map[interface{}]interface{}:
		if res, err := validateMapPermissions(val, path, logf); err != nil {
			return false, err
		} else if !res {
			return false, nil
		}

	// Invalid type.
	default:
		return false, ErrInvalidGitHubWorkflowFile
	}

	return true, nil
}

// Check file content.
func validateGitHubActionTokenPermissions(path string, content []byte,
	logf func(s string, f ...interface{})) (bool, error) {
	if len(content) == 0 {
		//nolint
		return false, sce.Create(sce.ErrRunFailure, sce.ErrInternalEmptyFile.Error())
	}

	var workflow map[interface{}]interface{}
	var r bool
	var err error
	err = yaml.Unmarshal(content, &workflow)
	if err != nil {
		return false, fmt.Errorf("!! token-permissions - cannot unmarshal file %v\n%v\n%v: %w",
			path, content, string(content), err)
	}

	// 1. Check that each file uses 'content: read' only or 'none'.
	//nolint
	// https://docs.github.com/en/actions/reference/authentication-in-a-workflow#example-1-passing-the-github_token-as-an-input,
	// https://github.blog/changelog/2021-04-20-github-actions-control-permissions-for-github_token/,
	// https://docs.github.com/en/actions/reference/authentication-in-a-workflow#modifying-the-permissions-for-the-github_token.
	if r, err = validateReadPermissions(workflow, path, logf); err != nil {
		return false, nil
	}
	if !r {
		return r, nil
	}

	// TODO(laurent): 2. Identify github actions that require write and add checks.

	// TODO(laurent): 3. Read a few runs and ensures they have the same permissions.

	return true, nil
}
