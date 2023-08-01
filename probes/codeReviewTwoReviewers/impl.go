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

package codeReviewTwoReviewers

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/utils"
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
	changesets := reviewData.DefaultBranchChangesets
	var findings []finding.Finding
	var leastFoundReviewers = 0
	var numChangesets = len(changesets)
	var numBotAuthors = 0
	if numChangesets == 0 {
		return nil, probeID, fmt.Errorf("%w", utils.NoChangesetsErr)
	}
	// Loops through all changesets, if an author login cannot be retrieved: returns OutcomeNotAvailabe.
	// leastFoundReviewers will be the lowest number of unique reviewers found among the changesets.
	for i := range changesets {
		data := &changesets[i]
		if data.Author.Login == "" {
			f, err := finding.NewNotAvailable(fs, probeID, fmt.Sprintf("Could not retrieve the author of a changeset."), nil)
			if err != nil {
				return nil, probeID, fmt.Errorf("create finding: %w", err)
			}
			findings = append(findings, *f)
			return findings, probeID, nil
		} else if data.Author.IsBot == true {
			numBotAuthors += 1
		}
		numReviewers, err := uniqueReviewers(data.Author.Login, data.Reviews)
		if err != nil {
			f, err := finding.NewNotAvailable(fs, probeID, fmt.Sprintf("Could not retrieve the reviewer of a changeset."), nil)
			if err != nil {
				return nil, probeID, fmt.Errorf("create finding: %w", err)
			}
			findings = append(findings, *f)
			return findings, probeID, nil
		} else if i == 0 || numReviewers < leastFoundReviewers {
			leastFoundReviewers = numReviewers
		}
	}
	if numBotAuthors == numChangesets {
		// returns a NotAvailable outcome if all changesets were authored by bots
		f, err := finding.NewNotAvailable(fs, probeID, fmt.Sprintf("All changesets authored by bot(s)."), nil)
		if err != nil {
			return nil, probeID, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
		return findings, probeID, nil
	} else if leastFoundReviewers >= minimumReviewers {
		// returns PositiveOutcome if the lowest number of unique reviewers is at least as high as minimumReviewers (2).
		f, err := finding.NewWith(fs, probeID, fmt.Sprintf("%v unique reviewers found for at least one changeset, %v wanted.", leastFoundReviewers, minimumReviewers),
			nil, positiveOutcome)
		if err != nil {
			return nil, probeID, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	} else {
		// returns NegativeOutcome if even a single changeset was reviewed by fewer than minimumReviewers (2).
		f, err := finding.NewWith(fs, probeID, fmt.Sprintf("%v unique reviewer(s) found for at least one changeset, %v wanted.", leastFoundReviewers, minimumReviewers),
			nil, negativeOutcome)
		if err != nil {
			return nil, probeID, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	}
	return findings, probeID, nil
}

// Loops through the reviews of a changeset, counting how many unique user logins are present.
// Reviews performed by the author don't count, and an error is returned if a reviewer login can't be retrieved.
func uniqueReviewers(authorLogin string, reviews []clients.Review) (int, error) {
	reviewersList := make([]string, len(reviews))
	for i := range reviewersList {
		reviewersList[i] = reviews[i].Author.Login
	}
	uniqueReviewers := 0
	for i := range reviewersList {
		duplicateCount := 0
		if reviewersList[i] == "" {
			return uniqueReviewers, fmt.Errorf("could not find the login of a reviewer")
		}
		for j := range reviewersList {
			if reviewersList[j] == reviewersList[i] && j > i {
				duplicateCount++
			}
		}
		if reviewersList[i] != authorLogin && duplicateCount == 0 {
			uniqueReviewers++
		}
	}
	return uniqueReviewers, nil
}
