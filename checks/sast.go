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
	"github.com/ossf/scorecard/lib"
)

const sastStr = "SAST"

var sastTools map[string]bool = map[string]bool{"github-code-scanning": true, "sonarcloud": true}

func init() {
	registerCheck(sastStr, SAST)
}

func SAST(c lib.CheckRequest) lib.CheckResult {
	return lib.MultiCheck(
		CodeQLInCheckDefinitions,
		SASTToolInCheckRuns,
	)(c)
}

func SASTToolInCheckRuns(c lib.CheckRequest) lib.CheckResult {
	prs, _, err := c.Client.PullRequests.List(c.Ctx, c.Owner, c.Repo, &github.PullRequestListOptions{
		State: "closed",
	})
	if err != nil {
		return lib.MakeRetryResult(sastStr, err)
	}

	totalMerged := 0
	totalTested := 0
	for _, pr := range prs {
		if pr.MergedAt == nil {
			continue
		}
		totalMerged++
		crs, _, err := c.Client.Checks.ListCheckRunsForRef(c.Ctx, c.Owner, c.Repo, pr.GetHead().GetSHA(), &github.ListCheckRunsOptions{})
		if err != nil {
			return lib.MakeRetryResult(sastStr, err)
		}
		if crs == nil {
			return lib.MakeInconclusiveResult(sastStr)
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
		return lib.MakeInconclusiveResult(sastStr)
	}
	return lib.MakeProportionalResult(sastStr, totalTested, totalMerged, .75)
}

func CodeQLInCheckDefinitions(c lib.CheckRequest) lib.CheckResult {
	searchQuery := ("github/codeql-action path:/.github/workflows repo:" + c.Owner + "/" + c.Repo)
	results, _, err := c.Client.Search.Code(c.Ctx, searchQuery, &github.SearchOptions{})
	if err != nil {
		return lib.MakeRetryResult(sastStr, err)
	}

	for _, result := range results.CodeResults {
		c.Logf("found CodeQL definition: %s", result.GetPath())
	}

	return lib.CheckResult{
		Name:       sastStr,
		Pass:       *results.Total > 0,
		Confidence: lib.MaxResultConfidence,
	}
}
