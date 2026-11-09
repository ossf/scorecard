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

// Package ospsbaseline holds Scorecard's evaluation functions for OSPS
// Baseline requirements. Each OSPS Baseline release has its own
// sub-package that lists the release's requirements and the mappings it
// uses from here.
package ospsbaseline

import (
	"fmt"

	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/conformance"
	"github.com/ossf/scorecard/v5/probes/blocksDeleteOnBranches"
	"github.com/ossf/scorecard/v5/probes/branchProtectionAppliesToAdmins"
	"github.com/ossf/scorecard/v5/probes/branchesAreProtected"
	"github.com/ossf/scorecard/v5/probes/requiresPRsToChangeCode"
)

const (
	// CatalogID is the OSPS Baseline catalog ID, as used by other Gemara
	// evaluators.
	CatalogID = "osps-baseline"
	// CatalogTitle is the OSPS Baseline catalog's title.
	CatalogTitle = "Open Source Project Security Baseline"

	// AC0301 is the requirement that direct commits to the primary branch
	// are prevented.
	AC0301 = "OSPS-AC-03.01"
	// AC0302 is the requirement that deleting the primary branch requires
	// explicit confirmation.
	AC0302 = "OSPS-AC-03.02"

	ac03Title = "Protect the Primary Branch from Accidental Modification"
)

// AC0301Mapping evaluates whether direct commits to the default branch are
// prevented.
func AC0301Mapping() conformance.Mapping {
	return conformance.Mapping{
		Requirement:  AC0301,
		ControlTitle: ac03Title,
		Description: "Checks that the default branch is protected, requires pull requests to " +
			"change code, and that the rule applies to administrators.",
		Probes: []string{
			branchProtectionAppliesToAdmins.Probe,
			branchesAreProtected.Probe,
			requiresPRsToChangeCode.Probe,
		},
		Steps: []string{
			"Read the branchesAreProtected finding for the default branch.",
			"Read the requiresPRsToChangeCode finding for the default branch.",
			"Read the branchProtectionAppliesToAdmins finding for the default branch.",
		},
		Evaluate: evaluateAC0301,
	}
}

// AC0302Mapping evaluates whether deleting the default branch is blocked.
func AC0302Mapping() conformance.Mapping {
	return conformance.Mapping{
		Requirement:  AC0302,
		ControlTitle: ac03Title,
		Description: "Checks that the default branch cannot be deleted. Blocking deletion is " +
			"stricter than requiring confirmation, so it meets the requirement. Allowing " +
			"deletion does not show whether the platform asks for confirmation.",
		Probes: []string{blocksDeleteOnBranches.Probe},
		Steps: []string{
			"Read the blocksDeleteOnBranches finding for the default branch.",
		},
		Evaluate: evaluateAC0302,
	}
}

