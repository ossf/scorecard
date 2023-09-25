// Copyright 2023 OpenSSF Scorecard Authors
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
package checker

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAggregateScores(t *testing.T) {
	t.Parallel()
	type args struct {
		scores []int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "single",
			args: args{
				scores: []int{1},
			},
			want: 1,
		},
		{
			name: "multiple",
			args: args{
				scores: []int{1, 2, 3},
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := AggregateScores(tt.args.scores...); got != tt.want { //nolint:govet
				t.Errorf("AggregateScores() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAggregateScoresWithWeight(t *testing.T) {
	t.Parallel()
	type args struct {
		scores map[int]int
	}
	tests := []struct { //nolint:govet
		name string
		args args
		want int
	}{
		{
			name: "single",
			args: args{
				scores: map[int]int{1: 1},
			},
			want: 1,
		},
		{
			name: "multiple",
			args: args{
				scores: map[int]int{1: 1, 2: 2, 3: 3},
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := AggregateScoresWithWeight(tt.args.scores); got != tt.want { //nolint:govet
				t.Errorf("AggregateScoresWithWeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateProportionalScore(t *testing.T) {
	t.Parallel()
	type args struct {
		success int
		total   int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "empty",
			args: args{
				success: 0,
			},
			want: 0,
		},
		{
			name: "single",
			args: args{
				success: 1,
				total:   1,
			},
			want: 10,
		},
		{
			name: "multiple",
			args: args{
				success: 1,
				total:   2,
			},
			want: 5,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := CreateProportionalScore(tt.args.success, tt.args.total); got != tt.want { //nolint:govet
				t.Errorf("CreateProportionalScore() = %v, want %v", got, tt.want) //nolint:govet
			}
		})
	}
}

func TestCreateProportionalScoreWeighted(t *testing.T) {
	t.Parallel()
	type want struct {
		score int
		err   bool
	}
	tests := []struct {
		name   string
		scores []ProportionalScoreWeighted
		want   want
	}{
		{
			name: "max result with 1 group and normal weight",
			scores: []ProportionalScoreWeighted{
				{
					Success: 1,
					Total:   1,
					Weight:  10,
				},
			},
			want: want{
				score: 10,
			},
		},
		{
			name: "min result with 1 group and normal weight",
			scores: []ProportionalScoreWeighted{
				{
					Success: 0,
					Total:   1,
					Weight:  10,
				},
			},
			want: want{
				score: 0,
			},
		},
		{
			name: "partial result with 1 group and normal weight",
			scores: []ProportionalScoreWeighted{
				{
					Success: 2,
					Total:   10,
					Weight:  10,
				},
			},
			want: want{
				score: 2,
			},
		},
		{
			name: "partial result with 2 groups and normal weights",
			scores: []ProportionalScoreWeighted{
				{
					Success: 2,
					Total:   10,
					Weight:  10,
				},
				{
					Success: 8,
					Total:   10,
					Weight:  10,
				},
			},
			want: want{
				score: 5,
			},
		},
		{
			name: "partial result with 2 groups and odd weights",
			scores: []ProportionalScoreWeighted{
				{
					Success: 2,
					Total:   10,
					Weight:  8,
				},
				{
					Success: 8,
					Total:   10,
					Weight:  2,
				},
			},
			want: want{
				score: 3,
			},
		},
		{
			name: "all groups with 0 weight, no groups matter for the score, results in max score",
			scores: []ProportionalScoreWeighted{
				{
					Success: 1,
					Total:   1,
					Weight:  0,
				},
				{
					Success: 1,
					Total:   1,
					Weight:  0,
				},
			},
			want: want{
				score: 10,
			},
		},
		{
			name: "not all groups with 0 weight, only groups with weight matter to the score",
			scores: []ProportionalScoreWeighted{
				{
					Success: 1,
					Total:   1,
					Weight:  0,
				},
				{
					Success: 2,
					Total:   10,
					Weight:  8,
				},
				{
					Success: 8,
					Total:   10,
					Weight:  2,
				},
			},
			want: want{
				score: 3,
			},
		},
		{
			name: "no total, results in inconclusive score",
			scores: []ProportionalScoreWeighted{
				{
					Success: 0,
					Total:   0,
					Weight:  10,
				},
			},
			want: want{
				score: -1,
			},
		},
		{
			name: "some groups with 0 total, only groups with total matter to the score",
			scores: []ProportionalScoreWeighted{
				{
					Success: 0,
					Total:   0,
					Weight:  10,
				},
				{
					Success: 2,
					Total:   10,
					Weight:  10,
				},
			},
			want: want{
				score: 2,
			},
		},
		{
			name: "any group with number of successes higher than total, results in inconclusive score and error",
			scores: []ProportionalScoreWeighted{
				{
					Success: 1,
					Total:   0,
					Weight:  10,
				},
			},
			want: want{
				score: -1,
				err:   true,
			},
		},
		{
			name: "only groups with weight and total matter to the score",
			scores: []ProportionalScoreWeighted{
				{
					Success: 1,
					Total:   1,
					Weight:  0,
				},
				{
					Success: 0,
					Total:   0,
					Weight:  10,
				},
				{
					Success: 2,
					Total:   10,
					Weight:  8,
				},
				{
					Success: 8,
					Total:   10,
					Weight:  2,
				},
			},
			want: want{
				score: 3,
			},
		},
		{
			name: "only groups with weight and total matter to the score but no groups have success, results in min score",
			scores: []ProportionalScoreWeighted{
				{
					Success: 1,
					Total:   1,
					Weight:  0,
				},
				{
					Success: 0,
					Total:   0,
					Weight:  10,
				},
				{
					Success: 0,
					Total:   10,
					Weight:  8,
				},
				{
					Success: 0,
					Total:   10,
					Weight:  2,
				},
			},
			want: want{
				score: 0,
			},
		},
		{
			name: "group with 0 weight counts as max score and group with 0 total does not count",
			scores: []ProportionalScoreWeighted{
				{
					Success: 2,
					Total:   8,
					Weight:  0,
				},
				{
					Success: 0,
					Total:   0,
					Weight:  10,
				},
			},
			want: want{
				score: 10,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := CreateProportionalScoreWeighted(tt.scores...)
			if err != nil && !tt.want.err {
				t.Errorf("CreateProportionalScoreWeighted unexpected error '%v'", err)
				t.Fail()
			}
			if err == nil && tt.want.err {
				t.Errorf("CreateProportionalScoreWeighted expected error and got none")
				t.Fail()
			}
			if got != tt.want.score {
				t.Errorf("CreateProportionalScoreWeighted() = %v, want %v", got, tt.want.score)
			}
		})
	}
}

func TestNormalizeReason(t *testing.T) {
	t.Parallel()
	type args struct {
		reason string
		score  int
	}
	tests := []struct { //nolint:govet
		name string
		args args
		want string
	}{
		{
			name: "empty",
			args: args{
				reason: "",
			},
			want: " -- score normalized to 0",
		},
		{
			name: "a reason",
			args: args{
				reason: "reason",
				score:  1,
			},
			want: "reason -- score normalized to 1",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := NormalizeReason(tt.args.reason, tt.args.score); got != tt.want { //nolint:govet
				t.Errorf("NormalizeReason() = %v, want %v", got, tt.want) //nolint:govet
			}
		})
	}
}

func TestCreateResultWithScore(t *testing.T) {
	t.Parallel()
	type args struct {
		name   string
		reason string
		score  int
	}
	tests := []struct {
		name string
		args args
		want CheckResult
	}{
		{
			name: "empty",
			args: args{
				name:   "",
				reason: "",
				score:  0,
			},
			want: CheckResult{
				Name:    "",
				Reason:  "",
				Score:   0,
				Version: 2,
			},
		},
		{
			name: "a reason",
			args: args{
				name:   "name",
				reason: "reason",
				score:  1,
			},
			want: CheckResult{
				Name:    "name",
				Reason:  "reason",
				Version: 2,
				Score:   1,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := CreateResultWithScore(tt.args.name, tt.args.reason, tt.args.score); !cmp.Equal(got, tt.want) { //nolint:lll,govet
				t.Errorf("CreateResultWithScore() = %v, want %v", got, cmp.Diff(got, tt.want)) //nolint:govet
			}
		})
	}
}

func TestCreateProportionalScoreResult(t *testing.T) {
	t.Parallel()
	type args struct {
		name   string
		reason string
		b      int
		t      int
	}
	tests := []struct { //nolint:govet
		name string
		args args
		want CheckResult
	}{
		{
			name: "empty",
			args: args{
				name:   "",
				reason: "",
				b:      0,
				t:      0,
			},
			want: CheckResult{
				Name:    "",
				Reason:  " -- score normalized to 0",
				Score:   0,
				Version: 2,
			},
		},
		{
			name: "a reason",
			args: args{
				name:   "name",
				reason: "reason",
				b:      1,
				t:      1,
			},
			want: CheckResult{
				Name:    "name",
				Reason:  "reason -- score normalized to 10",
				Score:   10,
				Version: 2,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := CreateProportionalScoreResult(tt.args.name, tt.args.reason, tt.args.b, tt.args.t); !cmp.Equal(got, tt.want) { //nolint:govet,lll
				t.Errorf("CreateProportionalScoreResult() = %v, want %v", got, cmp.Diff(got, tt.want)) //nolint:govet
			}
		})
	}
}

func TestCreateMaxScoreResult(t *testing.T) {
	t.Parallel()
	type args struct {
		name   string
		reason string
	}
	tests := []struct {
		name string
		args args
		want CheckResult
	}{
		{
			name: "empty",
			args: args{
				name:   "",
				reason: "",
			},
			want: CheckResult{
				Name:    "",
				Reason:  "",
				Score:   10,
				Version: 2,
			},
		},
		{
			name: "a reason",
			args: args{
				name:   "name",
				reason: "reason",
			},
			want: CheckResult{
				Name:    "name",
				Reason:  "reason",
				Score:   10,
				Version: 2,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := CreateMaxScoreResult(tt.args.name, tt.args.reason); !cmp.Equal(got, tt.want) { //nolint:govet
				t.Errorf("CreateMaxScoreResult() = %v, want %v", got, cmp.Diff(got, tt.want))
			}
		})
	}
}

func TestCreateMinScoreResult(t *testing.T) {
	t.Parallel()
	type args struct {
		name   string
		reason string
	}
	tests := []struct {
		name string
		args args
		want CheckResult
	}{
		{
			name: "empty",
			args: args{
				name:   "",
				reason: "",
			},
			want: CheckResult{
				Name:    "",
				Reason:  "",
				Score:   0,
				Version: 2,
			},
		},
		{
			name: "a reason",
			args: args{
				name:   "name",
				reason: "reason",
			},
			want: CheckResult{
				Name:    "name",
				Reason:  "reason",
				Score:   0,
				Version: 2,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := CreateMinScoreResult(tt.args.name, tt.args.reason); !cmp.Equal(got, tt.want) { //nolint:govet
				t.Errorf("CreateMinScoreResult() = %v, want %v", got, cmp.Diff(got, tt.want))
			}
		})
	}
}

func TestCreateInconclusiveResult(t *testing.T) {
	t.Parallel()
	type args struct {
		name   string
		reason string
	}
	tests := []struct {
		name string
		args args
		want CheckResult
	}{
		{
			name: "empty",
			args: args{
				name:   "",
				reason: "",
			},
			want: CheckResult{
				Name:    "",
				Reason:  "",
				Score:   -1,
				Version: 2,
			},
		},
		{
			name: "a reason",
			args: args{
				name:   "name",
				reason: "reason",
			},
			want: CheckResult{
				Name:    "name",
				Reason:  "reason",
				Score:   -1,
				Version: 2,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := CreateInconclusiveResult(tt.args.name, tt.args.reason); !cmp.Equal(got, tt.want) {
				t.Errorf("CreateInconclusiveResult() = %v, want %v", got, cmp.Diff(got, tt.want))
			}
		})
	}
}

func TestCreateRuntimeErrorResult(t *testing.T) {
	t.Parallel()
	type args struct { //nolint:govet
		name string
		e    error
	}
	tests := []struct {
		name string
		args args
		want CheckResult
	}{
		{
			name: "empty",
			args: args{
				name: "",
				e:    errors.New("runtime error"), //nolint:goerr113
			},
			want: CheckResult{
				Name:    "",
				Reason:  "runtime error",
				Score:   -1,
				Version: 2,
				Error:   errors.New("runtime error"), //nolint:goerr113
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := CreateRuntimeErrorResult(tt.args.name, tt.args.e); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateRuntimeErrorResult() = %v, want %v", got, cmp.Diff(got, tt.want))
			}
		})
	}
}
