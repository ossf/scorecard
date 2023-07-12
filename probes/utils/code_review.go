// © 2023 Nokia
// Licensed under the Apache License 2.0
// SPDX-License-Identifier: Apache-2.0

package utils
	
// returns the number of unique reviewers that aren't the changeset author
func UniqueReviewers(authorLogin string, reviewers []string) int {
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