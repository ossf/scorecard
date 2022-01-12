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
	"github.com/rhysd/actionlint"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/fileparser"
	sce "github.com/ossf/scorecard/v4/errors"
)

// CheckPinnedDependencies is the registered name for FrozenDeps.
const CheckPinnedDependencies = "Pinned-Dependencies"

// Structure to host information about pinned github
// or third party dependencies.
type worklowPinningResult struct {
	thirdParties pinnedResult
	gitHubOwned  pinnedResult
}

//nolint:gochecknoinits
func init() {
	registerCheck(CheckPinnedDependencies, PinnedDependencies)
}

// PinnedDependencies will check the repository if it contains frozen dependecies.
func PinnedDependencies(c *checker.CheckRequest) checker.CheckResult {
	// Lock file.
	/* WARNING: this code is inherently incorrect:
	- does not differentiate between libs and main
	- only looks at root folder.
	=> disabling to avoid false positives.
	lockScore, lockErr := isPackageManagerLockFilePresent(c)
	if lockErr != nil {
		return checker.CreateRuntimeErrorResult(CheckPinnedDependencies, lockErr)
	}
	*/

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
	actionScore = maxScore(0, actionScore)
	dockerFromScore = maxScore(0, dockerFromScore)
	dockerDownloadScore = maxScore(0, dockerDownloadScore)
	scriptScore = maxScore(0, scriptScore)
	actionScriptScore = maxScore(0, actionScriptScore)
	score := checker.AggregateScores(actionScore, dockerFromScore,
		dockerDownloadScore, scriptScore, actionScriptScore)

	if score == checker.MaxResultScore {
		return checker.CreateMaxScoreResult(CheckPinnedDependencies, "all dependencies are pinned")
	}
	return checker.CreateProportionalScoreResult(CheckPinnedDependencies,
		"dependency not pinned by hash detected", score, checker.MaxResultScore)
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

func dataAsWorkflowResultPointer(data fileparser.FileCbData) *worklowPinningResult {
	pdata, ok := data.(*worklowPinningResult)
	if !ok {
		// panic if it is not correct type
		panic("type need to be of worklowPinningResult")
	}
	return pdata
}

func createReturnValuesForGitHubActionsWorkflowPinned(r worklowPinningResult, infoMsg string,
	dl checker.DetailLogger, err error) (int, error) {
	if err != nil {
		return checker.InconclusiveResultScore, err
	}

	score := checker.MinResultScore

	if r.gitHubOwned != notPinned {
		score += 2
		dl.Info3(&checker.LogMessage{
			Type:   checker.FileTypeSource,
			Offset: checker.OffsetDefault,
			Text:   fmt.Sprintf("%s %s", "GitHub-owned", infoMsg),
		})
	}

	if r.thirdParties != notPinned {
		score += 8
		dl.Info3(&checker.LogMessage{
			Type:   checker.FileTypeSource,
			Offset: checker.OffsetDefault,
			Text:   fmt.Sprintf("%s %s", "Third-party", infoMsg),
		})
	}

	return score, nil
}

func dataAsResultPointer(data fileparser.FileCbData) *pinnedResult {
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
	err := fileparser.CheckFilesContent("*", false,
		c, validateShellScriptIsFreeOfInsecureDownloads, &r)
	return createReturnForIsShellScriptFreeOfInsecureDownloads(r, c.Dlogger, err)
}

func createReturnForIsShellScriptFreeOfInsecureDownloads(r pinnedResult,
	dl checker.DetailLogger, err error) (int, error) {
	return createReturnValues(r,
		"no insecure (not pinned by hash) dependency downloads found in shell scripts",
		dl, err)
}

func testValidateShellScriptIsFreeOfInsecureDownloads(pathfn string,
	content []byte, dl checker.DetailLogger) (int, error) {
	var r pinnedResult
	_, err := validateShellScriptIsFreeOfInsecureDownloads(pathfn, content, dl, &r)
	return createReturnForIsShellScriptFreeOfInsecureDownloads(r, dl, err)
}

func validateShellScriptIsFreeOfInsecureDownloads(pathfn string, content []byte,
	dl checker.DetailLogger, data fileparser.FileCbData) (bool, error) {
	pdata := dataAsResultPointer(data)

	// Validate the file type.
	if !isSupportedShellScriptFile(pathfn, content) {
		addPinnedResult(pdata, true)
		return true, nil
	}

	r, err := validateShellFile(pathfn, 0, 0 /*unknown*/, content, map[string]bool{}, dl)
	if err != nil {
		return false, err
	}

	addPinnedResult(pdata, r)
	return true, nil
}

func isDockerfileFreeOfInsecureDownloads(c *checker.CheckRequest) (int, error) {
	var r pinnedResult
	err := fileparser.CheckFilesContent("*Dockerfile*",
		false, c, validateDockerfileIsFreeOfInsecureDownloads, &r)
	return createReturnForIsDockerfileFreeOfInsecureDownloads(r, c.Dlogger, err)
}

// Create the result.
func createReturnForIsDockerfileFreeOfInsecureDownloads(r pinnedResult,
	dl checker.DetailLogger, err error) (int, error) {
	return createReturnValues(r,
		"no insecure (not pinned by hash) dependency downloads found in Dockerfiles",
		dl, err)
}

func testValidateDockerfileIsFreeOfInsecureDownloads(pathfn string,
	content []byte, dl checker.DetailLogger) (int, error) {
	var r pinnedResult
	_, err := validateDockerfileIsFreeOfInsecureDownloads(pathfn, content, dl, &r)
	return createReturnForIsDockerfileFreeOfInsecureDownloads(r, dl, err)
}

func validateDockerfileIsFreeOfInsecureDownloads(pathfn string, content []byte,
	dl checker.DetailLogger, data fileparser.FileCbData) (bool, error) {
	pdata := dataAsResultPointer(data)

	// Return early if this is a script, e.g. script_dockerfile_something.sh
	if isShellScriptFile(pathfn, content) {
		addPinnedResult(pdata, true)
		return true, nil
	}

	if !fileparser.CheckFileContainsCommands(content, "#") {
		addPinnedResult(pdata, true)
		return true, nil
	}

	contentReader := strings.NewReader(string(content))
	res, err := parser.Parse(contentReader)
	if err != nil {
		return false, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("%v: %v", errInternalInvalidDockerFile, err))
	}

	// Walk the Dockerfile's AST.
	taintedFiles := make(map[string]bool)
	for i := range res.AST.Children {
		var bytes []byte

		child := res.AST.Children[i]
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
			return false, sce.WithMessage(sce.ErrScorecardInternal, errInternalInvalidDockerFile.Error())
		}

		// Build a file content.
		cmd := strings.Join(valueList, " ")
		bytes = append(bytes, cmd...)
		r, err := validateShellFile(pathfn, uint(child.StartLine)-1, uint(child.EndLine)-1,
			bytes, taintedFiles, dl)
		if err != nil {
			return false, err
		}
		addPinnedResult(pdata, r)
	}

	return true, nil
}

