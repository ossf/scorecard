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
	"path"
	"regexp"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"gopkg.in/yaml.v3"

	"github.com/ossf/scorecard/v2/checker"
	sce "github.com/ossf/scorecard/v2/errors"
)

// CheckPinnedDependencies is the registered name for FrozenDeps.
const CheckPinnedDependencies = "Pinned-Dependencies"

// Structure for workflow config.
// We only declare the fields we need.
// Github workflows format: https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions
type gitHubActionWorkflowConfig struct {
	// nolint: govet
	Jobs map[string]struct {
		Name  string `yaml:"name"`
		Steps []struct {
			Name  string         `yaml:"name"`
			ID    string         `yaml:"id"`
			Uses  stringWithLine `yaml:"uses"`
			Shell string         `yaml:"shell"`
			Run   string         `yaml:"run"`
		}
		Defaults struct {
			Run struct {
				Shell string `yaml:"shell"`
			} `yaml:"run"`
		} `yaml:"defaults"`
	}
	Name string `yaml:"name"`
}

// stringWithLine is for when you want to keep track of the line number that the string came from.
type stringWithLine struct {
	Value string
	Line  int
}

func (ws *stringWithLine) UnmarshalYAML(value *yaml.Node) error {
	err := value.Decode(&ws.Value)
	if err != nil {
		//nolint:wrapcheck
		return sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("error decoding stringWithLine Value: %v", err))
	}
	ws.Line = value.Line

	return nil
}

//nolint:gochecknoinits
func init() {
	registerCheck(CheckPinnedDependencies, PinnedDependencies)
}

// PinnedDependencies will check the repository if it contains frozen dependecies.
func PinnedDependencies(c *checker.CheckRequest) checker.CheckResult {
	// Lock file.
	lockScore, lockErr := isPackageManagerLockFilePresent(c)
	if lockErr != nil {
		return checker.CreateRuntimeErrorResult(CheckPinnedDependencies, lockErr)
	}

	// GitHub actions.
	actionScore, actionErr := isGitHubActionsWorkflowPinned(c)
	if actionErr != nil {
		return checker.CreateRuntimeErrorResult(CheckPinnedDependencies, actionErr)
	}

	// Docker files.
	dockerFromScore, dockerFromErr := isDockerfilePinned(c)
	if dockerFromErr != nil {
		return checker.CreateRuntimeErrorResult(CheckPinnedDependencies, dockerFromErr)
	}

	// Docker downloads.
	dockerDownloadScore, dockerDownloadErr := isDockerfileFreeOfInsecureDownloads(c)
	if dockerDownloadErr != nil {
		return checker.CreateRuntimeErrorResult(CheckPinnedDependencies, dockerDownloadErr)
	}

	// Script downloads.
	scriptScore, scriptError := isShellScriptFreeOfInsecureDownloads(c)
	if scriptError != nil {
		return checker.CreateRuntimeErrorResult(CheckPinnedDependencies, scriptError)
	}

	// Action script downloads.
	actionScriptScore, actionScriptError := isGitHubWorkflowScriptFreeOfInsecureDownloads(c)
	if actionScriptError != nil {
		return checker.CreateRuntimeErrorResult(CheckPinnedDependencies, actionScriptError)
	}

	// Scores may be inconclusive.
	lockScore = maxScore(0, lockScore)
	actionScore = maxScore(0, actionScore)
	dockerFromScore = maxScore(0, dockerFromScore)
	dockerDownloadScore = maxScore(0, dockerDownloadScore)
	scriptScore = maxScore(0, scriptScore)
	actionScriptScore = maxScore(0, actionScriptScore)
	score := checker.AggregateScores(lockScore, actionScore, dockerFromScore,
		dockerDownloadScore, scriptScore, actionScriptScore)

	if score == checker.MaxResultScore {
		return checker.CreateMaxScoreResult(CheckPinnedDependencies, "all dependencies are pinned")
	}
	return checker.CreateProportionalScoreResult(CheckPinnedDependencies,
		"unpinned dependencies detected", score, checker.MaxResultScore)
}

// TODO(laurent): need to support GCB pinning.
//nolint
func maxScore(s1, s2 int) int {
	if s1 > s2 {
		return s1
	}
	return s2
}

type pinnedResult int

const (
	pinnedUndefined pinnedResult = iota
	pinned
	notPinned
)

