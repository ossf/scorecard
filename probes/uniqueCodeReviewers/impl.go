// © 2023 Nokia
// Licensed under the Apache License 2.0
// SPDX-License-Identifier: Apache-2.0

package uniqueCodeReviewers

import (
	"embed"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/utils"
)

//go:embed *.yml
var fs embed.FS

const probe = "minimumCodeReviewers"

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	reviewData := raw.CodeReviewResults.DefaultBranchChangesets
	return utils.CodeReviewRun(reviewData, fs, probe, finding.OutcomePositive, finding.OutcomeNegative)
}