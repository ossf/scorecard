// Copyright 2021 OpenSSF Scorecard Authors
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

package fileparser

import (
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rhysd/actionlint"

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
)

const (
	// defaultShellNonWindows is the default shell used for GitHub workflow actions for Linux and Mac.
	defaultShellNonWindows = "bash"
	// defaultShellWindows is the default shell used for GitHub workflow actions for Windows.
	defaultShellWindows = "pwsh"
	windows             = "windows"
	os                  = "os"
	matrixos            = "matrix.os"
)

// GetJobName returns Name.Value if non-nil, else returns "".
func GetJobName(job *actionlint.Job) string {
	if job != nil && job.Name != nil {
		return job.Name.Value
	}
	return ""
}

// GetStepName returns Name.Value if non-nil, else returns "".
func GetStepName(step *actionlint.Step) string {
	if step != nil && step.Name != nil {
		return step.Name.Value
	}
	return ""
}

// IsStepExecKind compares input `step` ExecKind with `kind` and returns true on a match.
func IsStepExecKind(step *actionlint.Step, kind actionlint.ExecKind) bool {
	if step == nil || step.Exec == nil {
		return false
	}
	return step.Exec.Kind() == kind
}

// GetLineNumber returns the line number for this position.
func GetLineNumber(pos *actionlint.Pos) uint {
	if pos == nil {
		return checker.OffsetDefault
	}
	return uint(pos.Line)
}

// GetUses returns the 'uses' statement in this step or nil if this step does not have one.
func GetUses(step *actionlint.Step) *actionlint.String {
	if step == nil {
		return nil
	}
	if !IsStepExecKind(step, actionlint.ExecKindAction) {
		return nil
	}
	execAction, ok := step.Exec.(*actionlint.ExecAction)
	if !ok || execAction == nil {
		return nil
	}
	return execAction.Uses
}

// getWith returns the 'with' statement in this step or nil if this step does not have one.
func getWith(step *actionlint.Step) map[string]*actionlint.Input {
	if step == nil {
		return nil
	}
	if !IsStepExecKind(step, actionlint.ExecKindAction) {
		return nil
	}
	execAction, ok := step.Exec.(*actionlint.ExecAction)
	if !ok || execAction == nil {
		return nil
	}
	return execAction.Inputs
}

// getRun returns the 'run' statement in this step or nil if this step does not have one.
func getRun(step *actionlint.Step) *actionlint.String {
	if step == nil {
		return nil
	}
	if !IsStepExecKind(step, actionlint.ExecKindRun) {
		return nil
	}
	execAction, ok := step.Exec.(*actionlint.ExecRun)
	if !ok || execAction == nil {
		return nil
	}
	return execAction.Run
}

func getExecRunShell(execRun *actionlint.ExecRun) string {
	if execRun != nil && execRun.Shell != nil {
		return execRun.Shell.Value
	}
	return ""
}

func getJobDefaultRunShell(job *actionlint.Job) string {
	if job != nil && job.Defaults != nil && job.Defaults.Run != nil && job.Defaults.Run.Shell != nil {
		return job.Defaults.Run.Shell.Value
	}
	return ""
}

func getJobRunsOnLabels(job *actionlint.Job) []*actionlint.String {
	if job != nil && job.RunsOn != nil {
		return job.RunsOn.Labels
	}
	return nil
}

func getJobStrategyMatrixRows(job *actionlint.Job) map[string]*actionlint.MatrixRow {
	if job != nil && job.Strategy != nil && job.Strategy.Matrix != nil {
		return job.Strategy.Matrix.Rows
	}
	return nil
}

func getJobStrategyMatrixIncludeCombinations(job *actionlint.Job) []*actionlint.MatrixCombination {
	if job != nil && job.Strategy != nil && job.Strategy.Matrix != nil && job.Strategy.Matrix.Include != nil &&
		job.Strategy.Matrix.Include.Combinations != nil {
		return job.Strategy.Matrix.Include.Combinations
	}
	return nil
}

