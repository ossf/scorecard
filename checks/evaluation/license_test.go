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

func Test_scoreLicenseCriteria(t *testing.T) {
	t.Parallel()
	type args struct {
		f  *checker.LicenseFile
		dl checker.DetailLogger
	}
	tests := []struct { //nolint:govet
		name string
		args args
		want int
	}{
		{
			name: "License Attribution Type API",
			args: args{
				f: &checker.LicenseFile{
					LicenseInformation: checker.License{
						Attribution: checker.LicenseAttributionTypeAPI,
						Approved:    true,
					},
				},
				dl: &scut.TestDetailLogger{},
			},
			want: 10,
		},
		{
			name: "License Attribution Type Heuristics",
			args: args{
				f: &checker.LicenseFile{
					LicenseInformation: checker.License{
						Attribution: checker.LicenseAttributionTypeHeuristics,
					},
				},
				dl: &scut.TestDetailLogger{},
			},
			want: 9,
		},
		{
			name: "License Attribution Type Other",
			args: args{
				f: &checker.LicenseFile{
					LicenseInformation: checker.License{
						Attribution: checker.LicenseAttributionTypeOther,
					},
				},
				dl: &scut.TestDetailLogger{},
			},
			want: 6,
		},
		{
			name: "License Attribution Type Unknown",
			args: args{
				f: &checker.LicenseFile{
					LicenseInformation: checker.License{
						Attribution: "Unknown",
					},
				},
				dl: &scut.TestDetailLogger{},
			},
			want: 6,
		},
	}
	for _, tt := range tests {
		tt := tt // Parallel testing scoping hack.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := scoreLicenseCriteria(tt.args.f, tt.args.dl); got != tt.want {
				t.Errorf("scoreLicenseCriteria() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLicense(t *testing.T) {
	t.Parallel()
	type args struct { //nolint:govet
		name string
		dl   checker.DetailLogger
		r    *checker.LicenseData
	}
	tests := []struct {
		name string
		args args
		want checker.CheckResult
	}{
		{
			name: "No License",
			args: args{
				name: "No License",
				dl:   &scut.TestDetailLogger{},
			},
			want: checker.CheckResult{
				Score:   -1,
				Version: 2,
				Reason:  "internal error: empty raw data",
				Name:    "No License",
			},
		},
		{
			name: "No License Files",
			args: args{
				name: "No License Files",
				dl:   &scut.TestDetailLogger{},
				r: &checker.LicenseData{
					LicenseFiles: []checker.LicenseFile{},
				},
			},
			want: checker.CheckResult{
				Score:   0,
				Version: 2,
				Reason:  "license file not detected",
				Name:    "No License Files",
			},
		},
		{
			name: "License Files Detected",
			args: args{
				name: "License Files Detected",
				dl:   &scut.TestDetailLogger{},
				r: &checker.LicenseData{
					LicenseFiles: []checker.LicenseFile{
						{
							LicenseInformation: checker.License{
								Attribution: checker.LicenseAttributionTypeAPI,
								Approved:    true,
							},
						},
					},
				},
			},
			want: checker.CheckResult{
				Score:   10,
				Version: 2,
				Reason:  "license file detected",
				Name:    "License Files Detected",
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Parallel testing scoping hack.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := License(tt.args.name, tt.args.dl, tt.args.r); !cmp.Equal(got, tt.want, cmpopts.IgnoreFields(checker.CheckResult{}, "Error")) { //nolint:lll
				t.Errorf("License() = %v, want %v", got, cmp.Diff(got, tt.want, cmpopts.IgnoreFields(checker.CheckResult{}, "Error"))) //nolint:lll
			}
		})
	}
}
