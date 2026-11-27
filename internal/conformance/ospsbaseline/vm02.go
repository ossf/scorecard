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

package ospsbaseline

import (
	"fmt"
	"strings"

	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/conformance"
	"github.com/ossf/scorecard/v5/probes/securityPolicyContainsLinks"
	"github.com/ossf/scorecard/v5/probes/securityPolicyPresent"
)

const (
	// VM0201 is the requirement that the project documentation contains
	// security contacts.
	VM0201 = "OSPS-VM-02.01"

	vm02Title = "Publish Contacts and Process for Reporting Vulnerabilities."

	// policyLocations is where the security policy probe looks.
	policyLocations = "SECURITY.md, .markdown, .adoc, or .rst at the repository root, in .github/, " +
		"or in docs/ (and doc/SECURITY.rst), then, where the platform supports it, in the " +
		"organization's .github repository"
)

// VM0201Mapping evaluates whether the project publishes a security policy
// that may name contacts.
func VM0201Mapping() conformance.Mapping {
	return conformance.Mapping{
		Requirement:  VM0201,
		ControlTitle: vm02Title,
		Description: "Checks for a security policy file. A policy is needed for security " +
			"contacts, but whether it names them needs review: the link check matches any " +
			"URL or email address.",
		Probes: []string{securityPolicyContainsLinks.Probe, securityPolicyPresent.Probe},
		Steps: []string{
			"Read the securityPolicyPresent findings.",
			"Record the securityPolicyContainsLinks findings as evidence.",
		},
		Evaluate: evaluateVM0201,
	}
}

func evaluateVM0201(in conformance.Input) conformance.Outcome {
	present := probeFindings(in, securityPolicyPresent.Probe)
	evidence := probeFindings(in, securityPolicyPresent.Probe, securityPolicyContainsLinks.Probe)
	switch {
	case anyOutcome(present, finding.OutcomeTrue):
		links := probeFindings(in, securityPolicyContainsLinks.Probe)
		detail := "No URL or email address was found in it."
		if anyOutcome(links, finding.OutcomeTrue) {
			detail = "It contains a URL or email address, which may or may not be a contact."
		}
		return conformance.Outcome{
			Result: conformance.NeedsReview,
			Explanation: fmt.Sprintf("A security policy was found (%s). %s",
				strings.Join(paths(withOutcome(present, finding.OutcomeTrue)), ", "), detail),
			Evidence: evidence,
		}
	case onlyOutcome(present, finding.OutcomeFalse):
		return conformance.Outcome{
			Result:      conformance.Fail,
			Explanation: "No security policy was found. Scorecard looked for " + policyLocations + ".",
			Evidence:    evidence,
		}
	default:
		return unknown(securityPolicyPresent.Probe, "whether the project has a security policy", evidence)
	}
}
