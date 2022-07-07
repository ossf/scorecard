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

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/fileparser"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/remediation"
)

// CheckTokenPermissions is the exported name for Token-Permissions check.
const (
	CheckTokenPermissions = "Token-Permissions"
	jobLevelPermission    = "job level"
	topLevelPermission    = "top level"
)

type permission string

const (
	permissionAll            = permission("all")
	permissionStatuses       = permission("statuses")
	permissionChecks         = permission("checks")
	permissionSecurityEvents = permission("security-events")
	permissionDeployments    = permission("deployments")
	permissionContents       = permission("contents")
	permissionPackages       = permission("packages")
	permissionActions        = permission("actions")
)

var permissionsOfInterest = []permission{
	permissionStatuses, permissionChecks,
	permissionSecurityEvents, permissionDeployments,
	permissionContents, permissionPackages, permissionActions,
}

//nolint:gochecknoinits
func init() {
	supportedRequestTypes := []checker.RequestType{
		checker.FileBased,
		checker.CommitBased,
	}
	if err := registerCheck(CheckTokenPermissions, TokenPermissions, supportedRequestTypes); err != nil {
		// This should never happen.
		panic(err)
	}
}

// Holds stateful data to pass thru callbacks.
// Each field correpsonds to a GitHub permission type, and
// will hold true if declared non-write, false otherwise.
type permissions struct {
	topLevelWritePermissions map[permission]bool
	jobLevelWritePermissions map[permission]bool
}

type permissionCbData struct {
	// map of filename to write permissions used.
	workflows map[string]permissions
}

// TokenPermissions runs Token-Permissions check.
func TokenPermissions(c *checker.CheckRequest) checker.CheckResult {
	// data is shared across all GitHub workflows.
	data := permissionCbData{
		workflows: make(map[string]permissions),
	}

	if err := remediation.Setup(c); err != nil {
		createResultForLeastPrivilegeTokens(data, err)
	}

	err := fileparser.OnMatchingFileContentDo(c.RepoClient, fileparser.PathMatcher{
		Pattern:       ".github/workflows/*",
		CaseSensitive: false,
	}, validateGitHubActionTokenPermissions, c.Dlogger, &data)
	return createResultForLeastPrivilegeTokens(data, err)
}

// Check file content.
var validateGitHubActionTokenPermissions fileparser.DoWhileTrueOnFileContent = func(path string,
	content []byte,
	args ...interface{},
) (bool, error) {
	if !fileparser.IsWorkflowFile(path) {
		return true, nil
	}
	// Verify the type of the data.
	if len(args) != 2 {
		return false, fmt.Errorf(
			"validateGitHubActionTokenPermissions requires exactly 2 arguments: %w", errInvalidArgLength)
	}
	pdata, ok := args[1].(*permissionCbData)
	if !ok {
		return false, fmt.Errorf(
			"validateGitHubActionTokenPermissions requires arg[0] of type *permissionCbData: %w", errInvalidArgType)
	}
	dl, ok := args[0].(checker.DetailLogger)
	if !ok {
		return false, fmt.Errorf(
			"validateGitHubActionTokenPermissions requires arg[1] of type checker.DetailLogger: %w", errInvalidArgType)
	}

	if !fileparser.CheckFileContainsCommands(content, "#") {
		return true, nil
	}

	workflow, errs := actionlint.Parse(content)
	if len(errs) > 0 && workflow == nil {
		return false, fileparser.FormatActionlintError(errs)
	}

	// 1. Top-level permission definitions.
	//nolint
	// https://docs.github.com/en/actions/reference/authentication-in-a-workflow#example-1-passing-the-github_token-as-an-input,
	// https://github.blog/changelog/2021-04-20-github-actions-control-permissions-for-github_token/,
	// https://docs.github.com/en/actions/reference/authentication-in-a-workflow#modifying-the-permissions-for-the-github_token.
	if err := validateTopLevelPermissions(workflow, path, dl, pdata); err != nil {
		return false, err
	}

	// 2. Run-level permission definitions,
	// see https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idpermissions.
	ignoredPermissions := createIgnoredPermissions(workflow, path, dl)
	if err := validatejobLevelPermissions(workflow, path, dl, pdata, ignoredPermissions); err != nil {
		return false, err
	}

	// TODO(laurent): 2. Identify github actions that require write and add checks.

	// TODO(laurent): 3. Read a few runs and ensures they have the same permissions.

	return true, nil
}

