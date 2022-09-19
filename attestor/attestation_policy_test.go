// Copyright 2022 Security Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package policy

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
	scut "github.com/ossf/scorecard/v4/utests"
)

func (a AttestationPolicy) ToJSON() string {
	jsonbytes, err := json.Marshal(a)
	if err != nil {
		return ""
	}

	return string(jsonbytes)
}

func TestCheckPreventBinaryArtifacts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                   string
		raw                    *checker.RawResults
		err                    error
		allowedBinaryArtifacts []string
		expected               PolicyResult
	}{
		{
			name: "test with no artifacts",
			raw: &checker.RawResults{
				BinaryArtifactResults: checker.BinaryArtifactData{Files: []checker.File{}},
			},
			expected: Pass,
			err:      nil,
		},
		{
			name: "test with multiple artifacts",
			raw: &checker.RawResults{
				BinaryArtifactResults: checker.BinaryArtifactData{Files: []checker.File{
					{Path: "a"},
					{Path: "b"},
				}},
			},
			expected: Fail,
			err:      nil,
		},
		{
			name:                   "test with multiple ignored artifacts",
			allowedBinaryArtifacts: []string{"a", "b"},
			raw: &checker.RawResults{
				BinaryArtifactResults: checker.BinaryArtifactData{Files: []checker.File{
					{Path: "a"},
					{Path: "b"},
				}},
			},
			expected: Pass,
			err:      nil,
		},
		{
			name:                   "test with some artifacts",
			allowedBinaryArtifacts: []string{"a"},
			raw: &checker.RawResults{
				BinaryArtifactResults: checker.BinaryArtifactData{Files: []checker.File{
					{Path: "a"},
					{Path: "b/a"},
				}},
			},
			expected: Fail,
			err:      nil,
		},

		{
			name:                   "test with glob ignored",
			allowedBinaryArtifacts: []string{"a/*", "b/*"},
			raw: &checker.RawResults{
				BinaryArtifactResults: checker.BinaryArtifactData{Files: []checker.File{
					{Path: "a/c/foo.txt"},
					{Path: "b/c/foo.txt"},
				}},
			},
			expected: Pass,
			err:      nil,
		},
	}

	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			actual, err := CheckPreventBinaryArtifacts(tt.allowedBinaryArtifacts, tt.raw, &dl)

			if !errors.Is(err, tt.err) {
				t.Fatalf("%s: expected %v, got %v", tt.name, tt.err, err)
			}
			if err != nil {
				return
			}

			// Compare outputs only if the error is nil.
			// TODO: compare objects.
			if actual != tt.expected {
				t.Fatalf("%s: invalid result", tt.name)
			}
		})
	}
}

func TestAttestationPolicyRead(t *testing.T) {
	t.Parallel()

	tests := []struct {
		err      error
		name     string
		filename string
		result   AttestationPolicy
	}{
		{
			name:     "default attestation policy with everything on",
			filename: "./testdata/policy-binauthz.yaml",
			err:      nil,
			result: AttestationPolicy{
				PreventBinaryArtifacts: true,
				AllowedBinaryArtifacts: []string{},
			},
		},
		{
			name:     "invalid attestation policy",
			filename: "./testdata/policy-binauthz-invalid.yaml",
			err:      sce.ErrScorecardInternal,
		},
		{
			name:     "policy with allowlist of binary artifacts",
			filename: "./testdata/policy-binauthz-allowlist.yaml",
			err:      nil,
			result: AttestationPolicy{
				PreventBinaryArtifacts: true,
				AllowedBinaryArtifacts: []string{"/a/b/c", "d"},
			},
		},
		{
			name:     "policy with allowlist of binary artifacts",
			filename: "./testdata/policy-binauthz-missingparam.yaml",
			err:      nil,
			result: AttestationPolicy{
				PreventBinaryArtifacts: true,
				AllowedBinaryArtifacts: nil,
			},
		},
	}

	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p, err := ParseAttestationPolicyFromFile(tt.filename)

			if !errors.Is(err, tt.err) {
				t.Fatalf("%s: expected %v, got %v", tt.name, tt.err, err)
			}
			if err != nil {
				return
			}

			// Compare outputs only if the error is nil.
			// TODO: compare objects.
			if p.ToJSON() != tt.result.ToJSON() {
				t.Fatalf("%s: invalid result", tt.name)
			}
		})
	}
}
