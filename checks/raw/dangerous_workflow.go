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
	triggerPullRequest              = triggerName("pull_request")
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
	}, validateGitHubActionWorkflowPatterns, &data)

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

	// 3. Check for secrets used in workflows triggered by pull requests.
	if err := validateSecretsInPullRequests(workflow, path, pdata); err != nil {
		return false, err
	}

	// TODO: Check other dangerous patterns.
	return true, nil
}

func validateSecretsInPullRequests(workflow *actionlint.Workflow, path string,
	pdata *checker.DangerousWorkflowData,
) error {
	triggers := make(map[triggerName]bool)

	// We need pull request trigger.
	usesPullRequest := usesEventTrigger(workflow, triggerPullRequest)
	usesPullRequestTarget := usesEventTrigger(workflow, triggerPullRequestTarget)
	usesWorkflowRun := usesEventTrigger(workflow, triggerWorkflowRun)

	if !usesPullRequest && !usesPullRequestTarget && !usesWorkflowRun {
		return nil
	}

	// Record the triggers.
	if usesPullRequest {
		triggers[triggerPullRequest] = usesPullRequest
	}
	if usesPullRequestTarget {
		triggers[triggerPullRequestTarget] = usesPullRequestTarget
	}
	if usesWorkflowRun {
		triggers[triggerWorkflowRun] = usesWorkflowRun
	}

	// Secrets used in env at the top of the wokflow.
	if err := checkWorkflowSecretInEnv(workflow, triggers, path, pdata); err != nil {
		return err
	}

	// Secrets used on jobs.
	for _, job := range workflow.Jobs {
		if err := checkJobForUsedSecrets(job, triggers, path, pdata); err != nil {
			return err
		}
	}

	return nil
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

func jobUsesEnvironment(job *actionlint.Job) bool {
	if job.Environment == nil {
		return false
	}

	return job.Environment.Name != nil &&
		job.Environment.Name.Value != ""
}

func checkJobForUsedSecrets(job *actionlint.Job, triggers map[triggerName]bool,
	path string, pdata *checker.DangerousWorkflowData,
) error {
	if job == nil {
		return nil
	}

	// If the job has an environment, assume it's an env secret gated by
	// some approval and don't alert.
	if !jobUsesCodeCheckoutAndNoEnvironment(job, triggers) {
		return nil
	}

	// https://docs.github.com/en/actions/security-guides/encrypted-secrets#naming-your-secrets
	for _, step := range job.Steps {
		if step == nil {
			continue
		}

		if err := checkSecretInActionArgs(step, job, path, pdata); err != nil {
			return err
		}

		if err := checkSecretInRun(step, job, path, pdata); err != nil {
			return err
		}

		if err := checkSecretInEnv(step.Env, job, path, pdata); err != nil {
			return err
		}
	}
	return nil
}

func workflowUsesCodeCheckoutAndNoEnvironment(workflow *actionlint.Workflow,
	triggers map[triggerName]bool,
) bool {
	if workflow == nil {
		return false
	}

	for _, job := range workflow.Jobs {
		if jobUsesCodeCheckoutAndNoEnvironment(job, triggers) {
			return true
		}
	}
	return false
}

func jobUsesCodeCheckoutAndNoEnvironment(job *actionlint.Job, triggers map[triggerName]bool,
) bool {
	if job == nil {
		return false
	}
	_, usesPullRequest := triggers[triggerPullRequest]
	_, usesPullRequestTarget := triggers[triggerPullRequestTarget]
	_, usesWorkflowRun := triggers[triggerWorkflowRun]

	chk, ref := jobUsesCodeCheckout(job)
	if !jobUsesEnvironment(job) {
		if (chk && usesPullRequest) ||
			(chk && usesPullRequestTarget && strings.Contains(ref, checkoutUntrustedPullRequestRef)) ||
			(chk && usesWorkflowRun && strings.Contains(ref, checkoutUntrustedWorkflowRunRef)) {
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
			pdata.UntrustedCheckouts = append(pdata.UntrustedCheckouts,
				checker.UntrustedCheckout{
					File: checker.File{
						Path:    path,
						Type:    checker.FileTypeSource,
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

func checkWorkflowSecretInEnv(workflow *actionlint.Workflow, triggers map[triggerName]bool,
	path string, pdata *checker.DangerousWorkflowData,
) error {
	// We need code checkout and not environment rule protection.
	if !workflowUsesCodeCheckoutAndNoEnvironment(workflow, triggers) {
		return nil
	}

	return checkSecretInEnv(workflow.Env, nil, path, pdata)
}

func checkSecretInEnv(env *actionlint.Env, job *actionlint.Job, path string,
	pdata *checker.DangerousWorkflowData,
) error {
	if env == nil {
		return nil
	}

	for _, v := range env.Vars {
		if err := checkSecretInScript(v.Value.Value, v.Value.Pos, job, path, pdata); err != nil {
			return err
		}
	}
	return nil
}

func checkSecretInRun(step *actionlint.Step, job *actionlint.Job, path string,
	pdata *checker.DangerousWorkflowData,
) error {
	if step == nil || step.Exec == nil {
		return nil
	}

	run, ok := step.Exec.(*actionlint.ExecRun)
	if ok && run.Run != nil {
		if err := checkSecretInScript(run.Run.Value, run.Run.Pos, job, path, pdata); err != nil {
			return err
		}
	}
	return nil
}

func checkSecretInActionArgs(step *actionlint.Step, job *actionlint.Job, path string,
	pdata *checker.DangerousWorkflowData,
) error {
	if step == nil || step.Exec == nil {
		return nil
	}

	e, ok := step.Exec.(*actionlint.ExecAction)
	if ok && e.Uses != nil {
		// Check for reference. If not defined for a pull_request_target event, this defaults to
		// the base branch of the pull request.
		for _, v := range e.Inputs {
			if v.Value != nil {
				if err := checkSecretInScript(v.Value.Value, v.Value.Pos, job, path, pdata); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func checkSecretInScript(script string, pos *actionlint.Pos,
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

		// Note: The default GitHub token is allowed, as it has
		// only read permission for `pull_request`.
		// For `pull_request_event`, we use other signals such as
		// whether checkout action is used.
		variable := strings.Trim(script[s:s+e+2], " ")
		if !strings.Contains(variable, "secrets.GITHUB_TOKEN") &&
			strings.Contains(variable, "secrets.") {
			line := fileparser.GetLineNumber(pos)
			pdata.SecretInPullRequests = append(pdata.SecretInPullRequests,
				checker.EncryptedSecret{
					File: checker.File{
						Path:    path,
						Type:    checker.FileTypeSource,
						Offset:  line,
						Snippet: variable,
					},
					Job: createJob(job),
				},
			)
		}
		script = script[s+e:]
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
			pdata.ScriptInjections = append(pdata.ScriptInjections,
				checker.ScriptInjection{
					File: checker.File{
						Path:    path,
						Type:    checker.FileTypeSource,
						Offset:  line,
						Snippet: variable,
					},
					Job: createJob(job),
				},
			)
		}
		script = script[s+e:]
	}
	return nil
}