// For the 'to' param, true means the file is pinning dependencies (or there are no dependencies),
// false means there are unpinned dependencies.
func addPinnedResult(r *pinnedResult, to bool) {
	// If the result is `notPinned`, we keep it.
	// In other cases, we always update the result.
	if *r == notPinned {
		return
	}

	switch to {
	case true:
		*r = pinned
	case false:
		*r = notPinned
	}
}

func dataAsResultPointer(data FileCbData) *pinnedResult {
	pdata, ok := data.(*pinnedResult)
	if !ok {
		// This never happens.
		panic("invalid type")
	}
	return pdata
}

func createReturnValues(r pinnedResult, infoMsg string, dl checker.DetailLogger, err error) (int, error) {
	if err != nil {
		return checker.InconclusiveResultScore, err
	}

	switch r {
	default:
		panic("invalid value")
	case pinned, pinnedUndefined:
		dl.Info(infoMsg)
		return checker.MaxResultScore, nil
	case notPinned:
		// No logging needed as it's done by the checks.
		return checker.MinResultScore, nil
	}
}

func isShellScriptFreeOfInsecureDownloads(c *checker.CheckRequest) (int, error) {
	var r pinnedResult
	err := CheckFilesContent("*", false, c, validateShellScriptIsFreeOfInsecureDownloads, &r)
	return createReturnForIsShellScriptFreeOfInsecureDownloads(r, c.Dlogger, err)
}

func createReturnForIsShellScriptFreeOfInsecureDownloads(r pinnedResult,
	dl checker.DetailLogger, err error) (int, error) {
	return createReturnValues(r,
		"no insecure (unpinned) dependency downloads found in shell scripts",
		dl, err)
}

func testValidateShellScriptIsFreeOfInsecureDownloads(pathfn string,
	content []byte, dl checker.DetailLogger) (int, error) {
	var r pinnedResult
	_, err := validateShellScriptIsFreeOfInsecureDownloads(pathfn, content, dl, &r)
	return createReturnForIsShellScriptFreeOfInsecureDownloads(r, dl, err)
}

func validateShellScriptIsFreeOfInsecureDownloads(pathfn string, content []byte,
	dl checker.DetailLogger, data FileCbData) (bool, error) {
	pdata := dataAsResultPointer(data)

	// Validate the file type.
	if !isSupportedShellScriptFile(pathfn, content) {
		addPinnedResult(pdata, true)
		return true, nil
	}

	r, err := validateShellFile(pathfn, content, dl)
	if err != nil {
		return false, err
	}

	addPinnedResult(pdata, r)
	return true, nil
}

func isDockerfileFreeOfInsecureDownloads(c *checker.CheckRequest) (int, error) {
	var r pinnedResult
	err := CheckFilesContent("*Dockerfile*", false, c, validateDockerfileIsFreeOfInsecureDownloads, &r)
	return createReturnForIsDockerfileFreeOfInsecureDownloads(r, c.Dlogger, err)
}

// Create the result.
func createReturnForIsDockerfileFreeOfInsecureDownloads(r pinnedResult,
	dl checker.DetailLogger, err error) (int, error) {
	return createReturnValues(r,
		"no insecure (unpinned) dependency downloads found in Dockerfiles",
		dl, err)
}

func testValidateDockerfileIsFreeOfInsecureDownloads(pathfn string,
	content []byte, dl checker.DetailLogger) (int, error) {
	var r pinnedResult
	_, err := validateDockerfileIsFreeOfInsecureDownloads(pathfn, content, dl, &r)
	return createReturnForIsDockerfileFreeOfInsecureDownloads(r, dl, err)
}

func validateDockerfileIsFreeOfInsecureDownloads(pathfn string, content []byte,
	dl checker.DetailLogger, data FileCbData) (bool, error) {
	pdata := dataAsResultPointer(data)

	// Return early if this is a script, e.g. script_dockerfile_something.sh
	if isShellScriptFile(pathfn, content) {
		addPinnedResult(pdata, true)
		return true, nil
	}

	if !CheckFileContainsCommands(content, "#") {
		addPinnedResult(pdata, true)
		return true, nil
	}

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

	r, err := validateShellFile(pathfn, bytes, dl)
	if err != nil {
		return false, err
	}

	addPinnedResult(pdata, r)
	return true, nil
}

func isDockerfilePinned(c *checker.CheckRequest) (int, error) {
	var r pinnedResult
	err := CheckFilesContent("*Dockerfile*", false, c, validateDockerfileIsPinned, &r)
	return createReturnForIsDockerfilePinned(r, c.Dlogger, err)
}

