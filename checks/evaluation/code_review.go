// Copyright 2021 OpenSSF Scorecard Authors
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

package evaluation

import (
	"fmt"
	"strings"

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
)

type reviewScore int

// TODO(raghavkaul) More partial credit? E.g. approval from non-contributor, discussion liveness,
// number of resolved comments, number of approvers (more eyes on a project).
const (
	noReview              reviewScore = 0 // No approving review before merge
	changesReviewed       reviewScore = 1 // Changes were reviewed
	reviewedOutsideGithub reviewScore = 1 // Full marks until we can check review platforms outside of GitHub
)

// CodeReview applies the score policy for the Code-Review check.
func CodeReview(name string, dl checker.DetailLogger, r *checker.CodeReviewData) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	if len(r.DefaultBranchChangesets) == 0 {
		return checker.CreateInconclusiveResult(name, "no commits found")
	}

	foundHumanReviewActivity := false
	foundBotReviewActivity := false
	nUnreviewedHumanChanges := 0
	nUnreviewedBotChanges := 0

	var botReviewActivityRevisionIDs, unreviewedHumanRevisionIDs, unreviewedBotRevisionIDs []string
	for i := range r.DefaultBranchChangesets {
		cs := &r.DefaultBranchChangesets[i]
		isReviewed := reviewScoreForChangeset(cs) >= changesReviewed
		isBotCommit := cs.Author.IsBot

		switch {
		case isReviewed && isBotCommit:
			foundBotReviewActivity = true
			botReviewActivityRevisionIDs = append(botReviewActivityRevisionIDs, cs.RevisionID)
		case isReviewed && !isBotCommit:
			foundHumanReviewActivity = true
		case !isReviewed && isBotCommit:
			nUnreviewedBotChanges += 1
			unreviewedBotRevisionIDs = append(unreviewedBotRevisionIDs, cs.RevisionID)
		case !isReviewed && !isBotCommit:
			nUnreviewedHumanChanges += 1
			unreviewedHumanRevisionIDs = append(unreviewedHumanRevisionIDs, cs.RevisionID)
		}
	}

	// Let's include non-compliant revision IDs in details
	if len(botReviewActivityRevisionIDs) > 0 {
		dl.Warn(&checker.LogMessage{
			Text: fmt.Sprintf("List of revision IDs by bots:%s", strings.Join(botReviewActivityRevisionIDs, ",")),
		})
	}

	if len(unreviewedHumanRevisionIDs) > 0 {
		dl.Warn(&checker.LogMessage{
			Text: fmt.Sprintf("List of revision IDs not reviewed by humans:%s", strings.Join(unreviewedHumanRevisionIDs, ",")),
		})
	}

	if len(unreviewedBotRevisionIDs) > 0 {
		dl.Warn(&checker.LogMessage{
			Text: fmt.Sprintf("List of bot revision IDs not reviewed by humans:%s", strings.Join(unreviewedBotRevisionIDs, ",")),
		})
	}

	if foundBotReviewActivity && !foundHumanReviewActivity {
		reason := fmt.Sprintf("found no human review activity in the last %v changesets", len(r.DefaultBranchChangesets))
		return checker.CreateInconclusiveResult(name, reason)
	}

	switch {
	case nUnreviewedBotChanges > 0 && nUnreviewedHumanChanges > 0:
		return checker.CreateMinScoreResult(
			name,
			fmt.Sprintf("found unreviewed changesets (%v human %v bot)", nUnreviewedHumanChanges, nUnreviewedBotChanges),
		)
	case nUnreviewedBotChanges > 0:
		return checker.CreateResultWithScore(
			name,
			fmt.Sprintf("all human changesets reviewed, found %v unreviewed bot changesets", nUnreviewedBotChanges),
			7,
		)
	case nUnreviewedHumanChanges > 0:
		score := 3
		if len(r.DefaultBranchChangesets) == 1 || nUnreviewedHumanChanges > 1 {
			score = 0
		}
		return checker.CreateResultWithScore(
			name,
			fmt.Sprintf("found %v unreviewed human changesets", nUnreviewedHumanChanges),
			score,
		)
	}

	return checker.CreateMaxScoreResult(name, "all changesets reviewed")
}

func reviewScoreForChangeset(changeset *checker.Changeset) (score reviewScore) {
	if changeset.ReviewPlatform != "" && changeset.ReviewPlatform != checker.ReviewPlatformGitHub {
		return reviewedOutsideGithub
	}

	if changeset.ReviewPlatform == checker.ReviewPlatformGitHub {
		for i := range changeset.Reviews {
			review := changeset.Reviews[i]
			if review.State == "APPROVED" && review.Author.Login != changeset.Author.Login {
				return changesReviewed
			}
		}
	}

	return noReview
}
