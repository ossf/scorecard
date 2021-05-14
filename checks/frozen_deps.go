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
	"regexp"
	"strings"

	"github.com/ossf/scorecard/checker"
	"gopkg.in/yaml.v2"
)

const frozenDepsStr = "Frozen-Deps"

func init() {
	registerCheck(frozenDepsStr, FrozenDeps)
}

// FrozenDeps will check the repository if it contains frozen dependecies.
func FrozenDeps(c *checker.CheckRequest) checker.CheckResult {
	return checker.MultiCheckAnd(
		lockFilesPresent,
		workflowActionsPinned,
	)(c)
}

// TODO: need to support Docker https://github.com/ossf/scorecard/issues/403

// Check file content
func onFileContent(path string, content []byte,
	logf func(s string, f ...interface{})) bool {

	// Structure for workflow config.
	// We only retrieve what we need for logging.
	// Github workflows format: https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions
	type YamlWorflowConfig struct {
		Name string `yaml:name`
		Jobs map[string]struct {
			Name  string `yaml:name`
			Steps []struct {
				Name string `yaml:name`
				Id   string `yaml:id`
				Uses string `yaml:uses`
			}
		}
	}

	var workflow YamlWorflowConfig
	err := yaml.Unmarshal(content, &workflow)
	if err != nil {
		panic(fmt.Errorf("!! frozen-deps - Cannot unmarshall file %v\n%v\n%v", path, content, string(content)))
		return false
	}

	r := true
	for jobName, job := range workflow.Jobs {
		if len(job.Name) > 0 {
			jobName = job.Name
		}
		for _, step := range job.Steps {
			if len(step.Uses) > 0 {
				// Ensure a hash at least as large as SHA1 is used (40 hex characters).
				match, err := regexp.Match("^.*@[a-f\\d]{40,}$", []byte(step.Uses))
				if err != nil {
					panic(fmt.Errorf("!! frozen-deps - Regex failed for %v", path))
				}
				if !match {
					r = false
					logf("!! frozen-deps - %v has non-pinned dependency '%v' (job \"%v\")", path, step.Uses, jobName)
				}
			}
		}

	}

	return r
}

// Check pinning of github actions in workflows.
func workflowActionsPinned(c *checker.CheckRequest) checker.CheckResult {
	return CheckFilesContent(frozenDepsStr, ".github/workflows/*", c, onFileContent)
}

// Check presence of lock files thru filePredicate().
func lockFilesPresent(c *checker.CheckRequest) checker.CheckResult {
	return CheckIfFileExists(frozenDepsStr, c, filePredicate)
}

// filePredicate will validate the if frozen dependecies file name exists.
// TODO: need to differentiate between libraries and programs.
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
	// TODO: I don't think requirement allows pinning
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
