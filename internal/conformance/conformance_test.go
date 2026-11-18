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

package conformance

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ossf/scorecard/v5/finding"
)

// fixed returns an evaluation function that always gives the outcome.
func fixed(r Result, explanation string, evidence ...finding.Finding) func(Input) Outcome {
	return func(Input) Outcome {
		return Outcome{Result: r, Explanation: explanation, Evidence: evidence}
	}
}

var testFinding = finding.Finding{
	Probe:   "testProbe",
	Message: "test probe observed the setting on branch 'main'",
	Outcome: finding.OutcomeTrue,
	Values:  map[string]string{"branchName": "main"},
}

// testRelease is a test-only catalog release. TEST-01.05 applies only to
// level-3, so it gives NOT_APPLICABLE for level-1. TEST-02.01 has no
// mapping and is in level-1, so it is not evaluated. TEST-03.01 has no
// mapping and is outside level-1, so it is not part of a level-1 run.
func testRelease() *Release {
	all := []string{"level-1", "level-2", "level-3"}
	return &Release{
		Catalog:  "test-catalog",
		Title:    "Test Catalog",
		Version:  "v1.0.0",
		URL:      "https://example.com/test-catalog/v1.0.0",
		Profiles: all,
		Requirements: []Requirement{
			{ID: "TEST-01.01", Control: "TEST-01", Applicability: all},
			{ID: "TEST-01.02", Control: "TEST-01", Applicability: all},
			{ID: "TEST-01.03", Control: "TEST-01", Applicability: all},
			{ID: "TEST-01.04", Control: "TEST-01", Applicability: all},
			{ID: "TEST-01.05", Control: "TEST-01", Applicability: []string{"level-3"}},
			{ID: "TEST-02.01", Control: "TEST-02", Applicability: all},
			{ID: "TEST-03.01", Control: "TEST-03", Applicability: []string{"level-3"}},
		},
		Mappings: []Mapping{
			{
				Requirement: "TEST-01.01", ControlTitle: "Test Control", Description: "Passes.",
				Probes: []string{"probeB", "probeA"}, Steps: []string{"Read probeA."},
				Evaluate: fixed(Pass, "The setting is on.", testFinding),
			},
			{
				Requirement: "TEST-01.02", ControlTitle: "Test Control", Description: "Fails.",
				Probes: []string{"probeA"}, Steps: []string{"Read probeA."},
				Evaluate: fixed(Fail, "The setting is off.", testFinding),
			},
			{
				Requirement: "TEST-01.03", ControlTitle: "Test Control", Description: "Is unknown.",
				Probes: []string{"probeA"}, Steps: []string{"Read probeA."},
				Evaluate: fixed(Unknown, "The setting could not be read."),
			},
			{
				Requirement: "TEST-01.04", ControlTitle: "Test Control", Description: "Needs review.",
				Probes: []string{"probeA"}, Steps: []string{"Read probeA."},
				Evaluate: fixed(NeedsReview, "The setting has an exception.", testFinding),
			},
			{
				Requirement: "TEST-01.05", ControlTitle: "Test Control", Description: "Applies at level 3.",
				Probes: []string{"probeLevel3"}, Steps: []string{"Read probeLevel3."},
				Evaluate: func(Input) Outcome { panic("evaluated a requirement outside the profile") },
			},
		},
	}
}

func TestEvaluate(t *testing.T) {
	t.Parallel()
	results, notEvaluated := Evaluate(testRelease(), "level-1", Input{})

	got := map[string]Result{}
	var order []string
	for _, r := range results {
		got[r.Requirement] = r.Result
		order = append(order, r.Requirement)
		if r.Evidence == nil {
			t.Errorf("%s: evidence is nil, want an empty list", r.Requirement)
		}
	}
	want := map[string]Result{
		"TEST-01.01": Pass,
		"TEST-01.02": Fail,
		"TEST-01.03": Unknown,
		"TEST-01.04": NeedsReview,
		"TEST-01.05": NotApplicable,
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("results (-want +got):\n%s", diff)
	}
	wantOrder := []string{"TEST-01.01", "TEST-01.02", "TEST-01.03", "TEST-01.04", "TEST-01.05"}
	if diff := cmp.Diff(wantOrder, order); diff != "" {
		t.Errorf("results are not in catalog order (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{"TEST-02.01"}, notEvaluated); diff != "" {
		t.Errorf("notEvaluated (-want +got):\n%s", diff)
	}
}

func TestEvaluateInsideProfile(t *testing.T) {
	t.Parallel()
	rel := testRelease()
	rel.Mappings[4].Evaluate = fixed(Pass, "The level 3 setting is on.")
	results, notEvaluated := Evaluate(rel, "level-3", Input{})
	if got := results[4]; got.Requirement != "TEST-01.05" || got.Result != Pass {
		t.Errorf("TEST-01.05 at level-3 = %s %s, want PASS", got.Requirement, got.Result)
	}
	if diff := cmp.Diff([]string{"TEST-02.01", "TEST-03.01"}, notEvaluated); diff != "" {
		t.Errorf("notEvaluated (-want +got):\n%s", diff)
	}
}

func TestReleaseProbes(t *testing.T) {
	t.Parallel()
	rel := testRelease()
	tests := []struct {
		profile string
		want    []string
	}{
		{profile: "level-1", want: []string{"probeA", "probeB"}},
		{profile: "level-3", want: []string{"probeA", "probeB", "probeLevel3"}},
	}
	for _, tt := range tests {
		t.Run(tt.profile, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, rel.Probes(tt.profile)); diff != "" {
				t.Errorf("Probes(%q) (-want +got):\n%s", tt.profile, diff)
			}
		})
	}
}

func TestDefaultBranchFinding(t *testing.T) {
	t.Parallel()
	onMain := finding.Finding{Probe: "p", Outcome: finding.OutcomeTrue, Values: map[string]string{"branchName": "main"}}
	onRelease := finding.Finding{
		Probe: "p", Outcome: finding.OutcomeFalse,
		Values: map[string]string{"branchName": "release/v1"},
	}
	noBranch := finding.Finding{Probe: "p", Outcome: finding.OutcomeNotApplicable}
	findings := []finding.Finding{onRelease, noBranch, onMain}

	tests := []struct {
		name          string
		defaultBranch string
		probe         string
		want          finding.Finding
		wantOK        bool
	}{
		{name: "default branch", defaultBranch: "main", probe: "p", want: onMain, wantOK: true},
		{name: "other branches ignored", defaultBranch: "develop", probe: "p"},
		{name: "unknown default branch", defaultBranch: "", probe: "p"},
		{name: "other probe", defaultBranch: "main", probe: "q"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := DefaultBranchFinding(Input{DefaultBranch: tt.defaultBranch, Findings: findings}, tt.probe, "branchName")
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(finding.Finding{})); diff != "" {
				t.Errorf("finding (-want +got):\n%s", diff)
			}
		})
	}
}
