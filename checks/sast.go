// Copyright 2020 Security Scorecard Authors
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
	"path/filepath"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/fileparser"
	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/rhysd/actionlint"
)

// CheckSAST is the registered name for SAST.
const CheckSAST = "SAST"

var allowedConclusions = map[string]bool{"success": true, "neutral": true}

type sastMatcher struct {
	// Name to display in results.
	display string
	// Slug of the app used by the tool to run.
	slug string
	// The GitHub action to look for.
	// Only valid if the slug is "github-actions".
	action string
	// TODO: support commands.
}

type sastCategory int

const (
	sastCategoryNone = iota
	sastCategoryLinter
	sastCategorySupplyChain
	sastCategoryCodeAnalysis
)

const (
	linterScoreMax                      = 1
	supplyChainScoreMax                 = 1
	supplyChainScoreAllMergeRequestsMax = 1
	codeAnalysisScoreMax                = 5
	codeAnalysisAllMergeRequestsMax     = 2
)

var sastTools = map[sastCategory][]sastMatcher{
	sastCategoryLinter: {
		// May be run via dockerfile `docker://github/super-linter:v3`,
		// or action `github/super-linter`.
		// See https://github.com/github/super-linter.
		// Example run: https://api.github.com/repos/systemd/systemd/check-runs/4744934229
		{
			display: "super-linter",
			slug:    "github-actions",
			action:  "github/super-linter",
		},
		{
			display: "megalinter",
			slug:    "github-actions",
			action:  "megalinter/megalinter",
		},
	},
	sastCategorySupplyChain: {
		// Example run: https://api.github.com/repos/laurentsimon/scorecard-action-test-2/check-runs/4743724614
		{
			display: "Scorecard",
			slug:    "github-actions",
			action:  "ossf/scorecard-action",
		},
	},
	sastCategoryCodeAnalysis: {
		{
			display: "lgtm.com",
			slug:    "lgtm-com",
		},
		{
			display: "Sonarcloud",
			slug:    "sonarcloud",
		},
		{
			display: "CodeQL",
			slug:    "github-actions",
			action:  "github/codeql-action/analyze",
		},
	},
}

type (
	// List SAST categories.
	categoryResults map[sastCategory]sastMatcher
	// List SAST categories for each pull request number.
	prResults map[int]categoryResults
	// Lists categories for each workflow.
	workflowCategories map[int64]map[sastCategory]sastMatcher
	// List app categories.
	appCategories map[string]map[sastCategory]sastMatcher
)

//nolint:gochecknoinits
func init() {
	if err := registerCheck(CheckSAST, SAST); err != nil {
		// This should never happen.
		panic(err)
	}
}

// SAST runs SAST check.
func SAST(c *checker.CheckRequest) checker.CheckResult {
	workflowsCat, err := readDefinedWorkflowCategories(c)
	// fmt.Println(workflowsCat)
	if err != nil {
		return checker.CreateRuntimeErrorResult(CheckSAST, sce.WithMessage(sce.ErrScorecardInternal, err.Error()))
	}

	appsCat, err := readAppCategories(c)
	// fmt.Println(appsCat)
	if err != nil {
		return checker.CreateRuntimeErrorResult(CheckSAST, sce.WithMessage(sce.ErrScorecardInternal, err.Error()))
	}

	// Check SAST tools run on pull requests.
	merged, results, err := readSuccessfulSastRunsInPRs(c, workflowsCat, appsCat)
	if err != nil {
		return checker.CreateRuntimeErrorResult(CheckSAST, sce.WithMessage(sce.ErrScorecardInternal, err.Error()))
	}

	// Score calculation.
	mrLinterScore := computeScore(results, sastCategoryLinter, float64(merged), c.Dlogger) * linterScoreMax
	mrSupplyChainScore := computeScore(results, sastCategorySupplyChain, float64(merged), c.Dlogger) * supplyChainScoreMax
	mrCodeAnalysisScore := computeScore(results, sastCategoryCodeAnalysis, float64(merged), c.Dlogger) * codeAnalysisScoreMax
	score := float64(0)
	linterUsed := false
	supplyChainUsed := false
	staticAnalysisUsed := false

	// 1. Linter.
	// A linter runs quickly and can be enabled on all merge requests.
	// linterScoreMax points are awarded.
	if mrLinterScore == linterScoreMax {
		linterUsed = true
	}
	score += float64(mrLinterScore)

	// 2. Supply chain tools.
	// supplyChainScoreMax points are awarded if a tool is used.
	// supplyChainScoreAllMergeRequestsMax bonus points awarded if a tool is run
	// on all merge commits.
	if isToolUsed(workflowsCat, sastCategorySupplyChain, mrSupplyChainScore) {
		supplyChainUsed = true
		score += supplyChainScoreMax
		score += supplyChainScoreAllMergeRequestsMax * mrSupplyChainScore / supplyChainScoreMax
	}

	// 3. static analysis.
	// codeAnalysisScoreMax points are awarded if a tool is used.
	// codeAnalysisAllMergeRequests bonus points awarded if a tool is run
	// on all merge commits.
	if isToolUsed(workflowsCat, sastCategoryCodeAnalysis, mrCodeAnalysisScore) {
		staticAnalysisUsed = true
		score += codeAnalysisScoreMax
		score += codeAnalysisAllMergeRequestsMax * mrCodeAnalysisScore / codeAnalysisScoreMax
	}

	if score == checker.MaxResultScore {
		return checker.CreateMaxScoreResult(CheckSAST, "SAST tools of each category are used")
	}

	if score == checker.MinResultScore {
		return checker.CreateMinScoreResult(CheckSAST, "no SAST tools detected")
	}

	switch {
	case linterUsed && staticAnalysisUsed && supplyChainUsed:
		return checker.CreateResultWithScore(CheckSAST,
			fmt.Sprintf("SAST tools used but %s and/or %s categories of tools are not run on all commits",
				categoryToString(sastCategoryCodeAnalysis), categoryToString(sastCategorySupplyChain)),
			int(score))
	default:
		return checker.CreateResultWithScore(CheckSAST,
			fmt.Sprintf("SAST category of tools are not used",
				categoryToString(sastCategoryCodeAnalysis), categoryToString(sastCategorySupplyChain)),
			int(score))
	}
}

