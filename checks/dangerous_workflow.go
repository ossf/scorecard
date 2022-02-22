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
	supportedRequestTypes := []checker.RequestType{
		checker.FileBased,
		checker.CommitBased,
	}
	if err := registerCheck(CheckDangerousWorkflow, DangerousWorkflow, supportedRequestTypes); err != nil {
		// this should never happen
		panic(err)
	}
}

type dangerousResults int

const (
	scriptInjection dangerousResults = iota
	untrustedCheckout
	secretsViaPullRequests
)

type triggerName string

var (
	triggerPullRequestTarget = triggerName("pull_request_target")
	triggerPullRequest       = triggerName("pull_request")
	checkoutUntrustedRef     = "github.event.pull_request"
)

// Holds stateful data to pass thru callbacks.
// Each field correpsonds to a dangerous GitHub workflow pattern, and
// will hold true if the pattern is avoided, false otherwise.
type patternCbData struct {
	workflowPattern map[dangerousResults]bool
}

// DangerousWorkflow runs Dangerous-Workflow check.
func DangerousWorkflow(c *checker.CheckRequest) checker.CheckResult {
	// data is shared across all GitHub workflows.
	data := patternCbData{
		workflowPattern: make(map[dangerousResults]bool),
	}
	err := fileparser.OnMatchingFileContentDo(c.RepoClient, fileparser.PathMatcher{
		Pattern:       ".github/workflows/*",
		CaseSensitive: false,
	},
		validateGitHubActionWorkflowPatterns, c.Dlogger, &data)
	return createResultForDangerousWorkflowPatterns(data, err)
}

