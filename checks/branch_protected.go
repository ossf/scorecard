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
	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"

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

func BranchProtection(c *checker.CheckRequest) checker.CheckResult {
	// Checks branch protection on both release and development branch
	repo, _, err := c.Client.Repositories.Get(c.Ctx, c.Owner, c.Repo)
	if err != nil {
		return checker.MakeRetryResult(CheckBranchProtection, err)
	}

	// Get release branches
	releases, _, err := c.Client.Repositories.ListReleases(c.Ctx, c.Owner, c.Repo, &github.ListOptions{})
	if err != nil {
		return checker.MakeRetryResult(CheckBranchProtection, err)
	}

	var checks []checker.CheckFn
	for _, release := range releases {
		if release.TargetCommitish != nil {
			res, err := getReleaseBranchProtection(c, *release.TargetCommitish)
			if err != nil {
				return checker.MakeRetryResult(CheckBranchProtection, err)
			}
			checks = append(checks, res)
		}
	}

	// Default development branch check
	res, err := getReleaseBranchProtection(c, *repo.DefaultBranch)
	if err != nil {
		return checker.MakeRetryResult(CheckBranchProtection, err)
	}
	checks = append(checks, res)

	return checker.MultiCheckAnd(checks...)(c)
}

func IsBranchProtected(protection *github.Protection, branch string, c *checker.CheckRequest) checker.CheckResult {
	totalChecks := 6
	totalSuccess := 0

	// This is disabled by default (good).
	if protection.GetAllowForcePushes() != nil &&
		protection.AllowForcePushes.Enabled {
		c.Logf("!! branch protection - AllowForcePushes should be disabled on %s", branch)
	} else {
		totalSuccess++
	}

	// This is disabled by default (good).
	if protection.GetAllowDeletions() != nil &&
		protection.AllowDeletions.Enabled {
		c.Logf("!! branch protection - AllowDeletions should be disabled on %s", branch)
	} else {
		totalSuccess++
	}

	// This is disabled by default (bad).
	if protection.GetEnforceAdmins() != nil &&
		protection.EnforceAdmins.Enabled {
		totalSuccess++
	} else {
		c.Logf("!! branch protection - EnforceAdmins should be enabled on %s", branch)
	}

	// This is disabled by default (bad).
	if protection.GetRequireLinearHistory() != nil &&
		protection.RequireLinearHistory.Enabled {
		totalSuccess++
	} else {
		c.Logf("!! branch protection - Linear history should be enabled on %s", branch)
	}

	if requiresStatusChecks(protection, branch, c) {
		totalSuccess++
	}

	if requiresThoroughReviews(protection, branch, c) {
		totalSuccess++
	}

	return checker.MakeProportionalResult(CheckBranchProtection, totalSuccess, totalChecks, 1.0)
}

// Returns true if several PR status checks requirements are enabled. Otherwise returns false and logs why it failed.
func requiresStatusChecks(protection *github.Protection, branch string, c *checker.CheckRequest) bool {
	// This is disabled by default (bad).
	if protection.GetRequiredStatusChecks() != nil &&
		protection.RequiredStatusChecks.Strict &&
		len(protection.RequiredStatusChecks.Contexts) > 0 {
		return true
	}
	switch {
	case protection.RequiredStatusChecks == nil ||
		!protection.RequiredStatusChecks.Strict:
		c.Logf("!! branch protection - Status checks for merging should be enabled on %s", branch)
	case len(protection.RequiredStatusChecks.Contexts) == 0:
		c.Logf("!! branch protection - Status checks for merging should have specific status to check for on %s", branch)
	default:
		panic("!! branch protection - Unhandled status checks error")
	}
	return false
}

// Returns true if several PR review requirements are enabled. Otherwise returns false and logs why it failed.
func requiresThoroughReviews(protection *github.Protection, branch string, c *checker.CheckRequest) bool {
	// This is disabled by default (bad).
	if protection.GetRequiredPullRequestReviews() != nil &&
		protection.RequiredPullRequestReviews.RequiredApprovingReviewCount >= minReviews &&
		protection.RequiredPullRequestReviews.DismissStaleReviews &&
		protection.RequiredPullRequestReviews.RequireCodeOwnerReviews {
		return true
	}
	switch {
	case protection.RequiredPullRequestReviews == nil:
		c.Logf("!! branch protection - Pullrequest reviews should be enabled on %s", branch)
		fallthrough
	case protection.RequiredPullRequestReviews.RequiredApprovingReviewCount < minReviews:
		c.Logf("!! branch protection - %v pullrequest reviews should be enabled on %s", minReviews, branch)
		fallthrough
	case !protection.RequiredPullRequestReviews.DismissStaleReviews:
		c.Logf("!! branch protection - Stale review dismissal should be enabled on %s", branch)
		fallthrough
	case !protection.RequiredPullRequestReviews.RequireCodeOwnerReviews:
		c.Logf("!! branch protection - Owner review should be enabled on %s", branch)
	default:
		panic("!! branch protection - Unhandled pull request error")
	}
	return false
}

func getReleaseBranchProtection(c *checker.CheckRequest, branch string) (checker.CheckFn, error) {
	protection, resp, err := c.Client.Repositories.GetBranchProtection(c.Ctx, c.Owner, c.Repo, branch)

	const fileNotFound = 404
	if resp.StatusCode == fileNotFound {
		return nil, errors.Wrap(err, "not found")
	}

	if err != nil {
		c.Logf("!! branch protection not enabled for branch %s", branch)
		const confidence = 10
		return func(*checker.CheckRequest) checker.CheckResult {
			return checker.CheckResult{
				Name:       CheckBranchProtection,
				Pass:       false,
				Confidence: confidence,
			}
		}, nil
	}

	return func(*checker.CheckRequest) checker.CheckResult { return IsBranchProtected(protection, branch, c) }, nil
}
