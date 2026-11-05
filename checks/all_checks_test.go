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
	"os"
	"testing"

	"github.com/ossf/scorecard/v5/checker"
)

func TestExperimentalChecks(t *testing.T) {
	for _, name := range []string{CheckBinaryArtifacts, "Not-A-Real-Check"} {
		if IsExperimental(name) {
			t.Errorf("IsExperimental(%q) = true, want false", name)
		}
	}
	for _, state := range []string{"unset", "empty", "enabled"} {
		t.Run(state, func(t *testing.T) {
			t.Setenv("SCORECARD_EXPERIMENTAL", "")
			if state == "unset" {
				if err := os.Unsetenv("SCORECARD_EXPERIMENTAL"); err != nil {
					t.Fatal(err)
				}
			} else if state == "enabled" {
				t.Setenv("SCORECARD_EXPERIMENTAL", "1")
			}
			defaults := GetAll()
			all := GetAllWithExperimental()
			for _, name := range []string{CheckWebHooks, CheckSBOM} {
				if !IsExperimental(name) {
					t.Errorf("IsExperimental(%q) = false, want true", name)
				}
				if _, ok := all[name]; !ok {
					t.Errorf("GetAllWithExperimental() is missing %q", name)
				}
				if _, ok := defaults[name]; ok != (state != "unset") {
					t.Errorf("GetAll() contains %q = %v with environment %s", name, ok, state)
				}
			}
		})
	}
}

func Test_registerCheck(t *testing.T) {
	t.Parallel()
	type args struct {
		fn   checker.CheckFn
		name string
	}
	tests := []struct {
		args    args
		name    string
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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := registerCheck(tt.args.name, tt.args.fn, nil /*supportedRequestTypes*/); (err != nil) != tt.wanterr {
				t.Errorf("registerCheck() error = %v, wantErr %v", err, tt.wanterr)
			}
		})
	}
}
