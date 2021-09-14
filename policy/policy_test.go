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
	"io/ioutil"
	"reflect"
	"testing"

	sce "github.com/ossf/scorecard/v2/errors"
)

//nolint
func TestPolicyRead(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		result   ScorecardPolicy
		err      error
	}{
		{
			name:     "correct",
			filename: "./testdata/policy-ok.yaml",
			err:      nil,
			result: ScorecardPolicy{
				Version: 1,
				Policies: map[string]CheckPolicy{
					"Token-Permissions": {
						Score: 3,
						Mode:  "disabled",
					},
					"Branch-Protection": {
						Score: 5,
						Mode:  "enforced",
					},
					"Vulnerabilities": {
						Score: 1,
						Mode:  "logging",
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

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error
			content, err = ioutil.ReadFile(tt.filename)
			if err != nil {
				t.Fatalf("cannot read file: %v", err)
			}

			var p ScorecardPolicy
			err = p.Read(content)

			if !errors.Is(err, tt.err) {
				t.Fatalf("%s: expected %v, got %v", tt.name, tt.err, err)
			}
			if err != nil {
				return
			}

			// Compare outputs only if the error is nil.
			if !reflect.DeepEqual(p, tt.result) {
				t.Fatalf("%s: invalid result", tt.name)
			}
		})
	}
}
