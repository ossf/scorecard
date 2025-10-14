// Copyright 2026 OpenSSF Scorecard Authors
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

func TestInactiveMaintainers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		findings []finding.Finding
		result   scut.TestReturn
	}{
		{
			name: "All maintainers active",
			findings: []finding.Finding{
				{
					Probe:   "hasInactiveMaintainers",
					Outcome: finding.OutcomeFalse,
					Values: map[string]string{
						"username": "active-maintainer-1",
					},
				},
				{
					Probe:   "hasInactiveMaintainers",
					Outcome: finding.OutcomeFalse,
					Values: map[string]string{
						"username": "active-maintainer-2",
					},
				},
			},
			result: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 2,
				NumberOfWarn: 0,
			},
		},
		{
			name: "All maintainers inactive",
			findings: []finding.Finding{
				{
					Probe:   "hasInactiveMaintainers",
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						"username": "inactive-maintainer-1",
					},
				},
				{
					Probe:   "hasInactiveMaintainers",
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						"username": "inactive-maintainer-2",
					},
				},
			},
			result: scut.TestReturn{
				Score:        checker.MinResultScore,
				NumberOfWarn: 2,
				NumberOfInfo: 0,
			},
		},
		{
			name: "Mixed active and inactive maintainers",
			findings: []finding.Finding{
				{
					Probe:   "hasInactiveMaintainers",
					Outcome: finding.OutcomeFalse,
					Values: map[string]string{
						"username": "active-maintainer",
					},
				},
				{
					Probe:   "hasInactiveMaintainers",
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						"username": "inactive-maintainer",
					},
				},
			},
			result: scut.TestReturn{
				Score:        5, // 50% active
				NumberOfInfo: 1,
				NumberOfWarn: 1,
			},
		},
		{
			name: "No maintainers found",
			findings: []finding.Finding{
				{
					Probe:   "hasInactiveMaintainers",
					Outcome: finding.OutcomeNotApplicable,
				},
			},
			result: scut.TestReturn{
				Score: checker.InconclusiveResultScore,
			},
		},
		{
			name: "Three out of four maintainers active",
			findings: []finding.Finding{
				{
					Probe:   "hasInactiveMaintainers",
					Outcome: finding.OutcomeFalse,
					Values: map[string]string{
						"username": "active-1",
					},
				},
				{
					Probe:   "hasInactiveMaintainers",
					Outcome: finding.OutcomeFalse,
					Values: map[string]string{
						"username": "active-2",
					},
				},
				{
					Probe:   "hasInactiveMaintainers",
					Outcome: finding.OutcomeFalse,
					Values: map[string]string{
						"username": "active-3",
					},
				},
				{
					Probe:   "hasInactiveMaintainers",
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						"username": "inactive-1",
					},
				},
			},
			result: scut.TestReturn{
				Score:        7, // 75% active = 7.5, truncated to 7
				NumberOfInfo: 3,
				NumberOfWarn: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			got := InactiveMaintainers(tt.name, tt.findings, &dl)
			scut.ValidateTestReturn(t, tt.name, &tt.result, &got, &dl)
		})
	}
}
