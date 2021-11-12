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
	CheckBranchProtection                         = "Branch-Protection"
	minReviews                                    = 2
	allowForcePushes      branchProtectionSetting = iota
	allowDeletions
	requireLinearHistory
	enforceAdmins
	requireStrictStatusChecks
	requireStatusChecksContexts
	requireApprovingReviewCount
	dismissStaleReviews
	requireCodeOwnerReviews
)

var branchProtectionSettingScores = map[branchProtectionSetting]int{
	allowForcePushes:            1,
	allowDeletions:              1,
	requireLinearHistory:        1,
	enforceAdmins:               3,
	requireStrictStatusChecks:   1,
	requireStatusChecksContexts: 1,
	requireApprovingReviewCount: 2,
	// This is a big deal to enabled, so let's reward 3 points.
	dismissStaleReviews:     3,
	requireCodeOwnerReviews: 2,
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
		if branch.Name != nil && *branch.Name != "" {
			ret[*branch.Name] = branch
		}
	}
	return ret
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
	checkBranches[*defaultBranch.Name] = true

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
		if branch.Protected != nil && !*branch.Protected {
			dl.Warn("branch protection not enabled for branch '%s'", b)
			return checker.CreateMinScoreResult(CheckBranchProtection,
				fmt.Sprintf("branch protection not enabled on development/release branch: %s", b))
		}
		// The branch is protected. Check the protection.
		score := isBranchProtected(&branch.BranchProtectionRule, b, dl)
		scores = append(scores, score)
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

// isBranchProtected checks branch protection rules on a Git branch.
func isBranchProtected(protection *clients.BranchProtectionRule, branch string, dl checker.DetailLogger) int {
	totalScore := 15
	score := 0

	if protection.AllowForcePushes != nil {
		switch *protection.AllowForcePushes {
		case true:
			dl.Warn("'force pushes' enabled on branch '%s'", branch)
		case false:
			dl.Info("'force pushes' disabled on branch '%s'", branch)
			score += branchProtectionSettingScores[allowForcePushes]
		}
	}

	if protection.AllowDeletions != nil {
		switch *protection.AllowDeletions {
		case true:
			dl.Warn("'allow deletion' enabled on branch '%s'", branch)
		case false:
			dl.Info("'allow deletion' disabled on branch '%s'", branch)
			score += branchProtectionSettingScores[allowDeletions]
		}
	}

	if protection.RequireLinearHistory != nil {
		switch *protection.RequireLinearHistory {
		case true:
			dl.Info("linear history enabled on branch '%s'", branch)
			score += branchProtectionSettingScores[requireLinearHistory]
		case false:
			dl.Warn("linear history disabled on branch '%s'", branch)
		}
	}

	if protection.EnforceAdmins != nil {
		switch *protection.EnforceAdmins {
		case true:
			dl.Info("'administrator' PRs need reviews before being merged on branch '%s'", branch)
			score += branchProtectionSettingScores[enforceAdmins]
		case false:
			dl.Warn("'administrator' PRs are exempt from reviews on branch '%s'", branch)
		}
	}

	score += requiresStatusChecks(protection, branch, dl)
	score += requiresThoroughReviews(protection, branch, dl)

	return checker.CreateProportionalScore(score, totalScore)
}

// Returns true if several PR status checks requirements are enabled. Otherwise returns false and logs why it failed.
// Maximum score returned is 2.
func requiresStatusChecks(protection *clients.BranchProtectionRule, branch string, dl checker.DetailLogger) int {
	score := 0

	if protection.RequiredStatusChecks.Strict != nil {
		switch *protection.RequiredStatusChecks.Strict {
		case false:
			dl.Warn("status checks for merging disabled on branch '%s'", branch)
			return score
		case true:
			dl.Info("strict status check enabled on branch '%s'", branch)
			score += branchProtectionSettingScores[requireStrictStatusChecks]
		}
	}

	if len(protection.RequiredStatusChecks.Contexts) > 0 {
		dl.Info("status checks for merging have specific status to check on branch '%s'", branch)
		score += branchProtectionSettingScores[requireStatusChecksContexts]
	} else {
		dl.Warn("status checks for merging have no specific status to check on branch '%s'", branch)
	}

	return score
}

// Returns true if several PR review requirements are enabled. Otherwise returns false and logs why it failed.
// Maximum score returned is 7.
func requiresThoroughReviews(protection *clients.BranchProtectionRule, branch string, dl checker.DetailLogger) int {
	score := 0

	if protection.RequiredPullRequestReviews.RequiredApprovingReviewCount != nil {
		if *protection.RequiredPullRequestReviews.RequiredApprovingReviewCount >= minReviews {
			dl.Info("number of required reviewers is %d on branch '%s'",
				*protection.RequiredPullRequestReviews.RequiredApprovingReviewCount, branch)
			score += branchProtectionSettingScores[requireApprovingReviewCount]
		} else {
			score += int(*protection.RequiredPullRequestReviews.RequiredApprovingReviewCount)
			dl.Warn("number of required reviewers is only %d on branch '%s'",
				*protection.RequiredPullRequestReviews.RequiredApprovingReviewCount, branch)
		}
	}

	if protection.RequiredPullRequestReviews.DismissStaleReviews != nil {
		switch *protection.RequiredPullRequestReviews.DismissStaleReviews {
		case true:
			dl.Info("Stale review dismissal enabled on branch '%s'", branch)
			score += branchProtectionSettingScores[dismissStaleReviews]
		case false:
			dl.Warn("Stale review dismissal disabled on branch '%s'", branch)
		}
	}

	if protection.RequiredPullRequestReviews.RequireCodeOwnerReviews != nil {
		switch *protection.RequiredPullRequestReviews.RequireCodeOwnerReviews {
		case true:
			score += branchProtectionSettingScores[requireCodeOwnerReviews]
			dl.Info("Owner review required on branch '%s'", branch)
		case false:
			dl.Warn("Owner review not required on branch '%s'", branch)
		}
	}

	return score
}
