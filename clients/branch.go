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

// RefType represents the
// type of reference (branch or tag).
type RefType string

const (
	// RefTypeBranch represents a
	// branch reference.
	RefTypeBranch RefType = "branch"
	// RefTypeTag represents a
	// tag reference.
	RefTypeTag RefType = "tag"
)

// BranchRef represents a single branch reference and its protection rules.
type BranchRef struct {
	Name                 *string
	Protected            *bool
	BranchProtectionRule BranchProtectionRule
}

// TagRef represents a single tag
// reference and its protection rules.
type TagRef struct {
	Name              *string
	Protected         *bool
	TagProtectionRule TagProtectionRule
}

// RefProtectionRule captures
// the core protection settings
// that can apply to both branches
// and tags.
// This is the shared foundation
// for BranchProtectionRule and
// TagProtectionRule.
type RefProtectionRule struct {
	AllowDeletions       *bool
	AllowForcePushes     *bool
	RequireLinearHistory *bool
	EnforceAdmins        *bool
	CheckRules           StatusChecksRule
}

// BranchProtectionRule captures the settings enabled on a branch for security.
// It extends RefProtectionRule with
// branch-specific settings.
type BranchProtectionRule struct {
	PullRequestRule         PullRequestRule
	RequireLastPushApproval *bool
	RefProtectionRule
}

// TagProtectionRule captures
// the settings enabled on a
// tag for security.
// It extends RefProtectionRule with
// tag-specific settings.
type TagProtectionRule struct {
	AllowUpdates      *bool
	RequireSignatures *bool
	RestrictCreation  *bool
	RefProtectionRule
}

// StatusChecksRule captures settings on status checks.
type StatusChecksRule struct {
	UpToDateBeforeMerge  *bool
	RequiresStatusChecks *bool
	Contexts             []string
}

// PullRequestRule captures settings on a PullRequest.
type PullRequestRule struct {
	Required                     *bool // are PRs required
	RequiredApprovingReviewCount *int32
	DismissStaleReviews          *bool
	RequireCodeOwnerReviews      *bool
}
