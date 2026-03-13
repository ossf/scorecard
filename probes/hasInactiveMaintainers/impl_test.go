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

package hasInactiveMaintainers

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/internal/utils/test"
	"github.com/ossf/scorecard/v5/probes/internal/utils/uerror"
)

func Test_Run(t *testing.T) {
	t.Parallel()
	tests := []struct {
		err      error
		raw      *checker.RawResults
		name     string
		outcomes []finding.Outcome
	}{
		{
			name: "All maintainers active",
			raw: &checker.RawResults{
				MaintainedResults: checker.MaintainedData{
					MaintainerActivity: map[string]bool{
						"active-maintainer-1": true,
						"active-maintainer-2": true,
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeFalse,
				finding.OutcomeFalse,
			},
		},
		{
			name: "All maintainers inactive",
			raw: &checker.RawResults{
				MaintainedResults: checker.MaintainedData{
					MaintainerActivity: map[string]bool{
						"inactive-maintainer-1": false,
						"inactive-maintainer-2": false,
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue,
				finding.OutcomeTrue,
			},
		},
		{
			name: "Mixed active and inactive maintainers",
			raw: &checker.RawResults{
				MaintainedResults: checker.MaintainedData{
					MaintainerActivity: map[string]bool{
						"active-maintainer":   true,
						"inactive-maintainer": false,
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeFalse,
				finding.OutcomeTrue,
			},
		},
		{
			name: "No maintainers found",
			raw: &checker.RawResults{
				MaintainedResults: checker.MaintainedData{
					MaintainerActivity: map[string]bool{},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeNotApplicable,
			},
		},
		{
			name: "Single active maintainer",
			raw: &checker.RawResults{
				MaintainedResults: checker.MaintainedData{
					MaintainerActivity: map[string]bool{
						"sole-maintainer": true,
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeFalse,
			},
		},
		{
			name: "Single inactive maintainer",
			raw: &checker.RawResults{
				MaintainedResults: checker.MaintainedData{
					MaintainerActivity: map[string]bool{
						"sole-maintainer": false,
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue,
			},
		},
		{
			name: "Nil raw results",
			raw:  nil,
			err:  uerror.ErrNil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			findings, s, err := Run(tt.raw)
			if !cmp.Equal(tt.err, err, cmpopts.EquateErrors()) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(tt.err, err, cmpopts.EquateErrors()))
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(Probe, s); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
			test.AssertOutcomes(t, findings, tt.outcomes)
		})
	}
}
