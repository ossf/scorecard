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
	"github.com/ossf/scorecard/v5/probes/hasBinaryArtifacts"
	"github.com/ossf/scorecard/v5/probes/hasUnverifiedBinaryArtifacts"
)

const (
	// QA0501 is the requirement that the version control system contains
	// no generated executable artifacts.
	QA0501 = "OSPS-QA-05.01"
	// QA0502 is the requirement that the version control system contains
	// no unreviewable binary artifacts.
	QA0502 = "OSPS-QA-05.02"

	qa05Title = "Prevent Executables in the Codebase"
)

// QA0501Mapping evaluates whether the repository contains generated
// executables.
func QA0501Mapping() conformance.Mapping {
	return conformance.Mapping{
		Requirement:  QA0501,
		ControlTitle: qa05Title,
		Description: "Checks the repository for binary files. Scorecard does not separate " +
			"generated executables from other binaries, so any unverified binary counts. " +
			"Validated Gradle wrappers are generated executables and need review.",
		Probes: []string{hasBinaryArtifacts.Probe, hasUnverifiedBinaryArtifacts.Probe},
		Steps: []string{
			"Read the hasUnverifiedBinaryArtifacts findings.",
			"Read the hasBinaryArtifacts findings.",
		},
		Evaluate: evaluateQA0501,
	}
}

// QA0502Mapping evaluates whether the repository contains binaries that
// can't be reviewed.
func QA0502Mapping() conformance.Mapping {
	return conformance.Mapping{
		Requirement:  QA0502,
		ControlTitle: qa05Title,
		Description: "Checks the repository for binary files other than Gradle wrapper JARs " +
			"that a workflow validates against Gradle's published checksums.",
		Probes: []string{hasBinaryArtifacts.Probe, hasUnverifiedBinaryArtifacts.Probe},
		Steps: []string{
			"Read the hasUnverifiedBinaryArtifacts findings.",
			"Record the hasBinaryArtifacts findings as evidence.",
		},
		Evaluate: evaluateQA0502,
	}
}

func evaluateQA0501(in conformance.Input) conformance.Outcome {
	bin := probeFindings(in, hasBinaryArtifacts.Probe)
	unverified := probeFindings(in, hasUnverifiedBinaryArtifacts.Probe)
	evidence := probeFindings(in, hasBinaryArtifacts.Probe, hasUnverifiedBinaryArtifacts.Probe)
	switch {
	case anyOutcome(unverified, finding.OutcomeTrue):
		return conformance.Outcome{
			Result:      conformance.Fail,
			Explanation: unverifiedBinaries(unverified),
			Evidence:    evidence,
		}
	case onlyOutcome(bin, finding.OutcomeFalse):
		return conformance.Outcome{
			Result:      conformance.Pass,
			Explanation: "No binary files were found in the repository.",
			Evidence:    evidence,
		}
	case anyOutcome(bin, finding.OutcomeTrue) && onlyOutcome(unverified, finding.OutcomeFalse):
		return conformance.Outcome{
			Result: conformance.NeedsReview,
			Explanation: fmt.Sprintf("The only binary files are Gradle wrapper JARs that a "+
				"workflow validates (%s). They are still generated executables.",
				strings.Join(paths(withOutcome(bin, finding.OutcomeTrue)), ", ")),
			Evidence: evidence,
		}
	default:
		return unknown(hasBinaryArtifacts.Probe+" and "+hasUnverifiedBinaryArtifacts.Probe,
			"whether the repository contains generated executables", evidence)
	}
}

func evaluateQA0502(in conformance.Input) conformance.Outcome {
	bin := probeFindings(in, hasBinaryArtifacts.Probe)
	unverified := probeFindings(in, hasUnverifiedBinaryArtifacts.Probe)
	evidence := probeFindings(in, hasBinaryArtifacts.Probe, hasUnverifiedBinaryArtifacts.Probe)
	switch {
	case anyOutcome(unverified, finding.OutcomeTrue):
		return conformance.Outcome{
			Result:      conformance.Fail,
			Explanation: unverifiedBinaries(unverified),
			Evidence:    evidence,
		}
	case onlyOutcome(unverified, finding.OutcomeFalse):
		explanation := "No binary files were found in the repository."
		if wrappers := paths(withOutcome(bin, finding.OutcomeTrue)); len(wrappers) > 0 {
			explanation = fmt.Sprintf("The only binary files are Gradle wrapper JARs that a "+
				"workflow validates against Gradle's published checksums (%s).", strings.Join(wrappers, ", "))
		}
		return conformance.Outcome{
			Result:      conformance.Pass,
			Explanation: explanation,
			Evidence:    evidence,
		}
	default:
		return unknown(hasUnverifiedBinaryArtifacts.Probe,
			"whether the repository contains unreviewable binaries", unverified)
	}
}

func unverifiedBinaries(unverified []finding.Finding) string {
	return fmt.Sprintf("The repository contains binary files: %s.",
		strings.Join(paths(withOutcome(unverified, finding.OutcomeTrue)), ", "))
}
