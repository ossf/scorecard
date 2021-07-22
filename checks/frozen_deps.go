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
	"fmt"
	"regexp"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"gopkg.in/yaml.v2"

	"github.com/ossf/scorecard/v2/checker"
	sce "github.com/ossf/scorecard/v2/errors"
)

// CheckFrozenDeps is the registered name for FrozenDeps.
const CheckFrozenDeps = "Frozen-Deps"

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
	registerCheck(CheckFrozenDeps, FrozenDeps)
}

// FrozenDeps will check the repository if it contains frozen dependecies.
func FrozenDeps(c *checker.CheckRequest) checker.CheckResult {
	score, reason, err := isPackageManagerLockFilePresent(c)
	if err != nil {
		return checker.CreateRuntimeErrorResult(CheckFrozenDeps, err)
	}
	// return checker.MultiCheckAnd2(
	// 	isGitHubActionsWorkflowPinned,
	// 	isDockerfilePinned,
	// 	isDockerfileFreeOfInsecureDownloads,
	// 	isShellScriptFreeOfInsecureDownloads,
	// 	isGitHubWorkflowScriptFreeOfInsecureDownloads,
	// )(c)
}

// TODO(laurent): need to support GCB pinning.

func isShellScriptFreeOfInsecureDownloads(c *checker.CheckRequest) checker.CheckResult {
	r, err := CheckFilesContent("*", false, c, validateShellScriptIsFreeOfInsecureDownloads)
	return createResultForIsShellScriptFreeOfInsecureDownloads(r, err)
}

func createResultForIsShellScriptFreeOfInsecureDownloads(r bool, err error) checker.CheckResult {
	if err != nil {
		return checker.CreateRuntimeErrorResult(CheckFrozenDeps, err)
	}
	if !r {
		return checker.CreateMinScoreResult(CheckFrozenDeps,
			"insecure (unpinned) dependency downloads found in shell scripts")
	}

	return checker.CreateMaxScoreResult(CheckFrozenDeps,
		"no insecure (unpinned) dependency downloads found in shell scripts")
}

func testValidateShellScriptIsFreeOfInsecureDownloads(pathfn string,
	content []byte, dl checker.DetailLogger) checker.CheckResult {
	r, err := validateShellScriptIsFreeOfInsecureDownloads(pathfn, content, dl)
	return createResultForIsShellScriptFreeOfInsecureDownloads(r, err)
}

func validateShellScriptIsFreeOfInsecureDownloads(pathfn string, content []byte,
	dl checker.DetailLogger) (bool, error) {
	// Validate the file type.
	if !isShellScriptFile(pathfn, content) {
		return true, nil
	}
	return validateShellFile(pathfn, content, dl)
}

func isDockerfileFreeOfInsecureDownloads(c *checker.CheckRequest) checker.CheckResult {
	r, err := CheckFilesContent("*Dockerfile*", false, c, validateDockerfileIsFreeOfInsecureDownloads)
	return createResultForIsDockerfileFreeOfInsecureDownloads(r, err)
}

// Create the result.
func createResultForIsDockerfileFreeOfInsecureDownloads(r bool, err error) checker.CheckResult {
	if err != nil {
		return checker.CreateRuntimeErrorResult(CheckFrozenDeps, err)
	}
	if !r {
		return checker.CreateMinScoreResult(CheckFrozenDeps,
			"insecure (unpinned) dependency downloads found in Dockerfiles")
	}

	return checker.CreateMaxScoreResult(CheckFrozenDeps,
		"no insecure (unpinned) dependency downloads found in Dockerfiles")
}

func testValidateDockerfileIsFreeOfInsecureDownloads(pathfn string,
	content []byte, dl checker.DetailLogger) checker.CheckResult {
	r, err := validateDockerfileIsFreeOfInsecureDownloads(pathfn, content, dl)
	return createResultForIsDockerfileFreeOfInsecureDownloads(r, err)
}

