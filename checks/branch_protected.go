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
	"fmt"
	"regexp"

	"github.com/google/go-github/v32/github"

	"github.com/ossf/scorecard/checker"
	sce "github.com/ossf/scorecard/errors"
)

const (
	CheckBranchProtection = "Branch-Protection"
	minReviews            = 2
)

//nolint:gochecknoinits
func init() {
	registerCheck(CheckBranchProtection, BranchProtection)
}

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

	// Get release branches
	releases, _, err := r.ListReleases(ctx, ownerStr, repoStr, &github.ListOptions{})
	if err != nil {
		return checker.CreateRuntimeErrorResult(CheckBranchProtection, err)
	}

	var checks []checker.CheckResult
	commit := regexp.MustCompile("^[a-f0-9]{40}$")
	checkBranches := make(map[string]bool)
	for _, release := range releases {
		if release.TargetCommitish == nil {
			// Log with a named error if target_commitish is nil.
			r := checker.CreateRuntimeErrorResult(CheckBranchProtection,
				sce.Create(sce.ErrRunFailure, sce.ErrCommitishNil.Error()))
			checks = append(checks, r)
			continue
		}

		// TODO: if this is a sha, get the associated branch. for now, ignore.
		if commit.Match([]byte(*release.TargetCommitish)) {
			continue
		}

		// Try to resolve the branch name.
		name, err := resolveBranchName(branches, *release.TargetCommitish)
		if err != nil {
			// If the commitish branch is still not found, fail.
			r := checker.CreateRuntimeErrorResult(CheckBranchProtection,
				sce.Create(sce.ErrRunFailure, sce.ErrBranchNotFound.Error()))
			checks = append(checks, r)
			continue
		}

		// Branch is valid, add to list of branches to check.
		checkBranches[*name] = true
	}

	// Add default branch.
	repo, _, err := r.Get(ctx, ownerStr, repoStr)
	if err != nil {
		return checker.CreateRuntimeErrorResult(CheckBranchProtection, err)
	}
	checkBranches[*repo.DefaultBranch] = true

	// Check protections on the branches.
	for b := range checkBranches {
		protected, err := isBranchProtected(branches, b)
		if err != nil {
			r := checker.CreateRuntimeErrorResult(CheckBranchProtection, sce.Create(sce.ErrRunFailure, sce.ErrBranchNotFound.Error()))
			checks = append(checks, r)
		}
		if !protected {
			r := checker.CreateMinScoreResult(CheckBranchProtection, fmt.Sprintf("branch protection not enabled for branch '%s'", b))
			checks = append(checks, r)
		} else {
			// The branch is protected. Check the protection.
			res := getProtectionAndCheck(ctx, r, dl, ownerStr, repoStr, b)
			checks = append(checks, res)
		}
	}

	return checker.MakeAndResult2(checks...)
}

func resolveBranchName(branches []*github.Branch, name string) (*string, error) {
	// First check list of branches.
	for _, b := range branches {
		if b.GetName() == name {
			return b.Name, nil
		}
	}
	// Ideally, we should check using repositories.GetBranch if there was a branch redirect.
	// See https://github.com/google/go-github/issues/1895
	// For now, handle the common master -> main redirect.
	if name == "master" {
		return resolveBranchName(branches, "main")
	}

	return nil, sce.Create(sce.ErrRunFailure, sce.ErrBranchNotFound.Error())
}

func isBranchProtected(branches []*github.Branch, name string) (bool, error) {
	// Returns bool indicating if protected.
	for _, b := range branches {
		if b.GetName() == name {
			return b.GetProtected(), nil
		}
	}

	return false, sce.Create(sce.ErrRunFailure, sce.ErrBranchNotFound.Error())
}

func getProtectionAndCheck(ctx context.Context, r repositories, dl checker.DetailLogger, ownerStr, repoStr,
	branch string) checker.CheckResult {
	// We only call this if the branch is protected. An error indicates not found.
	protection, resp, err := r.GetBranchProtection(ctx, ownerStr, repoStr, branch)

	const fileNotFound = 404
	if resp.StatusCode == fileNotFound {
		return checker.CreateRuntimeErrorResult(CheckBranchProtection, sce.Create(sce.ErrRunFailure, err.Error()))
	}

	return IsBranchProtected(protection, branch, dl)
}

