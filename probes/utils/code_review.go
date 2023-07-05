// © 2023 Nokia
// Licensed under the Apache License 2.0
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
)

func CodeReviewRun(reviewData []checker.Changeset, fs embed.FS, probeID string,
	foundOutcome, notFoundOutcome finding.Outcome,
) ([]finding.Finding, string, error) {
	var findings []finding.Finding
	for i := range reviewData {
		data := &reviewData[i]
		fmt.Printf("ABACKMAN%v\n\n", data)
		fmt.Printf("ABACKMAN%v", findings)
	}
	return findings, probeID, nil
}