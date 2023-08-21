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
	scut "github.com/ossf/scorecard/v4/utests"
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
						Probe:   "fuzzedWithClusterFuzzLite",
						Outcome: finding.OutcomeNegative,
					},
					{
						Probe:   "fuzzedWithGoNative",
						Outcome: finding.OutcomeNegative,
					},
					{
						Probe:   "fuzzedWithOneFuzz",
						Outcome: finding.OutcomeNegative,
					},
					{
						Probe:   "fuzzedWithOSSFuzz",
						Outcome: finding.OutcomeNegative,
					},
					{
						Probe:   "fuzzedWithPropertyBasedHaskell",
						Outcome: finding.OutcomeNegative,
					},
					{
						Probe:   "fuzzedWithPropertyBasedJavascript",
						Outcome: finding.OutcomeNegative,
					},
					{
						Probe:   "fuzzedWithPropertyBasedTypescript",
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
			name: "Fuzzing - fuzzing GoNative",
			args: args{
				name: "Fuzzing",
				findings: []finding.Finding{
					{
						Probe:   "fuzzedWithClusterFuzzLite",
						Outcome: finding.OutcomeNegative,
					},
					{
						Probe:   "fuzzedWithGoNative",
						Outcome: finding.OutcomePositive,
					},
					{
						Probe:   "fuzzedWithOneFuzz",
						Outcome: finding.OutcomeNegative,
					},
					{
						Probe:   "fuzzedWithOSSFuzz",
						Outcome: finding.OutcomeNegative,
					},
					{
						Probe:   "fuzzedWithPropertyBasedHaskell",
						Outcome: finding.OutcomeNegative,
					},
					{
						Probe:   "fuzzedWithPropertyBasedJavascript",
						Outcome: finding.OutcomeNegative,
					},
					{
						Probe:   "fuzzedWithPropertyBasedTypescript",
						Outcome: finding.OutcomeNegative,
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
			name: "Fuzzing - fuzzing missing GoNative finding",
			args: args{
				name: "Fuzzing",
				findings: []finding.Finding{
					{
						Probe:   "fuzzedWithClusterFuzzLite",
						Outcome: finding.OutcomeNegative,
					},
					{
						Probe:   "fuzzedWithOneFuzz",
						Outcome: finding.OutcomeNegative,
					},
					{
						Probe:   "fuzzedWithOSSFuzz",
						Outcome: finding.OutcomeNegative,
					},
					{
						Probe:   "fuzzedWithPropertyBasedHaskell",
						Outcome: finding.OutcomeNegative,
					},
					{
						Probe:   "fuzzedWithPropertyBasedJavascript",
						Outcome: finding.OutcomeNegative,
					},
					{
						Probe:   "fuzzedWithPropertyBasedTypescript",
						Outcome: finding.OutcomeNegative,
					},
				},
			},
			want: checker.CheckResult{
				Score:   -1,
				Name:    "Fuzzing",
				Version: 2,
				Reason:  "internal error: invalid probe results",
			},
		},
		{
			name: "Fuzzing - fuzzing invalid probe name",
			args: args{
				name: "Fuzzing",
				findings: []finding.Finding{
					{
						Probe:   "fuzzedWithClusterFuzzLite",
						Outcome: finding.OutcomeNegative,
					},
					{
						Probe:   "fuzzedWithGoNative",
						Outcome: finding.OutcomePositive,
					},
					{
						Probe:   "fuzzedWithOneFuzz",
						Outcome: finding.OutcomeNegative,
					},
					{
						Probe:   "fuzzedWithOSSFuzz",
						Outcome: finding.OutcomeNegative,
					},
					{
						Probe:   "fuzzedWithPropertyBasedHaskell",
						Outcome: finding.OutcomeNegative,
					},
					{
						Probe:   "fuzzedWithPropertyBasedJavascript",
						Outcome: finding.OutcomeNegative,
					},
					{
						Probe:   "fuzzedWithPropertyBasedTypescript",
						Outcome: finding.OutcomeNegative,
					},
					{
						Probe:   "fuzzedWithInvalidProbeName",
						Outcome: finding.OutcomePositive,
					},
				},
			},
			want: checker.CheckResult{
				Score:   -1,
				Name:    "Fuzzing",
				Version: 2,
				Reason:  "internal error: invalid probe results",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			if got := Fuzzing(tt.args.name, tt.args.findings, &dl); !cmp.Equal(got, tt.want, cmpopts.IgnoreFields(checker.CheckResult{}, "Error")) { //nolint:lll
				t.Errorf("Fuzzing() = %v, want %v", got, cmp.Diff(got, tt.want, cmpopts.IgnoreFields(checker.CheckResult{}, "Error"))) //nolint:lll
			}
		})
	}
}