func isDockerfilePinned(c *checker.CheckRequest) (int, error) {
	var r pinnedResult
	err := fileparser.CheckFilesContent("*Dockerfile*", false,
		c, validateDockerfileIsPinned, &r)
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
	dl checker.DetailLogger, data fileparser.FileCbData) (bool, error) {
	// Users may use various names, e.g.,
	// Dockerfile.aarch64, Dockerfile.template, Dockerfile_template, dockerfile, Dockerfile-name.template

	pdata := dataAsResultPointer(data)
	// Return early if this is a script, e.g. script_dockerfile_something.sh
	if isShellScriptFile(pathfn, content) {
		addPinnedResult(pdata, true)
		return true, nil
	}

	if !fileparser.CheckFileContainsCommands(content, "#") {
		addPinnedResult(pdata, true)
		return true, nil
	}

	if fileparser.IsTemplateFile(pathfn) {
		addPinnedResult(pdata, true)
		return true, nil
	}

	// We have what looks like a docker file.
	// Let's interpret the content as utf8-encoded strings.
	contentReader := strings.NewReader(string(content))
	// The dependency must be pinned by sha256 hash, e.g.,
	// FROM something@sha256:${ARG},
	// FROM something:@sha256:45b23dee08af5e43a7fea6c4cf9c25ccf269ee113168c19722f87876677c5cb2
	regex := regexp.MustCompile(`.*@sha256:([a-f\d]{64}|\${.*})`)

	ret := true
	pinnedAsNames := make(map[string]bool)
	res, err := parser.Parse(contentReader)
	if err != nil {
		return false, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("%v: %v", errInternalInvalidDockerFile, err))
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
			pinned := pinnedAsNames[name]
			if pinned || regex.Match([]byte(name)) {
				// Record the asName.
				pinnedAsNames[asName] = true
				continue
			}

			// Not pinned.
			ret = false
			dl.Warn3(&checker.LogMessage{
				Path:      pathfn,
				Type:      checker.FileTypeSource,
				Offset:    uint(child.StartLine),
				EndOffset: uint(child.EndLine),
				Text:      "docker image not pinned by hash",
				Snippet:   child.Original,
			})

		// FROM name.
		case len(valueList) == 1:
			name := valueList[0]
			pinned := pinnedAsNames[name]
			if !pinned && !regex.Match([]byte(name)) {
				ret = false
				dl.Warn3(&checker.LogMessage{
					Path:      pathfn,
					Type:      checker.FileTypeSource,
					Offset:    uint(child.StartLine),
					EndOffset: uint(child.EndLine),
					Text:      "docker image not pinned by hash",
					Snippet:   child.Original,
				})
			}

		default:
			// That should not happen.
			return false, sce.WithMessage(sce.ErrScorecardInternal, errInternalInvalidDockerFile.Error())
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
	err := fileparser.CheckFilesContent(".github/workflows/*", false,
		c, validateGitHubWorkflowIsFreeOfInsecureDownloads, &r)
	return createReturnForIsGitHubWorkflowScriptFreeOfInsecureDownloads(r, c.Dlogger, err)
}

