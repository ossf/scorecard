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

package rule

import (
	"embed"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v4/checker"
)

func errCmp(e1, e2 error) bool {
	return errors.Is(e1, e2) || errors.Is(e2, e1)
}

//go:embed testdata/*
var testfs embed.FS

func Test_RuleNew(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		id    string
		check string
		err   error
		rule  *Rule
	}{
		{
			name: "test",
			id:   "testdata/all-fields",
			rule: &Rule{
				Name:       "testdata/all-fields",
				Short:      "short description",
				Desc:       "description",
				Motivation: "line1 line2\n",
				Risk:       checker.RiskHigh,
				Remediation: &checker.Remediation{
					Text:         "step1\nstep2 https://www.google.com/something",
					TextMarkdown: "step1\nstep2 [google.com](https://www.google.com/something)",
					Effort:       checker.RemediationEffortLow,
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, err := RuleNew(testfs, tt.id)
			if err != nil || tt.err != nil {
				if !errCmp(err, tt.err) {
					t.Fatalf("unexpected error: %v", cmp.Diff(err, tt.err, cmpopts.EquateErrors()))
				}
				return
			}

			if diff := cmp.Diff(*tt.rule, *r); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
