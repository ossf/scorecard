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
	"strings"

	"github.com/google/go-github/v32/github"

	"github.com/ossf/scorecard/checker"
	sce "github.com/ossf/scorecard/errors"
)

// CheckSAST is the registered name for SAST.
const CheckSAST = "SAST"

var sastTools = map[string]bool{"github-code-scanning": true, "sonarcloud": true}

//nolint:gochecknoinits
func init() {
	registerCheck(CheckSAST, SAST)
}

func SAST(c *checker.CheckRequest) checker.CheckResult {
	r1 := SASTToolInCheckRuns(c)
	r2 := CodeQLInCheckDefinitions(c)
	if r1.Error2 != nil {
		return r1
	}
	if r2.Error2 != nil {
		return r2
	}
	// Merge the results.
	var result checker.CheckResult
	if r1.Score == checker.MaxResultScore {
		// All commits have a SAST tool check run,
		// score is maximum.
		result = r1
		result.Score = 10
		i := strings.Index(r2.Reason, "-- score normalized")
		c.Dlogger.Info(r2.Reason[:i])
	} else {
		// Not all commits have a check run,
		// We compute the final score as follows:
		// 5 points for CodeQL enabled, 5 points for
		// SAST run on commits.
		result = r2

		//nolint
		result.Score = 5 + r1.Score/10*5
		result.Reason = "not all commits are checked with a SAST tool"
		i := strings.Index(r1.Reason, "-- score normalized")
		c.Dlogger.Info(r1.Reason[:i])
	}

	return result
}

func SASTToolInCheckRuns(c *checker.CheckRequest) checker.CheckResult {
	prs, _, err := c.Client.PullRequests.List(c.Ctx, c.Owner, c.Repo, &github.PullRequestListOptions{
		State: "closed",
	})
	if err != nil {
		e := sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("Client.PullRequests.List: %v", err))
		return checker.CreateRuntimeErrorResult(CheckSAST, e)
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
			e := sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("Client.Checks.ListCheckRunsForRef: %v", err))
			return checker.CreateRuntimeErrorResult(CheckSAST, e)
		}
		if crs == nil {
			return checker.CreateInconclusiveResult(CheckSAST, "no merges detected")
		}
		for _, cr := range crs.CheckRuns {
			if cr.GetStatus() != "completed" {
				continue
			}
			if cr.GetConclusion() != "success" {
				continue
			}
			if sastTools[cr.GetApp().GetSlug()] {
				c.Dlogger.Info("tool detected: %s", cr.GetHTMLURL())
				totalTested++
				break
			}
		}
	}
	if totalMerged == 0 {
		return checker.CreateInconclusiveResult(CheckSAST, "no merges detected")
	}
	reason := fmt.Sprintf("%v commits out of %v are checked with a SAST tool", totalTested, totalMerged)
	return checker.CreateProportionalScoreResult(CheckSAST, reason, totalTested, totalMerged)
}

func CodeQLInCheckDefinitions(c *checker.CheckRequest) checker.CheckResult {
	searchQuery := ("github/codeql-action path:/.github/workflows repo:" + c.Owner + "/" + c.Repo)
	results, _, err := c.Client.Search.Code(c.Ctx, searchQuery, &github.SearchOptions{})
	if err != nil {
		e := sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("Client.Search.Code: %v", err))
		return checker.CreateRuntimeErrorResult(CheckSAST, e)
	}

	for _, result := range results.CodeResults {
		c.Dlogger.Info("CodeQL definition detected: %s", result.GetPath())
	}

	// TODO: check if it's enabled as cron or presubmit.
	// TODO: check which branches it is enabled on. We should find main.
	if *results.Total > 0 {
		return checker.CreateMaxScoreResult(CheckSAST, "tool detected: CodeQL")
	}
	return checker.CreateMinScoreResult(CheckSAST, "CodeQL tool not detected")
}
