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
	"github.com/ossf/scorecard/v4/finding"
	scut "github.com/ossf/scorecard/v4/utests"
)

func TestLicense(t *testing.T) {
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
					Probe:   "hasLicenseFile",
					Outcome: finding.OutcomePositive,
				},
			},
			result: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 0,
			},
		}, {
			name: "Negative outcome = Min score",
			findings: []finding.Finding{
				{
					Probe:   "hasLicenseFile",
					Outcome: finding.OutcomeNegative,
				},
			},
			result: scut.TestReturn{
				Score:        checker.MinResultScore,
				NumberOfInfo: 0,
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Parallel testing scoping hack.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			got := License(tt.name, tt.findings, &dl)
			if !scut.ValidateTestReturn(t, tt.name, &tt.result, &got, &dl) {
				t.Fatalf(tt.name)
			}
		})
	}
}
