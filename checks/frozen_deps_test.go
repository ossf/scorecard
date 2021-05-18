// Copyright 2020 Security Scorecard Authors
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

package checks

import (
	"errors"
	"fmt"
	"io/ioutil"
	"testing"
)

//nolint:dupl // repeating test cases that are slightly different is acceptable
func TestGithubWorkflowPinning(t *testing.T) {
	t.Parallel()
	type args struct {
		Filename string
		Logf     func(s string, f ...interface{})
	}

	type returnValue struct {
		Result bool
		Error  error
	}

	l := log{}
	tests := []struct {
		name string
		args args
		want returnValue
	}{
		{
			name: "Zero size content",
			args: args{
				Filename: "",
				Logf:     l.Logf,
			},
			want: returnValue{false, errors.New("file has no content")},
		},
		{
			name: "Pinned workflow",
			args: args{
				Filename: "./testdata/workflow-pinned.yaml",
				Logf:     l.Logf,
			},
			want: returnValue{true, nil},
		},
		{
			name: "Non-pinned workflow",
			args: args{
				Filename: "./testdata/workflow-not-pinned.yaml",
				Logf:     l.Logf,
			},
			want: returnValue{false, nil},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		l.messages = []string{}
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error
			if len(tt.args.Filename) == 0 {
				content = make([]byte, 0)
			} else {
				content, err = ioutil.ReadFile(tt.args.Filename)
				if err != nil {
					panic(fmt.Errorf("cannot read file: %w", err))
				}
			}
			r, err := validateGitHubActionWorkflow(tt.args.Filename, content, tt.args.Logf)

			if (err != nil && tt.want.Error == nil) ||
				(err == nil && tt.want.Error != nil) ||
				(err != nil && tt.want.Error != nil && err.Error() != tt.want.Error.Error()) ||
				r != tt.want.Result {
				t.Errorf("TestGithubWorkflowPinning:\"%v\": %v (%v,%v) want (%v, %v)", tt.name, tt.args.Filename, r, err, tt.want.Result, tt.want.Error)
			}
		})
	}
}
