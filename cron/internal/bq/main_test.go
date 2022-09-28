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
	"testing"
)

func TestIsCompleted(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name               string
		inputExpected      int
		inputCreated       int
		completedThreshold float64
		expectedCompleted  bool
	}{
		{
			name:               "All shards completed",
			inputExpected:      2,
			inputCreated:       2,
			completedThreshold: 0.5,
			expectedCompleted:  true,
		},
		{
			name:               "No expected shards",
			inputExpected:      0,
			inputCreated:       0,
			completedThreshold: 0.9,
			expectedCompleted:  false,
		},
		{
			name:               "Completed shards same as threshold",
			inputExpected:      10,
			inputCreated:       1,
			completedThreshold: 0.1,
			expectedCompleted:  true,
		},
	}
	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()
			completed := isCompleted(testcase.inputExpected, testcase.inputCreated, testcase.completedThreshold)
			if completed != testcase.expectedCompleted {
				t.Errorf("test failed - expected: %t, got: %t", testcase.expectedCompleted, completed)
			}
		})
	}
}
