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

func TestDockerfileScriptDownload(t *testing.T) {
	t.Parallel()
	type args struct {
		Logf     func(s string, f ...interface{})
		Filename string
	}

	type returnValue struct {
		Error  error
		Result bool
	}

	l := log{}
	tests := []struct {
		args args
		want returnValue
		name string
	}{
		{
			name: "sudo curl | bash",
			args: args{
				Filename: "testdata/Dockerfile-sudo-curl-bash",
				Logf:     l.Logf,
			},
			want: returnValue{
				Error:  nil,
				Result: false,
			},
		},
		{
			name: "curl | sudo bash",
			args: args{
				Filename: "testdata/Dockerfile-curl-sudo-bash",
				Logf:     l.Logf,
			},
			want: returnValue{
				Error:  nil,
				Result: false,
			},
		},
		{
			name: "curl | sh",
			args: args{
				Filename: "testdata/Dockerfile-curl-sh",
				Logf:     l.Logf,
			},
			want: returnValue{
				Error:  nil,
				Result: false,
			},
		},
		{
			name: "wget | /bin/sh",
			args: args{
				Filename: "testdata/Dockerfile-wget-bin-sh",
				Logf:     l.Logf,
			},
			want: returnValue{
				Error:  nil,
				Result: false,
			},
		},
		{
			name: "wget no exec",
			args: args{
				Filename: "testdata/Dockerfile-script-ok",
				Logf:     l.Logf,
			},
			want: returnValue{
				Error:  nil,
				Result: true,
			},
		},
		{
			name: "curl file sh",
			args: args{
				Filename: "testdata/Dockerfile-curl-file-sh",
				Logf:     l.Logf,
			},
			want: returnValue{
				Error:  nil,
				Result: false,
			},
		},
		{
			name: "proc substitution",
			args: args{
				Filename: "testdata/Dockerfile-proc-subs",
				Logf:     l.Logf,
			},
			want: returnValue{
				Error:  nil,
				Result: false,
			},
		},
		{
			name: "su proc substitution",
			args: args{
				Filename: "testdata/Dockerfile-su-proc-subs",
				Logf:     l.Logf,
			},
			want: returnValue{
				Error:  nil,
				Result: false,
			},
		},
		{
			name: "wget file",
			args: args{
				Filename: "testdata/Dockerfile-wget-file",
				Logf:     l.Logf,
			},
			want: returnValue{
				Error:  nil,
				Result: false,
			},
		},
		{
			name: "gsutil file",
			args: args{
				Filename: "testdata/Dockerfile-gsutil-file",
				Logf:     l.Logf,
			},
			want: returnValue{
				Error:  nil,
				Result: false,
			},
		},
		{
			name: "aws file",
			args: args{
				Filename: "testdata/Dockerfile-aws-file",
				Logf:     l.Logf,
			},
			want: returnValue{
				Error:  nil,
				Result: false,
			},
		},
	}
	//nolint
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		l.messages = []string{}
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error
			content, err = ioutil.ReadFile(tt.args.Filename)
			if err != nil {
				panic(fmt.Errorf("cannot read file: %w", err))
			}

			r, err := validateDockerfileDownloads(tt.args.Filename, content, tt.args.Logf)

			if !errors.Is(err, tt.want.Error) ||
				r != tt.want.Result {
				t.Errorf("TestDockerfileScriptDownload:\"%v\": %v (%v,%v) want (%v, %v)",
					tt.name, tt.args.Filename, r, err, tt.want.Result, tt.want.Error)
			}
		})
	}
}