func validatePermission(permissionKey permission, permissionValue *actionlint.PermissionScope,
	permLevel, path string, dl checker.DetailLogger, pPermissions map[permission]bool,
	ignoredPermissions map[permission]bool,
) error {
	if permissionValue.Value == nil {
		return sce.WithMessage(sce.ErrScorecardInternal, errInvalidGitHubWorkflow.Error())
	}
	val := permissionValue.Value.Value
	lineNumber := fileparser.GetLineNumber(permissionValue.Value.Pos)
	if strings.EqualFold(val, "write") {
		if isPermissionOfInterest(permissionKey, ignoredPermissions) {
			dl.Warn(&checker.LogMessage{
				Path:        path,
				Type:        checker.FileTypeSource,
				Offset:      lineNumber,
				Text:        fmt.Sprintf("%s '%v' permission set to '%v'", permLevel, permissionKey, val),
				Snippet:     val,
				Remediation: remediation.CreateWorkflowPermissionRemediation(path),
			})
			recordPermissionWrite(pPermissions, permissionKey)
		} else {
			// Only log for debugging, otherwise
			// it may confuse users.
			dl.Debug(&checker.LogMessage{
				Path:        path,
				Type:        checker.FileTypeSource,
				Offset:      lineNumber,
				Text:        fmt.Sprintf("%s '%v' permission set to '%v'", permLevel, permissionKey, val),
				Snippet:     val,
				Remediation: remediation.CreateWorkflowPermissionRemediation(path),
			})
		}
		return nil
	}

	dl.Info(&checker.LogMessage{
		Path:   path,
		Type:   checker.FileTypeSource,
		Offset: lineNumber,
		Text:   fmt.Sprintf("%s '%v' permission set to '%v'", permLevel, permissionKey, val),
		// TODO: set Snippet.
	})
	return nil
}

func validateMapPermissions(scopes map[string]*actionlint.PermissionScope, permLevel, path string,
	dl checker.DetailLogger, pPermissions map[permission]bool,
	ignoredPermissions map[permission]bool,
) error {
	for key, v := range scopes {
		if err := validatePermission(permission(key), v, permLevel, path, dl, pPermissions, ignoredPermissions); err != nil {
			return err
		}
	}
	return nil
}

func recordPermissionWrite(pPermissions map[permission]bool, perm permission) {
	pPermissions[perm] = true
}

func getWritePermissionsMap(p *permissionCbData, path, permLevel string) map[permission]bool {
	if _, exists := p.workflows[path]; !exists {
		p.workflows[path] = permissions{
			topLevelWritePermissions: make(map[permission]bool),
			jobLevelWritePermissions: make(map[permission]bool),
		}
	}
	if permLevel == jobLevelPermission {
		return p.workflows[path].jobLevelWritePermissions
	}
	return p.workflows[path].topLevelWritePermissions
}

func recordAllPermissionsWrite(p *permissionCbData, permLevel, path string) {
	// Special case: `all` does not correspond
	// to a GitHub permission.
	m := getWritePermissionsMap(p, path, permLevel)
	m[permissionAll] = true
}

