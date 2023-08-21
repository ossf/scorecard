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

// nolint:stylecheck
package fuzzedWithOSSFuzz

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/internal/utils/uerror"
)

func Test_Run(t *testing.T) {
	t.Parallel()
	// nolint:govet
	tests := []struct {
		name     string
		raw      *checker.RawResults
		outcomes []finding.Outcome
		err      error
	}{
		{
			name: "fuzzer present",
			raw: &checker.RawResults{
				FuzzingResults: checker.FuzzingData{
					Fuzzers: []checker.Tool{
						{
							Name: "OSSFuzz",
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomePositive,
			},
		},
		{
			name: "fuzzer present twice",
			raw: &checker.RawResults{
				FuzzingResults: checker.FuzzingData{
					Fuzzers: []checker.Tool{
						{
							Name: "OSSFuzz",
						},
						{
							Name: "OSSFuzz",
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomePositive,
				finding.OutcomePositive,
			},
		},
		{
			name: "fuzzer present and other present",
			raw: &checker.RawResults{
				FuzzingResults: checker.FuzzingData{
					Fuzzers: []checker.Tool{
						{
							Name: "OSSFuzz",
						},
						{
							Name: "not-OSSFuzz",
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomePositive,
			},
		},
		{
			name: "fuzzer not present",
			raw: &checker.RawResults{
				FuzzingResults: checker.FuzzingData{
					Fuzzers: []checker.Tool{
						{
							Name: "not-OSSFuzz",
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeNegative,
			},
		},
		{
			name: "no fuzzer",
			raw:  &checker.RawResults{},
			outcomes: []finding.Outcome{
				finding.OutcomeNegative,
			},
		},
		{
			name: "nil raw",
			err:  uerror.ErrNil,
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
			if diff := cmp.Diff(len(tt.outcomes), len(findings)); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
			for i := range tt.outcomes {
				outcome := &tt.outcomes[i]
				f := &findings[i]
				if diff := cmp.Diff(*outcome, f.Outcome); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
