// Copyright 2025 OpenSSF Scorecard Authors
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

package evaluation

import (
	"testing"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/MTTUDependenciesIsHigh"
	"github.com/ossf/scorecard/v5/probes/MTTUDependenciesIsLow"
	"github.com/ossf/scorecard/v5/probes/MTTUDependenciesIsVeryLow"
)

func f(probe string, outcome finding.Outcome) finding.Finding {
	return finding.Finding{Probe: probe, Outcome: outcome}
}

//nolint:paralleltest // test uses shared state
func TestEvaluation_Scoring(t *testing.T) {
	tests := []struct {
		name     string
		findings []finding.Finding
		want     int
	}{
		{
			name: "high-true → score 0",
			findings: []finding.Finding{
				f(MTTUDependenciesIsHigh.Probe, finding.OutcomeTrue),
				f(MTTUDependenciesIsLow.Probe, finding.OutcomeFalse),
				f(MTTUDependenciesIsVeryLow.Probe, finding.OutcomeFalse),
			},
			want: 0,
		},
		{
			name: "low-true → score 5",
			findings: []finding.Finding{
				f(MTTUDependenciesIsHigh.Probe, finding.OutcomeFalse),
				f(MTTUDependenciesIsLow.Probe, finding.OutcomeTrue),
				f(MTTUDependenciesIsVeryLow.Probe, finding.OutcomeFalse),
			},
			want: 5,
		},
		{
			name: "verylow-true → score 10",
			findings: []finding.Finding{
				f(MTTUDependenciesIsHigh.Probe, finding.OutcomeFalse),
				f(MTTUDependenciesIsLow.Probe, finding.OutcomeFalse),
				f(MTTUDependenciesIsVeryLow.Probe, finding.OutcomeTrue),
			},
			want: 10,
		},
	}

	//nolint:paralleltest // test cases use shared state
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var dl checker.DetailLogger = nil
			res := MTTUDependencies(CheckMTTUDependencies, tc.findings, dl)
			if res.Score != tc.want {
				t.Fatalf("want score %d, got %d", tc.want, res.Score)
			}
		})
	}
}

//nolint:paralleltest // test expected to panic
func TestEvaluation_InvalidProbeSet(t *testing.T) {
	// Missing one probe → runtime error expected.
	ff := []finding.Finding{
		f(MTTUDependenciesIsHigh.Probe, finding.OutcomeFalse),
		f(MTTUDependenciesIsLow.Probe, finding.OutcomeTrue),
		// verylow missing
	}
	var dl checker.DetailLogger = nil
	res := MTTUDependencies(CheckMTTUDependencies, ff, dl)
	if res.Error == nil {
		t.Fatalf("expected runtime error for invalid probe set, got nil")
	}
}
