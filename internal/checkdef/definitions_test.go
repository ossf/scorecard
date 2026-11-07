// Copyright 2026 OpenSSF Scorecard Authors
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

package checkdef

import (
	"slices"
	"testing"

	"github.com/ossf/scorecard/v5/internal/checknames"
)

// TestDefinitionNames requires exactly one definition for each name in
// checknames, and no definition under any other name.
func TestDefinitionNames(t *testing.T) {
	t.Parallel()
	count := map[string]int{}
	for _, d := range definitions {
		count[d.name]++
	}
	for _, name := range checknames.AllValidChecks {
		if n := count[name]; n != 1 {
			t.Errorf("check %q has %d definitions, want 1", name, n)
		}
	}
	for name := range count {
		if !slices.Contains(checknames.AllValidChecks, name) {
			t.Errorf("definition %q is not in checknames.AllValidChecks", name)
		}
	}
}
