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

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
)

// CheckSAST is the registered name for SAST.
const CheckSAST = "SAST"

var sastTools = map[string]bool{"github-code-scanning": true, "lgtm-com": true, "sonarcloud": true}

var allowedConclusions = map[string]bool{"success": true, "neutral": true}

//nolint:gochecknoinits
func init() {
	if err := registerCheck(CheckSAST, SAST); err != nil {
		// This should never happen.
		panic(err)
	}
}

// SAST runs SAST check.
func SAST(c *checker.CheckRequest) checker.CheckResult {
	sastScore, sastErr := sastToolInCheckRuns(c)
	if sastErr != nil {
		return checker.CreateRuntimeErrorResult(CheckSAST, sastErr)
	}

	codeQlScore, codeQlErr := codeQLInCheckDefinitions(c)
	if codeQlErr != nil {
		return checker.CreateRuntimeErrorResult(CheckSAST, codeQlErr)
	}

	// Both results are inconclusive.
	// Can never happen.
	if sastScore == checker.InconclusiveResultScore &&
		codeQlScore == checker.InconclusiveResultScore {
		// That can never happen since sastToolInCheckRuns can never
		// retun checker.InconclusiveResultScore.
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
		case sastScore == checker.MaxResultScore:
			return checker.CreateMaxScoreResult(CheckSAST, "SAST tool is run on all commits")
		case codeQlScore == checker.MinResultScore:
			return checker.CreateResultWithScore(CheckSAST,
				checker.NormalizeReason("SAST tool is not run on all commits", sastScore), sastScore)

		// codeQl is enabled and sast has 0+ (but not all) PRs checks.
		case codeQlScore == checker.MaxResultScore:
			const sastWeight = 3
			const codeQlWeight = 7
			score := checker.AggregateScoresWithWeight(map[int]int{sastScore: sastWeight, codeQlScore: codeQlWeight})
			return checker.CreateResultWithScore(CheckSAST, "SAST tool detected but not run on all commmits", score)
		default:
			return checker.CreateRuntimeErrorResult(CheckSAST, sce.WithMessage(sce.ErrScorecardInternal, "contact team"))
		}
	}

	// Sast inconclusive.
	if codeQlScore != checker.InconclusiveResultScore {
		if codeQlScore == checker.MaxResultScore {
			return checker.CreateMaxScoreResult(CheckSAST, "SAST tool detected")
		}
		return checker.CreateMinScoreResult(CheckSAST, "no SAST tool detected")
	}

	// CodeQl inconclusive.
	if sastScore != checker.InconclusiveResultScore {
		if sastScore == checker.MaxResultScore {
			return checker.CreateMaxScoreResult(CheckSAST, "SAST tool is run on all commits")
		}

		return checker.CreateResultWithScore(CheckSAST,
			checker.NormalizeReason("SAST tool is not run on all commits", sastScore), sastScore)
	}

	// Should never happen.
	return checker.CreateRuntimeErrorResult(CheckSAST, sce.WithMessage(sce.ErrScorecardInternal, "contact team"))
}

// nolint
func sastToolInCheckRuns(c *checker.CheckRequest) (int, error) {
	commits, err := c.RepoClient.ListCommits()
	if err != nil {
		//nolint
		return checker.InconclusiveResultScore,
			sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("RepoClient.ListCommits: %v", err))
	}

	totalMerged := 0
	totalTested := 0
	for _, commit := range commits {
		pr := commit.AssociatedMergeRequest
		// TODO(#575): We ignore associated PRs if Scorecard is being run on a fork
		// but the PR was created in the original repo.
		if pr.MergedAt.IsZero() || pr.Repository != c.Repo.String() {
			continue
		}
		totalMerged++
		crs, err := c.RepoClient.ListCheckRunsForRef(pr.HeadSHA)
		if err != nil {
			return checker.InconclusiveResultScore,
				sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("Client.Checks.ListCheckRunsForRef: %v", err))
		}
		// Note: crs may be `nil`: in this case
		// the loop below will be skipped.
		for _, cr := range crs {
			if cr.Status != "completed" {
				continue
			}
			if !allowedConclusions[cr.Conclusion] {
				continue
			}
			if sastTools[cr.App.Slug] {
				c.Dlogger.Debug3(&checker.LogMessage{
					Path: cr.URL,
					Type: checker.FileTypeURL,
					Text: "tool detected",
				})
				totalTested++
				break
			}
		}
	}
	if totalMerged == 0 {
		c.Dlogger.Warn3(&checker.LogMessage{
			Text: "no pull requests merged into dev branch",
		})
		return checker.InconclusiveResultScore, nil
	}

	if totalTested == totalMerged {
		c.Dlogger.Info3(&checker.LogMessage{
			Text: fmt.Sprintf("all commits (%v) are checked with a SAST tool", totalMerged),
		})
	} else {
		c.Dlogger.Warn3(&checker.LogMessage{
			Text: fmt.Sprintf("%v commits out of %v are checked with a SAST tool", totalTested, totalMerged),
		})
	}

	return checker.CreateProportionalScore(totalTested, totalMerged), nil
}

// nolint
func codeQLInCheckDefinitions(c *checker.CheckRequest) (int, error) {
	searchRequest := clients.SearchRequest{
		Query: "github/codeql-action/analyze",
		Path:  "/.github/workflows",
	}
	resp, err := c.RepoClient.Search(searchRequest)
	if err != nil {
		return checker.InconclusiveResultScore,
			sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("Client.Search.Code: %v", err))
	}

	for _, result := range resp.Results {
		c.Dlogger.Debug3(&checker.LogMessage{
			Path:   result.Path,
			Type:   checker.FileTypeSource,
			Offset: checker.OffsetDefault,
			Text:   "CodeQL detected",
		})
	}

	// TODO: check if it's enabled as cron or presubmit.
	// TODO: check which branches it is enabled on. We should find main.
	if resp.Hits > 0 {
		c.Dlogger.Info3(&checker.LogMessage{
			Text: "SAST tool detected: CodeQL",
		})
		return checker.MaxResultScore, nil
	}

	c.Dlogger.Warn3(&checker.LogMessage{
		Text: "CodeQL tool not detected",
	})
	return checker.MinResultScore, nil
}
