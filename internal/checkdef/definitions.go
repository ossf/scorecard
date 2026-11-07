// Copyright 2026 OpenSSF Scorecard Authors
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

package checkdef

import (
	"errors"
	"fmt"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/checks/evaluation"
	"github.com/ossf/scorecard/v5/checks/raw"
	"github.com/ossf/scorecard/v5/checks/raw/github"
	"github.com/ossf/scorecard/v5/checks/raw/gitlab"
	"github.com/ossf/scorecard/v5/clients/githubrepo"
	"github.com/ossf/scorecard/v5/clients/gitlabrepo"
	"github.com/ossf/scorecard/v5/clients/localdir"
	"github.com/ossf/scorecard/v5/internal/checknames"
	"github.com/ossf/scorecard/v5/probes"
)

// BinaryArtifacts is the Binary-Artifacts check.
var BinaryArtifacts = define(
	checknames.BinaryArtifacts,
	raw.BinaryArtifacts,
	func(r *checker.RawResults, data checker.BinaryArtifactData) { r.BinaryArtifactResults = data },
	probes.BinaryArtifacts,
	evaluation.BinaryArtifacts,
).supports(checker.CommitBased, checker.FileBased)

// BranchProtection is the Branch-Protection check.
var BranchProtection = define(
	checknames.BranchProtection,
	raw.BranchProtection,
	func(r *checker.RawResults, data checker.BranchProtectionsData) { r.BranchProtectionResults = data },
	probes.BranchProtection,
	evaluation.BranchProtection,
)

// CIIBestPractices is the CII-Best-Practices check.
var CIIBestPractices = define(
	checknames.CIIBestPractices,
	raw.CIIBestPractices,
	func(r *checker.RawResults, data checker.CIIBestPracticesData) { r.CIIBestPracticesResults = data },
	probes.CIIBestPractices,
	evaluation.CIIBestPractices,
)

// CITests is the CI-Tests check.
var CITests = define(
	checknames.CITests,
	func(c *checker.CheckRequest) (checker.CITestData, error) { return raw.CITests(c.RepoClient) },
	func(r *checker.RawResults, data checker.CITestData) { r.CITestResults = data },
	probes.CITests,
	evaluation.CITests,
).supports(checker.CommitBased)

// CodeReview is the Code-Review check.
var CodeReview = define(
	checknames.CodeReview,
	func(c *checker.CheckRequest) (checker.CodeReviewData, error) { return raw.CodeReview(c.RepoClient) },
	func(r *checker.RawResults, data checker.CodeReviewData) { r.CodeReviewResults = data },
	probes.CodeReview,
	evaluation.CodeReview,
).supports(checker.CommitBased)

// Contributors is the Contributors check.
var Contributors = define(
	checknames.Contributors,
	raw.Contributors,
	func(r *checker.RawResults, data checker.ContributorsData) { r.ContributorsResults = data },
	probes.Contributors,
	evaluation.Contributors,
)

// DangerousWorkflow is the Dangerous-Workflow check.
var DangerousWorkflow = define(
	checknames.DangerousWorkflow,
	raw.DangerousWorkflow,
	func(r *checker.RawResults, data checker.DangerousWorkflowData) { r.DangerousWorkflowResults = data },
	probes.DangerousWorkflows,
	evaluation.DangerousWorkflow,
).supports(checker.CommitBased, checker.FileBased)

// DependencyUpdateTool is the Dependency-Update-Tool check.
var DependencyUpdateTool = define(
	checknames.DependencyUpdateTool,
	func(c *checker.CheckRequest) (checker.DependencyUpdateToolData, error) {
		return raw.DependencyUpdateTool(c.RepoClient)
	},
	func(r *checker.RawResults, data checker.DependencyUpdateToolData) {
		r.DependencyUpdateToolResults = data
	},
	probes.DependencyToolUpdates,
	evaluation.DependencyUpdateTool,
).supports(checker.FileBased)

// Fuzzing is the Fuzzing check.
var Fuzzing = define(
	checknames.Fuzzing,
	raw.Fuzzing,
	func(r *checker.RawResults, data checker.FuzzingData) { r.FuzzingResults = data },
	probes.Fuzzing,
	evaluation.Fuzzing,
).supports(checker.FileBased)

// License is the License check.
var License = define(
	checknames.License,
	raw.License,
	func(r *checker.RawResults, data checker.LicenseData) { r.LicenseResults = data },
	probes.License,
	evaluation.License,
).supports(checker.CommitBased, checker.FileBased)

// Maintained is the Maintained check.
var Maintained = define(
	checknames.Maintained,
	raw.Maintained,
	func(r *checker.RawResults, data checker.MaintainedData) { r.MaintainedResults = data },
	probes.Maintained,
	evaluation.Maintained,
)

// Packaging is the Packaging check. Probe mode collects only from GitHub
// and GitLab clients.
var Packaging = withProbeMode(
	define(
		checknames.Packaging,
		collectPackaging,
		storePackaging,
		probes.Packaging,
		evaluation.Packaging,
	),
	collectPackagingProbeMode,
	storePackaging,
).supports(checker.FileBased)

func storePackaging(r *checker.RawResults, data checker.PackagingData) { r.PackagingResults = data }

var errPackagingProbeModeClient = errors.New("unsupported repository client")

