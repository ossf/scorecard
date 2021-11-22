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
	"strings"

	"github.com/rhysd/actionlint"

	"github.com/ossf/scorecard/v3/checker"
	"github.com/ossf/scorecard/v3/checks/fileparser"
)

// CheckDangerousWorkflow is the exported name for Dangerous-Workflow check.
const CheckDangerousWorkflow = "Dangerous-Workflow"

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
		if ok && e.Hook.Value == "pull_request_target" {
			return true
		}
	}

	return false
}

func checkJobForUntrustedCodeCheckout(job *actionlint.Job, path string,
	dl checker.DetailLogger, pdata *patternCbData) error {
	// Check each step, which is a map, for checkouts with untrusted ref
	for _, step := range job.Steps {
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
		if !ok {
			continue
		}
		if strings.Contains(ref.Value.Value, "github.event.pull_request") {
			dl.Warn3(&checker.LogMessage{
				Path:   path,
				Type:   checker.FileTypeSource,
				Offset: step.Pos.Line,
				Text:   fmt.Sprintf("untrusted code checkout '%v'", ref.Value.Value),
				// TODO: set Snippet.
			})
			// Detected untrusted checkout.
			pdata.workflowPattern["untrusted_checkout"] = true
		}
	}
	return nil
}

// Calculate the workflow score.
func calculateWorkflowScore(result patternCbData) int {
	// Start with a perfect score.
	score := float32(checker.MaxResultScore)

	// pull_request_event indicates untrusted code checkout
	if ok := result.workflowPattern["untrusted_checkout"]; ok {
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
