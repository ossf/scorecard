// Copyright 2021 Security Scorecard Authors
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
