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
	"math"
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
	// First level.
	allowForcePushes branchProtectionSetting = iota
	allowDeletions
	requiresPullRequests
	// Second level.
	requireStatusChecksContexts
	// Third level.
	requireApprovingReviewCount
	// Admin settings.
	// First level.
	enforceAdmins
	// Second level.
	requireUpToDateBeforeMerge
	// Third level.
	dismissStaleReviews
	// requireCodeOwnerReviews no longer used.
	// requireLinearHistory no longer used, see https://github.com/ossf/scorecard/issues/1027.
)

var branchProtectionSettingScores = map[branchProtectionSetting]int{
	// The basics.
	// We would like to have enforceAdmins,
	// but it's only available to admins.
	allowForcePushes:            1,
	allowDeletions:              1,
	requiresPullRequests:        1,
	requireStatusChecksContexts: 1, // Gated by requireStatusChecks=true.

	requireApprovingReviewCount: 2,

	// Need admin token, so we cannot rely on them in general.
	requireUpToDateBeforeMerge: 1, // Gated by requireStatusChecks=true.
	enforceAdmins:              1,
	dismissStaleReviews:        1,
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

	var scores []int
	// Check protections on all the branches.
	for b := range checkBranches {
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
			return checker.CreateMinScoreResult(CheckBranchProtection,
				fmt.Sprintf("branch protection not enabled on development/release branch: %s", b))
		}

		// The branch is protected. Check the protection.
		// naScore, naMax, adScore, adMax, naRev, naRevMax, adRev, adRevMax :=
		readBranchProtection(&branch.BranchProtectionRule, b, dl)

		// scores = append(scores, score)
	}

	if len(scores) == 0 {
		return checker.CreateInconclusiveResult(CheckBranchProtection, "unable to detect any development/release branches")
	}

	score := checker.AggregateScores(scores...)
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

	max += branchProtectionSettingScores[requiresPullRequests]
	switch protection.RequiredPullRequestReviews.RequiredApprovingReviewCount {
	case nil:
		dl.Warn("pull requests disabled on branch '%s'", branch)
	default:
		dl.Info("pull requests enabled on branch '%s'", branch)
		score += branchProtectionSettingScores[requiresPullRequests]
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

func nonAdminReviewProtection(protection *clients.BranchProtectionRule, branch string,
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
		dl.Info("no status check found to merge onto on branch '%s'", branch)
		score += branchProtectionSettingScores[requireStatusChecksContexts]
	default:
		dl.Warn("status checks found to merge onto branch '%s'", branch)
	}

	max += branchProtectionSettingScores[requireApprovingReviewCount]
	if protection.RequiredPullRequestReviews.RequiredApprovingReviewCount != nil {
		// We do not display anything here, it's done in nonAdminThoroughReviewProtection()
		score += int(math.Min(float64(*protection.RequiredPullRequestReviews.RequiredApprovingReviewCount), minReviews))
	}
	return score, max
}

func adminReviewProtection(protection *clients.BranchProtectionRule, branch string,
	dl checker.DetailLogger) (int, int) {
	score := 0
	max := 0

	if protection.CheckRules.UpToDateBeforeMerge != nil {
		// Note: `This setting will not take effect unless at least one status check is enabled`.
		// Even though it technically is not enforced by GitHub if Context=nil, we still
		// show a positive outcome for users. Otherwise it will be confusing when they compare
		// their settings to screcard results.
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
			// RequiredApprovingReviewCount is 0 or 1.
			score += int(*protection.RequiredPullRequestReviews.RequiredApprovingReviewCount)
		}
	} else {
		// This happens when pull requests are disabled entirely.
		dl.Warn("number of required reviewers is only %d on branch '%s'", 0, branch)
	}
	return score, max
}

// readBranchProtection reads branch protection rules on a Git branch.
func readBranchProtection(protection *clients.BranchProtectionRule,
	branch string, dl checker.DetailLogger) (int, int, int, int, int, int, int, int, int, int, int, int) {

	naScore, naMax := basicNonAdminProtection(protection, branch, dl)
	adScore, adMax := basicAdminProtection(protection, branch, dl)

	naRev, naRevMax := nonAdminReviewProtection(protection, branch, dl)
	adRev, adRevMax := adminReviewProtection(protection, branch, dl)

	a, b := nonAdminThoroughReviewProtection(protection, branch, dl)
	c, d := adminThoroughReviewProtection(protection, branch, dl) // Do we want this?
	return naScore, naMax, adScore, adMax, naRev, naRevMax, adRev, adRevMax, a, b, c, d
}
