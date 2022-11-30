// Copyright 2020 OpenSSF Scorecard Authors
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
	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
)

const (
	minReviews = 2
	// Points incremented at each level.
	adminNonAdminBasicLevel     = 3 // Level 1.
	adminNonAdminReviewLevel    = 3 // Level 2.
	nonAdminContextLevel        = 2 // Level 3.
	nonAdminThoroughReviewLevel = 1 // Level 4.
	adminThoroughReviewLevel    = 1 // Level 5.

)

type scoresInfo struct {
	basic               int
	adminBasic          int
	review              int
	adminReview         int
	context             int
	thoroughReview      int
	adminThoroughReview int
	codeownerReview     int
}

// Maximum score depending on whether admin token is used.
type levelScore struct {
	scores scoresInfo // Score result for a branch.
	maxes  scoresInfo // Maximum possible score for a branch.
}

// BranchProtection runs Branch-Protection check.
func BranchProtection(name string, dl checker.DetailLogger,
	r *checker.BranchProtectionsData,
) checker.CheckResult {
	var scores []levelScore

	// Check protections on all the branches.
	for i := range r.Branches {
		var score levelScore
		b := r.Branches[i]

		// Protected field only indates that the branch matches
		// one `Branch protection rules`. All settings may be disabled,
		// so it does not provide any guarantees.
		protected := !(b.Protected != nil && !*b.Protected)
		if !protected {
			dl.Warn(&checker.LogMessage{
				Text: fmt.Sprintf("branch protection not enabled for branch '%s'", *b.Name),
			})
		}
		score.scores.basic, score.maxes.basic = basicNonAdminProtection(&b, dl)
		score.scores.adminBasic, score.maxes.adminBasic = basicAdminProtection(&b, dl)
		score.scores.review, score.maxes.review = nonAdminReviewProtection(&b)
		score.scores.adminReview, score.maxes.adminReview = adminReviewProtection(&b, dl)
		score.scores.context, score.maxes.context = nonAdminContextProtection(&b, dl)
		score.scores.thoroughReview, score.maxes.thoroughReview = nonAdminThoroughReviewProtection(&b, dl)
		// Do we want this?
		score.scores.adminThoroughReview, score.maxes.adminThoroughReview = adminThoroughReviewProtection(&b, dl)
		score.scores.codeownerReview, score.maxes.codeownerReview = codeownerBranchProtection(&b, r.CodeownersFiles, dl)

		scores = append(scores, score)
	}

	if len(scores) == 0 {
		return checker.CreateInconclusiveResult(name, "unable to detect any development/release branches")
	}

	score, err := computeScore(scores)
	if err != nil {
		return checker.CreateRuntimeErrorResult(name, err)
	}

	switch score {
	case checker.MinResultScore:
		return checker.CreateMinScoreResult(name,
			"branch protection not enabled on development/release branches")
	case checker.MaxResultScore:
		return checker.CreateMaxScoreResult(name,
			"branch protection is fully enabled on development and all release branches")
	default:
		return checker.CreateResultWithScore(name,
			"branch protection is not maximal on development and all release branches", score)
	}
}

func computeNonAdminBasicScore(scores []levelScore) int {
	score := 0
	for i := range scores {
		s := scores[i]
		score += s.scores.basic
	}
	return score
}

func computeAdminBasicScore(scores []levelScore) int {
	score := 0
	for i := range scores {
		s := scores[i]
		score += s.scores.adminBasic
	}
	return score
}

func computeNonAdminReviewScore(scores []levelScore) int {
	score := 0
	for i := range scores {
		s := scores[i]
		score += s.scores.review
	}
	return score
}

func computeAdminReviewScore(scores []levelScore) int {
	score := 0
	for i := range scores {
		s := scores[i]
		score += s.scores.adminReview
	}
	return score
}

func computeNonAdminThoroughReviewScore(scores []levelScore) int {
	score := 0
	for i := range scores {
		s := scores[i]
		score += s.scores.thoroughReview
	}
	return score
}

func computeAdminThoroughReviewScore(scores []levelScore) int {
	score := 0
	for i := range scores {
		s := scores[i]
		score += s.scores.adminThoroughReview
	}
	return score
}

func computeNonAdminContextScore(scores []levelScore) int {
	score := 0
	for i := range scores {
		s := scores[i]
		score += s.scores.context
	}
	return score
}

func computeCodeownerThoroughReviewScore(scores []levelScore) int {
	score := 0
	for i := range scores {
		s := scores[i]
		score += s.scores.codeownerReview
	}
	return score
}

func noarmalizeScore(score, max, level int) float64 {
	if max == 0 {
		return float64(level)
	}
	return float64(score*level) / float64(max)
}

