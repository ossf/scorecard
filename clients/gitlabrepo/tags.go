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

package gitlabrepo

import (
	"fmt"
	"strings"
	"sync"

	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/clients/gitlabrepo/internal/fnmatch"
)

type tagsHandler struct {
	glClient          *gitlab.Client
	once              *sync.Once
	errSetup          error
	repourl           *Repo
	getProtectedTag   fnProtectedTag
	listProtectedTags fnListProtectedTags
	protectedPatterns []*gitlab.ProtectedTag
}

type (
	fnProtectedTag func(pid interface{}, tag string,
		options ...gitlab.RequestOptionFunc) (*gitlab.ProtectedTag, *gitlab.Response, error)
	fnListProtectedTags func(pid interface{}, opt *gitlab.ListProtectedTagsOptions,
		options ...gitlab.RequestOptionFunc) ([]*gitlab.ProtectedTag, *gitlab.Response, error)
)

func (handler *tagsHandler) init(repourl *Repo) {
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
	handler.getProtectedTag = handler.glClient.ProtectedTags.GetProtectedTag
	handler.listProtectedTags = handler.glClient.ProtectedTags.ListProtectedTags
	handler.protectedPatterns = nil
}

func (handler *tagsHandler) setup() error {
	handler.once.Do(func() {
		if !strings.EqualFold(handler.repourl.commitSHA, clients.HeadSHA) {
			handler.errSetup = fmt.Errorf(
				"%w: tags only supported for HEAD queries",
				clients.ErrUnsupportedFeature,
			)
			return
		}

		// Fetch all protected tag patterns for this project
		patterns, _, err := handler.listProtectedTags(
			handler.repourl.projectID,
			&gitlab.ListProtectedTagsOptions{},
		)
		if err != nil {
			// If we can't fetch protected tags,
			// we'll treat all tags as unprotected
			// rather than failing the entire check
			handler.protectedPatterns = nil
			handler.errSetup = nil
			return
		}
		handler.protectedPatterns = patterns
		handler.errSetup = nil
	})
	return handler.errSetup
}

func (handler *tagsHandler) getTag(tagName string) (*clients.TagRef, error) {
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during tagsHandler.setup: %w", err)
	}

	// Check if this tag matches any protected tag patterns
	matchedPatterns := handler.findMatchingPatterns(tagName)
	if len(matchedPatterns) == 0 {
		// Tag is not protected by any pattern
		notProtected := false
		noRestriction := false
		return &clients.TagRef{
			Name:      &tagName,
			Protected: &notProtected,
			TagProtectionRule: clients.TagProtectionRule{
				RefProtectionRule: clients.RefProtectionRule{
					AllowDeletions:       newTrue(),
					AllowForcePushes:     newTrue(),
					RequireLinearHistory: newFalse(),
					EnforceAdmins:        newFalse(),
					CheckRules:           clients.StatusChecksRule{},
				},
				AllowUpdates:      newTrue(),
				RequireSignatures: newFalse(),
				RestrictCreation:  &noRestriction,
			},
		}, nil
	}

	// Tag is protected, build the TagProtectionRule
	// from all matched patterns.
	// GitLab inheritance: if multiple patterns match,
	// settings are combined
	return makeTagRefFrom(tagName, matchedPatterns), nil
}

// findMatchingPatterns checks if the tag name
// matches any protected tag patterns.
// Returns all matching patterns. If multiple
// patterns match, GitLab inherits settings
// from all of them (union of permissions).
func (handler *tagsHandler) findMatchingPatterns(
	tagName string,
) []*gitlab.ProtectedTag {
	var matches []*gitlab.ProtectedTag
	for _, pattern := range handler.protectedPatterns {
		if pattern.Name == "" {
			continue
		}
		// Try exact match first
		if pattern.Name == tagName {
			matches = append(matches, pattern)
			continue
		}
		// Try wildcard match using fnmatch
		matched, err := fnmatch.Match(pattern.Name, tagName)
		if err != nil {
			// Pattern is invalid, skip it
			continue
		}
		if matched {
			matches = append(matches, pattern)
		}
	}
	return matches
}

// mergeCreateAccessLevels combines
// CreateAccessLevels from all
// matching patterns.
// GitLab's inheritance model:
// if multiple patterns match,
// settings are combined (union).
// Returns a deduplicated list of
// all access levels from all patterns.
func mergeCreateAccessLevels(
	patterns []*gitlab.ProtectedTag,
) []*gitlab.TagAccessDescription {
	if len(patterns) == 0 {
		return nil
	}

	// Use a map to deduplicate by access level value
	accessMap := make(map[gitlab.AccessLevelValue]*gitlab.TagAccessDescription)

	for _, pattern := range patterns {
		for _, access := range pattern.CreateAccessLevels {
			if access != nil {
				// Store by AccessLevel to deduplicate
				accessMap[access.AccessLevel] = access
			}
		}
	}

	// Convert map back to slice
	var result []*gitlab.TagAccessDescription
	for _, access := range accessMap {
		result = append(result, access)
	}

	return result
}

func makeTagRefFrom(
	tagName string,
	protectedTags []*gitlab.ProtectedTag,
) *clients.TagRef {
	protected := true

	// Merge CreateAccessLevels from all matching patterns
	// GitLab inheritance: if any pattern has creation
	// restrictions, combine them (union)
	mergedAccessLevels := mergeCreateAccessLevels(protectedTags)
	restrictCreation := len(mergedAccessLevels) > 0

	// Note: GitLab's protected tag API only
	// provides information about who can create tags.
	// However, GitLab implicitly provides the following
	// protections for protected tags:
	// - Deletion is always blocked (set to false)
	// - Update is always blocked (set to false)
	// - Admin enforcement is always on (set to true)
	//
	// While not explicitly exposed in the API response,
	// these are documented GitLab behaviors
	// for protected tags, so we set them accordingly
	// rather than leaving them as nil.

	return &clients.TagRef{
		Name:      &tagName,
		Protected: &protected,
		TagProtectionRule: clients.TagProtectionRule{
			RefProtectionRule: clients.RefProtectionRule{
				AllowDeletions:       newFalse(), // GitLab implicitly blocks deletion
				AllowForcePushes:     nil,        // Not applicable to tags
				RequireLinearHistory: nil,        // Not applicable to tags
				EnforceAdmins:        newTrue(),  // GitLab implicitly enforces for admins
				CheckRules:           clients.StatusChecksRule{},
			},
			AllowUpdates:      newFalse(), // GitLab implicitly blocks updates
			RequireSignatures: newFalse(), // GitLab doesn't support this
			RestrictCreation:  &restrictCreation,
		},
	}
}

// getProtectedPatterns returns all protected tag patterns for use by probes.
// This allows probes to analyze raw GitLab protection settings.
func (handler *tagsHandler) getProtectedPatterns() ([]*gitlab.ProtectedTag, error) {
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during tagsHandler.setup: %w", err)
	}
	return handler.protectedPatterns, nil
}
