// Copyright 2023 OpenSSF Scorecard Authors
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

package probes

import (
	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/blocksDeleteOnBranches"
	"github.com/ossf/scorecard/v4/probes/blocksForcePushOnBranches"
	"github.com/ossf/scorecard/v4/probes/branchProtectionAppliesToAdmins"
	"github.com/ossf/scorecard/v4/probes/branchesAreProtected"
	"github.com/ossf/scorecard/v4/probes/codeApproved"
	"github.com/ossf/scorecard/v4/probes/codeReviewOneReviewers"
	"github.com/ossf/scorecard/v4/probes/contributorsFromOrgOrCompany"
	"github.com/ossf/scorecard/v4/probes/dismissesStaleReviews"
	"github.com/ossf/scorecard/v4/probes/freeOfAnyBinaryArtifacts"
	"github.com/ossf/scorecard/v4/probes/freeOfUnverifiedBinaryArtifacts"
	"github.com/ossf/scorecard/v4/probes/fuzzed"
	"github.com/ossf/scorecard/v4/probes/hasDangerousWorkflowScriptInjection"
	"github.com/ossf/scorecard/v4/probes/hasDangerousWorkflowUntrustedCheckout"
	"github.com/ossf/scorecard/v4/probes/hasFSFOrOSIApprovedLicense"
	"github.com/ossf/scorecard/v4/probes/hasLicenseFile"
	"github.com/ossf/scorecard/v4/probes/hasLicenseFileAtTopDir"
	"github.com/ossf/scorecard/v4/probes/hasNoGitHubWorkflowPermissionUnknown"
	"github.com/ossf/scorecard/v4/probes/hasOSVVulnerabilities"
	"github.com/ossf/scorecard/v4/probes/hasOpenSSFBadge"
	"github.com/ossf/scorecard/v4/probes/hasRecentCommits"
	"github.com/ossf/scorecard/v4/probes/issueActivityByProjectMember"
	"github.com/ossf/scorecard/v4/probes/jobLevelPermissions"
	"github.com/ossf/scorecard/v4/probes/notArchived"
	"github.com/ossf/scorecard/v4/probes/notCreatedRecently"
	"github.com/ossf/scorecard/v4/probes/packagedWithAutomatedWorkflow"
	"github.com/ossf/scorecard/v4/probes/pinsDependencies"
	"github.com/ossf/scorecard/v4/probes/releasesAreSigned"
	"github.com/ossf/scorecard/v4/probes/releasesHaveProvenance"
	"github.com/ossf/scorecard/v4/probes/requiresApproversForPullRequests"
	"github.com/ossf/scorecard/v4/probes/requiresCodeOwnersReview"
	"github.com/ossf/scorecard/v4/probes/requiresLastPushApproval"
	"github.com/ossf/scorecard/v4/probes/requiresPRsToChangeCode"
	"github.com/ossf/scorecard/v4/probes/requiresUpToDateBranches"
	"github.com/ossf/scorecard/v4/probes/runsStatusChecksBeforeMerging"
	"github.com/ossf/scorecard/v4/probes/sastToolConfigured"
	"github.com/ossf/scorecard/v4/probes/sastToolRunsOnAllCommits"
	"github.com/ossf/scorecard/v4/probes/securityPolicyContainsLinks"
	"github.com/ossf/scorecard/v4/probes/securityPolicyContainsText"
	"github.com/ossf/scorecard/v4/probes/securityPolicyContainsVulnerabilityDisclosure"
	"github.com/ossf/scorecard/v4/probes/securityPolicyPresent"
	"github.com/ossf/scorecard/v4/probes/testsRunInCI"
	"github.com/ossf/scorecard/v4/probes/toolDependabotInstalled"
	"github.com/ossf/scorecard/v4/probes/toolPyUpInstalled"
	"github.com/ossf/scorecard/v4/probes/toolRenovateInstalled"
	"github.com/ossf/scorecard/v4/probes/topLevelPermissions"
	"github.com/ossf/scorecard/v4/probes/webhooksUseSecrets"
)

// ProbeImpl is the implementation of a probe.
type ProbeImpl func(*checker.RawResults) ([]finding.Finding, string, error)

