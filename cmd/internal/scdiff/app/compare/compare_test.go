// Copyright 2023 OpenSSF Scorecard Authors
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

package compare

import (
	"testing"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/pkg"
)

func TestResults(t *testing.T) {
	//nolint:govet // field alignment
	tests := []struct {
		name      string
		a, b      *pkg.ScorecardResult
		wantEqual bool
	}{
		{
			name:      "both nil",
			a:         nil,
			b:         nil,
			wantEqual: true,
		},
		{
			name:      "one nil",
			a:         nil,
			b:         &pkg.ScorecardResult{},
			wantEqual: false,
		},
		{
			name: "different repo name",
			a: &pkg.ScorecardResult{
				Repo: pkg.RepoInfo{
					Name: "a",
				},
			},
			b: &pkg.ScorecardResult{
				Repo: pkg.RepoInfo{
					Name: "b",
				},
			},
			wantEqual: false,
		},
		{
			name: "unequal amount of checks",
			a: &pkg.ScorecardResult{
				Checks: []checker.CheckResult{
					{
						Name: "a1",
					},
					{
						Name: "a2",
					},
				},
			},
			b: &pkg.ScorecardResult{
				Checks: []checker.CheckResult{
					{
						Name: "b",
					},
				},
			},
			wantEqual: false,
		},
		{
			name: "different check name",
			a: &pkg.ScorecardResult{
				Checks: []checker.CheckResult{
					{
						Name: "a",
					},
				},
			},
			b: &pkg.ScorecardResult{
				Checks: []checker.CheckResult{
					{
						Name: "b",
					},
				},
			},
			wantEqual: false,
		},
		{
			name: "different check score",
			a: &pkg.ScorecardResult{
				Checks: []checker.CheckResult{
					{
						Score: 1,
					},
				},
			},
			b: &pkg.ScorecardResult{
				Checks: []checker.CheckResult{
					{
						Score: 2,
					},
				},
			},
			wantEqual: false,
		},
		{
			name: "different check reason",
			a: &pkg.ScorecardResult{
				Checks: []checker.CheckResult{
					{
						Reason: "a",
					},
				},
			},
			b: &pkg.ScorecardResult{
				Checks: []checker.CheckResult{
					{
						Reason: "b",
					},
				},
			},
			wantEqual: false,
		},
		{
			name: "unequal number of details",
			a: &pkg.ScorecardResult{
				Checks: []checker.CheckResult{
					{
						Details: []checker.CheckDetail{},
					},
				},
			},
			b: &pkg.ScorecardResult{
				Checks: []checker.CheckResult{
					{
						Details: []checker.CheckDetail{
							{
								Type: checker.DetailWarn,
							},
						},
					},
				},
			},
			wantEqual: false,
		},
		{
			name: "details have differnet levels",
			a: &pkg.ScorecardResult{
				Checks: []checker.CheckResult{
					{
						Details: []checker.CheckDetail{
							{
								Type: checker.DetailInfo,
							},
						},
					},
				},
			},
			b: &pkg.ScorecardResult{
				Checks: []checker.CheckResult{
					{
						Details: []checker.CheckDetail{
							{
								Type: checker.DetailWarn,
							},
						},
					},
				},
			},
			wantEqual: false,
		},
		{
			name: "equal results",
			a: &pkg.ScorecardResult{
				Repo: pkg.RepoInfo{
					Name: "foo",
				},
				Checks: []checker.CheckResult{
					{
						Name: "bar",
						Details: []checker.CheckDetail{
							{
								Type: checker.DetailWarn,
							},
						},
					},
				},
			},
			b: &pkg.ScorecardResult{
				Repo: pkg.RepoInfo{
					Name: "foo",
				},
				Checks: []checker.CheckResult{
					{
						Name: "bar",
						Details: []checker.CheckDetail{
							{
								Type: checker.DetailWarn,
							},
						},
					},
				},
			},
			wantEqual: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Results(tt.a, tt.b); got != tt.wantEqual {
				t.Errorf("Results() = %v, want %v", got, tt.wantEqual)
			}
		})
	}
}
