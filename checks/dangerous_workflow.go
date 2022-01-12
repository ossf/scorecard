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

	"github.com/rhysd/actionlint"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/fileparser"
	sce "github.com/ossf/scorecard/v4/errors"
)

// CheckDangerousWorkflow is the exported name for Dangerous-Workflow check.
const CheckDangerousWorkflow = "Dangerous-Workflow"

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

//nolint:gochecknoinits
func init() {
	registerCheck(CheckDangerousWorkflow, DangerousWorkflow)
}

// Holds stateful data to pass thru callbacks.
// Each field correpsonds to a dangerous GitHub workflow pattern, and
// will hold true if the pattern is avoided, false otherwise.
type patternCbData struct {
	workflowPattern map[string]bool
}

// DangerousWorkflow runs Dangerous-Workflow check.
func DangerousWorkflow(c *checker.CheckRequest) checker.CheckResult {
	// data is shared across all GitHub workflows.
	data := patternCbData{
		workflowPattern: make(map[string]bool),
	}
	err := fileparser.CheckFilesContent(".github/workflows/*", false,
		c, validateGitHubActionWorkflowPatterns, &data)
	return createResultForDangerousWorkflowPatterns(data, err)
}

// Check file content.
func validateGitHubActionWorkflowPatterns(path string, content []byte, dl checker.DetailLogger,
	data fileparser.FileCbData) (bool, error) {
	if !fileparser.IsWorkflowFile(path) {
		return true, nil
	}

	// Verify the type of the data.
	pdata, ok := data.(*patternCbData)
	if !ok {
		// This never happens.
		panic("invalid type")
	}

	if !fileparser.CheckFileContainsCommands(content, "#") {
		return true, nil
	}

	workflow, errs := actionlint.Parse(content)
	if len(errs) > 0 && workflow == nil {
		return false, fileparser.FormatActionlintError(errs)
	}

	// 1. Check for untrusted code checkout with pull_request_target and a ref
	if err := validateUntrustedCodeCheckout(workflow, path, dl, pdata); err != nil {
		return false, err
	}

	// 2. Check for script injection in workflow inline scripts.
	if err := validateScriptInjection(workflow, path, dl, pdata); err != nil {
		return false, err
	}

	// TODO: Check other dangerous patterns.
	return true, nil
}

func validateUntrustedCodeCheckout(workflow *actionlint.Workflow, path string,
	dl checker.DetailLogger, pdata *patternCbData) error {
	if checkPullRequestTrigger(workflow) {
		for _, job := range workflow.Jobs {
			if err := checkJobForUntrustedCodeCheckout(job, path, dl, pdata); err != nil {
				return err
			}
		}
	}
	return nil
}

func checkPullRequestTrigger(workflow *actionlint.Workflow) bool {
	// Check if the webhook event trigger is a pull_request_target
	for _, event := range workflow.On {
		e, ok := event.(*actionlint.WebhookEvent)
		if ok && e.Hook != nil && e.Hook.Value == "pull_request_target" {
			return true
		}
	}

	return false
}

func checkJobForUntrustedCodeCheckout(job *actionlint.Job, path string,
	dl checker.DetailLogger, pdata *patternCbData) error {
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
			return nil
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
		if strings.Contains(ref.Value.Value, "github.event.pull_request") {
			line := fileparser.GetLineNumber(step.Pos)
			dl.Warn3(&checker.LogMessage{
				Path:   path,
				Type:   checker.FileTypeSource,
				Offset: line,
				Text:   fmt.Sprintf("untrusted code checkout '%v'", ref.Value.Value),
				// TODO: set Snippet.
			})
			// Detected untrusted checkout.
			pdata.workflowPattern["untrusted_checkout"] = true
		}
	}
	return nil
}

func validateScriptInjection(workflow *actionlint.Workflow, path string,
	dl checker.DetailLogger, pdata *patternCbData) error {
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
			if err := checkVariablesInScript(run.Run.Value, run.Run.Pos, path, dl, pdata); err != nil {
				return err
			}
		}
	}

	return nil
}

func checkVariablesInScript(script string, pos *actionlint.Pos, path string,
	dl checker.DetailLogger, pdata *patternCbData) error {
	for {
		s := strings.Index(script, "${{")
		if s == -1 {
			return nil
		}

		e := strings.Index(script[s:], "}}")
		if e == -1 {
			return sce.WithMessage(sce.ErrScorecardInternal, errInvalidGitHubWorkflow.Error())
		}

		// Check if the variable may be untrustworthy.
		variable := script[s+3 : s+e]
		if containsUntrustedContextPattern(variable) {
			line := fileparser.GetLineNumber(pos)
			dl.Warn3(&checker.LogMessage{
				Path:   path,
				Type:   checker.FileTypeSource,
				Offset: line,
				Text:   fmt.Sprintf("script injection with untrusted input '%v'", variable),
				// TODO: set Snippet.
			})
			pdata.workflowPattern["script_injection"] = true
		}
		script = script[s+e:]
	}
}

// Calculate the workflow score.
func calculateWorkflowScore(result patternCbData) int {
	// Start with a perfect score.
	score := float32(checker.MaxResultScore)

	// pull_request_event indicates untrusted code checkout
	if ok := result.workflowPattern["untrusted_checkout"]; ok {
		score -= 10
	}

	// script injection with an untrusted context
	if ok := result.workflowPattern["script_injection"]; ok {
		score -= 10
	}

	// We're done, calculate the final score.
	if score < checker.MinResultScore {
		return checker.MinResultScore
	}

	return int(score)
}

// Create the result.
func createResultForDangerousWorkflowPatterns(result patternCbData, err error) checker.CheckResult {
	if err != nil {
		return checker.CreateRuntimeErrorResult(CheckDangerousWorkflow, err)
	}

	score := calculateWorkflowScore(result)

	if score != checker.MaxResultScore {
		return checker.CreateResultWithScore(CheckDangerousWorkflow,
			"dangerous workflow patterns detected", score)
	}

	return checker.CreateMaxScoreResult(CheckDangerousWorkflow,
		"no dangerous workflow patterns detected")
}

func testValidateGitHubActionDangerousWorkflow(pathfn string,
	content []byte, dl checker.DetailLogger) checker.CheckResult {
	data := patternCbData{
		workflowPattern: make(map[string]bool),
	}
	_, err := validateGitHubActionWorkflowPatterns(pathfn, content, dl, &data)
	return createResultForDangerousWorkflowPatterns(data, err)
}
