// Copyright 2021 OpenSSF Scorecard Authors
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

package policy

import (
	"errors"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks"
	sce "github.com/ossf/scorecard/v4/errors"
)

func TestPolicyRead(t *testing.T) {
	t.Parallel()

	//nolint
	tests := []struct {
		err      error
		name     string
		filename string
		result   ScorecardPolicy
	}{
		{
			name:     "correct",
			filename: "./testdata/policy-ok.yaml",
			err:      nil,
			result: ScorecardPolicy{
				Version: 1,
				Policies: map[string]*CheckPolicy{
					"Token-Permissions": {
						Score: 3,
						Mode:  CheckPolicy_DISABLED,
					},
					"Branch-Protection": {
						Score: 5,
						Mode:  CheckPolicy_ENFORCED,
					},
					"Vulnerabilities": {
						Score: 1,
						Mode:  CheckPolicy_ENFORCED,
					},
				},
			},
		},
		{
			name:     "no score disabled",
			filename: "./testdata/policy-no-score-disabled.yaml",
			err:      nil,
			result: ScorecardPolicy{
				Version: 1,
				Policies: map[string]*CheckPolicy{
					"Token-Permissions": {
						Score: 0,
						Mode:  CheckPolicy_DISABLED,
					},
					"Branch-Protection": {
						Score: 5,
						Mode:  CheckPolicy_ENFORCED,
					},
					"Vulnerabilities": {
						Score: 1,
						Mode:  CheckPolicy_ENFORCED,
					},
				},
			},
		},
		{
			name:     "invalid score - 0",
			filename: "./testdata/policy-invalid-score-0.yaml",
			err:      sce.ErrScorecardInternal,
		},
		{
			name:     "invalid score + 10",
			filename: "./testdata/policy-invalid-score-10.yaml",
			err:      sce.ErrScorecardInternal,
		},
		{
			name:     "invalid mode",
			filename: "./testdata/policy-invalid-mode.yaml",
			err:      sce.ErrScorecardInternal,
		},
		{
			name:     "invalid check name",
			filename: "./testdata/policy-invalid-check.yaml",
			err:      sce.ErrScorecardInternal,
		},
		{
			name:     "multiple check definitions",
			filename: "./testdata/policy-multiple-defs.yaml",
			err:      sce.ErrScorecardInternal,
		},
	}

	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error
			content, err = os.ReadFile(tt.filename)
			if err != nil {
				t.Fatalf("cannot read file: %v", err)
			}

			p, err := parseFromYAML(content)

			if !errors.Is(err, tt.err) {
				t.Fatalf("%s: expected %v, got %v", tt.name, tt.err, err)
			}
			if err != nil {
				return
			}

			// Compare outputs only if the error is nil.
			// TODO: compare objects.
			if p.String() != tt.result.String() {
				t.Fatalf("%s: invalid result", tt.name)
			}
		})
	}
}

func TestChecksHavePolicies(t *testing.T) {
	// Create a sample ScorecardPolicy
	sp := &ScorecardPolicy{
		Version: 1,
		Policies: map[string]*CheckPolicy{
			"Binary-Artifacts": {
				// Set fields of the CheckPolicy struct accordingly
			},
		},
	}
	check := checker.CheckNameToFnMap{
		"Binary-Artifacts": checker.Check{
			Fn: checks.BinaryArtifacts,
		},
	}

	// Call the function being tested
	result := checksHavePolicies(sp, check)

	// Assert the result
	if !result {
		t.Error("Expected checks to have policies")
	}

	delete(sp.Policies, "Binary-Artifacts")
	// Call the function being tested
	result = checksHavePolicies(sp, check)

	if result {
		t.Error("Expected checks to have no policies")
	}
}

func TestEnableCheck(t *testing.T) {
	t.Parallel()

	// Create a sample check name
	checkName := "Binary-Artifacts"

	// Create a sample enabled checks map
	enabledChecks := make(checker.CheckNameToFnMap)

	// Call the function being tested
	result := enableCheck(checkName, &enabledChecks)

	// Assert the result
	if !result {
		t.Error("Expected the check to be enabled")
	}
	if _, ok := enabledChecks[checkName]; !ok {
		t.Error("Expected the check to be added to enabled checks")
	}

	// Try enabling a check that does not exist
	nonExistentCheck := "Non-Existent-Check"
	result = enableCheck(nonExistentCheck, &enabledChecks)

	// Assert the result
	if result {
		t.Error("Expected the check to not be enabled")
	}
	if _, ok := enabledChecks[nonExistentCheck]; ok {
		t.Error("Expected the check to not be added to enabled checks")
	}
}

