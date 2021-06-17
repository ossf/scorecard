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

func TestGithubWorkflowPinning(t *testing.T) {
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
			name: "Zero size content",
			args: args{
				Filename: "",
				Logf:     l.Logf,
			},
			want: returnValue{
				Error:  ErrEmptyFile,
				Result: false,
			},
		},
		{
			name: "Pinned workflow",
			args: args{
				Filename: "./testdata/workflow-pinned.yaml",
				Logf:     l.Logf,
			},
			want: returnValue{
				Error:  nil,
				Result: true,
			},
		},
		{
			name: "Non-pinned workflow",
			args: args{
				Filename: "./testdata/workflow-not-pinned.yaml",
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
			if tt.args.Filename == "" {
				content = make([]byte, 0)
			} else {
				content, err = ioutil.ReadFile(tt.args.Filename)
				if err != nil {
					panic(fmt.Errorf("cannot read file: %w", err))
				}
			}
			r, err := validateGitHubActionWorkflow(tt.args.Filename, content, tt.args.Logf)

			if !errors.Is(err, tt.want.Error) ||
				r != tt.want.Result {
				t.Errorf("TestGithubWorkflowPinning:\"%v\": %v (%v,%v) want (%v, %v)",
					tt.name, tt.args.Filename, r, err, tt.want.Result, tt.want.Error)
			}
		})
	}
}

func TestDockerfilePinning(t *testing.T) {
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
			name: "Invalid dockerfile",
			args: args{
				Filename: "./testdata/Dockerfile-invalid",
				Logf:     l.Logf,
			},
			want: returnValue{
				Error:  ErrInvalidDockerfile,
				Result: false,
			},
		},
		{
			name: "Pinned dockerfile",
			args: args{
				Filename: "./testdata/Dockerfile-pinned",
				Logf:     l.Logf,
			},
			want: returnValue{
				Error:  nil,
				Result: true,
			},
		},
		{
			name: "Pinned dockerfile as",
			args: args{
				Filename: "./testdata/Dockerfile-pinned-as",
				Logf:     l.Logf,
			},
			want: returnValue{
				Error:  nil,
				Result: true,
			},
		},
		{
			name: "Non-pinned dockerfile as",
			args: args{
				Filename: "./testdata/Dockerfile-not-pinned-as",
				Logf:     l.Logf,
			},
			want: returnValue{
				Error:  nil,
				Result: false,
			},
		},
		{
			name: "Non-pinned dockerfile",
			args: args{
				Filename: "./testdata/Dockerfile-not-pinned",
				Logf:     l.Logf,
			},
			want: returnValue{
				Error:  nil,
				Result: false,
			},
		},
	}
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

			r, err := validateDockerfile(tt.args.Filename, content, tt.args.Logf)

			if !errors.Is(err, tt.want.Error) ||
				r != tt.want.Result {
				t.Errorf("TestGithubWorkflowPinning:\"%v\": %v (%v,%v) want (%v, %v)",
					tt.name, tt.args.Filename, r, err, tt.want.Result, tt.want.Error)
			}
		})
	}
}

func TestDockerfileScriptDownload(t *testing.T) {
	t.Parallel()
	//nolint
	type args struct {
		// Note: this seems to be defined in e2e/e2e_suite_test.go
		Log      log
		Filename string
	}

	//nolint
	type returnValue struct {
		Error          error
		Result         bool
		NumberOfErrors int
	}

	tests := []struct {
		args args
		want returnValue
		name string
	}{
		{
			name: "curl | sh",
			args: args{
				Filename: "testdata/Dockerfile-curl-sh",
				Log:      log{},
			},
			want: returnValue{
				Error:          nil,
				Result:         false,
				NumberOfErrors: 3,
			},
		},
		{
			name: "wget | /bin/sh",
			args: args{
				Filename: "testdata/Dockerfile-wget-bin-sh",
				Log:      log{},
			},
			want: returnValue{
				Error:          nil,
				Result:         false,
				NumberOfErrors: 1,
			},
		},
		{
			name: "wget no exec",
			args: args{
				Filename: "testdata/Dockerfile-script-ok",
				Log:      log{},
			},
			want: returnValue{
				Error:          nil,
				Result:         true,
				NumberOfErrors: 0,
			},
		},
		{
			name: "curl file sh",
			args: args{
				Filename: "testdata/Dockerfile-curl-file-sh",
				Log:      log{},
			},
			want: returnValue{
				Error:          nil,
				Result:         false,
				NumberOfErrors: 8,
			},
		},
		{
			name: "proc substitution",
			args: args{
				Filename: "testdata/Dockerfile-proc-subs",
				Log:      log{},
			},
			want: returnValue{
				Error:          nil,
				Result:         false,
				NumberOfErrors: 4,
			},
		},
		{
			name: "wget file",
			args: args{
				Filename: "testdata/Dockerfile-wget-file",
				Log:      log{},
			},
			want: returnValue{
				Error:          nil,
				Result:         false,
				NumberOfErrors: 8,
			},
		},
		{
			name: "gsutil file",
			args: args{
				Filename: "testdata/Dockerfile-gsutil-file",
				Log:      log{},
			},
			want: returnValue{
				Error:          nil,
				Result:         false,
				NumberOfErrors: 14,
			},
		},
		{
			name: "aws file",
			args: args{
				Filename: "testdata/Dockerfile-aws-file",
				Log:      log{},
			},
			want: returnValue{
				Error:          nil,
				Result:         false,
				NumberOfErrors: 14,
			},
		},
	}
	//nolint
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error
			content, err = ioutil.ReadFile(tt.args.Filename)
			if err != nil {
				panic(fmt.Errorf("cannot read file: %w", err))
			}

			r, err := validateDockerfileDownloads(tt.args.Filename, content, tt.args.Log.Logf)

			if !errors.Is(err, tt.want.Error) ||
				r != tt.want.Result ||
				len(tt.args.Log.messages) != tt.want.NumberOfErrors {
				t.Errorf("TestDockerfileScriptDownload:\"%v\": %v (%v,%v,%v) want (%v, %v, %v)",
					tt.name, tt.args.Filename, r, err, len(tt.args.Log.messages), tt.want.Result, tt.want.Error, tt.want.NumberOfErrors)
			}
		})
	}
}
