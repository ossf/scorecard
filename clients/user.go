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

package clients

// User represents a Git user.
type User struct {
	Login string
}

// RepoAssociation is how a user is associated with a repository.
type RepoAssociation int32

// Values taken from https://docs.github.com/en/graphql/reference/enums#commentauthorassociation. Additional
// values may be added in the future for non-Github projects.
const (
	// Collaborator: Author has been invited to collaborate on the repository.
	RepoAssociationCollaborator RepoAssociation = iota
	// Contributor: Author has been a contributor to the repository.
	RepoAssociationContributor
	// FirstTimer: Author has previously committed to the repository.
	RepoAssociationFirstTimer
	// FirstTimeContributor: Author has not previously committed to the repository.
	RepoAssociationFirstTimeContributor
	// Mannequin: Author is a placeholder for an unclaimed user.
	RepoAssociationMannequin
	// Member: Author is a member of the organization that owns the repository.
	RepoAssociationMember
	// None: Author has no association with the repository.
	RepoAssociationNone
	// Owner: Author is the owner of the repository.
	RepoAssociationOwner
)
