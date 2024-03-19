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

package checks

import (
	"testing"

	"github.com/ossf/scorecard/v4/checker"
	scut "github.com/ossf/scorecard/v4/utests"
)

func TestSbomFileSubdirectory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		inputFolder string
		err         error
		expected    scut.TestReturn
	}{
		{
			name:        "With Sbom in release artifacts",
			inputFolder: "testdata/sbomdir/withsbom",
			expected: scut.TestReturn{
				Error:        nil,
				Score:        checker.MaxResultScore,
				NumberOfInfo: 1,
				NumberOfWarn: 1,
			},
			err: nil,
		},
		{
			name:        "With Sbom in source",
			inputFolder: "testdata/sbomdir/withsbom",
			expected: scut.TestReturn{
				Error:        nil,
				Score:        3, // Sbom maintained in source
				NumberOfInfo: 1,
				NumberOfWarn: 1,
			},
			err: nil,
		},
		{
			name:        "Without LICENSE",
			inputFolder: "testdata/sbomdir/withoutsbom",
			expected: scut.TestReturn{
				Error:        nil,
				Score:        checker.MinResultScore,
				NumberOfWarn: 0,
				NumberOfInfo: 2,
			},
			err: nil,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
		})
	}
}
