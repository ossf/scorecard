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

// Warning: config cannot import checks. This is why we declare a different package here
// and import both config and checks to test config.
package config_test

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ossf/scorecard/v5/config"
)

func Test_Parse_Checks(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		configPath string
		want       config.Config
		wantErr    bool
	}{
		{
			name:       "Annotation on a single check",
			configPath: "testdata/single_check.yml",
			want: config.Config{
				Annotations: []config.Annotation{
					{
						Checks:  []string{"Binary-Artifacts"},
						Reasons: []config.ReasonGroup{{Reason: "test-data"}},
					},
				},
			},
		},
		{
			name:       "Annotation on all checks (case insensitive)",
			configPath: "testdata/all_checks.yml",
			want: config.Config{
				Annotations: []config.Annotation{
					{
						Checks: []string{
							"Binary-Artifacts",
							"Branch-Protection",
							"CII-Best-Practices",
							"CI-Tests",
							"Code-Review",
							"Contributors",
							"Dangerous-Workflow",
							"Dependency-Update-Tool",
							"Fuzzing",
							"License",
							"Maintained",
							"Packaging",
							"Pinned-Dependencies",
							"SAST",
							"Security-Policy",
							"Signed-Releases",
							"Token-Permissions",
							"Vulnerabilities",
						},
						Reasons: []config.ReasonGroup{{Reason: "test-data"}},
					},
				},
			},
		},
		{
			name:       "Annotating all reasons",
			configPath: "testdata/all_reasons.yml",
			want: config.Config{
				Annotations: []config.Annotation{
					{
						Checks: []string{"Binary-Artifacts"},
						Reasons: []config.ReasonGroup{
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
			want: config.Config{
				Annotations: []config.Annotation{
					{
						Checks:  []string{"Binary-Artifacts"},
						Reasons: []config.ReasonGroup{{Reason: "test-data"}},
					},
					{
						Checks:  []string{"Pinned-Dependencies"},
						Reasons: []config.ReasonGroup{{Reason: "not-applicable"}},
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
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r, err := os.Open(tt.configPath)
			if err != nil {
				t.Fatalf("Could not open config test file: %s", tt.configPath)
			}
			result, err := config.Parse(r)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Unexpected error during Parse: got %v, wantErr %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, result); diff != "" {
				t.Errorf("Config mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
