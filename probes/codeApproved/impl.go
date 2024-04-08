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
	"errors"
	"fmt"
	"strconv"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/internal/probes"
	"github.com/ossf/scorecard/v4/probes/internal/utils/uerror"
)

func init() {
	probes.MustRegister(Probe, Run, []probes.CheckName{probes.CodeReview})
}

var (
	//go:embed *.yml
	fs embed.FS

	errNoAuthor   = errors.New("could not retrieve changeset author")
	errNoReviewer = errors.New("could not retrieve the changeset reviewer")
)

const (
	Probe          = "codeApproved"
	NumApprovedKey = "approvedChangesets"
	NumTotalKey    = "totalChangesets"
)

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}
	rawReviewData := &raw.CodeReviewResults
	return approvedRun(rawReviewData, fs, Probe)
}

// Looks through the data and validates that each changeset has been approved at least once.
func approvedRun(reviewData *checker.CodeReviewData, fs embed.FS, probeID string) ([]finding.Finding, string, error) {
	changesets := reviewData.DefaultBranchChangesets
	var findings []finding.Finding

	if len(changesets) == 0 {
		f, err := finding.NewWith(fs, Probe, "no changesets detected", nil, finding.OutcomeNotApplicable)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
		return findings, Probe, nil
	}

	foundHumanActivity := false
	nChangesets := len(changesets)
	nChanges := 0
	nApproved := 0

	for x := range changesets {
		data := &changesets[x]
		approvedChangeset, err := approved(data)
		if err != nil {
			f, err := finding.NewWith(fs, probeID, err.Error(), nil, finding.OutcomeError)
			if err != nil {
				return nil, probeID, fmt.Errorf("create finding: %w", err)
			}
			findings = append(findings, *f)
			return findings, probeID, nil
		}
		// skip bot authored changesets, which can skew single maintainer projects which otherwise dont code review
		// https://github.com/ossf/scorecard/issues/2450
		if approvedChangeset && data.Author.IsBot {
			continue
		}
		nChanges += 1
		if !data.Author.IsBot {
			foundHumanActivity = true
		}
		if approvedChangeset {
			nApproved += 1
		}
	}
	var outcome finding.Outcome
	var reason string
	switch {
	case nApproved != nChanges:
		outcome = finding.OutcomeFalse
		reason = fmt.Sprintf("Found %d/%d approved changesets", nApproved, nChanges)
	case !foundHumanActivity:
		outcome = finding.OutcomeNotApplicable
		reason = fmt.Sprintf("Found no human activity in the last %d changesets", nChangesets)
	default:
		outcome = finding.OutcomeTrue
		reason = "All changesets approved"
	}
	f, err := finding.NewWith(fs, probeID, reason, nil, outcome)
	if err != nil {
		return nil, probeID, fmt.Errorf("create finding: %w", err)
	}
	f.WithValue(NumApprovedKey, strconv.Itoa(nApproved))
	f.WithValue(NumTotalKey, strconv.Itoa(nChanges))
	findings = append(findings, *f)
	return findings, probeID, nil
}

func approved(c *checker.Changeset) (bool, error) {
	if c.Author.Login == "" {
		return false, errNoAuthor
	}
	for _, review := range c.Reviews {
		if review.Author.Login == "" {
			return false, errNoReviewer
		}
		if review.State == "APPROVED" && review.Author.Login != c.Author.Login {
			return true, nil
		}
	}
	return false, nil
}
