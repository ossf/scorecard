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

package raw

import (
	"io"
	"os"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	scut "github.com/ossf/scorecard/v5/utests"
)

func Test_isSecurityPolicyFilename(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{
			name:     "test1",
			filename: "test1",
			expected: false,
		},
		{
			name:     "docs/security.rst",
			filename: "docs/security.rst",
			expected: true,
		},
		{
			name:     "doc/security.rst",
			filename: "doc/security.rst",
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isSecurityPolicyFilename(tt.filename); got != tt.expected {
				t.Errorf("isSecurityPolicyFilename() = %v, want %v for %v", got, tt.expected, tt.name)
			}
		})
	}
}

// TestSecurityPolicy tests the security policy.
func TestSecurityPolicy(t *testing.T) {
	t.Parallel()
	//nolint:govet
	tests := []struct {
		name    string
		files   []string
		path    string
		result  checker.SecurityPolicyData
		wantErr bool
		want    scut.TestReturn
	}{
		{
			name: "security.md",
			files: []string{
				"security.md",
			},
			path: "",
		},
		{
			name: ".github/security.md",
			files: []string{
				".github/security.md",
			},
			path: "",
		},
		{
			name: "docs/security.md",
			files: []string{
				"docs/security.md",
			},
			path: "",
		},
		{
			name: "security.markdown",
			files: []string{
				"security.markdown",
			},
			path: "",
		},
		{
			name: ".github/security.markdown",
			files: []string{
				".github/security.markdown",
			},
			path: "",
		},
		{
			name: "docs/security.markdown",
			files: []string{
				"docs/security.markdown",
			},
			path: "",
		},
		{
			name: "docs/security.rst",
			files: []string{
				"docs/security.rst",
			},
			path: "",
		},
		{
			name: "doc/security.rst",
			files: []string{
				"doc/security.rst",
			},
			path: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
			mockRepo := mockrepo.NewMockRepo(ctrl)

			mockRepoClient.EXPECT().ListFiles(gomock.Any()).Return(tt.files, nil).AnyTimes()
			// the revised Security Policy will immediate go for the
			// file contents once found. This test will return that
			// mock file, but this specific unit test is not testing
			// for content. As such, this test will crash without
			// a mock GetFileReader, so this will return no content
			// for the existing file. content test are in overall check
			//
			mockRepoClient.EXPECT().GetFileReader(gomock.Any()).DoAndReturn(func(fn string) (io.ReadCloser, error) {
				if tt.path == "" {
					return io.NopCloser(strings.NewReader("")), nil
				}
				return os.Open(tt.path)
			}).AnyTimes()

			mockRepoClient.EXPECT().HasPrivateVulnerabilityReportingEnabled().Return(true, nil).AnyTimes()

			dl := scut.TestDetailLogger{}
			c := checker.CheckRequest{
				RepoClient: mockRepoClient,
				Repo:       mockRepo,
				Dlogger:    &dl,
			}

			res, err := SecurityPolicy(&c)

			scut.ValidateTestReturn(t, tt.name, &tt.want, &checker.CheckResult{}, &dl)

			if (err != nil) != tt.wantErr {
				t.Errorf("SecurityPolicy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if (res.PolicyFiles[0].File.Path) != (tt.files[0]) {
				t.Errorf("test failed: the file returned is not correct: %+v", res)
			}
		})
	}
}

