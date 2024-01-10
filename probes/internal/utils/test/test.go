// Copyright 2023 OpenSSF Scorecard Authors
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

package test

import (
	"testing"

	"github.com/ossf/scorecard/v4/finding"
)

// AssertOutcomes compares finding outcomes against expected outcomes.
func AssertOutcomes(t *testing.T, got []finding.Finding, want []finding.Outcome) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d findings, wanted %d", len(got), len(want))
	}
	for i := range got {
		if got[i].Outcome != want[i] {
			t.Errorf("got outcome %v, wanted %v for finding: %v", got[i].Outcome, want[i], got[i])
		}
	}
}
