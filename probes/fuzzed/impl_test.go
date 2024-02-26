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

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/internal/fuzzers"
	"github.com/ossf/scorecard/v4/probes/internal/utils/test"
	"github.com/ossf/scorecard/v4/probes/internal/utils/uerror"
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
			name: "negative outcome from no fuzzers",
			raw: &checker.RawResults{
				FuzzingResults: checker.FuzzingData{
					Fuzzers: []checker.Tool{},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeNegative,
			},
		},
		{
			name: "one fuzzer is a positive outcomes",
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
				finding.OutcomePositive,
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
				finding.OutcomePositive,
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

func TestRun_Location(t *testing.T) {
	t.Parallel()
	raw := &checker.RawResults{
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
	}
	findings, _, err := Run(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected one finding, got %v", findings)
	}
	if diff := cmp.Diff(findings[0].Location.Path, "foo.go"); diff != "" {
		t.Errorf("incorrect path: %s", diff)
	}
}
