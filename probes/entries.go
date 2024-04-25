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
	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/archived"
	"github.com/ossf/scorecard/v5/probes/blocksDeleteOnBranches"
	"github.com/ossf/scorecard/v5/probes/blocksForcePushOnBranches"
	"github.com/ossf/scorecard/v5/probes/branchProtectionAppliesToAdmins"
	"github.com/ossf/scorecard/v5/probes/branchesAreProtected"
	"github.com/ossf/scorecard/v5/probes/codeApproved"
	"github.com/ossf/scorecard/v5/probes/codeReviewOneReviewers"
	"github.com/ossf/scorecard/v5/probes/contributorsFromOrgOrCompany"
	"github.com/ossf/scorecard/v5/probes/createdRecently"
	"github.com/ossf/scorecard/v5/probes/dependencyUpdateToolConfigured"
	"github.com/ossf/scorecard/v5/probes/dismissesStaleReviews"
	"github.com/ossf/scorecard/v5/probes/fuzzed"
	"github.com/ossf/scorecard/v5/probes/hasBinaryArtifacts"
	"github.com/ossf/scorecard/v5/probes/hasDangerousWorkflowScriptInjection"
	"github.com/ossf/scorecard/v5/probes/hasDangerousWorkflowUntrustedCheckout"
	"github.com/ossf/scorecard/v5/probes/hasFSFOrOSIApprovedLicense"
	"github.com/ossf/scorecard/v5/probes/hasLicenseFile"
	"github.com/ossf/scorecard/v5/probes/hasNoGitHubWorkflowPermissionUnknown"
	"github.com/ossf/scorecard/v5/probes/hasOSVVulnerabilities"
	"github.com/ossf/scorecard/v5/probes/hasOpenSSFBadge"
	"github.com/ossf/scorecard/v5/probes/hasPermissiveLicense"
	"github.com/ossf/scorecard/v5/probes/hasRecentCommits"
	"github.com/ossf/scorecard/v5/probes/hasUnverifiedBinaryArtifacts"
	"github.com/ossf/scorecard/v5/probes/issueActivityByProjectMember"
	"github.com/ossf/scorecard/v5/probes/jobLevelPermissions"
	"github.com/ossf/scorecard/v5/probes/packagedWithAutomatedWorkflow"
	"github.com/ossf/scorecard/v5/probes/pinsDependencies"
	"github.com/ossf/scorecard/v5/probes/releasesAreSigned"
	"github.com/ossf/scorecard/v5/probes/releasesHaveProvenance"
	"github.com/ossf/scorecard/v5/probes/requiresApproversForPullRequests"
	"github.com/ossf/scorecard/v5/probes/requiresCodeOwnersReview"
	"github.com/ossf/scorecard/v5/probes/requiresLastPushApproval"
	"github.com/ossf/scorecard/v5/probes/requiresPRsToChangeCode"
	"github.com/ossf/scorecard/v5/probes/requiresUpToDateBranches"
	"github.com/ossf/scorecard/v5/probes/runsStatusChecksBeforeMerging"
	"github.com/ossf/scorecard/v5/probes/sastToolConfigured"
	"github.com/ossf/scorecard/v5/probes/sastToolRunsOnAllCommits"
	"github.com/ossf/scorecard/v5/probes/securityPolicyContainsLinks"
	"github.com/ossf/scorecard/v5/probes/securityPolicyContainsText"
	"github.com/ossf/scorecard/v5/probes/securityPolicyContainsVulnerabilityDisclosure"
	"github.com/ossf/scorecard/v5/probes/securityPolicyPresent"
	"github.com/ossf/scorecard/v5/probes/testsRunInCI"
	"github.com/ossf/scorecard/v5/probes/topLevelPermissions"
	"github.com/ossf/scorecard/v5/probes/webhooksUseSecrets"
)

// ProbeImpl is the implementation of a probe.
type ProbeImpl func(*checker.RawResults) ([]finding.Finding, string, error)

// IndependentProbeImpl is the implementation of an independent probe.
type IndependentProbeImpl func(*checker.CheckRequest) ([]finding.Finding, string, error)

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
		dependencyUpdateToolConfigured.Run,
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
	}
	Contributors = []ProbeImpl{
		contributorsFromOrgOrCompany.Run,
	}
	Vulnerabilities = []ProbeImpl{
		hasOSVVulnerabilities.Run,
	}
	CodeReview = []ProbeImpl{
		codeApproved.Run,
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
		archived.Run,
		hasRecentCommits.Run,
		issueActivityByProjectMember.Run,
		createdRecently.Run,
	}
	CIIBestPractices = []ProbeImpl{
		hasOpenSSFBadge.Run,
	}
	BinaryArtifacts = []ProbeImpl{
		hasUnverifiedBinaryArtifacts.Run,
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
		hasPermissiveLicense.Run,
		codeReviewOneReviewers.Run,
		hasBinaryArtifacts.Run,
	}

	// Probes which don't use pre-computed raw data but rather collect it themselves.
	Independent = []IndependentProbeImpl{}
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
