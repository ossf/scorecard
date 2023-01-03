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

// BranchRef represents a single branch reference and its protection rules.
type BranchRef struct {
	Name                 *string
	Protected            *bool
	BranchProtectionRule BranchProtectionRule
}

// BranchProtectionRule captures the settings enabled on a branch for security.
type BranchProtectionRule struct {
	RequiredPullRequestReviews PullRequestReviewRule
	AllowDeletions             *bool
	AllowForcePushes           *bool
	RequireLinearHistory       *bool
	EnforceAdmins              *bool
	RequireLastPushApproval    *bool
	CheckRules                 StatusChecksRule
}

// StatusChecksRule captures settings on status checks.
type StatusChecksRule struct {
	UpToDateBeforeMerge  *bool
	RequiresStatusChecks *bool
	Contexts             []string
}

// PullRequestReviewRule captures settings on a PullRequest.
type PullRequestReviewRule struct {
	RequiredApprovingReviewCount *int32
	DismissStaleReviews          *bool
	RequireCodeOwnerReviews      *bool
}
