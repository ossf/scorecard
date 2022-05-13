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
	"strings"

	"github.com/rhysd/actionlint"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/fileparser"
	sce "github.com/ossf/scorecard/v4/errors"
)

// CheckTokenPermissions is the exported name for Token-Permissions check.
const (
	jobLevelPermission = "job level"
	topLevelPermission = "top level"
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
	results   checker.TokenPermissionsData
}

// TokenPermissions runs Token-Permissions check.
func TokenPermissions(c *checker.CheckRequest) (checker.TokenPermissionsData, error) {
	// data is shared across all GitHub workflows.
	data := permissionCbData{
		workflows: make(map[string]permissions),
	}

	if err := remdiationSetup(c); err != nil {
		return data.results, err
	}

	err := fileparser.OnMatchingFileContentDo(c.RepoClient, fileparser.PathMatcher{
		Pattern:       ".github/workflows/*",
		CaseSensitive: false,
	}, validateGitHubActionTokenPermissions, &data)

	return data.results, err
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
	if len(args) != 1 {
		return false, fmt.Errorf(
			"validateGitHubActionTokenPermissions requires exactly 2 arguments: %w", errInvalidArgLength)
	}
	pdata, ok := args[0].(*permissionCbData)
	if !ok {
		return false, fmt.Errorf(
			"validateGitHubActionTokenPermissions requires arg[0] of type *permissionCbData: %w", errInvalidArgType)
	}

	if !fileparser.CheckFileContainsCommands(content, "#") {
		return true, nil
	}

	workflow, errs := actionlint.Parse(content)
	if len(errs) > 0 && workflow == nil {
		return false, fileparser.FormatActionlintError(errs)
	}

	// Temporary fix to have a DetailLogger necessary for calls to Packaging check.
	dl := checker.NewLogger()

	// 1. Top-level permission definitions.
	//nolint
	// https://docs.github.com/en/actions/reference/authentication-in-a-workflow#example-1-passing-the-github_token-as-an-input,
	// https://github.blog/changelog/2021-04-20-github-actions-control-permissions-for-github_token/,
	// https://docs.github.com/en/actions/reference/authentication-in-a-workflow#modifying-the-permissions-for-the-github_token.
	if err := validateTopLevelPermissions(workflow, path, pdata); err != nil {
		return false, err
	}

	// 2. Run-level permission definitions,
	// see https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idpermissions.
	ignoredPermissions := createIgnoredPermissions(workflow, path, dl, pdata)
	if err := validatejobLevelPermissions(workflow, path, pdata, ignoredPermissions); err != nil {
		return false, err
	}

	// Extract the logs.
	logs := dl.Flush()
	for _, l := range logs {
		pdata.results.TokenPermissions = append(pdata.results.TokenPermissions,
			checker.TokenPermission{
				Log: checker.DetailToRawLog(&l),
			})
	}

	// TODO(laurent): 2. Identify github actions that require write and add checks.

	// TODO(laurent): 3. Read a few runs and ensures they have the same permissions.

	return true, nil
}