func validatePermissions(permissions *actionlint.Permissions, permLevel, path string,
	dl checker.DetailLogger, pdata *permissionCbData,
	ignoredPermissions map[permission]bool,
) error {
	allIsSet := permissions != nil && permissions.All != nil && permissions.All.Value != ""
	scopeIsSet := permissions != nil && len(permissions.Scopes) > 0
	if permissions == nil || (!allIsSet && !scopeIsSet) {
		dl.Info(&checker.LogMessage{
			Path:   path,
			Type:   checker.FileTypeSource,
			Offset: checker.OffsetDefault,
			Text:   fmt.Sprintf("%s permissions set to 'none'", permLevel),
		})
	}
	if allIsSet {
		val := permissions.All.Value
		lineNumber := fileparser.GetLineNumber(permissions.All.Pos)
		if !strings.EqualFold(val, "read-all") && val != "" {
			dl.Warn(&checker.LogMessage{
				Path:        path,
				Type:        checker.FileTypeSource,
				Offset:      lineNumber,
				Text:        fmt.Sprintf("%s permissions set to '%v'", permLevel, val),
				Snippet:     val,
				Remediation: remediation.CreateWorkflowPermissionRemediation(path),
			})
			recordAllPermissionsWrite(pdata, permLevel, path)
			return nil
		}

		dl.Info(&checker.LogMessage{
			Path:        path,
			Type:        checker.FileTypeSource,
			Offset:      lineNumber,
			Text:        fmt.Sprintf("%s permissions set to '%v'", permLevel, val),
			Snippet:     val,
			Remediation: remediation.CreateWorkflowPermissionRemediation(path),
		})
	} else /* scopeIsSet == true */ if err := validateMapPermissions(permissions.Scopes,
		permLevel, path, dl, getWritePermissionsMap(pdata, path, permLevel), ignoredPermissions); err != nil {
		return err
	}
	return nil
}

func validateTopLevelPermissions(workflow *actionlint.Workflow, path string,
	dl checker.DetailLogger, pdata *permissionCbData,
) error {
	// Check if permissions are set explicitly.
	if workflow.Permissions == nil {
		dl.Warn(&checker.LogMessage{
			Path:        path,
			Type:        checker.FileTypeSource,
			Offset:      checker.OffsetDefault,
			Text:        fmt.Sprintf("no %s permission defined", topLevelPermission),
			Remediation: remediation.CreateWorkflowPermissionRemediation(path),
		})
		recordAllPermissionsWrite(pdata, topLevelPermission, path)
		return nil
	}

	return validatePermissions(workflow.Permissions, topLevelPermission, path, dl,
		pdata, map[permission]bool{})
}

func validatejobLevelPermissions(workflow *actionlint.Workflow, path string,
	dl checker.DetailLogger, pdata *permissionCbData,
	ignoredPermissions map[permission]bool,
) error {
	for _, job := range workflow.Jobs {
		// Run-level permissions may be left undefined.
		// For most workflows, no write permissions are needed,
		// so only top-level read-only permissions need to be declared.
		if job.Permissions == nil {
			dl.Debug(&checker.LogMessage{
				Path:        path,
				Type:        checker.FileTypeSource,
				Offset:      fileparser.GetLineNumber(job.Pos),
				Text:        fmt.Sprintf("no %s permission defined", jobLevelPermission),
				Remediation: remediation.CreateWorkflowPermissionRemediation(path),
			})
			recordAllPermissionsWrite(pdata, jobLevelPermission, path)
			continue
		}
		err := validatePermissions(job.Permissions, jobLevelPermission,
			path, dl, pdata, ignoredPermissions)
		if err != nil {
			return err
		}
	}
	return nil
}

func isPermissionOfInterest(name permission, ignoredPermissions map[permission]bool) bool {
	for _, p := range permissionsOfInterest {
		_, present := ignoredPermissions[p]
		if strings.EqualFold(string(name), string(p)) && !present {
			return true
		}
	}
	return false
}

func permissionIsPresent(perms permissions, name permission) bool {
	return permissionIsPresentInTopLevel(perms, name) ||
		permissionIsPresentInRunLevel(perms, name)
}

func permissionIsPresentInTopLevel(perms permissions, name permission) bool {
	_, ok := perms.topLevelWritePermissions[name]
	return ok
}

