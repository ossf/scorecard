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

//nolint:stylecheck
package dependencyUpdateToolConfigured

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/internal/utils/test"
	"github.com/ossf/scorecard/v4/probes/internal/utils/uerror"
)

func TestRun(t *testing.T) {
	t.Parallel()
	tests := []struct {
		err      error
		raw      *checker.RawResults
		name     string
		outcomes []finding.Outcome
	}{
		{
			name: "no raw data provided is an error",
			raw:  nil,
			err:  uerror.ErrNil,
		},
		{
			name: "negative outcome from no update tools",
			raw: &checker.RawResults{
				DependencyUpdateToolResults: checker.DependencyUpdateToolData{
					Tools: []checker.Tool{},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeNegative,
			},
		},
		{
			name: "one update tool is a positive outcomes",
			raw: &checker.RawResults{
				DependencyUpdateToolResults: checker.DependencyUpdateToolData{
					Tools: []checker.Tool{
						{
							Name: "Dependabot",
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomePositive,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			findings, s, err := Run(tt.raw)
			if diff := cmp.Diff(tt.err, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(Probe, s); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
			test.AssertOutcomes(t, findings, tt.outcomes)
		})
	}
}

// for tests that want to check more than just the outcome.
func TestRun_Detailed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		err      error
		raw      *checker.RawResults
		name     string
		expected []finding.Finding
	}{
		{
			name: "update tool file locations are propagated",
			raw: &checker.RawResults{
				DependencyUpdateToolResults: checker.DependencyUpdateToolData{
					Tools: []checker.Tool{
						{
							Name: "Dependabot",
							Files: []checker.File{
								{
									Path: ".github/dependabot.yml",
								},
							},
						},
					},
				},
			},
			expected: []finding.Finding{
				{
					Probe:   Probe,
					Message: "detected update tool: Dependabot",
					Outcome: finding.OutcomePositive,
					Values: map[string]string{
						ToolKey: "Dependabot",
					},
					Location: &finding.Location{
						LineStart: asPtr(uint(0)),
						Path:      ".github/dependabot.yml",
					},
				},
			},
		},
		{
			name: "update tool name is included as tool Value",
			raw: &checker.RawResults{
				DependencyUpdateToolResults: checker.DependencyUpdateToolData{
					Tools: []checker.Tool{
						{
							Name: "some update tool",
						},
					},
				},
			},
			expected: []finding.Finding{
				{
					Probe:   Probe,
					Message: "detected update tool: some update tool",
					Outcome: finding.OutcomePositive,
					Values: map[string]string{
						ToolKey: "some update tool",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			findings, _, err := Run(tt.raw)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(findings, tt.expected); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func asPtr[T any](x T) *T {
	return &x
}
