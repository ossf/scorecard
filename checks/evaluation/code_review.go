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
	"math"

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

	N := len(r.DefaultBranchChangesets)
	if N == 0 {
		return checker.CreateInconclusiveResult(name, "no commits found")
	}

	nUnreviewedChanges := 0
	nChanges := 0
	foundHumanActivity := false

	for i := range r.DefaultBranchChangesets {
		cs := &r.DefaultBranchChangesets[i]
		isReviewed := reviewScoreForChangeset(cs, dl) >= changesReviewed
		if isReviewed && cs.Author.IsBot {
			continue // ignore reviewed bot commits
		}

		nChanges += 1

		if !cs.Author.IsBot {
			foundHumanActivity = true
		}

		if !isReviewed {
			nUnreviewedChanges += 1
		}
	}

	switch {
	case nChanges == 0 || !foundHumanActivity:
		reason := fmt.Sprintf("found no human review activity in the last %v changesets", N)
		return checker.CreateInconclusiveResult(name, reason)
	case nUnreviewedChanges > 0:
		return checker.CreateProportionalScoreResult(
			name,
			fmt.Sprintf(
				"found %d unreviewed changesets out of %d", nUnreviewedChanges,
				nChanges,
			),
			int(math.Max(float64(nChanges-nUnreviewedChanges), 0)),
			nChanges,
		)
	}

	return checker.CreateMaxScoreResult(name, "all changesets reviewed")
}

func reviewScoreForChangeset(changeset *checker.Changeset, dl checker.DetailLogger) (score reviewScore) {
	plat := changeset.ReviewPlatform
	if plat != checker.ReviewPlatformUnknown &&
		plat != checker.ReviewPlatformGitHub {
		return reviewedOutsideGithub
	}

	if plat == checker.ReviewPlatformGitHub {
		for i := range changeset.Reviews {
			review := changeset.Reviews[i]
			if review.State == "APPROVED" && review.Author.Login != changeset.Author.Login {
				return changesReviewed
			}
		}
	}

	dl.Debug(
		&checker.LogMessage{
			Text: fmt.Sprintf(
				"couldn't find approvals for revision: %s platform: %s",
				changeset.RevisionID, plat,
			),
		},
	)
	return noReview
}
