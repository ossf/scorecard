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
	"github.com/ossf/scorecard/v5/probes/hasDangerousWorkflowScriptInjection"
	"github.com/ossf/scorecard/v5/probes/hasDangerousWorkflowUntrustedCheckout"
)

const (
	// BR0101 is the requirement that untrusted metadata is sanitized and
	// validated before a CI/CD pipeline uses it.
	BR0101 = "OSPS-BR-01.01"
	// BR0103 is the requirement that a CI/CD pipeline operating on
	// untrusted code snapshots prevents access to privileged credentials
	// and assets.
	BR0103 = "OSPS-BR-01.03"

	br01Title = "Prevent Untrusted Input When Building & Releasing"
)

// BR0101Mapping evaluates whether GitHub Actions workflows use untrusted
// metadata in scripts.
func BR0101Mapping() conformance.Mapping {
	return conformance.Mapping{
		Requirement:  BR0101,
		ControlTitle: br01Title,
		Description: "Checks GitHub Actions workflows for known script-injection patterns. " +
			"CI/CD defined elsewhere is not read, and finding no pattern does not show that " +
			"all untrusted metadata is sanitized.",
		Probes: []string{hasDangerousWorkflowScriptInjection.Probe},
		Steps: []string{
			"Read the hasDangerousWorkflowScriptInjection findings.",
		},
		Evaluate: evaluateBR0101,
	}
}

// BR0103Mapping evaluates whether workflows check out untrusted code in a
// privileged context.
func BR0103Mapping() conformance.Mapping {
	return conformance.Mapping{
		Requirement:  BR0103,
		ControlTitle: br01Title,
		Description: "Checks GitHub Actions workflows triggered by pull_request_target or " +
			"workflow_run for checkouts of pull request code. Conditions and job permissions " +
			"are not read.",
		Probes: []string{hasDangerousWorkflowUntrustedCheckout.Probe},
		Steps: []string{
			"Read the hasDangerousWorkflowUntrustedCheckout findings.",
		},
		Evaluate: evaluateBR0103,
	}
}

func evaluateBR0101(in conformance.Input) conformance.Outcome {
	probe := hasDangerousWorkflowScriptInjection.Probe
	fs := probeFindings(in, probe)
	switch {
	case anyOutcome(fs, finding.OutcomeTrue):
		return conformance.Outcome{
			Result: conformance.Fail,
			Explanation: fmt.Sprintf("Untrusted metadata is used in a workflow script without "+
				"sanitization: %s.", strings.Join(paths(withOutcome(fs, finding.OutcomeTrue)), ", ")),
			Evidence: fs,
		}
	case onlyOutcome(fs, finding.OutcomeFalse):
		return conformance.Outcome{
			Result: conformance.NeedsReview,
			Explanation: "No known script-injection pattern was found in the workflows. Only " +
				"known patterns were read, so other uses of untrusted metadata need review.",
			Evidence: fs,
		}
	case onlyOutcome(fs, finding.OutcomeNotApplicable):
		return noWorkflows(fs)
	default:
		return unknown(probe, "whether workflows use untrusted metadata safely", fs)
	}
}

func evaluateBR0103(in conformance.Input) conformance.Outcome {
	probe := hasDangerousWorkflowUntrustedCheckout.Probe
	fs := probeFindings(in, probe)
	switch {
	case anyOutcome(fs, finding.OutcomeTrue):
		return conformance.Outcome{
			Result: conformance.Fail,
			Explanation: fmt.Sprintf("A pull_request_target or workflow_run workflow checks out "+
				"pull request code in a context that can reach repository secrets: %s. "+
				"Conditions such as a required label, and job permissions, were not read.",
				strings.Join(paths(withOutcome(fs, finding.OutcomeTrue)), ", ")),
			Evidence: fs,
		}
	case onlyOutcome(fs, finding.OutcomeFalse):
		return conformance.Outcome{
			Result: conformance.NeedsReview,
			Explanation: "No checkout of pull request code was found in pull_request_target or " +
				"workflow_run workflows. Only those two triggers were read, so other ways " +
				"untrusted code could reach credentials need review.",
			Evidence: fs,
		}
	case onlyOutcome(fs, finding.OutcomeNotApplicable):
		return noWorkflows(fs)
	default:
		return unknown(probe, "whether workflows check out untrusted code", fs)
	}
}

// noWorkflows returns Unknown for a repository with no GitHub Actions
// workflows: it may run CI/CD on another system.
func noWorkflows(fs []finding.Finding) conformance.Outcome {
	return conformance.Outcome{
		Result: conformance.Unknown,
		Explanation: "No workflows were found in .github/workflows. CI/CD defined elsewhere, " +
			"such as GitLab CI or Azure Pipelines, is not read.",
		Evidence: fs,
	}
}
