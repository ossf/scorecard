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
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/ossf/scorecard/checker"
	"gopkg.in/yaml.v2"
)

const frozenDepsStr = "Frozen-Deps"

// ErrFrozenDepsInvalidDockerfile : Invalid docker file.
var ErrFrozenDepsInvalidDockerfile = errors.New("invalid docker file")

// ErrFrozenDepsEmptyFile : Invalid docker file.
var ErrFrozenDepsEmptyFile = errors.New("file has no content")

func init() {
	registerCheck(frozenDepsStr, FrozenDeps)
}

// FrozenDeps will check the repository if it contains frozen dependecies.
func FrozenDeps(c *checker.CheckRequest) checker.CheckResult {
	return checker.MultiCheckAnd(
		isPackageManagerLockFilePresent,
		isGitHubActionsWorkflowPinned,
		isDockerfilePinned,
	)(c)
}

// TODO(laurent): need to support GCB

// ============================================================
// ======================== Dockerfiles =======================
// ============================================================
func isDockerfilePinned(c *checker.CheckRequest) checker.CheckResult {
	return CheckFilesContent(frozenDepsStr, "*Dockerfile*", false, c, validateDockerfile)
}

func isPresent(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func validateDockerfile(path string, content []byte,
	logf func(s string, f ...interface{})) (bool, error) {
	// Users may use various names, e.g.,
	// Dockerfile.aarch64, Dockerfile.template, Dockerfile_template, dockerfile, Dockerfile-name.template
	// Templates may trigger false positives, e.g. FROM { NAME }.

	// We have what looks like a docker file.
	// Let's interpret the content as utf8-encoded strings.
<<<<<<< HEAD
	contentReader := strings.NewReader(string(content))
	regex := regexp.MustCompile(`.*@sha256:[a-f\d]{64}`)
=======
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
<<<<<<< HEAD
	asRegex := regexp.MustCompile(`^FROM\s+(.*)\s+AS\s+(.*)`)
	regex := regexp.MustCompile(`^FROM\s+(.*)`)
	hashAsRegex := regexp.MustCompile(`^FROM\s+.*@sha256:[a-f\d]{64}\s+AS\s+(.*)`)
	hashRegex := regexp.MustCompile(`^FROM\s+.*@sha256:[a-f\d]{64}`)
>>>>>>> e767aa7 (check docker files hash pinning)
=======
	asRegex := regexp.MustCompile(`^(?i)FROM\s+(.*)\s+AS\s+(.*)`)
	regex := regexp.MustCompile(`^(?i)FROM\s+(.*)`)
	hashAsRegex := regexp.MustCompile(`^(?i)FROM\s+.*@sha256:[a-f\d]{64}\s+AS\s+(.*)`)
	hashRegex := regexp.MustCompile(`^(?i)FROM\s+.*@sha256:[a-f\d]{64}`)
>>>>>>> 19851e8 (make keyword matches case-insensitive)

	r := true
<<<<<<< HEAD
	fromFound := false
	var pinnedAsNames []string

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

		// FROM name AS newname.
		if len(valueList) == 3 && strings.ToLower(valueList[1]) == "as" {
			name := valueList[0]
			asName := valueList[2]
			// Check if the name is pinned.
			// (1): name = <>@sha245:hash
			// (2): name = XXX where XXX was pinned
			if regex.Match([]byte(name)) || isPresent(pinnedAsNames, name) {
				// Record the asName.
				if !isPresent(pinnedAsNames, asName) {
					pinnedAsNames = append(pinnedAsNames, asName)
				}
				continue
			}

			// Not pinned.
=======
	nl := 0
	var al []string
	for scanner.Scan() {
		line := scanner.Text()
		// Only look at lines starting with FROM.
		if !strings.HasPrefix(strings.ToLower(line), "from ") {
			continue
		}

		// New line found
		nl += 1
		fmt.Printf("line: %v\n", line)
		// FROM name@sha256:hash AS newname.
		// In this case, we record newname. It's pinned
		// so it can be re-used as 'FROM new name' later.
		re := hashAsRegex.FindStringSubmatch(line)
		if len(re) == 2 {
			// Record the newname.
			al = append(al, re[1])
			continue
		}

		// FROM oldname AS newname
		// where oldname refers to a pinned image
		re = asRegex.FindStringSubmatch(line)
		if len(re) == 3 {
			oldname := re[1]
			newname := re[2]
			if !isPresent(al, oldname) {
				r = false
				logf("!! frozen-deps - %v has non-pinned dependency '%v'", path, line)
				continue
			}
			// Record the newname if not alresdy present in our list.
			if !isPresent(al, newname) {
				al = append(al, newname)
				fmt.Printf("array: %v\n", al)
			}
			continue
		}

		// FROM name
		// where name refers to a pinned image
		re = regex.FindStringSubmatch(line)
		if len(re) == 2 && isPresent(al, re[1]) {
			continue
		}

		// FROM name@sha256:hash
		if !hashRegex.Match([]byte(line)) {
>>>>>>> e767aa7 (check docker files hash pinning)
			r = false
			logf("!! frozen-deps - %v has non-pinned dependency '%v'", path, name)
			continue

		} else if len(valueList) == 1 {
			// FROM name
			name := valueList[0]
			if !regex.Match([]byte(name)) {
				r = false
				logf("!! frozen-deps - %v has non-pinned dependency '%v'", path, name)
				continue
			}
		} else {
			// That should not happen.
			return false, ErrFrozenDepsInvalidDockerfile
		}

	}

	// The file should have at least one FROM statement.
	if !fromFound {
		logf("end")
		return false, ErrFrozenDepsInvalidDockerfile
	}

	return r, nil
}

// ============================================================
// ===================== Github workflows =====================
// ============================================================

// Check pinning of github actions in workflows.
func isGitHubActionsWorkflowPinned(c *checker.CheckRequest) checker.CheckResult {
	return CheckFilesContent(frozenDepsStr, ".github/workflows/*", true, c, validateGitHubActionWorkflow)
}

// Check file content
func validateGitHubActionWorkflow(path string, content []byte,
	logf func(s string, f ...interface{})) (bool, error) {

	// Structure for workflow config.
	// We only retrieve what we need for logging.
	// Github workflows format: https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions
	type GitHubActionWorkflowConfig struct {
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

	if len(content) == 0 {
		return false, ErrFrozenDepsEmptyFile
	}

	var workflow GitHubActionWorkflowConfig
	err := yaml.Unmarshal(content, &workflow)
	if err != nil {
		return false, fmt.Errorf("!! frozen-deps - cannot unmarshal file %v\n%v\n%v: %w", path, content, string(content), err)
	}

	hashRegex := regexp.MustCompile(`^.*@[a-f\d]{40,}`)
	r := true
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
					r = false
					logf("!! frozen-deps - %v has non-pinned dependency '%v' (job \"%v\")", path, step.Uses, jobName)
				}
			}
		}

	}

	return r, nil
}

// ============================================================
// ================== Package manager lock files ==============
// ============================================================

// Check presence of lock files thru validatePackageManagerFile().
func isPackageManagerLockFilePresent(c *checker.CheckRequest) checker.CheckResult {
	return CheckIfFileExists(frozenDepsStr, c, validatePackageManagerFile)
}

// validatePackageManagerFile will validate the if frozen dependecies file name exists.
// TODO(laurent): need to differentiate between libraries and programs.
// TODO(laurent): handle multi-language repos
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
