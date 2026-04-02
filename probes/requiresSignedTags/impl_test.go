// Copyright 2025 OpenSSF Scorecard Authors
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

package requiresSignedTags

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/finding"
)

func Test_Run(t *testing.T) {
	t.Parallel()
	trueVal := true
	falseVal := false
	tagVal1 := "v1.0.0"
	tagVal2 := "v2.0.0"

	tests := []struct {
		err      error
		raw      *checker.RawResults
		name     string
		outcomes []finding.Outcome
	}{
		{
			name: "No release tags",
			raw: &checker.RawResults{
				TagProtectionResults: checker.TagProtectionsData{
					Tags: []clients.TagRef{},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeNotApplicable,
			},
		},
		{
			name: "Signed tags required",
			raw: &checker.RawResults{
				TagProtectionResults: checker.TagProtectionsData{
					Tags: []clients.TagRef{
						{
							Name:      &tagVal1,
							Protected: &trueVal,
							TagProtectionRule: clients.TagProtectionRule{
								RequireSignatures: &trueVal,
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
			name: "Signed tags not required",
			raw: &checker.RawResults{
				TagProtectionResults: checker.TagProtectionsData{
					Tags: []clients.TagRef{
						{
							Name:      &tagVal1,
							Protected: &trueVal,
							TagProtectionRule: clients.TagProtectionRule{
								RequireSignatures: &falseVal,
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
			name: "Unprotected tag",
			raw: &checker.RawResults{
				TagProtectionResults: checker.TagProtectionsData{
					Tags: []clients.TagRef{
						{
							Name:      &tagVal1,
							Protected: &falseVal,
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeFalse,
			},
		},
		{
			name: "Protected tag with nil RequireSignatures",
			raw: &checker.RawResults{
				TagProtectionResults: checker.TagProtectionsData{
					Tags: []clients.TagRef{
						{
							Name:      &tagVal1,
							Protected: &trueVal,
							TagProtectionRule: clients.TagProtectionRule{
								RequireSignatures: nil,
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
			name: "Multiple tags mixed signature requirements",
			raw: &checker.RawResults{
				TagProtectionResults: checker.TagProtectionsData{
					Tags: []clients.TagRef{
						{
							Name:      &tagVal1,
							Protected: &trueVal,
							TagProtectionRule: clients.TagProtectionRule{
								RequireSignatures: &trueVal,
							},
						},
						{
							Name:      &tagVal2,
							Protected: &trueVal,
							TagProtectionRule: clients.TagProtectionRule{
								RequireSignatures: &falseVal,
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue,
				finding.OutcomeFalse,
			},
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
			if diff := cmp.Diff(len(tt.outcomes), len(findings)); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
			for i := range tt.outcomes {
				if diff := cmp.Diff(tt.outcomes[i], findings[i].Outcome); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
