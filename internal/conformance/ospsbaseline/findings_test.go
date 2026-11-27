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
	"strings"
	"testing"

	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/conformance"
)

// repoFinding returns a finding with no location.
func repoFinding(probe string, outcome finding.Outcome) finding.Finding {
	return finding.Finding{Probe: probe, Outcome: outcome, Message: probe + " " + string(outcome)}
}

// fileFinding returns a finding located at path.
func fileFinding(probe string, outcome finding.Outcome, path string) finding.Finding {
	f := repoFinding(probe, outcome)
	f.Location = &finding.Location{Path: path}
	return f
}

// mappingCase is one evaluation of a mapping.
type mappingCase struct {
	name         string
	want         conformance.Result
	wantText     string // a substring the explanation must contain
	findings     []finding.Finding
	wantEvidence int
}

func runMappingCases(t *testing.T, evaluate func(conformance.Input) conformance.Outcome, tests []mappingCase) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := evaluate(conformance.Input{DefaultBranch: "main", Findings: tt.findings})
			if got.Result != tt.want {
				t.Errorf("result = %s (%s), want %s", got.Result, got.Explanation, tt.want)
			}
			if len(got.Evidence) != tt.wantEvidence {
				t.Errorf("evidence has %d findings, want %d: %v", len(got.Evidence), tt.wantEvidence, got.Evidence)
			}
			if got.Explanation == "" {
				t.Error("explanation is empty")
			}
			if !strings.Contains(got.Explanation, tt.wantText) {
				t.Errorf("explanation %q lacks %q", got.Explanation, tt.wantText)
			}
		})
	}
}

func TestIsLicenseName(t *testing.T) {
	t.Parallel()
	tests := map[string]bool{
		"LICENSE":            true,
		"license.md":         true,
		"LICENCE.txt":        true,
		"COPYING":            true,
		"copying.lesser":     true,
		"LICENSE-APACHE":     true,
		"Apache-2.0-LICENSE": true,
		"LICENSE_MIT.txt":    true,
		"LICENSES/MIT.txt":   true,
		"LICENSE/GPL-3.0.md": true,
		"COPYRIGHT":          false,
		"PATENTS":            false,
		"UNLICENSE":          false,
		"COPYRIGHT/LICENSE":  false, // only the top-level name counts
	}
	for path, want := range tests {
		if got := isLicenseName(path); got != want {
			t.Errorf("isLicenseName(%q) = %v, want %v", path, got, want)
		}
	}
}

func TestUnknownNamesOutcomes(t *testing.T) {
	t.Parallel()
	got := unknown("someProbe", "something", []finding.Finding{
		repoFinding("someProbe", finding.OutcomeError),
		repoFinding("someProbe", finding.OutcomeError),
		repoFinding("someProbe", finding.OutcomeNotAvailable),
	})
	if got.Result != conformance.Unknown {
		t.Errorf("result = %s, want UNKNOWN", got.Result)
	}
	if want := "(someProbe: Error, NotAvailable)"; !strings.Contains(got.Explanation, want) {
		t.Errorf("explanation %q lacks %q", got.Explanation, want)
	}
	if len(got.Evidence) != 3 {
		t.Errorf("evidence has %d findings, want 3", len(got.Evidence))
	}

	none := unknown("someProbe", "something", nil)
	if !strings.Contains(none.Explanation, "returned no findings") {
		t.Errorf("explanation %q does not say the probe returned no findings", none.Explanation)
	}
}
