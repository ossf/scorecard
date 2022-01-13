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

package checks

import (
	"os"
	"testing"

	scut "github.com/ossf/scorecard/v4/utests"
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
	dl := scut.TestDetailLogger{}
	_, err = validateShellFile(filename, 0, 0, content, map[string]bool{}, &dl)
	if err != nil {
		t.Errorf("failed to discard shell parsing error: %v", err)
	}
}
