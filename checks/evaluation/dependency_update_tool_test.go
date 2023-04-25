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

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
	scut "github.com/ossf/scorecard/v4/utests"
)

func TestDependencyUpdateTool(t *testing.T) {
	t.Parallel()
	//nolint
	type args struct {
		name string
		dl   checker.DetailLogger
		r    *checker.DependencyUpdateToolData
	}
	//nolint
	tests := []struct {
		name     string
		args     args
		want     checker.CheckResult
		err      bool
		expected scut.TestReturn
	}{
		{
			name: "DependencyUpdateTool",
			args: args{
				name: "DependencyUpdateTool",
				dl:   &scut.TestDetailLogger{},
				r: &checker.DependencyUpdateToolData{
					Tools: []checker.Tool{
						{
							Name: "DependencyUpdateTool",
						},
					},
				},
			},
			want: checker.CheckResult{
				Score: -1,
			},
			err: false,
			expected: scut.TestReturn{
				Error: sce.ErrScorecardInternal,
				Score: -1,
			},
		},
		{
			name: "empty tool list",
			args: args{
				name: "DependencyUpdateTool",
				dl:   &scut.TestDetailLogger{},
				r: &checker.DependencyUpdateToolData{
					Tools: []checker.Tool{},
				},
			},
			want: checker.CheckResult{
				Score: 0,
				Error: nil,
			},
			err: false,
			expected: scut.TestReturn{
				Score:        0,
				NumberOfWarn: 1,
			},
		},
		{
			name: "Valid tool",
			args: args{
				name: "DependencyUpdateTool",
				dl:   &scut.TestDetailLogger{},
				r: &checker.DependencyUpdateToolData{
					Tools: []checker.Tool{
						{
							Name: "DependencyUpdateTool",
							Files: []checker.File{
								{
									Path: "/etc/dependency-update-tool.conf",
									Snippet: `
										[dependency-update-tool]
										enabled = true
										`,
									Type: finding.FileTypeSource,
								},
							},
						},
					},
				},
			},
			want: checker.CheckResult{
				Score: 10,
				Error: nil,
			},
			expected: scut.TestReturn{
				Error:        nil,
				Score:        10,
				NumberOfInfo: 1,
			},
			err: false,
		},
		{
			name: "more than one tool in the list",
			args: args{
				name: "DependencyUpdateTool",
				dl:   &scut.TestDetailLogger{},
				r: &checker.DependencyUpdateToolData{
					Tools: []checker.Tool{
						{
							Name: "DependencyUpdateTool",
						},
						{
							Name: "DependencyUpdateTool",
						},
					},
				},
			},
			want: checker.CheckResult{
				Score: -1,
				Error: nil,
			},
			expected: scut.TestReturn{
				Error: sce.ErrScorecardInternal,
				Score: -1,
			},
			err: false,
		},
		{
			name: "Nil r",
			args: args{
				name: "nil r",
				dl:   &scut.TestDetailLogger{},
			},
			want: checker.CheckResult{
				Score: -1,
				Error: nil,
			},
			expected: scut.TestReturn{
				Error: sce.ErrScorecardInternal,
				Score: -1,
			},
			err: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dl := scut.TestDetailLogger{}
			got := DependencyUpdateTool(tt.args.name, &dl, tt.args.r)
			if tt.want.Score != got.Score {
				t.Errorf("DependencyUpdateTool() got Score = %v, want %v for %v", got.Score, tt.want.Score, tt.name)
			}
			if tt.err && got.Error == nil {
				t.Errorf("DependencyUpdateTool() error = %v, want %v for %v", got.Error, tt.want.Error, tt.name)
				return
			}

			if !scut.ValidateTestReturn(t, tt.name, &tt.expected, &got, &dl) {
				t.Fatalf(tt.name)
			}
		})
	}
}
