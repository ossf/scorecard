// Copyright 2020 Security Scorecard Authors
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

package checks

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/ossf/scorecard/v3/checker"
	"github.com/ossf/scorecard/v3/clients"
	sce "github.com/ossf/scorecard/v3/errors"
)

type branchProtectionSetting int

const (
	// CheckBranchProtection is the exported name for Branch-Protected check.
	CheckBranchProtection = "Branch-Protection"
	minReviews            = 2
	// Points incremented at each level.
	adminNonAdminBasicLevel     = 3 // Level 1.
	adminNonAdminReviewLevel    = 3 // Level 2.
	nonAdminContextLevel        = 2 // Level 3.
	nonAdminThoroughReviewLevel = 1 // Level 4.
	adminThoroughReviewLevel    = 1 // Level 5.
	// First level.
	allowForcePushes branchProtectionSetting = iota
	allowDeletions
	// Second and third level.
	requireApprovingReviewCount
	// Fourth level.
	requireStatusChecksContexts
	// Admin settings.
	// First level.
	enforceAdmins
	// Second level.
	requireUpToDateBeforeMerge
	// Fifth level.
	dismissStaleReviews
	// requireCodeOwnerReviews no longer used.
	// requireLinearHistory no longer used, see https://github.com/ossf/scorecard/issues/1027.
)

// Note: we don't need these since they are all `1`, so we may remove them.
var branchProtectionSettingScores = map[branchProtectionSetting]int{
	allowForcePushes: 1,
	allowDeletions:   1,
	// Used to check if required status checks before merging is non-empty,
	// rather than checking if requireStatusChecks is set.
	requireStatusChecksContexts: 1,
	requireApprovingReviewCount: 1,
	// Need admin token, so we cannot rely on them in general.
	enforceAdmins:              1,
	requireUpToDateBeforeMerge: 1, // Gated by requireStatusChecks=true.
	dismissStaleReviews:        1,
}

type scoresInfo struct {
	basic               int
	adminBasic          int
	review              int
	adminReview         int
	context             int
	thoroughReview      int
	adminThoroughReview int
}

// Maximum score depending on whether admin token is used.
type levelScore struct {
	scores    scoresInfo // Score result for a branch.
	maxes     scoresInfo // Maximum possible score for a branch.
	protected bool       // Protection enabled on the branch.
}

//nolint:gochecknoinits
func init() {
	registerCheck(CheckBranchProtection, BranchProtection)
}

type branchMap map[string]*clients.BranchRef

func (b branchMap) getBranchByName(name string) (*clients.BranchRef, error) {
	val, exists := b[name]
	if exists {
		return val, nil
	}

	// Ideally, we should check using repositories.GetBranch if there was a branch redirect.
	// See https://github.com/google/go-github/issues/1895
	// For now, handle the common master -> main redirect.
	if name == "master" {
		val, exists := b["main"]
		if exists {
			return val, nil
		}
	}
	return nil, fmt.Errorf("could not find branch name %s: %w", name, errInternalBranchNotFound)
}

func getBranchMapFrom(branches []*clients.BranchRef) branchMap {
	ret := make(branchMap)
	for _, branch := range branches {
		branchName := getBranchName(branch)
		if branchName != "" {
			ret[branchName] = branch
		}
	}
	return ret
}

func getBranchName(branch *clients.BranchRef) string {
	if branch == nil || branch.Name == nil {
		return ""
	}
	return *branch.Name
}

// BranchProtection runs Branch-Protection check.
func BranchProtection(c *checker.CheckRequest) checker.CheckResult {
	// Checks branch protection on both release and development branch.
	return checkReleaseAndDevBranchProtection(c.RepoClient, c.Dlogger)
}

func validateMaxScore(s1, s2 int) error {
	if s1 != s2 {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("invalid score %d != %d",
			s1, s2))
	}

	return nil
}

