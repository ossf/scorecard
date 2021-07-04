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

// CheckSecurityPolicy is the registred name for SecurityPolicy.
const CheckSecurityPolicy = "Security-Policy"

//nolint:gochecknoinits
func init() {
	registerCheck(CheckSecurityPolicy, SecurityPolicy)
}

func SecurityPolicy(c *checker.CheckRequest) checker.CheckResult {
	// check repository for repository-specific policy
	onFile := func(name string, logf func(s string, f ...interface{})) (bool, error) {
		if strings.EqualFold(name, "security.md") {
			logf("security policy : %s", name)
			return true, nil
		} else if isSecurityrstFound(name) {
			logf("security policy : %s", name)
			return true, nil
		}
		return false, nil
	}
	result := CheckIfFileExists(CheckSecurityPolicy, c, onFile)

	if result.Pass {
		return result
	}

	// checking for community default within the .github folder
	// https://docs.github.com/en/github/building-a-strong-community/creating-a-default-community-health-file
	dotGitHub := c
	dotGitHub.Repo = ".github"

	onFile = func(name string, logf func(s string, f ...interface{})) (bool, error) {
		if strings.EqualFold(name, "security.md") {
			logf("security policy within .github folder : %s", name)
			return true, nil
		}
		return false, nil
	}
	return CheckIfFileExists(CheckSecurityPolicy, dotGitHub, onFile)
}

func isSecurityrstFound(name string) bool {
	if strings.EqualFold(name, "doc/security.rst") {
		return true
	} else if strings.EqualFold(name, "docs/security.rst") {
		return true
	}
	return false
}
