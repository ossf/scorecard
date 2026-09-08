// Copyright 2025 OpenSSF Scorecard Authors
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

	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
)

// gitLabClient defines GitLab-specific methods for tag protection.
type gitLabClient interface {
	ListBranches() ([]string, error)
	GetProtectedTagPatterns() ([]*gitlab.ProtectedTag, error)
	GetMinimumAccessLevel(
		[]*gitlab.TagAccessDescription,
	) (gitlab.AccessLevelValue, error)
}

type tagSet struct {
	exists map[string]bool
	set    []clients.TagRef
}

func (set *tagSet) add(tag *clients.TagRef) bool {
	if tag != nil &&
		tag.Name != nil &&
		*tag.Name != "" &&
		!set.exists[*tag.Name] {
		set.set = append(set.set, *tag)
		set.exists[*tag.Name] = true
		return true
	}
	return false
}

func (set *tagSet) contains(tagName string) bool {
	_, contains := set.exists[tagName]
	return contains
}

// TagProtection retrieves the raw data for the Tag-Protection check.
func TagProtection(cr *checker.CheckRequest) (checker.TagProtectionsData, error) {
	c := cr.RepoClient
	tags := tagSet{
		exists: make(map[string]bool),
	}

	// Get releases to check for tag protection
	releases, err := c.ListReleases()
	if err != nil && !errors.Is(err, clients.ErrUnsupportedFeature) {
		return checker.TagProtectionsData{}, fmt.Errorf("%w", err)
	}

	// No releases means no tags to protect
	if len(releases) == 0 {
		return checker.TagProtectionsData{
			Tags: []clients.TagRef{},
		}, nil
	}

	// Extract and deduplicate tags from releases
	for _, release := range releases {
		if release.TagName == "" {
			continue
		}

		if tags.contains(release.TagName) {
			continue
		}

		tagRef, err := c.GetTag(release.TagName)
		if err != nil {
			if errors.Is(err, clients.ErrUnsupportedFeature) {
				return checker.TagProtectionsData{}, fmt.Errorf(
					"getting tag %s: %w",
					release.TagName,
					err,
				)
			}
			// Skip on other errors (network issues, etc.)
			continue
		}

		tags.add(tagRef)
	}

	result := checker.TagProtectionsData{
		Tags: tags.set,
	}

	// Collect GitLab-specific data if available
	if glClient, ok := c.(gitLabClient); ok {
		collectGitLabData(&result, glClient)
	}

	return result, nil
}

// collectGitLabData gathers GitLab-specific tag protection data.
func collectGitLabData(
	result *checker.TagProtectionsData,
	client gitLabClient,
) {
	// Get branches for branch shadowing analysis
	branches, err := client.ListBranches()
	if err == nil {
		result.GitLabBranches = branches
	}

	// Get protected tag patterns with access levels
	patterns, err := client.GetProtectedTagPatterns()
	if err != nil {
		return
	}

	for _, pattern := range patterns {
		info := convertPatternToInfo(pattern, client)
		if info != nil {
			result.GitLabProtectedTags = append(
				result.GitLabProtectedTags,
				*info,
			)
		}
	}
}

// convertPatternToInfo converts a GitLab pattern to tag info.
func convertPatternToInfo(
	pattern *gitlab.ProtectedTag,
	client gitLabClient,
) *checker.GitLabProtectedTagInfo {
	minAccess, err := client.GetMinimumAccessLevel(
		pattern.CreateAccessLevels,
	)
	if err != nil {
		return nil
	}

	return &checker.GitLabProtectedTagInfo{
		Pattern:           pattern.Name,
		CreateAccessLevel: int(minAccess),
	}
}
