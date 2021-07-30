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

	"github.com/ossf/scorecard/v2/checker"
)

const CheckAutomaticDependencyUpdate = "Automatic-Dependency-Update"

//nolint
func init() {
	registerCheck(CheckAutomaticDependencyUpdate, AutomaticDependencyUpdate)
}

// AutomaticDependencyUpdate will check the repository if it contains Automatic dependency update.
func AutomaticDependencyUpdate(c *checker.CheckRequest) checker.CheckResult {
	var r bool
	err := CheckIfFileExists(CheckAutomaticDependencyUpdate, c, fileExists, &r)
	if err != nil {
		return checker.CreateRuntimeErrorResult(CheckAutomaticDependencyUpdate, err)
	}
	if !r {
		c.Dlogger.Warn("dependabot not detected")
		c.Dlogger.Warn("renovatebot not detected")
		return checker.CreateMinScoreResult(CheckAutomaticDependencyUpdate, "no update tool detected")
	}

	// High score result.
	return checker.CreateMaxScoreResult(CheckAutomaticDependencyUpdate, "update tool detected")
}

// fileExists will validate the if frozen dependencies file name exists.
func fileExists(name string, dl checker.DetailLogger, data FileCbData) (bool, error) {
	pdata := FileCbDataAsBoolPointer(data)

	switch strings.ToLower(name) {
	case ".github/dependabot.yml":
		dl.Info("dependabot detected : %s", name)
		// https://docs.renovatebot.com/configuration-options/
	case ".github/renovate.json", ".github/renovate.json5", ".renovaterc.json", "renovate.json",
		"renovate.json5", ".renovaterc":
		dl.Info("renovate detected: %s", name)
	default:
		// Continue iterating.
		return true, nil
	}

	*pdata = true
	// We found the file, no ne dto continue iterating.
	return false, nil
}
