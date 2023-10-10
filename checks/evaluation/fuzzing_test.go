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

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
	scut "github.com/ossf/scorecard/v4/utests"
)

func TestFuzzing(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		findings []finding.Finding
		result   scut.TestReturn
	}{
		{
			name: "Fuzzing - no fuzzing",
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
					Probe:   "fuzzedWithPythonAtheris",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "fuzzedWithCLibFuzzer",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "fuzzedWithCppLibFuzzer",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "fuzzedWithRustCargofuzz",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "fuzzedWithSwiftLibFuzzer",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "fuzzedWithJavaJazzerFuzzer",
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
			result: scut.TestReturn{
				Score:        checker.MinResultScore,
				NumberOfWarn: 13,
			},
		},
		{
			name: "Fuzzing - fuzzing GoNative",
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
					Probe:   "fuzzedWithPythonAtheris",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "fuzzedWithCLibFuzzer",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "fuzzedWithCppLibFuzzer",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "fuzzedWithRustCargofuzz",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "fuzzedWithSwiftLibFuzzer",
					Outcome: finding.OutcomeNegative,
				},
				{
					Probe:   "fuzzedWithJavaJazzerFuzzer",
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
			result: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 1,
			},
		},

		{
			name: "Fuzzing - fuzzing missing GoNative finding",
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
			result: scut.TestReturn{
				Score: checker.InconclusiveResultScore,
				Error: sce.ErrScorecardInternal,
			},
		},
		{
			name: "Fuzzing - fuzzing invalid probe name",
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
			result: scut.TestReturn{
				Score: checker.InconclusiveResultScore,
				Error: sce.ErrScorecardInternal,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			got := Fuzzing(tt.name, tt.findings, &dl)
			if !scut.ValidateTestReturn(t, tt.name, &tt.result, &got, &dl) {
				t.Errorf("got %v, expected %v", got, tt.result)
			}
		})
	}
}