func IsBranchProtected(protection *github.Protection, branch string, dl checker.DetailLogger) checker.CheckResult {
	totalChecks := 10
	totalSuccess := 0

	// This is disabled by default (good).
	if protection.GetAllowForcePushes() != nil &&
		protection.AllowForcePushes.Enabled {
		dl.Warn("AllowForcePushes enabled on branch '%s'", branch)
	} else {
		dl.Info("AllowForcePushes disabled on branch '%s'", branch)
		totalSuccess++
	}

	// This is disabled by default (good).
	if protection.GetAllowDeletions() != nil &&
		protection.AllowDeletions.Enabled {
		dl.Warn("AllowDeletions enabled on branch '%s'", branch)
	} else {
		dl.Info("AllowDeletions disabled on branch '%s'", branch)
		totalSuccess++
	}

	// This is disabled by default (bad).
	if protection.GetEnforceAdmins() != nil &&
		protection.EnforceAdmins.Enabled {
		dl.Info("EnforceAdmins disabled on branch '%s'", branch)
		totalSuccess++
	} else {
		dl.Warn("EnforceAdmins disabled on branch '%s'", branch)
	}

	// This is disabled by default (bad).
	if protection.GetRequireLinearHistory() != nil &&
		protection.RequireLinearHistory.Enabled {
		dl.Info("Linear history enabled on branch '%s'", branch)
		totalSuccess++
	} else {
		dl.Warn("Linear history disabled on branch '%s'", branch)
	}

	if requiresStatusChecks(protection, branch, dl) {
		dl.Info("Strict status check enabled on branch '%s'", branch)
		totalSuccess++
	}

	if requiresThoroughReviews(protection, branch, dl) {
		totalSuccess++
	}

	return checker.CreateProportionalScoreResult(CheckBranchProtection,
		"%d out of %d branch protection settings are enabled", totalSuccess, totalChecks)
}

// Returns true if several PR status checks requirements are enabled. Otherwise returns false and logs why it failed.
func requiresStatusChecks(protection *github.Protection, branch string, dl checker.DetailLogger) bool {
	// This is disabled by default (bad).
	if protection.GetRequiredStatusChecks() != nil &&
		protection.RequiredStatusChecks.Strict &&
		len(protection.RequiredStatusChecks.Contexts) > 0 {
		return true
	}
	switch {
	case protection.RequiredStatusChecks == nil ||
		!protection.RequiredStatusChecks.Strict:
		dl.Warn("Status checks for merging disabled on branch '%s'", branch)
	case len(protection.RequiredStatusChecks.Contexts) == 0:
		dl.Warn("Status checks for merging have no specific status to check on branch '%s'", branch)
	default:
		panic("Unhandled status checks error")
	}
	return false
}

// Returns true if several PR review requirements are enabled. Otherwise returns false and logs why it failed.
func requiresThoroughReviews(protection *github.Protection, branch string, dl checker.DetailLogger) bool {
	// This is disabled by default (bad).
	if protection.GetRequiredPullRequestReviews() != nil &&
		protection.RequiredPullRequestReviews.RequiredApprovingReviewCount >= minReviews &&
		protection.RequiredPullRequestReviews.DismissStaleReviews &&
		protection.RequiredPullRequestReviews.RequireCodeOwnerReviews {
		return true
	}
	switch {
	case protection.RequiredPullRequestReviews == nil:
		dl.Warn("Pullrequest reviews disabled on branch '%s'", branch)
		fallthrough
	case protection.RequiredPullRequestReviews.RequiredApprovingReviewCount < minReviews:
		dl.Warn("Number of required reviewers is only %d on branch '%s'",
			protection.RequiredPullRequestReviews.RequiredApprovingReviewCount, branch)
		fallthrough
	case !protection.RequiredPullRequestReviews.DismissStaleReviews:
		dl.Warn("Stale review dismissal disabled on branch '%s'", branch)
		fallthrough
	case !protection.RequiredPullRequestReviews.RequireCodeOwnerReviews:
		dl.Warn("Owner review not required on branch '%s'", branch)
	default:
		panic("Unhandled pull request error")
	}
	return false
}
