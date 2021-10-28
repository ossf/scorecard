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

package data

import (
	"errors"
	"testing"
	"time"
)

const (
	inputTimeFormat string = "2006-01-02T15:04:05"
)

func TestGetBlobFilename(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name             string
		inputTime        string
		inputFilename    string
		expectedFilename string
	}{
		{
			name:             "Basic",
			inputTime:        "2021-04-23T15:06:43",
			inputFilename:    "file-000",
			expectedFilename: "2021.04.23/150643/file-000",
		},
	}
	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()
			datetime, err := time.Parse(inputTimeFormat, testcase.inputTime)
			if err != nil {
				t.Errorf("failed to parse testcase.inputTime %s: %v", testcase.inputTime, err)
			}
			gotFilename := GetBlobFilename(testcase.inputFilename, datetime)
			if gotFilename != testcase.expectedFilename {
				t.Errorf("test failed - expected: %s, got: %s", testcase.expectedFilename, gotFilename)
			}
		})
	}
}

func TestParseBlobFilename(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name         string
		input        string
		err          error
		expectedTime time.Time
		expectedName string
	}{
		{
			name:         "Basic",
			input:        "2021.06.09/165503/shard-00010",
			expectedTime: time.Date(2021, 6, 9, 16, 55, 3, 0, time.UTC),
			expectedName: "shard-00010",
		},
		{
			name:         "NoSuffix",
			input:        "2021.06.09/165503/",
			expectedTime: time.Date(2021, 6, 9, 16, 55, 3, 0, time.UTC),
			expectedName: "",
		},
		{
			name:  "ParseError",
			input: "2021.06.09/shard-00010",
			err:   errParseBlobName,
		},
		{
			name:         "SubDirectory",
			input:        "2021.06.09/165503/shards/shard-00010",
			expectedTime: time.Date(2021, 6, 9, 16, 55, 3, 0, time.UTC),
			expectedName: "shards/shard-00010",
		},
		{
			name:  "NoTrailingSlash",
			input: "2021.06.09/165503",
			err:   errShortBlobName,
		},
	}
	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()
			datetime, filename, err := ParseBlobFilename(testcase.input)
			if !errors.Is(err, testcase.err) {
				t.Errorf("Expected: %v, got: %v", testcase.err, err)
			}
			if testcase.err == nil {
				if datetime != testcase.expectedTime {
					t.Errorf("Expected: %q, got: %q", testcase.expectedTime, datetime)
				}
				if filename != testcase.expectedName {
					t.Errorf("Expected: %s, got: %s", testcase.expectedName, filename)
				}
			}
		})
	}
}
