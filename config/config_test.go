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

package config

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_Parse_Checks(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		configPath string
		want       Config
		wantErr    bool
	}{
		{
			name:       "Annotation on a single check",
			configPath: "testdata/single_check.yml",
			want: Config{
				Annotations: []Annotation{
					{
						Checks:  []string{"binary-artifacts"},
						Reasons: []ReasonGroup{{Reason: "test-data"}},
					},
				},
			},
		},
		{
			name:       "Annotation on all checks",
			configPath: "testdata/all_checks.yml",
			want: Config{
				Annotations: []Annotation{
					{
						Checks: []string{
							"binary-artifacts",
							"branch-protection",
							"cii-best-practices",
							"ci-tests",
							"code-review",
							"contributors",
							"dangerous-workflow",
							"dependency-update-tool",
							"fuzzing",
							"license",
							"maintained",
							"packaging",
							"pinned-dependencies",
							"sast",
							"security-policy",
							"signed-releases",
							"token-permissions",
							"vulnerabilities",
						},
						Reasons: []ReasonGroup{{Reason: "test-data"}},
					},
				},
			},
		},
		{
			name:       "Annotating all reasons",
			configPath: "testdata/all_reasons.yml",
			want: Config{
				Annotations: []Annotation{
					{
						Checks: []string{"binary-artifacts"},
						Reasons: []ReasonGroup{
							{Reason: "test-data"},
							{Reason: "remediated"},
							{Reason: "not-applicable"},
							{Reason: "not-supported"},
							{Reason: "not-detected"},
						},
					},
				},
			},
		},
		{
			name:       "Multiple annotations",
			configPath: "testdata/multiple_annotations.yml",
			want: Config{
				Annotations: []Annotation{
					{
						Checks:  []string{"binary-artifacts"},
						Reasons: []ReasonGroup{{Reason: "test-data"}},
					},
					{
						Checks:  []string{"pinned-dependencies"},
						Reasons: []ReasonGroup{{Reason: "not-applicable"}},
					},
				},
			},
		},
		{
			name:       "Invalid check",
			configPath: "testdata/invalid_check.yml",
			wantErr:    true,
		},
		{
			name:       "Invalid reason",
			configPath: "testdata/invalid_reason.yml",
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r, err := os.Open(tt.configPath)
			if err != nil {
				t.Fatalf("Could not open config test file: %s", tt.configPath)
			}
			result, err := Parse(r)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Unexpected error during Parse: got %v, wantErr %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, result); diff != "" {
				t.Errorf("Config mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
