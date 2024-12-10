// Copyright 2022 OpenSSF Scorecard Authors
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

package properties

import (
	"testing"
)

func TestIsValidFixedVersion(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		version string
		isFixed bool
	}{
		{"fixed version", "[10.1.1]", true},
		{"fixed beta version", "[10.1.1-beta]", true},
		{"fixed beta patch", "[10.1.1-beta.1]", true},
		{"fixed version label zzz", "[1.0.1-zzz]", true},
		{"fixed version RC with label", "[1.0.1-rc.10]", true},
		{"fixed version RC with label 2", "[1.0.1-rc.2]", true},
		{"fixed version with label open", "[1.0.1-open]", true},
		{"fixed version alpha", "[1.0.1-alpha2]", true},
		{"fixed version RC with label aaa", "[1.0.1-aaa]", true},
		{"fixed version range", "[1.0]", true},
		{"version as variable", "[$(ComponentDetectionPackageVersion)]", true},
		{"version range with inclusive min", "[1.0,)", false},
		{"version range with inclusive min without brackets", "1.0", false},
		{"version range with exclusive min", "(1.0,)", false},
		{"version range with inclusive max", "(,1.0]", false},
		{"version range with exclusive max", "[,1.0)", false},
		{"Exact range, inclusive", "[1.0,2.0]", false},
		{"Exact range, exclusive", "(1.0,2.0)", false},
		{"Mixed inclusive minimum and exclusive maximum version", "(1.0,2.0)", false},
		{"invalid", "(1.0)", false},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			isFixed := isValidFixedVersion(tt.version)
			if tt.isFixed != isFixed {
				t.Errorf("expected %v. Got %v", tt.isFixed, isFixed)
			}
		})
	}
}
