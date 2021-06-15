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
	"regexp"

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
	ListBranches(ctx context.Context, owner string, repo string,
		opts *github.BranchListOptions) ([]*github.Branch, *github.Response, error)
	ListReleases(ctx context.Context, owner string, repo string, opts *github.ListOptions) (
		[]*github.RepositoryRelease, *github.Response, error)
	GetBranchProtection(context.Context, string, string, string) (
		*github.Protection, *github.Response, error)
}

type logger func(s string, f ...interface{})

// ErrCommitishNil: TargetCommitish nil for release
var ErrCommitishNil = errors.New("target_commitish is nil for release")

// ErrBranchNotFound: branch from TargetCommitish not found
var ErrBranchNotFound = errors.New("branch not found")

func BranchProtection(c *checker.CheckRequest) checker.CheckResult {
	// Checks branch protection on both release and development branch
	return checkReleaseAndDevBranchProtection(c.Ctx, c.Client.Repositories, c.Logf, c.Owner, c.Repo)
}

func checkReleaseAndDevBranchProtection(ctx context.Context, r repositories, l logger, ownerStr,
	repoStr string) checker.CheckResult {
	// Get all branches. This will include information on whether they are protected.
	branches, _, err := r.ListBranches(ctx, ownerStr, repoStr, &github.BranchListOptions{})
	if err != nil {
		return checker.MakeRetryResult(CheckBranchProtection, err)
	}

	// Get release branches
	releases, _, err := r.ListReleases(ctx, ownerStr, repoStr, &github.ListOptions{})
	if err != nil {
		return checker.MakeRetryResult(CheckBranchProtection, err)
	}

	var checks []checker.CheckResult
	commit := regexp.MustCompile("^[a-f0-9]{40}$")
	checkBranches := make(map[string]bool)
	for _, release := range releases {
		if release.TargetCommitish == nil {
			// Log with a named error if target_commitish is nil.
			checks = append(checks, checker.MakeFailResult(CheckBranchProtection, ErrCommitishNil))
			continue
		}

		// TODO: if this is a sha, get the associated branch. for now, ignore.
		if commit.Match([]byte(*release.TargetCommitish)) {
			continue
		}

		// Try to resolve the branch name.
		name, err := resolveBranchName(ctx, r, ownerStr, repoStr, branches, *release.TargetCommitish)
		if err != nil {
			// If the commitish branch is still not found, fail.
			checks = append(checks, checker.MakeFailResult(CheckBranchProtection, ErrBranchNotFound))
			continue
		}

		// Branch is valid, add to list of branches to check.
		checkBranches[*name] = true
	}

	// Add default branch
	repo, _, err := r.Get(ctx, ownerStr, repoStr)
	if err != nil {
		return checker.MakeRetryResult(CheckBranchProtection, err)
	}
	checkBranches[*repo.DefaultBranch] = true

	// Check protections on the branches.
	for b := range checkBranches {
		protected, err := isBranchProtected(branches, b)
		if err != nil {
			checks = append(checks, checker.MakeFailResult(CheckBranchProtection, ErrBranchNotFound))
		}
		if !protected {
			l("!! branch protection not enabled for branch %s", b)
			checks = append(checks, checker.CheckResult{
				Name:       CheckBranchProtection,
				Pass:       false,
				Confidence: checker.MaxResultConfidence,
			})
		} else {
			// The branch is protected. Check the protection.
			res := getProtectionAndCheck(ctx, r, l, ownerStr, repoStr, b)
			checks = append(checks, res)
		}
	}

	return checker.MultiCheckResultAnd(checks...)
}

func resolveBranchName(ctx context.Context, r repositories, ownerStr, repoStr string,
	branches []*github.Branch, name string) (*string, error) {
	fmt.Printf("finding branch %s\n", name)
	// First check list of branches.
	for _, b := range branches {
		if b.GetName() == name {
			fmt.Printf("found branch in branches %s\n", b.GetName())
			return b.Name, nil
		}
	}
	// Ideally, we should check using repositories.GetBranch if there was a branch redirect.
	// See https://github.com/google/go-github/issues/1895
	// For now, handle the common master -> main redirect.
	if name == "master" {
		return resolveBranchName(ctx, r, ownerStr, repoStr, branches, "main")
	}

	return nil, errors.New("branch not found")
}

func isBranchProtected(branches []*github.Branch, name string) (bool, error) {
	// Returns bool indicating if protected.
	for _, b := range branches {
		if b.GetName() == name {
			return b.GetProtected(), nil
		}
	}
	return false, ErrBranchNotFound
}

func getProtectionAndCheck(ctx context.Context, r repositories, l logger, ownerStr, repoStr,
	branch string) checker.CheckResult {
	// We only call this if the branch is protected. An error indicates not found.
	protection, resp, err := r.GetBranchProtection(ctx, ownerStr, repoStr, branch)

	const fileNotFound = 404
	if resp.StatusCode == fileNotFound {
		return checker.MakeRetryResult(CheckBranchProtection, err)
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
