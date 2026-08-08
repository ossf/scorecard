// Copyright 2024 OpenSSF Scorecard Authors
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

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// accessLevelHandler handles access level verification for protected tags.
type accessLevelHandler struct {
	glClient *gitlab.Client
	repourl  *Repo
}

func (handler *accessLevelHandler) init(repourl *Repo) {
	handler.repourl = repourl
}

// getMinimumAccessLevel determines the most restrictive access level
// from a list of TagAccessDescriptions.
//
// Returns:
// - 0 if "No one" is specified (most restrictive)
// - Minimum access level among all users and groups
// - Error if levels cannot be determined.
func (handler *accessLevelHandler) getMinimumAccessLevel(
	accessLevels []*gitlab.TagAccessDescription,
) (gitlab.AccessLevelValue, error) {
	if len(accessLevels) == 0 {
		// No restrictions means Developer+ can create
		return gitlab.DeveloperPermissions, nil
	}

	minLevel := gitlab.AccessLevelValue(100)

	for _, access := range accessLevels {
		if access == nil {
			continue
		}

		level, err := handler.resolveAccessLevel(access)
		if err != nil {
			return 0, err
		}

		if level < minLevel {
			minLevel = level
		}
	}

	if minLevel == 100 {
		return gitlab.DeveloperPermissions, nil
	}

	return minLevel, nil
}

// resolveAccessLevel gets effective access level from a description.
func (handler *accessLevelHandler) resolveAccessLevel(
	access *gitlab.TagAccessDescription,
) (gitlab.AccessLevelValue, error) {
	// Check for direct access level (including NoPermissions = 0)
	hasNoIDs := access.UserID == 0 && access.GroupID == 0
	if access.AccessLevel >= 0 && hasNoIDs {
		return access.AccessLevel, nil
	}

	// Resolve user access level
	if access.UserID > 0 {
		level, err := handler.getUserAccessLevel(access.UserID)
		if err != nil {
			return 0, fmt.Errorf(
				"failed to get access for user %d: %w",
				access.UserID,
				err,
			)
		}
		return level, nil
	}

	// Resolve group access level
	if access.GroupID > 0 {
		level, err := handler.getGroupMinimumAccessLevel(access.GroupID)
		if err != nil {
			return 0, fmt.Errorf(
				"failed to get access for group %d: %w",
				access.GroupID,
				err,
			)
		}
		return level, nil
	}

	// Malformed entry - skip it
	return 100, nil
}

// getUserAccessLevel fetches the access level for a specific user
// in the project.
func (handler *accessLevelHandler) getUserAccessLevel(userID int64) (gitlab.AccessLevelValue, error) {
	member, _, err := handler.glClient.ProjectMembers.GetProjectMember(
		handler.repourl.projectID,
		userID,
		nil,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to get project member: %w", err)
	}

	return member.AccessLevel, nil
}

// getGroupMinimumAccessLevel gets the minimum access level among all members
// of a group who are also project members.
func (handler *accessLevelHandler) getGroupMinimumAccessLevel(groupID int64) (gitlab.AccessLevelValue, error) {
	// Get all project members
	projectMembers, _, err := handler.glClient.ProjectMembers.ListAllProjectMembers(
		handler.repourl.projectID,
		&gitlab.ListProjectMembersOptions{},
	)
	if err != nil {
		return 0, fmt.Errorf("failed to list project members: %w", err)
	}

	// Get all group members
	groupMembers, _, err := handler.glClient.Groups.ListGroupMembers(
		int(groupID),
		&gitlab.ListGroupMembersOptions{},
	)
	if err != nil {
		return 0, fmt.Errorf("failed to list group members: %w", err)
	}

	// Create a map of group member IDs for quick lookup
	groupMemberIDs := make(map[int64]bool)
	for _, gm := range groupMembers {
		groupMemberIDs[gm.ID] = true
	}

	// Find the minimum access level among group members
	// who are also project members
	minAccessLevel := gitlab.AccessLevelValue(100)
	foundAny := false

	for _, pm := range projectMembers {
		if groupMemberIDs[pm.ID] {
			foundAny = true
			if pm.AccessLevel < minAccessLevel {
				minAccessLevel = pm.AccessLevel
			}
		}
	}

	if !foundAny {
		// If no group members are project members, the group
		// effectively has no access. This shouldn't happen in practice,
		// but we'll return Developer as a safe default
		return gitlab.DeveloperPermissions, nil
	}

	return minAccessLevel, nil
}
