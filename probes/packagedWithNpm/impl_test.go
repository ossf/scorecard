// Copyright 2024 OpenSSF Scorecard Authors
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
package packagedWithNpm

import (
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	"github.com/ossf/scorecard/v5/finding"
)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func Test_Run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		files          []string
		packageContent string
		outcomes       []finding.Outcome
		err            string
		expectError    bool
		expectFileRead bool
	}{
		{
			name:        "nil check request",
			expectError: true,
			err:         "nil check request",
		},
		{
			name:     "no package.json found",
			files:    []string{"src/main.js", "README.md"},
			outcomes: []finding.Outcome{finding.OutcomeFalse},
		},
		{
			name:     "package.json in subdirectory (should not match)",
			files:    []string{"frontend/package.json", "README.md"},
			outcomes: []finding.Outcome{finding.OutcomeFalse},
		},
		{
			name:           "package.json with invalid JSON",
			files:          []string{"package.json"},
			packageContent: `{"name": "test-package",`,
			outcomes:       []finding.Outcome{finding.OutcomeFalse},
			expectFileRead: true,
		},
		{
			name:           "package.json without name",
			files:          []string{"package.json"},
			packageContent: `{"version": "1.0.0"}`,
			outcomes:       []finding.Outcome{finding.OutcomeFalse},
			expectFileRead: true,
		},
		{
			name:           "package.json with empty name",
			files:          []string{"package.json"},
			packageContent: `{"name": "", "version": "1.0.0"}`,
			outcomes:       []finding.Outcome{finding.OutcomeFalse},
			expectFileRead: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var request *checker.CheckRequest
			if !tt.expectError {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()
				mockClient := mockrepo.NewMockRepoClient(ctrl)

				// Set up expectations for ListFiles call
				mockClient.EXPECT().ListFiles(gomock.Any()).DoAndReturn(func(predicate func(string) (bool, error)) ([]string, error) {
					var matchedFiles []string
					for _, file := range tt.files {
						match, err := predicate(file)
						if err != nil {
							return nil, err
						}
						if match {
							matchedFiles = append(matchedFiles, file)
						}
					}
					return matchedFiles, nil
				})

				// Set up expectations for file reading if needed
				if tt.expectFileRead {
					reader := nopCloser{strings.NewReader(tt.packageContent)}
					mockClient.EXPECT().GetFileReader("package.json").Return(reader, nil)
				}

				request = &checker.CheckRequest{
					RepoClient: mockClient,
				}
			}

			findings, s, err := Run(request)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.err)
				} else if err.Error() != tt.err {
					t.Errorf("expected error %q, got %q", tt.err, err.Error())
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if s != Probe {
				t.Errorf("expected probe name %q, got %q", Probe, s)
			}
			if len(findings) != len(tt.outcomes) {
				t.Errorf("expected %d findings, got %d", len(tt.outcomes), len(findings))
				return
			}
			for i, f := range findings {
				if i >= len(tt.outcomes) {
					break
				}
				if diff := cmp.Diff(tt.outcomes[i], f.Outcome, cmpopts.EquateEmpty()); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
