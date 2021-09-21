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
	"regexp"
	"strings"

	"github.com/ossf/scorecard/v2/checker"
	sce "github.com/ossf/scorecard/v2/errors"
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

		if !isPackagingWorkflow(string(fc), fp, c.Dlogger) {
			continue
		}

		runs, err := c.RepoClient.ListSuccessfulWorkflowRuns(filepath.Base(fp))
		if err != nil {
			e := sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("Client.Actions.ListWorkflowRunsByFileName: %v", err))
			return checker.CreateRuntimeErrorResult(CheckPackaging, e)
		}
		if len(runs) > 0 {
			c.Dlogger.Info3(&checker.LogMessage{
				Path: fp,
				Type: checker.FileTypeSource,
				// Source file must have line number > 0.
				Offset: 1,
				Text:   fmt.Sprintf("GitHub publishing workflow used in run %s", runs[0].URL),
			})
			return checker.CreateMaxScoreResult(CheckPackaging,
				"publishing workflow detected")
		}
		c.Dlogger.Info3(&checker.LogMessage{
			Path: fp,
			Type: checker.FileTypeSource,
			// Source file must have line number > 0.
			Offset: 1,
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
func isPackagingWorkflow(s, fp string, dl checker.DetailLogger) bool {
	// Nodejs packages.
	if strings.Contains(s, "actions/setup-node@") {
		r1 := regexp.MustCompile(`(?s)registry-url.*https://registry\.npmjs\.org`)
		r2 := regexp.MustCompile(`(?s)npm.*publish`)

		if r1.MatchString(s) && r2.MatchString(s) {
			dl.Info3(&checker.LogMessage{
				Path: fp,
				Type: checker.FileTypeSource,
				// Source file must have line number > 0.
				Offset: 1,
				Text:   "candidate node publishing workflow using npm",
			})
			return true
		}
	}

	// Java packages.
	if strings.Contains(s, "actions/setup-java@") {
		// Java packages with maven.
		r1 := regexp.MustCompile(`(?s)mvn.*deploy`)
		if r1.MatchString(s) {
			dl.Info3(&checker.LogMessage{
				Path: fp,
				Type: checker.FileTypeSource,
				// Source file must have line number > 0.
				Offset: 1,
				Text:   "candidate java publishing workflow using maven",
			})
			return true
		}

		// Java packages with gradle.
		r2 := regexp.MustCompile(`(?s)gradle.*publish`)
		if r2.MatchString(s) {
			dl.Info3(&checker.LogMessage{
				Path: fp,
				Type: checker.FileTypeSource,
				// Source file must have line number > 0.
				Offset: 1,
				Text:   "candidate java publishing workflow using gradle",
			})
			return true
		}
	}

	// Ruby packages.
	r := regexp.MustCompile(`(?s)gem.*push`)
	if r.MatchString(s) {
		dl.Info3(&checker.LogMessage{
			Path: fp,
			Type: checker.FileTypeSource,
			// Source file must have line number > 0.
			Offset: 1,
			Text:   "candidate ruby publishing workflow using gem",
		})
		return true
	}

	// NuGet packages.
	r = regexp.MustCompile(`(?s)nuget.*push`)
	if r.MatchString(s) {
		dl.Info3(&checker.LogMessage{
			Path: fp,
			Type: checker.FileTypeSource,
			// Source file must have line number > 0.
			Offset: 1,
			Text:   "candidate nuget publishing workflow",
		})
		return true
	}

	// Docker packages.
	if strings.Contains(s, "docker/build-push-action@") {
		dl.Info3(&checker.LogMessage{
			Path: fp,
			Type: checker.FileTypeSource,
			// Source file must have line number > 0.
			Offset: 1,
			Text:   "candidate docker publishing workflow",
		})
		return true
	}

	r = regexp.MustCompile(`(?s)docker.*push`)
	if r.MatchString(s) {
		dl.Info3(&checker.LogMessage{
			Path: fp,
			Type: checker.FileTypeSource,
			// Source file must have line number > 0.
			Offset: 1,
			Text:   "candidate docker publishing workflow",
		})
		return true
	}

	// Python packages.
	if strings.Contains(s, "actions/setup-python@") && strings.Contains(s, "pypa/gh-action-pypi-publish@master") {
		dl.Info3(&checker.LogMessage{
			Path: fp,
			Type: checker.FileTypeSource,
			// Source file must have line number > 0.
			Offset: 1,
			Text:   "candidate python publishing workflow using pypi",
		})
		return true
	}

	// Go packages.
	if strings.Contains(s, "actions/setup-go") &&
		strings.Contains(s, "goreleaser/goreleaser-action@") {
		dl.Info3(&checker.LogMessage{
			Path: fp,
			Type: checker.FileTypeSource,
			// Source file must have line number > 0.
			Offset: 1,
			Text:   "candidate golang publishing workflow",
		})
		return true
	}

	// Rust packages.
	// https://doc.rust-lang.org/cargo/reference/publishing.html.
	r = regexp.MustCompile(`(?s)cargo.*publish`)
	if r.MatchString(s) {
		dl.Info3(&checker.LogMessage{
			Path: fp,
			Type: checker.FileTypeSource,
			// Source file must have line number > 0.
			Offset: 1,
			Text:   "candidate rust publishing workflow using cargo",
		})
		return true
	}

	dl.Debug3(&checker.LogMessage{
		Path: fp,
		Type: checker.FileTypeSource,
		// Source file must have line number > 0.
		Offset: 1,
		Text:   "not a publishing workflow",
	})
	return false
}
