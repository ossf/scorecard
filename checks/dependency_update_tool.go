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

// CheckDependencyUpdateTool is the exported name for Automatic-Depdendency-Update.
const CheckDependencyUpdateTool = "Dependency-Update-Tool"

//nolint
func init() {
	registerCheck(CheckDependencyUpdateTool, UsesDependencyUpdateTool)
}

// UsesDependencyUpdateTool will check the repository uses a dependency update tool.
func UsesDependencyUpdateTool(c *checker.CheckRequest) checker.CheckResult {
	var r bool
	err := CheckIfFileExists(CheckDependencyUpdateTool, c, fileExists, &r)
	if err != nil {
		return checker.CreateRuntimeErrorResult(CheckDependencyUpdateTool, err)
	}
	if !r {
		c.Dlogger.Warn3(&checker.LogMessage{
			Text: `dependabot config file not detected in source location.
			We recommend setting this configuration in code so it can be easily verified by others.`,
		})
		c.Dlogger.Warn3(&checker.LogMessage{
			Text: `renovatebot config file not detected in source location.
			We recommend setting this configuration in code so it can be easily verified by others.`,
		})
		return checker.CreateMinScoreResult(CheckDependencyUpdateTool, "no update tool detected")
	}

	// High score result.
	return checker.CreateMaxScoreResult(CheckDependencyUpdateTool, "update tool detected")
}

// fileExists will validate the if frozen dependencies file name exists.
func fileExists(name string, dl checker.DetailLogger, data FileCbData) (bool, error) {
	pdata := FileGetCbDataAsBoolPointer(data)

	switch strings.ToLower(name) {
	case ".github/dependabot.yml":
		dl.Info3(&checker.LogMessage{
			Path: name,
			Type: checker.FileTypeSource,
			// Source file must have line number > 0.
			Offset: 1,
			Text:   "dependabot detected",
		})
		// https://docs.renovatebot.com/configuration-options/
	case ".github/renovate.json", ".github/renovate.json5", ".renovaterc.json", "renovate.json",
		"renovate.json5", ".renovaterc":
		dl.Info3(&checker.LogMessage{
			Path: name,
			Type: checker.FileTypeSource,
			// Source file must have line number > 0.
			Offset: 1,
			Text:   "renovate detected",
		})
	default:
		// Continue iterating.
		return true, nil
	}

	*pdata = true
	// We found the file, no need to continue iterating.
	return false, nil
}
