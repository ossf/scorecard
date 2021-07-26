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

	"github.com/google/go-github/v32/github"

	"github.com/ossf/scorecard/v2/checker"
	sce "github.com/ossf/scorecard/v2/errors"
)

// CheckSAST is the registered name for SAST.
const CheckSAST = "SAST"

var sastTools = map[string]bool{"github-code-scanning": true, "sonarcloud": true}

//nolint:gochecknoinits
func init() {
	registerCheck(CheckSAST, SAST)
}

func SAST(c *checker.CheckRequest) checker.CheckResult {
	sastScore, sastReason, sastErr := SASTToolInCheckRuns(c)
	if sastErr != nil {
		return checker.CreateRuntimeErrorResult(sastReason, sastErr)
	}

	codeQlScore, codeQlReason, codeQlErr := CodeQLInCheckDefinitions(c)
	if codeQlErr != nil {
		return checker.CreateRuntimeErrorResult(codeQlReason, codeQlErr)
	}

	// Both results are inconclusive.
	if sastScore == checker.InconclusiveResultScore &&
		codeQlScore == checker.InconclusiveResultScore {
		c.Dlogger.Warn(sastReason)
		c.Dlogger.Warn(codeQlReason)
		return checker.CreateInconclusiveResult(CheckSAST, "internal error")
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
		// This only happens if:
		// - sastScore is maximum and codeQl is enabled OR
		// - sastScore is minimum and codeQl is not enabled
		// In both cases, sastReason gives the best reason to the user.
		case sastScore >= codeQlScore:
			// Add codeQlReason to the details.
			c.Dlogger.Warn(codeQlReason)
			return checker.CreateProportionalScoreResult(CheckSAST, sastReason, sastScore, checker.MaxResultScore)
		// codeQl is enabled and sast has 0+ (but not all) PRs checks.
		// In this case, codeQlReason provides good information.
		case codeQlScore == checker.MaxResultScore:
			// Add sastReason to the details.
			c.Dlogger.Warn(sastReason)
			const sastWeight = 3
			const codeQlWeight = 7
			score := checker.AggregateScoresWithWeight(map[int]int{sastScore: sastWeight, codeQlScore: codeQlWeight})
			return checker.CreateResultWithScore(CheckSAST, codeQlReason, score)
		default:
			return checker.CreateRuntimeErrorResult(CheckSAST, sce.Create(sce.ErrScorecardInternal, "contact team"))
		}
	}

	// CodeQl inconclusive.
	if codeQlScore != checker.InconclusiveResultScore {
		c.Dlogger.Warn(sastReason)
		return checker.CreateResultWithScore(CheckSAST, codeQlReason, codeQlScore)
	}

	// Sast inconclusive.
	if sastScore != checker.InconclusiveResultScore {
		c.Dlogger.Warn(codeQlReason)
		return checker.CreateResultWithScore(CheckSAST, sastReason, sastScore)
	}

	// Should never happen.
	return checker.CreateRuntimeErrorResult(CheckSAST, sce.Create(sce.ErrScorecardInternal, "contact team"))
}

//nolint
func SASTToolInCheckRuns(c *checker.CheckRequest) (int, string, error) {
	prs, _, err := c.Client.PullRequests.List(c.Ctx, c.Owner, c.Repo, &github.PullRequestListOptions{
		State: "closed",
	})
	if err != nil {
		//nolint
		return checker.InconclusiveResultScore, "",
			sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("Client.PullRequests.List: %v", err))
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
			return checker.InconclusiveResultScore, "",
				sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("Client.Checks.ListCheckRunsForRef: %v", err))
		}
		if crs == nil {
			return checker.InconclusiveResultScore, "no merges detected", nil
		}
		for _, cr := range crs.CheckRuns {
			if cr.GetStatus() != "completed" {
				continue
			}
			if cr.GetConclusion() != "success" {
				continue
			}
			if sastTools[cr.GetApp().GetSlug()] {
				c.Dlogger.Debug("tool detected: %s", cr.GetHTMLURL())
				totalTested++
				break
			}
		}
	}
	if totalMerged == 0 {
		return checker.InconclusiveResultScore, "no merges detected", nil
	}
	reason := fmt.Sprintf("%v commits out of %v are checked with a SAST tool", totalTested, totalMerged)
	return checker.CreateProportionalScore(totalTested, totalMerged), reason, nil
}

//nolint
func CodeQLInCheckDefinitions(c *checker.CheckRequest) (int, string, error) {
	searchQuery := ("github/codeql-action path:/.github/workflows repo:" + c.Owner + "/" + c.Repo)
	results, _, err := c.Client.Search.Code(c.Ctx, searchQuery, &github.SearchOptions{})
	if err != nil {
		return checker.InconclusiveResultScore, "",
			sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("Client.Search.Code: %v", err))
	}

	for _, result := range results.CodeResults {
		c.Dlogger.Info("CodeQL definition detected: %s", result.GetPath())
	}

	// TODO: check if it's enabled as cron or presubmit.
	// TODO: check which branches it is enabled on. We should find main.
	if *results.Total > 0 {
		return checker.MaxResultScore, "tool detected: CodeQL", nil
	}
	return checker.MinResultScore, "CodeQL tool not detected", nil
}
