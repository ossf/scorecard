// Copyright 2022 OpenSSF Scorecard Authors
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
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/rhysd/actionlint"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/checks/fileparser"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/dotnet/csproj"
	"github.com/ossf/scorecard/v5/internal/dotnet/properties"
	"github.com/ossf/scorecard/v5/remediation"
)

type dotnetCsprojLockedData struct {
	Path          string
	LockedModeSet bool
}

type nugetPostProcessData struct {
	CsprojConfigs []dotnetCsprojLockedData
	CpmConfig     properties.CentralPackageManagementConfig
}

// PinningDependencies checks for (un)pinned dependencies.
func PinningDependencies(c *checker.CheckRequest) (checker.PinningDependenciesData, error) {
	var results checker.PinningDependenciesData

	// GitHub actions.
	if err := collectGitHubActionsWorkflowPinning(c, &results); err != nil {
		return checker.PinningDependenciesData{}, err
	}

	// // Docker files.
	if err := collectDockerfilePinning(c, &results); err != nil {
		return checker.PinningDependenciesData{}, err
	}

	// Docker downloads.
	if err := collectDockerfileInsecureDownloads(c, &results); err != nil {
		return checker.PinningDependenciesData{}, err
	}

	// Script downloads.
	if err := collectShellScriptInsecureDownloads(c, &results); err != nil {
		return checker.PinningDependenciesData{}, err
	}

	// Action script downloads.
	if err := collectGitHubWorkflowScriptInsecureDownloads(c, &results); err != nil {
		return checker.PinningDependenciesData{}, err
	}

	// Nuget Post Processing
	if err := postProcessNugetDependencies(c, &results); err != nil {
		return checker.PinningDependenciesData{}, err
	}

	return results, nil
}

func postProcessNugetDependencies(c *checker.CheckRequest,
	pinningDependenciesData *checker.PinningDependenciesData,
) error {
	unpinnedDependencies := getUnpinnedNugetDependencies(pinningDependenciesData)
	if len(unpinnedDependencies) == 0 {
		return nil
	}
	var nugetPostProcessData nugetPostProcessData
	if err := retrieveNugetCentralPackageManagement(c, &nugetPostProcessData); err != nil {
		return err
	}
	if err := retrieveCsprojConfig(c, &nugetPostProcessData); err != nil {
		return err
	}
	if nugetPostProcessData.CpmConfig.IsCPMEnabled {
		collectPostProcessNugetCPMDependencies(unpinnedDependencies, &nugetPostProcessData)
	} else {
		collectPostProcessNugetCsprojDependencies(unpinnedDependencies, &nugetPostProcessData)
	}

	return nil
}

func collectPostProcessNugetCPMDependencies(unpinnedNugetDependencies []*checker.Dependency,
	postProcessingData *nugetPostProcessData,
) {
	packageVersions := postProcessingData.CpmConfig.PackageVersions

	numUnfixedVersions, unfixedVersions := countUnfixedVersions(packageVersions)
	// if all dependencies are fixed to specific versions, pin all dependencies
	if numUnfixedVersions == 0 {
		pinAllNugetDependencies(unpinnedNugetDependencies)
		return
	}
	// if some or all dependencies are not fixed to specific versions, update the remediation
	for i := range unpinnedNugetDependencies {
		(unpinnedNugetDependencies)[i].Remediation.Text = (unpinnedNugetDependencies)[i].Remediation.Text +
			": some of dependency versions are not fixes to specific versions: " + unfixedVersions
	}
}

func retrieveNugetCentralPackageManagement(c *checker.CheckRequest, nugetPostProcessData *nugetPostProcessData) error {
	if err := fileparser.OnMatchingFileContentDo(c.RepoClient, fileparser.PathMatcher{
		Pattern:       "Directory.*.props",
		CaseSensitive: false,
	}, processDirectoryPropsFile, nugetPostProcessData, c.Dlogger); err != nil {
		return err
	}

	return nil
}

