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
package securityPolicyContainsLinks

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/internal/utils"
)

func errCmp(e1, e2 error) bool {
	return errors.Is(e1, e2) || errors.Is(e2, e1)
}

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
			name: "file present on repo no link",
			raw: &checker.RawResults{
				SecurityPolicyResults: checker.SecurityPolicyData{
					PolicyFiles: []checker.SecurityPolicyFile{
						{
							File: checker.File{
								Path: "SECURITY.md",
								Type: finding.FileTypeText,
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
			name: "file present on repo link",
			raw: &checker.RawResults{
				SecurityPolicyResults: checker.SecurityPolicyData{
					PolicyFiles: []checker.SecurityPolicyFile{
						{
							File: checker.File{
								Path: "SECURITY.md",
								Type: finding.FileTypeText,
							},
							Information: []checker.SecurityPolicyInformation{
								{
									InformationType: checker.SecurityPolicyInformationTypeLink,
									InformationValue: checker.SecurityPolicyValueType{
										Match: "https://www.bla.com",
									},
								},
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
			name: "file present on repo email",
			raw: &checker.RawResults{
				SecurityPolicyResults: checker.SecurityPolicyData{
					PolicyFiles: []checker.SecurityPolicyFile{
						{
							File: checker.File{
								Path: "SECURITY.md",
								Type: finding.FileTypeText,
							},
							Information: []checker.SecurityPolicyInformation{
								{
									InformationType: checker.SecurityPolicyInformationTypeEmail,
									InformationValue: checker.SecurityPolicyValueType{
										Match: "hey@google.com",
									},
								},
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
			name: "file present on org no link",
			raw: &checker.RawResults{
				SecurityPolicyResults: checker.SecurityPolicyData{
					PolicyFiles: []checker.SecurityPolicyFile{
						{
							File: checker.File{
								Path: "SECURITY.md",
								Type: finding.FileTypeText,
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
			name: "file present on org link",
			raw: &checker.RawResults{
				SecurityPolicyResults: checker.SecurityPolicyData{
					PolicyFiles: []checker.SecurityPolicyFile{
						{
							File: checker.File{
								Path: "SECURITY.md",
								Type: finding.FileTypeURL,
							},
							Information: []checker.SecurityPolicyInformation{
								{
									InformationType: checker.SecurityPolicyInformationTypeLink,
									InformationValue: checker.SecurityPolicyValueType{
										Match: "https://www.bla.com",
									},
								},
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
			name: "file present on org email",
			raw: &checker.RawResults{
				SecurityPolicyResults: checker.SecurityPolicyData{
					PolicyFiles: []checker.SecurityPolicyFile{
						{
							File: checker.File{
								Path: "SECURITY.md",
								Type: finding.FileTypeURL,
							},
							Information: []checker.SecurityPolicyInformation{
								{
									InformationType: checker.SecurityPolicyInformationTypeEmail,
									InformationValue: checker.SecurityPolicyValueType{
										Match: "hey@google.com",
									},
								},
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
			name: "files present on org and repo",
			raw: &checker.RawResults{
				SecurityPolicyResults: checker.SecurityPolicyData{
					PolicyFiles: []checker.SecurityPolicyFile{
						{
							File: checker.File{
								Path: "SECURITY.md",
								Type: finding.FileTypeURL,
							},
						},
						{
							File: checker.File{
								Path: "SECURITY.md",
								Type: finding.FileTypeText,
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeNegative,
				finding.OutcomeNegative,
			},
		},
		{
			name: "files present on org and repo email",
			raw: &checker.RawResults{
				SecurityPolicyResults: checker.SecurityPolicyData{
					PolicyFiles: []checker.SecurityPolicyFile{
						{
							File: checker.File{
								Path: "SECURITY.md",
								Type: finding.FileTypeURL,
							},
							Information: []checker.SecurityPolicyInformation{
								{
									InformationType: checker.SecurityPolicyInformationTypeEmail,
									InformationValue: checker.SecurityPolicyValueType{
										Match: "hey@google.com",
									},
								},
							},
						},
						{
							File: checker.File{
								Path: "SECURITY.md",
								Type: finding.FileTypeText,
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomePositive,
				finding.OutcomeNegative,
			},
		},
		{
			name: "files present on org and repo link",
			raw: &checker.RawResults{
				SecurityPolicyResults: checker.SecurityPolicyData{
					PolicyFiles: []checker.SecurityPolicyFile{
						{
							File: checker.File{
								Path: "SECURITY.md",
								Type: finding.FileTypeURL,
							},
							Information: []checker.SecurityPolicyInformation{
								{
									InformationType: checker.SecurityPolicyInformationTypeLink,
									InformationValue: checker.SecurityPolicyValueType{
										Match: "https://www.bla.com",
									},
								},
							},
						},
						{
							File: checker.File{
								Path: "SECURITY.md",
								Type: finding.FileTypeText,
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomePositive,
				finding.OutcomeNegative,
			},
		},
		{
			name: "file not present",
			raw:  &checker.RawResults{},
			outcomes: []finding.Outcome{
				finding.OutcomeNegative,
			},
		},
		{
			name: "nil raw",
			err:  utils.ErrNil,
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
				finding := &findings[i]
				if diff := cmp.Diff(*outcome, finding.Outcome); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
