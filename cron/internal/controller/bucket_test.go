// Copyright 2022 Security Scorecard Authors
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

package main

import (
	"context"
	"path/filepath"
	"testing"
)

//nolint:tparallel,paralleltest // since t.Setenv is used
func TestGetPrefix(t *testing.T) {
	//nolint:govet
	testcases := []struct {
		name       string
		url        string
		prefix     string
		prefixFile string
		want       string
		wantErr    bool
	}{
		{
			name:       "no prefix",
			url:        "testdata/getPrefix",
			prefix:     "",
			prefixFile: "",
			want:       "",
			wantErr:    false,
		},
		{
			name:       "prefix set",
			url:        "testdata/getPrefix",
			prefix:     "foo",
			prefixFile: "",
			want:       "foo",
			wantErr:    false,
		},
		{
			name:       "prefix file set",
			url:        "testdata/getPrefix",
			prefix:     "",
			prefixFile: "marker",
			want:       "bar",
			wantErr:    false,
		},
		{
			name:       "prefix and prefix file set",
			url:        "testdata/getPrefix",
			prefix:     "foo",
			prefixFile: "marker",
			want:       "foo",
			wantErr:    false,
		},
		{
			name:       "non existent prefix file",
			url:        "testdata/getPrefix",
			prefix:     "",
			prefixFile: "baz",
			want:       "",
			wantErr:    true,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Setenv("INPUT_BUCKET_URL", testcase.url)
			t.Setenv("INPUT_BUCKET_PREFIX", testcase.prefix)
			t.Setenv("INPUT_BUCKET_PREFIX_FILE", testcase.prefixFile)
			testdataPath, err := filepath.Abs(testcase.url)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			got, err := getPrefix(context.Background(), "file:///"+testdataPath)
			if (err != nil) != testcase.wantErr {
				t.Errorf("unexpected error produced: %v", err)
			}
			if got != testcase.want {
				t.Errorf("test failed: expected - %s, got = %s", testcase.want, got)
			}
		})
	}
}
