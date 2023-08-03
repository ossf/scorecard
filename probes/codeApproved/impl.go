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

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	rawReviewData := &raw.CodeReviewResults
	return approvedRun(rawReviewData, fs, probe, finding.OutcomePositive, finding.OutcomeNegative)
}

// Looks through the data and validates that each changeset has been approved at least once.
func approvedRun(reviewData *checker.CodeReviewData, fs embed.FS, probeID string,
	positiveOutcome, negativeOutcome finding.Outcome,
) ([]finding.Finding, string, error) {
	changesets := reviewData.DefaultBranchChangesets
	var findings []finding.Finding
	nApprovedChangesets := 0
	foundHumanActivity := false
	nChangesets := len(changesets)
	nHumanChangesets := 0
	if nChangesets == 0 {
		return nil, probeID, utils.NoChangesetsErr
	}
	for x := range changesets {
		data := &changesets[x]
		approvedChangeset := false
		for y := range data.Reviews {
			if data.Reviews[y].State == "APPROVED" && data.Reviews[y].Author.Login != data.Author.Login {
				approvedChangeset = true
				break
			}
		}
		if approvedChangeset && data.Author.IsBot {
			continue
		}
		if !data.Author.IsBot {
			foundHumanActivity = true
			nHumanChangesets += 1
		}
		if approvedChangeset {
			nApprovedChangesets += 1
		}
	}
	switch {
	case nHumanChangesets == 0 || !foundHumanActivity:
		// returns a NotAvailable outcome if all changesets were authored by bots
		f, err := finding.NewNotAvailable(fs, probeID, fmt.Sprint("found no human review activity " +
		"in the last %v changesets", nChangesets), nil)
		if err != nil {
			return nil, probeID, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
		return findings, probeID, nil
	case nApprovedChangesets < nHumanChangesets:
		// returns NegativeOutcome if not all changesets were approved
		f, err := finding.NewWith(fs, probeID, fmt.Sprintf("Not all changesets approved. "+
		"Found %v approvals among %v changesets.", nApprovedChangesets, nChangesets), nil, negativeOutcome)
		if err != nil {
			return nil, probeID, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	default:
		// returns PositiveOutcome if all changesets have been approved
		f, err := finding.NewWith(fs, probeID, fmt.Sprintf("All changesets approved (%v out of %v).",
			nApprovedChangesets, nChangesets), nil, positiveOutcome)
		if err != nil {
			return nil, probeID, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	}
	return findings, probeID, nil
}