// FormatActionlintError combines the errors into a single one.
func FormatActionlintError(errs []*actionlint.Error) error {
	if len(errs) == 0 {
		return nil
	}
	builder := strings.Builder{}
	builder.WriteString(errInvalidGitHubWorkflow.Error() + ":")
	for _, err := range errs {
		builder.WriteString("\n" + err.Error())
	}
	return sce.WithMessage(sce.ErrScorecardInternal, builder.String())
}

// GetOSesForJob returns the OSes this job runs on.
func GetOSesForJob(job *actionlint.Job) ([]string, error) {
	// The 'runs-on' field either lists the OS'es directly, or it can have an expression '${{ matrix.os }}' which
	// is where the OS'es are actually listed.
	jobOSes := make([]string, 0)
	jobRunsOnLabels := getJobRunsOnLabels(job)
	getFromMatrix := len(jobRunsOnLabels) == 1 && strings.Contains(jobRunsOnLabels[0].Value, matrixos)
	if !getFromMatrix {
		// We can get the OSes straight from 'runs-on'.
		for _, os := range jobRunsOnLabels {
			jobOSes = append(jobOSes, os.Value)
		}
		return jobOSes, nil
	}

	jobStrategyMatrixRows := getJobStrategyMatrixRows(job)
	for rowKey, rowValue := range jobStrategyMatrixRows {
		if rowKey != os {
			continue
		}
		for _, os := range rowValue.Values {
			jobOSes = append(jobOSes, strings.Trim(os.String(), "'\""))
		}
	}

	matrixCombinations := getJobStrategyMatrixIncludeCombinations(job)
	for _, combination := range matrixCombinations {
		if combination.Assigns == nil {
			continue
		}
		for _, assign := range combination.Assigns {
			if assign.Key == nil || assign.Key.Value != os || assign.Value == nil {
				continue
			}
			jobOSes = append(jobOSes, strings.Trim(assign.Value.String(), "'\""))
		}
	}

	if len(jobOSes) == 0 {
		return jobOSes, sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("unable to determine OS for job: %v", GetJobName(job)))
	}
	return jobOSes, nil
}

// JobAlwaysRunsOnWindows returns true if the only OS that this job runs on is Windows.
func JobAlwaysRunsOnWindows(job *actionlint.Job) (bool, error) {
	jobOSes, err := GetOSesForJob(job)
	if err != nil {
		return false, err
	}
	for _, os := range jobOSes {
		if !strings.HasPrefix(strings.ToLower(os), windows) {
			return false, nil
		}
	}
	return true, nil
}

// GetShellForStep returns the shell that is used to run the given step.
func GetShellForStep(step *actionlint.Step, job *actionlint.Job) (string, error) {
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#using-a-specific-shell.
	execRun, ok := step.Exec.(*actionlint.ExecRun)
	if !ok {
		jobName := GetJobName(job)
		stepName := GetStepName(step)
		return "", sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("unable to parse step '%v' for job '%v'", jobName, stepName))
	}
	execRunShell := getExecRunShell(execRun)
	if execRunShell != "" {
		return execRun.Shell.Value, nil
	}
	jobDefaultRunShell := getJobDefaultRunShell(job)
	if jobDefaultRunShell != "" {
		return jobDefaultRunShell, nil
	}

	isStepWindows, err := IsStepWindows(step)
	if err != nil {
		return "", err
	}
	if isStepWindows {
		return defaultShellWindows, nil
	}

	alwaysRunsOnWindows, err := JobAlwaysRunsOnWindows(job)
	if err != nil {
		return "", err
	}
	if alwaysRunsOnWindows {
		return defaultShellWindows, nil
	}

	return defaultShellNonWindows, nil
}

// IsStepWindows returns true if the step will be run on Windows.
func IsStepWindows(step *actionlint.Step) (bool, error) {
	if step.If == nil {
		return false, nil
	}
	windowsRegexes := []string{
		// Looking for "if: runner.os == 'Windows'" (and variants)
		`(?i)runner\.os\s*==\s*['"]windows['"]`,
		// Looking for "if: ${{ startsWith(runner.os, 'Windows') }}" (and variants)
		`(?i)\$\{\{\s*startsWith\(runner\.os,\s*['"]windows['"]\)`,
		// Looking for "if: matrix.os == 'windows-2019'" (and variants)
		`(?i)matrix\.os\s*==\s*['"]windows-`,
	}

	for _, windowsRegex := range windowsRegexes {
		matches, err := regexp.MatchString(windowsRegex, step.If.Value)
		if err != nil {
			return false, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("error matching Windows regex: %v", err))
		}
		if matches {
			return true, nil
		}
	}

	return false, nil
}

