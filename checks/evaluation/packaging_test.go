// Copyright 2023 OpenSSF Scorecard Authors
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
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
	scut "github.com/ossf/scorecard/v4/utests"
)

func TestPackaging(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		findings []finding.Finding
		result   scut.TestReturn
	}{
		{
			name: "test positive outcome",
			findings: []finding.Finding{
				{
					Probe:   "packagedNpmWithGitHubWorkflow",
					Outcome: finding.OutcomePositive,
				},
			},
			result: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 1,
			},
		},
		{
			name: "test positive outcome with wrong probes",
			findings: []finding.Finding{
				{
					Probe:   "wrongProbe",
					Outcome: finding.OutcomePositive,
				},
			},
			result: scut.TestReturn{
				Score: -1,
				Error: sce.ErrScorecardInternal,
			},
		},
		{
			name: "test negative outcome",
			findings: []finding.Finding{
				{
					Probe:   "packagedNpmWithGitHubWorkflow",
					Outcome: finding.OutcomeNegative,
				},
			},
			result: scut.TestReturn{
				Score: checker.MinResultScore,
			},
		},
		{
			name: "test negative outcome with wrong probes",
			findings: []finding.Finding{
				{
					Probe:   "wrongProbe",
					Outcome: finding.OutcomeNegative,
				},
			},
			result: scut.TestReturn{
				Score: -1,
				Error: sce.ErrScorecardInternal,
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			got := Packaging(tt.name, tt.findings, &dl)
			if !scut.ValidateTestReturn(t, tt.name, &tt.result, &got, &dl) {
				t.Errorf("got %v, expected %v", got, tt.result)
			}
		})
	}
}