func isToolUsed(cats workflowCategories, c sastCategory, mrScore float64) bool {
	return isToolUsedInWorkflows(cats, c) ||
		mrScore > 0
}

func isToolUsedInWorkflows(cats workflowCategories, c sastCategory) bool {
	for _, cats := range cats {
		if _, exists := cats[sastCategoryCodeAnalysis]; exists {
			return true
		}
	}
	return false
}

func readDefinedWorkflowCategories(c *checker.CheckRequest) (workflowCategories, error) {
	wcat := make(map[int64]map[sastCategory]sastMatcher, 0)
	matchedFiles, err := c.RepoClient.ListFiles(fileparser.IsGithubWorkflowFileCb)
	if err != nil {
		return wcat,
			sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("RepoClient.ListFiles: %v", err))
	}

	// Get the list of SAST defined thru an action.
	actions := listSastActionsOfInterest()
	for _, fn := range matchedFiles {

		fc, err := c.RepoClient.GetFileContent(fn)
		if err != nil {
			return wcat,
				sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("RepoClient.GetFileContent: %v", err))
		}

		workflow, errs := actionlint.Parse(fc)
		if len(errs) > 0 && workflow == nil {
			return wcat, fileparser.FormatActionlintError(errs)
		}

		workflowActions := listDefinedSastActions(workflow, actions, fn, c.Dlogger)
		if len(workflowActions) == 0 {
			continue
		}

		// List workflow information.
		ghWorkflow, err := c.RepoClient.GetWorkflowByFileName(filepath.Base(fn))
		if err != nil {
			return wcat,
				sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("Client.Actions.ListWorkflowRunsByFileName: %v", err))
		}

		if _, exists := wcat[ghWorkflow.ID]; !exists {
			wcat[ghWorkflow.ID] = make(map[sastCategory]sastMatcher)
		}

		for act := range workflowActions {
			cats := readActionCategories(act)
			wcat[ghWorkflow.ID] = cats
			for k, v := range cats {
				c.Dlogger.Info3(&checker.LogMessage{
					Path: fn,
					Type: checker.FileTypeSource,
					Text: fmt.Sprintf("%s tool '%s' detected in workflow", categoryToString(k), v.display),
				})
			}

		}

	}

	return wcat, nil
}

func readAppCategories(c *checker.CheckRequest) (appCategories, error) {
	a := make(map[string]map[sastCategory]sastMatcher, 0)
	for k, v := range sastTools {
		for _, m := range v {
			if m.slug == "github-code-scanning" {
				return a,
					sce.WithMessage(sce.ErrScorecardInternal, "unexpected slug 'github-code-scanning'")
			}
			if m.slug == "github-actions" {
				continue
			}

			// fmt.Println(m.slug, k)
			if _, exists := a[m.slug]; !exists {
				a[m.slug] = make(map[sastCategory]sastMatcher, 0)
			}
			a[m.slug][k] = m

		}
	}
	return a, nil
}