// Create the result.
func createReturnForIsDockerfilePinned(r pinnedResult, dl checker.DetailLogger, err error) (int, error) {
	return createReturnValues(r,
		"Dockerfile dependencies are pinned",
		dl, err)
}

func testValidateDockerfileIsPinned(pathfn string, content []byte, dl checker.DetailLogger) (int, error) {
	var r pinnedResult
	_, err := validateDockerfileIsPinned(pathfn, content, dl, &r)
	return createReturnForIsDockerfilePinned(r, dl, err)
}

func validateDockerfileIsPinned(pathfn string, content []byte,
	dl checker.DetailLogger, data FileCbData) (bool, error) {
	// Users may use various names, e.g.,
	// Dockerfile.aarch64, Dockerfile.template, Dockerfile_template, dockerfile, Dockerfile-name.template
	// Templates may trigger false positives, e.g. FROM { NAME }.

	pdata := dataAsResultPointer(data)
	// Return early if this is a script, e.g. script_dockerfile_something.sh
	if isShellScriptFile(pathfn, content) {
		addPinnedResult(pdata, true)
		return true, nil
	}

	if !CheckFileContainsCommands(content, "#") {
		addPinnedResult(pdata, true)
		return true, nil
	}

	// We have what looks like a docker file.
	// Let's interpret the content as utf8-encoded strings.
	contentReader := strings.NewReader(string(content))
	regex := regexp.MustCompile(`.*@sha256:[a-f\d]{64}`)

	ret := true
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

	//nolint
	// The file need not have a FROM statement,
	// https://github.com/tensorflow/tensorflow/blob/master/tensorflow/tools/dockerfiles/partials/jupyter.partial.Dockerfile.

	addPinnedResult(pdata, ret)
	return true, nil
}

func isGitHubWorkflowScriptFreeOfInsecureDownloads(c *checker.CheckRequest) (int, error) {
	var r pinnedResult
	err := CheckFilesContent(".github/workflows/*", false, c, validateGitHubWorkflowIsFreeOfInsecureDownloads, &r)
	return createReturnForIsGitHubWorkflowScriptFreeOfInsecureDownloads(r, c.Dlogger, err)
}

// Create the result.
func createReturnForIsGitHubWorkflowScriptFreeOfInsecureDownloads(r pinnedResult,
	dl checker.DetailLogger, err error) (int, error) {
	return createReturnValues(r,
		"no insecure (unpinned) dependency downloads found in GitHub workflows",
		dl, err)
}

func testValidateGitHubWorkflowScriptFreeOfInsecureDownloads(pathfn string,
	content []byte, dl checker.DetailLogger) (int, error) {
	var r pinnedResult
	_, err := validateGitHubWorkflowIsFreeOfInsecureDownloads(pathfn, content, dl, &r)
	return createReturnForIsGitHubWorkflowScriptFreeOfInsecureDownloads(r, dl, err)
}

