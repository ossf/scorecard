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

//nolint:stylecheck
package hasFSFOrOSIApprovedLicense

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/internal/utils/test"
)

func Test_Run(t *testing.T) {
	t.Parallel()
	//nolint:govet
	tests := []struct {
		name     string
		raw      *checker.RawResults
		outcomes []finding.Outcome
		err      error
	}{
		{
			name: "License file found and is approved: outcome should be true",
			raw: &checker.RawResults{
				LicenseResults: checker.LicenseData{
					LicenseFiles: []checker.LicenseFile{
						{
							File: checker.File{
								Path: "LICENSE.md",
							},
							LicenseInformation: checker.License{
								Approved: true,
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue,
			},
		},
		{
			name: "License file found and is not approved: outcome should be false",
			raw: &checker.RawResults{
				LicenseResults: checker.LicenseData{
					LicenseFiles: []checker.LicenseFile{
						{
							File: checker.File{
								Path: "LICENSE.md",
							},
							LicenseInformation: checker.License{
								Approved: false,
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeFalse,
			},
		},
		{
			name: "License file not found and outcome should be false",
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
			name: "License file found but is not approved. Outcome should be False",
			raw: &checker.RawResults{
				LicenseResults: checker.LicenseData{
					LicenseFiles: []checker.LicenseFile{
						{
							File: checker.File{
								Path: "LICENSE.md",
							},
							LicenseInformation: checker.License{
								Attribution: "wrong attribution",
							},
						},
						{
							File: checker.File{
								Path: "COPYING.md",
							},
							LicenseInformation: checker.License{
								Attribution: "wrong attribution2",
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeFalse,
			},
		},
		{
			name: "nil license files and outcome should be false",
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
			name: "0 license files and outcome should be false",
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
			test.AssertOutcomes(t, findings, tt.outcomes)
		})
	}
}