// IsWorkflowFile returns true if this is a GitHub workflow file.
func IsWorkflowFile(pathfn string) bool {
	// From https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions:
	// "Workflow files use YAML syntax, and must have either a .yml or .yaml file extension."
	switch path.Ext(pathfn) {
	case ".yml", ".yaml":
		return filepath.Dir(strings.ToLower(pathfn)) == ".github/workflows"
	default:
		return false
	}
}

// IsGithubWorkflowFileCb determines if a file is a workflow
// as a callback to use for repo client's ListFiles() API.
func IsGithubWorkflowFileCb(pathfn string) (bool, error) {
	return IsWorkflowFile(pathfn), nil
}

// IsGitHubOwnedAction checks if this is a github specific action.
func IsGitHubOwnedAction(actionName string) bool {
	a := strings.HasPrefix(actionName, "actions/")
	c := strings.HasPrefix(actionName, "github/")
	return a || c
}

// JobMatcher is rule for matching a job.
type JobMatcher struct {
	// The text to be logged when a job match is found.
	LogText string
	// Each step in this field has a matching step in the job.
	Steps []*JobMatcherStep
}

// JobMatcherStep is a single step that needs to be matched.
type JobMatcherStep struct {
	// If set, the step's 'Uses' must match this field. Checks that the action name is the same.
	Uses string
	// If set, the step's 'With' have the keys and values that are in this field.
	With map[string]string
	// If set, the step's 'Run' must match this field. Does a regex match using this field.
	Run string
}

// JobMatchResult represents the result of a matche.
type JobMatchResult struct {
	Msg  string
	File checker.File
}

// AnyJobsMatch returns true if any of the jobs have a match in the given workflow.
func AnyJobsMatch(workflow *actionlint.Workflow, jobMatchers []JobMatcher, fp string,
	logMsgNoMatch string,
) (JobMatchResult, bool) {
	for _, job := range workflow.Jobs {
		for _, matcher := range jobMatchers {
			if !matcher.matches(job) {
				continue
			}

			return JobMatchResult{
				File: checker.File{
					Path:   fp,
					Type:   finding.FileTypeSource,
					Offset: GetLineNumber(job.Pos),
				},
				Msg: fmt.Sprintf("%v: %v", matcher.LogText, fp),
			}, true
		}
	}

	return JobMatchResult{
		File: checker.File{
			Path:   fp,
			Type:   finding.FileTypeSource,
			Offset: checker.OffsetDefault,
		},
		Msg: fmt.Sprintf("%v: %v", logMsgNoMatch, fp),
	}, false
}

// matches returns true if the job matches the job matcher.
func (m *JobMatcher) matches(job *actionlint.Job) bool {
	for _, stepToMatch := range m.Steps {
		hasMatch := false

		// First look for re-usable workflow calls.
		if job.WorkflowCall != nil &&
			job.WorkflowCall.Uses != nil &&
			strings.HasPrefix(job.WorkflowCall.Uses.Value, stepToMatch.Uses+"@") {
			return true
		}

		// Second looks for steps in the job.
		for _, step := range job.Steps {
			if stepsMatch(stepToMatch, step) {
				hasMatch = true
				break
			}
		}
		if !hasMatch {
			return false
		}
	}
	return true
}

