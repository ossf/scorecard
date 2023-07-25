// © 2023 Nokia
// Licensed under the BSD 3 Clause license
// SPDX-License-Identifier: BSD-3-Clause

package codeApproved

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
)

//go:embed *.yml
var fs embed.FS

const probe = "codeApproved"

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	rawReviewData := &raw.CodeReviewResults
	return approvedRun(rawReviewData, fs, probe, finding.OutcomePositive, finding.OutcomeNegative)
}

/*
** Looks through the data and validates that each changeset has been approved at least once.
*/

func approvedRun(reviewData *checker.CodeReviewData, fs embed.FS, probeID string,
	positiveOutcome, negativeOutcome finding.Outcome,
) ([]finding.Finding, string, error) {
	var findings []finding.Finding
	var approvedReviews = 0
	changesets := reviewData.DefaultBranchChangesets
	var numChangesets = len(changesets)
	for x := range changesets {
		data := &changesets[x]
		for y := range data.Reviews {
			if data.Reviews[y].State == "APPROVED" && data.Reviews[y].Author.Login != data.Author.Login {
				approvedReviews += 1
				break
			}
		}
	}
	if approvedReviews >= numChangesets {
		f, err := finding.NewWith(fs, probeID, fmt.Sprintf("All changesets approved (%v out of %v).", approvedReviews, numChangesets),
			nil, positiveOutcome)
		if err != nil {
			return nil, probeID, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	} else {
		f, err := finding.NewWith(fs, probeID, fmt.Sprintf("Not all changesets approved. Found %v approvals among %v changesets.", approvedReviews, numChangesets),
			nil, negativeOutcome)
		if err != nil {
			return nil, probeID, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	}
	return findings, probeID, nil
}
