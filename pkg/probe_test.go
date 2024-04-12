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

package pkg

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/ossf/scorecard/v4/finding"
)

func TestAsProbe(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		expected string
		result   ScorecardResult
	}{
		{
			name:     "multiple findings displayed",
			expected: "./testdata/probe1.json",
			result: ScorecardResult{
				Repo: RepoInfo{
					Name:      "foo",
					CommitSHA: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				},
				Scorecard: ScorecardInfo{
					Version:   "1.2.3",
					CommitSHA: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
				},
				Date: time.Date(2024, time.February, 1, 13, 48, 0, 0, time.UTC),
				Findings: []finding.Finding{
					{
						Probe:   "check for X",
						Outcome: finding.OutcomeTrue,
						Message: "found X",
						Location: &finding.Location{
							Path: "some/path/to/file",
							Type: finding.FileTypeText,
						},
					},
					{
						Probe:   "check for Y",
						Outcome: finding.OutcomeFalse,
						Message: "did not find Y",
					},
				},
			},
		},
	}
	// pretty print results so the test files are easier to read
	opt := &ProbeResultOption{Indent: "    "}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			expected, err := os.ReadFile(tt.expected)
			if err != nil {
				t.Fatalf("cannot read expected results file: %v", err)
			}

			var result bytes.Buffer
			if err = tt.result.AsProbe(&result, opt); err != nil {
				t.Fatalf("AsProbe: %v", err)
			}

			if diff := cmp.Diff(expected, result.Bytes()); diff != "" {
				t.Errorf("results differ: %s", diff)
			}
		})
	}
}
