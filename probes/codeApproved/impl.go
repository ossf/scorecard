// © 2023 Nokia
// Licensed under the Apache License 2.0
// SPDX-License-Identifier: Apache-2.0

package codeApproved

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/utils"
)

//go:embed *.yml
var fs embed.FS

const probe = "codeApproved"

const noReviewerFound = -1

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	rawReviewData := &raw.CodeReviewResults
	return approvedRun(rawReviewData, fs, probe, finding.OutcomePositive, finding.OutcomeNegative)
}

/*
** Looks through the data and validates that the changesets have been approved
*/

func approvedRun(reviewData *checker.CodeReviewData, fs embed.FS, probeID string,
	positiveOutcome, negativeOutcome finding.Outcome,
) ([]finding.Finding, string, error) {
	var findings []finding.Finding
	var unapprovedReviews = 0
	changesets := reviewData.DefaultBranchChangesets
	var numChangesets = len(changesets)
	for i := range changesets {
		data := &changesets[i]
		if data.Author.Login == "" {
			return utils.AuthorNotFound(findings, probeID, fs)
		}
		for i := range data.Reviews {
			if data.Reviews[i].State != "APPROVED" || data.Reviews[i].Author.Login == data.Author.Login {
				unapprovedReviews += 1
			}
		}
	}
	if unapprovedReviews == 0 {
		f, err := finding.NewWith(fs, probeID, fmt.Sprintf("All %v changesets approved.", numChangesets),
			nil, positiveOutcome)
		if err != nil {
			return nil, probeID, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	} else {
		f, err := finding.NewWith(fs, probeID, fmt.Sprintf("%v unapproved reviews found among %v changesets.", unapprovedReviews, numChangesets),
			nil, negativeOutcome)
		if err != nil {
			return nil, probeID, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	}
	return findings, probeID, nil
}
