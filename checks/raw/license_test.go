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
		name     string
		filename string
	}{
		{
			name:     "LICENSE.md",
			filename: "LICENSE.md",
		},
		{
			name:     "LICENSE",
			filename: "LICENSE",
		},
		{
			name:     "COPYING",
			filename: "COPYING",
		},
		{
			name:     "COPYING.md",
			filename: "COPYING.md",
		},
		{
			name:     "LICENSE.textile",
			filename: "LICENSE.textile",
		},
		{
			name:     "COPYING.textile",
			filename: "COPYING.textile",
		},
		{
			name:     "LICENSE-MIT",
			filename: "LICENSE-MIT",
		},
		{
			name:     "COPYING-MIT",
			filename: "COPYING-MIT",
		},
		{
			name:     "MIT-LICENSE-MIT",
			filename: "MIT-LICENSE-MIT",
		},
		{
			name:     "MIT-COPYING",
			filename: "MIT-COPYING",
		},
		{
			name:     "OFL.md",
			filename: "OFL.md",
		},
		{
			name:     "OFL.textile",
			filename: "OFL.textile",
		},
		{
			name:     "OFL",
			filename: "OFL",
		},
		{
			name:     "PATENTS",
			filename: "PATENTS",
		},
		{
			name:     "PATENTS.txt",
			filename: "PATENTS.txt",
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := TestLicense(tt.filename)
			if !s {
				t.Fail()
			}
		})
	}
}
