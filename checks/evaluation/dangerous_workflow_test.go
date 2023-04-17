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

func TestDangerousWorkflow(t *testing.T) {
	t.Parallel()
	type args struct { //nolint:govet
		name string
		dl   checker.DetailLogger
		r    *checker.DangerousWorkflowData
	}
	tests := []struct {
		name string
		args args
		want checker.CheckResult
	}{
		{
			name: "DangerousWorkflow - empty",
			args: args{
				name: "DangerousWorkflow",
				dl:   &scut.TestDetailLogger{},
				r:    &checker.DangerousWorkflowData{},
			},
			want: checker.CheckResult{
				Score:   10,
				Reason:  "no dangerous workflow patterns detected",
				Version: 2,
				Name:    "DangerousWorkflow",
			},
		},
		{
			name: "DangerousWorkflow - Dangerous workflow detected",
			args: args{
				name: "DangerousWorkflow",
				dl:   &scut.TestDetailLogger{},
				r: &checker.DangerousWorkflowData{
					Workflows: []checker.DangerousWorkflow{
						{
							Type: checker.DangerousWorkflowUntrustedCheckout,
							File: checker.File{
								Path:      "a",
								Snippet:   "a",
								Offset:    0,
								EndOffset: 0,
								Type:      0,
							},
						},
					},
				},
			},
			want: checker.CheckResult{
				Score:   0,
				Reason:  "dangerous workflow patterns detected",
				Version: 2,
				Name:    "DangerousWorkflow",
			},
		},
		{
			name: "DangerousWorkflow - Script injection detected",
			args: args{
				name: "DangerousWorkflow",
				dl:   &scut.TestDetailLogger{},
				r: &checker.DangerousWorkflowData{
					Workflows: []checker.DangerousWorkflow{
						{
							Type: checker.DangerousWorkflowScriptInjection,
							File: checker.File{
								Path:      "a",
								Snippet:   "a",
								Offset:    0,
								EndOffset: 0,
								Type:      0,
							},
						},
					},
				},
			},
			want: checker.CheckResult{
				Score:   0,
				Reason:  "dangerous workflow patterns detected",
				Version: 2,
				Name:    "DangerousWorkflow",
			},
		},
		{
			name: "DangerousWorkflow - unknown type",
			args: args{
				name: "DangerousWorkflow",
				dl:   &scut.TestDetailLogger{},
				r: &checker.DangerousWorkflowData{
					Workflows: []checker.DangerousWorkflow{
						{
							Type: "foobar",
							File: checker.File{
								Path:      "a",
								Snippet:   "a",
								Offset:    0,
								EndOffset: 0,
								Type:      0,
							},
						},
					},
				},
			},
			want: checker.CheckResult{
				Score:   -1,
				Reason:  "internal error: invalid type",
				Version: 2,
				Name:    "DangerousWorkflow",
			},
		},
		{
			name: "DangerousWorkflow - nil data",
			args: args{
				name: "DangerousWorkflow",
				dl:   &scut.TestDetailLogger{},
				r:    nil,
			},
			want: checker.CheckResult{
				Score:   -1,
				Reason:  "internal error: empty raw data",
				Name:    "DangerousWorkflow",
				Version: 2,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := DangerousWorkflow(tt.args.name, tt.args.dl, tt.args.r); !cmp.Equal(got, tt.want, cmpopts.IgnoreFields(checker.CheckResult{}, "Error")) { //nolint:lll
				t.Errorf("DangerousWorkflow() = %v, want %v", got, cmp.Diff(got, tt.want, cmpopts.IgnoreFields(checker.CheckResult{}, "Error"))) //nolint:lll
			}
		})
	}
}
