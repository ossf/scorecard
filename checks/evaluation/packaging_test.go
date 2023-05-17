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

func Test_createLogMessage(t *testing.T) {
	msg := "msg"
	t.Parallel()
	tests := []struct { //nolint:govet
		name    string
		args    checker.Package
		want    checker.LogMessage
		wantErr bool
	}{
		{
			name:    "nil package",
			args:    checker.Package{},
			want:    checker.LogMessage{},
			wantErr: true,
		},
		{
			name: "nil file",
			args: checker.Package{
				File: nil,
			},
			want:    checker.LogMessage{},
			wantErr: true,
		},
		{
			name: "msg is not nil",
			args: checker.Package{
				File: &checker.File{},
				Msg:  &msg,
			},
			want: checker.LogMessage{
				Text: "",
			},
			wantErr: true,
		},
		{
			name: "file is not nil",
			args: checker.Package{
				File: &checker.File{
					Path: "path",
				},
			},
			want: checker.LogMessage{
				Path: "path",
			},
			wantErr: true,
		},
		{
			name: "runs are not zero",
			args: checker.Package{
				File: &checker.File{
					Path: "path",
				},
				Runs: []checker.Run{
					{},
				},
			},
			want: checker.LogMessage{
				Text: "GitHub/GitLab publishing workflow used in run ",
				Path: "path",
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := createLogMessage(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("createLogMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("createLogMessage() got = %v, want %v", got, cmp.Diff(got, tt.want))
			}
		})
	}
}

func TestPackaging(t *testing.T) {
	t.Parallel()
	type args struct { //nolint:govet
		name string
		dl   checker.DetailLogger
		r    *checker.PackagingData
	}
	tests := []struct {
		name string
		args args
		want checker.CheckResult
	}{
		{
			name: "nil packaging data",
			args: args{
				name: "name",
				dl:   nil,
				r:    nil,
			},
			want: checker.CheckResult{
				Name:    "name",
				Version: 2,
				Score:   -1,
				Reason:  "internal error: empty raw data",
			},
		},
		{
			name: "empty packaging data",
			args: args{
				name: "name",
				dl:   &scut.TestDetailLogger{},
				r:    &checker.PackagingData{},
			},
			want: checker.CheckResult{
				Name:    "name",
				Version: 2,
				Score:   -1,
				Reason:  "no published package detected",
			},
		},
		{
			name: "runs are not zero",
			args: args{
				dl: &scut.TestDetailLogger{},
				r: &checker.PackagingData{
					Packages: []checker.Package{
						{
							File: &checker.File{
								Path: "path",
							},
							Runs: []checker.Run{
								{},
							},
						},
					},
				},
			},
			want: checker.CheckResult{
				Name:    "",
				Version: 2,
				Score:   10,
				Reason:  "publishing workflow detected",
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := Packaging(tt.args.name, tt.args.dl, tt.args.r); !cmp.Equal(got, tt.want, cmpopts.IgnoreFields(checker.CheckResult{}, "Error")) { //nolint:lll
				t.Errorf("Packaging() = %v, want %v", got, cmp.Diff(got, tt.want, cmpopts.IgnoreFields(checker.CheckResult{}, "Error"))) //nolint:lll
			}
		})
	}
}