func processDirectoryPropsFile(path string, content []byte, args ...interface{}) (bool, error) {
	pdata, ok := args[0].(*nugetPostProcessData)
	if !ok {
		// panic if it is not correct type
		panic(fmt.Sprintf("expected type nugetPostProcessData, got %v", reflect.TypeOf(args[0])))
	}

	cpmConfig, err := properties.GetCentralPackageManagementConfig(path, content)
	if err != nil {
		dl, ok := args[1].(checker.DetailLogger)
		if !ok {
			// panic if it is not correct type
			panic(fmt.Sprintf("expected type checker.DetailLogger, got %v", reflect.TypeOf(args[1])))
		}

		dl.Warn(&checker.LogMessage{
			Text: fmt.Sprintf("malformed properties file: %v", err),
		})
		return true, nil
	}
	pdata.CpmConfig = cpmConfig
	return false, nil
}

func getUnpinnedNugetDependencies(pinningDependenciesData *checker.PinningDependenciesData) []*checker.Dependency {
	var unpinnedNugetDependencies []*checker.Dependency
	nugetDependencies := getDependenciesByType(pinningDependenciesData, checker.DependencyUseTypeNugetCommand)
	for i := range nugetDependencies {
		if !*nugetDependencies[i].Pinned {
			unpinnedNugetDependencies = append(unpinnedNugetDependencies, nugetDependencies[i])
		}
	}
	return unpinnedNugetDependencies
}

func getDependenciesByType(p *checker.PinningDependenciesData,
	useType checker.DependencyUseType,
) []*checker.Dependency {
	var deps []*checker.Dependency
	for i := range p.Dependencies {
		if p.Dependencies[i].Type == useType {
			deps = append(deps, &p.Dependencies[i])
		}
	}
	return deps
}

func collectPostProcessNugetCsprojDependencies(unpinnedNugetDependencies []*checker.Dependency,
	postProcessingData *nugetPostProcessData,
) {
	unlockedCsprojDeps, unlockedPath := countUnlocked(postProcessingData.CsprojConfigs)
	switch unlockedCsprojDeps {
	case len(postProcessingData.CsprojConfigs):
		// none of the csproject files set RestoreLockedMode. Keep the same status of the nuget dependencies
		return
	case 0:
		// all csproj files set RestoreLockedMode, update the dependency pinning status of all nuget dependencies to pinned
		pinAllNugetDependencies(unpinnedNugetDependencies)
	default:
		// only some csproj files are locked, keep the same status of the nuget dependencies but create a remediation
		for i := range unpinnedNugetDependencies {
			(unpinnedNugetDependencies)[i].Remediation.Text = (unpinnedNugetDependencies)[i].Remediation.Text +
				": some of your csproj files set the RestoreLockedMode property to true, " +
				"while other do not set it: " + unlockedPath
		}
	}
}

func pinAllNugetDependencies(dependencies []*checker.Dependency) {
	for i := range dependencies {
		if dependencies[i].Type == checker.DependencyUseTypeNugetCommand {
			dependencies[i].Pinned = asBoolPointer(true)
			dependencies[i].Remediation = nil
		}
	}
}

func retrieveCsprojConfig(c *checker.CheckRequest, nugetPostProcessData *nugetPostProcessData) error {
	if err := fileparser.OnMatchingFileContentDo(c.RepoClient, fileparser.PathMatcher{
		Pattern:       "*.csproj",
		CaseSensitive: false,
	}, analyseCsprojLockedMode, &nugetPostProcessData.CsprojConfigs, c.Dlogger); err != nil {
		return err
	}

	return nil
}

func analyseCsprojLockedMode(path string, content []byte, args ...interface{}) (bool, error) {
	pdata, ok := args[0].(*[]dotnetCsprojLockedData)
	if !ok {
		// panic if it is not correct type
		panic(fmt.Sprintf("expected type *[]dotnetCsprojLockedData, got %v", reflect.TypeOf(args[0])))
	}

	pinned, err := csproj.IsRestoreLockedModeEnabled(content)
	if err != nil {
		dl, ok := args[1].(checker.DetailLogger)
		if !ok {
			// panic if it is not correct type
			panic(fmt.Sprintf("expected type checker.DetailLogger, got %v", reflect.TypeOf(args[1])))
		}

		dl.Warn(&checker.LogMessage{
			Text: fmt.Sprintf("malformed csproj file: %v", err),
		})
		return true, nil
	}

	csprojData := dotnetCsprojLockedData{
		Path:          path,
		LockedModeSet: pinned,
	}

	*pdata = append(*pdata, csprojData)
	return true, nil
}