func validatePermission(permissionKey permission, permissionValue *actionlint.PermissionScope,
	permLoc checker.PermissionLocation, path string, p *permissionCbData, pPermissions map[permission]bool,
	ignoredPermissions map[permission]bool,
) error {
	if permissionValue.Value == nil {
		return sce.WithMessage(sce.ErrScorecardInternal, errInvalidGitHubWorkflow.Error())
	}
	key := string(permissionKey)
	val := permissionValue.Value.Value
	lineNumber := fileparser.GetLineNumber(permissionValue.Value.Pos)
	if strings.EqualFold(val, "write") {
		if isPermissionOfInterest(permissionKey, ignoredPermissions) {
			p.results.TokenPermissions = append(p.results.TokenPermissions,
				checker.TokenPermission{
					File: &checker.File{
						Path:    path,
						Type:    checker.FileTypeSource,
						Offset:  lineNumber,
						Snippet: val,
					},
					LocationType: &permLoc,
					Name:         &key,
					Value:        &val,
					Remediation:  createWorkflowPermissionRemediation(path),
					Log: checker.Log{
						Level: checker.LogLevelWarn,
						Msg: fmt.Sprintf("%s '%v' permission set to '%v'",
							checker.PermissionLocationToString(permLoc), key, val),
					},
					// TODO: Job
				})
			recordPermissionWrite(pPermissions, permissionKey)
		} else {
			// Only log for debugging, otherwise
			// it may confuse users.
			p.results.TokenPermissions = append(p.results.TokenPermissions,
				checker.TokenPermission{
					File: &checker.File{
						Path:    path,
						Type:    checker.FileTypeSource,
						Offset:  lineNumber,
						Snippet: val,
					},
					LocationType: &permLoc,
					Name:         &key,
					Value:        &val,
					Log: checker.Log{
						Level: checker.LogLevelDebug,
						Msg: fmt.Sprintf("%s '%v' permission set to '%v'",
							checker.PermissionLocationToString(permLoc), key, val),
					},
					// TODO: Job
				})
		}
		return nil
	}

	p.results.TokenPermissions = append(p.results.TokenPermissions,
		checker.TokenPermission{
			File: &checker.File{
				Path:   path,
				Type:   checker.FileTypeSource,
				Offset: lineNumber,
				// TODO: set Snippet.
			},
			LocationType: &permLoc,
			Name:         &key,
			Value:        &val,
			Log: checker.Log{
				Level: checker.LogLevelInfo,
				Msg: fmt.Sprintf("%s '%v' permission set to '%v'",
					checker.PermissionLocationToString(permLoc), key, val),
			},
			// TODO: Job
		})
	return nil
}

func validateMapPermissions(scopes map[string]*actionlint.PermissionScope, permLoc checker.PermissionLocation,
	path string, pdata *permissionCbData, pPermissions map[permission]bool,
	ignoredPermissions map[permission]bool,
) error {
	for key, v := range scopes {
		if err := validatePermission(permission(key), v, permLoc, path, pdata, pPermissions, ignoredPermissions); err != nil {
			return err
		}
	}
	return nil
}

func recordPermissionWrite(pPermissions map[permission]bool, perm permission) {
	pPermissions[perm] = true
}

func getWritePermissionsMap(p *permissionCbData, path string, permLoc checker.PermissionLocation) map[permission]bool {
	if _, exists := p.workflows[path]; !exists {
		p.workflows[path] = permissions{
			topLevelWritePermissions: make(map[permission]bool),
			jobLevelWritePermissions: make(map[permission]bool),
		}
	}
	if permLoc == checker.PermissionLocationJob {
		return p.workflows[path].jobLevelWritePermissions
	}
	return p.workflows[path].topLevelWritePermissions
}

func recordAllPermissionsWrite(p *permissionCbData, permLoc checker.PermissionLocation, path string) {
	// Special case: `all` does not correspond
	// to a GitHub permission.
	m := getWritePermissionsMap(p, path, permLoc)
	m[permissionAll] = true
}

func validatePermissions(permissions *actionlint.Permissions, permLoc checker.PermissionLocation,
	path string, pdata *permissionCbData,
	ignoredPermissions map[permission]bool,
) error {
	allIsSet := permissions != nil && permissions.All != nil && permissions.All.Value != ""
	scopeIsSet := permissions != nil && len(permissions.Scopes) > 0
	none := "none"
	if permissions == nil || (!allIsSet && !scopeIsSet) {
		pdata.results.TokenPermissions = append(pdata.results.TokenPermissions,
			checker.TokenPermission{
				File: &checker.File{
					Path:   path,
					Type:   checker.FileTypeSource,
					Offset: checker.OffsetDefault,
				},
				LocationType: &permLoc,
				Log: checker.Log{
					Level: checker.LogLevelInfo,
					Msg: fmt.Sprintf("%s permissions set to 'none'",
						checker.PermissionLocationToString(permLoc)),
				},
				Value: &none,
				// TODO: Job, etc.
			})
	}
	if allIsSet {
		val := permissions.All.Value
		lineNumber := fileparser.GetLineNumber(permissions.All.Pos)
		if !strings.EqualFold(val, "read-all") && val != "" {
			pdata.results.TokenPermissions = append(pdata.results.TokenPermissions,
				checker.TokenPermission{
					File: &checker.File{
						Path:    path,
						Type:    checker.FileTypeSource,
						Offset:  lineNumber,
						Snippet: val,
					},
					LocationType: &permLoc,
					Value:        &val,
					Remediation:  createWorkflowPermissionRemediation(path),
					Log: checker.Log{
						Level: checker.LogLevelWarn,
						Msg: fmt.Sprintf("%s permissions set to '%v'",
							checker.PermissionLocationToString(permLoc), val),
					},
					// TODO: Job
				})

			recordAllPermissionsWrite(pdata, permLoc, path)
			return nil
		}

		pdata.results.TokenPermissions = append(pdata.results.TokenPermissions,
			checker.TokenPermission{
				File: &checker.File{
					Path:    path,
					Type:    checker.FileTypeSource,
					Offset:  lineNumber,
					Snippet: val,
				},
				LocationType: &permLoc,
				Value:        &val,
				Log: checker.Log{
					Level: checker.LogLevelInfo,
					Msg: fmt.Sprintf("%s permissions set to '%v'",
						checker.PermissionLocationToString(permLoc), val),
				},
				// TODO: Job
			})
	} else /* scopeIsSet == true */ if err := validateMapPermissions(permissions.Scopes,
		permLoc, path, pdata, getWritePermissionsMap(pdata, path, permLoc), ignoredPermissions); err != nil {
		return err
	}
	return nil
}

