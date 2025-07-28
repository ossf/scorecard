// Copyright 2023 OpenSSF Scorecard Authors
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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"path"
	"regexp"
	"strings"

	"github.com/rhysd/actionlint"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/checks/fileparser"
	"github.com/ossf/scorecard/v5/clients"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
)

const CheckSAST = "SAST"

var errInvalid = errors.New("invalid")

var sastTools = map[string]bool{
	"github-advanced-security": true,
	"github-code-scanning":     true,
	"lgtm-com":                 true,
	"sonarcloud":               true,
	"sonarqubecloud":           true,
}

var allowedConclusions = map[string]bool{"success": true, "neutral": true}

// SAST checks for presence of static analysis tools.
func SAST(c *checker.CheckRequest) (checker.SASTData, error) {
	var data checker.SASTData

	commits, err := sastToolInCheckRuns(c)
	if err != nil {
		return data, err
	}
	data.Commits = commits

	codeQLWorkflows, err := getSastUsesWorkflows(c, "^github/codeql-action/analyze$", checker.CodeQLWorkflow)
	if err != nil {
		return data, err
	}
	data.Workflows = append(data.Workflows, codeQLWorkflows...)

	sonarWorkflows, err := getSonarWorkflows(c)
	if err != nil {
		return data, err
	}
	data.Workflows = append(data.Workflows, sonarWorkflows...)

	snykWorkflows, err := getSastUsesWorkflows(c, "^snyk/actions/.*", checker.SnykWorkflow)
	if err != nil {
		return data, err
	}
	data.Workflows = append(data.Workflows, snykWorkflows...)

	pysaWorkflows, err := getSastUsesWorkflows(c, "^facebook/pysa-action$", checker.PysaWorkflow)
	if err != nil {
		return data, err
	}
	data.Workflows = append(data.Workflows, pysaWorkflows...)

	qodanaWorkflows, err := getSastUsesWorkflows(c, "^JetBrains/qodana-action$", checker.QodanaWorkflow)
	if err != nil {
		return data, err
	}
	data.Workflows = append(data.Workflows, qodanaWorkflows...)

	hadolintWorkflows, err := getSastUsesWorkflows(c, "^hadolint/hadolint-action$", checker.HadolintWorkflow)
	if err != nil {
		return data, err
	}
	data.Workflows = append(data.Workflows, hadolintWorkflows...)

	return data, nil
}

func sastToolInCheckRuns(c *checker.CheckRequest) ([]checker.SASTCommit, error) {
	var sastCommits []checker.SASTCommit
	commits, err := c.RepoClient.ListCommits()
	if err != nil {
		// ignoring check for local dir
		if errors.Is(err, clients.ErrUnsupportedFeature) {
			return sastCommits, nil
		}
		return sastCommits,
			sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("RepoClient.ListCommits: %v", err))
	}

	for i := range commits {
		pr := commits[i].AssociatedMergeRequest
		// TODO(#575): We ignore associated PRs if Scorecard is being run on a fork
		// but the PR was created in the original repo.
		if pr.MergedAt.IsZero() {
			continue
		}

		checked := false
		crs, err := c.RepoClient.ListCheckRunsForRef(pr.HeadSHA)
		if err != nil {
			return sastCommits,
				sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("Client.Checks.ListCheckRunsForRef: %v", err))
		}
		// Note: crs may be `nil`: in this case
		// the loop below will be skipped.
		for _, cr := range crs {
			if cr.Status != "completed" {
				continue
			}
			if !allowedConclusions[cr.Conclusion] {
				continue
			}
			if sastTools[cr.App.Slug] {
				if c.Dlogger != nil {
					c.Dlogger.Debug(&checker.LogMessage{
						Path: cr.URL,
						Type: finding.FileTypeURL,
						Text: fmt.Sprintf("tool detected: %v", cr.App.Slug),
					})
				}
				checked = true
				break
			}
		}
		sastCommit := checker.SASTCommit{
			CommittedDate:          commits[i].CommittedDate,
			Message:                commits[i].Message,
			SHA:                    commits[i].SHA,
			AssociatedMergeRequest: commits[i].AssociatedMergeRequest,
			Committer:              commits[i].Committer,
			Compliant:              checked,
		}
		sastCommits = append(sastCommits, sastCommit)
	}
	return sastCommits, nil
}

// getSastUsesWorkflows matches if the "uses" field of a GitHub action matches
// a given regex by way of usesRegex. Each workflow that matches the usesRegex
// is appended to the slice that is returned.
func getSastUsesWorkflows(
	c *checker.CheckRequest,
	usesRegex string,
	checkerType checker.SASTWorkflowType,
) ([]checker.SASTWorkflow, error) {
	var workflowPaths []string
	var sastWorkflows []checker.SASTWorkflow
	err := fileparser.OnMatchingFileContentDo(c.RepoClient, fileparser.PathMatcher{
		Pattern:       ".github/workflows/*",
		CaseSensitive: false,
	}, searchGitHubActionWorkflowUseRegex, &workflowPaths, usesRegex)
	if err != nil {
		return sastWorkflows, err
	}
	for _, path := range workflowPaths {
		sastWorkflow := checker.SASTWorkflow{
			File: checker.File{
				Path:   path,
				Offset: checker.OffsetDefault,
				Type:   finding.FileTypeSource,
			},
			Type: checkerType,
		}

		sastWorkflows = append(sastWorkflows, sastWorkflow)
	}
	return sastWorkflows, nil
}

