// Copyright 2020 OpenSSF Scorecard Authors
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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/rhysd/actionlint"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/fileparser"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
)

// CheckSAST is the registered name for SAST.
const CheckSAST = "SAST"

var errInvalid = errors.New("invalid")

var sastTools = map[string]bool{"github-code-scanning": true, "lgtm-com": true, "sonarcloud": true}

var allowedConclusions = map[string]bool{"success": true, "neutral": true}

//nolint:gochecknoinits
func init() {
	if err := registerCheck(CheckSAST, SAST, nil); err != nil {
		// This should never happen.
		panic(err)
	}
}

// SAST runs SAST check.
func SAST(c *checker.CheckRequest) checker.CheckResult {
	sastScore, nonCompliantPRs, sastErr := sastToolInCheckRuns(c)
	if sastErr != nil {
		return checker.CreateRuntimeErrorResult(CheckSAST, sastErr)
	}

	codeQlScore, codeQlErr := codeQLInCheckDefinitions(c)
	if codeQlErr != nil {
		return checker.CreateRuntimeErrorResult(CheckSAST, codeQlErr)
	}
	sonarScore, sonarErr := sonarEnabled(c)
	if sonarErr != nil {
		return checker.CreateRuntimeErrorResult(CheckSAST, sonarErr)
	}

	if sonarScore == checker.MaxResultScore {
		return checker.CreateMaxScoreResult(CheckSAST, "SAST tool detected")
	}

	// Both results are inconclusive.
	// Can never happen.
	if sastScore == checker.InconclusiveResultScore &&
		codeQlScore == checker.InconclusiveResultScore {
		// That can never happen since sastToolInCheckRuns can never
		// retun checker.InconclusiveResultScore.
		return checker.CreateRuntimeErrorResult(CheckSAST, sce.ErrScorecardInternal)
	}

	// Both scores are conclusive.
	// We assume the CodeQl config uses a cron and is not enabled as pre-submit.
	// TODO: verify the above comment in code.
	// We encourage developers to have sast check run on every pre-submit rather
	// than as cron jobs thru the score computation below.
	// Warning: there is a hidden assumption that *any* sast tool is equally good.
	if sastScore != checker.InconclusiveResultScore &&
		codeQlScore != checker.InconclusiveResultScore {
		switch {
		case sastScore == checker.MaxResultScore:
			return checker.CreateMaxScoreResult(CheckSAST, "SAST tool is run on all commits")
		case codeQlScore == checker.MinResultScore:
			return checker.CreateResultWithScore(CheckSAST,
				checker.NormalizeReason("SAST tool is not run on all commits", sastScore), sastScore)

		// codeQl is enabled and sast has 0+ (but not all) PRs checks.
		case codeQlScore == checker.MaxResultScore:
			const sastWeight = 3
			const codeQlWeight = 7
			c.Dlogger.Debug(&checker.LogMessage{
				Text: getNonCompliantPRMessage(nonCompliantPRs),
			})
			score := checker.AggregateScoresWithWeight(map[int]int{sastScore: sastWeight, codeQlScore: codeQlWeight})
			return checker.CreateResultWithScore(CheckSAST, "SAST tool detected but not run on all commits", score)
		default:
			return checker.CreateRuntimeErrorResult(CheckSAST, sce.WithMessage(sce.ErrScorecardInternal, "contact team"))
		}
	}

	// Sast inconclusive.
	if codeQlScore != checker.InconclusiveResultScore {
		if codeQlScore == checker.MaxResultScore {
			return checker.CreateMaxScoreResult(CheckSAST, "SAST tool detected")
		}
		return checker.CreateMinScoreResult(CheckSAST, "no SAST tool detected")
	}

	// CodeQl inconclusive.
	if sastScore != checker.InconclusiveResultScore {
		if sastScore == checker.MaxResultScore {
			return checker.CreateMaxScoreResult(CheckSAST, "SAST tool is run on all commits")
		}

		c.Dlogger.Debug(&checker.LogMessage{
			Text: getNonCompliantPRMessage(nonCompliantPRs),
		})
		return checker.CreateResultWithScore(CheckSAST,
			checker.NormalizeReason("SAST tool is not run on all commits", sastScore), sastScore)
	}

	// Should never happen.
	return checker.CreateRuntimeErrorResult(CheckSAST, sce.WithMessage(sce.ErrScorecardInternal, "contact team"))
}