// validateGitHubWorkflowIsFreeOfInsecureDownloads checks if the workflow file downloads dependencies that are unpinned.
// Returns true if the check should continue executing after this file.
func validateGitHubWorkflowIsFreeOfInsecureDownloads(pathfn string, content []byte,
	dl checker.DetailLogger, data FileCbData) (bool, error) {
	if !isWorkflowFile(pathfn) {
		return true, nil
	}

	pdata := dataAsResultPointer(data)

	if !CheckFileContainsCommands(content, "#") {
		addPinnedResult(pdata, true)
		return true, nil
	}

	var workflow gitHubActionWorkflowConfig
	err := yaml.Unmarshal(content, &workflow)
	if err != nil {
		//nolint
		return false, sce.Create(sce.ErrScorecardInternal,
			fmt.Sprintf("%v: %v", errInternalInvalidYamlFile, err))
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

	addPinnedResult(pdata, validated)
	return true, nil
}

// Check pinning of github actions in workflows.
func isGitHubActionsWorkflowPinned(c *checker.CheckRequest) (int, error) {
	var r pinnedResult
	err := CheckFilesContent(".github/workflows/*", true, c, validateGitHubActionWorkflow, &r)
	return createReturnForIsGitHubActionsWorkflowPinned(r, c.Dlogger, err)
}

// Create the result.
func createReturnForIsGitHubActionsWorkflowPinned(r pinnedResult, dl checker.DetailLogger, err error) (int, error) {
	return createReturnValues(r,
		"GitHub actions are pinned",
		dl, err)
}

func testIsGitHubActionsWorkflowPinned(pathfn string, content []byte, dl checker.DetailLogger) (int, error) {
	var r pinnedResult
	_, err := validateGitHubActionWorkflow(pathfn, content, dl, &r)
	return createReturnForIsGitHubActionsWorkflowPinned(r, dl, err)
}

// validateGitHubActionWorkflow checks if the workflow file contains unpinned actions. Returns true if the check
// should continue executing after this file.
func validateGitHubActionWorkflow(pathfn string, content []byte,
	dl checker.DetailLogger, data FileCbData) (bool, error) {
	if !isWorkflowFile(pathfn) {
		return true, nil
	}

	pdata := dataAsResultPointer(data)

	if !CheckFileContainsCommands(content, "#") {
		addPinnedResult(pdata, true)
		return true, nil
	}

	var workflow gitHubActionWorkflowConfig
	err := yaml.Unmarshal(content, &workflow)
	if err != nil {
		//nolint
		return false, sce.Create(sce.ErrScorecardInternal,
			fmt.Sprintf("%v: %v", errInternalInvalidYamlFile, err))
	}

	hashRegex := regexp.MustCompile(`^.*@[a-f\d]{40,}`)
	ret := true
	for jobName, job := range workflow.Jobs {
		if len(job.Name) > 0 {
			jobName = job.Name
		}
		for _, step := range job.Steps {
			if len(step.Uses.Value) > 0 {
				// Ensure a hash at least as large as SHA1 is used (40 hex characters).
				// Example: action-name@hash
				match := hashRegex.Match([]byte(step.Uses.Value))
				if !match {
					ret = false
					dl.Warn3(&checker.LogMessage{
						Path: pathfn, Type: checker.FileTypeSource, Offset: step.Uses.Line, Snippet: step.Uses.Value,
						Text: fmt.Sprintf("unpinned dependency detected (job '%v')", jobName),
					})
				}
			}
		}
	}

	addPinnedResult(pdata, ret)
	return true, nil
}

// isWorkflowFile returns true if this is a GitHub workflow file.
func isWorkflowFile(pathfn string) bool {
	// From https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions:
	// "Workflow files use YAML syntax, and must have either a .yml or .yaml file extension."
	switch path.Ext(pathfn) {
	case ".yml", ".yaml":
		return true
	default:
		return false
	}
}

// Check presence of lock files thru validatePackageManagerFile().
func isPackageManagerLockFilePresent(c *checker.CheckRequest) (int, error) {
	var r pinnedResult
	err := CheckIfFileExists(CheckPinnedDependencies, c, validatePackageManagerFile, &r)
	if err != nil {
		return checker.InconclusiveResultScore, err
	}
	if r != pinned {
		c.Dlogger.Warn("no lock files detected for a package manager")
		return checker.InconclusiveResultScore, nil
	}

	return checker.MaxResultScore, nil
}

// validatePackageManagerFile will validate the if frozen dependecies file name exists.
// TODO(laurent): need to differentiate between libraries and programs.
// TODO(laurent): handle multi-language repos.
func validatePackageManagerFile(name string, dl checker.DetailLogger, data FileCbData) (bool, error) {
	switch strings.ToLower(name) {
	// TODO(laurent): "go.mod" is for libraries
	default:
		return true, nil
	case "go.sum":
		dl.Info("go lock file detected: %s", name)
	case "vendor/", "third_party/", "third-party/":
		dl.Info("vendoring detected in: %s", name)
	case "package-lock.json", "npm-shrinkwrap.json":
		dl.Info("javascript lock file detected: %s", name)
	// TODO(laurent): add check for hashbased pinning in requirements.txt - https://davidwalsh.name/hashin
	// Note: because requirements.txt does not handle transitive dependencies, we consider it
	// not a lock file, until we have remediation steps for pip-build.
	case "pipfile.lock":
		dl.Info("python lock file detected: %s", name)
	case "gemfile.lock":
		dl.Info("ruby lock file detected: %s", name)
	case "cargo.lock":
		dl.Info("rust lock file detected: %s", name)
	case "yarn.lock":
		dl.Info("yarn lock file detected: %s", name)
	case "composer.lock":
		dl.Info("composer lock file detected: %s", name)
	}

	pdata := dataAsResultPointer(data)
	addPinnedResult(pdata, true)
	return false, nil
}
