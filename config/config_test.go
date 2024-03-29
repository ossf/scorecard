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
	"errors"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ossf/scorecard/v4/checks"
	"github.com/ossf/scorecard/v4/config"
)

func Test_Parse_Checks(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		configPath string
		want       config.Config
		wantErr    error
	}{
		{
			name:       "Annotation on Binary-Artifacts check",
			configPath: "testdata/binary_artifacts.yml",
			want: config.Config{
				Annotations: []config.Annotation{
					{
						Checks:  []string{"binary-artifacts"},
						Reasons: []config.ReasonGroup{{Reason: "test-data"}},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		allChecks := []string{}
		for check := range checks.GetAll() {
			allChecks = append(allChecks, check)
		}
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r, err := os.Open(tt.configPath)
			if err != nil {
				t.Fatalf("Could not open config test file: %s", tt.configPath)
			}
			result, err := config.Parse(r, allChecks)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, result); diff != "" {
				t.Errorf("Config mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