// stepsMatch returns true if the fields on 'stepToMatch' match what's in 'step'.
func stepsMatch(stepToMatch *JobMatcherStep, step *actionlint.Step) bool {
	// Make sure 'uses' matches if present.
	if stepToMatch.Uses != "" {
		uses := GetUses(step)
		if uses == nil {
			return false
		}
		if !strings.HasPrefix(uses.Value, stepToMatch.Uses+"@") {
			return false
		}
	}

	// Make sure 'with' matches if present.
	if len(stepToMatch.With) > 0 {
		with := getWith(step)
		if with == nil {
			return false
		}
		for keyToMatch, valToMatch := range stepToMatch.With {
			input, ok := with[keyToMatch]
			if !ok || input == nil || input.Value == nil || input.Value.Value != valToMatch {
				return false
			}
		}
	}

	// Make sure 'run' matches if present.
	if stepToMatch.Run != "" {
		run := getRun(step)
		if run == nil {
			return false
		}
		withoutLineContinuations := regexp.MustCompile("\\\\(\n|\r|\r\n)").ReplaceAllString(run.Value, "")
		r := regexp.MustCompile(stepToMatch.Run)
		if !r.MatchString(withoutLineContinuations) {
			return false
		}
	}

	return true
}

// IsPackagingWorkflow checks for a packaging workflow.
func IsPackagingWorkflow(workflow *actionlint.Workflow, fp string) (JobMatchResult, bool) {
	jobMatchers := []JobMatcher{
		{
			Steps: []*JobMatcherStep{
				{
					Uses: "actions/setup-node",
					With: map[string]string{"registry-url": "https://registry.npmjs.org"},
				},
				{
					Run: "npm.*publish",
				},
			},
			LogText: "candidate node publishing workflow using npm",
		},
		{
			// Java packages with maven.
			Steps: []*JobMatcherStep{
				{
					Uses: "actions/setup-java",
				},
				{
					Run: "mvn.*deploy",
				},
			},
			LogText: "candidate java publishing workflow using maven",
		},
		{
			// Java packages with gradle.
			Steps: []*JobMatcherStep{
				{
					Uses: "actions/setup-java",
				},
				{
					Run: "gradle.*publish",
				},
			},
			LogText: "candidate java publishing workflow using gradle",
		},
		{
			// Ruby packages.
			Steps: []*JobMatcherStep{
				{
					Run: "gem.*push",
				},
			},
			LogText: "candidate ruby publishing workflow using gem",
		},
		{
			// NuGet packages.
			Steps: []*JobMatcherStep{
				{
					Run: "nuget.*push",
				},
			},
			LogText: "candidate nuget publishing workflow",
		},
		{
			// Docker packages.
			Steps: []*JobMatcherStep{
				{
					Run: "docker.*push",
				},
			},
			LogText: "candidate docker publishing workflow",
		},
		{
			// Docker packages.
			Steps: []*JobMatcherStep{
				{
					Uses: "docker/build-push-action",
				},
			},
			LogText: "candidate docker publishing workflow",
		},
		{
			// Python packages.
			Steps: []*JobMatcherStep{
				{
					Uses: "pypa/gh-action-pypi-publish",
				},
			},
			LogText: "candidate python publishing workflow using pypi",
		},
		{
			// Python packages.
			// This is a custom Python packaging workflow based on semantic versioning.
			// TODO(#1642): accept custom workflows through a separate configuration.
			Steps: []*JobMatcherStep{
				{
					Uses: "relekang/python-semantic-release",
				},
			},
			LogText: "candidate python publishing workflow using python-semantic-release",
		},
		{
			// Go packages.
			Steps: []*JobMatcherStep{
				{
					Uses: "actions/setup-go",
				},
				{
					Uses: "goreleaser/goreleaser-action",
				},
			},
			LogText: "candidate golang publishing workflow",
		},
		{
			// Rust packages. https://doc.rust-lang.org/cargo/reference/publishing.html
			Steps: []*JobMatcherStep{
				{
					Run: "cargo.*publish",
				},
			},
			LogText: "candidate rust publishing workflow using cargo",
		},
		{
			// Ko container action. https://github.com/google/ko
			Steps: []*JobMatcherStep{
				{
					Uses: "imjasonh/setup-ko",
				},
				{
					Uses: "ko-build/setup-ko",
				},
			},
			LogText: "candidate container publishing workflow using ko",
		},
	}

	return AnyJobsMatch(workflow, jobMatchers, fp, "not a publishing workflow")
}
