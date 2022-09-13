// Copyright 2021 Security Scorecard Authors
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

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
)

type ReviewScore = int

// TODO More partial credit? E.g. approval from non-contributor, discussion liveness,
// number of resolved comments, number of approvers (more eyes on a project).
const (
	NoReview              ReviewScore = 0 // No approving review before merge
	Reviewed              ReviewScore = 1 // Changes were reviewed
	ReviewedOutsideGithub ReviewScore = 1 // Full marks until we can check review platforms outside of GitHub
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

	score := 0
	numReviewed := 0
	for i := range r.DefaultBranchChangesets {
		score += reviewScoreForChangeset(&r.DefaultBranchChangesets[i])
		if score >= Reviewed {
			numReviewed += 1
		}
	}
	reason := fmt.Sprintf(
		"%v out of last %v changesets reviewed before merge", numReviewed, len(r.DefaultBranchChangesets),
	)

	return checker.CreateProportionalScoreResult(name, reason, score, len(r.DefaultBranchChangesets))
}

func reviewScoreForChangeset(changeset *checker.Changeset) (score ReviewScore) {
	if changeset.ReviewPlatform != "" && changeset.ReviewPlatform != checker.ReviewPlatformGitHub {
		return ReviewedOutsideGithub
	}

	mr := changeset.Commits[0].AssociatedMergeRequest

	for _, review := range mr.Reviews {
		if review.State == "APPROVED" {
			return Reviewed
		}
	}

	return NoReview
}
