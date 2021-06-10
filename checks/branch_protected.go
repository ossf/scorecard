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

	"github.com/google/go-github/v32/github"

	"github.com/ossf/scorecard/checker"
)

const (
	// CheckBranchProtection is the registered name for BranchProtection.
	CheckBranchProtection = "Branch-Protection"
	minReviews            = 1
)

//nolint:gochecknoinits
func init() {
	registerCheck(CheckBranchProtection, BranchProtection)
}

type repositories interface {
	Get(context.Context, string, string) (*github.Repository,
		*github.Response, error)
	ListReleases(ctx context.Context, owner string, repo string, opts *github.ListOptions) (
		[]*github.RepositoryRelease, *github.Response, error)
	GetBranchProtection(context.Context, string, string, string) (
		*github.Protection, *github.Response, error)
}

type logger func(s string, f ...interface{})

func BranchProtection(c *checker.CheckRequest) checker.CheckResult {
	// Checks branch protection on both release and development branch
	return checkReleaseAndDevBranchProtection(c.Ctx, c.Client.Repositories, c.Logf, c.Owner, c.Repo)
}

func checkReleaseAndDevBranchProtection(ctx context.Context, r repositories, l logger, ownerStr,
	repoStr string) checker.CheckResult {
	// Get release branches
	releases, _, err := r.ListReleases(ctx, ownerStr, repoStr, &github.ListOptions{})
	if err != nil {
		return checker.MakeRetryResult(CheckBranchProtection, err)
	}

	var checks []checker.CheckResult
	for _, release := range releases {
		if release.TargetCommitish != nil {
			res := getProtectionAndCheck(ctx, r, l, ownerStr, repoStr, *release.TargetCommitish)
			checks = append(checks, res)
		}
	}

	// Default development branch check
	repo, _, err := r.Get(ctx, ownerStr, repoStr)
	if err != nil {
		return checker.MakeRetryResult(CheckBranchProtection, err)
	}
	res := getProtectionAndCheck(ctx, r, l, ownerStr, repoStr, *repo.DefaultBranch)
	checks = append(checks, res)

	return checker.MultiCheckResultAnd(checks...)
}

func getProtectionAndCheck(ctx context.Context, r repositories, l logger, ownerStr, repoStr,
	branch string) checker.CheckResult {
	protection, resp, err := r.GetBranchProtection(ctx, ownerStr, repoStr, branch)

	const fileNotFound = 404
	if resp.StatusCode == fileNotFound {
		return checker.MakeRetryResult(CheckBranchProtection, err)
	}

	if err != nil {
		l("!! branch protection not enabled for branch %s", branch)
		const confidence = 10

		return checker.CheckResult{
			Name:       CheckBranchProtection,
			Pass:       false,
			Confidence: confidence,
		}
	}

	return IsBranchProtected(protection, branch, l)
}

func IsBranchProtected(protection *github.Protection, branch string, l logger) checker.CheckResult {
	totalChecks := 6
	totalSuccess := 0

	// This is disabled by default (good).
	if protection.GetAllowForcePushes() != nil &&
		protection.AllowForcePushes.Enabled {
		l("!! branch protection - AllowForcePushes should be disabled on %s", branch)
	} else {
		totalSuccess++
	}

	// This is disabled by default (good).
	if protection.GetAllowDeletions() != nil &&
		protection.AllowDeletions.Enabled {
		l("!! branch protection - AllowDeletions should be disabled on %s", branch)
	} else {
		totalSuccess++
	}

	// This is disabled by default (bad).
	if protection.GetEnforceAdmins() != nil &&
		protection.EnforceAdmins.Enabled {
		totalSuccess++
	} else {
		l("!! branch protection - EnforceAdmins should be enabled on %s", branch)
	}

	// This is disabled by default (bad).
	if protection.GetRequireLinearHistory() != nil &&
		protection.RequireLinearHistory.Enabled {
		totalSuccess++
	} else {
		l("!! branch protection - Linear history should be enabled on %s", branch)
	}

	if requiresStatusChecks(protection, branch, l) {
		totalSuccess++
	}

	if requiresThoroughReviews(protection, branch, l) {
		totalSuccess++
	}

	return checker.MakeProportionalResult(CheckBranchProtection, totalSuccess, totalChecks, 1.0)
}

// Returns true if several PR status checks requirements are enabled. Otherwise returns false and logs why it failed.
func requiresStatusChecks(protection *github.Protection, branch string, l logger) bool {
	// This is disabled by default (bad).
	if protection.GetRequiredStatusChecks() != nil &&
		protection.RequiredStatusChecks.Strict &&
		len(protection.RequiredStatusChecks.Contexts) > 0 {
		return true
	}
	switch {
	case protection.RequiredStatusChecks == nil ||
		!protection.RequiredStatusChecks.Strict:
		l("!! branch protection - Status checks for merging should be enabled on %s", branch)
	case len(protection.RequiredStatusChecks.Contexts) == 0:
		l("!! branch protection - Status checks for merging should have specific status to check for on %s", branch)
	default:
		panic("!! branch protection - Unhandled status checks error")
	}
	return false
}

// Returns true if several PR review requirements are enabled. Otherwise returns false and logs why it failed.
func requiresThoroughReviews(protection *github.Protection, branch string, l logger) bool {
	// This is disabled by default (bad).
	if protection.GetRequiredPullRequestReviews() != nil &&
		protection.RequiredPullRequestReviews.RequiredApprovingReviewCount >= minReviews &&
		protection.RequiredPullRequestReviews.DismissStaleReviews &&
		protection.RequiredPullRequestReviews.RequireCodeOwnerReviews {
		return true
	}
	switch {
	case protection.RequiredPullRequestReviews == nil:
		l("!! branch protection - Pullrequest reviews should be enabled on %s", branch)
		fallthrough
	case protection.RequiredPullRequestReviews.RequiredApprovingReviewCount < minReviews:
		l("!! branch protection - %v pullrequest reviews should be enabled on %s", minReviews, branch)
		fallthrough
	case !protection.RequiredPullRequestReviews.DismissStaleReviews:
		l("!! branch protection - Stale review dismissal should be enabled on %s", branch)
		fallthrough
	case !protection.RequiredPullRequestReviews.RequireCodeOwnerReviews:
		l("!! branch protection - Owner review should be enabled on %s", branch)
	default:
		panic("!! branch protection - Unhandled pull request error")
	}
	return false
}
