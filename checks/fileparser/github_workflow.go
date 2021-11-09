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

package fileparser

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/rhysd/actionlint"

	sce "github.com/ossf/scorecard/v3/errors"
)

// defaultShellNonWindows is the default shell used for GitHub workflow actions for Linux and Mac.
const defaultShellNonWindows = "bash"

// defaultShellWindows is the default shell used for GitHub workflow actions for Windows.
const defaultShellWindows = "pwsh"

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
	getFromMatrix := len(job.RunsOn.Labels) == 1 && strings.Contains(job.RunsOn.Labels[0].Value, "matrix.os")
	if !getFromMatrix {
		// We can get the OSes straight from 'runs-on'.
		for _, os := range job.RunsOn.Labels {
			jobOSes = append(jobOSes, os.Value)
		}
		return jobOSes, nil
	}

	for rowKey, rowValue := range job.Strategy.Matrix.Rows {
		if rowKey != "os" {
			continue
		}
		for _, os := range rowValue.Values {
			jobOSes = append(jobOSes, strings.Trim(os.String(), "'\""))
		}
	}

	if len(jobOSes) == 0 {
		return jobOSes, sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("unable to determine OS for job: %v", job.Name.Value))
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
		if !strings.HasPrefix(strings.ToLower(os), "windows") {
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
		jobName := ""
		if job.Name != nil {
			jobName = job.Name.Value
		}
		stepName := ""
		if step.Name != nil {
			stepName = step.Name.Value
		}
		return "", sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("unable to parse step '%v' for job '%v'", jobName, stepName))
	}
	if execRun != nil && execRun.Shell != nil && execRun.Shell.Value != "" {
		return execRun.Shell.Value, nil
	}
	if job.Defaults != nil && job.Defaults.Run != nil && job.Defaults.Run.Shell != nil &&
		job.Defaults.Run.Shell.Value != "" {
		return job.Defaults.Run.Shell.Value, nil
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
		return true
	default:
		return false
	}
}

// IsGitHubOwnedAction checks if this is a github specific action.
func IsGitHubOwnedAction(actionName string) bool {
	a := strings.HasPrefix(actionName, "actions/")
	c := strings.HasPrefix(actionName, "github/")
	return a || c
}