// This function validates that all maximum scores are the same
// for each level and branch. An error would mean a logic bug in our
// implementation.
func getMaxScores(scores []levelScore) (scoresInfo, error) {
	if len(scores) == 0 {
		return scoresInfo{}, sce.WithMessage(sce.ErrScorecardInternal, "empty score")
	}

	score := scores[0]
	for _, s := range scores[1:] {
		// Only validate the maximum scores if both entries have the same protection status.
		if s.protected != score.protected {
			continue
		}
		if err := validateMaxScore(score.maxes.basic, s.maxes.basic); err != nil {
			return scoresInfo{}, err
		}

		if err := validateMaxScore(score.maxes.adminBasic, s.maxes.adminBasic); err != nil {
			return scoresInfo{}, err
		}

		if err := validateMaxScore(score.maxes.review, s.maxes.review); err != nil {
			return scoresInfo{}, err
		}

		if err := validateMaxScore(score.maxes.adminReview, s.maxes.adminReview); err != nil {
			return scoresInfo{}, err
		}

		if err := validateMaxScore(score.maxes.context, s.maxes.context); err != nil {
			return scoresInfo{}, err
		}

		if err := validateMaxScore(score.maxes.thoroughReview, s.maxes.thoroughReview); err != nil {
			return scoresInfo{}, err
		}

		if err := validateMaxScore(score.maxes.adminThoroughReview,
			s.maxes.adminThoroughReview); err != nil {
			return scoresInfo{}, err
		}
	}

	return scoresInfo{
		basic:               score.maxes.basic,
		adminBasic:          score.maxes.adminBasic,
		review:              score.maxes.review,
		adminReview:         score.maxes.adminReview,
		context:             score.maxes.context,
		thoroughReview:      score.maxes.thoroughReview,
		adminThoroughReview: score.maxes.adminThoroughReview,
	}, nil
}

func computeNonAdminBasicScore(scores []levelScore, maxes scoresInfo) (int, int) {
	score := 0
	max := 0
	for _, s := range scores {
		score += s.scores.basic
		max += maxes.basic
	}
	return score, max
}

func computeAdminBasicScore(scores []levelScore, maxes scoresInfo) (int, int) {
	score := 0
	max := 0
	for _, s := range scores {
		score += s.scores.adminBasic
		max += maxes.adminBasic
	}
	return score, max
}

func computeNonAdminReviewScore(scores []levelScore, maxes scoresInfo) (int, int) {
	score := 0
	max := 0
	for _, s := range scores {
		score += s.scores.review
		max += maxes.review
	}
	return score, max
}

func computeAdminReviewScore(scores []levelScore, maxes scoresInfo) (int, int) {
	score := 0
	max := 0
	for _, s := range scores {
		score += s.scores.adminReview
		max += maxes.adminReview
	}
	return score, max
}

func computeNonAdminThoroughReviewScore(scores []levelScore, maxes scoresInfo) (int, int) {
	score := 0
	max := 0
	for _, s := range scores {
		score += s.scores.thoroughReview
		max += maxes.thoroughReview
	}
	return score, max
}

func computeAdminThoroughReviewScore(scores []levelScore, maxes scoresInfo) (int, int) {
	score := 0
	max := 0
	for _, s := range scores {
		score += s.scores.adminThoroughReview
		max += maxes.adminThoroughReview
	}
	return score, max
}

func computeNonAdminContextScore(scores []levelScore, maxes scoresInfo) (int, int) {
	score := 0
	max := 0
	for _, s := range scores {
		score += s.scores.context
		max += maxes.context
	}
	return score, max
}

func noarmalizeScore(score, max, level int) float64 {
	if max == 0 {
		return float64(level)
	}
	return float64(score*level) / float64(max)
}

