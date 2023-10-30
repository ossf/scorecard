// Copyright 2023 OpenSSF Scorecard Authors
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

// nolint:stylecheck
package hasLicenseFileAtTopDir

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
)

func Test_Run(t *testing.T) {
	t.Parallel()
	// nolint:govet
	tests := []struct {
		name     string
		raw      *checker.RawResults
		outcomes []finding.Outcome
		err      error
	}{
		{
			name: "License file found and correct attribution 1",
			raw: &checker.RawResults{
				LicenseResults: checker.LicenseData{
					LicenseFiles: []checker.LicenseFile{
						{
							File: checker.File{
								Path: "LICENSE.md",
							},
							LicenseInformation: checker.License{
								Attribution: checker.LicenseAttributionTypeAPI,
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomePositive,
			},
		},
		{
			name: "License file not found and outcome should be OutcomeNotApplicable",
			raw: &checker.RawResults{
				LicenseResults: checker.LicenseData{
					LicenseFiles: []checker.LicenseFile{},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeNotApplicable,
			},
		},
		{
			name: "License file found and correct attribution 2",
			raw: &checker.RawResults{
				LicenseResults: checker.LicenseData{
					LicenseFiles: []checker.LicenseFile{
						{
							File: checker.File{
								Path: "LICENSE.md",
							},
							LicenseInformation: checker.License{
								Attribution: checker.LicenseAttributionTypeHeuristics,
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomePositive,
			},
		},
		{
			name: "License file found and wrong attribution",
			raw: &checker.RawResults{
				LicenseResults: checker.LicenseData{
					LicenseFiles: []checker.LicenseFile{
						{
							File: checker.File{
								Path: "LICENSE.md",
							},
							LicenseInformation: checker.License{
								Attribution: "wrong_attribution",
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeNegative,
			},
		},
		{
			name: "nil license files and outcome should be OutcomeNotApplicable",
			raw: &checker.RawResults{
				LicenseResults: checker.LicenseData{
					LicenseFiles: nil,
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeNotApplicable,
			},
		},
		{
			name: "0 license files and outcome should be OutcomeNotApplicable",
			raw: &checker.RawResults{
				LicenseResults: checker.LicenseData{
					LicenseFiles: []checker.LicenseFile{},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeNotApplicable,
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
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
			if diff := cmp.Diff(len(tt.outcomes), len(findings)); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
			for i := range tt.outcomes {
				outcome := &tt.outcomes[i]
				f := &findings[i]
				if diff := cmp.Diff(*outcome, f.Outcome); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
