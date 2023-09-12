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

package raw

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/rhysd/actionlint"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/fileparser"
	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
)

func containsUntrustedContextPattern(variable string) bool {
	// GitHub event context details that may be attacker controlled.
	// See https://securitylab.github.com/research/github-actions-untrusted-input/
	untrustedContextPattern := regexp.MustCompile(
		`.*(issue\.title|` +
			`issue\.body|` +
			`pull_request\.title|` +
			`pull_request\.body|` +
			`comment\.body|` +
			`review\.body|` +
			`review_comment\.body|` +
			`pages.*\.page_name|` +
			`commits.*\.message|` +
			`head_commit\.message|` +
			`head_commit\.author\.email|` +
			`head_commit\.author\.name|` +
			`commits.*\.author\.email|` +
			`commits.*\.author\.name|` +
			`pull_request\.head\.ref|` +
			`pull_request\.head\.label|` +
			`pull_request\.head\.repo\.default_branch).*`)

	if strings.Contains(variable, "github.head_ref") {
		return true
	}
	return strings.Contains(variable, "github.event.") && untrustedContextPattern.MatchString(variable)
}

type triggerName string

var (
	triggerPullRequestTarget        = triggerName("pull_request_target")
	triggerWorkflowRun              = triggerName("workflow_run")
	checkoutUntrustedPullRequestRef = "github.event.pull_request"
	checkoutUntrustedWorkflowRunRef = "github.event.workflow_run"
)

// DangerousWorkflow retrieves the raw data for the DangerousWorkflow check.
func DangerousWorkflow(c clients.RepoClient) (checker.DangerousWorkflowData, error) {
	// data is shared across all GitHub workflows.
	var data checker.DangerousWorkflowData
	err := fileparser.OnMatchingFileContentDo(c, fileparser.PathMatcher{
		Pattern:       ".github/workflows/*",
		CaseSensitive: false,
	}, validateGitHubActionWorkflowPatterns, true, &data)

	return data, err
}

// Check file content.
var validateGitHubActionWorkflowPatterns fileparser.DoWhileTrueOnFileContent = func(path string,
	content []byte,
	args ...interface{},
) (bool, error) {
	if !fileparser.IsWorkflowFile(path) {
		return true, nil
	}

	if len(args) != 1 {
		return false, fmt.Errorf(
			"validateGitHubActionWorkflowPatterns requires exactly 2 arguments: %w", errInvalidArgLength)
	}

	// Verify the type of the data.
	pdata, ok := args[0].(*checker.DangerousWorkflowData)
	if !ok {
		return false, fmt.Errorf(
			"validateGitHubActionWorkflowPatterns expects arg[0] of type *patternCbData: %w", errInvalidArgType)
	}

	if !fileparser.CheckFileContainsCommands(content, "#") {
		return true, nil
	}

	pdata.NumWorkflows += 1

	workflow, errs := actionlint.Parse(content)
	if len(errs) > 0 && workflow == nil {
		return false, fileparser.FormatActionlintError(errs)
	}

	// 1. Check for untrusted code checkout with pull_request_target and a ref
	if err := validateUntrustedCodeCheckout(workflow, path, pdata); err != nil {
		return false, err
	}

	// 2. Check for script injection in workflow inline scripts.
	if err := validateScriptInjection(workflow, path, pdata); err != nil {
		return false, err
	}

	// TODO: Check other dangerous patterns.
	return true, nil
}

func validateUntrustedCodeCheckout(workflow *actionlint.Workflow, path string,
	pdata *checker.DangerousWorkflowData,
) error {
	if !usesEventTrigger(workflow, triggerPullRequestTarget) && !usesEventTrigger(workflow, triggerWorkflowRun) {
		return nil
	}

	for _, job := range workflow.Jobs {
		if err := checkJobForUntrustedCodeCheckout(job, path, pdata); err != nil {
			return err
		}
	}

	return nil
}

func usesEventTrigger(workflow *actionlint.Workflow, name triggerName) bool {
	// Check if the webhook event trigger is a pull_request_target
	for _, event := range workflow.On {
		if event.EventName() == string(name) {
			return true
		}
	}

	return false
}

func createJob(job *actionlint.Job) *checker.WorkflowJob {
	if job == nil {
		return nil
	}
	var r checker.WorkflowJob
	if job.Name != nil {
		r.Name = &job.Name.Value
	}
	if job.ID != nil {
		r.ID = &job.ID.Value
	}
	return &r
}

func checkJobForUntrustedCodeCheckout(job *actionlint.Job, path string,
	pdata *checker.DangerousWorkflowData,
) error {
	if job == nil {
		return nil
	}

	// Check each step, which is a map, for checkouts with untrusted ref
	for _, step := range job.Steps {
		if step == nil || step.Exec == nil {
			continue
		}
		// Check for a step that uses actions/checkout
		e, ok := step.Exec.(*actionlint.ExecAction)
		if !ok || e.Uses == nil {
			continue
		}
		if !strings.Contains(e.Uses.Value, "actions/checkout") {
			continue
		}
		// Check for reference. If not defined for a pull_request_target event, this defaults to
		// the base branch of the pull request.
		ref, ok := e.Inputs["ref"]
		if !ok || ref.Value == nil {
			continue
		}

		if strings.Contains(ref.Value.Value, checkoutUntrustedPullRequestRef) ||
			strings.Contains(ref.Value.Value, checkoutUntrustedWorkflowRunRef) {
			line := fileparser.GetLineNumber(step.Pos)
			pdata.Workflows = append(pdata.Workflows,
				checker.DangerousWorkflow{
					Type: checker.DangerousWorkflowUntrustedCheckout,
					File: checker.File{
						Path:    path,
						Type:    finding.FileTypeSource,
						Offset:  line,
						Snippet: ref.Value.Value,
					},
					Job: createJob(job),
				},
			)
		}
	}
	return nil
}

func validateScriptInjection(workflow *actionlint.Workflow, path string,
	pdata *checker.DangerousWorkflowData,
) error {
	for _, job := range workflow.Jobs {
		if job == nil {
			continue
		}
		for _, step := range job.Steps {
			if step == nil {
				continue
			}
			run, ok := step.Exec.(*actionlint.ExecRun)
			if !ok || run.Run == nil {
				continue
			}
			// Check Run *String for user-controllable (untrustworthy) properties.
			if err := checkVariablesInScript(run.Run.Value, run.Run.Pos, job, path, pdata); err != nil {
				return err
			}
		}
	}
	return nil
}

func checkVariablesInScript(script string, pos *actionlint.Pos,
	job *actionlint.Job, path string,
	pdata *checker.DangerousWorkflowData,
) error {
	for {
		s := strings.Index(script, "${{")
		if s == -1 {
			break
		}

		e := strings.Index(script[s:], "}}")
		if e == -1 {
			return sce.WithMessage(sce.ErrScorecardInternal, errInvalidGitHubWorkflow.Error())
		}

		// Check if the variable may be untrustworthy.
		variable := script[s+3 : s+e]
		if containsUntrustedContextPattern(variable) {
			line := fileparser.GetLineNumber(pos)
			pdata.Workflows = append(pdata.Workflows,
				checker.DangerousWorkflow{
					File: checker.File{
						Path:    path,
						Type:    finding.FileTypeSource,
						Offset:  line,
						Snippet: variable,
					},
					Job:  createJob(job),
					Type: checker.DangerousWorkflowScriptInjection,
				},
			)
		}
		script = script[s+e:]
	}
	return nil
}