func readActionCategories(action string) map[sastCategory]sastMatcher {
	a := make(map[sastCategory]sastMatcher, 0)
	for k, v := range sastTools {
		for _, m := range v {
			if m.slug == "github-actions" &&
				m.action == action {
				// fmt.Println(action, k)
				a[k] = m
			}
		}
	}
	return a
}

func listSastActionsOfInterest() map[string]bool {
	a := make(map[string]bool, 0)
	for _, v := range sastTools {
		for _, m := range v {
			if m.slug == "github-actions" {
				// TODO: get category
				a[m.action] = true
			}
		}
	}
	return a
}

func listDefinedSastActions(workflow *actionlint.Workflow, acts map[string]bool,
	fp string, dl checker.DetailLogger,
) map[string]bool {
	r := make(map[string]bool, 0)
	jobMatchers := []fileparser.JobMatcher{}
	names := make(map[string]string)
	for k := range acts {
		s := fileparser.JobMatcherStep{Uses: k}
		m := fileparser.JobMatcher{
			LogText: fmt.Sprintf("SAST action '%s' found", k),
			Steps:   []*fileparser.JobMatcherStep{&s},
		}
		jobMatchers = append(jobMatchers, m)
		names[fmt.Sprintf("%+v", m)] = k
	}

	for _, job := range workflow.Jobs {
		for _, matcher := range jobMatchers {
			if !matcher.Matches(job) {
				continue
			}

			dl.Debug3(&checker.LogMessage{
				Path:   fp,
				Type:   checker.FileTypeSource,
				Offset: fileparser.GetLineNumber(job.Pos),
				Text:   matcher.LogText,
			})

			n, _ := names[fmt.Sprintf("%+v", matcher)]
			r[n] = true
		}
	}

	dl.Debug3(&checker.LogMessage{
		Path:   fp,
		Type:   checker.FileTypeSource,
		Offset: checker.OffsetDefault,
		Text:   "no SAST action found in workflow",
	})
	return r
}

// nolint
func readSuccessfulSastRunsInPRs(c *checker.CheckRequest,
	wcats workflowCategories, acats appCategories,
) (int, prResults, error) {
	results := make(prResults, 0)
	totalMerged := 0

	commits, err := c.RepoClient.ListCommits()
	if err != nil {
		//nolint
		return totalMerged, results,
			sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("RepoClient.ListCommits: %v", err))
	}

	for _, commit := range commits {
		pr := commit.AssociatedMergeRequest
		if pr.MergedAt.IsZero() {
			continue
		}

		totalMerged++
		crs, err := c.RepoClient.ListCheckRunsForRef(pr.HeadSHA)
		if err != nil {
			return totalMerged, results,
				sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("Client.Checks.ListCheckRunsForRef: %v", err))
		}
		// Note: crs may be `nil`: in this case
		// the loop below will be skipped.
		// fmt.Println(pr.Number)
		for _, cr := range crs {
			// fmt.Println()
			// fmt.Println(" ", cr.Status, cr.Conclusion, cr.App.Slug, cr.Name)
			if cr.Status != "completed" {
				continue
			}
			if !allowedConclusions[cr.Conclusion] {
				continue
			}
			// fmt.Println("", cr.App.Slug, cr.Name, *cr.CheckSuiteID)
			category, name, err := getSastCategory(c, &cr, wcats, acats)
			if err != nil {
				return totalMerged, results, err
			}
			if category != sastCategoryNone {
				if _, exists := results[pr.Number]; !exists {
					results[pr.Number] = categoryResults{
						sastCategoryLinter:       sastMatcher{},
						sastCategorySupplyChain:  sastMatcher{},
						sastCategoryCodeAnalysis: sastMatcher{},
					}
				}
				results[pr.Number][category] = sastMatcher{display: name}
				c.Dlogger.Debug3(&checker.LogMessage{
					Path: cr.URL,
					Type: checker.FileTypeURL,
					Text: fmt.Sprintf("%s detected on PR#%d", name, pr.Number),
				})
			}
		}
	}
	if totalMerged == 0 {
		c.Dlogger.Warn3(&checker.LogMessage{
			Text: "no pull requests merged into dev branch",
		})
		return totalMerged, results, nil
	}
	// fmt.Println("merged:", totalMerged)
	return totalMerged, results, nil
}

func categoryToString(c sastCategory) string {
	switch c {
	case sastCategoryLinter:
		return "linter"
	case sastCategorySupplyChain:
		return "supply-chain"
	case sastCategoryCodeAnalysis:
		return "code-analysis"
	case sastCategoryNone:
		return "none"
	}
	return "none"
}

