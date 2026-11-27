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
	"slices"
	"strings"

	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/conformance"
)

// probeFindings returns the findings from the named probes, in input
// order.
func probeFindings(in conformance.Input, probes ...string) []finding.Finding {
	var out []finding.Finding
	for i := range in.Findings {
		if slices.Contains(probes, in.Findings[i].Probe) {
			out = append(out, in.Findings[i])
		}
	}
	return out
}

// anyOutcome reports whether any finding has the outcome.
func anyOutcome(findings []finding.Finding, o finding.Outcome) bool {
	for i := range findings {
		if findings[i].Outcome == o {
			return true
		}
	}
	return false
}

// onlyOutcome reports whether there is at least one finding and every
// finding has the outcome.
func onlyOutcome(findings []finding.Finding, o finding.Outcome) bool {
	if len(findings) == 0 {
		return false
	}
	for i := range findings {
		if findings[i].Outcome != o {
			return false
		}
	}
	return true
}

// withOutcome returns the findings that have the outcome.
func withOutcome(findings []finding.Finding, o finding.Outcome) []finding.Finding {
	var out []finding.Finding
	for i := range findings {
		if findings[i].Outcome == o {
			out = append(out, findings[i])
		}
	}
	return out
}

// paths returns the location paths of the findings, skipping findings
// without one.
func paths(findings []finding.Finding) []string {
	var out []string
	for i := range findings {
		if loc := findings[i].Location; loc != nil && loc.Path != "" {
			out = append(out, loc.Path)
		}
	}
	return out
}

// unknown returns Unknown for findings that are missing or that no rule
// decides, naming the probe and the outcomes it reported.
func unknown(probe, what string, findings []finding.Finding) conformance.Outcome {
	if len(findings) == 0 {
		return conformance.Outcome{
			Result:      conformance.Unknown,
			Explanation: fmt.Sprintf("Could not determine %s: %s returned no findings.", what, probe),
		}
	}
	var outcomes []string
	for i := range findings {
		o := string(findings[i].Outcome)
		if !slices.Contains(outcomes, o) {
			outcomes = append(outcomes, o)
		}
	}
	return conformance.Outcome{
		Result:      conformance.Unknown,
		Explanation: fmt.Sprintf("Could not determine %s (%s: %s).", what, probe, strings.Join(outcomes, ", ")),
		Evidence:    findings,
	}
}
