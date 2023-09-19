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
	basicLevel                  = 3 // Level 1.
	adminNonAdminReviewLevel    = 3 // Level 2.
	nonAdminContextLevel        = 2 // Level 3.
	nonAdminThoroughReviewLevel = 1 // Level 4.
	adminThoroughReviewLevel    = 1 // Level 5.

)

type scoresInfo struct {
	basic               int
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

// Evaluates if Scorecard is being run with an administrator token.
func isUserAdmin(branchProtectionData *clients.BranchRef) bool {
	// When Scorecard is run without the admin token, Github retrieves both of the following fields as nil,
	// so we're using them to evaluate if Scorecard is run using admin token or not.
	return branchProtectionData.BranchProtectionRule.CheckRules.UpToDateBeforeMerge != nil ||
		branchProtectionData.BranchProtectionRule.RequireLastPushApproval != nil
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

	score, err := computeFinalScore(scores, dl)
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

func sumUpScoreForTier(tier int, scoresData []levelScore, dl checker.DetailLogger) int {
	sum := 0
	for i := range scoresData {
		score := scoresData[i]
		switch tier {
		case 1:
			sum += score.scores.basic
		case 2:
			sum += score.scores.review + score.scores.adminReview
		case 3:
			sum += score.scores.context
		case 4:
			sum += score.scores.thoroughReview + score.scores.codeownerReview
		case 5:
			sum += score.scores.adminThoroughReview
		default:
			debug(dl, true, "Function sumUpScoreForTier called with the invalid parameter: '%d';"+
				"BranchProtection score won't be accurate.", tier)
		}
	}
	return sum
}

func normalizeScore(score, max, level int) float64 {
	if max == 0 {
		return float64(level)
	}
	return float64(score*level) / float64(max)
}

func computeFinalScore(scores []levelScore, dl checker.DetailLogger) (int, error) {
	if len(scores) == 0 {
		return 0, sce.WithMessage(sce.ErrScorecardInternal, "scores are empty")
	}

	score := float64(0)
	maxScore := scores[0].maxes

	// First, check if they all pass the basic (admin and non-admin) checks.
	maxBasicScore := maxScore.basic * len(scores)
	basicScore := sumUpScoreForTier(1, scores, dl)
	score += normalizeScore(basicScore, maxBasicScore, basicLevel)
	if basicScore < maxBasicScore {
		return int(score), nil
	}

	// Second, check the (admin and non-admin) reviews.
	maxReviewScore := maxScore.review * len(scores)
	maxAdminReviewScore := maxScore.adminReview * len(scores)
	adminNonAdminReviewScore := sumUpScoreForTier(2, scores, dl)
	score += normalizeScore(adminNonAdminReviewScore, maxReviewScore+maxAdminReviewScore, adminNonAdminReviewLevel)
	if adminNonAdminReviewScore < maxReviewScore+maxAdminReviewScore {
		return int(score), nil
	}

	// Third, check the use of non-admin context.
	maxContextScore := maxScore.context * len(scores)
	contextScore := sumUpScoreForTier(3, scores, dl)
	score += normalizeScore(contextScore, maxContextScore, nonAdminContextLevel)
	if contextScore < maxContextScore {
		return int(score), nil
	}

	// Fourth, check the thorough non-admin reviews.
	// Also check whether this repo requires codeowner review
	maxThoroughReviewScore := maxScore.thoroughReview * len(scores)
	maxCodeownerReviewScore := maxScore.codeownerReview * len(scores)
	tier4Score := sumUpScoreForTier(4, scores, dl)
	score += normalizeScore(tier4Score, maxThoroughReviewScore+maxCodeownerReviewScore, nonAdminThoroughReviewLevel)
	if tier4Score < maxThoroughReviewScore+maxCodeownerReviewScore {
		return int(score), nil
	}

	// Lastly, check the thorough admin review config.
	// This one is controversial and has usability issues
	// https://github.com/ossf/scorecard/issues/1027, so we may remove it.
	maxAdminThoroughReviewScore := maxScore.adminThoroughReview * len(scores)
	adminThoroughReviewScore := sumUpScoreForTier(5, scores, dl)
	score += normalizeScore(adminThoroughReviewScore, maxAdminThoroughReviewScore, adminThoroughReviewLevel)
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

	max += 2
	if branch.BranchProtectionRule.RequiredPullRequestReviews.RequiredApprovingReviewCount != nil &&
		*branch.BranchProtectionRule.RequiredPullRequestReviews.RequiredApprovingReviewCount > 0 {
		// We do not display anything here, it's done in nonAdminThoroughReviewProtection()
		score += 2
	}
	return score, max
}

func adminReviewProtection(branch *clients.BranchRef, dl checker.DetailLogger) (int, int) {
	score := 0
	max := 0

	// Only log information if the branch is protected.
	log := branch.Protected != nil && *branch.Protected

	// Process UpToDateBeforeMerge value.
	if branch.BranchProtectionRule.CheckRules.UpToDateBeforeMerge == nil {
		debug(dl, log, "unable to retrieve whether up-to-date branches are needed to merge on branch '%s'", *branch.Name)
	} else {
		// Note: `This setting will not take effect unless at least one status check is enabled`.
		max++
		if *branch.BranchProtectionRule.CheckRules.UpToDateBeforeMerge {
			info(dl, log, "status checks require up-to-date branches for '%s'", *branch.Name)
			score++
		} else {
			warn(dl, log, "status checks do not require up-to-date branches for '%s'", *branch.Name)
		}
	}

	// Process RequireLastPushApproval value.
	if branch.BranchProtectionRule.RequireLastPushApproval == nil {
		debug(dl, log, "unable to retrieve whether 'last push approval' is required to merge on branch '%s'", *branch.Name)
	} else {
		max++
		if *branch.BranchProtectionRule.RequireLastPushApproval {
			info(dl, log, "'last push approval' enabled on branch '%s'", *branch.Name)
			score++
		} else {
			warn(dl, log, "'last push approval' disabled on branch '%s'", *branch.Name)
		}
	}

	if isUserAdmin(branch) {
		// If Scorecard is run with admin token, we can interprete GitHub's response to say
		// if the branch requires PRs prior to code changes.
		max++
		if branch.BranchProtectionRule.RequiredPullRequestReviews.RequiredApprovingReviewCount != nil {
			score++
			info(dl, log, "PRs are required in order to make changes on branch '%s'", *branch.Name)
		} else {
			warn(dl, log, "PRs are not required in order to make changes on branch '%s'", *branch.Name)
		}
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
			info(dl, log, "stale review dismissal enabled on branch '%s'", *branch.Name)
			score++
		case false:
			warn(dl, log, "stale review dismissal disabled on branch '%s'", *branch.Name)
		}
	} else {
		debug(dl, log, "unable to retrieve review dismissal on branch '%s'", *branch.Name)
	}

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

func nonAdminThoroughReviewProtection(branch *clients.BranchRef, dl checker.DetailLogger) (int, int) {
	score := 0
	max := 0

	// Only log information if the branch is protected.
	log := branch.Protected != nil && *branch.Protected

	max++

	// On this first check we exclude the case of PRs don't being required, covered on adminReviewProtection function
	if !(isUserAdmin(branch) &&
		branch.BranchProtectionRule.RequiredPullRequestReviews.RequiredApprovingReviewCount == nil) {
		switch {
		// If not running as admin, the nil value can both mean that no reviews are required or no PR are required,
		// so here we assume no reviews are required.
		case (!isUserAdmin(branch) &&
			branch.BranchProtectionRule.RequiredPullRequestReviews.RequiredApprovingReviewCount == nil) ||
			*branch.BranchProtectionRule.RequiredPullRequestReviews.RequiredApprovingReviewCount == 0:
			warn(dl, log, "number of required reviewers is 0 on branch '%s'", *branch.Name)
		case *branch.BranchProtectionRule.RequiredPullRequestReviews.RequiredApprovingReviewCount >= minReviews:
			info(dl, log, "number of required reviewers is %d on branch '%s'",
				*branch.BranchProtectionRule.RequiredPullRequestReviews.RequiredApprovingReviewCount, *branch.Name)
			score++
		default:
			warn(dl, log, "number of required reviewers is only %d on branch '%s'",
				*branch.BranchProtectionRule.RequiredPullRequestReviews.RequiredApprovingReviewCount, *branch.Name)
		}
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
