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
	"github.com/ossf/scorecard/clients/githubrepo"
)

// CheckSecurityPolicy is the registred name for SecurityPolicy.
const CheckSecurityPolicy = "Security-Policy"

//nolint:gochecknoinits
func init() {
	registerCheck(CheckSecurityPolicy, SecurityPolicy)
}

func SecurityPolicy(c *checker.CheckRequest) checker.CheckResult {
	// check repository for repository-specific policy
	onFile := func(name string, dl checker.DetailLogger) (bool, error) {
		if strings.EqualFold(name, "security.md") {
			c.Dlogger.Info("security policy detected: %s", name)
			return true, nil
		} else if isSecurityRstFound(name) {
			c.Dlogger.Info("security policy detected: %s", name)
			return true, nil
		}
		return false, nil
	}
	r, err := CheckIfFileExists2(CheckSecurityPolicy, c, onFile)
	if err != nil {
		return checker.CreateRuntimeErrorResult(CheckSecurityPolicy, err)
	}
	if r {
		return checker.CreateMaxScoreResult(CheckSecurityPolicy, "security policy file detected")
	}

	// checking for community default within the .github folder
	// https://docs.github.com/en/github/building-a-strong-community/creating-a-default-community-health-file
	dotGitHub := c
	dotGitHub.Repo = ".github"
	dotGitHubClient := githubrepo.CreateGithubRepoClient(c.Ctx, c.Client)
	if err := dotGitHubClient.InitRepo(c.Owner, c.Repo); err != nil {
		return checker.CreateRuntimeErrorResult(CheckSecurityPolicy, err)
	}
	defer dotGitHubClient.Close()
	dotGitHub.RepoClient = dotGitHubClient

	onFile = func(name string, dl checker.DetailLogger) (bool, error) {
		if strings.EqualFold(name, "security.md") {
			dl.Info("security policy detected in .github folder: %s", name)
			return true, nil
		}
		return false, nil
	}
	r, err = CheckIfFileExists2(CheckSecurityPolicy, dotGitHub, onFile)
	if err != nil {
		return checker.CreateRuntimeErrorResult(CheckSecurityPolicy, err)
	}

	if r {
		return checker.CreateMaxScoreResult(CheckSecurityPolicy, "security policy file detected")
	}
	return checker.CreateMinScoreResult(CheckSecurityPolicy, "security policy file not detected")
}

func isSecurityRstFound(name string) bool {
	if strings.EqualFold(name, "doc/security.rst") {
		return true
	} else if strings.EqualFold(name, "docs/security.rst") {
		return true
	}
	return false
}
