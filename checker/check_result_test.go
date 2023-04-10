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
			name: "empty",
			args: args{
				scores: []int{},
			},
			want: 0,
		},
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
			name: "empty",
			args: args{
				scores: map[int]int{},
			},
			want: 0,
		},
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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := CreateProportionalScore(tt.args.success, tt.args.total); got != tt.want { //nolint:govet
				t.Errorf("CreateProportionalScore() = %v, want %v", got, tt.want) //nolint:govet
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
				Name:   "",
				Reason: "",
				Score:  0,
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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := CreateMinScoreResult(tt.args.name, tt.args.reason); !cmp.Equal(got, tt.want) {
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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := CreateRuntimeErrorResult(tt.args.name, tt.args.e); !reflect.DeepEqual(got, tt.want) { //nolint:govet
				t.Errorf("CreateRuntimeErrorResult() = %v, want %v", got, cmp.Diff(got, tt.want)) //nolint:govet
			}
		})
	}
}
