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
			name: "Positive outcome = Max Score",
			findings: []finding.Finding{
				{
					Probe:   "hasSbomFile",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "hasSbomReleaseArtifact",
					Outcome: finding.OutcomePositive,
				},
			},
			result: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 2,
			},
		}, {
			name: "Negative outcomes from all probes = Min score",
			findings: []finding.Finding{
				{
					Probe:   "hasLicenseFile",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "hasSbomReleaseArtifact",
					Outcome: finding.OutcomeNegative,
				},
			},
			result: scut.TestReturn{
				Score:        checker.MinResultScore,
				NumberOfWarn: 2,
			},
		}, {
			name: "Has license file but not a top level or in OSI/FSF format",
			findings: []finding.Finding{
				{
					Probe:   "hasLicenseFile",
					Outcome: finding.OutcomePositive,
				},
				{
					Probe:   "hasSbomReleaseArtifact",
					Outcome: finding.OutcomeNegative,
				},
			},
			result: scut.TestReturn{
				Score:        3,
				NumberOfWarn: 1,
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Parallel testing scoping hack.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
		})
	}
}