func computeScore(t prResults, c sastCategory, merged float64,
	dl checker.DetailLogger,
) float64 {
	score := float64(0)
	if merged == 0 {
		return 0
	}
	for n := range t {
		r, _ := t[n][c]
		if r.display != "" {
			score++
		}
	}
	// fmt.Println("score / merged /", c, ":", score, merged)
	if score != merged {
		dl.Warn3(&checker.LogMessage{
			Text: fmt.Sprintf("%s tool run on %d commits out of %d commits", categoryToString(c),
				int(score), int(merged)),
		})
	} else {
		dl.Info3(&checker.LogMessage{
			Text: fmt.Sprintf("%s tool run on the last %d commits", categoryToString(c), int(merged)),
		})
	}
	return score / merged
}

func getWorkflowIDForCheckRun(c *checker.CheckRequest, cr *clients.CheckRun) (int64, error) {
	// List workflow runs by CheckSuiteID.
	runs, err := c.RepoClient.ListWorkflowRuns(&clients.ListWorkflowRunOptions{CheckSuiteID: cr.CheckSuiteID})
	if err != nil {
		return 0, err
	}
	// Get the workflow ID.
	// Note: this typically only contains a single run.
	for _, r := range runs {
		if r.WorkflowID == nil {
			continue
		}
		return *r.WorkflowID, nil
	}
	return 0, nil
}

func getSastCategory(c *checker.CheckRequest, cr *clients.CheckRun,
	wcats workflowCategories, acats appCategories,
) (sastCategory, string, error) {
	if cr.App.Slug != "github-actions" {
		cats, exists := acats[cr.App.Slug]
		if !exists {
			return sastCategoryNone, "", nil
		}
		// Multiple entries may happen if we have declared the same app in multiple categories.
		if len(cats) > 1 {
			return sastCategoryNone, "",
				sce.WithMessage(sce.ErrScorecardInternal,
					fmt.Sprintf("multiple categories for an app not supported (%s)", cr.App.Slug))
		}
		// Take the first entry.
		for k := range cats {
			return k, cr.App.Slug, nil
		}

	} else {
		// GitHub actions.
		// fmt.Println("getWorkflowIDForCheckRun", *cr.CheckSuiteID)
		workflowID, err := getWorkflowIDForCheckRun(c, cr)
		if err != nil {
			return sastCategoryNone, "", err
		}

		cats, exists := wcats[workflowID]
		if !exists {
			return sastCategoryNone, "", nil
		}

		// Multiple entries may happen if we have declared the same app in multiple categories.
		if len(cats) > 1 {
			return sastCategoryNone, "",
				sce.WithMessage(sce.ErrScorecardInternal,
					fmt.Sprintf("multiple categories for an app not supported (%s)", cr.App.Slug))
		}

		// Take the first entry.
		for k, v := range cats {
			return k, v.display, nil
		}
	}

	return sastCategoryNone, "", nil
}

/*
// nolint
func toolUsedInWorkflows(c *checker.CheckRequest, name, action, category string) (int, error) {
	matchedFiles, err := c.RepoClient.ListFiles(fileparser.IsWorkflowFileCb)
	if err != nil {
		return checker.MinResultScore,
			sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("RepoClient.ListFiles: %v", err))
	}

	// fmt.Println(actions)
	for _, fn := range matchedFiles {

		if !fileparser.IsWorkflowFile(fn) {
			continue
		}

		fc, err := c.RepoClient.GetFileContent(fn)
		if err != nil {
			return checker.MinResultScore,
				sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("RepoClient.GetFileContent: %v", err))
		}

		workflow, errs := actionlint.Parse(fc)
		if len(errs) > 0 && workflow == nil {
			return checker.MinResultScore, fileparser.FormatActionlintError(errs)
		}

		actions := make(map[string]bool)
		actions[action] = true
		workflowActions := listDefinedSastActions(workflow, actions, fn, c.Dlogger)
		if len(workflowActions) == 0 {
			continue
		}

		c.Dlogger.Info3(&checker.LogMessage{
			Text: fmt.Sprintf("%s tool '%s' detected in workflow", category, name),
			Path: fn,
			// TODO: add line number.
		})
		return checker.MaxResultScore, nil
	}

	c.Dlogger.Debug3(&checker.LogMessage{
		Text: fmt.Sprintf("%s tool %s not detected as workflow", category, name),
	})
	return checker.MinResultScore, nil
}
*/
