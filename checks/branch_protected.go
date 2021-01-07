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

func init() {
	registerCheck("Branch-Protection", BranchProtection)
}

func BranchProtection(c checker.Checker) checker.CheckResult {
	repo, _, err := c.Client.Repositories.Get(c.Ctx, c.Owner, c.Repo)
	if err != nil {
		return checker.RetryResult(err)
	}

	protection, _, err := c.Client.Repositories.
		GetBranchProtection(c.Ctx, c.Owner, c.Repo, *repo.DefaultBranch)

	if err != nil {
		c.Logf("!! branch protection not enabled")
		return checker.CheckResult{
			Pass:       false,
			Confidence: 10,
		}
	}
	return IsBranchProtected(protection, c)

}

func IsBranchProtected(protection *github.Protection, c checker.Checker) checker.CheckResult {
	totalChecks := 6
	totalSuccess := 0

	if protection.GetAllowForcePushes() != nil {
		if protection.AllowForcePushes.Enabled {
			c.Logf("!! branch protection AllowForcePushes enabled")
		} else {
			totalSuccess++
		}
	}

	if protection.GetAllowDeletions() != nil {
		if protection.AllowDeletions.Enabled {
			c.Logf("!! branch protection AllowDeletions enabled")
		} else {
			totalSuccess++
		}
	}

	if protection.GetEnforceAdmins() != nil {
		if !protection.EnforceAdmins.Enabled {
			c.Logf("!! branch protection EnforceAdmins not enabled")
		} else {
			totalSuccess++
		}
	}

	if protection.GetRequireLinearHistory() != nil {
		if !protection.RequireLinearHistory.Enabled {
			c.Logf("!! branch protection require linear history not enabled")
		} else {
			totalSuccess++
		}
	}

	if protection.GetRequiredStatusChecks() != nil {
		if !protection.RequiredStatusChecks.Strict {
			c.Logf("!! branch protection require status checks to pass before merging not enabled")
		} else {
			if len(protection.RequiredStatusChecks.Contexts) == 0 {
				c.Logf("!! branch protection require status checks to pass before merging has no specific status to check for")
			} else {
				totalSuccess++
			}
		}
	}

	if protection.GetRequiredPullRequestReviews() != nil {
		if protection.RequiredPullRequestReviews.RequiredApprovingReviewCount < 1 {
			c.Logf("!! branch protection require pullrequest before merging not enabled")
		} else {
			totalSuccess++
		}
	}

	return checker.ProportionalResult(totalSuccess, totalChecks, 1.0)
}