func countUnlocked(csprojFiles []dotnetCsprojLockedData) (int, string) {
	var unlockedPaths []string

	for i := range csprojFiles {
		if !csprojFiles[i].LockedModeSet {
			unlockedPaths = append(unlockedPaths, csprojFiles[i].Path)
		}
	}
	return len(unlockedPaths), strings.Join(unlockedPaths, ", ")
}

func countUnfixedVersions(packages []properties.NugetPackage) (int, string) {
	var unfixedVersions []string

	for i := range packages {
		if !packages[i].IsFixed {
			unfixedVersions = append(unfixedVersions, packages[i].Version)
		}
	}
	return len(unfixedVersions), strings.Join(unfixedVersions, ", ")
}

func dataAsPinnedDependenciesPointer(data interface{}) *checker.PinningDependenciesData {
	pdata, ok := data.(*checker.PinningDependenciesData)
	if !ok {
		// panic if it is not correct type
		panic(fmt.Sprintf("expected type PinningDependenciesData, got %v", reflect.TypeOf(data)))
	}
	return pdata
}

func collectShellScriptInsecureDownloads(c *checker.CheckRequest, r *checker.PinningDependenciesData) error {
	return fileparser.OnMatchingFileContentDo(c.RepoClient, fileparser.PathMatcher{
		Pattern:       "*",
		CaseSensitive: false,
	}, validateShellScriptIsFreeOfInsecureDownloads, r)
}

var validateShellScriptIsFreeOfInsecureDownloads fileparser.DoWhileTrueOnFileContent = func(
	pathfn string,
	content []byte,
	args ...interface{},
) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf(
			"validateShellScriptIsFreeOfInsecureDownloads requires exactly 1 arguments: got %v: %w",
			len(args), errInvalidArgLength)
	}

	pdata := dataAsPinnedDependenciesPointer(args[0])

	// Validate the file type.
	if !isSupportedShellScriptFile(pathfn, content) {
		return true, nil
	}

	if err := validateShellFile(pathfn, 0, 0, content, map[string]bool{}, pdata); err != nil {
		return false, nil
	}

	return true, nil
}

func collectDockerfileInsecureDownloads(c *checker.CheckRequest, r *checker.PinningDependenciesData) error {
	return fileparser.OnMatchingFileContentDo(c.RepoClient, fileparser.PathMatcher{
		Pattern:       "*Dockerfile*",
		CaseSensitive: false,
	}, validateDockerfileInsecureDownloads, r)
}

func fileIsInVendorDir(pathfn string) bool {
	cleanedPath := filepath.Clean(pathfn)
	splitCleanedPath := strings.Split(cleanedPath, "/")

	for _, d := range splitCleanedPath {
		if strings.EqualFold(d, "vendor") {
			return true
		}
		if strings.EqualFold(d, "third_party") {
			return true
		}
	}
	return false
}

