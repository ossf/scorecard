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
	"github.com/ossf/scorecard/v5/probes/hasFSFOrOSIApprovedLicense"
	"github.com/ossf/scorecard/v5/probes/hasLicenseFile"
)

const (
	// LE0201 is the requirement that the source code license meets the
	// OSI Open Source Definition or the FSF Free Software Definition.
	LE0201 = "OSPS-LE-02.01"
	// LE0202 is the requirement that the license of the released software
	// assets meets the OSI or FSF definition.
	LE0202 = "OSPS-LE-02.02"
	// LE0301 is the requirement that the source code license is kept in
	// the repository's LICENSE file, COPYING file, LICENSES/ directory, or
	// LICENSE/ directory.
	LE0301 = "OSPS-LE-03.01"
	// LE0302 is the requirement that the license is included in the
	// released source code or alongside the release assets.
	LE0302 = "OSPS-LE-03.02"

	le02Title = "Ensure Project Licenses are Fully Open Source"
	le03Title = "Maintain and Release Licenses in a Well Known Location"
)

// LE0201Mapping evaluates whether the source code license is OSI or FSF
// approved.
func LE0201Mapping() conformance.Mapping {
	return conformance.Mapping{
		Requirement:  LE0201,
		ControlTitle: le02Title,
		Description:  "Checks that the repository's license is OSI or FSF approved.",
		Probes:       []string{hasFSFOrOSIApprovedLicense.Probe},
		Steps: []string{
			"Read the hasFSFOrOSIApprovedLicense finding.",
		},
		Evaluate: evaluateLE0201,
	}
}

// LE0202Mapping evaluates the released assets' license from the default
// branch's license.
func LE0202Mapping() conformance.Mapping {
	return conformance.Mapping{
		Requirement:  LE0202,
		ControlTitle: le02Title,
		Description: "Checks that the repository's license is OSI or FSF approved. The " +
			"license is read from the default branch, not from the released assets.",
		Probes: []string{hasFSFOrOSIApprovedLicense.Probe},
		Steps: []string{
			"Read the hasFSFOrOSIApprovedLicense finding.",
		},
		Evaluate: evaluateLE0202,
	}
}

// LE0301Mapping evaluates whether the license is in a LICENSE or COPYING
// file or directory.
func LE0301Mapping() conformance.Mapping {
	return conformance.Mapping{
		Requirement:  LE0301,
		ControlTitle: le03Title,
		Description: "Checks that the repository has a license file named LICENSE, LICENCE, " +
			"LICENSES, or COPYING, or a directory with one of those names.",
		Probes: []string{hasLicenseFile.Probe},
		Steps: []string{
			"Read the hasLicenseFile findings.",
			"Check the name of each license file found.",
		},
		Evaluate: evaluateLE0301,
	}
}

// LE0302Mapping evaluates whether the license ships with releases from the
// default branch's license file.
func LE0302Mapping() conformance.Mapping {
	return conformance.Mapping{
		Requirement:  LE0302,
		ControlTitle: le03Title,
		Description: "Checks that the repository has a license file. The default branch is " +
			"read, not the released source code or the release assets.",
		Probes: []string{hasLicenseFile.Probe},
		Steps: []string{
			"Read the hasLicenseFile findings.",
		},
		Evaluate: evaluateLE0302,
	}
}

func evaluateLE0201(in conformance.Input) conformance.Outcome {
	probe := hasFSFOrOSIApprovedLicense.Probe
	fs := probeFindings(in, probe)
	switch {
	case onlyOutcome(fs, finding.OutcomeTrue):
		return conformance.Outcome{
			Result:      conformance.Pass,
			Explanation: "The repository's license is OSI or FSF approved.",
			Evidence:    fs,
		}
	case anyOutcome(fs, finding.OutcomeFalse):
		return conformance.Outcome{
			Result:      conformance.Fail,
			Explanation: "The repository's license is not one Scorecard recognizes as OSI or FSF approved.",
			Evidence:    fs,
		}
	case onlyOutcome(fs, finding.OutcomeNotApplicable):
		return noLicense(fs)
	default:
		return unknown(probe, "whether the license is OSI or FSF approved", fs)
	}
}

