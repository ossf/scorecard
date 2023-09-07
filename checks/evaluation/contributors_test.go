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

func TestContributors(t *testing.T) {
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
			name: "Only has two positive outcomes",
			args: args{
				name: "Contributors",
				findings: []finding.Finding{
					{
						Probe:   "contributorsFromOrgOrCompany",
						Outcome: finding.OutcomePositive,
					},
					{
						Probe:   "contributorsFromOrgOrCompany",
						Outcome: finding.OutcomePositive,
					},
				},
			},
			want: checker.CheckResult{
				Score:   0,
				Version: 2,
				Name:    "contributorsFromOrgOrCompany",
				Reason:  "project has 2 contributing companies or organizations",
			},
		}, {
			name: "Has two positive outcomes",
			args: args{
				name: "Contributors",
				findings: []finding.Finding{
					{
						Probe:   "contributorsFromOrgOrCompany",
						Outcome: finding.OutcomePositive,
					},
					{
						Probe:   "contributorsFromOrgOrCompany",
						Outcome: finding.OutcomePositive,
					},
					{
						Probe:   "contributorsFromOrgOrCompany",
						Outcome: finding.OutcomePositive,
					},
				},
			},
			want: checker.CheckResult{
				Score:   10,
				Version: 2,
				Name:    "contributorsFromOrgOrCompany",
				Reason:  "project has 3 contributing companies or organizations",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := Contributors("contributorsFromOrgOrCompany", tt.args.findings)
			if !cmp.Equal(result, tt.want, cmpopts.IgnoreFields(checker.CheckResult{}, "Error")) { //nolint:govet
				t.Errorf("expected %v, got %v", tt.want, cmp.Diff(tt.want, result, cmpopts.IgnoreFields(checker.CheckResult{}, "Error"))) //nolint:lll
			}
		})
	}
}