var validateDockerfileInsecureDownloads fileparser.DoWhileTrueOnFileContent = func(
	pathfn string,
	content []byte,
	args ...interface{},
) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf(
			"validateDockerfileInsecureDownloads requires exactly 1 arguments: got %v: %w",
			len(args), errInvalidArgLength)
	}

	if fileIsInVendorDir(pathfn) {
		return true, nil
	}

	pdata := dataAsPinnedDependenciesPointer(args[0])

	// Return early if this is not a docker file.
	if !isDockerfile(pathfn, content) {
		return true, nil
	}

	if !fileparser.CheckFileContainsCommands(content, "#") {
		return true, nil
	}

	contentReader := bytes.NewReader(content)
	res, err := parser.Parse(contentReader)
	if err != nil {
		return false, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("%v: %v", errInternalInvalidDockerFile, err))
	}

	// Walk the Dockerfile's AST.
	taintedFiles := make(map[string]bool)
	for i := range res.AST.Children {

		child := res.AST.Children[i]
		cmdType := child.Value

		// Only look for the 'RUN' command.
		if cmdType != "RUN" {
			continue
		}

		if len(child.Heredocs) > 0 {
			startOffset := 1
			for _, heredoc := range child.Heredocs {
				cmd := heredoc.Content
				lineCount := startOffset + strings.Count(cmd, "\n")
				if err := validateShellFile(pathfn, uint(child.StartLine+startOffset)-1, uint(child.StartLine+lineCount)-2,
					[]byte(cmd), taintedFiles, pdata); err != nil {
					return false, err
				}
				startOffset += lineCount
			}
		} else {
			var valueList []string
			for n := child.Next; n != nil; n = n.Next {
				valueList = append(valueList, n.Value)
			}

			if len(valueList) == 0 {
				return false, sce.WithMessage(sce.ErrScorecardInternal, errInternalInvalidDockerFile.Error())
			}

			// Build a file content.
			cmd := strings.Join(valueList, " ")
			if err := validateShellFile(pathfn, uint(child.StartLine)-1, uint(child.EndLine)-1,
				[]byte(cmd), taintedFiles, pdata); err != nil {
				return false, err
			}
		}
	}

	return true, nil
}

func isDockerfile(pathfn string, content []byte) bool {
	if strings.HasSuffix(pathfn, ".go") ||
		strings.HasSuffix(pathfn, ".c") ||
		strings.HasSuffix(pathfn, ".cpp") ||
		strings.HasSuffix(pathfn, ".rs") ||
		strings.HasSuffix(pathfn, ".js") ||
		strings.HasSuffix(pathfn, ".py") ||
		strings.HasSuffix(pathfn, ".pyc") ||
		strings.HasSuffix(pathfn, ".java") ||
		isShellScriptFile(pathfn, content) {
		return false
	}
	return true
}

func collectDockerfilePinning(c *checker.CheckRequest, r *checker.PinningDependenciesData) error {
	err := fileparser.OnMatchingFileContentDo(c.RepoClient, fileparser.PathMatcher{
		Pattern:       "*Dockerfile*",
		CaseSensitive: false,
	}, validateDockerfilesPinning, r)
	if err != nil {
		return err
	}

	applyDockerfilePinningRemediations(r.Dependencies)
	return nil
}

func applyDockerfilePinningRemediations(d []checker.Dependency) {
	for i := range d {
		rr := &d[i]
		if rr.Type == checker.DependencyUseTypeDockerfileContainerImage && !*rr.Pinned {
			remediate := remediation.CreateDockerfilePinningRemediation(rr, remediation.CraneDigester{})
			rr.Remediation = remediate
		}
	}
}