func validateTopLevelPermissions(workflow *actionlint.Workflow, path string,
	pdata *permissionCbData,
) error {
	// Check if permissions are set explicitly.
	if workflow.Permissions == nil {
		var permLoc checker.PermissionLocation = checker.PermissionLocationTop
		var alertType checker.PermissionAlertType = checker.PermissionAlertTypeUndeclared
		pdata.results.TokenPermissions = append(pdata.results.TokenPermissions,
			checker.TokenPermission{
				File: &checker.File{
					Path:   path,
					Type:   checker.FileTypeSource,
					Offset: checker.OffsetDefault,
				},
				LocationType: &permLoc,
				AlertType:    &alertType,
				Remediation:  createWorkflowPermissionRemediation(path),
				Log: checker.Log{
					Level: checker.LogLevelWarn,
					Msg:   fmt.Sprintf("no %s permission defined", checker.PermissionLocationToString(permLoc)),
				},
				// TODO: Job
			})

		recordAllPermissionsWrite(pdata, checker.PermissionLocationTop, path)
		return nil
	}

	return validatePermissions(workflow.Permissions, checker.PermissionLocationTop, path,
		pdata, map[permission]bool{})
}

func validatejobLevelPermissions(workflow *actionlint.Workflow, path string,
	pdata *permissionCbData,
	ignoredPermissions map[permission]bool,
) error {
	for _, job := range workflow.Jobs {
		// Run-level permissions may be left undefined.
		// For most workflows, no write permissions are needed,
		// so only top-level read-only permissions need to be declared.
		if job.Permissions == nil {
			var permLoc checker.PermissionLocation = checker.PermissionLocationJob
			var alertType checker.PermissionAlertType = checker.PermissionAlertTypeUndeclared
			pdata.results.TokenPermissions = append(pdata.results.TokenPermissions,
				checker.TokenPermission{
					File: &checker.File{
						Path:   path,
						Type:   checker.FileTypeSource,
						Offset: fileparser.GetLineNumber(job.Pos),
					},
					LocationType: &permLoc,
					AlertType:    &alertType,
					Log: checker.Log{
						Level: checker.LogLevelDebug,
						Msg:   fmt.Sprintf("no %s permission defined", checker.PermissionLocationToString(permLoc)),
					},
					// TODO: Job
				})

			recordAllPermissionsWrite(pdata, permLoc, path)
			continue
		}
		err := validatePermissions(job.Permissions, checker.PermissionLocationJob,
			path, pdata, ignoredPermissions)
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

func createIgnoredPermissions(workflow *actionlint.Workflow, fp string, dl checker.DetailLogger,
	pdata *permissionCbData,
) map[permission]bool {
	ignoredPermissions := make(map[permission]bool)
	if requiresPackagesPermissions(workflow, fp, dl) {
		ignoredPermissions[permissionPackages] = true
	}
	if requiresContentsPermissions(workflow, fp, dl) {
		ignoredPermissions[permissionContents] = true
	}
	if isSARIFUploadWorkflow(workflow, fp, pdata) {
		ignoredPermissions[permissionSecurityEvents] = true
	}

	return ignoredPermissions
}

// Scanning tool run externally and SARIF file uploaded.
func isSARIFUploadWorkflow(workflow *actionlint.Workflow, fp string, pdata *permissionCbData) bool {
	//nolint
	// CodeQl analysis workflow automatically sends sarif file to GitHub.
	// https://docs.github.com/en/code-security/secure-coding/integrating-with-code-scanning/uploading-a-sarif-file-to-github#about-sarif-file-uploads-for-code-scanning.
	// `The CodeQL action uploads the SARIF file automatically when it completes analysis`.
	if isCodeQlAnalysisWorkflow(workflow, fp, pdata) {
		return true
	}

	//nolint
	// Third-party scanning tools use the SARIF-upload action from code-ql.
	// https://docs.github.com/en/code-security/secure-coding/integrating-with-code-scanning/uploading-a-sarif-file-to-github#uploading-a-code-scanning-analysis-with-github-actions
	// We only support CodeQl today.
	if isSARIFUploadAction(workflow, fp, pdata) {
		return true
	}

	// TODO: some third party tools may upload directly thru their actions.
	// Very unlikely.
	// See https://github.com/marketplace for tools.

	return false
}

// CodeQl run externally and SARIF file uploaded.
func isSARIFUploadAction(workflow *actionlint.Workflow, fp string, pdata *permissionCbData) bool {
	for _, job := range workflow.Jobs {
		for _, step := range job.Steps {
			uses := fileparser.GetUses(step)
			if uses == nil {
				continue
			}
			if strings.HasPrefix(uses.Value, "github/codeql-action/upload-sarif@") {
				pdata.results.TokenPermissions = append(pdata.results.TokenPermissions,
					checker.TokenPermission{
						File: &checker.File{
							Path:   fp,
							Type:   checker.FileTypeSource,
							Offset: fileparser.GetLineNumber(uses.Pos),
							// TODO: set Snippet.
						},
						Log: checker.Log{
							Level: checker.LogLevelDebug,
							Msg:   "codeql SARIF upload workflow detected",
						},
						// TODO: Job
					})

				return true
			}
		}
	}
	pdata.results.TokenPermissions = append(pdata.results.TokenPermissions,
		checker.TokenPermission{
			File: &checker.File{
				Path:   fp,
				Type:   checker.FileTypeSource,
				Offset: checker.OffsetDefault,
			},
			Log: checker.Log{
				Level: checker.LogLevelDebug,
				Msg:   "not a codeql upload SARIF workflow",
			},
			// TODO: Job
		})

	return false
}

//nolint
// CodeQl run within GitHub worklow automatically bubbled up to
// security events, see
// https://docs.github.com/en/code-security/secure-coding/automatically-scanning-your-code-for-vulnerabilities-and-errors/configuring-code-scanning.
func isCodeQlAnalysisWorkflow(workflow *actionlint.Workflow, fp string, pdata *permissionCbData) bool {
	for _, job := range workflow.Jobs {
		for _, step := range job.Steps {
			uses := fileparser.GetUses(step)
			if uses == nil {
				continue
			}
			if strings.HasPrefix(uses.Value, "github/codeql-action/analyze@") {
				pdata.results.TokenPermissions = append(pdata.results.TokenPermissions,
					checker.TokenPermission{
						File: &checker.File{
							Path:   fp,
							Type:   checker.FileTypeSource,
							Offset: fileparser.GetLineNumber(uses.Pos),
						},
						Log: checker.Log{
							Level: checker.LogLevelDebug,
							Msg:   "codeql workflow detected",
						},
						// TODO: Job
					})

				return true
			}
		}
	}

	pdata.results.TokenPermissions = append(pdata.results.TokenPermissions,
		checker.TokenPermission{
			File: &checker.File{
				Path:   fp,
				Type:   checker.FileTypeSource,
				Offset: checker.OffsetDefault,
			},
			Log: checker.Log{
				Level: checker.LogLevelDebug,
				Msg:   "not a codeql workflow",
			},
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
	return false
	// TODO: isPackagingWorkflow(workflow, fp, dl)
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
