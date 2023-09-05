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
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

//nolint:lll // results are long
func Test_countScores(t *testing.T) {
	t.Parallel()

	validResults := `{"date":"0001-01-01T00:00:00Z","repo":{"name":"github.com/repo/one","commit":""},"scorecard":{"version":"","commit":""},"score":0,"checks":[{"details":null,"score":10,"reason":"no bars detected","name":"Foo"}],"metadata":null}
	{"date":"0001-01-01T00:00:00Z","repo":{"name":"github.com/repo/two","commit":""},"scorecard":{"version":"","commit":""},"score":-1,"checks":[{"details":null,"score":9,"reason":"some bars detected","name":"Foo"}],"metadata":null}
`
	invalidScore := `{"date":"0001-01-01T00:00:00Z","repo":{"name":"github.com/repo/one","commit":""},"scorecard":{"version":"","commit":""},"score":-2,"metadata":null}
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
			results: validResults,
			want:    [12]int{1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name:    "check score used when check specified",
			check:   "Foo",
			results: validResults,
			want:    [12]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1},
		},
		{
			name:    "check name case insensitive",
			check:   "fOo",
			results: validResults,
			want:    [12]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1},
		},
		{
			name:    "non existent check",
			check:   "not present",
			results: validResults,
			wantErr: true,
		},
		{
			name:    "score outside of [-1, 10] rejected",
			results: invalidScore,
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
