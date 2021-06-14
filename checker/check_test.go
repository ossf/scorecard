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

package checker

import (
	"errors"
	"testing"
)

const checkTest = "Check-Test"

var errorTest = errors.New("test error")

func TestMakeCheckAnd(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		checks []CheckResult
		want   CheckResult
	}{
		{
			name: "Multiple passing",
			checks: []CheckResult{
				{
					Name:        checkTest,
					Pass:        true,
					Details:     nil,
					Confidence:  5,
					ShouldRetry: false,
					Error:       errorTest,
				},
				{
					Name:        checkTest,
					Pass:        true,
					Details:     nil,
					Confidence:  10,
					ShouldRetry: false,
					Error:       errorTest,
				},
			},
			want: CheckResult{
				Name:        checkTest,
				Pass:        true,
				Details:     nil,
				Confidence:  5,
				ShouldRetry: false,
				Error:       nil,
			},
		},
		{
			name: "Multiple failing",
			checks: []CheckResult{
				{
					Name:        checkTest,
					Pass:        false,
					Details:     nil,
					Confidence:  10,
					ShouldRetry: false,
					Error:       errorTest,
				},
				{
					Name:        checkTest,
					Pass:        false,
					Details:     nil,
					Confidence:  5,
					ShouldRetry: false,
					Error:       errorTest,
				},
			},
			want: CheckResult{
				Name:        checkTest,
				Pass:        false,
				Details:     nil,
				Confidence:  10,
				ShouldRetry: false,
				Error:       errorTest,
			},
		},
		{
			name: "Passing and failing",
			checks: []CheckResult{
				{
					Name:        checkTest,
					Pass:        true,
					Details:     nil,
					Confidence:  10,
					ShouldRetry: false,
					Error:       nil,
				},
				{
					Name:        checkTest,
					Pass:        false,
					Details:     nil,
					Confidence:  5,
					ShouldRetry: false,
					Error:       errorTest,
				},
				{
					Name:        checkTest,
					Pass:        false,
					Details:     nil,
					Confidence:  10,
					ShouldRetry: false,
					Error:       errorTest,
				},
			},
			want: CheckResult{
				Name:        checkTest,
				Pass:        false,
				Details:     nil,
				Confidence:  10,
				ShouldRetry: false,
				Error:       errorTest,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := MakeAndResult(tt.checks...)
			if result.Pass != tt.want.Pass || result.Confidence != tt.want.Confidence {
				t.Errorf("MakeAndResult failed (%s): got %v, expected %v", tt.name, result, tt.want)
			}

			// Also test CheckFn variant
			var fns []CheckFn
			for _, c := range tt.checks {
				check := c
				fns = append(fns, func(*CheckRequest) CheckResult { return check })
			}
			c := CheckRequest{}
			resultfn := MultiCheckAnd(fns...)(&c)
			if resultfn.Pass != tt.want.Pass || resultfn.Confidence != tt.want.Confidence {
				t.Errorf("MultiCheckAnd failed (%s): got %v, expected %v", tt.name, resultfn, tt.want)
			}
		})
	}
}
