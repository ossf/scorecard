// Copyright 2021 Security Scorecard Authors
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

package raw

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
)

type branchMap map[string]*clients.BranchRef

// BranchProtection retrieves the raw data for the Branch-Protection check.
func BranchProtection(c clients.RepoClient) (checker.BranchProtectionsData, error) {
	// Checks branch protection on both release and development branch.
	// Get all branches. This will include information on whether they are protected.
	branches, err := c.ListBranches()
	if err != nil {
		return checker.BranchProtectionsData{}, fmt.Errorf("%w", err)
	}
	branchesMap := getBranchMapFrom(branches)

	// Get release branches.
	releases, err := c.ListReleases()
	if err != nil {
		return checker.BranchProtectionsData{}, fmt.Errorf("%w", err)
	}

	commit := regexp.MustCompile("^[a-f0-9]{40}$")
	checkBranches := make(map[string]bool)
	for _, release := range releases {
		if release.TargetCommitish == "" {
			// Log with a named error if target_commitish is nil.
			return checker.BranchProtectionsData{}, fmt.Errorf("%w", errInternalCommitishNil)
		}

		// TODO: if this is a sha, get the associated branch. for now, ignore.
		if commit.Match([]byte(release.TargetCommitish)) {
			continue
		}

		// Try to resolve the branch name.
		b, err := branchesMap.getBranchByName(release.TargetCommitish)
		if err != nil {
			// If the commitish branch is still not found, fail.
			return checker.BranchProtectionsData{}, err
		}

		// Branch is valid, add to list of branches to check.
		checkBranches[*b.Name] = true
	}

	// Add default branch.
	defaultBranch, err := c.GetDefaultBranch()
	if err != nil {
		return checker.BranchProtectionsData{}, fmt.Errorf("%w", err)
	}
	defaultBranchName := getBranchName(defaultBranch)
	if defaultBranchName != "" {
		checkBranches[defaultBranchName] = true
	}

	rawData := checker.BranchProtectionsData{}
	// Check protections on all the branches.
	for b := range checkBranches {
		branch, err := branchesMap.getBranchByName(b)
		if err != nil {
			if errors.Is(err, errInternalBranchNotFound) {
				continue
			}
			return checker.BranchProtectionsData{}, err
		}

		// Protected field only indates that the branch matches
		// one `Branch protection rules`. All settings may be disabled,
		// so it does not provide any guarantees.
		protected := !(branch.Protected != nil && !*branch.Protected)
		bpData := checker.BranchProtectionData{Name: b}
		bp := branch.BranchProtectionRule
		bpData.Protected = &protected
		bpData.RequiresLinearHistory = bp.RequireLinearHistory
		bpData.AllowsForcePushes = bp.AllowForcePushes
		bpData.AllowsDeletions = bp.AllowDeletions
		bpData.EnforcesAdmins = bp.EnforceAdmins
		bpData.RequiresCodeOwnerReviews = bp.RequiredPullRequestReviews.RequireCodeOwnerReviews
		bpData.DismissesStaleReviews = bp.RequiredPullRequestReviews.DismissStaleReviews
		bpData.RequiresUpToDateBranchBeforeMerging = bp.CheckRules.UpToDateBeforeMerge
		if bp.RequiredPullRequestReviews.RequiredApprovingReviewCount != nil {
			v := int(*bp.RequiredPullRequestReviews.RequiredApprovingReviewCount)
			bpData.RequiredApprovingReviewCount = &v
		}
		bpData.StatusCheckContexts = bp.CheckRules.Contexts

		rawData.Branches = append(rawData.Branches, bpData)
	}

	// No error, return the data.
	return rawData, nil
}

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
	return nil, sce.WithMessage(sce.ErrScorecardInternal,
		fmt.Sprintf("could not find branch name %s: %v", name, errInternalBranchNotFound))
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