func permissionIsPresentInRunLevel(perms permissions, name permission) bool {
	_, ok := perms.jobLevelWritePermissions[name]
	return ok
}

// Calculate the score.
func calculateScore(result permissionCbData) int {
	// See list https://github.blog/changelog/2021-04-20-github-actions-control-permissions-for-github_token/.
	// Note: there are legitimate reasons to use some of the permissions like checks, deployments, etc.
	// in CI/CD systems https://docs.travis-ci.com/user/github-oauth-scopes/.

	// Start with a perfect score.
	score := float32(checker.MaxResultScore)

	// Retrieve the overall results.
	for _, perms := range result.workflows {
		// If no top level permissions are defined, all the permissions
		// are enabled by default, hence permissionAll. In this case,
		if permissionIsPresentInTopLevel(perms, permissionAll) {
			if permissionIsPresentInRunLevel(perms, permissionAll) {
				// ... give lowest score if no run level permissions are defined either.
				return checker.MinResultScore
			}
			// ... reduce score if run level permissions are defined.
			score -= 0.5
		}

		// status: https://docs.github.com/en/rest/reference/repos#statuses.
		// May allow an attacker to change the result of pre-submit and get a PR merged.
		// Low risk: -0.5.
		if permissionIsPresent(perms, permissionStatuses) {
			score -= 0.5
		}

		// checks.
		// May allow an attacker to edit checks to remove pre-submit and introduce a bug.
		// Low risk: -0.5.
		if permissionIsPresent(perms, permissionChecks) {
			score -= 0.5
		}

		// secEvents.
		// May allow attacker to read vuln reports before patch available.
		// Low risk: -1
		if permissionIsPresent(perms, permissionSecurityEvents) {
			score--
		}

		// deployments: https://docs.github.com/en/rest/reference/repos#deployments.
		// May allow attacker to charge repo owner by triggering VM runs,
		// and tiny chance an attacker can trigger a remote
		// service with code they own if server accepts code/location var unsanitized.
		// Low risk: -1
		if permissionIsPresent(perms, permissionDeployments) {
			score--
		}

		// contents.
		// Allows attacker to commit unreviewed code.
		// High risk: -10
		if permissionIsPresent(perms, permissionContents) {
			score -= checker.MaxResultScore
		}

		// packages: https://docs.github.com/en/packages/learn-github-packages/about-permissions-for-github-packages.
		// Allows attacker to publish packages.
		// High risk: -10
		if permissionIsPresent(perms, permissionPackages) {
			score -= checker.MaxResultScore
		}

		// actions.
		// May allow an attacker to steal GitHub secrets by approving to run an action that needs approval.
		// High risk: -10
		if permissionIsPresent(perms, permissionActions) {
			score -= checker.MaxResultScore
		}

		if score < checker.MinResultScore {
			break
		}
	}

	// We're done, calculate the final score.
	if score < checker.MinResultScore {
		return checker.MinResultScore
	}

	return int(score)
}

// Create the result.
func createResultForLeastPrivilegeTokens(result permissionCbData, err error) checker.CheckResult {
	if err != nil {
		return checker.CreateRuntimeErrorResult(CheckTokenPermissions, err)
	}

	score := calculateScore(result)

	if score != checker.MaxResultScore {
		return checker.CreateResultWithScore(CheckTokenPermissions,
			"non read-only tokens detected in GitHub workflows", score)
	}

	return checker.CreateMaxScoreResult(CheckTokenPermissions,
		"tokens are read-only in GitHub workflows")
}

func createIgnoredPermissions(workflow *actionlint.Workflow, fp string, dl checker.DetailLogger) map[permission]bool {
	ignoredPermissions := make(map[permission]bool)
	if requiresPackagesPermissions(workflow, fp, dl) {
		ignoredPermissions[permissionPackages] = true
	}
	if requiresContentsPermissions(workflow, fp, dl) {
		ignoredPermissions[permissionContents] = true
	}
	if isSARIFUploadWorkflow(workflow, fp, dl) {
		ignoredPermissions[permissionSecurityEvents] = true
	}

	return ignoredPermissions
}

