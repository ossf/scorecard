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

package memorysafe

import (
	"io"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	"github.com/ossf/scorecard/v5/finding"
)

func Test_Run(t *testing.T) {
	t.Parallel()
	//nolint:govet
	tests := []struct {
		name          string
		repoLanguages []clients.Language
		filenames     []string
		expected      []finding.Finding
		err           error
	}{
		// no languages
		{
			name:          "no languages",
			repoLanguages: []clients.Language{},
			filenames:     []string{},
			expected:      []finding.Finding{},
			err:           nil,
		},
		// unimplemented languages
		{
			name: "unimplemented languages",
			repoLanguages: []clients.Language{
				{Name: clients.Erlang, NumLines: 0},
			},
			filenames: []string{},
			expected:  []finding.Finding{},
			err:       nil,
		},
		// golang
		{
			name: "golang - no files",
			repoLanguages: []clients.Language{
				{Name: clients.Go, NumLines: 0},
			},
			filenames: []string{},
			expected: []finding.Finding{
				{
					Probe:   Probe,
					Message: "Golang code does not use the unsafe package",
					Outcome: finding.OutcomeTrue,
				},
			},
			err: nil,
		},
		{
			name: "golang - safe no imports",
			repoLanguages: []clients.Language{
				{Name: clients.Go, NumLines: 0},
			},
			filenames: []string{
				"testdata/safe-no-imports.go",
			},
			expected: []finding.Finding{
				{
					Probe:   Probe,
					Message: "Golang code does not use the unsafe package",
					Outcome: finding.OutcomeTrue,
				},
			},
			err: nil,
		},
		{
			name: "golang - safe with imports",
			repoLanguages: []clients.Language{
				{Name: clients.Go, NumLines: 0},
			},
			filenames: []string{
				"testdata/safe-with-imports.go",
			},
			expected: []finding.Finding{
				{
					Probe:   Probe,
					Message: "Golang code does not use the unsafe package",
					Outcome: finding.OutcomeTrue,
				},
			},
			err: nil,
		},
		{
			name: "golang - unsafe",
			repoLanguages: []clients.Language{
				{Name: clients.Go, NumLines: 0},
			},
			filenames: []string{
				"testdata/unsafe.go",
			},
			expected: []finding.Finding{
				{
					Probe:   Probe,
					Message: "Golang code uses the unsafe package",
					Outcome: finding.OutcomeFalse,
					Remediation: &finding.Remediation{
						Text:   "Visit the OpenSSF Memory Safety SIG guidance on how to make your project memory safe.\nGuidance for [Memory-Safe By Default Languages](https://github.com/ossf/Memory-Safety/blob/main/docs/best-practice-memory-safe-by-default-languages.md)\nGuidance for [Non Memory-Safe By Default Languages](https://github.com/ossf/Memory-Safety/blob/main/docs/best-practice-non-memory-safe-by-default-languages.md)",
						Effort: 2,
					},
					Location: &finding.Location{Path: "testdata/unsafe.go"},
				},
			},
			err: nil,
		},
		{
			name: "golang - unsafe with safe",
			repoLanguages: []clients.Language{
				{Name: clients.Go, NumLines: 0},
			},
			filenames: []string{
				"testdata/unsafe.go",
				"testdata/safe-no-imports.go",
			},
			expected: []finding.Finding{
				{
					Probe:   Probe,
					Message: "Golang code uses the unsafe package",
					Outcome: finding.OutcomeFalse,
					Remediation: &finding.Remediation{
						Text:   "Visit the OpenSSF Memory Safety SIG guidance on how to make your project memory safe.\nGuidance for [Memory-Safe By Default Languages](https://github.com/ossf/Memory-Safety/blob/main/docs/best-practice-memory-safe-by-default-languages.md)\nGuidance for [Non Memory-Safe By Default Languages](https://github.com/ossf/Memory-Safety/blob/main/docs/best-practice-non-memory-safe-by-default-languages.md)",
						Effort: 2,
					},
					Location: &finding.Location{Path: "testdata/unsafe.go"},
				},
			},
			err: nil,
		},
		{
			name: "golang - malformed file with unsafe",
			repoLanguages: []clients.Language{
				{Name: clients.Go, NumLines: 0},
			},
			filenames: []string{
				"testdata/malformed.go",
				"testdata/unsafe.go",
			},
			expected: []finding.Finding{
				{
					Probe:   Probe,
					Message: "Golang code uses the unsafe package",
					Outcome: finding.OutcomeFalse,
					Remediation: &finding.Remediation{
						Text:   "Visit the OpenSSF Memory Safety SIG guidance on how to make your project memory safe.\nGuidance for [Memory-Safe By Default Languages](https://github.com/ossf/Memory-Safety/blob/main/docs/best-practice-memory-safe-by-default-languages.md)\nGuidance for [Non Memory-Safe By Default Languages](https://github.com/ossf/Memory-Safety/blob/main/docs/best-practice-non-memory-safe-by-default-languages.md)",
						Effort: 2,
					},
					Location: &finding.Location{Path: "testdata/unsafe.go"},
				},
			},
			err: nil,
		},
		// csharp
		{
			name: "C# - no files",
			repoLanguages: []clients.Language{
				{Name: clients.CSharp, NumLines: 0},
			},
			filenames: []string{},
			expected: []finding.Finding{
				{
					Probe:   Probe,
					Message: "C# code does not allow unsafe blocks",
					Outcome: finding.OutcomeTrue,
				},
			},
			err: nil,
		},
		{
			name: "C# - safe explicit",
			repoLanguages: []clients.Language{
				{Name: clients.CSharp, NumLines: 0},
			},
			filenames: []string{
				"testdata/safe-explicit.csproj",
			},
			expected: []finding.Finding{
				{
					Probe:   Probe,
					Message: "C# code does not allow unsafe blocks",
					Outcome: finding.OutcomeTrue,
				},
			},
			err: nil,
		},
		{
			name: "C# - safe implicit",
			repoLanguages: []clients.Language{
				{Name: clients.CSharp, NumLines: 0},
			},
			filenames: []string{
				"testdata/safe-implicit.csproj",
			},
			expected: []finding.Finding{
				{
					Probe:   Probe,
					Message: "C# code does not allow unsafe blocks",
					Outcome: finding.OutcomeTrue,
				},
			},
			err: nil,
		},
		{
			name: "C# - unsafe",
			repoLanguages: []clients.Language{
				{Name: clients.CSharp, NumLines: 0},
			},
			filenames: []string{
				"testdata/unsafe.csproj",
			},
			expected: []finding.Finding{
				{
					Probe:   Probe,
					Message: "C# code allows the use of unsafe blocks",
					Outcome: finding.OutcomeFalse,
					Remediation: &finding.Remediation{
						Text:   "Visit the OpenSSF Memory Safety SIG guidance on how to make your project memory safe.\nGuidance for [Memory-Safe By Default Languages](https://github.com/ossf/Memory-Safety/blob/main/docs/best-practice-memory-safe-by-default-languages.md)\nGuidance for [Non Memory-Safe By Default Languages](https://github.com/ossf/Memory-Safety/blob/main/docs/best-practice-non-memory-safe-by-default-languages.md)",
						Effort: 2,
					},
					Location: &finding.Location{Path: "testdata/unsafe.csproj"},
				},
			},
			err: nil,
		},
		{
			name: "C# - unsafe with safe",
			repoLanguages: []clients.Language{
				{Name: clients.CSharp, NumLines: 0},
			},
			filenames: []string{
				"testdata/unsafe.csproj",
				"testdata/safe-explicit.csproj",
			},
			expected: []finding.Finding{
				{
					Probe:   Probe,
					Message: "C# code allows the use of unsafe blocks",
					Outcome: finding.OutcomeFalse,
					Remediation: &finding.Remediation{
						Text:   "Visit the OpenSSF Memory Safety SIG guidance on how to make your project memory safe.\nGuidance for [Memory-Safe By Default Languages](https://github.com/ossf/Memory-Safety/blob/main/docs/best-practice-memory-safe-by-default-languages.md)\nGuidance for [Non Memory-Safe By Default Languages](https://github.com/ossf/Memory-Safety/blob/main/docs/best-practice-non-memory-safe-by-default-languages.md)",
						Effort: 2,
					},
					Location: &finding.Location{Path: "testdata/unsafe.csproj"},
				},
			},
			err: nil,
		},
		{
			name: "C# - malformed file with unsafe",
			repoLanguages: []clients.Language{
				{Name: clients.CSharp, NumLines: 0},
			},
			filenames: []string{
				"testdata/malformed.csproj",
				"testdata/unsafe.csproj",
			},
			expected: []finding.Finding{
				{
					Probe:   Probe,
					Message: "C# code allows the use of unsafe blocks",
					Outcome: finding.OutcomeFalse,
					Remediation: &finding.Remediation{
						Text:   "Visit the OpenSSF Memory Safety SIG guidance on how to make your project memory safe.\nGuidance for [Memory-Safe By Default Languages](https://github.com/ossf/Memory-Safety/blob/main/docs/best-practice-memory-safe-by-default-languages.md)\nGuidance for [Non Memory-Safe By Default Languages](https://github.com/ossf/Memory-Safety/blob/main/docs/best-practice-non-memory-safe-by-default-languages.md)",
						Effort: 2,
					},
					Location: &finding.Location{Path: "testdata/unsafe.csproj"},
				},
			},
			err: nil,
		},

		// all languages
		{
			name: "All Languages - no files",
			repoLanguages: []clients.Language{
				{Name: clients.All, NumLines: 0},
			},
			filenames: []string{},
			expected: []finding.Finding{
				{
					Probe:   Probe,
					Message: "Golang code does not use the unsafe package",
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   Probe,
					Message: "C# code does not allow unsafe blocks",
					Outcome: finding.OutcomeTrue,
				},
			},
			err: nil,
		},
		{
			name: "All Languages - all safe",
			repoLanguages: []clients.Language{
				{Name: clients.All, NumLines: 0},
			},
			filenames: []string{
				"testdata/safe-no-imports.go",
				"testdata/safe-explicit.csproj",
			},
			expected: []finding.Finding{
				{
					Probe:   Probe,
					Message: "Golang code does not use the unsafe package",
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   Probe,
					Message: "C# code does not allow unsafe blocks",
					Outcome: finding.OutcomeTrue,
				},
			},
			err: nil,
		},
		{
			name: "All Languages - go safe csharp unsafe",
			repoLanguages: []clients.Language{
				{Name: clients.All, NumLines: 0},
			},
			filenames: []string{
				"testdata/safe-no-imports.go",
				"testdata/unsafe.csproj",
			},
			expected: []finding.Finding{
				{
					Probe:   Probe,
					Message: "Golang code does not use the unsafe package",
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   Probe,
					Message: "C# code allows the use of unsafe blocks",
					Outcome: finding.OutcomeFalse,
					Remediation: &finding.Remediation{
						Text:   "Visit the OpenSSF Memory Safety SIG guidance on how to make your project memory safe.\nGuidance for [Memory-Safe By Default Languages](https://github.com/ossf/Memory-Safety/blob/main/docs/best-practice-memory-safe-by-default-languages.md)\nGuidance for [Non Memory-Safe By Default Languages](https://github.com/ossf/Memory-Safety/blob/main/docs/best-practice-non-memory-safe-by-default-languages.md)",
						Effort: 2,
					},
					Location: &finding.Location{Path: "testdata/unsafe.csproj"},
				},
			},
			err: nil,
		},
		{
			name: "All Languages - go unsafe csharp safe",
			repoLanguages: []clients.Language{
				{Name: clients.All, NumLines: 0},
			},
			filenames: []string{
				"testdata/unsafe.go",
				"testdata/safe-explicit.csproj",
			},
			expected: []finding.Finding{
				{
					Probe:   Probe,
					Message: "Golang code uses the unsafe package",
					Outcome: finding.OutcomeFalse,
					Remediation: &finding.Remediation{
						Text:   "Visit the OpenSSF Memory Safety SIG guidance on how to make your project memory safe.\nGuidance for [Memory-Safe By Default Languages](https://github.com/ossf/Memory-Safety/blob/main/docs/best-practice-memory-safe-by-default-languages.md)\nGuidance for [Non Memory-Safe By Default Languages](https://github.com/ossf/Memory-Safety/blob/main/docs/best-practice-non-memory-safe-by-default-languages.md)",
						Effort: 2,
					},
					Location: &finding.Location{Path: "testdata/unsafe.go"},
				},
				{
					Probe:   Probe,
					Message: "C# code does not allow unsafe blocks",
					Outcome: finding.OutcomeTrue,
				},
			},
			err: nil,
		},
		{
			name: "All Languages - unsafe",
			repoLanguages: []clients.Language{
				{Name: clients.All, NumLines: 0},
			},
			filenames: []string{
				"testdata/unsafe.go",
				"testdata/unsafe.csproj",
			},
			expected: []finding.Finding{
				{
					Probe:   Probe,
					Message: "Golang code uses the unsafe package",
					Outcome: finding.OutcomeFalse,
					Remediation: &finding.Remediation{
						Text:   "Visit the OpenSSF Memory Safety SIG guidance on how to make your project memory safe.\nGuidance for [Memory-Safe By Default Languages](https://github.com/ossf/Memory-Safety/blob/main/docs/best-practice-memory-safe-by-default-languages.md)\nGuidance for [Non Memory-Safe By Default Languages](https://github.com/ossf/Memory-Safety/blob/main/docs/best-practice-non-memory-safe-by-default-languages.md)",
						Effort: 2,
					},
					Location: &finding.Location{Path: "testdata/unsafe.go"},
				},
				{
					Probe:   Probe,
					Message: "C# code allows the use of unsafe blocks",
					Outcome: finding.OutcomeFalse,
					Remediation: &finding.Remediation{
						Text:   "Visit the OpenSSF Memory Safety SIG guidance on how to make your project memory safe.\nGuidance for [Memory-Safe By Default Languages](https://github.com/ossf/Memory-Safety/blob/main/docs/best-practice-memory-safe-by-default-languages.md)\nGuidance for [Non Memory-Safe By Default Languages](https://github.com/ossf/Memory-Safety/blob/main/docs/best-practice-non-memory-safe-by-default-languages.md)",
						Effort: 2,
					},
					Location: &finding.Location{Path: "testdata/unsafe.csproj"},
				},
			},
			err: nil,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			raw := &checker.CheckRequest{}
			mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
			mockRepoClient.EXPECT().ListProgrammingLanguages().DoAndReturn(func() ([]clients.Language, error) {
				return tt.repoLanguages, nil
			}).AnyTimes()
			mockRepoClient.EXPECT().ListFiles(gomock.Any()).DoAndReturn(func(predicate func(string) (bool, error)) ([]string, error) {
				return tt.filenames, nil
			}).AnyTimes()
			mockRepoClient.EXPECT().GetFileReader(gomock.Any()).DoAndReturn(func(file string) (io.ReadCloser, error) {
				return os.Open(file)
			}).AnyTimes()
			raw.RepoClient = mockRepoClient
			raw.Dlogger = checker.NewLogger()
			findings, _, err := Run(raw)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			sortFindings := func(a, b finding.Finding) bool { return a.Message < b.Message }
			if diff := cmp.Diff(findings, tt.expected, cmpopts.IgnoreUnexported(finding.Finding{}), cmpopts.SortSlices(sortFindings)); diff != "" {
				t.Error(diff)
			}
		})
	}
}
