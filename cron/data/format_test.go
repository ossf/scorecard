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

package data

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestToString(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name   string
		input  CSVStrings
		output []string
	}{
		{
			name:   "Basic",
			input:  []string{"str1", "str2"},
			output: []string{"str1", "str2"},
		},
		{
			name:   "NilInput",
			input:  nil,
			output: nil,
		},
		{
			name:   "EmptyString",
			input:  []string{""},
			output: []string{""},
		},
		{
			name:   "EmptySlice",
			input:  make([]string, 0),
			output: nil,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()
			actual := testcase.input.ToString()
			if !cmp.Equal(testcase.output, actual) {
				t.Errorf("testcase failed: expected equal, got diff: %s", cmp.Diff(testcase.output, actual))
			}
		})
	}
}

func TestUnmarshalCsv(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name   string
		input  []byte
		output CSVStrings
	}{
		{
			name:   "Basic",
			input:  []byte("str1,str2"),
			output: []string{"str1", "str2"},
		},
		{
			name:   "NilInput",
			input:  nil,
			output: nil,
		},
		{
			name:   "EmptyString",
			input:  []byte(""),
			output: nil,
		},
		{
			name:   "EmptySlice",
			input:  make([]byte, 0),
			output: nil,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()
			s := new(CSVStrings)
			if err := s.UnmarshalCSV(testcase.input); err != nil {
				t.Errorf("testcase failed: %v", err)
			}
			if !cmp.Equal(testcase.output, *s) {
				t.Errorf("testcase failed: expected - %q, got - %q", testcase.output, *s)
			}
		})
	}
}