// Scanning tool run externally and SARIF file uploaded.
func isSARIFUploadWorkflow(workflow *actionlint.Workflow, fp string, dl checker.DetailLogger) bool {
	//nolint
	// CodeQl analysis workflow automatically sends sarif file to GitHub.
	// https://docs.github.com/en/code-security/secure-coding/integrating-with-code-scanning/uploading-a-sarif-file-to-github#about-sarif-file-uploads-for-code-scanning.
	// `The CodeQL action uploads the SARIF file automatically when it completes analysis`.
	if isCodeQlAnalysisWorkflow(workflow, fp, dl) {
		return true
	}

	//nolint
	// Third-party scanning tools use the SARIF-upload action from code-ql.
	// https://docs.github.com/en/code-security/secure-coding/integrating-with-code-scanning/uploading-a-sarif-file-to-github#uploading-a-code-scanning-analysis-with-github-actions
	// We only support CodeQl today.
	if isSARIFUploadAction(workflow, fp, dl) {
		return true
	}

	// TODO: some third party tools may upload directly thru their actions.
	// Very unlikely.
	// See https://github.com/marketplace for tools.

	return false
}

// CodeQl run externally and SARIF file uploaded.
func isSARIFUploadAction(workflow *actionlint.Workflow, fp string, dl checker.DetailLogger) bool {
	for _, job := range workflow.Jobs {
		for _, step := range job.Steps {
			uses := fileparser.GetUses(step)
			if uses == nil {
				continue
			}
			if strings.HasPrefix(uses.Value, "github/codeql-action/upload-sarif@") {
				dl.Debug(&checker.LogMessage{
					Path:   fp,
					Type:   checker.FileTypeSource,
					Offset: fileparser.GetLineNumber(uses.Pos),
					Text:   "codeql SARIF upload workflow detected",
					// TODO: set Snippet.
				})
				return true
			}
		}
	}
	dl.Debug(&checker.LogMessage{
		Path:   fp,
		Type:   checker.FileTypeSource,
		Offset: checker.OffsetDefault,
		Text:   "not a codeql upload SARIF workflow",
	})
	return false
}

//nolint
// CodeQl run within GitHub worklow automatically bubbled up to
// security events, see
// https://docs.github.com/en/code-security/secure-coding/automatically-scanning-your-code-for-vulnerabilities-and-errors/configuring-code-scanning.
func isCodeQlAnalysisWorkflow(workflow *actionlint.Workflow, fp string, dl checker.DetailLogger) bool {
	for _, job := range workflow.Jobs {
		for _, step := range job.Steps {
			uses := fileparser.GetUses(step)
			if uses == nil {
				continue
			}
			if strings.HasPrefix(uses.Value, "github/codeql-action/analyze@") {
				dl.Debug(&checker.LogMessage{
					Path:   fp,
					Type:   checker.FileTypeSource,
					Offset: fileparser.GetLineNumber(uses.Pos),
					Text:   "codeql workflow detected",
					// TODO: set Snippet.
				})
				return true
			}
		}
	}
	dl.Debug(&checker.LogMessage{
		Path:   fp,
		Type:   checker.FileTypeSource,
		Offset: checker.OffsetDefault,
		Text:   "not a codeql workflow",
	})
	return false
}

// A packaging workflow using GitHub's supported packages:
// https://docs.github.com/en/packages.
func requiresPackagesPermissions(workflow *actionlint.Workflow, fp string, dl checker.DetailLogger) bool {
	// TODO: add support for GitHub registries.
	// Example: https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-npm-registry.
	// This feature requires parsing actions properly.
	// For now, we just re-use the Packaging check to verify that the
	// workflow is a packaging workflow.
	return isPackagingWorkflow(workflow, fp, dl)
}

