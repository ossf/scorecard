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

// checkFrozenDeps is the registered name for FrozenDeps.
const checkFrozenDeps = "Frozen-Deps"

// ErrInvalidDockerfile : Invalid docker file.
var ErrInvalidDockerfile = errors.New("invalid docker file")

// ErrEmptyFile : Invalid docker file.
var ErrEmptyFile = errors.New("file has no content")

// Structure for workflow config.
// We only declare the fields we need.
// Github workflows format: https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions
type gitHubActionWorkflowConfig struct {
	// nolint: govet
	Jobs map[string]struct {
		Name  string `yaml:"name"`
		Steps []struct {
			Name  string `yaml:"name"`
			ID    string `yaml:"id"`
			Uses  string `yaml:"uses"`
			Shell string `yaml:"shell"`
			Run   string `yaml:"run"`
		}
		Defaults struct {
			Run struct {
				Shell string `yaml:"shell"`
			} `yaml:"run"`
		} `yaml:"defaults"`
	}
	Name string `yaml:"name"`
}

//nolint:gochecknoinits
func init() {
	registerCheck(checkFrozenDeps, FrozenDeps)
}

// FrozenDeps will check the repository if it contains frozen dependecies.
func FrozenDeps(c *checker.CheckRequest) checker.CheckResult {
	return checker.MultiCheckAnd(
		isPackageManagerLockFilePresent,
		isGitHubActionsWorkflowPinned,
		isDockerfilePinned,
		isDockerfileFreeOfInsecureDownloads,
		isShellScriptFreeOfInsecureDownloads,
		isGitHubWorkflowScriptFreeOfInsecureDownloads,
	)(c)
}

// TODO(laurent): need to support GCB pinning.

func isShellScriptFreeOfInsecureDownloads(c *checker.CheckRequest) checker.CheckResult {
	r, err := CheckFilesContent2("*", false, c, validateShellScriptDownloads)
	if err != nil {
		// TODO: check for the repo retry error, which should be a common
		// scorecard error independent of the underlying implementation.
		return checker.MakeInternalErrorResult(checkFrozenDeps, err)
	}
	if !r {
		// We need not provide a reason/code because it's already done
		// in validateDockerfile via `Fail` call.
		return checker.MakeFailResultWithHighConfidence(checkBinaryArtifacts)
	}

	return checker.MakePassResultWithHighConfidenceAndReasonAndCode(checkBinaryArtifacts, c,
		"BinaryDownload", "no binary downloads found in shell scripts")
}

func validateShellScriptDownloads(pathfn string, content []byte,
	cl checker.CheckLogger) (bool, error) {
	// Validate the file type.
	if !isShellScriptFile(pathfn, content) {
		return true, nil
	}
	return validateShellFile(pathfn, content, cl)
}

func isDockerfileFreeOfInsecureDownloads(c *checker.CheckRequest) checker.CheckResult {
	r, err := CheckFilesContent2("*Dockerfile*", false, c, validateDockerfileDownloads)
	if err != nil {
		// TODO: check for the repo retry error, which should be a common
		// scorecard error independent of the underlying implementation.
		return checker.MakeInternalErrorResult(checkFrozenDeps, err)
	}
	if !r {
		// We need not provide a reason/code because it's already done
		// in validateDockerfile via `Fail` call.
		return checker.MakeFailResultWithHighConfidence(checkBinaryArtifacts)
	}

	return checker.MakePassResultWithHighConfidenceAndReasonAndCode(checkBinaryArtifacts, c,
		"BinaryDownload", "no binary downloads found in Dockerfiles")
}

func validateDockerfileDownloads(pathfn string, content []byte,
	cl checker.CheckLogger) (bool, error) {
	contentReader := strings.NewReader(string(content))
	res, err := parser.Parse(contentReader)
	if err != nil {
		return false, fmt.Errorf("cannot read dockerfile content: %w", err)
	}

	// nolint: prealloc
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
		bytes = append(bytes, '\n')
	}
	return validateShellFile(pathfn, bytes, cl)
}

func isDockerfilePinned(c *checker.CheckRequest) checker.CheckResult {
	r, err := CheckFilesContent2("*Dockerfile*", false, c, validateDockerfile)
	if err != nil {
		// TODO: check for the repo retry error, which should be a common
		// scorecard error independent of the underlying implementation.
		return checker.MakeInternalErrorResult(checkFrozenDeps, err)
	}
	if !r {
		// We need not provide a reason/code because it's already done
		// in validateDockerfile via `Fail` call.
		return checker.MakeFailResultWithHighConfidence(checkBinaryArtifacts)
	}

	return checker.MakePassResultWithHighConfidenceAndReasonAndCode(checkBinaryArtifacts, c,
		"Dockerfile", "Dockerfile dependencies are pinned")
}