func sastToolInCheckRuns(c *checker.CheckRequest) (int, map[int]int, error) {
	commits, err := c.RepoClient.ListCommits()
	if err != nil {
		return checker.InconclusiveResultScore, nil,
			sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("RepoClient.ListCommits: %v", err))
	}

	totalMerged := 0
	totalTested := 0
	nonCompliantPRs := make(map[int]int)
	for i := range commits {
		pr := commits[i].AssociatedMergeRequest
		// TODO(#575): We ignore associated PRs if Scorecard is being run on a fork
		// but the PR was created in the original repo.
		if pr.MergedAt.IsZero() {
			continue
		}
		totalMerged++
		checked := false
		crs, err := c.RepoClient.ListCheckRunsForRef(pr.HeadSHA)
		if err != nil {
			return checker.InconclusiveResultScore, nil,
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
				c.Dlogger.Debug(&checker.LogMessage{
					Path: cr.URL,
					Type: finding.FileTypeURL,
					Text: fmt.Sprintf("tool detected: %v", cr.App.Slug),
				})
				totalTested++
				checked = true
				break
			}
		}
		if !checked {
			nonCompliantPRs[pr.Number] = pr.Number
		}
	}
	if totalMerged == 0 {
		c.Dlogger.Warn(&checker.LogMessage{
			Text: "no pull requests merged into dev branch",
		})
		return checker.InconclusiveResultScore, nil, nil
	}

	if totalTested == totalMerged {
		c.Dlogger.Info(&checker.LogMessage{
			Text: fmt.Sprintf("all commits (%v) are checked with a SAST tool", totalMerged),
		})
	} else {
		c.Dlogger.Warn(&checker.LogMessage{
			Text: fmt.Sprintf("%v commits out of %v are checked with a SAST tool", totalTested, totalMerged),
		})
	}

	return checker.CreateProportionalScore(totalTested, totalMerged), nonCompliantPRs, nil
}

func codeQLInCheckDefinitions(c *checker.CheckRequest) (int, error) {
	var workflowPaths []string
	err := fileparser.OnMatchingFileContentDo(c.RepoClient, fileparser.PathMatcher{
		Pattern:       ".github/workflows/*",
		CaseSensitive: false,
	}, searchGitHubActionWorkflowCodeQL, &workflowPaths)
	if err != nil {
		return checker.InconclusiveResultScore, err
	}

	for _, path := range workflowPaths {
		c.Dlogger.Debug(&checker.LogMessage{
			Path:   path,
			Type:   finding.FileTypeSource,
			Offset: checker.OffsetDefault,
			Text:   "CodeQL detected",
		})
	}

	// TODO: check if it's enabled as cron or presubmit.
	// TODO: check which branches it is enabled on. We should find main.
	if len(workflowPaths) > 0 {
		c.Dlogger.Info(&checker.LogMessage{
			Text: "SAST tool detected: CodeQL",
		})
		return checker.MaxResultScore, nil
	}

	c.Dlogger.Warn(&checker.LogMessage{
		Text: "CodeQL tool not detected",
	})
	return checker.MinResultScore, nil
}

// Check file content.
var searchGitHubActionWorkflowCodeQL fileparser.DoWhileTrueOnFileContent = func(path string,
	content []byte,
	args ...interface{},
) (bool, error) {
	if !fileparser.IsWorkflowFile(path) {
		return true, nil
	}

	if len(args) != 1 {
		return false, fmt.Errorf(
			"searchGitHubActionWorkflowCodeQL requires exactly 1 arguments: %w", errInvalid)
	}

	// Verify the type of the data.
	paths, ok := args[0].(*[]string)
	if !ok {
		return false, fmt.Errorf(
			"searchGitHubActionWorkflowCodeQL expects arg[0] of type *[]string: %w", errInvalid)
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
			if action == "github/codeql-action/analyze" {
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

func sonarEnabled(c *checker.CheckRequest) (int, error) {
	var config []sonarConfig
	err := fileparser.OnMatchingFileContentDo(c.RepoClient, fileparser.PathMatcher{
		Pattern:       "*",
		CaseSensitive: false,
	}, validateSonarConfig, &config)
	if err != nil {
		return checker.InconclusiveResultScore, err
	}
	for _, result := range config {
		c.Dlogger.Info(&checker.LogMessage{
			Path:      result.file.Path,
			Type:      result.file.Type,
			Offset:    result.file.Offset,
			EndOffset: result.file.EndOffset,
			Text:      "Sonar configuration detected",
			Snippet:   result.url,
		})
	}

	if len(config) > 0 {
		return checker.MaxResultScore, nil
	}

	return checker.MinResultScore, nil
}

// Check file content.
var validateSonarConfig fileparser.DoWhileTrueOnFileContent = func(pathfn string,
	content []byte,
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

	line := 0
	// https://golang.org/pkg/bufio/#Scanner.Scan
	for scanner.Scan() {
		line++
		if strings.Contains(scanner.Text(), string(data)) {
			return uint(line), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("scanner.Err(): %w", err)
	}

	return 0, nil
}

func getNonCompliantPRMessage(intMap map[int]int) string {
	var sb strings.Builder
	for _, value := range intMap {
		if len(sb.String()) != 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%d", value))
	}
	return fmt.Sprintf("List of pull requests without CI test: %s", sb.String())
}