var validateDockerfilesPinning fileparser.DoWhileTrueOnFileContent = func(
	pathfn string,
	content []byte,
	args ...interface{},
) (bool, error) {
	// Users may use various names, e.g.,
	// Dockerfile.aarch64, Dockerfile.template, Dockerfile_template, dockerfile, Dockerfile-name.template

	if len(args) != 1 {
		return false, fmt.Errorf(
			"validateDockerfilesPinning requires exactly 2 arguments: got %v: %w", len(args), errInvalidArgLength)
	}

	if fileIsInVendorDir(pathfn) {
		return true, nil
	}

	pdata := dataAsPinnedDependenciesPointer(args[0])

	// Return early if this is not a dockerfile.
	if !isDockerfile(pathfn, content) {
		return true, nil
	}

	if !fileparser.CheckFileContainsCommands(content, "#") {
		return true, nil
	}

	if fileparser.IsTemplateFile(pathfn) {
		return true, nil
	}

	// We have what looks like a docker file.
	// Let's interpret the content as utf8-encoded strings.
	contentReader := bytes.NewReader(content)
	// The dependency must be pinned by sha256 hash, e.g.,
	// FROM something@sha256:${ARG},
	// FROM something:@sha256:45b23dee08af5e43a7fea6c4cf9c25ccf269ee113168c19722f87876677c5cb2
	regex := regexp.MustCompile(`.*@sha256:([a-f\d]{64}|\${.*})`)

	pinnedAsNames := make(map[string]bool)
	res, err := parser.Parse(contentReader)
	if err != nil {
		return false, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("%v: %v", errInternalInvalidDockerFile, err))
	}

	for _, child := range res.AST.Children {
		cmdType := child.Value
		if cmdType != "FROM" {
			continue
		}

		var valueList []string
		for n := child.Next; n != nil; n = n.Next {
			valueList = append(valueList, n.Value)
		}

		switch {
		// scratch is no-op.
		case len(valueList) > 0 && strings.EqualFold(valueList[0], "scratch"):
			if len(valueList) == 3 && strings.EqualFold(valueList[1], "as") {
				pinnedAsNames[valueList[2]] = true
			}
			continue

		// FROM name AS newname.
		case len(valueList) == 3 && strings.EqualFold(valueList[1], "as"):
			name := valueList[0]
			asName := valueList[2]
			// Check if the name is pinned.
			// (1): name = <>@sha245:hash
			// (2): name = XXX where XXX was pinned
			pinned := pinnedAsNames[name]
			// Record the asName.
			if pinned || regex.MatchString(name) {
				pinnedAsNames[asName] = true
			} else {
				pinnedAsNames[asName] = false
			}

			pdata.Dependencies = append(pdata.Dependencies,
				checker.Dependency{
					Location: &checker.File{
						Path:      pathfn,
						Type:      finding.FileTypeSource,
						Offset:    uint(child.StartLine),
						EndOffset: uint(child.EndLine),
						Snippet:   child.Original,
					},
					Name:     asPointer(name),
					PinnedAt: asPointer(asName),
					Pinned:   asBoolPointer(pinnedAsNames[asName]),
					Type:     checker.DependencyUseTypeDockerfileContainerImage,
				},
			)

		// FROM name.
		case len(valueList) == 1:
			name := valueList[0]
			pinned := pinnedAsNames[name]

			dep := checker.Dependency{
				Location: &checker.File{
					Path:      pathfn,
					Type:      finding.FileTypeSource,
					Offset:    uint(child.StartLine),
					EndOffset: uint(child.EndLine),
					Snippet:   child.Original,
				},
				Pinned: asBoolPointer(pinned || regex.MatchString(name)),
				Type:   checker.DependencyUseTypeDockerfileContainerImage,
			}
			parts := strings.SplitN(name, ":", 2)
			if len(parts) > 0 {
				dep.Name = asPointer(parts[0])
				if len(parts) > 1 {
					dep.PinnedAt = asPointer(parts[1])
				}
			}
			pdata.Dependencies = append(pdata.Dependencies, dep)
		default:
			// That should not happen.
			return false, sce.WithMessage(sce.ErrScorecardInternal, errInternalInvalidDockerFile.Error())
		}
	}

	//nolint:lll
	// The file need not have a FROM statement,
	// https://github.com/tensorflow/tensorflow/blob/master/tensorflow/tools/dockerfiles/partials/jupyter.partial.Dockerfile.

	return true, nil
}

func collectGitHubWorkflowScriptInsecureDownloads(c *checker.CheckRequest, r *checker.PinningDependenciesData) error {
	return fileparser.OnMatchingFileContentDo(c.RepoClient, fileparser.PathMatcher{
		Pattern:       ".github/workflows/*",
		CaseSensitive: false,
	}, validateGitHubWorkflowIsFreeOfInsecureDownloads, r)
}

