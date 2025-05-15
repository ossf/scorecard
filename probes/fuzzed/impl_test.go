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

package fuzzed

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/fuzzers"
	"github.com/ossf/scorecard/v5/probes/internal/utils/test"
	"github.com/ossf/scorecard/v5/probes/internal/utils/uerror"
)

func Test_Run(t *testing.T) {
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
			name: "false outcome from no fuzzers",
			raw: &checker.RawResults{
				FuzzingResults: checker.FuzzingData{
					Fuzzers: []checker.Tool{},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeFalse,
			},
		},
		{
			name: "one fuzzer is a true outcomes",
			raw: &checker.RawResults{
				FuzzingResults: checker.FuzzingData{
					Fuzzers: []checker.Tool{
						{
							Name: fuzzers.BuiltInGo,
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue,
			},
		},
		{
			name: "same fuzzer twice results in two outcomes",
			raw: &checker.RawResults{
				FuzzingResults: checker.FuzzingData{
					Fuzzers: []checker.Tool{
						{
							Name: fuzzers.OSSFuzz,
						},
						{
							Name: fuzzers.OSSFuzz,
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue,
				finding.OutcomeTrue,
			},
		},
	}
	for _, tt := range tests {
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
			name: "fuzzer file locations are propagated",
			raw: &checker.RawResults{
				FuzzingResults: checker.FuzzingData{
					Fuzzers: []checker.Tool{
						{
							Name: fuzzers.BuiltInGo,
							Files: []checker.File{
								{
									Path: "foo.go",
								},
							},
						},
					},
				},
			},
			expected: []finding.Finding{
				{
					Probe:   Probe,
					Message: "GoBuiltInFuzzer integration found",
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						ToolKey: fuzzers.BuiltInGo,
					},
					Location: &finding.Location{
						LineStart: asPtr(uint(0)),
						Path:      "foo.go",
					},
				},
			},
		},
		{
			name: "fuzzer name is included as tool Value",
			raw: &checker.RawResults{
				FuzzingResults: checker.FuzzingData{
					Fuzzers: []checker.Tool{
						{
							Name: "some fuzzer",
						},
					},
				},
			},
			expected: []finding.Finding{
				{
					Probe:   Probe,
					Message: "some fuzzer integration found",
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						ToolKey: "some fuzzer",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			findings, _, err := Run(tt.raw)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(findings, tt.expected, cmpopts.IgnoreUnexported(finding.Finding{})); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func asPtr[T any](x T) *T {
	return &x
}
