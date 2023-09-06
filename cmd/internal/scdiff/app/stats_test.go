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

package app

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_countScores(t *testing.T) {
	t.Parallel()

	common := `{"date":"0001-01-01T00:00:00Z","repo":{"name":"repo1"},"score":0,"checks":[{"score":10,"name":"Foo"}]}
{"date":"0001-01-01T00:00:00Z","repo":{"name":"repo2"},"score":-1,"checks":[{"score":9,"name":"Foo"}]}
`
	tests := []struct {
		name    string
		check   string
		results string
		want    [12]int
		wantErr bool
	}{
		{
			name:    "aggregate score used when no check specified",
			results: common,
			want:    [12]int{1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name:    "check score used when check specified",
			check:   "Foo",
			results: common,
			want:    [12]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1},
		},
		{
			name:    "check name case insensitive",
			check:   "fOo",
			results: common,
			want:    [12]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1},
		},
		{
			name:    "non existent check",
			check:   "not present",
			results: common,
			wantErr: true,
		},
		{
			name:    "score outside of [-1, 10] rejected",
			results: `{"date":"0001-01-01T00:00:00Z","repo":{"name":"repo1"},"score":12}`,
			wantErr: true,
		},
		{
			name:    "result fails to parse",
			results: `]}[{}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := strings.NewReader(tt.results)
			got, err := countScores(input, tt.check)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: got %v, wantedErr: %t", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("counts differ: %v", cmp.Diff(got, tt.want))
			}
		})
	}
}

func removeRepeatedSpaces(t *testing.T, s string) string {
	t.Helper()
	return strings.Join(strings.Fields(s), " ")
}

func Test_calcStats(t *testing.T) {
	t.Parallel()
	input := strings.NewReader(`{"date":"0001-01-01T00:00:00Z","repo":{"name":"repo1"},"score":10}`)
	var output bytes.Buffer
	if err := calcStats(input, &output); err != nil {
		t.Fatalf("unexepected error: %v", err)
	}
	got := output.String()
	// this is a bit of a simplification, but keeps the test simple
	// without needing to update every minor formatting change.
	got = removeRepeatedSpaces(t, got)
	want := "10 1" // score 10 should have count 1
	if !strings.Contains(got, want) {
		t.Errorf("didn't contain expected count")
	}
}
