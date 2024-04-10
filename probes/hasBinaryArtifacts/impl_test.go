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

//nolint:stylecheck
package hasBinaryArtifacts

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/internal/utils/test"
)

func Test_Run(t *testing.T) {
	t.Parallel()
	//nolint:govet
	tests := []struct {
		name     string
		raw      *checker.RawResults
		outcomes []finding.Outcome
		err      error
	}{
		{
			name: "1 binary artifact",
			raw: &checker.RawResults{
				BinaryArtifactResults: checker.BinaryArtifactData{
					Files: []checker.File{
						{
							Path: "test_binary_artifacts_check_pass",
							Type: finding.FileTypeBinary,
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue,
			},
		},
		{
			name: "No binary artifacts",
			raw: &checker.RawResults{
				BinaryArtifactResults: checker.BinaryArtifactData{},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeFalse,
			},
		},
		{
			name: "many binary artifact",
			raw: &checker.RawResults{
				BinaryArtifactResults: checker.BinaryArtifactData{
					Files: []checker.File{
						{
							Path: "test_binary_artifacts_check_pass",
							Type: finding.FileTypeBinary,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Type: finding.FileTypeBinary,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Type: finding.FileTypeBinary,
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue,
				finding.OutcomeTrue,
				finding.OutcomeTrue,
			},
		},
		{
			name: "many binary artifact including gradle wrappers",
			raw: &checker.RawResults{
				BinaryArtifactResults: checker.BinaryArtifactData{
					Files: []checker.File{
						{
							Path: "test_binary_artifacts_check_pass",
							Type: finding.FileTypeBinary,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Type: finding.FileTypeBinary,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Type: finding.FileTypeBinary,
						},
						{
							Path: "gradle-wrapper.jar",
							Type: finding.FileTypeBinaryVerified,
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue,
				finding.OutcomeTrue,
				finding.OutcomeTrue,
				finding.OutcomeTrue,
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			findings, s, err := Run(tt.raw)
			if !cmp.Equal(tt.err, err, cmpopts.EquateErrors()) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(tt.err, err, cmpopts.EquateErrors()))
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(Probe, s); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
			test.AssertOutcomes(t, findings, tt.outcomes)
		})
	}
}
