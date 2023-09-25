// Copyright 2021 OpenSSF Scorecard Authors
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

import (
	"os"
	"testing"

	"github.com/ossf/scorecard/v4/checker"
)

func TestIsSupportedShellScriptFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{
			name:     "awk script with .sh extension",
			filename: "testdata/shell_file_awk_shebang.sh",
			expected: false,
		},
		{
			name:     "bash script with #!/bin/bash",
			filename: "testdata/shell_file_bash_shebang1.sh",
			expected: true,
		},
		{
			name:     "bash script with #!/usr/bin/bash",
			filename: "testdata/shell_file_bash_shebang2.sh",
			expected: true,
		},
		{
			name:     "bash script with #!/bin/env bash",
			filename: "testdata/shell_file_bash_shebang3.sh",
			expected: true,
		},
		{
			name:     "mksh script",
			filename: "testdata/shell_file_mksh_shebang.sh",
			expected: true,
		},
		{
			name:     "script with .sh extension but no shebang",
			filename: "testdata/shell_file_no_shebang.sh",
			expected: true,
		},
		{
			name:     "shell script with #!/bin/sh",
			filename: "testdata/shell_file_sh_shebang.sh",
			expected: true,
		},
		{
			name:     "zsn script with #!/bin/zsh",
			filename: "testdata/shell_file_zsh_shebang.sh",
			expected: false,
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.filename, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error
			content, err = os.ReadFile(tt.filename)
			if err != nil {
				t.Errorf("cannot read file: %v", err)
			}
			result := isSupportedShellScriptFile(tt.filename, content)
			if result != tt.expected {
				t.Errorf("%v: Got (%v) expected (%v)", tt.name, result, tt.expected)
			}
		})
	}
}

func TestValidateShellFile(t *testing.T) {
	t.Parallel()
	filename := "testdata/script-invalid.sh"
	var content []byte
	var err error

	content, err = os.ReadFile(filename)
	if err != nil {
		t.Errorf("cannot read file: %v", err)
	}

	var r checker.PinningDependenciesData
	err = validateShellFile(filename, 0, 0, content, map[string]bool{}, &r)
	if err == nil {
		t.Errorf("failed to detect shell parsing error: %v", err)
	}
}

func Test_isDotNetUnpinnedDownload(t *testing.T) {
	type args struct {
		cmd []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "nuget install",
			args: args{
				cmd: []string{"nuget", "install", "Newtonsoft.Json"},
			},
			want: true,
		},
		{
			name: "nuget restore",
			args: args{
				cmd: []string{"nuget", "restore"},
			},
			want: false,
		},
		{
			name: "nuget install with -Version",
			args: args{
				cmd: []string{"nuget", "install", "Newtonsoft.Json", "-Version", "2"},
			},
			want: false,
		},
		{
			name: "nuget install with packages.config",
			args: args{
				cmd: []string{"nuget", "install", "config\\packages.config"},
			},
			want: false,
		},
		{
			name: "dotnet add",
			args: args{
				cmd: []string{"dotnet", "add", "package", "Newtonsoft.Json"},
			},
			want: true,
		},
		{
			name: "dotnet add to project",
			args: args{
				cmd: []string{"dotnet", "add", "project1", "package", "Newtonsoft.Json"},
			},
			want: true,
		},
		{
			name: "dotnet add reference to project",
			args: args{
				cmd: []string{"dotnet", "add", "project1", "reference", "OtherProject"},
			},
			want: false,
		},
		{
			name: "dotnet build",
			args: args{
				cmd: []string{"dotnet", "build"},
			},
			want: false,
		},
		{
			name: "dotnet add with -v",
			args: args{
				cmd: []string{"dotnet", "add", "package", "Newtonsoft.Json", "-v", "2.0"},
			},
			want: false,
		},
		{
			name: "dotnet add to project with -v",
			args: args{
				cmd: []string{"dotnet", "add", "project1", "package", "Newtonsoft.Json", "-v", "2.0"},
			},
			want: false,
		},
		{
			name: "dotnet add reference to project with -v",
			args: args{
				cmd: []string{"dotnet", "add", "project1", "reference", "Newtonsoft.Json", "-v", "2.0"},
			},
			want: false,
		},
		{
			name: "dotnet add with --version",
			args: args{
				cmd: []string{"dotnet", "add", "package", "Newtonsoft.Json", "--version", "2.0"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNugetUnpinnedDownload(tt.args.cmd); got != tt.want {
				t.Errorf("isNugetUnpinnedDownload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isGoUnpinnedDownload(t *testing.T) {
	type args struct {
		cmd []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "go get",
			args: args{
				cmd: []string{"go", "get", "github.com/ossf/scorecard"},
			},
			want: true,
		},
		{
			name: "go get with -d -v",
			args: args{
				cmd: []string{"go", "get", "-d", "-v"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isGoUnpinnedDownload(tt.args.cmd); got != tt.want {
				t.Errorf("isGoUnpinnedDownload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isNpmDownload(t *testing.T) {
	type args struct {
		cmd []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "npm install",
			args: args{
				cmd: []string{"npm", "install"},
			},
			want: true,
		},
		{
			name: "npm ci",
			args: args{
				cmd: []string{"npm", "ci"},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNpmDownload(tt.args.cmd); got != tt.want {
				t.Errorf("isNpmDownload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isNpmUnpinnedDownload(t *testing.T) {
	type args struct {
		cmd []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "npm install",
			args: args{
				cmd: []string{"npm", "install"},
			},
			want: true,
		},
		{
			name: "npm ci",
			args: args{
				cmd: []string{"npm", "ci"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNpmUnpinnedDownload(tt.args.cmd); got != tt.want {
				t.Errorf("isNpmUnpinnedDownload() = %v, want %v", got, tt.want)
			}
		})
	}
}