func validateDockerfile(pathfn string, content []byte,
	cl checker.CheckLogger) (bool, error) {
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
			cl.FailWithCode("Dockerfile", "%v has non-pinned dependency '%v'", pathfn, name)

		// FROM name.
		case len(valueList) == 1:
			name := valueList[0]
			if !regex.Match([]byte(name)) {
				ret = false
				cl.FailWithCode("Dockerfile", "%v has non-pinned dependency '%v'", pathfn, name)
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

func isGitHubWorkflowScriptFreeOfInsecureDownloads(c *checker.CheckRequest) checker.CheckResult {
	r, err := CheckFilesContent2(".github/workflows/*", false, c, validateGitHubWorkflowShellScriptDownloads)
	if err != nil {
		// TODO: check for the repo retry error, which should be a common
		// scorecard error independent of the underlying implementation.
		return checker.MakeInternalErrorResult(checkFrozenDeps, err)
	}
	if !r {
		// We need not provide a reason/code because it's already done
		// in validateGitHubWorkflowShellScriptDownloads via `Fail` call.
		return checker.MakeFailResultWithHighConfidence(checkBinaryArtifacts)
	}

	return checker.MakePassResultWithHighConfidenceAndReasonAndCode(checkBinaryArtifacts, c,
		"BinaryDownload", "no binary download found in GitHub workflows")
}

func validateGitHubWorkflowShellScriptDownloads(pathfn string, content []byte,
	cl checker.CheckLogger) (bool, error) {
	if len(content) == 0 {
		return false, ErrEmptyFile
	}

	var workflow gitHubActionWorkflowConfig
	err := yaml.Unmarshal(content, &workflow)
	if err != nil {
		return false, fmt.Errorf("cannot unmarshal file %v\n%v: %w", pathfn, string(content), err)
	}

	githubVarRegex := regexp.MustCompile(`{{[^{}]*}}`)
	validated := true
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#using-a-specific-shell.
	defaultShell := "bash"
	scriptContent := ""
	for _, job := range workflow.Jobs {
		if job.Defaults.Run.Shell != "" {
			defaultShell = job.Defaults.Run.Shell
		}

		for _, step := range job.Steps {
			if step.Run == "" {
				continue
			}

			shell := defaultShell
			if step.Shell != "" {
				shell = step.Shell
			}

			// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstepsrun.
			// Skip unsupported shells. We don't support Windows shells.
			if !isSupportedShell(shell) {
				continue
			}

			run := step.Run
			// We replace the `${{ github.variable }}` to avoid shell parising failures.
			script := githubVarRegex.ReplaceAll([]byte(run), []byte("GITHUB_REDACTED_VAR"))
			scriptContent = fmt.Sprintf("%v\n%v", scriptContent, string(script))
		}
	}

	if scriptContent != "" {
		validated, err = validateShellFile(pathfn, []byte(scriptContent), cl)
		if err != nil {
			return false, err
		}
	}

	return validated, nil
}

// Check pinning of github actions in workflows.
func isGitHubActionsWorkflowPinned(c *checker.CheckRequest) checker.CheckResult {
	r, err := CheckFilesContent2(".github/workflows/*", true, c, validateGitHubActionWorkflow)
	if err != nil {
		// TODO: check for the repo retry error, which should be a common
		// scorecard error independent of the underlying implementation.
		return checker.MakeInternalErrorResult(checkFrozenDeps, err)
	}
	if !r {
		// We need not provide a reason/code because it's already done
		// in validateGitHubActionWorkflow via `Fail` call.
		return checker.MakeFailResultWithHighConfidence(checkBinaryArtifacts)
	}

	// High confidence result.
	// We provide a reason to help the user.
	return checker.MakePassResultWithHighConfidenceAndReasonAndCode(checkBinaryArtifacts, c,
		"GitHubActions", "GitHub actions' dependencies are pinned")
}

// Check file content.
func validateGitHubActionWorkflow(pathfn string, content []byte, cl checker.CheckLogger) (bool, error) {
	if len(content) == 0 {
		return false, ErrEmptyFile
	}

	var workflow gitHubActionWorkflowConfig
	err := yaml.Unmarshal(content, &workflow)
	if err != nil {
		return false, fmt.Errorf("cannot unmarshal file %v\n%v: %w", pathfn, string(content), err)
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
					cl.FailWithCode("GitHubActions", "%v has non-pinned dependency '%v' (job '%v')", pathfn, step.Uses, jobName)
				}
			}
		}
	}

	return ret, nil
}

// Check presence of lock files thru validatePackageManagerFile().
func isPackageManagerLockFilePresent(c *checker.CheckRequest) checker.CheckResult {
	r, err := CheckIfFileExists2(checkFrozenDeps, c, validatePackageManagerFile)
	if err != nil {
		return checker.MakeInternalErrorResult(checkFrozenDeps, err)
	}
	if !r {
		return checker.MakeFailResultWithHighConfidenceAndReasonAndCode(checkFrozenDeps, c,
			"LockFile", "no lock file found in the repo")
	}

	// High confidence result.
	// We don't pass a `reason` because it's already done
	// thru calls to `Pass` in validatePackageManagerFile.
	return checker.MakePassResultWithHighConfidence(checkFrozenDeps)
}

// validatePackageManagerFile will validate the if frozen dependecies file name exists.
// TODO(laurent): need to differentiate between libraries and programs.
// TODO(laurent): handle multi-language repos.
func validatePackageManagerFile(name string, cl checker.CheckLogger) (bool, error) {
	switch strings.ToLower(name) {
	case "go.mod", "go.sum":
		cl.PassWithCode("LockFile", "go modules found: %s", name)
		return true, nil
	case "vendor/", "third_party/", "third-party/":
		cl.PassWithCode("LockFile", "vendor dir found: %s", name)
		return true, nil
	case "package-lock.json", "npm-shrinkwrap.json":
		cl.PassWithCode("LockFile", "nodejs packages found: %s", name)
		return true, nil
	// TODO(laurent): add check for hashbased pinning in requirements.txt - https://davidwalsh.name/hashin
	case "requirements.txt", "pipfile.lock":
		cl.PassWithCode("LockFile", "python requirements found: %s", name)
		return true, nil
	case "gemfile.lock":
		cl.PassWithCode("LockFile", "ruby gems found: %s", name)
		return true, nil
	case "cargo.lock":
		cl.PassWithCode("LockFile", "rust crates found: %s", name)
		return true, nil
	case "yarn.lock":
		cl.PassWithCode("LockFile", "yarn packages found: %s", name)
		return true, nil
	case "composer.lock":
		cl.PassWithCode("LockFile", "composer packages found: %s", name)
		return true, nil
	default:
		return false, nil
	}
}
