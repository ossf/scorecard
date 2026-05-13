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

package jobLevelPermissions

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/probes/internal/utils/test"
)

func Test_Run(t *testing.T) {
	t.Parallel()

	tests := test.GetTests(checker.PermissionLocationJob, checker.PermissionLevelWrite, "actions")

	tests = append(tests, test.GetTests(checker.PermissionLocationJob, checker.PermissionLevelWrite, "checks")...)
	tests = append(tests, test.GetTests(checker.PermissionLocationJob, checker.PermissionLevelWrite, "contents")...)
	tests = append(tests, test.GetTests(checker.PermissionLocationJob, checker.PermissionLevelWrite, "deployments")...)
	tests = append(tests, test.GetTests(checker.PermissionLocationJob, checker.PermissionLevelWrite, "packages")...)
	tests = append(tests, test.GetTests(checker.PermissionLocationJob, checker.PermissionLevelWrite, "security-events")...)

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			findings, s, err := Run(tt.Raw)
			if !cmp.Equal(tt.Err, err, cmpopts.EquateErrors()) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(tt.Err, err, cmpopts.EquateErrors()))
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(Probe, s); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
			test.AssertOutcomes(t, findings, tt.Outcomes)
		})
	}
}

func TestRunSecureWorkflowRemediationLink(t *testing.T) {
	t.Parallel()

	locationType := checker.PermissionLocationJob
	name := "contents"
	value := "write"
	workflow := "scorecard-analysis.yml"
	raw := &checker.RawResults{
		Metadata: checker.MetadataData{
			Metadata: map[string]string{
				"repository.uri":           "github.com/ossf/scorecard",
				"repository.defaultBranch": "main",
			},
		},
		TokenPermissionsResults: checker.TokenPermissionsData{
			NumTokens: 1,
			TokenPermissions: []checker.TokenPermission{
				{
					LocationType: &locationType,
					Name:         &name,
					Value:        &value,
					Type:         checker.PermissionLevelWrite,
					File:         &checker.File{Path: ".github/workflows/" + workflow},
				},
			},
		},
	}

	findings, _, err := Run(raw)
	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("Run() got %d findings, want 1", len(findings))
	}
	test.AssertSecureWorkflowRemediation(t, findings[0].Remediation, workflow)
}