var searchGitHubActionWorkflowUseRegex fileparser.DoWhileTrueOnFileContent = func(path string,
	content []byte,
	args ...interface{},
) (bool, error) {
	if !fileparser.IsWorkflowFile(path) {
		return true, nil
	}

	if len(args) != 2 {
		return false, fmt.Errorf(
			"searchGitHubActionWorkflowUseRegex requires exactly 2 arguments: %w", errInvalid)
	}

	// Verify the type of the data.
	paths, ok := args[0].(*[]string)
	if !ok {
		return false, fmt.Errorf(
			"searchGitHubActionWorkflowUseRegex expects arg[0] of type *[]string: %w", errInvalid)
	}

	usesRegex, ok := args[1].(string)
	if !ok {
		return false, fmt.Errorf(
			"searchGitHubActionWorkflowUseRegex expects arg[1] of type string: %w", errInvalid)
	}

	workflow, errs := actionlint.Parse(content)
	if len(errs) > 0 && workflow == nil {
		return false, fileparser.FormatActionlintError(errs)
	}

	for _, job := range workflow.Jobs {
		for _, step := range job.Steps {
			e, ok := step.Exec.(*actionlint.ExecAction)
			if !ok || e == nil || e.Uses == nil {
				continue
			}
			// Parse out repo / SHA.
			uses := strings.TrimPrefix(e.Uses.Value, "actions://")
			action, _, _ := strings.Cut(uses, "@")
			re := regexp.MustCompile(usesRegex)
			if re.MatchString(action) {
				*paths = append(*paths, path)
			}
		}
	}
	return true, nil
}

type sonarConfig struct {
	url  string
	file checker.File
}

func getSonarWorkflows(c *checker.CheckRequest) ([]checker.SASTWorkflow, error) {
	var config []sonarConfig
	var sastWorkflows []checker.SASTWorkflow
	// in the future, we may want to use ListFiles instead, so we don't open every file
	err := fileparser.OnMatchingFileReaderDo(c.RepoClient, fileparser.PathMatcher{
		Pattern:       "*",
		CaseSensitive: false,
	}, validateSonarConfig, &config)
	if err != nil {
		return sastWorkflows, err
	}
	for _, result := range config {
		sastWorkflow := checker.SASTWorkflow{
			File: checker.File{
				Path:      result.file.Path,
				Offset:    result.file.Offset,
				EndOffset: result.file.EndOffset,
				Type:      result.file.Type,
				Snippet:   result.url,
			},
			Type: checker.SonarWorkflow,
		}

		sastWorkflows = append(sastWorkflows, sastWorkflow)
	}
	return sastWorkflows, nil
}

// Check file content.
var validateSonarConfig fileparser.DoWhileTrueOnFileReader = func(pathfn string,
	reader io.Reader,
	args ...interface{},
) (bool, error) {
	if !strings.EqualFold(path.Base(pathfn), "pom.xml") {
		return true, nil
	}

	if len(args) != 1 {
		return false, fmt.Errorf(
			"validateSonarConfig requires exactly 1 argument: %w", errInvalid)
	}

	// Verify the type of the data.
	pdata, ok := args[0].(*[]sonarConfig)
	if !ok {
		return false, fmt.Errorf(
			"validateSonarConfig expects arg[0] of type *[]sonarConfig]: %w", errInvalid)
	}

	content, err := io.ReadAll(reader)
	if err != nil {
		return false, fmt.Errorf("read file: %w", err)
	}
	regex := regexp.MustCompile(`<sonar\.host\.url>\s*(\S+)\s*<\/sonar\.host\.url>`)
	match := regex.FindSubmatch(content)

	if len(match) < 2 {
		return true, nil
	}
	offset, err := findLine(content, []byte("<sonar.host.url>"))
	if err != nil {
		return false, err
	}

	endOffset, err := findLine(content, []byte("</sonar.host.url>"))
	if err != nil {
		return false, err
	}

	*pdata = append(*pdata, sonarConfig{
		url: string(match[1]),
		file: checker.File{
			Path:      pathfn,
			Type:      finding.FileTypeSource,
			Offset:    offset,
			EndOffset: endOffset,
		},
	})

	return true, nil
}

func findLine(content, data []byte) (uint, error) {
	r := bytes.NewReader(content)
	scanner := bufio.NewScanner(r)

	var line uint
	for scanner.Scan() {
		line++
		if bytes.Contains(scanner.Bytes(), data) {
			return line, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("scanner.Err(): %w", err)
	}

	return 0, nil
}
