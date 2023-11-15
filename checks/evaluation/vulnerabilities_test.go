// Copyright 2022 OpenSSF Scorecard Authors
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

	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
	scut "github.com/ossf/scorecard/v4/utests"
)

// TestVulnerabilities tests the vulnerabilities checker.
func TestVulnerabilities(t *testing.T) {
	t.Parallel()
	//nolint:govet
	tests := []struct {
		name     string
		findings []finding.Finding
		result   scut.TestReturn
		expected []struct {
			lineNumber uint
		}
	}{
		{
			name: "no vulnerabilities",
			findings: []finding.Finding{
				{
					Probe:   "hasOSVVulnerabilities",
					Outcome: finding.OutcomePositive,
				},
			},
			result: scut.TestReturn{
				Score: 10,
			},
		},
		{
			name: "three vulnerabilities",
			findings: []finding.Finding{
				{
					Probe:   "hasOSVVulnerabilities",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "hasOSVVulnerabilities",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "hasOSVVulnerabilities",
					Outcome: finding.OutcomeNegative,
				},
			},
			result: scut.TestReturn{
				Score:        7,
				NumberOfWarn: 3,
			},
		},
		{
			name: "twelve vulnerabilities to check that score is not less than 0",
			findings: []finding.Finding{
				{
					Probe:   "hasOSVVulnerabilities",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "hasOSVVulnerabilities",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "hasOSVVulnerabilities",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "hasOSVVulnerabilities",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "hasOSVVulnerabilities",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "hasOSVVulnerabilities",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "hasOSVVulnerabilities",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "hasOSVVulnerabilities",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "hasOSVVulnerabilities",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "hasOSVVulnerabilities",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "hasOSVVulnerabilities",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "hasOSVVulnerabilities",
					Outcome: finding.OutcomeNegative,
				},
			},
			result: scut.TestReturn{
				Score:        0,
				NumberOfWarn: 12,
			},
		},
		{
			name:     "invalid findings",
			findings: []finding.Finding{},
			result: scut.TestReturn{
				Score: -1,
				Error: sce.ErrScorecardInternal,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			got := Vulnerabilities(tt.name, tt.findings, &dl)
			if !scut.ValidateTestReturn(t, tt.name, &tt.result, &got, &dl) {
				t.Errorf("got %v, expected %v", got, tt.result)
			}
		})
	}
}
