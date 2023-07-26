// © 2023 Nokia
// Licensed under the Apache License 2.0
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"github.com/ossf/scorecard/v4/finding"
	"embed"
	"fmt"
)

// returns the number of unique reviewers that aren't the changeset author
func UniqueReviewers(authorLogin string, reviewers []string) int {
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

func AuthorNotFound(findings []finding.Finding, probeID string,
	fs embed.FS) ([]finding.Finding, string, error) {
	f, err := finding.NewNotAvailable(fs, probeID, fmt.Sprintf("Could not retrieve the author of a changeset."), nil)
	if err != nil {
		return nil, probeID, fmt.Errorf("create finding: %w", err)
	}
	findings = append(findings, *f)
	return findings, probeID, nil
}

func ReviewerNotFound(findings []finding.Finding, probeID string,
	fs embed.FS) ([]finding.Finding, string, error) {
		f, err := finding.NewNegative(fs, probeID, fmt.Sprintf("Could not retrieve reviewers of a changeset."), nil)
		if err != nil {
			return nil, probeID, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
		return findings, probeID, nil
	}