// Create the result.
func createReturnForIsGitHubWorkflowScriptFreeOfInsecureDownloads(r pinnedResult,
	dl checker.DetailLogger, err error) (int, error) {
	return createReturnValues(r,
		"no insecure (not pinned by hash) dependency downloads found in GitHub workflows",
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
// nolint: gocognit
func validateGitHubWorkflowIsFreeOfInsecureDownloads(pathfn string, content []byte,
	dl checker.DetailLogger, data fileparser.FileCbData) (bool, error) {
	if !fileparser.IsWorkflowFile(pathfn) {
		return true, nil
	}

	pdata := dataAsResultPointer(data)

	if !fileparser.CheckFileContainsCommands(content, "#") {
		addPinnedResult(pdata, true)
		return true, nil
	}

	workflow, errs := actionlint.Parse(content)
	if len(errs) > 0 && workflow == nil {
		// actionlint is a linter, so it will return errors when the yaml file does not meet its linting standards.
		// Often we don't care about these errors.
		return false, fileparser.FormatActionlintError(errs)
	}

	githubVarRegex := regexp.MustCompile(`{{[^{}]*}}`)
	for jobName, job := range workflow.Jobs {
		jobName := jobName
		job := job
		if len(fileparser.GetJobName(job)) > 0 {
			jobName = fileparser.GetJobName(job)
		}
		taintedFiles := make(map[string]bool)
		for _, step := range job.Steps {
			step := step
			if !fileparser.IsStepExecKind(step, actionlint.ExecKindRun) {
				continue
			}

			execRun, ok := step.Exec.(*actionlint.ExecRun)
			if !ok {
				stepName := fileparser.GetStepName(step)
				return false, sce.WithMessage(sce.ErrScorecardInternal,
					fmt.Sprintf("unable to parse step '%v' for job '%v'", jobName, stepName))
			}

			if execRun == nil || execRun.Run == nil {
				// Cannot check further, continue.
				continue
			}

			run := execRun.Run.Value
			// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstepsrun.
			shell, err := fileparser.GetShellForStep(step, job)
			if err != nil {
				return false, err
			}
			// Skip unsupported shells. We don't support Windows shells or some Unix shells.
			if !isSupportedShell(shell) {
				continue
			}

			// We replace the `${{ github.variable }}` to avoid shell parsing failures.
			script := githubVarRegex.ReplaceAll([]byte(run), []byte("GITHUB_REDACTED_VAR"))
			validated, err := validateShellFile(pathfn, uint(execRun.Run.Pos.Line), uint(execRun.Run.Pos.Line),
				script, taintedFiles, dl)
			if err != nil {
				return false, err
			}
			addPinnedResult(pdata, validated)
		}
	}

	return true, nil
}

// Check pinning of github actions in workflows.
func isGitHubActionsWorkflowPinned(c *checker.CheckRequest) (int, error) {
	var r worklowPinningResult
	err := fileparser.CheckFilesContent(".github/workflows/*",
		true, c, validateGitHubActionWorkflow, &r)
	return createReturnForIsGitHubActionsWorkflowPinned(r, c.Dlogger, err)
}

// Create the result.
func createReturnForIsGitHubActionsWorkflowPinned(r worklowPinningResult, dl checker.DetailLogger,
	err error) (int, error) {
	return createReturnValuesForGitHubActionsWorkflowPinned(r,
		"actions are pinned",
		dl, err)
}

func testIsGitHubActionsWorkflowPinned(pathfn string, content []byte, dl checker.DetailLogger) (int, error) {
	var r worklowPinningResult
	_, err := validateGitHubActionWorkflow(pathfn, content, dl, &r)
	return createReturnForIsGitHubActionsWorkflowPinned(r, dl, err)
}

func generateOwnerToDisplay(gitHubOwned bool) string {
	if gitHubOwned {
		return "GitHub-owned"
	}
	return "third-party"
}

// validateGitHubActionWorkflow checks if the workflow file contains unpinned actions. Returns true if the check
// should continue executing after this file.
func validateGitHubActionWorkflow(pathfn string, content []byte,
	dl checker.DetailLogger, data fileparser.FileCbData) (bool, error) {
	if !fileparser.IsWorkflowFile(pathfn) {
		return true, nil
	}

	pdata := dataAsWorkflowResultPointer(data)

	if !fileparser.CheckFileContainsCommands(content, "#") {
		addWorkflowPinnedResult(pdata, true, true)
		addWorkflowPinnedResult(pdata, true, true)
		return true, nil
	}

	workflow, errs := actionlint.Parse(content)
	if len(errs) > 0 && workflow == nil {
		// actionlint is a linter, so it will return errors when the yaml file does not meet its linting standards.
		// Often we don't care about these errors.
		return false, fileparser.FormatActionlintError(errs)
	}

	hashRegex := regexp.MustCompile(`^.*@[a-f\d]{40,}`)
	for jobName, job := range workflow.Jobs {
		jobName := jobName
		job := job
		if len(fileparser.GetJobName(job)) > 0 {
			jobName = fileparser.GetJobName(job)
		}
		for _, step := range job.Steps {
			if !fileparser.IsStepExecKind(step, actionlint.ExecKindAction) {
				continue
			}
			execAction, ok := step.Exec.(*actionlint.ExecAction)
			if !ok {
				stepName := fileparser.GetStepName(step)
				return false, sce.WithMessage(sce.ErrScorecardInternal,
					fmt.Sprintf("unable to parse step '%v' for job '%v'", jobName, stepName))
			}

			if execAction == nil || execAction.Uses == nil {
				// Cannot check further, continue.
				continue
			}

			// nolint:lll
			// Check whether this is an action defined in the same repo,
			// https://docs.github.com/en/actions/learn-github-actions/finding-and-customizing-actions#referencing-an-action-in-the-same-repository-where-a-workflow-file-uses-the-action.
			if strings.HasPrefix(execAction.Uses.Value, "./") {
				continue
			}

			// Check if we are dealing with a GitHub action or a third-party one.
			gitHubOwned := fileparser.IsGitHubOwnedAction(execAction.Uses.Value)
			owner := generateOwnerToDisplay(gitHubOwned)

			// Ensure a hash at least as large as SHA1 is used (40 hex characters).
			// Example: action-name@hash
			match := hashRegex.Match([]byte(execAction.Uses.Value))
			if !match {
				dl.Warn3(&checker.LogMessage{
					Path: pathfn, Type: checker.FileTypeSource,
					Offset:    uint(execAction.Uses.Pos.Line),
					EndOffset: uint(execAction.Uses.Pos.Line), // `Uses` always span a single line.
					Snippet:   execAction.Uses.Value,
					Text:      fmt.Sprintf("%s action not pinned by hash", owner),
				})
			}

			addWorkflowPinnedResult(pdata, match, gitHubOwned)
		}
	}

	return true, nil
}

func addWorkflowPinnedResult(w *worklowPinningResult, to, isGitHub bool) {
	if isGitHub {
		addPinnedResult(&w.gitHubOwned, to)
	} else {
		addPinnedResult(&w.thirdParties, to)
	}
}

// Check presence of lock files thru validatePackageManagerFile().
//nolint:unused,deadcode
func isPackageManagerLockFilePresent(c *checker.CheckRequest) (int, error) {
	var r pinnedResult
	err := fileparser.CheckIfFileExists(c, validatePackageManagerFile, &r)
	if err != nil {
		return checker.InconclusiveResultScore, fmt.Errorf("%w", err)
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
//nolint:unused
func validatePackageManagerFile(name string, dl checker.DetailLogger, data fileparser.FileCbData) (bool, error) {
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