func computeScore(scores []levelScore) (int, error) {
	// Validate and retrieve the maximum scores.
	maxScores, err := getMaxScores(scores)
	if err != nil {
		return 0, err
	}

	score := float64(0)

	// First, check if they all pass the basic (admin and non-admin) checks.
	basicScore, maxBasicScore := computeNonAdminBasicScore(scores, maxScores)
	adminBasicScore, maxAdminBasicScore := computeAdminBasicScore(scores, maxScores)
	score += noarmalizeScore(basicScore+adminBasicScore, maxBasicScore+maxAdminBasicScore, adminNonAdminBasicLevel)
	if basicScore != maxBasicScore ||
		adminBasicScore != maxAdminBasicScore {
		return int(score), nil
	}

	// Second, check the (admin and non-admin) reviews.
	reviewScore, maxReviewScore := computeNonAdminReviewScore(scores, maxScores)
	adminReviewScore, maxAdminReviewScore := computeAdminReviewScore(scores, maxScores)
	score += noarmalizeScore(reviewScore+adminReviewScore, maxReviewScore+maxAdminReviewScore, adminNonAdminReviewLevel)
	if reviewScore != maxReviewScore ||
		adminReviewScore != maxAdminReviewScore {
		return int(score), nil
	}

	// Third, check the use of non-admin context.
	contextScore, maxContextScore := computeNonAdminContextScore(scores, maxScores)
	score += noarmalizeScore(contextScore, maxContextScore, nonAdminContextLevel)
	if contextScore != maxContextScore {
		return int(score), nil
	}

	// Fourth, check the thorough non-admin reviews.
	thoroughReviewScore, maxThoroughReviewScore := computeNonAdminThoroughReviewScore(scores, maxScores)
	score += noarmalizeScore(thoroughReviewScore, maxThoroughReviewScore, nonAdminThoroughReviewLevel)
	if thoroughReviewScore != maxThoroughReviewScore {
		return int(score), nil
	}

	// Last, check the thorough admin review config.
	// This one is controversial and has usability issues
	// https://github.com/ossf/scorecard/issues/1027, so we may remove it.
	adminThoroughReviewScore, maxAdminThoroughReviewScore := computeAdminThoroughReviewScore(scores, maxScores)
	score += noarmalizeScore(adminThoroughReviewScore, maxAdminThoroughReviewScore, adminThoroughReviewLevel)
	if adminThoroughReviewScore != maxAdminThoroughReviewScore {
		return int(score), nil
	}

	return int(score), nil
}

