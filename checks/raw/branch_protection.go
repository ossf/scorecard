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
	"fmt"
	"regexp"
	"strings"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
)

const master = "master"

var commit = regexp.MustCompile("^[a-f0-9]{40}$")

type branchSet struct {
	exists map[string]bool
	set    []clients.BranchRef
}

func (set *branchSet) add(branch *clients.BranchRef) bool {
	if branch != nil &&
		branch.Name != nil &&
		*branch.Name != "" &&
		!set.exists[*branch.Name] {
		set.set = append(set.set, *branch)
		set.exists[*branch.Name] = true
		return true
	}
	return false
}

func (set branchSet) contains(branch string) bool {
	_, contains := set.exists[branch]
	return contains
}

// BranchProtection retrieves the raw data for the Branch-Protection check.
func BranchProtection(c clients.RepoClient) (checker.BranchProtectionsData, error) {
	branches := branchSet{
		exists: make(map[string]bool),
	}
	// Add default branch.
	defaultBranch, err := c.GetDefaultBranch()
	if err != nil {
		return checker.BranchProtectionsData{}, fmt.Errorf("%w", err)
	}
	branches.add(defaultBranch)

	// Get release branches.
	releases, err := c.ListReleases()
	if err != nil {
		return checker.BranchProtectionsData{}, fmt.Errorf("%w", err)
	}
	for _, release := range releases {
		if release.TargetCommitish == "" {
			// Log with a named error if target_commitish is nil.
			return checker.BranchProtectionsData{}, fmt.Errorf("%w", errInternalCommitishNil)
		}

		// TODO: if this is a sha, get the associated branch. for now, ignore.
		if commit.MatchString(release.TargetCommitish) && !strings.Contains(c.URI(), "gitlab.") {
			continue
		}

		if branches.contains(release.TargetCommitish) ||
			branches.contains(branchRedirect(release.TargetCommitish)) {
			continue
		}

		// Get the associated release branch.
		branchRef, err := c.GetBranch(release.TargetCommitish)
		if err != nil {
			return checker.BranchProtectionsData{},
				fmt.Errorf("error during GetBranch(%s): %w", release.TargetCommitish, err)
		}
		if branches.add(branchRef) {
			continue
		}

		// Couldn't find the branch check for redirects.
		redirectBranch := branchRedirect(release.TargetCommitish)
		if redirectBranch == "" {
			continue
		}
		branchRef, err = c.GetBranch(redirectBranch)
		if err != nil {
			return checker.BranchProtectionsData{},
				fmt.Errorf("error during GetBranch(%s) %w", redirectBranch, err)
		}
		branches.add(branchRef)
		// Branch doesn't exist or was deleted. Continue.
	}

	// No error, return the data.
	return checker.BranchProtectionsData{
		Branches: branches.set,
	}, nil
}

func branchRedirect(name string) string {
	// Ideally, we should check using repositories.GetBranch if there was a branch redirect.
	// See https://github.com/google/go-github/issues/1895
	// For now, handle the common master -> main redirect.
	if name == master {
		return "main"
	}
	return ""
}