func evaluateAC0301(in conformance.Input) conformance.Outcome {
	prs, ok := conformance.DefaultBranchFinding(in, requiresPRsToChangeCode.Probe, requiresPRsToChangeCode.BranchNameKey)
	if !ok {
		return missing(in, requiresPRsToChangeCode.Probe)
	}
	branch := in.DefaultBranch
	protected, protectedOK := conformance.DefaultBranchFinding(
		in, branchesAreProtected.Probe, branchesAreProtected.BranchNameKey)
	admins, adminsOK := conformance.DefaultBranchFinding(
		in, branchProtectionAppliesToAdmins.Probe, branchProtectionAppliesToAdmins.BranchNameKey)
	var evidence []finding.Finding
	if protectedOK {
		evidence = append(evidence, protected)
	}
	evidence = append(evidence, prs)
	if adminsOK {
		evidence = append(evidence, admins)
	}

	// A branch with no protection rule or ruleset leaves the pull request
	// setting unset, so requiresPRsToChangeCode alone would report it as
	// not available. branchesAreProtected also reports False when the
	// platform client can't tell whether the branch is protected (Azure
	// DevOps policies it can't read); the Branch-Protection check treats
	// that finding the same way.
	if protectedOK && protected.Outcome == finding.OutcomeFalse && prs.Outcome != finding.OutcomeTrue {
		return conformance.Outcome{
			Result:      conformance.Fail,
			Explanation: fmt.Sprintf("The default branch %q is not protected.", branch),
			Evidence:    evidence,
		}
	}

	switch prs.Outcome {
	case finding.OutcomeFalse:
		return conformance.Outcome{
			Result:      conformance.Fail,
			Explanation: fmt.Sprintf("Pull requests are not required to change the default branch %q.", branch),
			Evidence:    evidence,
		}
	case finding.OutcomeTrue:
		// Decided below from the administrator finding.
	default:
		return conformance.Outcome{
			Result: conformance.Unknown,
			Explanation: fmt.Sprintf("Could not determine whether the default branch %q requires pull requests "+
				"(requiresPRsToChangeCode: %s).", branch, prs.Outcome),
			Evidence: evidence,
		}
	}

	switch {
	case adminsOK && admins.Outcome == finding.OutcomeTrue:
		return conformance.Outcome{
			Result: conformance.Pass,
			Explanation: fmt.Sprintf("The default branch %q requires pull requests, and the rule applies "+
				"to administrators.", branch),
			Evidence: evidence,
		}
	case adminsOK && admins.Outcome == finding.OutcomeFalse:
		return conformance.Outcome{
			Result: conformance.NeedsReview,
			Explanation: fmt.Sprintf("The default branch %q requires pull requests, but administrators can "+
				"bypass the rule. The requirement does not say whether an administrator exception is "+
				"allowed.", branch),
			Evidence: evidence,
		}
	default:
		return conformance.Outcome{
			Result: conformance.Unknown,
			Explanation: fmt.Sprintf("The default branch %q requires pull requests, but whether the rule "+
				"applies to administrators could not be determined.", branch),
			Evidence: evidence,
		}
	}
}

func evaluateAC0302(in conformance.Input) conformance.Outcome {
	f, ok := conformance.DefaultBranchFinding(in, blocksDeleteOnBranches.Probe, blocksDeleteOnBranches.BranchNameKey)
	if !ok {
		return missing(in, blocksDeleteOnBranches.Probe)
	}
	branch := in.DefaultBranch
	evidence := []finding.Finding{f}

	switch f.Outcome {
	case finding.OutcomeTrue:
		return conformance.Outcome{
			Result:      conformance.Pass,
			Explanation: fmt.Sprintf("The default branch %q cannot be deleted.", branch),
			Evidence:    evidence,
		}
	case finding.OutcomeFalse:
		// The rule allows deletion, but the requirement asks only for
		// explicit confirmation, which no probe observes.
		return conformance.Outcome{
			Result: conformance.Unknown,
			Explanation: fmt.Sprintf("The branch rules allow deleting the default branch %q. Whether the "+
				"platform requires explicit confirmation to delete it could not be determined.", branch),
			Evidence: evidence,
		}
	default:
		return conformance.Outcome{
			Result: conformance.Unknown,
			Explanation: fmt.Sprintf("Could not determine whether the default branch %q can be deleted "+
				"(blocksDeleteOnBranches: %s).", branch, f.Outcome),
			Evidence: evidence,
		}
	}
}

// missing returns Unknown when a probe has no finding for the default
// branch. Findings from the probe that name no branch, such as "no
// branches found", are kept as evidence.
func missing(in conformance.Input, probe string) conformance.Outcome {
	var evidence []finding.Finding
	for i := range in.Findings {
		f := &in.Findings[i]
		if f.Probe == probe && len(f.Values) == 0 {
			evidence = append(evidence, *f)
		}
	}
	explanation := fmt.Sprintf("%s has no finding for the default branch %q.", probe, in.DefaultBranch)
	if in.DefaultBranch == "" {
		explanation = "The default branch is unknown."
	}
	return conformance.Outcome{
		Result:      conformance.Unknown,
		Explanation: explanation,
		Evidence:    evidence,
	}
}
