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

//nolint:stylecheck
package codeReviewed

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/utils"
)

//go:embed *.yml
var fs embed.FS

const probe = "codeReviewed"

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	rawReviewData := &raw.CodeReviewResults
	return reviewedRun(rawReviewData, fs, probe, finding.OutcomePositive, finding.OutcomeNegative)
}

// Looks through the data and validates that each changeset has been approved at least once.
func reviewedRun(reviewData *checker.CodeReviewData, fs embed.FS, probeID string,
	positiveOutcome, negativeOutcome finding.Outcome,
) ([]finding.Finding, string, error) {
	changesets := reviewData.DefaultBranchChangesets
	var findings []finding.Finding
	numReviews := 0
	numBotAuthors := 0
	numChangesets := len(changesets)
	if numChangesets == 0 {
		return nil, probeID, utils.NoChangesetsErr
	}
	for x := range changesets {
		data := &changesets[x]
		if data.Author.IsBot {
			numBotAuthors += 1
		}
		if len(data.Reviews) > 0 {
			numReviews += 1
		}
	}
	switch {
	case numBotAuthors == numChangesets:
		// returns a NotAvailable outcome if all changesets were authored by bots
		f, err := finding.NewNotAvailable(fs, probeID, "All changesets authored by bot(s).", nil)
		if err != nil {
			return nil, probeID, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
		return findings, probeID, nil
	case numReviews >= numChangesets:
		// returns PositiveOutcome if all changesets had review activity
		f, err := finding.NewWith(fs, probeID, fmt.Sprintf("All changesets have review activity "+
			"(%v out of %v).", numReviews, numChangesets), nil, positiveOutcome)
		if err != nil {
			return nil, probeID, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	default:
		// returns NegativeOutcome if some changesets did not have review activity
		f, err := finding.NewWith(fs, probeID, fmt.Sprintf("Not all changesets have review activity. "+
			"Found %v reviews among %v changesets.", numReviews, numChangesets), nil, negativeOutcome)
		if err != nil {
			return nil, probeID, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	}
	return findings, probeID, nil
}
