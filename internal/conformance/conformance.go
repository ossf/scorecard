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

// Package conformance evaluates catalog requirements, such as those in the
// OSPS Baseline, from Scorecard probe findings.
//
// Conformance mode is experimental. The package is internal so that no
// importer depends on it before it settles.
package conformance

import (
	"fmt"
	"slices"
	"time"

	"github.com/ossf/scorecard/v5/finding"
)

// Result is the outcome of evaluating one requirement.
type Result string

const (
	// Pass means the evidence shows the requirement is met.
	Pass Result = "PASS"
	// Fail means the evidence shows the requirement is not met.
	Fail Result = "FAIL"
	// Unknown means evidence needed to decide is missing, or a probe errored
	// or could not observe it.
	Unknown Result = "UNKNOWN"
	// NotApplicable means the requirement does not apply to the selected
	// profile.
	NotApplicable Result = "NOT_APPLICABLE"
	// NeedsReview means evidence exists, but deciding needs judgment that the
	// evaluator does not make. A person or another tool can make it.
	NeedsReview Result = "NEEDS_REVIEW"
)

// Requirement is one assessment requirement in a catalog release.
type Requirement struct {
	// ID is the requirement's identifier in the catalog, such as
	// "OSPS-AC-03.01".
	ID string
	// Control is the identifier of the control the requirement belongs to,
	// such as "OSPS-AC-03".
	Control string
	// Applicability lists the applicability groups the requirement belongs
	// to, such as "maturity-1".
	Applicability []string
}

// AppliesTo reports whether the requirement belongs to the given
// applicability group.
func (r *Requirement) AppliesTo(profile string) bool {
	return slices.Contains(r.Applicability, profile)
}

// Input is the data an evaluation function reads.
type Input struct {
	// DefaultBranch is the name of the repository's default branch, or
	// empty if it is unknown.
	DefaultBranch string
	// Findings are the findings of the probes the mappings name.
	Findings []finding.Finding
}

// Outcome is what an evaluation function returns for one requirement.
type Outcome struct {
	Result      Result
	Explanation string
	Evidence    []finding.Finding
}

// Mapping says how Scorecard evaluates one requirement.
type Mapping struct {
	// Evaluate decides the requirement from the input. It must return
	// Unknown, never Pass or Fail, when a finding it needs is missing or
	// has an outcome other than True or False.
	Evaluate func(Input) Outcome
	// Requirement is the ID of the requirement this mapping evaluates.
	Requirement string
	// ControlTitle is the title of the requirement's control.
	ControlTitle string
	// Description summarizes how the requirement is assessed.
	Description string
	// Probes are the probes whose findings Evaluate reads.
	Probes []string
	// Steps describe the assessment in order.
	Steps []string
}

// Release is one released version of a catalog, with Scorecard's mappings
// for it. A release must not change meaning after it ships: requirement
// IDs and applicability stay as published.
type Release struct {
	// Catalog is the catalog ID, such as "osps-baseline".
	Catalog string
	// Title is the catalog's human-readable title.
	Title string
	// Version is the release version, such as "v2026.08.28".
	Version string
	// URL points at the published release.
	URL string
	// Profiles are the catalog's applicability group IDs. They are not
	// ordered.
	Profiles []string
	// Requirements lists every requirement in the release, in catalog
	// order.
	Requirements []Requirement
	// Mappings lists the requirements Scorecard evaluates.
	Mappings []Mapping
}

// HasProfile reports whether the release defines the applicability group.
func (r *Release) HasProfile(profile string) bool {
	return slices.Contains(r.Profiles, profile)
}

// Requirement returns the requirement with the given ID.
func (r *Release) Requirement(id string) (*Requirement, bool) {
	for i := range r.Requirements {
		if r.Requirements[i].ID == id {
			return &r.Requirements[i], true
		}
	}
	return nil, false
}

// Mapping returns the mapping for the requirement with the given ID.
func (r *Release) Mapping(id string) (*Mapping, bool) {
	for i := range r.Mappings {
		if r.Mappings[i].Requirement == id {
			return &r.Mappings[i], true
		}
	}
	return nil, false
}

// Probes returns the sorted, de-duplicated probes that the mappings for the
// profile's requirements name. Mappings for requirements outside the
// profile run no probes.
func (r *Release) Probes(profile string) []string {
	var probes []string
	for i := range r.Mappings {
		m := &r.Mappings[i]
		req, ok := r.Requirement(m.Requirement)
		if !ok || !req.AppliesTo(profile) {
			continue
		}
		probes = append(probes, m.Probes...)
	}
	slices.Sort(probes)
	return slices.Compact(probes)
}

// RequirementResult is the result for one evaluated requirement.
type RequirementResult struct {
	Requirement string            `json:"requirement"`
	Result      Result            `json:"result"`
	Explanation string            `json:"explanation"`
	Evidence    []finding.Finding `json:"evidence"`
}

// Report is the outcome of one conformance run.
type Report struct {
	Date             time.Time
	Release          *Release
	Repo             string
	Commit           string
	ScorecardVersion string
	ScorecardCommit  string
	Profile          string
	Results          []RequirementResult
	NotEvaluated     []string
}

// Evaluate evaluates every mapped requirement in the release for the
// profile. It returns results in catalog order, and the IDs of the
// requirements in the profile that have no mapping.
func Evaluate(release *Release, profile string, in Input) ([]RequirementResult, []string) {
	var results []RequirementResult
	var notEvaluated []string
	for i := range release.Requirements {
		req := &release.Requirements[i]
		m, mapped := release.Mapping(req.ID)
		switch {
		case !mapped && req.AppliesTo(profile):
			notEvaluated = append(notEvaluated, req.ID)
		case !mapped:
			// Outside the profile and not mapped: not part of this run.
		case !req.AppliesTo(profile):
			results = append(results, RequirementResult{
				Requirement: req.ID,
				Result:      NotApplicable,
				Explanation: fmt.Sprintf("%s does not apply to %s.", req.ID, profile),
				Evidence:    []finding.Finding{},
			})
		default:
			out := m.Evaluate(in)
			evidence := out.Evidence
			if evidence == nil {
				evidence = []finding.Finding{}
			}
			results = append(results, RequirementResult{
				Requirement: req.ID,
				Result:      out.Result,
				Explanation: out.Explanation,
				Evidence:    evidence,
			})
		}
	}
	return results, notEvaluated
}

// DefaultBranchFinding returns the finding from the named probe for the
// default branch. Findings for other branches are ignored. It returns
// false if the default branch is unknown or no finding matches.
func DefaultBranchFinding(in Input, probe, branchKey string) (finding.Finding, bool) {
	if in.DefaultBranch == "" {
		return finding.Finding{}, false
	}
	for i := range in.Findings {
		f := &in.Findings[i]
		if f.Probe == probe && f.Values[branchKey] == in.DefaultBranch {
			return *f, true
		}
	}
	return finding.Finding{}, false
}
