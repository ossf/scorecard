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

package probe

import (
	"errors"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func errCmp(e1, e2 error) bool {
	return errors.Is(e1, e2) || errors.Is(e2, e1)
}

func Test_New(t *testing.T) {
	t.Parallel()
	// nolint: govet
	tests := []struct {
		name  string
		id    string
		path  string
		err   error
		probe *Probe
	}{
		{
			name: "all fields set",
			id:   "all-fields",
			path: "testdata/all-fields.yml",
			probe: &Probe{
				Name:           "all-fields",
				Short:          "short description",
				Implementation: "impl1 impl2\n",
				Motivation:     "mot1 mot2\n",
				Remediation: &Remediation{
					Text:     "step1\nstep2 https://www.google.com/something",
					Markdown: "step1\nstep2 [google.com](https://www.google.com/something)",
					Effort:   RemediationEffortLow,
				},
			},
		},
		{
			name: "mismatch probe ID",
			id:   "mismatch-id",
			path: "testdata/all-fields.yml",
			probe: &Probe{
				Name:           "all-fields",
				Short:          "short description",
				Implementation: "impl1 impl2\n",
				Motivation:     "mot1 mot2\n",
				Remediation: &Remediation{
					Text:     "step1\nstep2 https://www.google.com/something",
					Markdown: "step1\nstep2 [google.com](https://www.google.com/something)",
					Effort:   RemediationEffortLow,
				},
			},
			err: errInvalid,
		},
		{
			name: "invalid risk",
			id:   "invalid-risk",
			path: "testdata/invalid-risk.yml",
			err:  errInvalid,
		},
		{
			name: "invalid effort",
			id:   "invalid-effort",
			path: "testdata/invalid-effort.yml",
			err:  errInvalid,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			content, err := os.ReadFile(tt.path)
			if err != nil {
				panic(err)
			}

			r, err := FromBytes(content, tt.id)
			if err != nil || tt.err != nil {
				if !errCmp(err, tt.err) {
					t.Fatalf("unexpected error: %v", cmp.Diff(err, tt.err, cmpopts.EquateErrors()))
				}
				return
			}

			if diff := cmp.Diff(*tt.probe, *r); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
