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

	"gopkg.in/yaml.v3"

	"github.com/ossf/scorecard/v3/checker"
	"github.com/ossf/scorecard/v3/checks/fileparser"
	sce "github.com/ossf/scorecard/v3/errors"
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

	var workflow map[interface{}]interface{}
	err := yaml.Unmarshal(content, &workflow)
	if err != nil {
		return false,
			sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("yaml.Unmarshal: %v", err))
	}

	// 1. Check for untrusted code checkout with pull_request_target and a ref
	if err := validateUntrustedCodeCheckout(workflow, path, dl, pdata); err != nil {
		return false, err
	}

	// TODO: Check other dangerous patterns.
	return true, nil
}

func validateUntrustedCodeCheckout(config map[interface{}]interface{}, path string,
	dl checker.DetailLogger, pdata *patternCbData) error {
	checkPullRequestTrigger, err := checkPullRequestTrigger(config)
	if err != nil {
		return err
	}

	if checkPullRequestTrigger {
		return validateUntrustedCodeCheckoutRef(config, path, dl, pdata)
	}

	return nil
}

func validateUntrustedCodeCheckoutRef(config map[interface{}]interface{}, path string,
	dl checker.DetailLogger, pdata *patternCbData) error {
	var jobs interface{}

	// Now check if this is used with untrusted code checkout ref in jobs
	jobs, ok := config["jobs"]
	if !ok {
		return nil
	}

	mjobs, ok := jobs.(map[string]interface{})
	if !ok {
		return sce.WithMessage(sce.ErrScorecardInternal, errInvalidGitHubWorkflow.Error())
	}

	for _, value := range mjobs {
		job, ok := value.(map[string]interface{})
		if !ok {
			return sce.WithMessage(sce.ErrScorecardInternal, errInvalidGitHubWorkflow.Error())
		}

		if err := checkJobForUntrustedCodeCheckout(job, path, dl, pdata); err != nil {
			return err
		}
	}
	return nil
}

func checkPullRequestTrigger(config map[interface{}]interface{}) (bool, error) {
	// Check event trigger (required) is pull_request_target
	trigger, ok := config["on"]
	if !ok {
		return false, sce.WithMessage(sce.ErrScorecardInternal, errInvalidGitHubWorkflow.Error())
	}

	isPullRequestTrigger := false
	switch val := trigger.(type) {
	case string:
		if strings.EqualFold(val, "pull_request_target") {
			isPullRequestTrigger = true
		}
	case []string:
		for _, onVal := range val {
			if strings.EqualFold(onVal, "pull_request_target") {
				isPullRequestTrigger = true
			}
		}
	case map[interface{}]interface{}:
		for k := range val {
			key, ok := k.(string)
			if !ok {
				return false, sce.WithMessage(sce.ErrScorecardInternal, errInvalidGitHubWorkflow.Error())
			}
			if strings.EqualFold(key, "pull_request_target") {
				isPullRequestTrigger = true
			}
		}
	default:
		return false, sce.WithMessage(sce.ErrScorecardInternal, errInvalidGitHubWorkflow.Error())
	}
	return isPullRequestTrigger, nil
}

func checkJobForUntrustedCodeCheckout(job map[string]interface{}, path string,
	dl checker.DetailLogger, pdata *patternCbData) error {
	steps, ok := job["steps"]
	if !ok {
		return nil
	}
	msteps, ok := steps.([]interface{})
	if !ok {
		return sce.WithMessage(sce.ErrScorecardInternal, errInvalidGitHubWorkflow.Error())
	}
	// Check each step, which is a map, for checkouts with untrusted ref
	for _, step := range msteps {
		mstep, ok := step.(map[string]interface{})
		if !ok {
			return sce.WithMessage(sce.ErrScorecardInternal, errInvalidGitHubWorkflow.Error())
		}
		// Check for a step that uses actions/checkout
		uses, ok := mstep["uses"]
		if !ok {
			continue
		}
		muses, ok := uses.(string)
		if !ok {
			return sce.WithMessage(sce.ErrScorecardInternal, errInvalidGitHubWorkflow.Error())
		}
		// Uses defaults if not defined.
		with, ok := mstep["with"]
		if !ok {
			continue
		}
		mwith, ok := with.(map[string]interface{})
		if !ok {
			return sce.WithMessage(sce.ErrScorecardInternal, errInvalidGitHubWorkflow.Error())
		}
		// Check for reference. If not defined for a pull_request_target event, this defaults to
		// the base branch of the pull request.
		ref, ok := mwith["ref"]
		if !ok {
			continue
		}
		mref, ok := ref.(string)
		if !ok {
			return sce.WithMessage(sce.ErrScorecardInternal, errInvalidGitHubWorkflow.Error())
		}
		if strings.Contains(muses, "actions/checkout") &&
			strings.Contains(mref, "github.event.pull_request.head.sha") {
			dl.Warn3(&checker.LogMessage{
				Path: path,
				Type: checker.FileTypeSource,
				// TODO: set line correctly.
				Offset: 1,
				Text:   fmt.Sprintf("untrusted code checkout '%v'", mref),
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

func testValidateGitHubActionDangerousWOrkflow(pathfn string,
	content []byte, dl checker.DetailLogger) checker.CheckResult {
	data := patternCbData{
		workflowPattern: make(map[string]bool),
	}
	_, err := validateGitHubActionWorkflowPatterns(pathfn, content, dl, &data)
	return createResultForDangerousWorkflowPatterns(data, err)
}