var (
	// All represents all the probes.
	All []ProbeImpl
	// SecurityPolicy is all the probes for the
	// SecurityPolicy check.
	SecurityPolicy = []ProbeImpl{
		securityPolicyPresent.Run,
		securityPolicyContainsLinks.Run,
		securityPolicyContainsVulnerabilityDisclosure.Run,
		securityPolicyContainsText.Run,
	}
	// DependencyToolUpdates is all the probes for the
	// DependencyUpdateTool check.
	DependencyToolUpdates = []ProbeImpl{
		toolRenovateInstalled.Run,
		toolDependabotInstalled.Run,
		toolPyUpInstalled.Run,
	}
	Fuzzing = []ProbeImpl{
		fuzzed.Run,
	}
	Packaging = []ProbeImpl{
		packagedWithAutomatedWorkflow.Run,
	}
	License = []ProbeImpl{
		hasLicenseFile.Run,
		hasFSFOrOSIApprovedLicense.Run,
		hasLicenseFileAtTopDir.Run,
	}
	Contributors = []ProbeImpl{
		contributorsFromOrgOrCompany.Run,
	}
	Vulnerabilities = []ProbeImpl{
		hasOSVVulnerabilities.Run,
	}
	CodeReview = []ProbeImpl{
		codeApproved.Run,
		codeReviewOneReviewers.Run,
	}
	SAST = []ProbeImpl{
		sastToolConfigured.Run,
		sastToolRunsOnAllCommits.Run,
	}
	DangerousWorkflows = []ProbeImpl{
		hasDangerousWorkflowScriptInjection.Run,
		hasDangerousWorkflowUntrustedCheckout.Run,
	}
	Maintained = []ProbeImpl{
		notArchived.Run,
		hasRecentCommits.Run,
		issueActivityByProjectMember.Run,
		notCreatedRecently.Run,
	}
	CIIBestPractices = []ProbeImpl{
		hasOpenSSFBadge.Run,
	}
	BinaryArtifacts = []ProbeImpl{
		freeOfUnverifiedBinaryArtifacts.Run,
	}
	Webhook = []ProbeImpl{
		webhooksUseSecrets.Run,
	}
	CITests = []ProbeImpl{
		testsRunInCI.Run,
	}
	SignedReleases = []ProbeImpl{
		releasesAreSigned.Run,
		releasesHaveProvenance.Run,
	}
	BranchProtection = []ProbeImpl{
		blocksDeleteOnBranches.Run,
		blocksForcePushOnBranches.Run,
		branchesAreProtected.Run,
		branchProtectionAppliesToAdmins.Run,
		dismissesStaleReviews.Run,
		requiresApproversForPullRequests.Run,
		requiresCodeOwnersReview.Run,
		requiresLastPushApproval.Run,
		requiresUpToDateBranches.Run,
		runsStatusChecksBeforeMerging.Run,
		requiresPRsToChangeCode.Run,
	}
	PinnedDependencies = []ProbeImpl{
		pinsDependencies.Run,
	}
	TokenPermissions = []ProbeImpl{
		hasNoGitHubWorkflowPermissionUnknown.Run,
		jobLevelPermissions.Run,
		topLevelPermissions.Run,
	}

	// Probes which aren't included by any checks.
	// These still need to be listed so they can be called with --probes.
	Uncategorized = []ProbeImpl{
		freeOfAnyBinaryArtifacts.Run,
	}
)

//nolint:gochecknoinits
func init() {
	All = concatMultipleProbes([][]ProbeImpl{
		BinaryArtifacts,
		CIIBestPractices,
		CITests,
		CodeReview,
		Contributors,
		DangerousWorkflows,
		DependencyToolUpdates,
		Fuzzing,
		License,
		Maintained,
		Packaging,
		SAST,
		SecurityPolicy,
		SignedReleases,
		Uncategorized,
		Vulnerabilities,
		Webhook,
	})
}

func concatMultipleProbes(slices [][]ProbeImpl) []ProbeImpl {
	var totalLen int
	for _, s := range slices {
		totalLen += len(s)
	}
	tmp := make([]ProbeImpl, 0, totalLen)
	for _, s := range slices {
		tmp = append(tmp, s...)
	}
	return tmp
}
