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

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/evaluation"
	"github.com/ossf/scorecard/v4/checks/raw"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/remediation"
)

// CheckPinnedDependencies is the registered name for FrozenDeps.
const CheckPinnedDependencies = "Pinned-Dependencies"

//nolint:gochecknoinits
func init() {
	supportedRequestTypes := []checker.RequestType{
		checker.FileBased,
		checker.CommitBased,
	}
	if err := registerCheck(CheckPinnedDependencies, PinningDependencies, supportedRequestTypes); err != nil {
		// This should never happen.
		panic(err)
	}
}

<<<<<<< HEAD
// PinningDependencies will check the repository for its use of dependencies.
func PinningDependencies(c *checker.CheckRequest) checker.CheckResult {
	if err := remediation.Setup(c); err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("remdiationSetup: %v", err))
		return checker.CreateRuntimeErrorResult(CheckPinnedDependencies, e)
=======
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

	if remErr := remediation.Setup(c); remErr != nil {
		return checker.CreateRuntimeErrorResult(CheckPinnedDependencies, remErr)
>>>>>>> 465fc76a (updates)
	}

	rawData, err := raw.PinningDependencies(c)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckPinnedDependencies, e)
	}

	// Set the raw results.
	if c.RawResults != nil {
		c.RawResults.PinningDependenciesResults = rawData
	}

<<<<<<< HEAD
	return evaluation.PinningDependencies(CheckPinnedDependencies, c.Dlogger, &rawData)
=======
	if r.thirdParties != notPinned {
		score += 8
		dl.Info(&checker.LogMessage{
			Type:   checker.FileTypeSource,
			Offset: checker.OffsetDefault,
			Text:   fmt.Sprintf("%s %s", "Third-party", infoMsg),
		})
	}

	return score, nil
}

func createReturnValues(r pinnedResult, infoMsg string, dl checker.DetailLogger, err error) (int, error) {
	if err != nil {
		return checker.InconclusiveResultScore, err
	}

	switch r {
	default:
		panic("invalid value")
	case pinned, pinnedUndefined:
		dl.Info(&checker.LogMessage{
			Text: infoMsg,
		})
		return checker.MaxResultScore, nil
	case notPinned:
		// No logging needed as it's done by the checks.
		return checker.MinResultScore, nil
	}
}

func isShellScriptFreeOfInsecureDownloads(c *checker.CheckRequest) (int, error) {
	var r pinnedResult
	err := fileparser.OnMatchingFileContentDo(c.RepoClient, fileparser.PathMatcher{
		Pattern:       "*",
		CaseSensitive: false,
	}, validateShellScriptIsFreeOfInsecureDownloads, c.Dlogger, &r)
	return createReturnForIsShellScriptFreeOfInsecureDownloads(r, c.Dlogger, err)
}

func createReturnForIsShellScriptFreeOfInsecureDownloads(r pinnedResult,
	dl checker.DetailLogger, err error,
) (int, error) {
	return createReturnValues(r,
		"no insecure (not pinned by hash) dependency downloads found in shell scripts",
		dl, err)
}

var validateShellScriptIsFreeOfInsecureDownloads fileparser.DoWhileTrueOnFileContent = func(
	pathfn string,
	content []byte,
	args ...interface{},
) (bool, error) {
	if len(args) != 2 {
		return false, fmt.Errorf(
			"validateShellScriptIsFreeOfInsecureDownloads requires exactly 2 arguments: %w", errInvalidArgLength)
	}

	pdata := dataAsResultPointer(args[1])
	dl := dataAsDetailLogger(args[0])

	// Validate the file type.
	if !isSupportedShellScriptFile(pathfn, content) {
		addPinnedResult(pdata, true)
		return true, nil
	}
	r, err := validateShellFile(pathfn, 0, 0 /*unknown*/, content, map[string]bool{}, dl)
	if err != nil {
		// Ignore parsing errors.
		if errors.Is(err, sce.ErrorShellParsing) {
			addPinnedResult(pdata, true)
		}

		return false, nil
	}

	addPinnedResult(pdata, r)
	return true, nil
}

