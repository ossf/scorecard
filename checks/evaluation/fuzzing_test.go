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

	"github.com/ossf/scorecard/v5/checker"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/fuzzers"
	"github.com/ossf/scorecard/v5/probes/fuzzed"
	scut "github.com/ossf/scorecard/v5/utests"
)

func TestFuzzing(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		findings []finding.Finding
		result   scut.TestReturn
	}{
		{
			name: "no fuzzers",
			findings: []finding.Finding{
				{
					Probe:   fuzzed.Probe,
					Outcome: finding.OutcomeFalse,
				},
			},
			result: scut.TestReturn{
				Score:        checker.MinResultScore,
				NumberOfWarn: 1,
			},
		},
		{
			name: "single fuzzer gives max score",
			findings: []finding.Finding{
				fuzzTool(fuzzers.BuiltInGo),
			},
			result: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 1,
			},
		},
		{
			name: "one info per fuzzer",
			findings: []finding.Finding{
				fuzzTool(fuzzers.BuiltInGo),
				fuzzTool(fuzzers.OSSFuzz),
				fuzzTool(fuzzers.ClusterFuzzLite),
			},
			result: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 3,
			},
		},
		{
			name: "extra probe not part of check",
			findings: []finding.Finding{
				{
					Probe:   "someUnrelatedProbe",
					Outcome: finding.OutcomeFalse,
				},
				fuzzTool(fuzzers.RustCargoFuzz),
			},
			result: scut.TestReturn{
				Score: checker.InconclusiveResultScore,
				Error: sce.ErrScorecardInternal,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			got := Fuzzing(tt.name, tt.findings, &dl)
			scut.ValidateTestReturn(t, tt.name, &tt.result, &got, &dl)
		})
	}
}

func fuzzTool(name string) finding.Finding {
	return finding.Finding{
		Probe:   fuzzed.Probe,
		Outcome: finding.OutcomeTrue,
		Values: map[string]string{
			fuzzed.ToolKey: name,
		},
	}
}
