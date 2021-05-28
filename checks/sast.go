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
	"errors"

	"github.com/google/go-github/v32/github"

	"github.com/ossf/scorecard/checker"
)

// CheckSAST is the registered name for SAST.
const CheckSAST = "SAST"

var (
	sastTools = map[string]bool{"github-code-scanning": true, "sonarcloud": true}
	// ErrorNoChecks indicates no GitHub Check runs were found for this repo.
	ErrorNoChecks = errors.New("no check runs found")
	// ErrorNoMerges indicates no merges with SAST tool runs were found for this repo.
	ErrorNoMerges = errors.New("no merges found")
)

//nolint:gochecknoinits
func init() {
	registerCheck(CheckSAST, SAST)
}

func SAST(c *checker.CheckRequest) checker.CheckResult {
	return checker.MultiCheckOr(
		CodeQLInCheckDefinitions,
		SASTToolInCheckRuns,
	)(c)
}

func SASTToolInCheckRuns(c *checker.CheckRequest) checker.CheckResult {
	prs, _, err := c.Client.PullRequests.List(c.Ctx, c.Owner, c.Repo, &github.PullRequestListOptions{
		State: "closed",
	})
	if err != nil {
		return checker.MakeRetryResult(CheckSAST, err)
	}

	totalMerged := 0
	totalTested := 0
	for _, pr := range prs {
		if pr.MergedAt == nil {
			continue
		}
		totalMerged++
		crs, _, err := c.Client.Checks.ListCheckRunsForRef(c.Ctx, c.Owner, c.Repo, pr.GetHead().GetSHA(),
			&github.ListCheckRunsOptions{})
		if err != nil {
			return checker.MakeRetryResult(CheckSAST, err)
		}
		if crs == nil {
			return checker.MakeInconclusiveResult(CheckSAST, ErrorNoChecks)
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
		return checker.MakeInconclusiveResult(CheckSAST, ErrorNoMerges)
	}
	return checker.MakeProportionalResult(CheckSAST, totalTested, totalMerged, .75)
}

func CodeQLInCheckDefinitions(c *checker.CheckRequest) checker.CheckResult {
	searchQuery := ("github/codeql-action path:/.github/workflows repo:" + c.Owner + "/" + c.Repo)
	results, _, err := c.Client.Search.Code(c.Ctx, searchQuery, &github.SearchOptions{})
	if err != nil {
		return checker.MakeRetryResult(CheckSAST, err)
	}

	for _, result := range results.CodeResults {
		c.Logf("found CodeQL definition: %s", result.GetPath())
	}

	return checker.CheckResult{
		Name:       CheckSAST,
		Pass:       *results.Total > 0,
		Confidence: checker.MaxResultConfidence,
	}
}
