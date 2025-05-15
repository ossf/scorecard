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

package evaluation

import (
	"testing"

	"github.com/ossf/scorecard/v5/checker"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/dependencyUpdateToolConfigured"
	scut "github.com/ossf/scorecard/v5/utests"
)

func TestDependencyUpdateTool(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		findings []finding.Finding
		result   scut.TestReturn
	}{
		{
			name: "one update tool is max score",
			findings: []finding.Finding{
				depUpdateTool("Dependabot"),
			},
			result: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 1,
			},
		},
		{
			name: "multiple update tools both logged",
			findings: []finding.Finding{
				depUpdateTool("RenovateBot"),
				depUpdateTool("PyUp"),
			},
			result: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 2,
			},
		},
		{
			name: "no update tool is min score",
			findings: []finding.Finding{
				{
					Probe:   dependencyUpdateToolConfigured.Probe,
					Outcome: finding.OutcomeFalse,
				},
			},
			result: scut.TestReturn{
				Score:        checker.MinResultScore,
				NumberOfWarn: 1,
			},
		},
		{
			name: "invalid probe name is an error",
			findings: []finding.Finding{
				{
					Probe:   "notARealProbe",
					Outcome: finding.OutcomeFalse,
				},
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
			got := DependencyUpdateTool(tt.name, tt.findings, &dl)
			scut.ValidateTestReturn(t, tt.name, &tt.result, &got, &dl)
		})
	}
}

func depUpdateTool(name string) finding.Finding {
	return finding.Finding{
		Probe:   dependencyUpdateToolConfigured.Probe,
		Outcome: finding.OutcomeTrue,
		Values: map[string]string{
			dependencyUpdateToolConfigured.ToolKey: name,
		},
	}
}
