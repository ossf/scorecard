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
	"github.com/ossf/scorecard/checker"
)

const (
	branchProtectionStr = "Branch-Protection"
	minReviews          = 2
)

func init() {
	registerCheck(branchProtectionStr, BranchProtection)
}

func BranchProtection(c *checker.CheckRequest) checker.CheckResult {
	repo, _, err := c.Client.Repositories.Get(c.Ctx, c.Owner, c.Repo)
	if err != nil {
		return checker.MakeRetryResult(branchProtectionStr, err)
	}

	protection, resp, err := c.Client.Repositories.
		GetBranchProtection(c.Ctx, c.Owner, c.Repo, *repo.DefaultBranch)
	const fileNotFound = 404
	if resp.StatusCode == fileNotFound {
		return checker.MakeRetryResult(branchProtectionStr, err)
	}

	if err != nil {
		c.Logf("!! branch protection not enabled")
		const confidence = 10
		return checker.CheckResult{
			Name:       branchProtectionStr,
			Pass:       false,
			Confidence: confidence,
		}
	}
	return IsBranchProtected(protection, c)
}

func IsBranchProtected(protection *github.Protection, c *checker.CheckRequest) checker.CheckResult {
	totalChecks := 6
	totalSuccess := 0

	// This is disabled by default (good).
	if protection.GetAllowForcePushes() != nil &&
		protection.AllowForcePushes.Enabled {
		c.Logf("!! branch protection - AllowForcePushes should be disabled")
	} else {
		totalSuccess++
	}

	// This is disabled by default (good).
	if protection.GetAllowDeletions() != nil &&
		protection.AllowDeletions.Enabled {
		c.Logf("!! branch protection - AllowDeletions should be disabled")
	} else {
		totalSuccess++
	}

	// This is disabled by default (bad).
	if protection.GetEnforceAdmins() != nil &&
		protection.EnforceAdmins.Enabled {
		totalSuccess++
	} else {
		c.Logf("!! branch protection - EnforceAdmins should be enabled")
	}

	// This is disabled by default (bad).
	if protection.GetRequireLinearHistory() != nil &&
		protection.RequireLinearHistory.Enabled {
		totalSuccess++
	} else {
		c.Logf("!! branch protection - Linear history should be enabled")
	}

	// This is disabled by default (bad).
	if protection.GetRequiredStatusChecks() != nil &&
		protection.RequiredStatusChecks.Strict &&
		len(protection.RequiredStatusChecks.Contexts) > 0 {
		totalSuccess++
	} else {
		switch {
		case protection.RequiredStatusChecks == nil ||
			!protection.RequiredStatusChecks.Strict:
			c.Logf("!! branch protection - Status checks for merging should be enabled")
		case len(protection.RequiredStatusChecks.Contexts) == 0:
			c.Logf("!! branch protection - Status checks for merging should have specific status to check for")
		default:
			panic("!! branch protection - Unhandled status checks error")
		}
	}

	// This is disabled by default (bad).
	if protection.GetRequiredPullRequestReviews() != nil &&
		protection.RequiredPullRequestReviews.RequiredApprovingReviewCount >= minReviews &&
		protection.RequiredPullRequestReviews.DismissStaleReviews &&
		protection.RequiredPullRequestReviews.RequireCodeOwnerReviews {
		totalSuccess++
	} else {
		switch {
		case protection.RequiredPullRequestReviews == nil:
			c.Logf("!! branch protection - Pullrequest reviews should be enabled")
			fallthrough
		case protection.RequiredPullRequestReviews.RequiredApprovingReviewCount < minReviews:
			c.Logf("!! branch protection - %v pullrequest reviews should be enabled", minReviews)
			fallthrough
		case !protection.RequiredPullRequestReviews.DismissStaleReviews:
			c.Logf("!! branch protection - Stale review dismissal should be enabled")
			fallthrough
		case !protection.RequiredPullRequestReviews.RequireCodeOwnerReviews:
			c.Logf("!! branch protection - Owner review should be enabled")
		default:
			panic("!! branch protection - Unhandled pull request error")
		}
	}

	return checker.MakeProportionalResult(branchProtectionStr, totalSuccess, totalChecks, 1.0)
}
