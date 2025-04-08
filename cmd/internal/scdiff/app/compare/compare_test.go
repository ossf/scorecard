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

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/pkg/scorecard"
)

func TestResults(t *testing.T) {
	t.Parallel()
	//nolint:govet // field alignment
	tests := []struct {
		name      string
		a, b      *scorecard.Result
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
			b:         &scorecard.Result{},
			wantEqual: false,
		},
		{
			name: "different repo name",
			a: &scorecard.Result{
				Repo: scorecard.RepoInfo{
					Name: "a",
				},
			},
			b: &scorecard.Result{
				Repo: scorecard.RepoInfo{
					Name: "b",
				},
			},
			wantEqual: false,
		},
		{
			name: "unequal amount of checks",
			a: &scorecard.Result{
				Checks: []checker.CheckResult{
					{
						Name: "a1",
					},
					{
						Name: "a2",
					},
				},
			},
			b: &scorecard.Result{
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
			a: &scorecard.Result{
				Checks: []checker.CheckResult{
					{
						Name: "a",
					},
				},
			},
			b: &scorecard.Result{
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
			a: &scorecard.Result{
				Checks: []checker.CheckResult{
					{
						Score: 1,
					},
				},
			},
			b: &scorecard.Result{
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
			a: &scorecard.Result{
				Checks: []checker.CheckResult{
					{
						Reason: "a",
					},
				},
			},
			b: &scorecard.Result{
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
			a: &scorecard.Result{
				Checks: []checker.CheckResult{
					{
						Details: []checker.CheckDetail{},
					},
				},
			},
			b: &scorecard.Result{
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
			name: "details have different levels",
			a: &scorecard.Result{
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
			b: &scorecard.Result{
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
			a: &scorecard.Result{
				Repo: scorecard.RepoInfo{
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
			b: &scorecard.Result{
				Repo: scorecard.RepoInfo{
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
			t.Parallel()
			if got := Results(tt.a, tt.b); got != tt.wantEqual {
				t.Errorf("Results() = %v, want %v", got, tt.wantEqual)
			}
		})
	}
}