func isDockerfileFreeOfInsecureDownloads(c *checker.CheckRequest) (int, error) {
	var r pinnedResult
	err := fileparser.OnMatchingFileContentDo(c.RepoClient, fileparser.PathMatcher{
		Pattern:       "*Dockerfile*",
		CaseSensitive: false,
	}, validateDockerfileIsFreeOfInsecureDownloads, c.Dlogger, &r)
	return createReturnForIsDockerfileFreeOfInsecureDownloads(r, c.Dlogger, err)
}

// Create the result.
func createReturnForIsDockerfileFreeOfInsecureDownloads(r pinnedResult,
	dl checker.DetailLogger, err error,
) (int, error) {
	return createReturnValues(r,
		"no insecure (not pinned by hash) dependency downloads found in Dockerfiles",
		dl, err)
}

func isDockerfile(pathfn string, content []byte) bool {
	if strings.HasSuffix(pathfn, ".go") ||
		strings.HasSuffix(pathfn, ".c") ||
		strings.HasSuffix(pathfn, ".cpp") ||
		strings.HasSuffix(pathfn, ".rs") ||
		strings.HasSuffix(pathfn, ".js") ||
		strings.HasSuffix(pathfn, ".py") ||
		strings.HasSuffix(pathfn, ".pyc") ||
		strings.HasSuffix(pathfn, ".java") ||
		isShellScriptFile(pathfn, content) {
		return false
	}
	return true
}

var validateDockerfileIsFreeOfInsecureDownloads fileparser.DoWhileTrueOnFileContent = func(
	pathfn string,
	content []byte,
	args ...interface{},
) (bool, error) {
	if len(args) != 2 {
		return false, fmt.Errorf(
			"validateDockerfileIsFreeOfInsecureDownloads requires exactly 2 arguments: %w", errInvalidArgLength)
	}
	pdata := dataAsResultPointer(args[1])
	dl := dataAsDetailLogger(args[0])

	// Return early if this is not a docker file.
	if !isDockerfile(pathfn, content) {
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
		if cmdType != "RUN" {
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
			// Ignore parsing errors.
			if errors.Is(err, sce.ErrorShellParsing) {
				addPinnedResult(pdata, true)
			}
			return false, err
		}
		addPinnedResult(pdata, r)
	}

	return true, nil
}

func isDockerfilePinned(c *checker.CheckRequest) (int, error) {
	var r pinnedResult
	err := fileparser.OnMatchingFileContentDo(c.RepoClient, fileparser.PathMatcher{
		Pattern:       "*Dockerfile*",
		CaseSensitive: false,
	}, validateDockerfileIsPinned, c.Dlogger, &r)
	return createReturnForIsDockerfilePinned(r, c.Dlogger, err)
}

// Create the result.
func createReturnForIsDockerfilePinned(r pinnedResult, dl checker.DetailLogger, err error) (int, error) {
	return createReturnValues(r,
		"Dockerfile dependencies are pinned",
		dl, err)
}

