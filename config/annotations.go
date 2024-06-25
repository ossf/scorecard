// Copyright 2024 OpenSSF Scorecard Authors
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

package config

// Reason is the reason behind an annotation.
type Reason string

const (
	// TestData is to annotate when a check or probe is targeting a danger
	// in files or code snippets only used for test or example purposes.
	TestData Reason = "test-data"
	// Remediated is to annotate when a check or probe correctly identified a
	// danger and, even though the danger is necessary, a remediation was already applied.
	// E.g. a workflow is dangerous but only run under maintainers verification and approval,
	// or a binary is needed but it is signed or has provenance.
	Remediated Reason = "remediated"
	// NotApplicable is to annotate when a check or probe is not applicable for the case.
	// E.g. the dependencies should not be pinned because the project is a library.
	NotApplicable Reason = "not-applicable"
	// NotSupported is to annotate when the maintainer fulfills a check or probe in a way
	// that is not supported by Scorecard. E.g. Clang-Tidy is used as SAST tool but not identified
	// because its not supported.
	NotSupported Reason = "not-supported"
	// NotDetected is to annotate when the maintainer fulfills a check or probe in a way
	// that is supported by Scorecard but not identified. E.g. Dependabot is configured in the
	// repository settings and not in a file.
	NotDetected Reason = "not-detected"
)

// ReasonGroup groups the annotation reason and, in the future, the related probe.
// If there is a probe, the reason applies to the probe.
// If there is not a probe, the reason applies to the check or checks in
// the group.
type ReasonGroup struct {
	Reason Reason `yaml:"reason"`
}

// Annotation defines a group of checks that are being annotated for various reasons.
type Annotation struct {
	Checks  []string      `yaml:"checks"`
	Reasons []ReasonGroup `yaml:"reasons"`
}

// Doc maps a reason to its human-readable explanation.
func (r *Reason) Doc() string {
	switch *r {
	case TestData:
		return "The files or code snippets are only used for test or example purposes."
	case Remediated:
		return "The dangerous files or code snippets are necessary but remediations were already applied."
	case NotApplicable:
		return "The check or probe is not applicable in this case."
	case NotSupported:
		return "The check or probe is fulfilled but in a way that is not supported by Scorecard."
	case NotDetected:
		return "The check or probe is fulfilled but in a way that is supported by Scorecard but it was not detected."
	default:
		return string(*r)
	}
}

// isValidReason checks if a reason can be used by a config file.
func isValidReason(r Reason) bool {
	// the reason must be one of the preselected options
	switch r {
	case TestData, Remediated, NotApplicable, NotSupported, NotDetected:
		return true
	default:
		return false
	}
}
