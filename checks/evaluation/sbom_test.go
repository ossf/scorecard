// Copyright 2024 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
	scut "github.com/ossf/scorecard/v5/utests"
)

func TestSBOM(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		findings []finding.Finding
		result   scut.TestReturn
	}{
		{
			name: "No SBOM. Min Score",
			findings: []finding.Finding{
				{
					Probe:   "hasSBOM",
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   "hasReleaseSBOM",
					Outcome: finding.OutcomeFalse,
				},
			},
			result: scut.TestReturn{
				Score:        checker.MinResultScore,
				NumberOfInfo: 0,
				NumberOfWarn: 2,
			},
		},
		{
			name: "Only Source SBOM. Half Points",
			findings: []finding.Finding{
				{
					Probe:   "hasSBOM",
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   "hasReleaseSBOM",
					Outcome: finding.OutcomeFalse,
				},
			},
			result: scut.TestReturn{
				Score:        5,
				NumberOfInfo: 1,
				NumberOfWarn: 1,
			},
		},
		{
			name: "SBOM in Release Assets. Max score",
			findings: []finding.Finding{
				{
					Probe:   "hasSBOM",
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   "hasReleaseSBOM",
					Outcome: finding.OutcomeTrue,
				},
			},
			result: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 2,
				NumberOfWarn: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			got := SBOM(tt.name, tt.findings, &dl)
			scut.ValidateTestReturn(t, tt.name, &tt.result, &got, &dl)
		})
	}
}