// Check file content.
var validateGitHubActionWorkflowPatterns fileparser.DoWhileTrueOnFileContent = func(path string,
	content []byte,
	args ...interface{}) (bool, error) {
	if !fileparser.IsWorkflowFile(path) {
		return true, nil
	}

	if len(args) != 2 {
		return false, fmt.Errorf(
			"validateGitHubActionWorkflowPatterns requires exactly 2 arguments: %w", errInvalidArgLength)
	}

	// Verify the type of the data.
	pdata, ok := args[1].(*patternCbData)
	if !ok {
		return false, fmt.Errorf(
			"validateGitHubActionWorkflowPatterns expects arg[0] of type *patternCbData: %w", errInvalidArgType)
	}
	dl, ok := args[0].(checker.DetailLogger)
	if !ok {
		return false, fmt.Errorf(
			"validateGitHubActionWorkflowPatterns expects arg[1] of type checker.DetailLogger: %w", errInvalidArgType)
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

	// 3. Check for secrets used in workflows triggered by pull requests.
	if err := validateSecretsInPullRequests(workflow, path, dl, pdata); err != nil {
		return false, err
	}

	// TODO: Check other dangerous patterns.
	return true, nil
}

func validateSecretsInPullRequests(workflow *actionlint.Workflow, path string,
	dl checker.DetailLogger, pdata *patternCbData) error {
	triggers := make(map[triggerName]bool)

	// We need pull request trigger.
	usesPullRequest := usesEventTrigger(workflow, triggerPullRequest)
	usesPullRequestTarget := usesEventTrigger(workflow, triggerPullRequestTarget)
	if !usesPullRequest && !usesPullRequestTarget {
		return nil
	}

	// Record the triggers.
	if usesPullRequest {
		triggers[triggerPullRequest] = usesPullRequest
	}
	if usesPullRequestTarget {
		triggers[triggerPullRequestTarget] = usesPullRequestTarget
	}

	// Secrets used in env at the top of the wokflow.
	if err := checkWorkflowSecretInEnv(workflow, triggers, path, dl, pdata); err != nil {
		return err
	}

	// Secrets used on jobs.
	for _, job := range workflow.Jobs {
		if err := checkJobForUsedSecrets(job, triggers, path, dl, pdata); err != nil {
			return err
		}
	}

	return nil
}

func validateUntrustedCodeCheckout(workflow *actionlint.Workflow, path string,
	dl checker.DetailLogger, pdata *patternCbData) error {
	if !usesEventTrigger(workflow, triggerPullRequestTarget) {
		return nil
	}

	for _, job := range workflow.Jobs {
		if err := checkJobForUntrustedCodeCheckout(job, path, dl, pdata); err != nil {
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

func jobUsesEnvironment(job *actionlint.Job) bool {
	if job.Environment == nil {
		return false
	}

	return job.Environment.Name != nil &&
		job.Environment.Name.Value != ""
}

func checkJobForUsedSecrets(job *actionlint.Job, triggers map[triggerName]bool,
	path string, dl checker.DetailLogger, pdata *patternCbData) error {
	if job == nil {
		return nil
	}

	// If the job has an environment, assume it's an env secret gated by
	// some approval and don't alert.
	if jobUsesEnvironment(job) {
		return nil
	}

	// For pull request target, we need a ref to the pull request.
	_, usesPullRequest := triggers[triggerPullRequest]
	_, usesPullRequestTarget := triggers[triggerPullRequestTarget]
	chk, ref := jobUsesCodeCheckout(job)
	if !((chk && usesPullRequest) ||
		(chk && usesPullRequestTarget && strings.Contains(ref, checkoutUntrustedRef))) {
		return nil
	}

	// https://docs.github.com/en/actions/security-guides/encrypted-secrets#naming-your-secrets
	for _, step := range job.Steps {
		if step == nil {
			continue
		}

		if err := checkSecretInActionArgs(step, path, dl, pdata); err != nil {
			return err
		}

		if err := checkSecretInRun(step, path, dl, pdata); err != nil {
			return err
		}

		if err := checkSecretInEnv(step.Env, path, dl, pdata); err != nil {
			return err
		}
	}
	return nil
}

func workflowUsesCodeCheckoutAndNoEnvironment(workflow *actionlint.Workflow,
	triggers map[triggerName]bool) bool {
	if workflow == nil {
		return false
	}

	_, usesPullRequest := triggers[triggerPullRequest]
	_, usesPullRequestTarget := triggers[triggerPullRequestTarget]

	for _, job := range workflow.Jobs {
		chk, ref := jobUsesCodeCheckout(job)
		if ((chk && usesPullRequest) ||
			(chk && usesPullRequestTarget && strings.Contains(ref, checkoutUntrustedRef))) &&
			!jobUsesEnvironment(job) {
			return true
		}
	}
	return false
}

func jobUsesCodeCheckout(job *actionlint.Job) (bool, string) {
	if job == nil {
		return false, ""
	}

	hasCheckout := false
	for _, step := range job.Steps {
		if step == nil || step.Exec == nil {
			continue
		}
		// Check for a step that uses actions/checkout
		e, ok := step.Exec.(*actionlint.ExecAction)
		if !ok || e.Uses == nil {
			continue
		}
		if strings.Contains(e.Uses.Value, "actions/checkout") {
			hasCheckout = true
			ref, ok := e.Inputs["ref"]
			if !ok || ref.Value == nil {
				continue
			}
			return true, ref.Value.Value
		}
	}
	return hasCheckout, ""
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
		if strings.Contains(ref.Value.Value, checkoutUntrustedRef) {
			line := fileparser.GetLineNumber(step.Pos)
			dl.Warn(&checker.LogMessage{
				Path:   path,
				Type:   checker.FileTypeSource,
				Offset: line,
				Text:   fmt.Sprintf("untrusted code checkout '%v'", ref.Value.Value),
				// TODO: set Snippet.
			})
			// Detected untrusted checkout.
			pdata.workflowPattern[untrustedCheckout] = true
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

func checkWorkflowSecretInEnv(workflow *actionlint.Workflow, triggers map[triggerName]bool,
	path string, dl checker.DetailLogger, pdata *patternCbData) error {
	// We need code checkout and not environment rule protection.
	if !workflowUsesCodeCheckoutAndNoEnvironment(workflow, triggers) {
		return nil
	}

	return checkSecretInEnv(workflow.Env, path, dl, pdata)
}

func checkSecretInEnv(env *actionlint.Env, path string,
	dl checker.DetailLogger, pdata *patternCbData) error {
	if env == nil {
		return nil
	}

	for _, v := range env.Vars {
		if err := checkSecretInScript(v.Value.Value, v.Value.Pos, path, dl, pdata); err != nil {
			return err
		}
	}
	return nil
}

func checkSecretInRun(step *actionlint.Step, path string,
	dl checker.DetailLogger, pdata *patternCbData) error {
	if step == nil || step.Exec == nil {
		return nil
	}

	run, ok := step.Exec.(*actionlint.ExecRun)
	if ok && run.Run != nil {
		if err := checkSecretInScript(run.Run.Value, run.Run.Pos, path, dl, pdata); err != nil {
			return err
		}
	}
	return nil
}

func checkSecretInActionArgs(step *actionlint.Step, path string,
	dl checker.DetailLogger, pdata *patternCbData) error {
	if step == nil || step.Exec == nil {
		return nil
	}

	e, ok := step.Exec.(*actionlint.ExecAction)
	if ok && e.Uses != nil {
		// Check for reference. If not defined for a pull_request_target event, this defaults to
		// the base branch of the pull request.
		for _, v := range e.Inputs {
			if v.Value != nil {
				if err := checkSecretInScript(v.Value.Value, v.Value.Pos, path, dl, pdata); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func checkSecretInScript(script string, pos *actionlint.Pos, path string,
	dl checker.DetailLogger, pdata *patternCbData) error {
	for {
		s := strings.Index(script, "${{")
		if s == -1 {
			break
		}

		e := strings.Index(script[s:], "}}")
		if e == -1 {
			return sce.WithMessage(sce.ErrScorecardInternal, errInvalidGitHubWorkflow.Error())
		}

		variable := strings.Trim(script[s:s+e+2], " ")
		if strings.Contains(variable, "secrets.") {
			line := fileparser.GetLineNumber(pos)
			dl.Warn(&checker.LogMessage{
				Path:   path,
				Type:   checker.FileTypeSource,
				Offset: line,
				Text:   fmt.Sprintf("secret accessible to pull requests '%v'", variable),
				// TODO: set Snippet.
			})
			pdata.workflowPattern[secretsViaPullRequests] = true
		}
		script = script[s+e:]
	}
	return nil
}

func checkVariablesInScript(script string, pos *actionlint.Pos, path string,
	dl checker.DetailLogger, pdata *patternCbData) error {
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
			dl.Warn(&checker.LogMessage{
				Path:   path,
				Type:   checker.FileTypeSource,
				Offset: line,
				Text:   fmt.Sprintf("script injection with untrusted input '%v'", variable),
				// TODO: set Snippet.
			})
			pdata.workflowPattern[scriptInjection] = true
		}
		script = script[s+e:]
	}
	return nil
}

// Calculate the workflow score.
func calculateWorkflowScore(result patternCbData) int {
	// Start with a perfect score.
	score := float32(checker.MaxResultScore)

	// Pull_request_event indicates untrusted code checkout.
	if ok := result.workflowPattern[untrustedCheckout]; ok {
		score -= 10
	}

	// Script injection with an untrusted context.
	if ok := result.workflowPattern[scriptInjection]; ok {
		score -= 10
	}

	// Secrets available by pull requests.
	if ok := result.workflowPattern[secretsViaPullRequests]; ok {
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
		workflowPattern: make(map[dangerousResults]bool),
	}
	_, err := validateGitHubActionWorkflowPatterns(pathfn, content, dl, &data)
	return createResultForDangerousWorkflowPatterns(data, err)
}