// validateGitHubWorkflowIsFreeOfInsecureDownloads checks if the workflow file downloads dependencies that are unpinned.
// Returns true if the check should continue executing after this file.
var validateGitHubWorkflowIsFreeOfInsecureDownloads fileparser.DoWhileTrueOnFileContent = func(
	pathfn string,
	content []byte,
	args ...interface{},
) (bool, error) {
	if !fileparser.IsWorkflowFile(pathfn) {
		return true, nil
	}

	if len(args) != 1 {
		return false, fmt.Errorf(
			"validateGitHubWorkflowIsFreeOfInsecureDownloads requires exactly 1 arguments: got %v: %w",
			len(args), errInvalidArgLength)
	}

	pdata := dataAsPinnedDependenciesPointer(args[0])
	if !fileparser.CheckFileContainsCommands(content, "#") {
		return true, nil
	}

	workflow, errs := actionlint.Parse(content)
	if len(errs) > 0 && workflow == nil {
		// actionlint is a linter, so it will return errors when the yaml file does not meet its linting standards.
		// Often we don't care about these errors.
		return false, fileparser.FormatActionlintError(errs)
	}

	githubVarRegex := regexp.MustCompile(`{{[^{}]*}}`)
	for jobName, job := range workflow.Jobs {
		if len(fileparser.GetJobName(job)) > 0 {
			jobName = fileparser.GetJobName(job)
		}
		taintedFiles := make(map[string]bool)

		for _, step := range job.Steps {
			if !fileparser.IsStepExecKind(step, actionlint.ExecKindRun) {
				continue
			}

			execRun, ok := step.Exec.(*actionlint.ExecRun)
			if !ok {
				stepName := fileparser.GetStepName(step)
				return false, sce.WithMessage(sce.ErrScorecardInternal,
					fmt.Sprintf("unable to parse step '%v' for job '%v'", jobName, stepName))
			}

			if execRun == nil || execRun.Run == nil {
				// Cannot check further, continue.
				continue
			}

			run := execRun.Run.Value
			// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstepsrun.
			shell, err := fileparser.GetShellForStep(step, job)
			if err != nil {
				var elementError *checker.ElementError
				if errors.As(err, &elementError) {
					// Add the workflow name and step ID to the element
					lineStart := uint(step.Pos.Line)
					elementError.Location = finding.Location{
						Path:      pathfn,
						Snippet:   elementError.Location.Snippet,
						LineStart: &lineStart,
						Type:      finding.FileTypeSource,
					}

					pdata.ProcessingErrors = append(pdata.ProcessingErrors, *elementError)

					// continue instead of break because other `run` steps may declare
					// a valid shell we can scan
					continue
				}
				return false, err
			}
			// Skip unsupported shells. We don't support Windows shells or some Unix shells.
			if !isSupportedShell(shell) {
				continue
			}

			// We replace the `${{ github.variable }}` to avoid shell parsing failures.
			script := githubVarRegex.ReplaceAll([]byte(run), []byte("GITHUB_REDACTED_VAR"))
			if err := validateShellFile(pathfn, uint(execRun.Run.Pos.Line), uint(execRun.Run.Pos.Line),
				script, taintedFiles, pdata); err != nil {
				pdata.Dependencies = append(pdata.Dependencies, checker.Dependency{
					Msg: asPointer(err.Error()),
				})
			}
		}
	}

	return true, nil
}

// Check pinning of github actions in workflows.
func collectGitHubActionsWorkflowPinning(c *checker.CheckRequest, r *checker.PinningDependenciesData) error {
	err := fileparser.OnMatchingFileContentDo(c.RepoClient, fileparser.PathMatcher{
		Pattern:       ".github/workflows/*",
		CaseSensitive: true,
	}, validateGitHubActionWorkflow, r)
	if err != nil {
		return err
	}
	//nolint:errcheck
	remediationMetadata, _ := remediation.New(c)

	applyWorkflowPinningRemediations(remediationMetadata, r.Dependencies)
	return nil
}

func applyWorkflowPinningRemediations(rm *remediation.RemediationMetadata, d []checker.Dependency) {
	for i := range d {
		rr := &d[i]
		if rr.Type == checker.DependencyUseTypeGHAction && !*rr.Pinned {
			remediate := rm.CreateWorkflowPinningRemediation(rr.Location.Path)
			rr.Remediation = remediate
		}
	}
}