func evaluateLE0202(in conformance.Input) conformance.Outcome {
	probe := hasFSFOrOSIApprovedLicense.Probe
	fs := probeFindings(in, probe)
	switch {
	case anyOutcome(fs, finding.OutcomeFalse):
		return conformance.Outcome{
			Result: conformance.Fail,
			Explanation: "The repository's license, read from the default branch, is not one " +
				"Scorecard recognizes as OSI or FSF approved.",
			Evidence: fs,
		}
	case onlyOutcome(fs, finding.OutcomeTrue):
		return conformance.Outcome{
			Result: conformance.NeedsReview,
			Explanation: "The repository's license is OSI or FSF approved. It was read from the " +
				"default branch, not from the released assets.",
			Evidence: fs,
		}
	case onlyOutcome(fs, finding.OutcomeNotApplicable):
		return noLicense(fs)
	default:
		return unknown(probe, "whether the released assets' license is OSI or FSF approved", fs)
	}
}

func evaluateLE0301(in conformance.Input) conformance.Outcome {
	probe := hasLicenseFile.Probe
	fs := probeFindings(in, probe)
	found := withOutcome(fs, finding.OutcomeTrue)
	var named, other []string
	for _, p := range paths(found) {
		if isLicenseName(p) {
			named = append(named, p)
		} else {
			other = append(other, p)
		}
	}
	switch {
	case len(named) > 0:
		return conformance.Outcome{
			Result:      conformance.Pass,
			Explanation: fmt.Sprintf("The license is in %s.", strings.Join(named, ", ")),
			Evidence:    fs,
		}
	case len(found) > 0:
		return conformance.Outcome{
			Result: conformance.NeedsReview,
			Explanation: fmt.Sprintf("A license file was found (%s), but not under a name the "+
				"requirement lists: LICENSE, COPYING, LICENSES/, or LICENSE/.",
				strings.Join(other, ", ")),
			Evidence: fs,
		}
	case onlyOutcome(fs, finding.OutcomeFalse):
		return conformance.Outcome{
			Result:      conformance.Fail,
			Explanation: "No license file was found.",
			Evidence:    fs,
		}
	default:
		return unknown(probe, "whether the repository has a license file", fs)
	}
}

func evaluateLE0302(in conformance.Input) conformance.Outcome {
	probe := hasLicenseFile.Probe
	fs := probeFindings(in, probe)
	switch {
	case anyOutcome(fs, finding.OutcomeTrue):
		return conformance.Outcome{
			Result: conformance.NeedsReview,
			Explanation: fmt.Sprintf("The default branch has a license file (%s). The released "+
				"source code and the release assets were not read.",
				strings.Join(paths(withOutcome(fs, finding.OutcomeTrue)), ", ")),
			Evidence: fs,
		}
	case onlyOutcome(fs, finding.OutcomeFalse):
		return conformance.Outcome{
			Result:      conformance.Fail,
			Explanation: "No license file was found on the default branch.",
			Evidence:    fs,
		}
	default:
		return unknown(probe, "whether releases include the license", fs)
	}
}

// noLicense returns Unknown when no license file was found to judge.
// OSPS-LE-03.01 reports the missing file.
func noLicense(fs []finding.Finding) conformance.Outcome {
	return conformance.Outcome{
		Result:      conformance.Unknown,
		Explanation: "No license file was found, so its license could not be judged. OSPS-LE-03.01 reports the missing file.",
		Evidence:    fs,
	}
}

// isLicenseName reports whether a license file's top-level name is LICENSE,
// LICENCE, LICENSES, LICENCES, or COPYING, alone or joined to an SPDX
// identifier or extension by "-", "_", or ".". hasLicenseFile also accepts
// names such as PATENTS and COPYRIGHT, which OSPS-LE-03.01 doesn't list.
func isLicenseName(path string) bool {
	top, _, _ := strings.Cut(path, "/")
	for _, part := range strings.FieldsFunc(top, func(r rune) bool {
		return r == '-' || r == '_' || r == '.'
	}) {
		switch strings.ToUpper(part) {
		case "LICENSE", "LICENCE", "LICENSES", "LICENCES", "COPYING":
			return true
		}
	}
	return false
}
