// © 2023 Nokia
// Licensed under the Apache License 2.0
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"embed"
	"fmt"
	//"os"
	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
)

var minimumReviewers = 2

// Looks through the data and validates author and reviewers of a changeset
func CodeReviewRun(reviewData []checker.Changeset, fs embed.FS, probeID string,
	positiveOutcome, negativeOutcome finding.Outcome,
	) ([]finding.Finding, string, error) {
	var findings []finding.Finding
	leastFoundReviewers := 0
	for i := range reviewData {
		data := &reviewData[i]
		if data.ReviewPlatform == "Unknown" && data.Author.Login == "" {
			continue
		}
		reviewersList := make([]string, len(data.Reviews))
		for i := range data.Reviews {
			reviewersList[i] = data.Reviews[i].Author.Login
		}
		numReviewers := uniqueReviewers(data.Author.Login, reviewersList)
		if i == 0 || numReviewers < leastFoundReviewers {
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
	
// returns the number of unique reviewers that aren't the changeset author
func uniqueReviewers(authorLogin string, reviewers []string) int {
	uniqueReviewers := 0
	for i := range reviewers {
		duplicateCount := 0
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