func validateDockerfileIsFreeOfInsecureDownloads(pathfn string, content []byte,
	dl checker.DetailLogger) (bool, error) {
	contentReader := strings.NewReader(string(content))
	res, err := parser.Parse(contentReader)
	if err != nil {
		//nolint
		return false, sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("%v: %v", errInternalInvalidDockerFile, err))
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
			//nolint
			return false, sce.Create(sce.ErrScorecardInternal, errInternalInvalidDockerFile.Error())
		}

		// Build a file content.
		cmd := strings.Join(valueList, " ")
		bytes = append(bytes, cmd...)
		bytes = append(bytes, '\n')
	}
	return validateShellFile(pathfn, bytes, dl)
}

func isDockerfilePinned(c *checker.CheckRequest) checker.CheckResult {
	r, err := CheckFilesContent("*Dockerfile*", false, c, validateDockerfileIsPinned)
	return createResultForIsDockerfilePinned(r, err)
}

// Create the result.
func createResultForIsDockerfilePinned(r bool, err error) checker.CheckResult {
	if err != nil {
		return checker.CreateRuntimeErrorResult(CheckFrozenDeps, err)
	}
	if r {
		return checker.CreateMaxScoreResult(CheckFrozenDeps, "Dockerfile dependencies are pinned")
	}

	return checker.CreateMinScoreResult(CheckFrozenDeps, "unpinned dependencies found Dockerfiles")
}

func testValidateDockerfileIsPinned(pathfn string, content []byte, dl checker.DetailLogger) checker.CheckResult {
	r, err := validateDockerfileIsPinned(pathfn, content, dl)
	return createResultForIsDockerfilePinned(r, err)
}

func validateDockerfileIsPinned(pathfn string, content []byte,
	dl checker.DetailLogger) (bool, error) {
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
		//nolint
		return false, sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("%v: %v", errInternalInvalidDockerFile, err))
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
			dl.Warn("unpinned dependency detected in %v: '%v'", pathfn, name)

		// FROM name.
		case len(valueList) == 1:
			name := valueList[0]
			if !regex.Match([]byte(name)) {
				ret = false
				dl.Warn("unpinned dependency detected in %v: '%v'", pathfn, name)
			}

		default:
			// That should not happen.
			//nolint
			return false, sce.Create(sce.ErrScorecardInternal, errInternalInvalidDockerFile.Error())
		}
	}

	// The file should have at least one FROM statement.
	if !fromFound {
		//nolint
		return false, sce.Create(sce.ErrScorecardInternal, errInternalInvalidDockerFile.Error())
	}

	return ret, nil
}

func isGitHubWorkflowScriptFreeOfInsecureDownloads(c *checker.CheckRequest) checker.CheckResult {
	r, err := CheckFilesContent(".github/workflows/*", false, c, validateGitHubWorkflowIsFreeOfInsecureDownloads)
	return createResultForIsGitHubWorkflowScriptFreeOfInsecureDownloads(r, err)
}

// Create the result.
func createResultForIsGitHubWorkflowScriptFreeOfInsecureDownloads(r bool, err error) checker.CheckResult {
	if err != nil {
		return checker.CreateRuntimeErrorResult(CheckFrozenDeps, err)
	}
	if !r {
		return checker.CreateMinScoreResult(CheckFrozenDeps,
			"insecure (unpinned) dependency downloads found in GitHub workflows")
	}

	return checker.CreateMaxScoreResult(CheckFrozenDeps,
		"no insecure (unpinned) dependency downloads found in GitHub workflows")
}

func testValidateGitHubWorkflowScriptFreeOfInsecureDownloads(pathfn string,
	content []byte, dl checker.DetailLogger) checker.CheckResult {
	r, err := validateGitHubWorkflowIsFreeOfInsecureDownloads(pathfn, content, dl)
	return createResultForIsGitHubWorkflowScriptFreeOfInsecureDownloads(r, err)
}

func validateGitHubWorkflowIsFreeOfInsecureDownloads(pathfn string, content []byte,
	dl checker.DetailLogger) (bool, error) {
	if len(content) == 0 {
		//nolint
		return false, sce.Create(sce.ErrScorecardInternal, errInternalEmptyFile.Error())
	}

	var workflow gitHubActionWorkflowConfig
	err := yaml.Unmarshal(content, &workflow)
	if err != nil {
		//nolint
		return false, sce.Create(sce.ErrScorecardInternal,
			fmt.Sprintf("%v:%s:%s:%v", errInternalInvalidYamlFile, pathfn, string(content), err))
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
		validated, err = validateShellFile(pathfn, []byte(scriptContent), dl)
		if err != nil {
			return false, err
		}
	}

	return validated, nil
}

