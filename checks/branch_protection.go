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
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/google/go-github/v38/github"

	"github.com/ossf/scorecard/v2/checker"
	sce "github.com/ossf/scorecard/v2/errors"
)

const (
	// CheckBranchProtection is the exported name for Branch-Protected check.
	CheckBranchProtection = "Branch-Protection"
	minReviews            = 2
)

//nolint:gochecknoinits
func init() {
	registerCheck(CheckBranchProtection, BranchProtection)
}

// TODO: Use RepoClient interface instead of this.
type repositories interface {
	Get(context.Context, string, string) (*github.Repository,
		*github.Response, error)
	ListBranches(ctx context.Context, owner string, repo string,
		opts *github.BranchListOptions) ([]*github.Branch, *github.Response, error)
	ListReleases(ctx context.Context, owner string, repo string, opts *github.ListOptions) (
		[]*github.RepositoryRelease, *github.Response, error)
	GetBranchProtection(context.Context, string, string, string) (
		*github.Protection, *github.Response, error)
}

type branchMap map[string]*github.Branch

func (b branchMap) getBranchByName(name string) (*github.Branch, error) {
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

func getBranchMapFrom(branches []*github.Branch) branchMap {
	ret := make(branchMap)
	for _, branch := range branches {
		ret[branch.GetName()] = branch
	}
	return ret
}

// BranchProtection runs Branch-Protection check.
func BranchProtection(c *checker.CheckRequest) checker.CheckResult {
	// Checks branch protection on both release and development branch.
	return checkReleaseAndDevBranchProtection(c.Ctx, c.Client.Repositories, c.Dlogger, c.Owner, c.Repo)
}

func checkReleaseAndDevBranchProtection(ctx context.Context, r repositories, dl checker.DetailLogger, ownerStr,
	repoStr string) checker.CheckResult {
	// Get all branches. This will include information on whether they are protected.
	branches, _, err := r.ListBranches(ctx, ownerStr, repoStr, &github.BranchListOptions{})
	if err != nil {
		return checker.CreateRuntimeErrorResult(CheckBranchProtection, err)
	}
	branchesMap := getBranchMapFrom(branches)

	// Get release branches.
	releases, _, err := r.ListReleases(ctx, ownerStr, repoStr, &github.ListOptions{})
	if err != nil {
		return checker.CreateRuntimeErrorResult(CheckBranchProtection, err)
	}

	var scores []int
	commit := regexp.MustCompile("^[a-f0-9]{40}$")
	checkBranches := make(map[string]bool)
	for _, release := range releases {
		if release.TargetCommitish == nil {
			// Log with a named error if target_commitish is nil.
			e := sce.Create(sce.ErrScorecardInternal, errInternalCommitishNil.Error())
			return checker.CreateRuntimeErrorResult(CheckBranchProtection, e)
		}

		// TODO: if this is a sha, get the associated branch. for now, ignore.
		if commit.Match([]byte(*release.TargetCommitish)) {
			continue
		}

		// Try to resolve the branch name.
		b, err := branchesMap.getBranchByName(release.GetTargetCommitish())
		if err != nil {
			// If the commitish branch is still not found, fail.
			return checker.CreateRuntimeErrorResult(CheckBranchProtection, err)
		}

		// Branch is valid, add to list of branches to check.
		checkBranches[b.GetName()] = true
	}

	// Add default branch.
	repo, _, err := r.Get(ctx, ownerStr, repoStr)
	if err != nil {
		return checker.CreateRuntimeErrorResult(CheckBranchProtection, err)
	}
	checkBranches[repo.GetDefaultBranch()] = true

	protected := true
	unknown := false
	// Check protections on all the branches.
	for b := range checkBranches {
		branch, err := branchesMap.getBranchByName(b)
		if err != nil {
			if errors.Is(err, errInternalBranchNotFound) {
				continue
			}
			return checker.CreateRuntimeErrorResult(CheckBranchProtection, err)
		}
		// nolint
		if !branch.GetProtected() {
			protected = false
			dl.Warn("branch protection not enabled for branch '%s'", b)
		} else {
			// The branch is protected. Check the protection.
			score, err := getProtectionAndCheck(ctx, r, dl, ownerStr, repoStr, b)
			if err != nil {
				if errors.Is(err, errInternalBranchNotFound) {
					// Without an admin token, you only get information on the protection boolean.
					// Add a score of 1 (minimal branch protection) for this protected branch.
					unknown = true
					scores = append(scores, 1)
					dl.Warn("no detailed settings available for branch protection '%s'", b)
					continue
				} else {
					// Github timeout or other error
					return checker.CreateRuntimeErrorResult(CheckBranchProtection, err)
				}
			}
			scores = append(scores, score)
		}
	}

	if !protected {
		return checker.CreateMinScoreResult(CheckBranchProtection,
			"branch protection not enabled on development/release branches")
	}

	score := checker.AggregateScores(scores...)
	if score == checker.MinResultScore {
		return checker.CreateMinScoreResult(CheckBranchProtection,
			"branch protection not enabled on development/release branches")
	}

	if score == checker.MaxResultScore {
		return checker.CreateMaxScoreResult(CheckBranchProtection,
			"branch protection is fully enabled on development and all release branches")
	}

	if unknown {
		return checker.CreateResultWithScore(CheckBranchProtection,
			"branch protection is enabled on development and all release branches but settings are unknown", score)
	}

	return checker.CreateResultWithScore(CheckBranchProtection,
		"branch protection is not maximal on development and all release branches", score)
}

func getProtectionAndCheck(ctx context.Context, r repositories, dl checker.DetailLogger, ownerStr, repoStr,
	branch string) (int, error) {
	// We only call this if the branch is protected.
	protection, resp, err := r.GetBranchProtection(ctx, ownerStr, repoStr, branch)
	if err != nil {
		// Check the type of error. A not found error indicates that permissions are denied.
		if resp.StatusCode == http.StatusNotFound {
			//nolint
			return 1, sce.Create(errInternalBranchNotFound, errInternalBranchNotFound.Error())
		}
		//nolint
		return checker.InconclusiveResultScore, sce.Create(sce.ErrScorecardInternal, err.Error())
	}

	return IsBranchProtected(protection, branch, dl), nil
}

// IsBranchProtected checks branch protection rules on a Git branch.
func IsBranchProtected(protection *github.Protection, branch string, dl checker.DetailLogger) int {
	totalScore := 15
	score := 0

	if protection.GetAllowForcePushes() != nil &&
		protection.AllowForcePushes.Enabled {
		dl.Warn("'force pushes' enabled on branch '%s'", branch)
	} else {
		dl.Info("'force pushes' disabled on branch '%s'", branch)
		score++
	}

	if protection.GetAllowDeletions() != nil &&
		protection.AllowDeletions.Enabled {
		dl.Warn("'allow deletion' enabled on branch '%s'", branch)
	} else {
		dl.Info("'allow deletion' disabled on branch '%s'", branch)
		score++
	}

	if protection.GetRequireLinearHistory() != nil &&
		protection.RequireLinearHistory.Enabled {
		dl.Info("linear history enabled on branch '%s'", branch)
		score++
	} else {
		dl.Warn("linear history disabled on branch '%s'", branch)
	}

	score += requiresStatusChecks(protection, branch, dl)

	score += requiresThoroughReviews(protection, branch, dl)

	if protection.GetEnforceAdmins() != nil &&
		protection.EnforceAdmins.Enabled {
		dl.Info("'admininistrator' PRs need reviews before being merged on branch '%s'", branch)
		score += 3
	} else {
		dl.Warn("'admininistrator' PRs are exempt from reviews on branch '%s'", branch)
	}

	if score == totalScore {
		return checker.MaxResultScore
	}

	return checker.CreateProportionalScore(score, totalScore)
}

// Returns true if several PR status checks requirements are enabled. Otherwise returns false and logs why it failed.
// Maximum score returned is 2.
func requiresStatusChecks(protection *github.Protection, branch string, dl checker.DetailLogger) int {
	score := 0

	if protection.GetRequiredStatusChecks() == nil ||
		!protection.RequiredStatusChecks.Strict {
		dl.Warn("status checks for merging disabled on branch '%s'", branch)
		return score
	}

	dl.Info("strict status check enabled on branch '%s'", branch)
	score++

	if len(protection.RequiredStatusChecks.Contexts) > 0 {
		dl.Warn("status checks for merging have specific status to check on branch '%s'", branch)
		score++
	} else {
		dl.Warn("status checks for merging have no specific status to check on branch '%s'", branch)
	}

	return score
}

// Returns true if several PR review requirements are enabled. Otherwise returns false and logs why it failed.
// Maximum score returned is 7.
func requiresThoroughReviews(protection *github.Protection, branch string, dl checker.DetailLogger) int {
	score := 0

	if protection.GetRequiredPullRequestReviews() == nil {
		dl.Warn("pull request reviews disabled on branch '%s'", branch)
		return score
	}

	if protection.RequiredPullRequestReviews.RequiredApprovingReviewCount >= minReviews {
		dl.Info("number of required reviewers is %d on branch '%s'",
			protection.RequiredPullRequestReviews.RequiredApprovingReviewCount, branch)
		score += 2
	} else {
		score += protection.RequiredPullRequestReviews.RequiredApprovingReviewCount
		dl.Warn("number of required reviewers is only %d on branch '%s'",
			protection.RequiredPullRequestReviews.RequiredApprovingReviewCount, branch)
	}

	if protection.RequiredPullRequestReviews.DismissStaleReviews {
		// This is a big deal to enabled, so let's reward 3 points.
		dl.Info("Stale review dismissal enabled on branch '%s'", branch)
		score += 3
	} else {
		dl.Warn("Stale review dismissal disabled on branch '%s'", branch)
	}

	if protection.RequiredPullRequestReviews.RequireCodeOwnerReviews {
		score += 2
		dl.Info("Owner review required on branch '%s'", branch)
	} else {
		dl.Warn("Owner review not required on branch '%s'", branch)
	}

	return score
}
