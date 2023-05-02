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

package gitlab

import (
	"os"
	"testing"
)

func TestGitlabPackagingYamlCheck(t *testing.T) {
	t.Parallel()

	//nolint
	tests := []struct {
		name       string
		lineNumber uint
		filename   string
		exists     bool
	}{
		{
			name:       "No Publishing Detected",
			filename:   "./testdata/no-publishing.yaml",
			lineNumber: 1,
			exists:     false,
		},
		{
			name:       "Docker",
			filename:   "./testdata/docker.yaml",
			lineNumber: 31,
			exists:     true,
		},
		{
			name:       "Nuget",
			filename:   "./testdata/nuget.yaml",
			lineNumber: 21,
			exists:     true,
		},
		{
			name:       "Poetry",
			filename:   "./testdata/poetry.yaml",
			lineNumber: 30,
			exists:     true,
		},
		{
			name:       "Twine",
			filename:   "./testdata/twine.yaml",
			lineNumber: 26,
			exists:     true,
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error

			content, err = os.ReadFile(tt.filename)
			if err != nil {
				t.Errorf("cannot read file: %v", err)
			}

			file, found := isGitlabPackagingWorkflow(content, tt.filename)

			if tt.exists && !found {
				t.Errorf("Packaging %q should exist", tt.name)
			} else if !tt.exists && found {
				t.Errorf("No packaging information should have been found in %q", tt.name)
			}

			if file.Offset != tt.lineNumber {
				t.Errorf("Expected line number: %v != %v", tt.lineNumber, file.Offset)
			}

			if err != nil {
				return
			}
		})
	}
}
