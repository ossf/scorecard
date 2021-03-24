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
	"github.com/google/go-github/v32/github"
	"github.com/ossf/scorecard/checker"
)

var sastTools map[string]bool = map[string]bool{"github-code-scanning": true, "sonarcloud": true}

const (
	//nolint
	description string = `This check tries to determine if the project uses static code analysis systems. It currently works by looking for well-known results (CodeQL, etc.) in GitHub pull requests.`
	helpURL     string = "https://github.com/ossf/scorecard/blob/main/checks.md#sast"
)

func init() {
	registerCheck("SAST", SAST)
}

func SAST(c checker.Checker) checker.CheckResult {
	return checker.MultiCheck(
		CodeQLInCheckDefinitions,
		SASTToolInCheckRuns,
	)(c)
}

func SASTToolInCheckRuns(c checker.Checker) checker.CheckResult {
	prs, _, err := c.Client.PullRequests.List(c.Ctx, c.Owner, c.Repo, &github.PullRequestListOptions{
		State: "closed",
	})
	if err != nil {
		r := checker.RetryResult(err)
		r.Description = description
		r.HelpURL = helpURL
		return r
	}

	totalMerged := 0
	totalTested := 0
	for _, pr := range prs {
		if pr.MergedAt == nil {
			continue
		}
		totalMerged++
		crs, _, err := c.Client.Checks.
			ListCheckRunsForRef(c.Ctx, c.Owner, c.Repo, pr.GetHead().GetSHA(), &github.ListCheckRunsOptions{})
		if err != nil {
			r := checker.RetryResult(err)
			r.Description = description
			r.HelpURL = helpURL
			return r
		}
		if crs == nil {
			r := checker.InconclusiveResult
			r.Description = description
			r.HelpURL = helpURL
			return r
		}
		for _, cr := range crs.CheckRuns {
			if cr.GetStatus() != "completed" {
				continue
			}
			if cr.GetConclusion() != "success" {
				continue
			}
			if sastTools[cr.GetApp().GetSlug()] {
				c.Logf("SAST Tool found: %s", cr.GetHTMLURL())
				totalTested++
				break
			}
		}
	}
	if totalTested == 0 {
		r := checker.InconclusiveResult
		r.Description = description
		r.HelpURL = helpURL
		return r
	}
	r := checker.ProportionalResult(totalTested, totalMerged, .75)
	r.Description = description
	r.HelpURL = helpURL
	return r
}

func CodeQLInCheckDefinitions(c checker.Checker) checker.CheckResult {
	searchQuery := ("github/codeql-action path:/.github/workflows repo:" + c.Owner + "/" + c.Repo)
	results, _, err := c.Client.Search.Code(c.Ctx, searchQuery, &github.SearchOptions{})
	if err != nil {
		r := checker.RetryResult(err)
		r.Description = description
		r.HelpURL = helpURL
		return r
	}

	for _, result := range results.CodeResults {
		c.Logf("found CodeQL definition: %s", result.GetPath())
	}

	const confidence = 10
	return checker.CheckResult{
		Pass:        *results.Total > 0,
		Confidence:  confidence,
		Description: description,
		HelpURL:     helpURL,
	}
}