func TestIsSupportedCheck(t *testing.T) {
	t.Parallel()

	// Create a sample check name
	checkName := "Binary-Artifacts"

	// Create a sample list of required request types
	requiredRequestTypes := []checker.RequestType{
		checker.FileBased,
	}

	// Call the function being tested
	result := isSupportedCheck(checkName, requiredRequestTypes)

	// Assert the result
	expectedResult := true
	if result != expectedResult {
		t.Errorf("Unexpected result: got %v, want %v", result, expectedResult)
	}

	// Try with an unsupported check
	unsupportedCheckName := "Unsupported-Check"
	result = isSupportedCheck(unsupportedCheckName, requiredRequestTypes)

	// Assert the result
	expectedResult = false
	if diff := cmp.Diff(result, expectedResult); diff != "" {
		t.Errorf("Unexpected result (-got +want):\n%s", diff)
	}

	// Additional test cases can be added to cover more scenarios
}

func TestParseFromFile(t *testing.T) {
	t.Parallel()

	// Provide the path to the policy file
	policyFile := "testdata/policy-ok.yaml"

	// Call the function being tested
	sp, err := ParseFromFile(policyFile)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(sp.Policies) != 3 {
		t.Errorf("Unexpected number of policies: got %v, want %v", len(sp.Policies), 3)
	}
	invalidPolicy := "testdata/policy-invalid-score-0.yaml"
	_, err = ParseFromFile(invalidPolicy)
	if err == nil {
		t.Error("Expected an error")
	}
	invalidMode := "testdata/policy-invalid-mode.yaml"
	_, err = ParseFromFile(invalidMode)
	if err == nil {
		t.Error("Expected an error")
	}
	invalidFile := "testdata/invalid-file.yaml"
	_, err = ParseFromFile(invalidFile)
	if err == nil {
		t.Error("Expected an error")
	}
}

func TestModeToProto(t *testing.T) {
	t.Parallel()

	// Call the function being tested
	mode := modeToProto("enforced")

	// Check the result
	expectedMode := CheckPolicy_ENFORCED
	if mode != expectedMode {
		t.Errorf("Unexpected mode. Got: %v, Want: %v", mode, expectedMode)
	}

	// Call the function again with a different mode
	mode = modeToProto("disabled")

	// Check the result
	expectedMode = CheckPolicy_DISABLED
	if mode != expectedMode {
		t.Errorf("Unexpected mode. Got: %v, Want: %v", mode, expectedMode)
	}

	// Test panic with an unknown mode
	testPanic := func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic, but no panic occurred")
			}
		}()
		modeToProto("unknown")
	}

	// Run the panic test
	testPanic()

	// Additional test cases can be added to cover more scenarios
}

func TestGetEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		policyFile            string
		argsChecks            []string
		requiredRequestTypes  []checker.RequestType
		expectedEnabledChecks int
		expectedError         bool
	}{
		{
			name:                  "checks limited to those specified by checks arg",
			argsChecks:            []string{"Binary-Artifacts"},
			requiredRequestTypes:  []checker.RequestType{checker.FileBased},
			expectedEnabledChecks: 1,
			expectedError:         false,
		},
		{
			name:                  "mix of supported and unsupported checks",
			argsChecks:            []string{"Binary-Artifacts", "UnsupportedCheck"},
			requiredRequestTypes:  []checker.RequestType{checker.FileBased, checker.CommitBased},
			expectedEnabledChecks: 1,
			expectedError:         true,
		},
		{
			name:                  "request types limit enabled checks",
			argsChecks:            []string{},
			requiredRequestTypes:  []checker.RequestType{checker.FileBased, checker.CommitBased},
			expectedEnabledChecks: 5, // All checks which are FileBased and CommitBased
			expectedError:         false,
		},
		{
			name:                  "all checks in policy file enabled",
			policyFile:            "testdata/policy-ok.yaml",
			argsChecks:            []string{},
			requiredRequestTypes:  []checker.RequestType{},
			expectedEnabledChecks: 3,
			expectedError:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sp *ScorecardPolicy
			if tt.policyFile != "" {
				policyBytes, err := os.ReadFile(tt.policyFile)
				if err != nil {
					t.Fatalf("reading policy file: %v", err)
				}
				pol, err := parseFromYAML(policyBytes)
				if err != nil {
					t.Fatalf("parsing policy file: %v", err)
				}
				sp = pol
			}

			enabledChecks, err := GetEnabled(sp, tt.argsChecks, tt.requiredRequestTypes)

			if len(enabledChecks) != tt.expectedEnabledChecks {
				t.Errorf("Unexpected number of enabled checks: got %v, want %v", len(enabledChecks), tt.expectedEnabledChecks)
			}
			if tt.expectedError && err == nil {
				t.Errorf("Expected an error, but got none")
			} else if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
