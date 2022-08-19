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
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ossf/scorecard/v4/clients"
)

var errInvalidArg = errors.New("invalid argument")

// RawResults contains results before a policy
// is applied.
//
//nolint:govet
type RawResults struct {
	// only
	metadataOnce        sync.Once
	RemediationMetadata RemediationMetadata

	PackagingResults            PackagingData
	CIIBestPracticesResults     CIIBestPracticesData
	DangerousWorkflowResults    DangerousWorkflowData
	VulnerabilitiesResults      VulnerabilitiesData
	BinaryArtifactResults       BinaryArtifactData
	SecurityPolicyResults       SecurityPolicyData
	DependencyUpdateToolResults DependencyUpdateToolData
	BranchProtectionResults     BranchProtectionsData
	CodeReviewResults           CodeReviewData
	PinningDependenciesResults  PinningDependenciesData
	WebhookResults              WebhooksData
	ContributorsResults         ContributorsData
	MaintainedResults           MaintainedData
	SignedReleasesResults       SignedReleasesData
	FuzzingResults              FuzzingData
	LicenseResults              LicenseData
	TokenPermissionsResults     TokenPermissionsData
}

// FuzzingData represents different fuzzing done.
type FuzzingData struct {
	Fuzzers []Tool
}

// TODO: Add Msg to all results.

// PackagingData contains results for the Packaging check.
type PackagingData struct {
	Packages []Package
}

// Package represents a package.
type Package struct {
	// TODO: not supported yet. This needs to be unique across
	// ecosystems: purl, OSV, CPE, etc.
	Name *string
	Job  *WorkflowJob
	File *File
	// Note: Msg is populated only for debug messages.
	Msg  *string
	Runs []Run
}

// DependencyUseType reprensets a type of dependency use.
type DependencyUseType string

const (
	// DependencyUseTypeGHAction is an action.
	DependencyUseTypeGHAction DependencyUseType = "GitHubAction"
	// DependencyUseTypeDockerfileContainerImage a container image used via FROM.
	DependencyUseTypeDockerfileContainerImage DependencyUseType = "containerImage"
	// DependencyUseTypeDownloadThenRun is a download followed by a run.
	DependencyUseTypeDownloadThenRun DependencyUseType = "downloadThenRun"
	// DependencyUseTypeGoCommand is a go command.
	DependencyUseTypeGoCommand DependencyUseType = "goCommand"
	// DependencyUseTypeChocoCommand is a choco command.
	DependencyUseTypeChocoCommand DependencyUseType = "chocoCommand"
	// DependencyUseTypeNpmCommand is an npm command.
	DependencyUseTypeNpmCommand DependencyUseType = "npmCommand"
	// DependencyUseTypePipCommand is a pipp command.
	DependencyUseTypePipCommand DependencyUseType = "pipCommand"
)

// PinningDependenciesData represents pinned dependency data.
type PinningDependenciesData struct {
	Dependencies []Dependency
}

// Dependency represents a dependency.
type Dependency struct {
	// TODO: unique dependency name.
	// TODO: Job         *WorkflowJob
	Name     *string
	PinnedAt *string
	Location *File
	Msg      *string // Only for debug messages.
	Type     DependencyUseType
}

// MaintainedData contains the raw results
// for the Maintained check.
type MaintainedData struct {
	CreatedAt            time.Time
	Issues               []clients.Issue
	DefaultBranchCommits []clients.Commit
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
	DefaultBranchCommits []clients.Commit
}

// ContributorsData represents contributor information.
type ContributorsData struct {
	Users []clients.User
}

// VulnerabilitiesData contains the raw results
// for the Vulnerabilities check.
type VulnerabilitiesData struct {
	Vulnerabilities []clients.Vulnerability
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
	Webhooks []clients.Webhook
}

// BranchProtectionsData contains the raw results
// for the Branch-Protection check.
type BranchProtectionsData struct {
	Branches []clients.BranchRef
}

// Tool represents a tool.
type Tool struct {
	URL   *string
	Desc  *string
	Files []File
	Name  string
	// Runs of the tool.
	Runs []Run
	// Issues created by the tool.
	Issues []clients.Issue
	// Merge requests created by the tool.
	MergeRequests []clients.PullRequest

	// TODO: CodeCoverage, jsonWorkflowJob.
}

// Run represents a run.
type Run struct {
	URL string
	// TODO: add fields, e.g., Result=["success", "failure"]
}

// ArchivedStatus definess the archived status.
type ArchivedStatus struct {
	Status bool
	// TODO: add fields, e.g., date of archival.
}

// File represents a file.
type File struct {
	Path      string
	Snippet   string   // Snippet of code
	Offset    uint     // Offset in the file of Path (line for source/text files).
	EndOffset uint     // End of offset in the file, e.g. if the command spans multiple lines.
	Type      FileType // Type of file.
	// TODO: add hash.
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

// TokenPermissionsData represents data about a permission failure.
type TokenPermissionsData struct {
	TokenPermissions []TokenPermission
}

// PermissionLocation represents a declaration type.
type PermissionLocation string

const (
	// PermissionLocationTop is top-level workflow permission.
	PermissionLocationTop PermissionLocation = "topLevel"
	// PermissionLocationJob is job-level workflow permission.
	PermissionLocationJob PermissionLocation = "jobLevel"
)

// PermissionLevel represents a permission type.
type PermissionLevel string

const (
	// PermissionLevelUndeclared is an undecleared permission.
	PermissionLevelUndeclared PermissionLevel = "undeclared"
	// PermissionLevelWrite is a permission set to `write` for a permission we consider potentially dangerous.
	PermissionLevelWrite PermissionLevel = "write"
	// PermissionLevelRead is a permission set to `read`.
	PermissionLevelRead PermissionLevel = "read"
	// PermissionLevelNone is a permission set to `none`.
	PermissionLevelNone PermissionLevel = "none"
	// PermissionLevelUnknown is for other kinds of alerts, mostly to support debug messages.
	// TODO: remove it once we have implemented severity (#1874).
	PermissionLevelUnknown PermissionLevel = "unknown"
)

// TokenPermission defines a token permission result.
type TokenPermission struct {
	Job          *WorkflowJob
	LocationType *PermissionLocation
	Name         *string
	Value        *string
	File         *File
	Msg          *string
	Type         PermissionLevel
}

type RemediationMetadata struct {
	Branch string
	Repo   string
}

// remediationMetadata returns remediation relevant metadata from a CheckRequest.
func (r *RawResults) SetupRemediationMetadata(c *CheckRequest) error {
	var e error
	r.metadataOnce.Do(func() {
		if r == nil || c.RepoClient == nil {
			return
		}
		r.RemediationMetadata = RemediationMetadata{}

		// Get the branch for remediation.
		branch, err := c.RepoClient.GetDefaultBranchName()
		if err != nil {
			e = fmt.Errorf("GetDefaultBranchName: %w", err)
			return
		}

		uri := c.RepoClient.URI()
		parts := strings.Split(uri, "/")

		if len(parts) != 3 {
			e = fmt.Errorf("%w: empty: %s", errInvalidArg, uri)
			return
		}
		repo := fmt.Sprintf("%s/%s", parts[1], parts[2])
		r.RemediationMetadata = RemediationMetadata{Branch: branch, Repo: repo}
	})

	return e
}