// Test_collectPolicyHits tests the regexes in collectPolicyHits for positive and negative cases.
func Test_collectPolicyHits(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected []checker.SecurityPolicyInformation
	}{
		// URL regex
		{
			name:  "URL positive",
			input: "See https://example.com for details.",
			expected: []checker.SecurityPolicyInformation{
				{
					InformationType: checker.SecurityPolicyInformationTypeLink,
					InformationValue: checker.SecurityPolicyValueType{
						Match:      "https://example.com",
						LineNumber: 1,
						Offset:     4,
					},
				},
			},
		},
		{
			name:     "URL negative",
			input:    "No links here.",
			expected: nil,
		},
		// Email regex
		{
			name:  "Email positive with @",
			input: "Contact us at security@example.org.",
			expected: []checker.SecurityPolicyInformation{
				{
					InformationType: checker.SecurityPolicyInformationTypeEmail,
					InformationValue: checker.SecurityPolicyValueType{
						Match:      "security@example.org",
						LineNumber: 1,
						Offset:     14,
					},
				},
			},
		},
		{
			name:  "Email positive with [at] unescaped",
			input: "Contact us at security[at]example.org.",
			expected: []checker.SecurityPolicyInformation{
				{
					InformationType: checker.SecurityPolicyInformationTypeEmail,
					InformationValue: checker.SecurityPolicyValueType{
						Match:      "security[at]example.org",
						LineNumber: 1,
						Offset:     14,
					},
				},
			},
		},
		{
			name:  "Email positive with [at] escaped",
			input: "Contact us at security\\[at\\]example.org.",
			expected: []checker.SecurityPolicyInformation{
				{
					InformationType: checker.SecurityPolicyInformationTypeEmail,
					InformationValue: checker.SecurityPolicyValueType{
						Match:      "security\\[at\\]example.org",
						LineNumber: 1,
						Offset:     14,
					},
				},
			},
		},
		{
			name:     "Email negative",
			input:    "No email address here.",
			expected: nil,
		},
		// Disclosure/vuln/number regex
		{
			name:  "Disclosure positive (word)",
			input: "Please see our vulnerability disclosure policy.",
			expected: []checker.SecurityPolicyInformation{
				{
					InformationType: checker.SecurityPolicyInformationTypeText,
					InformationValue: checker.SecurityPolicyValueType{
						Match:      "vuln",
						LineNumber: 1,
						Offset:     15,
					},
				},
				{
					InformationType: checker.SecurityPolicyInformationTypeText,
					InformationValue: checker.SecurityPolicyValueType{
						Match:      "disclos",
						LineNumber: 1,
						Offset:     29,
					},
				},
			},
		},
		{
			name:  "Disclosure positive (number)",
			input: "Report issues to 1234.",
			expected: []checker.SecurityPolicyInformation{
				{
					InformationType: checker.SecurityPolicyInformationTypeText,
					InformationValue: checker.SecurityPolicyValueType{
						Match:      "1234",
						LineNumber: 1,
						Offset:     17,
					},
				},
			},
		},
		{
			name:     "Disclosure negative",
			input:    "No relevant keywords or numbers here.",
			expected: nil,
		},
		// Multi-line input
		{
			name:  "Multi-line, all types",
			input: "Contact: sec@ex.com\nPolicy: https://foo.com\nID: 42 vuln Disclos",
			expected: []checker.SecurityPolicyInformation{
				{
					InformationType: checker.SecurityPolicyInformationTypeEmail,
					InformationValue: checker.SecurityPolicyValueType{
						Match:      "sec@ex.com",
						LineNumber: 1,
						Offset:     9,
					},
				},
				{
					InformationType: checker.SecurityPolicyInformationTypeLink,
					InformationValue: checker.SecurityPolicyValueType{
						Match:      "https://foo.com",
						LineNumber: 2,
						Offset:     8,
					},
				},
				{
					InformationType: checker.SecurityPolicyInformationTypeText,
					InformationValue: checker.SecurityPolicyValueType{
						Match:      "42",
						LineNumber: 3,
						Offset:     4,
					},
				},
				{
					InformationType: checker.SecurityPolicyInformationTypeText,
					InformationValue: checker.SecurityPolicyValueType{
						Match:      "vuln",
						LineNumber: 3,
						Offset:     7,
					},
				},
				{
					InformationType: checker.SecurityPolicyInformationTypeText,
					InformationValue: checker.SecurityPolicyValueType{
						Match:      "Disclos",
						LineNumber: 3,
						Offset:     12,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			hits := collectPolicyHits([]byte(tt.input))
			if len(hits) != len(tt.expected) {
				t.Errorf("expected %d hits, got %d: %+v", len(tt.expected), len(hits), hits)
				return
			}
			for i, want := range tt.expected {
				got := hits[i]
				if got.InformationType != want.InformationType ||
					got.InformationValue.Match != want.InformationValue.Match ||
					got.InformationValue.LineNumber != want.InformationValue.LineNumber ||
					got.InformationValue.Offset != want.InformationValue.Offset {
					t.Errorf("hit %d: got %+v, want %+v", i, got, want)
				}
			}
		})
	}
}
