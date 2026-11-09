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
	"testing"

	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/conformance"
	"github.com/ossf/scorecard/v5/probes/blocksDeleteOnBranches"
	"github.com/ossf/scorecard/v5/probes/branchProtectionAppliesToAdmins"
	"github.com/ossf/scorecard/v5/probes/requiresPRsToChangeCode"
)

func branchFinding(probe, branch string, outcome finding.Outcome) finding.Finding {
	return finding.Finding{
		Probe:   probe,
		Outcome: outcome,
		Message: probe + " on " + branch,
		Values:  map[string]string{"branchName": branch},
	}
}

func noBranches(probe string) finding.Finding {
	return finding.Finding{Probe: probe, Outcome: finding.OutcomeNotApplicable, Message: "no branches found"}
}

const (
	prs    = requiresPRsToChangeCode.Probe
	admins = branchProtectionAppliesToAdmins.Probe
	del    = blocksDeleteOnBranches.Probe
)

func TestAC0301(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		defaultBranch string
		want          conformance.Result
		findings      []finding.Finding
		wantEvidence  int
	}{
		{
			name:          "protected for everyone",
			defaultBranch: "main",
			findings: []finding.Finding{
				branchFinding(prs, "main", finding.OutcomeTrue),
				branchFinding(admins, "main", finding.OutcomeTrue),
			},
			want:         conformance.Pass,
			wantEvidence: 2,
		},
		{
			name:          "pull requests not required",
			defaultBranch: "main",
			findings: []finding.Finding{
				branchFinding(prs, "main", finding.OutcomeFalse),
				branchFinding(admins, "main", finding.OutcomeTrue),
			},
			want:         conformance.Fail,
			wantEvidence: 2,
		},
		{
			name:          "pull requests not required, admin finding missing",
			defaultBranch: "main",
			findings:      []finding.Finding{branchFinding(prs, "main", finding.OutcomeFalse)},
			want:          conformance.Fail,
			wantEvidence:  1,
		},
		{
			name:          "administrators can bypass",
			defaultBranch: "main",
			findings: []finding.Finding{
				branchFinding(prs, "main", finding.OutcomeTrue),
				branchFinding(admins, "main", finding.OutcomeFalse),
			},
			want:         conformance.NeedsReview,
			wantEvidence: 2,
		},
		{
			name:          "administrator setting not available",
			defaultBranch: "main",
			findings: []finding.Finding{
				branchFinding(prs, "main", finding.OutcomeTrue),
				branchFinding(admins, "main", finding.OutcomeNotAvailable),
			},
			want:         conformance.Unknown,
			wantEvidence: 2,
		},
		{
			name:          "administrator finding missing",
			defaultBranch: "main",
			findings:      []finding.Finding{branchFinding(prs, "main", finding.OutcomeTrue)},
			want:          conformance.Unknown,
			wantEvidence:  1,
		},
		{
			name:          "probe could not read settings",
			defaultBranch: "main",
			findings: []finding.Finding{
				branchFinding(prs, "main", finding.OutcomeNotAvailable),
				branchFinding(admins, "main", finding.OutcomeTrue),
			},
			want:         conformance.Unknown,
			wantEvidence: 2,
		},
		{
			name:          "probe errored",
			defaultBranch: "main",
			findings:      []finding.Finding{branchFinding(prs, "main", finding.OutcomeError)},
			want:          conformance.Unknown,
			wantEvidence:  1,
		},
		{
			name:          "platform not supported",
			defaultBranch: "main",
			findings:      []finding.Finding{branchFinding(prs, "main", finding.OutcomeNotSupported)},
			want:          conformance.Unknown,
			wantEvidence:  1,
		},
		{
			name:          "probe found no branches",
			defaultBranch: "main",
			findings:      []finding.Finding{noBranches(prs), noBranches(admins)},
			want:          conformance.Unknown,
			wantEvidence:  1,
		},
		{
			name:          "only another branch is unprotected",
			defaultBranch: "main",
			findings: []finding.Finding{
				branchFinding(prs, "release/v1", finding.OutcomeFalse),
				branchFinding(prs, "main", finding.OutcomeTrue),
				branchFinding(admins, "release/v1", finding.OutcomeFalse),
				branchFinding(admins, "main", finding.OutcomeTrue),
			},
			want:         conformance.Pass,
			wantEvidence: 2,
		},
		{
			name:          "default branch unknown",
			defaultBranch: "unknown",
			findings: []finding.Finding{
				branchFinding(prs, "main", finding.OutcomeTrue),
				branchFinding(admins, "main", finding.OutcomeTrue),
			},
			want: conformance.Unknown,
		},
		{
			name:          "default branch empty",
			defaultBranch: "",
			findings:      []finding.Finding{branchFinding(prs, "main", finding.OutcomeTrue)},
			want:          conformance.Unknown,
		},
	}
	m := AC0301Mapping()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := m.Evaluate(conformance.Input{DefaultBranch: tt.defaultBranch, Findings: tt.findings})
			if got.Result != tt.want {
				t.Errorf("result = %s (%s), want %s", got.Result, got.Explanation, tt.want)
			}
			if len(got.Evidence) != tt.wantEvidence {
				t.Errorf("evidence has %d findings, want %d: %v", len(got.Evidence), tt.wantEvidence, got.Evidence)
			}
			if got.Explanation == "" {
				t.Error("explanation is empty")
			}
		})
	}
}

func TestAC0302(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		want     conformance.Result
		findings []finding.Finding
	}{
		{
			name:     "deletion blocked",
			findings: []finding.Finding{branchFinding(del, "main", finding.OutcomeTrue)},
			want:     conformance.Pass,
		},
		{
			name:     "deletion allowed",
			findings: []finding.Finding{branchFinding(del, "main", finding.OutcomeFalse)},
			want:     conformance.Fail,
		},
		{
			name:     "setting not available",
			findings: []finding.Finding{branchFinding(del, "main", finding.OutcomeNotAvailable)},
			want:     conformance.Unknown,
		},
		{
			name:     "probe found no branches",
			findings: []finding.Finding{noBranches(del)},
			want:     conformance.Unknown,
		},
		{
			name: "only another branch allows deletion",
			findings: []finding.Finding{
				branchFinding(del, "release/v1", finding.OutcomeFalse),
				branchFinding(del, "main", finding.OutcomeTrue),
			},
			want: conformance.Pass,
		},
		{
			name:     "no finding for the default branch",
			findings: []finding.Finding{branchFinding(del, "release/v1", finding.OutcomeTrue)},
			want:     conformance.Unknown,
		},
	}
	m := AC0302Mapping()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := m.Evaluate(conformance.Input{DefaultBranch: "main", Findings: tt.findings})
			if got.Result != tt.want {
				t.Errorf("result = %s (%s), want %s", got.Result, got.Explanation, tt.want)
			}
			if got.Explanation == "" {
				t.Error("explanation is empty")
			}
		})
	}
}
