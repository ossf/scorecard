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

	"github.com/google/go-github/v32/github"

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
		e := sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("RepoClient.ListFiles: %v", err))
		return checker.CreateRuntimeErrorResult(CheckPackaging, e)
	}

	for _, fp := range matchedFiles {
		fc, err := c.RepoClient.GetFileContent(fp)
		if err != nil {
			e := sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("RepoClient.GetFileContent: %v", err))
			return checker.CreateRuntimeErrorResult(CheckPackaging, e)
		}

		if !isPackagingWorkflow(fc, fp, c.Dlogger) {
			continue
		}

		runs, _, err := c.Client.Actions.ListWorkflowRunsByFileName(c.Ctx, c.Owner, c.Repo, filepath.Base(fp),
			&github.ListWorkflowRunsOptions{
				Status: "success",
			})
		if err != nil {
			e := sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("Client.Actions.ListWorkflowRunsByFileName: %v", err))
			return checker.CreateRuntimeErrorResult(CheckPackaging, e)
		}
		if *runs.TotalCount > 0 {
			c.Dlogger.Info("workflow %v used in run: %s", fp, runs.WorkflowRuns[0].GetHTMLURL())
			return checker.CreateMaxScoreResult(CheckPackaging,
				"packaging workflow detected")
		}
		c.Dlogger.Info("workflow %v not used in runs", fp)
	}

	return checker.CreateMinScoreResult(CheckPackaging,
		"no packaging workflow used")
}

// A packaging workflow.
func isPackagingWorkflow(s, fp string, dl checker.DetailLogger) bool {
	// Nodejs packages.
	if strings.Contains(s, "actions/setup-node@") {
		r1 := regexp.MustCompile(`(?s)registry-url.*https://registry\.npmjs\.org`)
		r2 := regexp.MustCompile(`(?s)npm.*publish`)

		if r1.MatchString(s) && r2.MatchString(s) {
			dl.Info("candidate node packaging workflow using npm: %s", fp)
			return true
		}
	}

	// Java packages.
	if strings.Contains(s, "actions/setup-java@") {
		// Java packages with maven.
		r1 := regexp.MustCompile(`(?s)mvn.*deploy`)
		if r1.MatchString(s) {
			dl.Info("candidate java packaging workflow using maven: %s", fp)
			return true
		}

		// Java packages with gradle.
		r2 := regexp.MustCompile(`(?s)gradle.*publish`)
		if r2.MatchString(s) {
			dl.Info("candidate java packaging workflow using gradle: %s", fp)
			return true
		}
	}

	// Ruby packages.
	r := regexp.MustCompile(`(?s)gem.*push`)
	if r.MatchString(s) {
		dl.Info("ruby publishing workflow using gem: %s", fp)
		return true
	}

	// NuGet packages.
	r = regexp.MustCompile(`(?s)nuget.*push`)
	if r.MatchString(s) {
		dl.Info("nuget publishing workflow: %s", fp)
		return true
	}

	// Docker packages.
	if strings.Contains(s, "docker/build-push-action@") {
		dl.Info("candidate docker publishing workflow: %s", fp)
		return true
	}

	r = regexp.MustCompile(`(?s)docker.*push`)
	if r.MatchString(s) {
		dl.Info("candidate docker publishing workflow: %s", fp)
		return true
	}

	// Python packages.
	if strings.Contains(s, "actions/setup-python@") && strings.Contains(s, "pypa/gh-action-pypi-publish@master") {
		dl.Info("candidate python packaging workflow using pypi: %s", fp)
		return true
	}

	// Go packages.
	if strings.Contains(s, "actions/setup-go") &&
		strings.Contains(s, "goreleaser/goreleaser-action@") {
		dl.Info("candidate golang packaging workflow: %s", fp)
		return true
	}

	dl.Debug("not a packaging workflow: %s", fp)
	return false
}
