// © 2023 Nokia
// Licensed under the Apache License 2.0
// SPDX-License-Identifier: Apache-2.0

package codeReviewTwoReviewers

import (
	"embed"
	"fmt"
	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
)

//go:embed *.yml
var fs embed.FS

const probe = "codeReviewTwoReviewers"
const minimumReviewers = 2
const noReviewerFound = -1

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	rawReviewData := &raw.CodeReviewResults
	return codeReviewRun(rawReviewData, fs, probe, finding.OutcomePositive, finding.OutcomeNegative)
}

/*
** Looks through the data and validates author and reviewers of a changeset
** Scorecard currently only supports GitHub revisions and generates a positive
** score in the case of other platforms. This probe is created to ensure that
** there are a number of unique reviewers for each changeset.
 */

func codeReviewRun(reviewData *checker.CodeReviewData, fs embed.FS, probeID string,
	positiveOutcome, negativeOutcome finding.Outcome,
) ([]finding.Finding, string, error) {
	var findings []finding.Finding
	leastFoundReviewers := 0
	changesets := reviewData.DefaultBranchChangesets
	for i := range changesets {
		data := &changesets[i]
		if data.Author.Login == "" {
			return authorNotFound(findings, probeID, fs)
		}
		reviewersList := make([]string, len(data.Reviews))
		for i := range data.Reviews {
			reviewersList[i] = data.Reviews[i].Author.Login
		}
		numReviewers := uniqueReviewers(data.Author.Login, reviewersList)
		if numReviewers == noReviewerFound {
			return reviewerNotFound(findings, probeID, fs)
		} else if i == 0 || numReviewers < leastFoundReviewers {
			leastFoundReviewers = numReviewers
		}
	}
	if leastFoundReviewers >= minimumReviewers {
		f, err := finding.NewWith(fs, probeID, fmt.Sprintf("%v unique reviewers found for at least one changeset, %v wanted.", leastFoundReviewers, minimumReviewers),
			nil, positiveOutcome)
		if err != nil {
			return nil, probeID, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	} else {
		f, err := finding.NewWith(fs, probeID, fmt.Sprintf("%v unique reviewer(s) found for at least one changeset, %v wanted.", leastFoundReviewers, minimumReviewers),
			nil, negativeOutcome)
		if err != nil {
			return nil, probeID, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	}
	return findings, probeID, nil
}

func uniqueReviewers(authorLogin string, reviewers []string) int {
	uniqueReviewers := 0
	for i := range reviewers {
		duplicateCount := 0
		if (reviewers[i] == "") {
			return -1
		}
		for j := range reviewers {
			if reviewers[j] == reviewers[i] && j > i {
				duplicateCount++
			}
		}
		if reviewers[i] != authorLogin && duplicateCount == 0 {
			uniqueReviewers++
		}
	}
	return uniqueReviewers
}

func authorNotFound(findings []finding.Finding, probeID string,
	fs embed.FS) ([]finding.Finding, string, error) {
	f, err := finding.NewNotAvailable(fs, probeID, fmt.Sprintf("Could not retrieve the author of a changeset."), nil)
	if err != nil {
		return nil, probeID, fmt.Errorf("create finding: %w", err)
	}
	findings = append(findings, *f)
	return findings, probeID, nil
}

func reviewerNotFound(findings []finding.Finding, probeID string,
	fs embed.FS) ([]finding.Finding, string, error) {
		f, err := finding.NewNegative(fs, probeID, fmt.Sprintf("Could not retrieve reviewers of a changeset."), nil)
		if err != nil {
			return nil, probeID, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
		return findings, probeID, nil
}
