// Copyright 2023 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package evaluation

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
)

func TestFuzzing(t *testing.T) {
	t.Parallel()
	type args struct { //nolint
		name     string
		findings []finding.Finding
	}
	tests := []struct {
		name string
		args args
		want checker.CheckResult
	}{
		{
			name: "Fuzzing - no fuzzing",
			args: args{
				name: "Fuzzing",
				findings: []finding.Finding{
					{
						Probe:   "probeID",
						Outcome: finding.OutcomeNegative,
					},
				},
			},
			want: checker.CheckResult{
				Score:   0,
				Name:    "Fuzzing",
				Version: 2,
				Reason:  "project is not fuzzed",
			},
		},
		{
			name: "Fuzzing - fuzzing",
			args: args{
				name: "Fuzzing",
				findings: []finding.Finding{
					{
						Probe:   "probeID",
						Outcome: finding.OutcomePositive,
					},
				},
			},
			want: checker.CheckResult{
				Score:   10,
				Name:    "Fuzzing",
				Version: 2,
				Reason:  "project is fuzzed",
			},
		},
		{
			name: "Fuzzing - fuzzing no finding",
			args: args{
				name: "Fuzzing",
			},
			want: checker.CheckResult{
				Score:   -1,
				Name:    "Fuzzing",
				Version: 2,
				Reason:  "internal error: no findings",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := Fuzzing(tt.args.name, tt.args.findings); !cmp.Equal(got, tt.want, cmpopts.IgnoreFields(checker.CheckResult{}, "Error")) { //nolint:lll
				t.Errorf("Fuzzing() = %v, want %v", got, cmp.Diff(got, tt.want, cmpopts.IgnoreFields(checker.CheckResult{}, "Error"))) //nolint:lll
			}
		})
	}
}
