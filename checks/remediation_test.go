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
	"testing"

	"github.com/ossf/scorecard/checker"
)

// TestRemediationSteps tests that every check has remediation steps.
func TestRemediationSteps(t *testing.T) {
	t.Parallel()
	for checkName := range AllChecks {
		result := checker.CheckResult{
			Name:       checkName,
			Confidence: checker.MaxResultConfidence,
		}
		steps, err := GetRemediationSteps(&result)
		if err != nil {
			t.Errorf("GetRemediationSteps() for %s gave error: %s", checkName, err.Error())
		} else if len(steps) == 0 {
			// If there is a check where all the remediation steps have a code assigned, then this test should be
			// modified to use one of those codes for that check.
			t.Errorf("GetRemediationSteps() for %s returned 0 steps", checkName)
		}
	}
}
