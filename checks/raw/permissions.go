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
	"strings"

	"github.com/rhysd/actionlint"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/fileparser"
	sce "github.com/ossf/scorecard/v4/errors"
)

type permission string

const (
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

type permissionCbData struct {
	results checker.TokenPermissionsData
}

// TokenPermissions runs Token-Permissions check.
func TokenPermissions(c *checker.CheckRequest) (checker.TokenPermissionsData, error) {
	// data is shared across all GitHub workflows.
	var data permissionCbData

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
	ignoredPermissions := createIgnoredPermissions(workflow, path, pdata)
	if err := validatejobLevelPermissions(workflow, path, pdata, ignoredPermissions); err != nil {
		return false, err
	}

	// TODO(laurent): 2. Identify github actions that require write and add checks.

	// TODO(laurent): 3. Read a few runs and ensures they have the same permissions.

	return true, nil
}

func validatePermission(permissionKey permission, permissionValue *actionlint.PermissionScope,
	permLoc checker.PermissionLocation, path string, p *permissionCbData,
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
					Type:         checker.PermissionLevelWrite,
					// TODO: Job
				})
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
					// It's a write but not considered dangerous.
					Type: checker.PermissionLevelUnknown,
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
			Type:         typeOfPermission(val),
			// TODO: Job
		})
	return nil
}

func typeOfPermission(val string) checker.PermissionLevel {
	switch val {
	case "read", "read-all":
		return checker.PermissionLevelRead
	case "none": //nolint:goconst
		return checker.PermissionLevelNone
	}
	return checker.PermissionLevelUnknown
}

func validateMapPermissions(scopes map[string]*actionlint.PermissionScope, permLoc checker.PermissionLocation,
	path string, pdata *permissionCbData,
	ignoredPermissions map[permission]bool,
) error {
	for key, v := range scopes {
		if err := validatePermission(permission(key), v, permLoc, path, pdata, ignoredPermissions); err != nil {
			return err
		}
	}
	return nil
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
				Type:         checker.PermissionLevelNone,
				Value:        &none,
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
					Type:         checker.PermissionLevelWrite,
					// TODO: Job
				})

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
				Type:         typeOfPermission(val),
				// TODO: Job
			})
	} else /* scopeIsSet == true */ if err := validateMapPermissions(permissions.Scopes,
		permLoc, path, pdata, ignoredPermissions); err != nil {
		return err
	}
	return nil
}

