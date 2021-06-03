// Copyright 2020 Security Scorecard Authors
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

	"github.com/ossf/scorecard/checker"
	"gopkg.in/yaml.v2"
)

const CheckPermissions = "Token-Permissions"

//nolint:gochecknoinits
func init() {
	registerCheck(CheckPermissions, leastPrivilegedTokens)
}

func leastPrivilegedTokens(c *checker.CheckRequest) checker.CheckResult {
	return CheckFilesContent(CheckPermissions, ".github/workflows/*", true, c, validateGitHubActionTokenPermissions)
}

type GitHubActionWorkflowConfig struct {
	Name        string `yaml:"name"`
	Permissions *struct {
		Contents       *string `yaml:"contents"`
		Actions        *string `yaml:"actions"`
		Checks         *string `yaml:"checks"`
		Deploymments   *string `yaml:"deployments"`
		Issues         *string `yaml:"issues"`
		Metadata       *string `yaml:"metadata"`
		Packages       *string `yaml:"packages"`
		PullRequests   *string `yaml:"pull requests"`
		Projects       *string `yaml:"repository projects"`
		SecurityEvents *string `yaml:"security events"`
		Statuses       *string `yaml:"statuses"`
	} `yaml:"permissions"`
}

func isPermissionWrite(permission *string) bool {
	return permission != nil &&
		*permission != "read" &&
		*permission != "none"
}

func permissionToString(permission *string) string {
	if permission != nil {
		return *permission
	}
	return "read"
}

func validateDefaultReadPermissions(config GitHubActionWorkflowConfig, path string,
	logf func(s string, f ...interface{})) bool {
	if config.Permissions == nil {
		logf("!! token-permissions/token - no permission defined in %v", path)
		return false
	}

	r := true
	if isPermissionWrite(config.Permissions.Contents) {
		logf("!! token-permissions/token - contents permission set to '%v' in %v", permissionToString(config.Permissions.Contents), path)
		r = false
	}

	if isPermissionWrite(config.Permissions.Actions) {
		logf("!! token-permissions/token - actions permission set to '%v' in %v", permissionToString(config.Permissions.Actions), path)
		r = false
	}

	if isPermissionWrite(config.Permissions.Checks) {
		logf("!! token-permissions/token - checks permission set to '%v' in %v", permissionToString(config.Permissions.Checks), path)
		r = false
	}

	if isPermissionWrite(config.Permissions.Deploymments) {
		logf("!! token-permissions/token - deployments permission set to '%v' in %v", permissionToString(config.Permissions.Deploymments), path)
		r = false
	}

	if isPermissionWrite(config.Permissions.Issues) {
		logf("!! token-permissions/token - issue permission set to '%v' in %v", permissionToString(config.Permissions.Issues), path)
		r = false
	}

	if isPermissionWrite(config.Permissions.Metadata) {
		logf("!! token-permissions/token - metadata permission set to '%v' in %v", permissionToString(config.Permissions.Metadata), path)
		r = false
	}

	if isPermissionWrite(config.Permissions.Packages) {
		logf("!! token-permissions/token - packages permission set to '%v' in %v", permissionToString(config.Permissions.Packages), path)
		r = false
	}

	if isPermissionWrite(config.Permissions.PullRequests) {
		logf("!! token-permissions/token - pull requests permission set to '%v' in %v",
			permissionToString(config.Permissions.PullRequests), path)
		r = false
	}

	if isPermissionWrite(config.Permissions.Projects) {
		logf("!! token-permissions/token - repository projects permission set to '%v' in %v",
			permissionToString(config.Permissions.Projects), path)
		r = false
	}

	if isPermissionWrite(config.Permissions.SecurityEvents) {
		logf("!! token-permissions/token - security events permission set to '%v' in %v",
			permissionToString(config.Permissions.SecurityEvents), path)
		r = false
	}

	if isPermissionWrite(config.Permissions.Statuses) {
		logf("!! token-permissions/token - statuses permission set to '%v' in %v",
			permissionToString(config.Permissions.Statuses), path)
		r = false
	}

	return r
}

// Check file content.
func validateGitHubActionTokenPermissions(path string, content []byte,
	logf func(s string, f ...interface{})) (bool, error) {
	if len(content) == 0 {
		return false, ErrEmptyFile
	}

	var workflow GitHubActionWorkflowConfig
	err := yaml.Unmarshal(content, &workflow)
	if err != nil {
		return false, fmt.Errorf("!! token-permissions - cannot unmarshal file %v\n%v\n%v: %w",
			path, content, string(content), err)
	}

	// 1. Check that each file uses 'content: read' only or 'none'.
	// https://docs.github.com/en/actions/reference/authentication-in-a-workflow#example-1-passing-the-github_token-as-an-input,
	// https://github.blog/changelog/2021-04-20-github-actions-control-permissions-for-github_token/,
	// https://docs.github.com/en/actions/reference/authentication-in-a-workflow#modifying-the-permissions-for-the-github_token.
	if !validateDefaultReadPermissions(workflow, path, logf) {
		return false, nil
	}

	// TODO(laurent): 2. Read a few runs and ensures they have the same permissions.

	// TODO(laurent): 3. Identify github actions that require write and add checks.
	return true, nil
}
