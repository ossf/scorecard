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
	"fmt"
	"io/ioutil"
	"testing"
)

func TestIsSupportedShellScriptFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		filename string
		expected bool
	}{
		{
			filename: "testdata/shell_file_awk_shebang.sh",
			expected: false,
		},
		{
			filename: "testdata/shell_file_bash_shebang1.sh",
			expected: true,
		},
		{
			filename: "testdata/shell_file_bash_shebang2.sh",
			expected: true,
		},
		{
			filename: "testdata/shell_file_bash_shebang3.sh",
			expected: true,
		},
		{
			filename: "testdata/shell_file_mksh_shebang.sh",
			expected: true,
		},
		{
			filename: "testdata/shell_file_no_shebang.sh",
			expected: true,
		},
		{
			filename: "testdata/shell_file_sh_shebang.sh",
			expected: true,
		},
		{
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
			if tt.filename == "" {
				content = make([]byte, 0)
			} else {
				content, err = ioutil.ReadFile(tt.filename)
				if err != nil {
					panic(fmt.Errorf("cannot read file: %w", err))
				}
			}
			result := isSupportedShellScriptFile(tt.filename, content)
			if result != tt.expected {
				t.Fail()
			}
		})
	}
}