func checkReleaseAndDevBranchProtection(
	repoClient clients.RepoClient, dl checker.DetailLogger) checker.CheckResult {
	// Get all branches. This will include information on whether they are protected.
	branches, err := repoClient.ListBranches()
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckBranchProtection, e)
	}
	branchesMap := getBranchMapFrom(branches)

	// Get release branches.
	releases, err := repoClient.ListReleases()
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckBranchProtection, e)
	}

	commit := regexp.MustCompile("^[a-f0-9]{40}$")
	checkBranches := make(map[string]bool)
	for _, release := range releases {
		if release.TargetCommitish == "" {
			// Log with a named error if target_commitish is nil.
			e := sce.WithMessage(sce.ErrScorecardInternal, errInternalCommitishNil.Error())
			return checker.CreateRuntimeErrorResult(CheckBranchProtection, e)
		}

		// TODO: if this is a sha, get the associated branch. for now, ignore.
		if commit.Match([]byte(release.TargetCommitish)) {
			continue
		}

		// Try to resolve the branch name.
		b, err := branchesMap.getBranchByName(release.TargetCommitish)
		if err != nil {
			// If the commitish branch is still not found, fail.
			return checker.CreateRuntimeErrorResult(CheckBranchProtection, err)
		}

		// Branch is valid, add to list of branches to check.
		checkBranches[*b.Name] = true
	}

	// Add default branch.
	defaultBranch, err := repoClient.GetDefaultBranch()
	if err != nil {
		return checker.CreateRuntimeErrorResult(CheckBranchProtection, err)
	}
	defaultBranchName := getBranchName(defaultBranch)
	if defaultBranchName != "" {
		checkBranches[defaultBranchName] = true
	}

	var scores []levelScore

	// Check protections on all the branches.
	for b := range checkBranches {
		var score levelScore
		branch, err := branchesMap.getBranchByName(b)
		if err != nil {
			if errors.Is(err, errInternalBranchNotFound) {
				continue
			}
			return checker.CreateRuntimeErrorResult(CheckBranchProtection, err)
		}
		// Protected field only indates that the branch matches
		// one `Branch protection rules`. All settings may be disabled,
		// so it does not provide any guarantees.
		if branch.Protected != nil && !*branch.Protected {
			dl.Warn("branch protection not enabled for branch '%s'", b)
			scores = append(scores, score)
			continue
		}

		// The branch is protected. Check the protection.
		score.protected = true
		score.scores.basic, score.maxes.basic = basicNonAdminProtection(&branch.BranchProtectionRule, b, dl)
		score.scores.adminBasic, score.maxes.adminBasic = basicAdminProtection(&branch.BranchProtectionRule, b, dl)
		score.scores.review, score.maxes.review = nonAdminReviewProtection(&branch.BranchProtectionRule)
		score.scores.adminReview, score.maxes.adminReview = adminReviewProtection(&branch.BranchProtectionRule, b, dl)
		score.scores.context, score.maxes.context = nonAdminContextProtection(&branch.BranchProtectionRule, b, dl)
		score.scores.thoroughReview, score.maxes.thoroughReview =
			nonAdminThoroughReviewProtection(&branch.BranchProtectionRule, b, dl)
		score.scores.adminThoroughReview, score.maxes.adminThoroughReview =
			adminThoroughReviewProtection(&branch.BranchProtectionRule, b, dl) // Do we want this?

		scores = append(scores, score)
	}

	if len(scores) == 0 {
		return checker.CreateInconclusiveResult(CheckBranchProtection, "unable to detect any development/release branches")
	}

	score, err := computeScore(scores)
	if err != nil {
		return checker.CreateRuntimeErrorResult(CheckBranchProtection, err)
	}

	switch score {
	case checker.MinResultScore:
		return checker.CreateMinScoreResult(CheckBranchProtection,
			"branch protection not enabled on development/release branches")
	case checker.MaxResultScore:
		return checker.CreateMaxScoreResult(CheckBranchProtection,
			"branch protection is fully enabled on development and all release branches")
	default:
		return checker.CreateResultWithScore(CheckBranchProtection,
			"branch protection is not maximal on development and all release branches", score)
	}
}

func basicNonAdminProtection(protection *clients.BranchProtectionRule,
	branch string, dl checker.DetailLogger) (int, int) {
	score := 0
	max := 0

	max += branchProtectionSettingScores[allowForcePushes]
	if protection.AllowForcePushes != nil {
		switch *protection.AllowForcePushes {
		case true:
			dl.Warn("'force pushes' enabled on branch '%s'", branch)
		case false:
			dl.Info("'force pushes' disabled on branch '%s'", branch)
			score += branchProtectionSettingScores[allowForcePushes]
		}
	}

	max += branchProtectionSettingScores[allowDeletions]
	if protection.AllowDeletions != nil {
		switch *protection.AllowDeletions {
		case true:
			dl.Warn("'allow deletion' enabled on branch '%s'", branch)
		case false:
			dl.Info("'allow deletion' disabled on branch '%s'", branch)
			score += branchProtectionSettingScores[allowDeletions]
		}
	}

	return score, max
}

func basicAdminProtection(protection *clients.BranchProtectionRule,
	branch string, dl checker.DetailLogger) (int, int) {
	score := 0
	max := 0

	// nil typically means we do not have access to the value.
	if protection.EnforceAdmins != nil {
		// Note: we don't inrecase max possible score for non-admin viewers.
		max += branchProtectionSettingScores[enforceAdmins]
		switch *protection.EnforceAdmins {
		case true:
			dl.Info("settings apply to administrators on branch '%s'", branch)
			score += branchProtectionSettingScores[enforceAdmins]
		case false:
			dl.Warn("settings do not apply to administrators on branch '%s'", branch)
		}
	} else {
		dl.Debug("unable to retrieve whether or not settings apply to administrators on branch '%s'", branch)
	}

	return score, max
}