var validateDockerfileIsPinned fileparser.DoWhileTrueOnFileContent = func(
	pathfn string,
	content []byte,
	args ...interface{},
) (bool, error) {
	// Users may use various names, e.g.,
	// Dockerfile.aarch64, Dockerfile.template, Dockerfile_template, dockerfile, Dockerfile-name.template

	if len(args) != 2 {
		return false, fmt.Errorf(
			"validateDockerfileIsPinned requires exactly 2 arguments: %w", errInvalidArgLength)
	}
	pdata := dataAsResultPointer(args[1])
	dl := dataAsDetailLogger(args[0])
	// Return early if this is not a dockerfile.
	if !isDockerfile(pathfn, content) {
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
		if cmdType != "FROM" {
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
			if pinned || regex.MatchString(name) {
				// Record the asName.
				pinnedAsNames[asName] = true
				continue
			}

			// Not pinned.
			ret = false
			dl.Warn(&checker.LogMessage{
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
			if !pinned && !regex.MatchString(name) {
				ret = false
				dl.Warn(&checker.LogMessage{
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
	err := fileparser.OnMatchingFileContentDo(c.RepoClient, fileparser.PathMatcher{
		Pattern:       ".github/workflows/*",
		CaseSensitive: false,
	}, validateGitHubWorkflowIsFreeOfInsecureDownloads, c.Dlogger, &r)
	return createReturnForIsGitHubWorkflowScriptFreeOfInsecureDownloads(r, c.Dlogger, err)
}

// Create the result.
func createReturnForIsGitHubWorkflowScriptFreeOfInsecureDownloads(r pinnedResult,
	dl checker.DetailLogger, err error,
) (int, error) {
	return createReturnValues(r,
		"no insecure (not pinned by hash) dependency downloads found in GitHub workflows",
		dl, err)
}

// validateGitHubWorkflowIsFreeOfInsecureDownloads checks if the workflow file downloads dependencies that are unpinned.
// Returns true if the check should continue executing after this file.
var validateGitHubWorkflowIsFreeOfInsecureDownloads fileparser.DoWhileTrueOnFileContent = func(
	pathfn string,
	content []byte,
	args ...interface{},
) (bool, error) {
	if !fileparser.IsWorkflowFile(pathfn) {
		return true, nil
	}

	if len(args) != 2 {
		return false, fmt.Errorf(
			"validateGitHubWorkflowIsFreeOfInsecureDownloads requires exactly 2 arguments: %w", errInvalidArgLength)
	}
	pdata := dataAsResultPointer(args[1])
	dl := dataAsDetailLogger(args[0])

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
				// Ignore parsing errors.
				if errors.Is(err, sce.ErrorShellParsing) {
					addPinnedResult(pdata, true)
				}
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
	err := fileparser.OnMatchingFileContentDo(c.RepoClient, fileparser.PathMatcher{
		Pattern:       ".github/workflows/*",
		CaseSensitive: true,
	}, validateGitHubActionWorkflow, c.Dlogger, &r)
	return createReturnForIsGitHubActionsWorkflowPinned(r, c.Dlogger, err)
}

// Create the result.
func createReturnForIsGitHubActionsWorkflowPinned(r worklowPinningResult, dl checker.DetailLogger,
	err error,
) (int, error) {
	return createReturnValuesForGitHubActionsWorkflowPinned(r,
		"actions are pinned",
		dl, err)
}

func generateOwnerToDisplay(gitHubOwned bool) string {
	if gitHubOwned {
		return "GitHub-owned"
	}
	return "third-party"
}

// validateGitHubActionWorkflow checks if the workflow file contains unpinned actions. Returns true if the check
// should continue executing after this file.
var validateGitHubActionWorkflow fileparser.DoWhileTrueOnFileContent = func(
	pathfn string,
	content []byte,
	args ...interface{},
) (bool, error) {
	if !fileparser.IsWorkflowFile(pathfn) {
		return true, nil
	}

	if len(args) != 2 {
		return false, fmt.Errorf(
			"validateGitHubActionWorkflow requires exactly 2 arguments: %w", errInvalidArgLength)
	}
	pdata := dataAsWorkflowResultPointer(args[1])
	dl := dataAsDetailLogger(args[0])

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
			match := hashRegex.MatchString(execAction.Uses.Value)
			if !match {
				dl.Warn(&checker.LogMessage{
					Path: pathfn, Type: checker.FileTypeSource,
					Offset:      uint(execAction.Uses.Pos.Line),
					EndOffset:   uint(execAction.Uses.Pos.Line), // `Uses` always span a single line.
					Snippet:     execAction.Uses.Value,
					Text:        fmt.Sprintf("%s action not pinned by hash", owner),
					Remediation: remediation.CreateWorkflowPinningRemediation(pathfn),
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
>>>>>>> 465fc76a (updates)
}