// Check pinning of github actions in workflows.
func isGitHubActionsWorkflowPinned(c *checker.CheckRequest) checker.CheckResult {
	r, err := CheckFilesContent(".github/workflows/*", true, c, validateGitHubActionWorkflow)
	return createResultForIsGitHubActionsWorkflowPinned(r, err)
}

// Create the result.
func createResultForIsGitHubActionsWorkflowPinned(r bool, err error) (int, string, error) {
	if err != nil {
		return checker.InconclusiveResultScore, "", err
	}
	if r {
		return checker.MaxResultScore, "GitHub actions are pinned", nil
	}

	return checker.MinResultScore, "GitHub actions are not pinned", nil
}

func testIsGitHubActionsWorkflowPinned(pathfn string, content []byte, dl checker.DetailLogger) checker.CheckResult {
	r, err := validateGitHubActionWorkflow(pathfn, content, dl)
	return createResultForIsGitHubActionsWorkflowPinned(r, err)
}

// Check file content.
func validateGitHubActionWorkflow(pathfn string, content []byte, dl checker.DetailLogger) (bool, error) {
	if len(content) == 0 {
		//nolint
		return false, sce.Create(sce.ErrScorecardInternal, errInternalEmptyFile.Error())
	}

	var workflow gitHubActionWorkflowConfig
	err := yaml.Unmarshal(content, &workflow)
	if err != nil {
		//nolint
		return false, sce.Create(sce.ErrScorecardInternal,
			fmt.Sprintf("%v:%s:%s:%v", errInternalInvalidYamlFile, pathfn, string(content), err))
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
					dl.Warn("unpinned dependency detected in %v: '%v' (job '%v')", pathfn, step.Uses, jobName)
				}
			}
		}
	}

	return ret, nil
}

// Check presence of lock files thru validatePackageManagerFile().
func isPackageManagerLockFilePresent(c *checker.CheckRequest) checker.CheckResult {
	r, err := CheckIfFileExists(CheckFrozenDeps, c, validatePackageManagerFile)
	if err != nil {
		return checker.InconclusiveResultScore, "", err
	}
	if !r {
		return checker.InconclusiveResultScore, "no lock files detected for a package manager", nil
	}

	return checker.MaxResultScore, "lock file detected for a package manager", nil
}

// validatePackageManagerFile will validate the if frozen dependecies file name exists.
// TODO(laurent): need to differentiate between libraries and programs.
// TODO(laurent): handle multi-language repos.
func validatePackageManagerFile(name string, dl checker.DetailLogger) (bool, error) {
	switch strings.ToLower(name) {
	// TODO(laurent): "go.mod" is for libraries
	case "go.sum":
		dl.Info("go lock file detected: %s", name)
		return true, nil
	case "vendor/", "third_party/", "third-party/":
		dl.Info("vendoring detected in: %s", name)
		return true, nil
	case "package-lock.json", "npm-shrinkwrap.json":
		dl.Info("javascript lock file detected: %s", name)
		return true, nil
	// TODO(laurent): add check for hashbased pinning in requirements.txt - https://davidwalsh.name/hashin
	// Note: because requirements.txt does not handle transitive dependencies, we consider it
	// not a lock file, until we have remediation steps for pip-build.
	case "pipfile.lock":
		dl.Info("python lock file detected: %s", name)
		return true, nil
	case "gemfile.lock":
		dl.Info("ruby lock file detected: %s", name)
		return true, nil
	case "cargo.lock":
		dl.Info("rust lock file detected: %s", name)
		return true, nil
	case "yarn.lock":
		dl.Info("yarn lock file detected: %s", name)
		return true, nil
	case "composer.lock":
		dl.Info("composer lock file detected: %s", name)
		return true, nil
	default:
		return false, nil
	}
}
