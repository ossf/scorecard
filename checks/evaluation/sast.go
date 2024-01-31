// Copyright 2023 OpenSSF Scorecard Authors
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

package evaluation

import (
	"fmt"
	"strconv"

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/sastToolCodeQLInstalled"
	"github.com/ossf/scorecard/v4/probes/sastToolPysaInstalled"
	"github.com/ossf/scorecard/v4/probes/sastToolQodanaInstalled"
	"github.com/ossf/scorecard/v4/probes/sastToolRunsOnAllCommits"
	"github.com/ossf/scorecard/v4/probes/sastToolSnykInstalled"
	"github.com/ossf/scorecard/v4/probes/sastToolSonarInstalled"
)

// SAST applies the score policy for the SAST check.
func SAST(name string,
	findings []finding.Finding, dl checker.DetailLogger,
) checker.CheckResult {
	// We have 3 unique probes, each should have a finding.
	expectedProbes := []string{
		sastToolCodeQLInstalled.Probe,
		sastToolPysaInstalled.Probe,
		sastToolQodanaInstalled.Probe,
		sastToolRunsOnAllCommits.Probe,
		sastToolSonarInstalled.Probe,
		sastToolSnykInstalled.Probe,
	}

	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	var sastScore, codeQlScore, pysaScore, qodanaScore, snykScore, sonarScore int
	var err error
	// Assign sastScore, codeQlScore and sonarScore
	for i := range findings {
		f := &findings[i]
		switch f.Probe {
		case sastToolRunsOnAllCommits.Probe:
			sastScore, err = getSASTScore(f, dl)
			if err != nil {
				return checker.CreateRuntimeErrorResult(name, sce.WithMessage(sce.ErrScorecardInternal, err.Error()))
			}
		case sastToolCodeQLInstalled.Probe:
			codeQlScore = getSastToolScore(f, dl)
		case sastToolSnykInstalled.Probe:
			snykScore = getSastToolScore(f, dl)
		case sastToolPysaInstalled.Probe:
			pysaScore = getSastToolScore(f, dl)
		case sastToolQodanaInstalled.Probe:
			qodanaScore = getSastToolScore(f, dl)
		case sastToolSonarInstalled.Probe:
			if f.Outcome == finding.OutcomePositive {
				sonarScore = checker.MaxResultScore
				dl.Info(&checker.LogMessage{
					Text:      f.Message,
					Type:      f.Location.Type,
					Path:      f.Location.Path,
					Offset:    *f.Location.LineStart,
					EndOffset: *f.Location.LineEnd,
					Snippet:   *f.Location.Snippet,
				})
			} else if f.Outcome == finding.OutcomeNegative {
				sonarScore = checker.MinResultScore
			}
		}
	}

	if sonarScore == checker.MaxResultScore {
		return checker.CreateMaxScoreResult(name, "SAST tool detected")
	}
	if snykScore == checker.MaxResultScore {
		return checker.CreateMaxScoreResult(name, "SAST tool detected: Snyk")
	}
	if pysaScore == checker.MaxResultScore {
		return checker.CreateMaxScoreResult(name, "SAST tool detected: Pysa")
	}
	if qodanaScore == checker.MaxResultScore {
		return checker.CreateMaxScoreResult(name, "SAST tool detected: Qodana")
	}

	if sastScore == checker.InconclusiveResultScore &&
		codeQlScore == checker.InconclusiveResultScore {
		// That can never happen since sastToolInCheckRuns can never
		// return checker.InconclusiveResultScore.
		return checker.CreateRuntimeErrorResult(name, sce.ErrScorecardInternal)
	}

	// Both scores are conclusive.
	// We assume the CodeQl config uses a cron and is not enabled as pre-submit.
	// TODO: verify the above comment in code.
	// We encourage developers to have sast check run on every pre-submit rather
	// than as cron jobs through the score computation below.
	// Warning: there is a hidden assumption that *any* sast tool is equally good.
	if sastScore != checker.InconclusiveResultScore &&
		codeQlScore != checker.InconclusiveResultScore {
		switch {
		case sastScore == checker.MaxResultScore:
			return checker.CreateMaxScoreResult(name, "SAST tool is run on all commits")
		case codeQlScore == checker.MinResultScore:
			return checker.CreateResultWithScore(name,
				checker.NormalizeReason("SAST tool is not run on all commits", sastScore), sastScore)

		// codeQl is enabled and sast has 0+ (but not all) PRs checks.
		case codeQlScore == checker.MaxResultScore:
			const sastWeight = 3
			const codeQlWeight = 7
			score := checker.AggregateScoresWithWeight(map[int]int{sastScore: sastWeight, codeQlScore: codeQlWeight})
			return checker.CreateResultWithScore(name, "SAST tool detected but not run on all commits", score)
		default:
			return checker.CreateRuntimeErrorResult(name, sce.WithMessage(sce.ErrScorecardInternal, "contact team"))
		}
	}

	// Sast inconclusive.
	if codeQlScore != checker.InconclusiveResultScore {
		if codeQlScore == checker.MaxResultScore {
			return checker.CreateMaxScoreResult(name, "SAST tool detected: CodeQL")
		}
		return checker.CreateMinScoreResult(name, "no SAST tool detected")
	}

	// CodeQl inconclusive.
	if sastScore != checker.InconclusiveResultScore {
		if sastScore == checker.MaxResultScore {
			return checker.CreateMaxScoreResult(name, "SAST tool is run on all commits")
		}

		return checker.CreateResultWithScore(name,
			checker.NormalizeReason("SAST tool is not run on all commits", sastScore), sastScore)
	}

	// Should never happen.
	return checker.CreateRuntimeErrorResult(name, sce.WithMessage(sce.ErrScorecardInternal, "contact team"))
}

// getSASTScore returns the proportional score of how many commits
// run SAST tools.
func getSASTScore(f *finding.Finding, dl checker.DetailLogger) (int, error) {
	switch f.Outcome {
	case finding.OutcomeNotApplicable:
		dl.Warn(&checker.LogMessage{
			Text: f.Message,
		})
		return checker.InconclusiveResultScore, nil
	case finding.OutcomePositive:
		dl.Info(&checker.LogMessage{
			Text: f.Message,
		})
	case finding.OutcomeNegative:
		dl.Warn(&checker.LogMessage{
			Text: f.Message,
		})
	default:
	}
	analyzed, err := strconv.Atoi(f.Values[sastToolRunsOnAllCommits.AnalyzedPRsKey])
	if err != nil {
		return 0, fmt.Errorf("parsing analyzed PR count: %w", err)
	}
	total, err := strconv.Atoi(f.Values[sastToolRunsOnAllCommits.TotalPRsKey])
	if err != nil {
		return 0, fmt.Errorf("parsing total PR count: %w", err)
	}
	return checker.CreateProportionalScore(analyzed, total), nil
}

// getSastToolScore returns positive if the project runs the Sast tool
// and negative if it doesn't.
func getSastToolScore(f *finding.Finding, dl checker.DetailLogger) int {
	switch f.Outcome {
	case finding.OutcomePositive:
		dl.Info(&checker.LogMessage{
			Text: f.Message,
		})
		return checker.MaxResultScore
	case finding.OutcomeNegative:
		return checker.MinResultScore
	default:
		return checker.InconclusiveResultScore
	}
}