func computeScore(scores []levelScore) (int, error) {
	if len(scores) == 0 {
		return 0, sce.WithMessage(sce.ErrScorecardInternal, "scores are empty")
	}

	score := float64(0)
	maxScore := scores[0].maxes

	// First, check if they all pass the basic (admin and non-admin) checks.
	maxBasicScore := maxScore.basic * len(scores)
	maxAdminBasicScore := maxScore.adminBasic * len(scores)
	basicScore := computeNonAdminBasicScore(scores)
	adminBasicScore := computeAdminBasicScore(scores)
	score += noarmalizeScore(basicScore+adminBasicScore, maxBasicScore+maxAdminBasicScore, adminNonAdminBasicLevel)
	if basicScore != maxBasicScore ||
		adminBasicScore != maxAdminBasicScore {
		return int(score), nil
	}

	// Second, check the (admin and non-admin) reviews.
	maxReviewScore := maxScore.review * len(scores)
	maxAdminReviewScore := maxScore.adminReview * len(scores)
	reviewScore := computeNonAdminReviewScore(scores)
	adminReviewScore := computeAdminReviewScore(scores)
	score += noarmalizeScore(reviewScore+adminReviewScore, maxReviewScore+maxAdminReviewScore, adminNonAdminReviewLevel)
	if reviewScore != maxReviewScore ||
		adminReviewScore != maxAdminReviewScore {
		return int(score), nil
	}

	// Third, check the use of non-admin context.
	maxContextScore := maxScore.context * len(scores)
	contextScore := computeNonAdminContextScore(scores)
	score += noarmalizeScore(contextScore, maxContextScore, nonAdminContextLevel)
	if contextScore != maxContextScore {
		return int(score), nil
	}

	// Fourth, check the thorough non-admin reviews.
	// Also check whether this repo requires codeowner review
	maxThoroughReviewScore := maxScore.thoroughReview * len(scores)
	maxCodeownerReviewScore := maxScore.codeownerReview * len(scores)
	thoroughReviewScore := computeNonAdminThoroughReviewScore(scores)
	codeownerReviewScore := computeCodeownerThoroughReviewScore(scores)
	score += noarmalizeScore(thoroughReviewScore+codeownerReviewScore, maxThoroughReviewScore+maxCodeownerReviewScore,
		nonAdminThoroughReviewLevel)
	if thoroughReviewScore != maxThoroughReviewScore {
		return int(score), nil
	}

	// Lastly, check the thorough admin review config.
	// This one is controversial and has usability issues
	// https://github.com/ossf/scorecard/issues/1027, so we may remove it.
	maxAdminThoroughReviewScore := maxScore.adminThoroughReview * len(scores)
	adminThoroughReviewScore := computeAdminThoroughReviewScore(scores)
	score += noarmalizeScore(adminThoroughReviewScore, maxAdminThoroughReviewScore, adminThoroughReviewLevel)
	if adminThoroughReviewScore != maxAdminThoroughReviewScore {
		return int(score), nil
	}

	return int(score), nil
}

func info(dl checker.DetailLogger, doLogging bool, desc string, args ...interface{}) {
	if !doLogging {
		return
	}

	dl.Info(&checker.LogMessage{
		Text: fmt.Sprintf(desc, args...),
	})
}

func debug(dl checker.DetailLogger, doLogging bool, desc string, args ...interface{}) {
	if !doLogging {
		return
	}

	dl.Debug(&checker.LogMessage{
		Text: fmt.Sprintf(desc, args...),
	})
}

func warn(dl checker.DetailLogger, doLogging bool, desc string, args ...interface{}) {
	if !doLogging {
		return
	}

	dl.Warn(&checker.LogMessage{
		Text: fmt.Sprintf(desc, args...),
	})
}

func basicNonAdminProtection(branch *clients.BranchRef, dl checker.DetailLogger) (int, int) {
	score := 0
	max := 0
	// Only log information if the branch is protected.
	log := branch.Protected != nil && *branch.Protected

	max++
	if branch.BranchProtectionRule.AllowForcePushes != nil {
		switch *branch.BranchProtectionRule.AllowForcePushes {
		case true:
			warn(dl, log, "'force pushes' enabled on branch '%s'", *branch.Name)
		case false:
			info(dl, log, "'force pushes' disabled on branch '%s'", *branch.Name)
			score++
		}
	}

	max++
	if branch.BranchProtectionRule.AllowDeletions != nil {
		switch *branch.BranchProtectionRule.AllowDeletions {
		case true:
			warn(dl, log, "'allow deletion' enabled on branch '%s'", *branch.Name)
		case false:
			info(dl, log, "'allow deletion' disabled on branch '%s'", *branch.Name)
			score++
		}
	}

	return score, max
}

func basicAdminProtection(branch *clients.BranchRef, dl checker.DetailLogger) (int, int) {
	score := 0
	max := 0
	// Only log information if the branch is protected.
	log := branch.Protected != nil && *branch.Protected

	// nil typically means we do not have access to the value.
	if branch.BranchProtectionRule.EnforceAdmins != nil {
		// Note: we don't inrecase max possible score for non-admin viewers.
		max++
		switch *branch.BranchProtectionRule.EnforceAdmins {
		case true:
			info(dl, log, "settings apply to administrators on branch '%s'", *branch.Name)
			score++
		case false:
			warn(dl, log, "settings do not apply to administrators on branch '%s'", *branch.Name)
		}
	} else {
		debug(dl, log, "unable to retrieve whether or not settings apply to administrators on branch '%s'", *branch.Name)
	}

	return score, max
}

