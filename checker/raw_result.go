// Copyright 2020 OpenSSF Scorecard Authors
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
	"fmt"
	"time"

	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/finding"
)

// RawResults contains results before a policy
// is applied.
// nolint
type RawResults struct {
	BinaryArtifactResults       BinaryArtifactData
	BranchProtectionResults     BranchProtectionsData
	CIIBestPracticesResults     CIIBestPracticesData
	CITestResults               CITestData
	CodeReviewResults           CodeReviewData
	ContributorsResults         ContributorsData
	DangerousWorkflowResults    DangerousWorkflowData
	DependencyUpdateToolResults DependencyUpdateToolData
	FuzzingResults              FuzzingData
	LicenseResults              LicenseData
	MaintainedResults           MaintainedData
	Metadata                    MetadataData
	PackagingResults            PackagingData
	PinningDependenciesResults  PinningDependenciesData
	SecurityPolicyResults       SecurityPolicyData
	SignedReleasesResults       SignedReleasesData
	TokenPermissionsResults     TokenPermissionsData
	VulnerabilitiesResults      VulnerabilitiesData
	WebhookResults              WebhooksData
}

type MetadataData struct {
	Metadata map[string]string
}

type RevisionCIInfo struct {
	HeadSHA           string
	CheckRuns         []clients.CheckRun
	Statuses          []clients.Status
	PullRequestNumber int
}

type CITestData struct {
	CIInfo []RevisionCIInfo
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
// nolint
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

// DependencyUseType represents a type of dependency use.
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
	// DependencyUseTypePipCommand is a pip command.
	DependencyUseTypePipCommand DependencyUseType = "pipCommand"
	// DependencyUseTypeNugetCommand is a nuget command.
	DependencyUseTypeNugetCommand DependencyUseType = "nugetCommand"
)

// PinningDependenciesData represents pinned dependency data.
type PinningDependenciesData struct {
	Dependencies     []Dependency
	ProcessingErrors []ElementError // jobs or files with errors may have incomplete results
}

// Dependency represents a dependency.
type Dependency struct {
	// TODO: unique dependency name.
	// TODO: Job         *WorkflowJob
	Name     *string
	PinnedAt *string
	Location *File
	Msg      *string // Only for debug messages.
	Pinned   *bool
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

type LicenseAttributionType string

const (
	// sources of license information used to assert repo's license.
	LicenseAttributionTypeOther      LicenseAttributionType = "other"
	LicenseAttributionTypeAPI        LicenseAttributionType = "repositoryAPI"
	LicenseAttributionTypeHeuristics LicenseAttributionType = "builtinHeuristics"
)

// license details.
type License struct {
	Name        string                 // OSI standardized license name
	SpdxID      string                 // SPDX standardized identifier
	Attribution LicenseAttributionType // source of licensing information
	Approved    bool                   // FSF or OSI Approved License
}

// one file contains one license.
type LicenseFile struct {
	LicenseInformation License
	File               File
}

// LicenseData contains the raw results
// for the License check.
// Some repos may have more than one license.
type LicenseData struct {
	LicenseFiles []LicenseFile
}

// CodeReviewData contains the raw results
// for the Code-Review check.
type CodeReviewData struct {
	DefaultBranchChangesets []Changeset
}
type ReviewPlatform = string

const (
	ReviewPlatformGitHub      ReviewPlatform = "GitHub"
	ReviewPlatformProw        ReviewPlatform = "Prow"
	ReviewPlatformGerrit      ReviewPlatform = "Gerrit"
	ReviewPlatformPhabricator ReviewPlatform = "Phabricator"
	ReviewPlatformPiper       ReviewPlatform = "Piper"
	ReviewPlatformUnknown     ReviewPlatform = "Unknown"
)

type Changeset struct {
	ReviewPlatform string
	RevisionID     string
	Commits        []clients.Commit
	Reviews        []clients.Review
	Author         clients.User
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

type SecurityPolicyInformationType string

const (
	// forms of security policy hints being evaluated.
	SecurityPolicyInformationTypeEmail SecurityPolicyInformationType = "emailAddress"
	SecurityPolicyInformationTypeLink  SecurityPolicyInformationType = "httpLink"
	SecurityPolicyInformationTypeText  SecurityPolicyInformationType = "vulnDisclosureText"
)

type SecurityPolicyValueType struct {
	Match      string // Snippet of match
	LineNumber uint   // Line number in policy file of match
	Offset     uint   // Offset in the line of the match
}

type SecurityPolicyInformation struct {
	InformationType  SecurityPolicyInformationType
	InformationValue SecurityPolicyValueType
}

type SecurityPolicyFile struct {
	// security policy information found in repo or org
	Information []SecurityPolicyInformation
	// file that contains the security policy information
	File File
}

// SecurityPolicyData contains the raw results
// for the Security-Policy check.
type SecurityPolicyData struct {
	PolicyFiles []SecurityPolicyFile
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
	Branches        []clients.BranchRef
	CodeownersFiles []string
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
}

// ArchivedStatus definess the archived status.
type ArchivedStatus struct {
	Status bool
	// TODO: add fields, e.g., date of archival.
}

// File represents a file.
type File struct {
	Path      string
	Snippet   string           // Snippet of code
	Offset    uint             // Offset in the file of Path (line for source/text files).
	EndOffset uint             // End of offset in the file, e.g. if the command spans multiple lines.
	FileSize  uint             // Total size of file.
	Type      finding.FileType // Type of file.
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
	Workflows    []DangerousWorkflow
	NumWorkflows int
}

// DangerousWorkflow represents a dangerous workflow.
type DangerousWorkflow struct {
	Job  *WorkflowJob
	Type DangerousWorkflowType
	File File
}

// WorkflowJob represents a workflow job.
type WorkflowJob struct {
	Name *string
	ID   *string
}

// TokenPermissionsData represents data about a permission failure.
type TokenPermissionsData struct {
	TokenPermissions []TokenPermission
	NumTokens        int
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
	// PermissionLevelUndeclared is an undeclared permission.
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

// Location generates location from a file.
func (f *File) Location() *finding.Location {
	// TODO(2626): merge location and path.
	if f == nil {
		return nil
	}
	loc := &finding.Location{
		Type:      f.Type,
		Path:      f.Path,
		LineStart: &f.Offset,
	}
	if f.EndOffset != 0 {
		loc.LineEnd = &f.EndOffset
	}
	if f.Snippet != "" {
		loc.Snippet = &f.Snippet
	}

	return loc
}

// ElementError allows us to identify the "element" that led to the given error.
// The "element" is the specific "code under analysis" that caused the error. It should
// describe what caused the error as precisely as possible.
//
// For example, if a shell parsing error occurs while parsing a Dockerfile `RUN` block
// or a GitHub workflow's `run:` step, the "element" should point to the Dockerfile
// lines or workflow job step that caused the failure, not just the file path.
type ElementError struct {
	Err     error
	Element *finding.Location
}

func (e *ElementError) Error() string {
	return fmt.Sprintf("%s: %v", e.Err, *e.Element)
}

func (e *ElementError) Unwrap() error {
	return e.Err
}
