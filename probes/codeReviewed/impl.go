// © 2023 Nokia
// Licensed under the BSD 3 Clause license
// SPDX-License-Identifier: BSD-3-Clause

package codeReviewed

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
)

//go:embed *.yml
var fs embed.FS

const probe = "codeReviewed"

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	rawReviewData := &raw.CodeReviewResults
	return reviewedRun(rawReviewData, fs, probe, finding.OutcomePositive, finding.OutcomeNegative)
}

/*
** Looks through the data and validates that each changeset has been approved at least once.
*/

func reviewedRun(reviewData *checker.CodeReviewData, fs embed.FS, probeID string,
	positiveOutcome, negativeOutcome finding.Outcome,
) ([]finding.Finding, string, error) {
	var findings []finding.Finding
	var numReviews = 0
	changesets := reviewData.DefaultBranchChangesets
	var numChangesets = len(changesets)
	for x := range changesets {
		data := &changesets[x]
		if len(data.Reviews) > 0 {
			numReviews += 1
		}
	}
	if numReviews >= numChangesets {
		f, err := finding.NewWith(fs, probeID, fmt.Sprintf("All changesets have review activity (%v out of %v).", numReviews, numChangesets),
			nil, positiveOutcome)
		if err != nil {
			return nil, probeID, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	} else {
		f, err := finding.NewWith(fs, probeID, fmt.Sprintf("Not all changesets have review activity. Found %v reviews among %v changesets.", numReviews, numChangesets),
			nil, negativeOutcome)
		if err != nil {
			return nil, probeID, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	}
	return findings, probeID, nil
}
