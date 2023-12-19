// Copyright 2020 OpenSSF Scorecard Authors
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

// Package checks defines all Scorecard checks.
package checks

import (
	"testing"

	"github.com/ossf/scorecard/v4/checker"
)

func Test_registerCheck(t *testing.T) {
	t.Parallel()
	//nolint:govet
	type args struct {
		name string
		fn   checker.CheckFn
	}
	//nolint:govet
	tests := []struct {
		name    string
		args    args
		wanterr bool
	}{
		{
			name: "registerCheck",
			args: args{
				name: "test",
				fn:   func(x *checker.CheckRequest) checker.CheckResult { return checker.CheckResult{} },
			},
			wanterr: false,
		},
		{
			name: "empty func",
			args: args{
				name: "test",
			},
			wanterr: true,
		},
		{
			name: "empty name",
			args: args{
				name: "",
				fn:   func(x *checker.CheckRequest) checker.CheckResult { return checker.CheckResult{} },
			},
			wanterr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := registerCheck(tt.args.name, tt.args.fn, nil /*supportedRequestTypes*/); (err != nil) != tt.wanterr {
				t.Errorf("registerCheck() error = %v, wantErr %v", err, tt.wanterr)
			}
		})
	}
}
