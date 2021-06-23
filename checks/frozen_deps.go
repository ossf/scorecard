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

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"gopkg.in/yaml.v2"

	"github.com/ossf/scorecard/checker"
)

// CheckFrozenDeps is the registered name for FrozenDeps.
const CheckFrozenDeps = "Frozen-Deps"

// ErrInvalidDockerfile : Invalid docker file.
var ErrInvalidDockerfile = errors.New("invalid docker file")

// ErrEmptyFile : Invalid docker file.
var ErrEmptyFile = errors.New("file has no content")

//nolint:gochecknoinits
func init() {
	registerCheck(CheckFrozenDeps, FrozenDeps)
}

// FrozenDeps will check the repository if it contains frozen dependecies.
func FrozenDeps(c *checker.CheckRequest) checker.CheckResult {
	return checker.MultiCheckAnd(
		isPackageManagerLockFilePresent,
		isGitHubActionsWorkflowPinned,
		isDockerfilePinned,
		isDockerfileFreeOfInsecureDownloads,
		isShellScriptFreeOfInsecureDownloads,
	)(c)
}

// TODO(laurent): need to support GCB pinning.

func isShellScriptFreeOfInsecureDownloads(c *checker.CheckRequest) checker.CheckResult {
	return CheckFilesContent(CheckFrozenDeps, "*", false, c, validateShellScriptDownloads)
}

func validateShellScriptDownloads(pathfn string, content []byte,
	logf func(s string, f ...interface{})) (bool, error) {
	// Validate the file type.
	if !isShellScriptFile(pathfn, content) {
		return false, nil
	}

	return validateShellFile(pathfn, content, logf)
}

func isDockerfileFreeOfInsecureDownloads(c *checker.CheckRequest) checker.CheckResult {
	return CheckFilesContent(CheckFrozenDeps, "*Dockerfile*", false, c, validateDockerfileDownloads)
}

func validateDockerfileDownloads(pathfn string, content []byte,
	logf func(s string, f ...interface{})) (bool, error) {
	contentReader := strings.NewReader(string(content))
	res, err := parser.Parse(contentReader)
	if err != nil {
		return false, fmt.Errorf("cannot read dockerfile content: %w", err)
	}

	var bytes []byte

	// Walk the Dockerfile's AST.
	for _, child := range res.AST.Children {
		cmdType := child.Value
		// Only look for the 'RUN' command.
		if cmdType != "run" {
			continue
		}

		var valueList []string
		for n := child.Next; n != nil; n = n.Next {
			valueList = append(valueList, n.Value)
		}

		if len(valueList) == 0 {
			return false, ErrParsingDockerfile
		}

		// Build a file content.
		cmd := strings.Join(valueList, " ")
		bytes = append(bytes, cmd...)
		bytes = append(bytes, []byte("\n")...)
	}
	return validateShellFile(pathfn, bytes, logf)
}

func isDockerfilePinned(c *checker.CheckRequest) checker.CheckResult {
	return CheckFilesContent(CheckFrozenDeps, "*Dockerfile*", false, c, validateDockerfile)
}

func validateDockerfile(path string, content []byte,
	logf func(s string, f ...interface{})) (bool, error) {
	// Users may use various names, e.g.,
	// Dockerfile.aarch64, Dockerfile.template, Dockerfile_template, dockerfile, Dockerfile-name.template
	// Templates may trigger false positives, e.g. FROM { NAME }.

	// We have what looks like a docker file.
	// Let's interpret the content as utf8-encoded strings.
	contentReader := strings.NewReader(string(content))
	regex := regexp.MustCompile(`.*@sha256:[a-f\d]{64}`)

	ret := true
	fromFound := false
	pinnedAsNames := make(map[string]bool)
	res, err := parser.Parse(contentReader)
	if err != nil {
		return false, fmt.Errorf("cannot read dockerfile content: %w", err)
	}

	for _, child := range res.AST.Children {
		cmdType := child.Value
		if cmdType != "from" {
			continue
		}

		// New 'FROM' line found.
		fromFound = true

		var valueList []string
		for n := child.Next; n != nil; n = n.Next {
			valueList = append(valueList, n.Value)
		}

		switch {
		// scratch is no-op.
		case len(valueList) > 0 && strings.EqualFold(valueList[0], "scratch"):
			continue

		// FROM name AS newname.
		case len(valueList) == 3 && strings.EqualFold(valueList[1], "as"):
			name := valueList[0]
			asName := valueList[2]
			// Check if the name is pinned.
			// (1): name = <>@sha245:hash
			// (2): name = XXX where XXX was pinned
			_, pinned := pinnedAsNames[name]
			if pinned || regex.Match([]byte(name)) {
				// Record the asName.
				pinnedAsNames[asName] = true
				continue
			}

			// Not pinned.
			ret = false
			logf("!! frozen-deps/docker - %v has non-pinned dependency '%v'", path, name)

		// FROM name.
		case len(valueList) == 1:
			name := valueList[0]
			if !regex.Match([]byte(name)) {
				ret = false
				logf("!! frozen-deps/docker - %v has non-pinned dependency '%v'", path, name)
			}

		default:
			// That should not happen.
			return false, ErrInvalidDockerfile
		}
	}

	// The file should have at least one FROM statement.
	if !fromFound {
		return false, ErrInvalidDockerfile
	}

	return ret, nil
}

