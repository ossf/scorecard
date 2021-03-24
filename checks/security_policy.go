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
	"strings"

	"github.com/ossf/scorecard/checker"
)

func init() {
	registerCheck("Security-Policy", SecurityPolicy)
}

func SecurityPolicy(c checker.Checker) checker.CheckResult {
	// check repository for repository-specific policy
	result := CheckIfFileExists(c, func(name string, logf func(s string, f ...interface{})) bool {
		if strings.EqualFold(name, "security.md") {
			logf("security policy : %s", name)
			return true
		}
		return false
	})
	//nolint
	result.Description = `This check tries to determine if a project has published a security policy. It works by looking for a file named SECURITY.md (case-insensitive) in a few well-known directories.`
	result.HelpURL = "https://github.com/ossf/scorecard/blob/main/checks.md#security-policy"
	if result.Pass {
		return result
	}

	// checking for community default within the .github folder
	// https://docs.github.com/en/github/building-a-strong-community/creating-a-default-community-health-file
	dotGitHub := c
	dotGitHub.Repo = ".github"

	result = CheckIfFileExists(dotGitHub, func(name string, logf func(s string, f ...interface{})) bool {
		if strings.EqualFold(name, "security.md") {
			logf("security policy within .github folder : %s", name)
			return true
		}
		return false
	})

	//nolint
	result.Description = `This check tries to determine if a project has published a security policy. It works by looking for a file named SECURITY.md (case-insensitive) in a few well-known directories.`
	result.HelpURL = "https://github.com/ossf/scorecard/blob/main/checks.md#security-policy"
	return result
}
