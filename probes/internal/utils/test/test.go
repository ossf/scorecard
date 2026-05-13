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
	"strings"
	"testing"

	"github.com/ossf/scorecard/v5/checker"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
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

// AssertSecureWorkflowRemediation checks that remediation markdown points to the
// generated workflow URL without using an auto-linked bare URL as the label.
func AssertSecureWorkflowRemediation(t *testing.T, got *finding.Remediation, workflow string) {
	t.Helper()
	if got == nil {
		t.Fatal("expected remediation")
	}

	want := "Visit [StepSecurity Secure Workflow]" +
		"(https://app.stepsecurity.io/secureworkflow/github.com/ossf/scorecard/" +
		workflow + "/main?enable=permissions)."
	if !strings.Contains(got.Markdown, want) {
		t.Errorf("remediation markdown %q does not contain %q", got.Markdown, want)
	}
	if strings.Contains(got.Markdown, "[https://app.stepsecurity.io/secureworkflow]") {
		t.Errorf("remediation markdown contains a bare secureworkflow link label: %q", got.Markdown)
	}
}

// Tests for permissions-probes.
type TestData struct {
	Name     string
	Err      error
	Raw      *checker.RawResults
	Outcomes []finding.Outcome
}

func GetTests(locationType checker.PermissionLocation,
	permissionType checker.PermissionLevel,
	tokenName string,
) []TestData {
	name := tokenName // Should come from each probe test.
	value := "value"
	var wrongPermissionLocation checker.PermissionLocation
	if locationType == checker.PermissionLocationTop {
		wrongPermissionLocation = checker.PermissionLocationJob
	} else {
		wrongPermissionLocation = checker.PermissionLocationTop
	}

	return []TestData{
		{
			Name: "No Tokens",
			Raw: &checker.RawResults{
				TokenPermissionsResults: checker.TokenPermissionsData{
					NumTokens: 0,
				},
			},
			Outcomes: []finding.Outcome{
				finding.OutcomeNotApplicable,
			},
		},
		{
			Name: "Correct name",
			Raw: &checker.RawResults{
				TokenPermissionsResults: checker.TokenPermissionsData{
					NumTokens: 1,
					TokenPermissions: []checker.TokenPermission{
						{
							LocationType: &locationType,
							Name:         &name,
							Value:        &value,
							Msg:          nil,
							Type:         permissionType,
						},
					},
				},
			},
			Outcomes: []finding.Outcome{
				finding.OutcomeFalse,
			},
		},
		{
			Name: "Two tokens",
			Raw: &checker.RawResults{
				TokenPermissionsResults: checker.TokenPermissionsData{
					NumTokens: 2,
					TokenPermissions: []checker.TokenPermission{
						{
							LocationType: &locationType,
							Name:         &name,
							Value:        &value,
							Msg:          nil,
							Type:         permissionType,
						},
						{
							LocationType: &locationType,
							Name:         &name,
							Value:        &value,
							Msg:          nil,
							Type:         permissionType,
						},
					},
				},
			},
			Outcomes: []finding.Outcome{
				finding.OutcomeFalse, finding.OutcomeFalse,
			},
		},
		{
			Name: "Value is nil - Everything else correct",
			Raw: &checker.RawResults{
				TokenPermissionsResults: checker.TokenPermissionsData{
					NumTokens: 1,
					TokenPermissions: []checker.TokenPermission{
						{
							LocationType: &locationType,
							Name:         &name,
							Value:        nil,
							Msg:          nil,
							Type:         permissionType,
						},
					},
				},
			},
			Outcomes: []finding.Outcome{
				finding.OutcomeFalse,
			},
			Err: sce.ErrScorecardInternal,
		},
		{
			Name: "Wrong locationType wrong type",
			Raw: &checker.RawResults{
				TokenPermissionsResults: checker.TokenPermissionsData{
					NumTokens: 1,
					TokenPermissions: []checker.TokenPermission{
						{
							LocationType: &wrongPermissionLocation,
							Name:         &name,
							Value:        nil,
							Msg:          nil,
							Type:         checker.PermissionLevel("999"),
						},
					},
				},
			},
			Outcomes: []finding.Outcome{
				finding.OutcomeTrue,
			},
		},
		{
			Name: "Wrong locationType correct type",
			Raw: &checker.RawResults{
				TokenPermissionsResults: checker.TokenPermissionsData{
					NumTokens: 1,
					TokenPermissions: []checker.TokenPermission{
						{
							LocationType: &wrongPermissionLocation,
							Name:         &name,
							Value:        nil,
							Msg:          nil,
							Type:         permissionType,
						},
					},
				},
			},
			Outcomes: []finding.Outcome{
				finding.OutcomeTrue,
			},
		},
	}
}