// requiresContentsPermissions returns true if the workflow requires the `contents: write` permission.
func requiresContentsPermissions(workflow *actionlint.Workflow, fp string, dl checker.DetailLogger) bool {
	return isReleasingWorkflow(workflow, fp, dl) || isGitHubPagesDeploymentWorkflow(workflow, fp, dl)
}

// isGitHubPagesDeploymentWorkflow returns true if the workflow involves pushing static pages to GitHub pages.
func isGitHubPagesDeploymentWorkflow(workflow *actionlint.Workflow, fp string, dl checker.DetailLogger) bool {
	jobMatchers := []fileparser.JobMatcher{
		{
			Steps: []*fileparser.JobMatcherStep{
				{
					Uses: "peaceiris/actions-gh-pages",
				},
			},
			LogText: "candidate GitHub page deployment workflow using peaceiris/actions-gh-pages",
		},
	}
	return fileparser.AnyJobsMatch(workflow, jobMatchers, fp, dl, "not a GitHub Pages deployment workflow")
}

// isReleasingWorkflow returns true if the workflow involves creating a release on GitHub.
func isReleasingWorkflow(workflow *actionlint.Workflow, fp string, dl checker.DetailLogger) bool {
	jobMatchers := []fileparser.JobMatcher{
		{
			// Python packages.
			// This is a custom Python packaging/releasing workflow based on semantic versioning.
			Steps: []*fileparser.JobMatcherStep{
				{
					Uses: "relekang/python-semantic-release",
				},
			},
			LogText: "candidate python publishing workflow using python-semantic-release",
		},
		{
			// Go packages.
			Steps: []*fileparser.JobMatcherStep{
				{
					Uses: "actions/setup-go",
				},
				{
					Uses: "goreleaser/goreleaser-action",
				},
			},
			LogText: "candidate golang publishing workflow",
		},
	}

	return fileparser.AnyJobsMatch(workflow, jobMatchers, fp, dl, "not a releasing workflow")
}

// TODO: remove when migrated to raw results.
// Should be using the definition in raw/packaging.go.
func isPackagingWorkflow(workflow *actionlint.Workflow, fp string, dl checker.DetailLogger) bool {
	jobMatchers := []fileparser.JobMatcher{
		{
			Steps: []*fileparser.JobMatcherStep{
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
			Steps: []*fileparser.JobMatcherStep{
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
			Steps: []*fileparser.JobMatcherStep{
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
			Steps: []*fileparser.JobMatcherStep{
				{
					Run: "gem.*push",
				},
			},
			LogText: "candidate ruby publishing workflow using gem",
		},
		{
			// NuGet packages.
			Steps: []*fileparser.JobMatcherStep{
				{
					Run: "nuget.*push",
				},
			},
			LogText: "candidate nuget publishing workflow",
		},
		{
			// Docker packages.
			Steps: []*fileparser.JobMatcherStep{
				{
					Run: "docker.*push",
				},
			},
			LogText: "candidate docker publishing workflow",
		},
		{
			// Docker packages.
			Steps: []*fileparser.JobMatcherStep{
				{
					Uses: "docker/build-push-action",
				},
			},
			LogText: "candidate docker publishing workflow",
		},
		{
			// Python packages.
			Steps: []*fileparser.JobMatcherStep{
				{
					Uses: "actions/setup-python",
				},
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
			Steps: []*fileparser.JobMatcherStep{
				{
					Uses: "relekang/python-semantic-release",
				},
			},
			LogText: "candidate python publishing workflow using python-semantic-release",
		},
		{
			// Go packages.
			Steps: []*fileparser.JobMatcherStep{
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
			Steps: []*fileparser.JobMatcherStep{
				{
					Run: "cargo.*publish",
				},
			},
			LogText: "candidate rust publishing workflow using cargo",
		},
	}

	return fileparser.AnyJobsMatch(workflow, jobMatchers, fp, dl, "not a publishing workflow")
}
