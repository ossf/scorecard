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

package format

import (
	"bytes"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_GenerateBQSchema(t *testing.T) {
	t.Parallel()

	//nolint:govet
	tests := []struct {
		name      string
		path      string
		structure interface{}
	}{
		{
			name: "valid structure",
			path: "testdata/bq-valid.schema",
			structure: struct {
				String string `bigquery:"string"`
				Bool   bool   `bigquery:"bool"`
				Number int    `bigquery:"number"`
				Struct struct {
					String string `bigquery:"string"`
					Bool   bool   `bigquery:"bool"`
					Number int    `bigquery:"number"`
				} `bigquery:"Struct"`
			}{},
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var content []byte
			var err error
			content, err = os.ReadFile(tt.path)
			if err != nil {
				t.Fatalf("cannot read file: %v", err)
			}

			var expected bytes.Buffer
			n, err := expected.Write(content)
			if err != nil {
				t.Fatalf("%s: cannot write buffer: %v", tt.name, err)
			}
			if n != len(content) {
				t.Fatalf("%s: write %d bytes but expected %d", tt.name, n, len(content))
			}

			result, err := GenerateBQSchema(tt.structure)
			if err != nil {
				t.Fatalf("%s: GenerateBQSchema: %v", tt.name, err)
			}

			if !cmp.Equal(result, expected.String()) {
				t.Errorf(cmp.Diff(result, expected.String()))
			}
		})
	}
}

func Test_GenerateJSONSchema(t *testing.T) {
	t.Parallel()

	//nolint:govet
	tests := []struct {
		name      string
		path      string
		structure interface{}
	}{
		{
			name: "valid structure",
			path: "testdata/valid.schema",
			structure: struct {
				String string `json:"string"`
				Bool   bool   `json:"bool"`
				Number int    `json:"number"`
				Struct struct {
					String string `json:"string"`
					Bool   bool   `json:"bool"`
					Number int    `json:"number"`
				} `json:"Struct"`
			}{},
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var content []byte
			var err error
			content, err = os.ReadFile(tt.path)
			if err != nil {
				t.Fatalf("cannot read file: %v", err)
			}

			var expected bytes.Buffer
			n, err := expected.Write(content)
			if err != nil {
				t.Fatalf("%s: cannot write buffer: %v", tt.name, err)
			}
			if n != len(content) {
				t.Fatalf("%s: write %d bytes but expected %d", tt.name, n, len(content))
			}

			result := GenerateJSONSchema(tt.structure)

			if !cmp.Equal(result, expected.String()) {
				t.Errorf(cmp.Diff(result, expected.String()))
			}
		})
	}
}
