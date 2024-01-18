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
	"errors"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/contributorsFromOrgOrCompany"
	"github.com/ossf/scorecard/v4/probes/freeOfUnverifiedBinaryArtifacts"
	"github.com/ossf/scorecard/v4/probes/fuzzedWithCLibFuzzer"
	"github.com/ossf/scorecard/v4/probes/fuzzedWithClusterFuzzLite"
	"github.com/ossf/scorecard/v4/probes/fuzzedWithCppLibFuzzer"
	"github.com/ossf/scorecard/v4/probes/fuzzedWithGoNative"
	"github.com/ossf/scorecard/v4/probes/fuzzedWithJavaJazzerFuzzer"
	"github.com/ossf/scorecard/v4/probes/fuzzedWithOSSFuzz"
	"github.com/ossf/scorecard/v4/probes/fuzzedWithPropertyBasedHaskell"
	"github.com/ossf/scorecard/v4/probes/fuzzedWithPropertyBasedJavascript"
	"github.com/ossf/scorecard/v4/probes/fuzzedWithPropertyBasedTypescript"
	"github.com/ossf/scorecard/v4/probes/fuzzedWithPythonAtheris"
	"github.com/ossf/scorecard/v4/probes/fuzzedWithRustCargofuzz"
	"github.com/ossf/scorecard/v4/probes/fuzzedWithSwiftLibFuzzer"
	"github.com/ossf/scorecard/v4/probes/hasDangerousWorkflowScriptInjection"
	"github.com/ossf/scorecard/v4/probes/hasDangerousWorkflowUntrustedCheckout"
	"github.com/ossf/scorecard/v4/probes/hasFSFOrOSIApprovedLicense"
	"github.com/ossf/scorecard/v4/probes/hasLicenseFile"
	"github.com/ossf/scorecard/v4/probes/hasLicenseFileAtTopDir"
	"github.com/ossf/scorecard/v4/probes/hasOSVVulnerabilities"
	"github.com/ossf/scorecard/v4/probes/hasOpenSSFBadge"
	"github.com/ossf/scorecard/v4/probes/hasRecentCommits"
	"github.com/ossf/scorecard/v4/probes/issueActivityByProjectMember"
	"github.com/ossf/scorecard/v4/probes/notArchived"
	"github.com/ossf/scorecard/v4/probes/notCreatedRecently"
	"github.com/ossf/scorecard/v4/probes/packagedWithAutomatedWorkflow"
	"github.com/ossf/scorecard/v4/probes/releasesAreSigned"
	"github.com/ossf/scorecard/v4/probes/releasesHaveProvenance"
	"github.com/ossf/scorecard/v4/probes/sastToolCodeQLInstalled"
	"github.com/ossf/scorecard/v4/probes/sastToolPysaInstalled"
	"github.com/ossf/scorecard/v4/probes/sastToolQodanaInstalled"
	"github.com/ossf/scorecard/v4/probes/sastToolRunsOnAllCommits"
	"github.com/ossf/scorecard/v4/probes/sastToolSnykInstalled"
	"github.com/ossf/scorecard/v4/probes/sastToolSonarInstalled"
	"github.com/ossf/scorecard/v4/probes/securityPolicyContainsLinks"
	"github.com/ossf/scorecard/v4/probes/securityPolicyContainsText"
	"github.com/ossf/scorecard/v4/probes/securityPolicyContainsVulnerabilityDisclosure"
	"github.com/ossf/scorecard/v4/probes/securityPolicyPresent"
	"github.com/ossf/scorecard/v4/probes/testsRunInCI"
	"github.com/ossf/scorecard/v4/probes/toolDependabotInstalled"
	"github.com/ossf/scorecard/v4/probes/toolPyUpInstalled"
	"github.com/ossf/scorecard/v4/probes/toolRenovateInstalled"
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
		fuzzedWithOSSFuzz.Run,
		fuzzedWithGoNative.Run,
		fuzzedWithPythonAtheris.Run,
		fuzzedWithCLibFuzzer.Run,
		fuzzedWithCppLibFuzzer.Run,
		fuzzedWithSwiftLibFuzzer.Run,
		fuzzedWithRustCargofuzz.Run,
		fuzzedWithJavaJazzerFuzzer.Run,
		fuzzedWithClusterFuzzLite.Run,
		fuzzedWithPropertyBasedHaskell.Run,
		fuzzedWithPropertyBasedTypescript.Run,
		fuzzedWithPropertyBasedJavascript.Run,
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
	SAST = []ProbeImpl{
		sastToolCodeQLInstalled.Run,
		sastToolPysaInstalled.Run,
		sastToolQodanaInstalled.Run,
		sastToolSnykInstalled.Run,
		sastToolRunsOnAllCommits.Run,
		sastToolSonarInstalled.Run,
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

	probeRunners = map[string]func(*checker.RawResults) ([]finding.Finding, string, error){
		securityPolicyPresent.Probe:                         securityPolicyPresent.Run,
		securityPolicyContainsLinks.Probe:                   securityPolicyContainsLinks.Run,
		securityPolicyContainsVulnerabilityDisclosure.Probe: securityPolicyContainsVulnerabilityDisclosure.Run,
		securityPolicyContainsText.Probe:                    securityPolicyContainsText.Run,
		toolRenovateInstalled.Probe:                         toolRenovateInstalled.Run,
		toolDependabotInstalled.Probe:                       toolDependabotInstalled.Run,
		toolPyUpInstalled.Probe:                             toolPyUpInstalled.Run,
		fuzzedWithOSSFuzz.Probe:                             fuzzedWithOSSFuzz.Run,
		fuzzedWithGoNative.Probe:                            fuzzedWithGoNative.Run,
		fuzzedWithPythonAtheris.Probe:                       fuzzedWithPythonAtheris.Run,
		fuzzedWithCLibFuzzer.Probe:                          fuzzedWithCLibFuzzer.Run,
		fuzzedWithCppLibFuzzer.Probe:                        fuzzedWithCppLibFuzzer.Run,
		fuzzedWithSwiftLibFuzzer.Probe:                      fuzzedWithSwiftLibFuzzer.Run,
		fuzzedWithRustCargofuzz.Probe:                       fuzzedWithRustCargofuzz.Run,
		fuzzedWithJavaJazzerFuzzer.Probe:                    fuzzedWithJavaJazzerFuzzer.Run,
		fuzzedWithClusterFuzzLite.Probe:                     fuzzedWithClusterFuzzLite.Run,
		fuzzedWithPropertyBasedHaskell.Probe:                fuzzedWithPropertyBasedHaskell.Run,
		fuzzedWithPropertyBasedTypescript.Probe:             fuzzedWithPropertyBasedTypescript.Run,
		fuzzedWithPropertyBasedJavascript.Probe:             fuzzedWithPropertyBasedJavascript.Run,
		packagedWithAutomatedWorkflow.Probe:                 packagedWithAutomatedWorkflow.Run,
		hasLicenseFile.Probe:                                hasLicenseFile.Run,
		hasFSFOrOSIApprovedLicense.Probe:                    hasFSFOrOSIApprovedLicense.Run,
		hasLicenseFileAtTopDir.Probe:                        hasLicenseFileAtTopDir.Run,
		contributorsFromOrgOrCompany.Probe:                  contributorsFromOrgOrCompany.Run,
		hasOSVVulnerabilities.Probe:                         hasOSVVulnerabilities.Run,
		sastToolCodeQLInstalled.Probe:                       sastToolCodeQLInstalled.Run,
		sastToolRunsOnAllCommits.Probe:                      sastToolRunsOnAllCommits.Run,
		sastToolSonarInstalled.Probe:                        sastToolSonarInstalled.Run,
		hasDangerousWorkflowScriptInjection.Probe:           hasDangerousWorkflowScriptInjection.Run,
		hasDangerousWorkflowUntrustedCheckout.Probe:         hasDangerousWorkflowUntrustedCheckout.Run,
		notArchived.Probe:                                   notArchived.Run,
		hasRecentCommits.Probe:                              hasRecentCommits.Run,
		issueActivityByProjectMember.Probe:                  issueActivityByProjectMember.Run,
		notCreatedRecently.Probe:                            notCreatedRecently.Run,
	}

	CheckMap = map[string]string{
		securityPolicyPresent.Probe:                         "Security-Policy",
		securityPolicyContainsLinks.Probe:                   "Security-Policy",
		securityPolicyContainsVulnerabilityDisclosure.Probe: "Security-Policy",
		securityPolicyContainsText.Probe:                    "Security-Policy",
		toolRenovateInstalled.Probe:                         "Dependency-Update-Tool",
		toolDependabotInstalled.Probe:                       "Dependency-Update-Tool",
		toolPyUpInstalled.Probe:                             "Dependency-Update-Tool",
		fuzzedWithOSSFuzz.Probe:                             "Fuzzing",
		fuzzedWithGoNative.Probe:                            "Fuzzing",
		fuzzedWithPythonAtheris.Probe:                       "Fuzzing",
		fuzzedWithCLibFuzzer.Probe:                          "Fuzzing",
		fuzzedWithCppLibFuzzer.Probe:                        "Fuzzing",
		fuzzedWithSwiftLibFuzzer.Probe:                      "Fuzzing",
		fuzzedWithRustCargofuzz.Probe:                       "Fuzzing",
		fuzzedWithJavaJazzerFuzzer.Probe:                    "Fuzzing",
		fuzzedWithClusterFuzzLite.Probe:                     "Fuzzing",
		fuzzedWithPropertyBasedHaskell.Probe:                "Fuzzing",
		fuzzedWithPropertyBasedTypescript.Probe:             "Fuzzing",
		fuzzedWithPropertyBasedJavascript.Probe:             "Fuzzing",
		packagedWithAutomatedWorkflow.Probe:                 "Packaging",
		hasLicenseFile.Probe:                                "License",
		hasFSFOrOSIApprovedLicense.Probe:                    "License",
		hasLicenseFileAtTopDir.Probe:                        "License",
		contributorsFromOrgOrCompany.Probe:                  "Contributors",
		hasOSVVulnerabilities.Probe:                         "Vulnerabilities",
		sastToolCodeQLInstalled.Probe:                       "SAST",
		sastToolRunsOnAllCommits.Probe:                      "SAST",
		sastToolSonarInstalled.Probe:                        "SAST",
		hasDangerousWorkflowScriptInjection.Probe:           "Dangerous-Workflow",
		hasDangerousWorkflowUntrustedCheckout.Probe:         "Dangerous-Workflow",
		notArchived.Probe:                                   "Maintained",
		hasRecentCommits.Probe:                              "Maintained",
		issueActivityByProjectMember.Probe:                  "Maintained",
		notCreatedRecently.Probe:                            "Maintained",
	}

	errProbeNotFound = errors.New("probe not found")
)

//nolint:gochecknoinits
func init() {
	All = concatMultipleProbes([][]ProbeImpl{
		DependencyToolUpdates,
		SecurityPolicy,
		Fuzzing,
		License,
		Contributors,
	})
}

func GetProbeRunner(probeName string) (func(*checker.RawResults) ([]finding.Finding, string, error), error) {
	if runner, ok := probeRunners[probeName]; ok {
		return runner, nil
	}
	return nil, errProbeNotFound
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
