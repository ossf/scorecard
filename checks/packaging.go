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
	"strings"

	"github.com/rhysd/actionlint"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/fileparser"
	sce "github.com/ossf/scorecard/v4/errors"
)

// CheckPackaging is the registered name for Packaging.
const CheckPackaging = "Packaging"

//nolint:gochecknoinits
func init() {
	registerCheck(CheckPackaging, Packaging)
}

func isGithubWorkflowFile(filename string) (bool, error) {
	return strings.HasPrefix(strings.ToLower(filename), ".github/workflows"), nil
}

// Packaging runs Packaging check.
func Packaging(c *checker.CheckRequest) checker.CheckResult {
	matchedFiles, err := c.RepoClient.ListFiles(isGithubWorkflowFile)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("RepoClient.ListFiles: %v", err))
		return checker.CreateRuntimeErrorResult(CheckPackaging, e)
	}

	for _, fp := range matchedFiles {
		fc, err := c.RepoClient.GetFileContent(fp)
		if err != nil {
			e := sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("RepoClient.GetFileContent: %v", err))
			return checker.CreateRuntimeErrorResult(CheckPackaging, e)
		}

		workflow, errs := actionlint.Parse(fc)
		if len(errs) > 0 && workflow == nil {
			e := fileparser.FormatActionlintError(errs)
			return checker.CreateRuntimeErrorResult(CheckPackaging, e)
		}
		if !isPackagingWorkflow(workflow, fp, c.Dlogger) {
			continue
		}

		runs, err := c.RepoClient.ListSuccessfulWorkflowRuns(filepath.Base(fp))
		if err != nil {
			e := sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("Client.Actions.ListWorkflowRunsByFileName: %v", err))
			return checker.CreateRuntimeErrorResult(CheckPackaging, e)
		}
		if len(runs) > 0 {
			c.Dlogger.Info3(&checker.LogMessage{
				Path:   fp,
				Type:   checker.FileTypeSource,
				Offset: checker.OffsetDefault,
				Text:   fmt.Sprintf("GitHub publishing workflow used in run %s", runs[0].URL),
			})
			return checker.CreateMaxScoreResult(CheckPackaging,
				"publishing workflow detected")
		}
		c.Dlogger.Debug3(&checker.LogMessage{
			Path:   fp,
			Type:   checker.FileTypeSource,
			Offset: checker.OffsetDefault,
			Text:   "GitHub publishing workflow not used in runs",
		})
	}

	c.Dlogger.Warn3(&checker.LogMessage{
		Text: "no GitHub publishing workflow detected",
	})

	return checker.CreateInconclusiveResult(CheckPackaging,
		"no published package detected")
}

// A packaging workflow.
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

	for _, job := range workflow.Jobs {
		for _, matcher := range jobMatchers {
			if !matcher.Matches(job) {
				continue
			}

			dl.Info3(&checker.LogMessage{
				Path:   fp,
				Type:   checker.FileTypeSource,
				Offset: fileparser.GetLineNumber(job.Pos),
				Text:   matcher.LogText,
			})
			return true
		}
	}

	dl.Debug3(&checker.LogMessage{
		Path:   fp,
		Type:   checker.FileTypeSource,
		Offset: checker.OffsetDefault,
		Text:   "not a publishing workflow",
	})
	return false
}
