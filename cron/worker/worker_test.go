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

package worker

import (
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ossf/scorecard/v4/cron/data"
)

func asIntPointer(i int32) *int32 {
	return &i
}

func TestResultFilename(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name string
		req  *data.ScorecardBatchRequest
		want string
	}{
		{
			name: "Basic",
			req: &data.ScorecardBatchRequest{
				JobTime:  timestamppb.New(time.Date(1979, time.October, 12, 1, 2, 3, 0, time.UTC)),
				ShardNum: asIntPointer(42),
			},
			want: "1979.10.12/010203/shard-0000042",
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()
			got := ResultFilename(testcase.req)
			if got != testcase.want {
				t.Errorf("\nexpected: \n%s \ngot: \n%s", testcase.want, got)
			}
		})
	}
}
