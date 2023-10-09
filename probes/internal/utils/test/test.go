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

	"github.com/google/go-cmp/cmp"

	"github.com/ossf/scorecard/v4/finding"
)

// AssertCorrect is a suitable for all probes to compare raw
// results against expected outcomes.
func AssertCorrect(t *testing.T, probeExpect, probeGot string,
	findings []finding.Finding, expectedOutcomes []finding.Outcome,
) {
	t.Helper()
	if diff := cmp.Diff(probeExpect, probeGot); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(len(expectedOutcomes), len(findings)); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
	for i := range expectedOutcomes {
		outcome := &expectedOutcomes[i]
		f := &findings[i]
		if diff := cmp.Diff(*outcome, f.Outcome); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
	}
}
