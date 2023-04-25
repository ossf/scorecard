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
	scut "github.com/ossf/scorecard/v4/utests"
)

func TestFuzzing(t *testing.T) {
	t.Parallel()
	type args struct { //nolint
		name string
		dl   checker.DetailLogger
		r    *checker.FuzzingData
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
				dl:   &scut.TestDetailLogger{},
				r:    &checker.FuzzingData{},
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
				dl:   &scut.TestDetailLogger{},
				r: &checker.FuzzingData{
					Fuzzers: []checker.Tool{
						{
							Name: "Fuzzing",
							Files: []checker.File{
								{
									Path:    "Fuzzing",
									Type:    0,
									Offset:  1,
									Snippet: "Fuzzing",
								},
							},
						},
					},
				},
			},
			want: checker.CheckResult{
				Score:   10,
				Name:    "Fuzzing",
				Version: 2,
				Reason:  "project is fuzzed with [Fuzzing]",
			},
		},
		{
			name: "Fuzzing - fuzzing data nil",
			args: args{
				name: "Fuzzing",
				dl:   &scut.TestDetailLogger{},
				r:    nil,
			},
			want: checker.CheckResult{
				Score:   -1,
				Name:    "Fuzzing",
				Version: 2,
				Reason:  "internal error: empty raw data",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := Fuzzing(tt.args.name, tt.args.dl, tt.args.r); !cmp.Equal(got, tt.want, cmpopts.IgnoreFields(checker.CheckResult{}, "Error")) { //nolint:lll
				t.Errorf("Fuzzing() = %v, want %v", got, cmp.Diff(got, tt.want, cmpopts.IgnoreFields(checker.CheckResult{}, "Error"))) //nolint:lll
			}
		})
	}
}