func validateTopLevelPermissions(workflow *actionlint.Workflow, path string,
	pdata *permissionCbData,
) error {
	// Check if permissions are set explicitly.
	if workflow.Permissions == nil {
		permLoc := checker.PermissionLocationTop
		pdata.results.TokenPermissions = append(pdata.results.TokenPermissions,
			checker.TokenPermission{
				File: &checker.File{
					Path:   path,
					Type:   checker.FileTypeSource,
					Offset: checker.OffsetDefault,
				},
				LocationType: &permLoc,
				Type:         checker.PermissionLevelUndeclared,
				// TODO: Job
			})

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
			permLoc := checker.PermissionLocationJob
			pdata.results.TokenPermissions = append(pdata.results.TokenPermissions,
				checker.TokenPermission{
					File: &checker.File{
						Path:   path,
						Type:   checker.FileTypeSource,
						Offset: fileparser.GetLineNumber(job.Pos),
					},
					LocationType: &permLoc,
					Type:         checker.PermissionLevelUndeclared,
					Msg:          stringPointer(fmt.Sprintf("no %s permission defined", permLoc)),
					// TODO: Job
				})

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

func createIgnoredPermissions(workflow *actionlint.Workflow, fp string,
	pdata *permissionCbData,
) map[permission]bool {
	ignoredPermissions := make(map[permission]bool)
	if requiresPackagesPermissions(workflow, fp, pdata) {
		ignoredPermissions[permissionPackages] = true
	}
	if requiresContentsPermissions(workflow, fp, pdata) {
		ignoredPermissions[permissionContents] = true
	}
	if isSARIFUploadWorkflow(workflow, fp, pdata) {
		ignoredPermissions[permissionSecurityEvents] = true
	}

	return ignoredPermissions
}

// Scanning tool run externally and SARIF file uploaded.
func isSARIFUploadWorkflow(workflow *actionlint.Workflow, fp string, pdata *permissionCbData) bool {
	// TODO: some third party tools may upload directly thru their actions.
	// Very unlikely.
	// See https://github.com/marketplace for tools.
	return isAllowedWorkflow(workflow, fp, pdata)
}

func isAllowedWorkflow(workflow *actionlint.Workflow, fp string, pdata *permissionCbData) bool {
	allowlist := map[string]bool{
		//nolint
		// CodeQl analysis workflow automatically sends sarif file to GitHub.
		// https://docs.github.com/en/code-security/secure-coding/integrating-with-code-scanning/uploading-a-sarif-file-to-github#about-sarif-file-uploads-for-code-scanning.
		// `The CodeQL action uploads the SARIF file automatically when it completes analysis`.
		"github/codeql-action/analyze": true,

		//nolint
		// Third-party scanning tools use the SARIF-upload action from code-ql.
		// https://docs.github.com/en/code-security/secure-coding/integrating-with-code-scanning/uploading-a-sarif-file-to-github#uploading-a-code-scanning-analysis-with-github-actions
		// We only support CodeQl today.
		"github/codeql-action/upload-sarif": true,

		// allow our own action, which writes sarif files
		// https://github.com/ossf/scorecard-action
		"ossf/scorecard-action": true,
	}

	tokenPermissions := checker.TokenPermission{
		File: &checker.File{
			Path:   fp,
			Type:   checker.FileTypeSource,
			Offset: checker.OffsetDefault,
			// TODO: set Snippet.
		},
		Type: checker.PermissionLevelUnknown,
		// TODO: Job
	}

	for _, job := range workflow.Jobs {
		for _, step := range job.Steps {
			uses := fileparser.GetUses(step)
			if uses == nil {
				continue
			}
			// remove any version pinning for the comparison
			uses.Value = strings.Split(uses.Value, "@")[0]
			if allowlist[uses.Value] {
				tokenPermissions.File.Offset = fileparser.GetLineNumber(uses.Pos)
				tokenPermissions.Msg = stringPointer("allowed SARIF workflow detected")
				pdata.results.TokenPermissions = append(pdata.results.TokenPermissions, tokenPermissions)
				return true
			}
		}
	}
	tokenPermissions.Msg = stringPointer("not a SARIF workflow, or not an allowed one")
	pdata.results.TokenPermissions = append(pdata.results.TokenPermissions, tokenPermissions)
	return false
}

// A packaging workflow using GitHub's supported packages:
// https://docs.github.com/en/packages.
func requiresPackagesPermissions(workflow *actionlint.Workflow, fp string, pdata *permissionCbData) bool {
	// TODO: add support for GitHub registries.
	// Example: https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-npm-registry.
	match, ok := fileparser.IsPackagingWorkflow(workflow, fp)
	// Print debug messages.
	pdata.results.TokenPermissions = append(pdata.results.TokenPermissions,
		checker.TokenPermission{
			File: &checker.File{
				Path:   fp,
				Type:   checker.FileTypeSource,
				Offset: checker.OffsetDefault,
			},
			Msg:  &match.Msg,
			Type: checker.PermissionLevelUnknown,
		})

	return ok
}

// requiresContentsPermissions returns true if the workflow requires the `contents: write` permission.
func requiresContentsPermissions(workflow *actionlint.Workflow, fp string, pdata *permissionCbData) bool {
	return isReleasingWorkflow(workflow, fp, pdata) || isGitHubPagesDeploymentWorkflow(workflow, fp, pdata)
}

// isGitHubPagesDeploymentWorkflow returns true if the workflow involves pushing static pages to GitHub pages.
func isGitHubPagesDeploymentWorkflow(workflow *actionlint.Workflow, fp string, pdata *permissionCbData) bool {
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

	return isWorkflowOf(workflow, fp, jobMatchers,
		"not a GitHub Pages deployment workflow", pdata)
}

// isReleasingWorkflow returns true if the workflow involves creating a release on GitHub.
func isReleasingWorkflow(workflow *actionlint.Workflow, fp string, pdata *permissionCbData) bool {
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
			// Go binaries.
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
			// SLSA Go builder. https://github.com/slsa-framework/slsa-github-generator
			Steps: []*fileparser.JobMatcherStep{
				{
					Uses: "slsa-framework/slsa-github-generator/.github/workflows/builder_go_slsa3.yml",
				},
			},
			LogText: "candidate SLSA publishing workflow using slsa-github-generator",
		},
		{
			// SLSA generic generator. https://github.com/slsa-framework/slsa-github-generator
			Steps: []*fileparser.JobMatcherStep{
				{
					Uses: "slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml",
				},
			},
			LogText: "candidate SLSA publishing workflow using slsa-github-generator",
		},
		{
			// Running mvn release:prepare requires committing changes.
			// https://maven.apache.org/maven-release/maven-release-plugin/examples/prepare-release.html
			Steps: []*fileparser.JobMatcherStep{
				{
					Run: ".*mvn.*release:prepare.*",
				},
			},
			LogText: "candidate mvn release workflow",
		},
	}

	return isWorkflowOf(workflow, fp, jobMatchers, "not a releasing workflow", pdata)
}

func isWorkflowOf(workflow *actionlint.Workflow, fp string,
	jobMatchers []fileparser.JobMatcher, msg string,
	pdata *permissionCbData,
) bool {
	match, ok := fileparser.AnyJobsMatch(workflow, jobMatchers, fp, msg)

	// Print debug messages.
	pdata.results.TokenPermissions = append(pdata.results.TokenPermissions,
		checker.TokenPermission{
			File: &checker.File{
				Path:   fp,
				Type:   checker.FileTypeSource,
				Offset: checker.OffsetDefault,
			},
			Msg:  &match.Msg,
			Type: checker.PermissionLevelUnknown,
		})

	return ok
}
