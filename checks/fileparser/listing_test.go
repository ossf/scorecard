// Copyright 2021 Security Scorecard Authors
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

package fileparser

import (
	"testing"
)

func TestIsTemplateFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		filename   string
		isTemplate bool
	}{
		{
			filename:   "Dockerfile.template",
			isTemplate: true,
		},
		{
			filename:   "Dockerfile.tmpl",
			isTemplate: true,
		},
		{
			filename:   "Dockerfile.template-debian",
			isTemplate: true,
		},
		{
			filename:   "Dockerfile.tmpl.",
			isTemplate: true,
		},
		{
			filename:   "Dockerfile-template",
			isTemplate: true,
		},
		{
			filename:   "tmpl.Dockerfile",
			isTemplate: true,
		},
		{
			filename:   "template.Dockerfile",
			isTemplate: true,
		},
		{
			filename:   "Dockerfile_template",
			isTemplate: true,
		},
		{
			filename:   "Dockerfile.tmpl.prod",
			isTemplate: true,
		},
		{
			filename:   "Dockerfile.Template",
			isTemplate: true,
		},
		{
			filename:   "dockerfile.tpl",
			isTemplate: true,
		},
		{
			filename:   "build/Dockerfile.tpl",
			isTemplate: true,
		},
		{
			filename:   "build/tpl.Dockerfile",
			isTemplate: true,
		},
		{
			filename:   "DockerfileTemplate",
			isTemplate: false,
		},
		{
			filename:   "Dockerfile.linux",
			isTemplate: false,
		},
		{
			filename:   "tmp.Dockerfile",
			isTemplate: false,
		},
		{
			filename:   "Dockerfile",
			isTemplate: false,
		},
		{
			filename:   "Dockerfile.temp.late",
			isTemplate: false,
		},
		{
			filename:   "Dockerfile.temp",
			isTemplate: false,
		},
		{
			filename:   "template/Dockerfile",
			isTemplate: false,
		},
		{
			filename:   "linux.Dockerfile",
			isTemplate: false,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.filename, func(t *testing.T) {
			t.Parallel()
			if got := IsTemplateFile(tt.filename); got != tt.isTemplate {
				t.Errorf("%v: Got (%v) expected (%v)", tt.filename, got, tt.isTemplate)
			}
		})
	}
}

// TestCheckFileContainsCommands tests if the content starts with a comment.
func TestCheckFileContainsCommands(t *testing.T) {
	t.Parallel()
	//nolint
	type args struct {
		content []byte
		comment string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Dockerfile.template",
			args: args{
				content: []byte(`FROM golang:1.12.4`),
			},
			want: false,
		},
		{
			name: "Dockerfile.template with a comment",
			args: args{
				content: []byte(`# This is a comment
				FROM golang:1.12.4`),
				// start with a comment
				comment: "#",
			},
			want: true,
		},
		{
			name: "empty file",
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := CheckFileContainsCommands(tt.args.content, tt.args.comment); got != tt.want {
				t.Errorf("CheckFileContainsCommands() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isMatchingPath(t *testing.T) {
	t.Parallel()
	type args struct {
		pattern       string
		fullpath      string
		caseSensitive bool
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "matching path",
			args: args{
				pattern:       "Dockerfile",
				fullpath:      "Dockerfile",
				caseSensitive: true,
			},
			want: true,
		},
		{
			name: "matching path with case insensitive",
			args: args{
				pattern:       "Dockerfile",
				fullpath:      "dockerfile",
				caseSensitive: false,
			},
			want: true,
		},
		{
			name: "matching path with case insensitive",
			args: args{
				pattern:       "Dockerfile",
				fullpath:      "dockerfile",
				caseSensitive: true,
			},
			want: false,
		},
		{
			name: "matching path with case insensitive",
			args: args{
				pattern:       "Dockerfile",
				fullpath:      "Dockerfile.template",
				caseSensitive: false,
			},
			want: false,
		},
		{
			name: "matching path with case insensitive",
			args: args{
				pattern:       "Dockerfile",
				fullpath:      "Dockerfile.template",
				caseSensitive: true,
			},
			want: false,
		},
		{
			name: "matching path with case insensitive",
			args: args{
				pattern:       "Dockerfile",
				fullpath:      "Dockerfile.template",
				caseSensitive: false,
			},
			want: false,
		},
		{
			name: "matching path with case insensitive",
			args: args{
				pattern:       "Dockerfile",
				fullpath:      "Dockerfile.template",
				caseSensitive: true,
			},
		},
		{
			name: "matching path with case insensitive",
			args: args{
				pattern:       "Dockerfile",
				fullpath:      "Dockerfile.template",
				caseSensitive: false,
			},
		},
		{
			name: "matching path with case insensitive",
			args: args{
				pattern:       "Dockerfile",
				fullpath:      "Dockerfile.template",
				caseSensitive: true,
			},
		},
		{
			name: "matching path with case insensitive",
			args: args{
				pattern:       "Dockerfile",
				fullpath:      "Dockerfile.template",
				caseSensitive: false,
			},
		},
		{
			name: "matching path with case insensitive",
			args: args{
				pattern:       "Dockerfile",
				fullpath:      "Dockerfile.template",
				caseSensitive: true,
			},
		},
		{
			name: "matching path with case insensitive",
			args: args{
				pattern:       "Dockerfile",
				fullpath:      "Dockerfile.template",
				caseSensitive: false,
			},
		},
		{
			name: "matching path with case insensitive",
			args: args{
				pattern:       "Dockerfile",
				fullpath:      "Dockerfile.template",
				caseSensitive: true,
			},
		},
		{
			name: "matching path with case insensitive",
			args: args{
				pattern:       "Dockerfile",
				fullpath:      "Dockerfile.template",
				caseSensitive: false,
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := isMatchingPath(tt.args.pattern, tt.args.fullpath, tt.args.caseSensitive)
			if (err != nil) != tt.wantErr {
				t.Errorf("isMatchingPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("isMatchingPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isTestdataFile(t *testing.T) {
	t.Parallel()
	type args struct {
		fullpath string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "testdata file",
			args: args{
				fullpath: "testdata/Dockerfile",
			},
			want: true,
		},
		{
			name: "testdata file",
			args: args{
				fullpath: "testdata/Dockerfile.template",
			},
			want: true,
		},
		{
			name: "testdata file",
			args: args{
				fullpath: "testdata/Dockerfile.template.template",
			},
			want: true,
		},
		{
			name: "testdata file",
			args: args{
				fullpath: "testdata/Dockerfile.template.template.template",
			},
			want: true,
		},
		{
			name: "testdata file",
			args: args{
				fullpath: "testdata/Dockerfile.template.template.template.template",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isTestdataFile(tt.args.fullpath); got != tt.want {
				t.Errorf("isTestdataFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
