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

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	scut "github.com/ossf/scorecard/v4/utests"
)

func TestSbom(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		findings []finding.Finding
		result   scut.TestReturn
	}{
		{
			name: "Negative outcome = Min Score",
			findings: []finding.Finding{
				{
					Probe:   "sbomExists",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "sbomReleaseAssetExists",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "sbomStandardsFileUsed",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "sbomCICDArtifactExists",
					Outcome: finding.OutcomeNegative,
				},
			},
			result: scut.TestReturn{
				Score:        checker.MinResultScore,
				NumberOfInfo: 0,
				NumberOfWarn: 4,
			},
		},
		{
			name: "Exists in Source: Positive outcome.",
			findings: []finding.Finding{
				{
					Probe:   "sbomExists",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "sbomReleaseAssetExists",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "sbomStandardsFileUsed",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "sbomCICDArtifactExists",
					Outcome: finding.OutcomeNegative,
				},
			},
			result: scut.TestReturn{
				Score:        3,
				NumberOfInfo: 1,
				NumberOfWarn: 3,
			},
		},
		{
			name: "Exists in Release Assets: Positive outcome.",
			findings: []finding.Finding{
				{
					Probe:   "sbomExists",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "sbomReleaseAssetExists",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "sbomStandardsFileUsed",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "sbomCICDArtifactExists",
					Outcome: finding.OutcomeNegative,
				},
			},
			result: scut.TestReturn{
				Score:        6,
				NumberOfInfo: 2,
				NumberOfWarn: 2,
			},
		},
		{
			name: "Exists in Standards File: Positive outcome.",
			findings: []finding.Finding{
				{
					Probe:   "sbomExists",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "sbomReleaseAssetExists",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "sbomStandardsFileUsed",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "sbomCICDArtifactExists",
					Outcome: finding.OutcomeNegative,
				},
			},
			result: scut.TestReturn{
				Score:        4,
				NumberOfInfo: 2,
				NumberOfWarn: 2,
			},
		},
		{
			name: "Exists in CICD Artifacts: Positive outcome.",
			findings: []finding.Finding{
				{
					Probe:   "sbomExists",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "sbomReleaseAssetExists",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "sbomStandardsFileUsed",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "sbomCICDArtifactExists",
					Outcome: finding.OutcomePositive,
				},
			},
			result: scut.TestReturn{
				Score:        6,
				NumberOfInfo: 2,
				NumberOfWarn: 2,
			},
		},
		{
			name: "Exists in Release Assets and Standards File: Positive outcome.",
			findings: []finding.Finding{
				{
					Probe:   "sbomExists",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "sbomReleaseAssetExists",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "sbomStandardsFileUsed",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "sbomCICDArtifactExists",
					Outcome: finding.OutcomeNegative,
				},
			},
			result: scut.TestReturn{
				Score:        7,
				NumberOfInfo: 3,
				NumberOfWarn: 1,
			},
		},
		{
			name: "Exists in Release Assets and CICD Artifacts: Positive outcome.",
			findings: []finding.Finding{
				{
					Probe:   "sbomExists",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "sbomReleaseAssetExists",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "sbomStandardsFileUsed",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "sbomCICDArtifactExists",
					Outcome: finding.OutcomePositive,
				},
			},
			result: scut.TestReturn{
				Score:        9,
				NumberOfInfo: 3,
				NumberOfWarn: 1,
			},
		},
		{
			name: "Exists in CICD Artifacts and Standards File: Positive outcome.",
			findings: []finding.Finding{
				{
					Probe:   "sbomExists",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "sbomReleaseAssetExists",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "sbomStandardsFileUsed",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "sbomCICDArtifactExists",
					Outcome: finding.OutcomePositive,
				},
			},
			result: scut.TestReturn{
				Score:        7,
				NumberOfInfo: 3,
				NumberOfWarn: 1,
			},
		},
		{
			name: "Positive outcome = Max Score",
			findings: []finding.Finding{
				{
					Probe:   "sbomExists",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "sbomReleaseAssetExists",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "sbomStandardsFileUsed",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "sbomCICDArtifactExists",
					Outcome: finding.OutcomePositive,
				},
			},
			result: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 4,
				NumberOfWarn: 0,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			got := Sbom(tt.name, tt.findings, &dl)
			scut.ValidateTestReturn(t, tt.name, &tt.result, &got, &dl)
		})
	}
}
