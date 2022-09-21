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

package raw

import (
	"testing"
)

func TestLicenseFileCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		filename   string
		extensions []string
	}{
		{
			name:     "LICENSE",
			filename: "LICENSE",
			extensions: []string{
				"",
				".textile",
				".txt",
				".rst",
				".PSF",
				".APACHE",
				".BSD",
				".md",
				"-MIT",
			},
		},
		{
			name:     "LICENCE",
			filename: "LICENCE",
			extensions: []string{
				"",
			},
		},
		{
			name:     "COPYING",
			filename: "COPYING",
			extensions: []string{
				"",
				".md",
				".textile",
				"-MIT",
			},
		},
		{
			name:     "MIT-LICENSE-MIT",
			filename: "MIT-LICENSE-MIT",
			extensions: []string{
				"",
			},
		},
		{
			name:     "MIT-COPYING",
			filename: "MIT-COPYING",
			extensions: []string{
				"",
			},
		},
		{
			name:     "OFL",
			filename: "OFL",
			extensions: []string{
				"",
				".md",
				".textile",
			},
		},
		{
			name:     "PATENTS",
			filename: "PATENTS",
			extensions: []string{
				"",
				".txt",
			},
		},
		{
			name:     "GPL",
			filename: "GPL",
			extensions: []string{
				"v1",
				"-1.0",
				"v2",
				"-2.0",
				"v3",
				"-3.0",
			},
		},
	}

	//nolint: paralleltest
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		for _, ext := range tt.extensions {
			name := tt.name + ext
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				s := TestLicense(name)
				if !s {
					t.Fail()
				}
			})
		}
	}
}