// ============================================================
// ===================== Github workflows =====================
// ============================================================.

// Check pinning of github actions in workflows.
func isGitHubActionsWorkflowPinned(c *checker.CheckRequest) checker.CheckResult {
	return CheckFilesContent(CheckFrozenDeps, ".github/workflows/*", true, c, validateGitHubActionWorkflow)
}

// Check file content.
func validateGitHubActionWorkflow(path string, content []byte, logf func(s string, f ...interface{})) (bool, error) {
	// Structure for workflow config.
	// We only retrieve what we need for logging.
	// Github workflows format: https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions
	type GitHubActionWorkflowConfig struct {
		Jobs map[string]struct {
			Name  string `yaml:"name"`
			Steps []struct {
				Name string `yaml:"name"`
				ID   string `yaml:"id"`
				Uses string `yaml:"uses"`
			}
		}
		Name string `yaml:"name"`
	}

	if len(content) == 0 {
		return false, ErrEmptyFile
	}

	var workflow GitHubActionWorkflowConfig
	err := yaml.Unmarshal(content, &workflow)
	if err != nil {
		return false, fmt.Errorf("!! frozen-deps - cannot unmarshal file %v\n%v\n%v: %w", path, content, string(content), err)
	}

	hashRegex := regexp.MustCompile(`^.*@[a-f\d]{40,}`)
	ret := true
	for jobName, job := range workflow.Jobs {
		if len(job.Name) > 0 {
			jobName = job.Name
		}
		for _, step := range job.Steps {
			if len(step.Uses) > 0 {
				// Ensure a hash at least as large as SHA1 is used (40 hex characters).
				// Example: action-name@hash
				match := hashRegex.Match([]byte(step.Uses))
				if !match {
					ret = false
					logf("!! frozen-deps/github-actions - %v has non-pinned dependency '%v' (job '%v')", path, step.Uses, jobName)
				}
			}
		}
	}

	return ret, nil
}

// ============================================================
// ================== Package manager lock files ==============
// ============================================================.

// Check presence of lock files thru validatePackageManagerFile().
func isPackageManagerLockFilePresent(c *checker.CheckRequest) checker.CheckResult {
	return CheckIfFileExists(CheckFrozenDeps, c, validatePackageManagerFile)
}

// validatePackageManagerFile will validate the if frozen dependecies file name exists.
// TODO(laurent): need to differentiate between libraries and programs.
// TODO(laurent): handle multi-language repos.
func validatePackageManagerFile(name string, logf func(s string, f ...interface{})) (bool, error) {
	switch strings.ToLower(name) {
	case "go.mod", "go.sum":
		logf("go modules found: %s", name)
		return true, nil
	case "vendor/", "third_party/", "third-party/":
		logf("vendor dir found: %s", name)
		return true, nil
	case "package-lock.json", "npm-shrinkwrap.json":
		logf("nodejs packages found: %s", name)
		return true, nil
	// TODO(laurent): add check for hashbased pinning in requirements.txt - https://davidwalsh.name/hashin
	case "requirements.txt", "pipfile.lock":
		logf("python requirements found: %s", name)
		return true, nil
	case "gemfile.lock":
		logf("ruby gems found: %s", name)
		return true, nil
	case "cargo.lock":
		logf("rust crates found: %s", name)
		return true, nil
	case "yarn.lock":
		logf("yarn packages found: %s", name)
		return true, nil
	case "composer.lock":
		logf("composer packages found: %s", name)
		return true, nil
	default:
		return false, nil
	}
}
