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
	registerCheck("Frozen-Deps", FrozenDeps)
}

// FrozenDeps will check the repository if it contains frozen dependecies.
func FrozenDeps(c checker.Checker) checker.CheckResult {
	r := CheckIfFileExists(c, filePredicate)
	//nolint
	r.Description = `This check tries to determine if a project has declared and pinned its dependencies. It works by looking for a set of well-known package management lock files.`
	r.HelpURL = "https://github.com/ossf/scorecard/blob/main/checks.md#frozen-deps"
	return r
}

// filePredicate will validate the if frozen dependecies file name exists.
func filePredicate(name string, logf func(s string, f ...interface{})) bool {
	switch strings.ToLower(name) {
	case "go.mod", "go.sum":
		logf("go modules found: %s", name)
		return true
	case "vendor/", "third_party/", "third-party/":
		logf("vendor dir found: %s", name)
		return true
	case "package-lock.json", "npm-shrinkwrap.json":
		logf("nodejs packages found: %s", name)
		return true
	case "requirements.txt", "pipfile.lock":
		logf("python requirements found: %s", name)
		return true
	case "gemfile.lock":
		logf("ruby gems found: %s", name)
		return true
	case "cargo.lock":
		logf("rust crates found: %s", name)
		return true
	case "yarn.lock":
		logf("yarn packages found: %s", name)
		return true
	case "composer.lock":
		logf("composer packages found: %s", name)
		return true
	default:
		return false
	}
}
