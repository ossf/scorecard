// Copyright 2022 Security Scorecard Authors
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

package raw

import "testing"

func Test_isSecurityRstFound(t *testing.T) {
	t.Parallel()
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test1",
			args: args{
				name: "test1",
			},
			want: false,
		},
		{
			name: "docs/security.rst",
			args: args{
				name: "docs/security.rst",
			},
			want: true,
		},
		{
			name: "doc/security.rst",
			args: args{
				name: "doc/security.rst",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isSecurityRstFound(tt.args.name); got != tt.want {
				t.Errorf("isSecurityRstFound() = %v, want %v for %v", got, tt.want, tt.name)
			}
		})
	}
}
