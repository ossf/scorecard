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

import (
	"time"

	"github.com/ossf/scorecard/v4/clients"
)

// RawResults contains results before a policy
// is applied.
//nolint
type RawResults struct {
	CIIBestPracticesResults     CIIBestPracticesData
	DangerousWorkflowResults    DangerousWorkflowData
	VulnerabilitiesResults      VulnerabilitiesData
	BinaryArtifactResults       BinaryArtifactData
	SecurityPolicyResults       SecurityPolicyData
	DependencyUpdateToolResults DependencyUpdateToolData
	BranchProtectionResults     BranchProtectionsData
	CodeReviewResults           CodeReviewData
	WebhookResults              WebhooksData
	ContributorsResults         ContributorsData
	MaintainedResults           MaintainedData
	SignedReleasesResults       SignedReleasesData
	FuzzingResults              FuzzingData
	LicenseResults              LicenseData
}

// FuzzingData represents different fuzzing done.
type FuzzingData struct {
	Fuzzers []Tool
}

// MaintainedData contains the raw results
// for the Maintained check.
type MaintainedData struct {
	Issues               []Issue
	DefaultBranchCommits []DefaultBranchCommit
	ArchivedStatus       ArchivedStatus
}

// LicenseData contains the raw results
// for the License check.
type LicenseData struct {
	Files []File
}

// CodeReviewData contains the raw results
// for the Code-Review check.
type CodeReviewData struct {
	DefaultBranchCommits []DefaultBranchCommit
}

// ContributorsData represents contributor information.
type ContributorsData struct {
	Users []User
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
	Releases []clients.Release
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
	URL  *string
	Desc *string
	File *File
	Name string
	// Runs of the tool.
	Runs []Run
	// Issues created by the tool.
	Issues []Issue
	// Merge requests created by the tool.
	MergeRequests []MergeRequest

	// TODO: CodeCoverage, jsonWorkflowJob.
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
	State    string
	Reviewer User
	// TODO(Review): add fields here if needed.
}

// User represent a user.
type User struct {
	RepoAssociation *RepoAssociation
	Login           string
	// Orgnization refers to a GitHub org.
	Organizations []Organization
	// Companies refer to a claim by a user in their profile.
	Companies        []Company
	NumContributions uint
}

// Organization represents a GitHub organization.
type Organization struct {
	Login string
	// TODO: other info.
}

// Company represents a company in a user's profile.
type Company struct {
	Name string
	// TODO: other info.
}

// RepoAssociation represents a user relationship with a repo.
type RepoAssociation string

const (
	// RepoAssociationCollaborator has been invited to collaborate on the repository.
	RepoAssociationCollaborator RepoAssociation = "collaborator"
	// RepoAssociationContributor is an contributor to the repository.
	RepoAssociationContributor RepoAssociation = "contributor"
	// RepoAssociationOwner is an owner of the repository.
	RepoAssociationOwner RepoAssociation = "owner"
	// RepoAssociationMember is a member of the organization that owns the repository.
	RepoAssociationMember RepoAssociation = "member"
	// RepoAssociationFirstTimer has previously committed to the repository.
	RepoAssociationFirstTimer RepoAssociation = "first-timer"
	// RepoAssociationFirstTimeContributor has not previously committed to the repository.
	RepoAssociationFirstTimeContributor RepoAssociation = "first-timer-contributor"
	// RepoAssociationMannequin is a placeholder for an unclaimed user.
	RepoAssociationMannequin RepoAssociation = "unknown"
	// RepoAssociationNone has no association with the repository.
	RepoAssociationNone RepoAssociation = "none"
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

// CIIBestPracticesData contains data foor CIIBestPractices check.
type CIIBestPracticesData struct {
	Badge clients.BadgeLevel
}

// DangerousWorkflowType represents a type of dangerous workflow.
type DangerousWorkflowType string

const (
	// DangerousWorkflowScriptInjection represents a script injection.
	DangerousWorkflowScriptInjection DangerousWorkflowType = "scriptInjection"
	// DangerousWorkflowUntrustedCheckout represents an untrusted checkout.
	DangerousWorkflowUntrustedCheckout DangerousWorkflowType = "untrustedCheckout"
)

// DangerousWorkflowData contains raw results
// for dangerous workflow check.
type DangerousWorkflowData struct {
	Workflows []DangerousWorkflow
}

// DangerousWorkflow represents a dangerous workflow.
type DangerousWorkflow struct {
	Job  *WorkflowJob
	Type DangerousWorkflowType
	File File
}

// WorkflowJob reprresents a workflow job.
type WorkflowJob struct {
	Name *string
	ID   *string
}