// validateGitHubActionWorkflow checks if the workflow file contains unpinned actions. Returns true if the check
// should continue executing after this file.
var validateGitHubActionWorkflow fileparser.DoWhileTrueOnFileContent = func(
	pathfn string,
	content []byte,
	args ...interface{},
) (bool, error) {
	if !fileparser.IsWorkflowFile(pathfn) {
		return true, nil
	}

	if len(args) != 1 {
		return false, fmt.Errorf(
			"validateGitHubActionWorkflow requires exactly 1 arguments: got %v: %w", len(args), errInvalidArgLength)
	}
	pdata := dataAsPinnedDependenciesPointer(args[0])

	if !fileparser.CheckFileContainsCommands(content, "#") {
		return true, nil
	}

	workflow, errs := actionlint.Parse(content)
	if len(errs) > 0 && workflow == nil {
		// actionlint is a linter, so it will return errors when the yaml file does not meet its linting standards.
		// Often we don't care about these errors.
		return false, fileparser.FormatActionlintError(errs)
	}

	for jobName, job := range workflow.Jobs {
		if len(fileparser.GetJobName(job)) > 0 {
			jobName = fileparser.GetJobName(job)
		}

		if job.WorkflowCall != nil && job.WorkflowCall.Uses != nil {
			//nolint:lll
			// Check whether this is an action defined in the same repo,
			// https://docs.github.com/en/actions/learn-github-actions/finding-and-customizing-actions#referencing-an-action-in-the-same-repository-where-a-workflow-file-uses-the-action.
			if !strings.HasPrefix(job.WorkflowCall.Uses.Value, "./") {
				dep := newGHActionDependency(job.WorkflowCall.Uses.Value, pathfn, job.WorkflowCall.Uses.Pos.Line)
				pdata.Dependencies = append(pdata.Dependencies, dep)
			}
		}

		for _, step := range job.Steps {
			if !fileparser.IsStepExecKind(step, actionlint.ExecKindAction) {
				continue
			}

			execAction, ok := step.Exec.(*actionlint.ExecAction)
			if !ok {
				stepName := fileparser.GetStepName(step)
				return false, sce.WithMessage(sce.ErrScorecardInternal,
					fmt.Sprintf("unable to parse step '%v' for job '%v'", jobName, stepName))
			}

			if execAction == nil || execAction.Uses == nil {
				// Cannot check further, continue.
				continue
			}

			//nolint:lll
			// Check whether this is an action defined in the same repo,
			// https://docs.github.com/en/actions/learn-github-actions/finding-and-customizing-actions#referencing-an-action-in-the-same-repository-where-a-workflow-file-uses-the-action.
			if strings.HasPrefix(execAction.Uses.Value, "./") {
				continue
			}
			dep := newGHActionDependency(execAction.Uses.Value, pathfn, execAction.Uses.Pos.Line)
			pdata.Dependencies = append(pdata.Dependencies, dep)
		}
	}

	return true, nil
}

func newGHActionDependency(uses, pathfn string, line int) checker.Dependency {
	dep := checker.Dependency{
		Location: &checker.File{
			Path:      pathfn,
			Type:      finding.FileTypeSource,
			Offset:    uint(line),
			EndOffset: uint(line), // `Uses` always span a single line.
			Snippet:   uses,
		},
		Pinned: asBoolPointer(isActionDependencyPinned(uses)),
		Type:   checker.DependencyUseTypeGHAction,
	}
	parts := strings.SplitN(uses, "@", 2)
	if len(parts) > 0 {
		dep.Name = asPointer(parts[0])
		if len(parts) > 1 {
			dep.PinnedAt = asPointer(parts[1])
		}
	}
	return dep
}

func isActionDependencyPinned(actionUses string) bool {
	localActionRegex := regexp.MustCompile(`^\..+[^/]`)
	if localActionRegex.MatchString(actionUses) {
		return true
	}

	publicActionRegex := regexp.MustCompile(`.*@[a-fA-F\d]{40,}`)
	if publicActionRegex.MatchString(actionUses) {
		return true
	}

	dockerhubActionRegex := regexp.MustCompile(`docker://.*@sha256:[a-fA-F\d]{64}`)
	return dockerhubActionRegex.MatchString(actionUses)
}
