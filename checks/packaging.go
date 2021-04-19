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
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/go-github/v32/github"
	"github.com/ossf/scorecard/checker"
)

const packagingStr = "Packaging"

func init() {
	registerCheck(packagingStr, Packaging)
}

func Packaging(c *checker.CheckRequest) checker.CheckResult {
	_, dc, _, err := c.Client.Repositories.GetContents(c.Ctx, c.Owner, c.Repo, ".github/workflows", &github.RepositoryContentGetOptions{})
	if err != nil {
		return checker.MakeRetryResult(packagingStr, err)
	}

	for _, f := range dc {
		fp := f.GetPath()
		fo, _, _, err := c.Client.Repositories.GetContents(c.Ctx, c.Owner, c.Repo, fp, &github.RepositoryContentGetOptions{})
		if err != nil {
			return checker.MakeRetryResult(packagingStr, err)
		}
		if fo == nil {
			// path is a directory, not a file. skip.
			continue
		}
		fc, err := fo.GetContent()
		if err != nil {
			return checker.MakeRetryResult(packagingStr, err)
		}

		if !isPackagingWorkflow(fc, fp, c) {
			continue
		}

		runs, _, err := c.Client.Actions.ListWorkflowRunsByFileName(c.Ctx, c.Owner, c.Repo, filepath.Base(fp), &github.ListWorkflowRunsOptions{
			Status: "success",
		})
		if err != nil {
			return checker.MakeRetryResult(packagingStr, err)
		}
		if *runs.TotalCount > 0 {
			c.Logf("found a completed run: %s", runs.WorkflowRuns[0].GetHTMLURL())
			return checker.CheckResult{
				Name:       packagingStr,
				Pass:       true,
				Confidence: checker.MaxResultConfidence,
			}
		}
		c.Logf("!! no run completed")
	}

	return checker.CheckResult{
		Name:       packagingStr,
		Pass:       false,
		Confidence: checker.MaxResultConfidence,
	}
}

func isPackagingWorkflow(s, fp string, c *checker.CheckRequest) bool {
	// nodejs packages
	if strings.Contains(s, "uses: actions/setup-node@") {
		r1 := regexp.MustCompile(`(?s)registry-url.*https://registry\.npmjs\.org`)
		r2 := regexp.MustCompile(`(?s)npm.*publish`)

		if r1.MatchString(s) && r2.MatchString(s) {
			c.Logf("found node packaging workflow using npm: %s", fp)
			return true
		}
	}

	if strings.Contains(s, "uses: actions/setup-java@") {
		// java packages with maven
		r1 := regexp.MustCompile(`(?s)mvn.*deploy`)
		if r1.MatchString(s) {
			c.Logf("found java packaging workflow using maven: %s", fp)
			return true
		}

		// java packages with gradle
		r2 := regexp.MustCompile(`(?s)gradle.*publish`)
		if r2.MatchString(s) {
			c.Logf("found java packaging workflow using gradle: %s", fp)
			return true
		}
	}

	if strings.Contains(s, "actions/setup-python@") && strings.Contains(s, "pypa/gh-action-pypi-publish@master") {
		c.Logf("found python packaging workflow using pypi: %s", fp)
		return true
	}

	if strings.Contains(s, "uses: docker/build-push-action@") {
		c.Logf("found docker publishing workflow: %s", fp)
		return true
	}

	if strings.Contains(s, "docker push") {
		c.Logf("found docker publishing workflow: %s", fp)
		return true
	}

	c.Logf("!! not a packaging workflow: %s", fp)
	return false
}