func nonAdminContextProtection(branch *clients.BranchRef, dl checker.DetailLogger) (int, int) {
	score := 0
	max := 0
	// Only log information if the branch is protected.
	log := branch.Protected != nil && *branch.Protected

	// This means there are specific checks enabled.
	// If only `Requires status check to pass before merging` is enabled
	// but no specific checks are declared, it's equivalent
	// to having no status check at all.
	max++
	switch {
	case len(branch.BranchProtectionRule.CheckRules.Contexts) > 0:
		info(dl, log, "status check found to merge onto on branch '%s'", *branch.Name)
		score++
	default:
		warn(dl, log, "no status checks found to merge onto branch '%s'", *branch.Name)
	}
	return score, max
}

func nonAdminReviewProtection(branch *clients.BranchRef) (int, int) {
	score := 0
	max := 0

	max++
	if branch.BranchProtectionRule.RequiredPullRequestReviews.RequiredApprovingReviewCount != nil &&
		*branch.BranchProtectionRule.RequiredPullRequestReviews.RequiredApprovingReviewCount > 0 {
		// We do not display anything here, it's done in nonAdminThoroughReviewProtection()
		score++
	}
	return score, max
}

func adminReviewProtection(branch *clients.BranchRef, dl checker.DetailLogger) (int, int) {
	score := 0
	max := 0

	// Only log information if the branch is protected.
	log := branch.Protected != nil && *branch.Protected

	if branch.BranchProtectionRule.CheckRules.UpToDateBeforeMerge != nil {
		// Note: `This setting will not take effect unless at least one status check is enabled`.
		max++
		switch *branch.BranchProtectionRule.CheckRules.UpToDateBeforeMerge {
		case true:
			info(dl, log, "status checks require up-to-date branches for '%s'", *branch.Name)
			score++
		default:
			warn(dl, log, "status checks do not require up-to-date branches for '%s'", *branch.Name)
		}
	} else {
		debug(dl, log, "unable to retrieve whether up-to-date branches are needed to merge on branch '%s'", *branch.Name)
	}

	return score, max
}

func adminThoroughReviewProtection(branch *clients.BranchRef, dl checker.DetailLogger) (int, int) {
	score := 0
	max := 0
	// Only log information if the branch is protected.
	log := branch.Protected != nil && *branch.Protected

	if branch.BranchProtectionRule.RequiredPullRequestReviews.DismissStaleReviews != nil {
		// Note: we don't inrecase max possible score for non-admin viewers.
		max++
		switch *branch.BranchProtectionRule.RequiredPullRequestReviews.DismissStaleReviews {
		case true:
			info(dl, log, "Stale review dismissal enabled on branch '%s'", *branch.Name)
			score++
		case false:
			warn(dl, log, "Stale review dismissal disabled on branch '%s'", *branch.Name)
		}
	} else {
		debug(dl, log, "unable to retrieve review dismissal on branch '%s'", *branch.Name)
	}
	return score, max
}

func nonAdminThoroughReviewProtection(branch *clients.BranchRef, dl checker.DetailLogger) (int, int) {
	score := 0
	max := 0

	// Only log information if the branch is protected.
	log := branch.Protected != nil && *branch.Protected

	max++
	if branch.BranchProtectionRule.RequiredPullRequestReviews.RequiredApprovingReviewCount != nil {
		switch *branch.BranchProtectionRule.RequiredPullRequestReviews.RequiredApprovingReviewCount >= minReviews {
		case true:
			info(dl, log, "number of required reviewers is %d on branch '%s'",
				*branch.BranchProtectionRule.RequiredPullRequestReviews.RequiredApprovingReviewCount, *branch.Name)
			score++
		default:
			warn(dl, log, "number of required reviewers is only %d on branch '%s'",
				*branch.BranchProtectionRule.RequiredPullRequestReviews.RequiredApprovingReviewCount, *branch.Name)
		}
	} else {
		warn(dl, log, "number of required reviewers is 0 on branch '%s'", *branch.Name)
	}
	return score, max
}

func codeownerBranchProtection(
	branch *clients.BranchRef, codeownersFiles []string, dl checker.DetailLogger,
) (int, int) {
	score := 0
	max := 1

	log := branch.Protected != nil && *branch.Protected

	if branch.BranchProtectionRule.RequiredPullRequestReviews.RequireCodeOwnerReviews != nil {
		switch *branch.BranchProtectionRule.RequiredPullRequestReviews.RequireCodeOwnerReviews {
		case true:
			info(dl, log, "codeowner review is required on branch '%s'", *branch.Name)
			if len(codeownersFiles) == 0 {
				warn(dl, log, "codeowners branch protection is being ignored - but no codeowners file found in repo")
			} else {
				score++
			}
		default:
			warn(dl, log, "codeowner review is not required on branch '%s'", *branch.Name)
		}
	}

	return score, max
}
