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

	"github.com/ossf/scorecard/v5/checker"
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
	if err != nil {
		t.Errorf("error validating shell file: %v", err)
	}

	if r.ProcessingErrors == nil {
		t.Errorf("failed to register shell parsing error")
	}
}

func Test_isDotNetUnpinnedDownload(t *testing.T) {
	t.Parallel()
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
			name: "nuget.exe install",
			args: args{
				cmd: []string{"nuget.exe", "install", "Newtonsoft.Json"},
			},
			want: true,
		},
		{
			name: "nuget restore",
			args: args{
				cmd: []string{"nuget", "restore"},
			},
			want: true,
		},
		{
			name: "nuget.exe restore",
			args: args{
				cmd: []string{"nuget.exe", "restore"},
			},
			want: true,
		},
		{
			name: "msbuild restore",
			args: args{
				cmd: []string{"msbuild", "/t:restore"},
			},
			want: true,
		},
		{
			name: "msbuild.exe restore",
			args: args{
				cmd: []string{"msbuild.exe", "/t:restore"},
			},
			want: true,
		},
		{
			name: "nuget restore locked",
			args: args{
				cmd: []string{"nuget", "restore", "-LockedMode"},
			},
			want: false,
		},
		{
			name: "nuget.exe restore locked",
			args: args{
				cmd: []string{"nuget.exe", "restore", "-LockedMode"},
			},
			want: false,
		},
		{
			name: "msbuild restore locked",
			args: args{
				cmd: []string{"msbuild", "/t:restore", "/p:RestoreLockedMode=true"},
			},
			want: false,
		},
		{
			name: "msbuild.exe restore locked",
			args: args{
				cmd: []string{"msbuild.exe", "/t:restore", "/p:RestoreLockedMode=true"},
			},
			want: false,
		},
		{
			name: "dotnet restore",
			args: args{
				cmd: []string{"dotnet", "restore"},
			},
			want: true,
		},
		{
			name: "dotnet.exe restore",
			args: args{
				cmd: []string{"dotnet.exe", "restore"},
			},
			want: true,
		},
		{
			name: "dotnet restore locked",
			args: args{
				cmd: []string{"dotnet", "restore", "--locked-mode"},
			},
			want: false,
		},
		{
			name: "dotnet.exe restore locked",
			args: args{
				cmd: []string{"dotnet.exe", "restore", "--locked-mode"},
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
			name: "nuget.exe install with -Version",
			args: args{
				cmd: []string{"nuget.exe", "install", "Newtonsoft.Json", "-Version", "2"},
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
			name: "nuget.exe install with packages.config",
			args: args{
				cmd: []string{"nuget.exe", "install", "config\\packages.config"},
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
			name: "dotnet.exe add",
			args: args{
				cmd: []string{"dotnet.exe", "add", "package", "Newtonsoft.Json"},
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
			name: "dotnet.exe add to project",
			args: args{
				cmd: []string{"dotnet.exe", "add", "project1", "package", "Newtonsoft.Json"},
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
			name: "dotnet.exe add reference to project",
			args: args{
				cmd: []string{"dotnet.exe", "add", "project1", "reference", "OtherProject"},
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
			name: "dotnet.exe add with -v",
			args: args{
				cmd: []string{"dotnet.exe", "add", "package", "Newtonsoft.Json", "-v", "2.0"},
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
			name: "dotnet.exe add to project with -v",
			args: args{
				cmd: []string{"dotnet.exe", "add", "project1", "package", "Newtonsoft.Json", "-v", "2.0"},
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
			name: "dotnet.exe add reference to project with -v",
			args: args{
				cmd: []string{"dotnet.exe", "add", "project1", "reference", "Newtonsoft.Json", "-v", "2.0"},
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
		{
			name: "dotnet.exe add with --version",
			args: args{
				cmd: []string{"dotnet.exe", "add", "package", "Newtonsoft.Json", "--version", "2.0"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isNugetUnpinned(tt.args.cmd); got != tt.want {
				t.Errorf("isNugetUnpinnedDownload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isGoUnpinnedDownload(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			if got := isGoUnpinnedDownload(tt.args.cmd); got != tt.want {
				t.Errorf("isGoUnpinnedDownload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isNpmDownload(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			if got := isNpmDownload(tt.args.cmd); got != tt.want {
				t.Errorf("isNpmDownload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isNpmUnpinnedDownload(t *testing.T) {
	t.Parallel()
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
		{
			name: "npm install with github: prefix and valid hash",
			args: args{
				cmd: []string{"npm", "install", "github:nodeca/js-yaml#2cef47bebf60da141b78b085f3dea3b5733dcc12"},
			},
			want: false,
		},
		{
			name: "npm install with git: prefix and valid hash",
			args: args{
				cmd: []string{"npm", "install", "git:nodeca/js-yaml#2cef47bebf60da141b78b085f3dea3b5733dcc12"},
			},
			want: true,
		},
		{
			name: "npm install with git:// prefix and valid hash",
			args: args{
				cmd: []string{"npm", "install", "git://nodeca/js-yaml#2cef47bebf60da141b78b085f3dea3b5733dcc12"},
			},
			want: false,
		},
		{
			name: "npm install with http: prefix and valid hash",
			args: args{
				cmd: []string{"npm", "install", "http://nodeca/js-yaml#2cef47bebf60da141b78b085f3dea3b5733dcc12"},
			},
			want: true,
		},
		{
			name: "npm install with https: prefix and valid hash",
			args: args{
				cmd: []string{"npm", "install", "https://nodeca/js-yaml#2cef47bebf60da141b78b085f3dea3b5733dcc12"},
			},
			want: true,
		},
		{
			name: "npm install invalid github url (has too many hash characters)",
			args: args{
				cmd: []string{"npm", "install", "githu#b:n#odeca/js-yaml#2cef47bebf60da141b78b085f3dea3b5733dcc12"},
			},
			want: true,
		},
		{
			name: "npm install with wrong prefix (githu instead of github)",
			args: args{
				cmd: []string{"npm", "install", "githu:nodeca/js-yaml#2cef47bebf60da141b78b085f3dea3b5733dcc12"},
			},
			want: true,
		},
		{
			name: "npm install with wrong hash length",
			args: args{
				cmd: []string{"npm", "install", "githu:nodeca/js-yaml#2cef47bebf60d"},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isNpmUnpinnedDownload(tt.args.cmd); got != tt.want {
				t.Errorf("isNpmUnpinnedDownload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_hasUnpinnedURLs(t *testing.T) {
	t.Parallel()
	type args struct {
		cmd []string
	}
	tests := []struct {
		name     string
		args     args
		expected bool
	}{
		{
			name: "Unpinned URL",
			args: args{
				cmd: []string{
					"curl",
					"-sSL",
					"https://dot.net/v1/dotnet-install.sh",
				},
			},
			expected: true,
		},
		{
			name: "GitHub content URL but no path",
			args: args{
				cmd: []string{
					"wget",
					"-0",
					"-",
					"https://raw.githubusercontent.com",
				},
			},
			expected: true,
		},
		{
			name: "GitHub content URL but no ref",
			args: args{
				cmd: []string{
					"wget",
					"-0",
					"-",
					"https://raw.githubusercontent.com/dotnet/install-scripts",
				},
			},
			expected: true,
		},
		{
			name: "Unpinned GitHub content URL",
			args: args{
				cmd: []string{
					"curl",
					"-sSL",
					"https://raw.githubusercontent.com/dotnet/install-scripts/main/src/dotnet-install.sh",
				},
			},
			expected: true,
		},
		{
			name: "Pinned GitHub content URL but invalid SHA",
			args: args{
				cmd: []string{
					"wget",
					"-0",
					"-",
					"https://raw.githubusercontent.com/dotnet/install-scripts/zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz/src/dotnet-install.sh",
				},
			},
			expected: true,
		},
		{
			name: "Pinned GitHub content URL but no file path",
			args: args{
				cmd: []string{
					"wget",
					"-0",
					"-",
					"https://raw.githubusercontent.com/dotnet/install-scripts/5b142a1e445a6f060d6430b661408989e9580b85",
				},
			},
			expected: true,
		},
		{
			name: "Pinned GitHub content URL",
			args: args{
				cmd: []string{
					"wget",
					"-0",
					"-",
					"https://raw.githubusercontent.com/dotnet/install-scripts/5b142a1e445a6f060d6430b661408989e9580b85/src/dotnet-install.sh",
				},
			},
			expected: false,
		},
		{
			name: "Pinned GitHub content URL but HTTP",
			args: args{
				cmd: []string{
					"wget",
					"-0",
					"-",
					"http://raw.githubusercontent.com/dotnet/install-scripts/5b142a1e445a6f060d6430b661408989e9580b85/src/dotnet-install.sh",
				},
			},
			expected: true,
		},
		{
			name: "Pinned GitHub URL but not raw content",
			args: args{
				cmd: []string{
					"wget",
					"-0",
					"-",
					"https://github.com/dotnet/install-scripts/blob/5b142a1e445a6f060d6430b661408989e9580b85/src/dotnet-install.sh",
				},
			},
			expected: true,
		},
		{
			name: "Single-quoted unpinned URL",
			args: args{
				cmd: []string{
					"curl",
					"--proto",
					"'=https'",
					"--tlsv1.2",
					"-sSf",
					"'https://raw.githubusercontent.com/rust-lang/rustup/main/rustup-init.sh'",
				},
			},
			expected: true,
		},
		{
			name: "Single-quoted pinned URL",
			args: args{
				cmd: []string{
					"curl",
					"--proto",
					"'=https'",
					"--tlsv1.2",
					"-sSf",
					"'https://raw.githubusercontent.com/rust-lang/rustup/f7935a8ad24a445629ceedb2cb706a4469e1e5b3/rustup-init.sh'",
				},
			},
			expected: false,
		},
		{
			name: "Double-quoted unpinned URL",
			args: args{
				cmd: []string{
					"curl",
					"--proto",
					"\"=https\"",
					"--tlsv1.2",
					"-sSf",
					"\"https://raw.githubusercontent.com/rust-lang/rustup/main/rustup-init.sh\"",
				},
			},
			expected: true,
		},
		{
			name: "Double-quoted pinned URL",
			args: args{
				cmd: []string{
					"curl",
					"--proto",
					"\"=https\"",
					"--tlsv1.2",
					"-sSf",
					"\"https://raw.githubusercontent.com/rust-lang/rustup/f7935a8ad24a445629ceedb2cb706a4469e1e5b3/rustup-init.sh\"",
				},
			},
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if actual := hasUnpinnedURLs(tt.args.cmd); actual != tt.expected {
				t.Errorf("hasUnpinnedURLs() = %v, expected %v for %v", actual, tt.expected, tt.name)
			}
		})
	}
}