func nonAdminContextProtection(protection *clients.BranchProtectionRule, branch string,
	dl checker.DetailLogger) (int, int) {
	score := 0
	max := 0
	// This means there are specific checks enabled.
	// If only `Requires status check to pass before merging` is enabled
	// but no specific checks are declared, it's equivalent
	// to having no status check at all.
	max += branchProtectionSettingScores[requireStatusChecksContexts]
	switch {
	case len(protection.CheckRules.Contexts) > 0:
		dl.Info("status check found to merge onto on branch '%s'", branch)
		score += branchProtectionSettingScores[requireStatusChecksContexts]
	default:
		dl.Warn("no status checks found to merge onto branch '%s'", branch)
	}
	return score, max
}

func nonAdminReviewProtection(protection *clients.BranchProtectionRule) (int, int) {
	score := 0
	max := 0

	max += branchProtectionSettingScores[requireApprovingReviewCount]
	if protection.RequiredPullRequestReviews.RequiredApprovingReviewCount != nil &&
		*protection.RequiredPullRequestReviews.RequiredApprovingReviewCount > 0 {
		// We do not display anything here, it's done in nonAdminThoroughReviewProtection()
		score += branchProtectionSettingScores[requireApprovingReviewCount]
	}
	return score, max
}

func adminReviewProtection(protection *clients.BranchProtectionRule, branch string,
	dl checker.DetailLogger) (int, int) {
	score := 0
	max := 0

	if protection.CheckRules.UpToDateBeforeMerge != nil {
		// Note: `This setting will not take effect unless at least one status check is enabled`.
		max += branchProtectionSettingScores[requireUpToDateBeforeMerge]
		switch *protection.CheckRules.UpToDateBeforeMerge {
		case true:
			dl.Info("status checks require up-to-date branches for '%s'", branch)
			score += branchProtectionSettingScores[requireUpToDateBeforeMerge]
		default:
			dl.Warn("status checks do not require up-to-date branches for '%s'", branch)
		}
	} else {
		dl.Debug("unable to retrieve whether up-to-date branches are needed to merge on branch '%s'", branch)
	}

	return score, max
}

func adminThoroughReviewProtection(protection *clients.BranchProtectionRule, branch string,
	dl checker.DetailLogger) (int, int) {
	score := 0
	max := 0
	if protection.RequiredPullRequestReviews.DismissStaleReviews != nil {
		// Note: we don't inrecase max possible score for non-admin viewers.
		max += branchProtectionSettingScores[dismissStaleReviews]
		switch *protection.RequiredPullRequestReviews.DismissStaleReviews {
		case true:
			dl.Info("Stale review dismissal enabled on branch '%s'", branch)
			score += branchProtectionSettingScores[dismissStaleReviews]
		case false:
			dl.Warn("Stale review dismissal disabled on branch '%s'", branch)
		}
	} else {
		dl.Debug("unable to retrieve review dismissal on branch '%s'", branch)
	}
	return score, max
}

func nonAdminThoroughReviewProtection(protection *clients.BranchProtectionRule, branch string,
	dl checker.DetailLogger) (int, int) {
	score := 0
	max := 0

	max += branchProtectionSettingScores[requireApprovingReviewCount]
	if protection.RequiredPullRequestReviews.RequiredApprovingReviewCount != nil {
		switch *protection.RequiredPullRequestReviews.RequiredApprovingReviewCount >= minReviews {
		case true:
			dl.Info("number of required reviewers is %d on branch '%s'",
				*protection.RequiredPullRequestReviews.RequiredApprovingReviewCount, branch)
			score += branchProtectionSettingScores[requireApprovingReviewCount]
		default:
			dl.Warn("number of required reviewers is only %d on branch '%s'",
				*protection.RequiredPullRequestReviews.RequiredApprovingReviewCount, branch)
		}
	} else {
		dl.Warn("number of required reviewers is 0 on branch '%s'", branch)
	}
	return score, max
}
