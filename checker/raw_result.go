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

// RawResults contains results before a policy
// is applied.
type RawResults struct {
	BinaryArtifactResults       BinaryArtifactData
	SecurityPolicyResults       SecurityPolicyData
	DependencyUpdateToolResults DependencyUpdateToolData
	BranchProtectionResults     BranchProtectionsData
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

// DependencyUpdateToolData contains the raw results
// for the Dependency-Update-Tool check.
type DependencyUpdateToolData struct {
	// Tools contains a list of tools.
	// Note: we only populate one entry at most.
	Tools []Tool
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
	// Merges requests created by the tool.
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

// Issue represents an issue.
type Issue struct {
	URL string
	// TODO: add fields, e.g., state=[opened|closed]
}

// MergeRequest represents a merge request.
type MergeRequest struct {
	URL string
	// TODO: add fields, e.g., State=["merged"|"closed"]
}

// File represents a file.
type File struct {
	Path    string
	Snippet string   // Snippet of code
	Offset  uint     // Offset in the file of Path (line for source/text files).
	Type    FileType // Type of file.
	// TODO: add hash.
}
