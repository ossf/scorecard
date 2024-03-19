// Copyright 2024 OpenSSF Scorecard Authors
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
	"testing"
)

func TestSbomFileCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		filename   string
		extensions []string
		shouldFail bool
	}{
		{
			name:       "LICENSE",
			filename:   "LICENSE",
			shouldFail: false,
			extensions: []string{
				"",
				".adoc",
				".asc",
				".docx",
				".doc",
				".ext",
				".html",
				".markdown",
				".md",
				".rst",
				".txt",
				".xml",
			},
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		for _, ext := range tt.extensions {
			name := tt.name + ext
			t.Run(name, func(t *testing.T) {
				t.Parallel()
			})
		}
	}
}
