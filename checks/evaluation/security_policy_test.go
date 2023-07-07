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

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	scut "github.com/ossf/scorecard/v4/utests"
)

func TestSecurityPolicy(t *testing.T) {
	t.Parallel()
	//nolint
	type args struct {
		name string
		r    *checker.SecurityPolicyData
	}
	//nolint
	tests := []struct {
		name string
		args args
		err  bool
		want checker.CheckResult
	}{
		{
			name: "test_security_policy_1",
			args: args{
				name: "test_security_policy_1",
			},
			want: checker.CheckResult{
				Score: -1,
			},
		},
		{
			name: "test_security_policy_2",
			args: args{
				name: "test_security_policy_2",
				r:    &checker.SecurityPolicyData{},
			},
			want: checker.CheckResult{
				Score: 0,
			},
		},
		{
			name: "test_security_policy_3",
			args: args{
				name: "test_security_policy_3",
				r: &checker.SecurityPolicyData{
					PolicyFiles: []checker.SecurityPolicyFile{
						{
							File: checker.File{
								Path: "/etc/security/pam_env.conf",
								Type: finding.FileTypeURL,
							},
							Information: make([]checker.SecurityPolicyInformation, 0),
						},
					},
				},
			},
			want: checker.CheckResult{
				Score: 0,
			},
		},
		{
			name: "test_security_policy_4",
			args: args{
				name: "test_security_policy_4",
				r: &checker.SecurityPolicyData{
					PolicyFiles: []checker.SecurityPolicyFile{
						{
							File: checker.File{
								Path: "/etc/security/pam_env.conf",
							},
							Information: make([]checker.SecurityPolicyInformation, 0),
						},
					},
				},
			},
			want: checker.CheckResult{
				Score: 0,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			x := checker.CheckRequest{Dlogger: &scut.TestDetailLogger{}}

			got := SecurityPolicy(tt.args.name, x.Dlogger, tt.args.r)
			if tt.err {
				if got.Score != -1 {
					t.Errorf("SecurityPolicy() = %v, want %v", got, tt.want)
				}
			}
			if got.Score != tt.want.Score {
				t.Errorf("SecurityPolicy() = %v, want %v for %v", got.Score, tt.want.Score, tt.name)
			}
		})
	}
}

func TestScoreSecurityCriteria(t *testing.T) {
	t.Parallel()
	tests := []struct { //nolint:govet
		name          string
		file          checker.File
		info          []checker.SecurityPolicyInformation
		expectedScore int
	}{
		{
			name: "Full score",
			file: checker.File{
				Path:     "/path/to/security_policy.md",
				FileSize: 100,
			},
			info: []checker.SecurityPolicyInformation{
				{
					InformationType: checker.SecurityPolicyInformationTypeEmail,
					InformationValue: checker.SecurityPolicyValueType{
						Match:      "security@example.com",
						LineNumber: 2,
						Offset:     0,
					},
				},
				{
					InformationType: checker.SecurityPolicyInformationTypeLink,
					InformationValue: checker.SecurityPolicyValueType{
						Match:      "https://example.com/report",
						LineNumber: 4,
						Offset:     0,
					},
				},
				{
					InformationType: checker.SecurityPolicyInformationTypeText,
					InformationValue: checker.SecurityPolicyValueType{
						Match:      "Disclose vulnerability",
						LineNumber: 6,
						Offset:     0,
					},
				},
				{
					InformationType: checker.SecurityPolicyInformationTypeText,
					InformationValue: checker.SecurityPolicyValueType{
						Match:      "30 days",
						LineNumber: 7,
						Offset:     0,
					},
				},
			},
			expectedScore: 10,
		},
		{
			name: "Partial score",
			file: checker.File{
				Path:     "/path/to/security_policy.md",
				FileSize: 50,
			},
			info: []checker.SecurityPolicyInformation{
				{
					InformationType: checker.SecurityPolicyInformationTypeLink,
					InformationValue: checker.SecurityPolicyValueType{
						Match:      "https://example.com/report",
						LineNumber: 4,
						Offset:     0,
					},
				},
				{
					InformationType: checker.SecurityPolicyInformationTypeText,
					InformationValue: checker.SecurityPolicyValueType{
						Match:      "Disclose vulnerability",
						LineNumber: 6,
						Offset:     0,
					},
				},
			},
			expectedScore: 9,
		},
		{
			name: "Low score",
			file: checker.File{
				Path:     "/path/to/security_policy.md",
				FileSize: 10,
			},
			info: []checker.SecurityPolicyInformation{
				{
					InformationType: checker.SecurityPolicyInformationTypeEmail,
					InformationValue: checker.SecurityPolicyValueType{
						Match:      "security@example.com",
						LineNumber: 2,
						Offset:     0,
					},
				},
			},
			expectedScore: 6,
		},
		{
			name: "Low score",
			file: checker.File{
				Path:     "/path/to/security_policy.md",
				FileSize: 5,
			},
			info:          []checker.SecurityPolicyInformation{},
			expectedScore: 3,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mockDetailLogger := &scut.TestDetailLogger{}
			score := scoreSecurityCriteria(tc.file, tc.info, mockDetailLogger)

			if score != tc.expectedScore {
				t.Errorf("scoreSecurityCriteria() mismatch, expected score: %d, got: %d", tc.expectedScore, score)
			}
		})
	}
}
