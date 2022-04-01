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

package checker

import "time"

// RawResults contains results before a policy
// is applied.
//nolint
type RawResults struct {
	VulnerabilitiesResults      VulnerabilitiesData
	BinaryArtifactResults       BinaryArtifactData
	SecurityPolicyResults       SecurityPolicyData
	DependencyUpdateToolResults DependencyUpdateToolData
	BranchProtectionResults     BranchProtectionsData
	CodeReviewResults           CodeReviewData
	WebhookResults              WebhooksData
	MaintainedResults           MaintainedData
	SignedReleasesResults       SignedReleasesData
}

// MaintainedData contains the raw results
// for the Maintained check.
type MaintainedData struct {
	Issues               []Issue
	DefaultBranchCommits []DefaultBranchCommit
	ArchivedStatus       ArchivedStatus
}

// CodeReviewData contains the raw results
// for the Code-Review check.
type CodeReviewData struct {
	DefaultBranchCommits []DefaultBranchCommit
}

// VulnerabilitiesData contains the raw results
// for the Vulnerabilities check.
type VulnerabilitiesData struct {
	Vulnerabilities []Vulnerability
}

// SecurityPolicyData contains the raw results
// for the Security-Policy check.
type SecurityPolicyData struct {
	// Files contains a list of files.
	Files []File
}

// BinaryArtifactData contains the raw results
// for the Binary-Artifact check.
type BinaryArtifactData struct {
	// Files contains a list of files.
	Files []File
}

// SignedReleasesData contains the raw results
// for the Signed-Releases check.
type SignedReleasesData struct {
	Releases []Release
}

// DependencyUpdateToolData contains the raw results
// for the Dependency-Update-Tool check.
type DependencyUpdateToolData struct {
	// Tools contains a list of tools.
	// Note: we only populate one entry at most.
	Tools []Tool
}

// WebhooksData contains the raw results
// for the Webhook check.
type WebhooksData struct {
	Webhook []WebhookData
}

// WebhookData contains the raw results
// for webhook check.
type WebhookData struct {
	Path           string
	ID             int64
	UsesAuthSecret bool
}

// BranchProtectionsData contains the raw results
// for the Branch-Protection check.
type BranchProtectionsData struct {
	Branches []BranchProtectionData
}

// BranchProtectionData contains the raw results
// for one branch.
//nolint:govet
type BranchProtectionData struct {
	Protected                           *bool
	AllowsDeletions                     *bool
	AllowsForcePushes                   *bool
	RequiresCodeOwnerReviews            *bool
	RequiresLinearHistory               *bool
	DismissesStaleReviews               *bool
	EnforcesAdmins                      *bool
	RequiresStatusChecks                *bool
	RequiresUpToDateBranchBeforeMerging *bool
	RequiredApprovingReviewCount        *int
	// StatusCheckContexts is always available, so
	// we don't use a pointer.
	StatusCheckContexts []string
	Name                string
}

// Tool represents a tool.
type Tool struct {
	// Runs of the tool.
	Runs []Run
	// Issues created by the tool.
	Issues []Issue
	// Merge requests created by the tool.
	MergeRequests []MergeRequest
	Name          string
	URL           string
	Desc          string
	ConfigFiles   []File
}

// Run represents a run.
type Run struct {
	URL string
	// TODO: add fields, e.g., Result=["success", "failure"]
}

// Comment represents a comment for a pull request or an issue.
type Comment struct {
	CreatedAt *time.Time
	Author    *User
	// TODO: add ields if needed, e.g., content.
}

// ArchivedStatus definess the archived status.
type ArchivedStatus struct {
	Status bool
	// TODO: add fields, e.g., date of archival.
}

// Issue represents an issue.
type Issue struct {
	CreatedAt *time.Time
	Author    *User
	URL       string
	Comments  []Comment
	// TODO: add fields, e.g., state=[opened|closed]
}

// DefaultBranchCommit represents a commit
// to the default branch.
type DefaultBranchCommit struct {
	// Fields below are taken directly from cloud
	// version control systems, e.g. GitHub.
	SHA           string
	CommitMessage string
	MergeRequest  *MergeRequest
	CommitDate    *time.Time
	Committer     User
}

// MergeRequest represents a merge request.
// nolint:govet
type MergeRequest struct {
	Number   int
	Labels   []string
	Reviews  []Review
	Author   User
	MergedAt time.Time
}

// Review represent a review using the built-in review system.
type Review struct {
	Reviewer User
	State    string
	// TODO(Review): add fields here if needed.
}

// User represent a user.
type User struct {
	RepoAssociation *RepoAssociation
	Login           string
}

// RepoAssociation represents a user relationship with a repo.
type RepoAssociation string

const (
	// RepoAssociationCollaborator has been invited to collaborate on the repository.
	RepoAssociationCollaborator RepoAssociation = RepoAssociation("collaborator")
	// RepoAssociationContributor is an contributor to the repository.
	RepoAssociationContributor RepoAssociation = RepoAssociation("contributor")
	// RepoAssociationOwner is an owner of the repository.
	RepoAssociationOwner RepoAssociation = RepoAssociation("owner")
	// RepoAssociationMember is a member of the organization that owns the repository.
	RepoAssociationMember RepoAssociation = RepoAssociation("member")
	// RepoAssociationFirstTimer has previously committed to the repository.
	RepoAssociationFirstTimer RepoAssociation = RepoAssociation("first-timer")
	// RepoAssociationFirstTimeContributor has not previously committed to the repository.
	RepoAssociationFirstTimeContributor RepoAssociation = RepoAssociation("first-timer-contributor")
	// RepoAssociationMannequin is a placeholder for an unclaimed user.
	RepoAssociationMannequin RepoAssociation = RepoAssociation("unknown")
	// RepoAssociationNone has no association with the repository.
	RepoAssociationNone RepoAssociation = RepoAssociation("none")
)

// File represents a file.
type File struct {
	Path    string
	Snippet string   // Snippet of code
	Offset  uint     // Offset in the file of Path (line for source/text files).
	Type    FileType // Type of file.
	// TODO: add hash.
}

// Vulnerability defines a vulnerability
// from a database.
type Vulnerability struct {
	// For OSV: OSV-2020-484
	// For CVE: CVE-2022-23945
	ID string
	// TODO(vuln): Add additional fields, if needed.
}

// Release represents a project release.
type Release struct {
	Tag    string
	URL    string
	Assets []ReleaseAsset
	// TODO: add needed fields, e.g. Path.
}

// ReleaseAsset represents a release asset.
type ReleaseAsset struct {
	Name string
	URL  string
}
