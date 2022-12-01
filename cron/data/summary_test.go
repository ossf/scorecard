// Copyright 2022 OpenSSF Scorecard Authors
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
	"context"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
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
			shards := &ShardSummary{
				shardsExpected: testcase.inputExpected,
				shardsCreated:  testcase.inputCreated,
			}
			completed := shards.IsCompleted(testcase.completedThreshold)
			if completed != testcase.expectedCompleted {
				t.Errorf("test failed - expected: %t, got: %t", testcase.expectedCompleted, completed)
			}
		})
	}
}

func TestGetBucketSummary(t *testing.T) {
	t.Parallel()
	//nolint:govet
	testcases := []struct {
		name     string
		blobPath string
		want     *BucketSummary
		wantErr  bool
	}{
		{
			name:     "basic",
			blobPath: "testdata/summary_test/basic",
			want: &BucketSummary{
				shards: map[time.Time]*ShardSummary{
					time.Date(2022, 9, 19, 2, 0, 1, 0, time.UTC): {
						creationTime:   time.Date(2022, 9, 19, 2, 0, 1, 0, time.UTC),
						shardMetadata:  []byte(`{"shardLoc":"test","numShard":3,"commitSha":"2231d1f722454c6c9aa6ad77377d2936803216ff"}`),
						shardsExpected: 3,
						shardsCreated:  2,
						isTransferred:  true,
					},
					time.Date(2022, 9, 26, 2, 0, 3, 0, time.UTC): {
						creationTime:   time.Date(2022, 9, 26, 2, 0, 3, 0, time.UTC),
						shardMetadata:  []byte(`{"shardLoc":"test","numShard":5,"commitSha":"2231d1f722454c6c9aa6ad77377d2936803216ff"}`),
						shardsExpected: 5,
						shardsCreated:  3,
						isTransferred:  false,
					},
				},
			},
			wantErr: false,
		},
		{
			name:     "invalid file present",
			blobPath: "testdata/summary_test/invalid",
			want:     nil,
			wantErr:  true,
		},
	}

	exporter := func(t reflect.Type) bool { return strings.HasPrefix(t.PkgPath(), "github.com/ossf/scorecard") }

	for i := range testcases {
		tt := &testcases[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// convert local to absolute path, which is needed for the fileblob bucket
			testdataPath, err := filepath.Abs(tt.blobPath)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			summary, err := GetBucketSummary(context.Background(), "file:///"+testdataPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("getBucketSummary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(summary, tt.want, cmp.Exporter(exporter)) {
				t.Errorf("Got diff: %s", cmp.Diff(summary, tt.want, cmp.Exporter(exporter)))
			}
		})
	}
}
