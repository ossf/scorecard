// Copyright 2021 OpenSSF Scorecard Authors
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

package clients

// User represents a Git user.
type User struct {
	Login            string
	Companies        []string
	Organizations    []User
	NumContributions int
	ID               int64
	IsBot            bool
}

// RepoAssociation is how a user is associated with a repository.
type RepoAssociation uint32

// Values taken from https://docs.github.com/en/graphql/reference/enums#commentauthorassociation.
// Additional values may be added in the future for non-Github projects.
// NOTE: Values are present in increasing order of privilege. If adding new values
// maintain the order of privilege to ensure Gte() functionality is preserved.
const (
	// Mannequin: Author is a placeholder for an unclaimed user.
	RepoAssociationMannequin RepoAssociation = iota
	// None: Author has no association with the repository.
	// NoPermissions: (GitLab).
	RepoAssociationNone
	// FirstTimer: Author has not previously committed to the VCS.
	RepoAssociationFirstTimer
	// FirstTimeContributor: Author has not previously committed to the repository.
	// MinimalAccessPermissions: (Gitlab).
	RepoAssociationFirstTimeContributor
	// Contributor: Author has been a contributor to the repository.
	RepoAssociationContributor
	// Collaborator: Author has been invited to collaborate on the repository.
	RepoAssociationCollaborator
	// Member: Author is a member of the organization that owns the repository.
	// DeveloperAccessPermissions: (GitLab).
	RepoAssociationMember
	// Maintainer: Author is part of the maintenance team for the repository (GitLab).
	RepoAssociationMaintainer
	// Owner: Author is the owner of the repository.
	// (Owner): (GitLab).
	RepoAssociationOwner
)

// Gte is >= comparator for RepoAssociation enum.
func (r RepoAssociation) Gte(val RepoAssociation) bool {
	return r >= val
}

// String returns an string value for RepoAssociation enum.
func (r RepoAssociation) String() string {
	switch r {
	case RepoAssociationMannequin:
		return "unknown"
	case RepoAssociationNone:
		return "none"
	case RepoAssociationFirstTimer:
		return "first-timer"
	case RepoAssociationFirstTimeContributor:
		return "first-time-contributor"
	case RepoAssociationContributor:
		return "contributor"
	case RepoAssociationCollaborator:
		return "collaborator"
	case RepoAssociationMember:
		return "member"
	case RepoAssociationOwner:
		return "owner"
	default:
		return ""
	}
}