// collectPackagingProbeMode picks the collector by client type and
// returns an error for any other client.
func collectPackagingProbeMode(c *checker.CheckRequest) (checker.PackagingData, error) {
	var data checker.PackagingData
	var err error
	switch c.RepoClient.(type) {
	case *githubrepo.Client:
		data, err = github.Packaging(c)
	case *gitlabrepo.Client:
		data, err = gitlab.Packaging(c)
	default:
		return data, errPackagingProbeModeClient
	}
	if err != nil {
		// CollectRawData adds the context, so keep the collector's error text.
		return data, fmt.Errorf("%w", err)
	}
	return data, nil
}

// collectPackaging picks the collector by client type. A local directory
// may hold either platform's workflows, so it runs both and reports the
// GitHub error first. Other clients get no packaging data.
func collectPackaging(c *checker.CheckRequest) (checker.PackagingData, error) {
	var data checker.PackagingData
	var err error
	switch c.RepoClient.(type) {
	case *localdir.Client:
		dataGithub, errGithub := github.Packaging(c)
		dataGitlab, errGitlab := gitlab.Packaging(c)
		data.Packages = append(data.Packages, dataGithub.Packages...)
		data.Packages = append(data.Packages, dataGitlab.Packages...)
		if errGithub != nil {
			err = errGithub
		} else {
			err = errGitlab
		}
	case *githubrepo.Client:
		data, err = github.Packaging(c)
	case *gitlabrepo.Client:
		data, err = gitlab.Packaging(c)
	}
	if err != nil {
		// Run adds the context, so keep the collector's error text.
		return data, fmt.Errorf("%w", err)
	}
	return data, nil
}

// PinnedDependencies is the Pinned-Dependencies check.
var PinnedDependencies = define(
	checknames.PinnedDependencies,
	raw.PinningDependencies,
	func(r *checker.RawResults, data checker.PinningDependenciesData) { r.PinningDependenciesResults = data },
	probes.PinnedDependencies,
	evaluation.PinningDependencies,
).supports(checker.CommitBased, checker.FileBased)

// SAST is the SAST check.
var SAST = define(
	checknames.SAST,
	raw.SAST,
	func(r *checker.RawResults, data checker.SASTData) { r.SASTResults = data },
	probes.SAST,
	evaluation.SAST,
).supports(checker.FileBased)

// SBOM is the SBOM check. It runs only when SCORECARD_EXPERIMENTAL is set.
var SBOM = define(
	checknames.SBOM,
	raw.SBOM,
	func(r *checker.RawResults, data checker.SBOMData) { r.SBOMResults = data },
	probes.SBOM,
	evaluation.SBOM,
).experimental("SBOM")

// SecurityPolicy is the Security-Policy check.
var SecurityPolicy = define(
	checknames.SecurityPolicy,
	raw.SecurityPolicy,
	func(r *checker.RawResults, data checker.SecurityPolicyData) { r.SecurityPolicyResults = data },
	probes.SecurityPolicy,
	evaluation.SecurityPolicy,
).supports(checker.CommitBased, checker.FileBased)

// SignedReleases is the Signed-Releases check.
var SignedReleases = define(
	checknames.SignedReleases,
	raw.SignedReleases,
	func(r *checker.RawResults, data checker.SignedReleasesData) { r.SignedReleasesResults = data },
	probes.SignedReleases,
	evaluation.SignedReleases,
)

// TokenPermissions is the Token-Permissions check.
var TokenPermissions = define(
	checknames.TokenPermissions,
	raw.TokenPermissions,
	func(r *checker.RawResults, data checker.TokenPermissionsData) { r.TokenPermissionsResults = data },
	probes.TokenPermissions,
	evaluation.TokenPermissions,
).supports(checker.CommitBased, checker.FileBased)

// Vulnerabilities is the Vulnerabilities check.
var Vulnerabilities = define(
	checknames.Vulnerabilities,
	raw.Vulnerabilities,
	func(r *checker.RawResults, data checker.VulnerabilitiesData) { r.VulnerabilitiesResults = data },
	probes.Vulnerabilities,
	evaluation.Vulnerabilities,
).supports(checker.CommitBased, checker.FileBased)

// Webhooks is the Webhooks check. It runs only when SCORECARD_EXPERIMENTAL
// is set.
var Webhooks = define(
	checknames.Webhooks,
	raw.WebHook,
	func(r *checker.RawResults, data checker.WebhooksData) { r.WebhookResults = data },
	probes.Webhook,
	evaluation.Webhooks,
).experimental("Webhook")

// byName looks up a definition by check name for probe mode.
var byName = func() map[string]*Definition {
	all := []*Definition{
		BinaryArtifacts,
		BranchProtection,
		CIIBestPractices,
		CITests,
		CodeReview,
		Contributors,
		DangerousWorkflow,
		DependencyUpdateTool,
		Fuzzing,
		License,
		Maintained,
		Packaging,
		PinnedDependencies,
		SAST,
		SBOM,
		SecurityPolicy,
		SignedReleases,
		TokenPermissions,
		Vulnerabilities,
		Webhooks,
	}
	m := make(map[string]*Definition, len(all))
	for _, d := range all {
		m[d.name] = d
	}
	return m
}()
