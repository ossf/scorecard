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
	"github.com/ossf/scorecard/v4/clients"
	scut "github.com/ossf/scorecard/v4/utests"
)

// TestVulnerabilities tests the vulnerabilities checker.
func TestVulnerabilities(t *testing.T) {
	t.Parallel()
	//nolint
	type args struct {
		name string
		r    *checker.VulnerabilitiesData
	}
	tests := []struct {
		name     string
		args     args
		want     checker.CheckResult
		expected []struct {
			lineNumber uint
		}
	}{
		{
			name: "no vulnerabilities",
			args: args{
				name: "vulnerabilities_test.go",
				r: &checker.VulnerabilitiesData{
					Vulnerabilities: []clients.Vulnerability{},
				},
			},
			want: checker.CheckResult{
				Score: 10,
			},
		},
		{
			name: "one vulnerability",
			args: args{
				name: "vulnerabilities_test.go",
				r: &checker.VulnerabilitiesData{
					Vulnerabilities: []clients.Vulnerability{
						{
							ID: "CVE-2019-1234",
						},
					},
				},
			},
			want: checker.CheckResult{
				Score: 9,
			},
		},
		{
			name: "one vulnerability",
			args: args{
				name: "vulnerabilities_test.go",
			},
			want: checker.CheckResult{
				Score: -1,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			res := Vulnerabilities(tt.args.name, &dl, tt.args.r)
			if res.Score != tt.want.Score {
				t.Errorf("Vulnerabilities() = %v, want %v", res.Score, tt.want.Score)
			}
		})
	}
}
