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

const checkAutomaticDependencyUpdate = "Automatic-Dependency-Update"

//nolint
func init() {
	registerCheck(checkAutomaticDependencyUpdate, AutomaticDependencyUpdate)
}

// AutomaticDependencyUpdate will check the repository if it contains Automatic dependency update.
func AutomaticDependencyUpdate(c *checker.CheckRequest) checker.CheckResult {
	r, err := CheckIfFileExists2(checkAutomaticDependencyUpdate, c, fileExists)
	if err != nil {
		return checker.MakeInconclusiveResult(checkAutomaticDependencyUpdate, err)
	}
	if !r {
		c.CLogger.Fail("E01", "no configuration file found in the repo")
		return checker.MakeInconclusiveResult(checkAutomaticDependencyUpdate, nil)
	}

	// We're confident it's correct.
	return checker.MakePassResult(checkAutomaticDependencyUpdate)
}

// fileExists will validate the if frozen dependencies file name exists.
func fileExists(name string, cl checker.CheckLogger) (bool, error) {
	switch strings.ToLower(name) {
	case ".github/dependabot.yml":
		cl.Pass("dependabot config found: %s", name)
		return true, nil
		// https://docs.renovatebot.com/configuration-options/
	case ".github/renovate.json", ".github/renovate.json5", ".renovaterc.json", "renovate.json",
		"renovate.json5", ".renovaterc":
		cl.Pass("renovate config found: %s", name)
		return true, nil
	default:
		return false, nil
	}
}
