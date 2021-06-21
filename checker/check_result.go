// Copyright 2020 Security Scorecard Authors
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

	scorecarderrors "github.com/ossf/scorecard/errors"
)

const MaxResultConfidence = 10

// ErrorDemoninatorZero indicates the denominator for a proportional result is 0.
var ErrorDemoninatorZero = errors.New("internal error: denominator is 0")

type CheckResult struct {
	Error       error `json:"-"`
	Name        string
	Details     []string
	Confidence  int
	Pass        bool
	ShouldRetry bool `json:"-"`
}

func MakeInconclusiveResult(name string, err error) CheckResult {
	return CheckResult{
		Name:       name,
		Pass:       false,
		Confidence: 0,
		Error:      scorecarderrors.MakeZeroConfidenceError(err),
	}
}

func MakePassResult(name string) CheckResult {
	return CheckResult{
		Name:       name,
		Pass:       true,
		Confidence: MaxResultConfidence,
	}
}

func MakeFailResult(name string, err error) CheckResult {
	return CheckResult{
		Name:       name,
		Pass:       false,
		Confidence: MaxResultConfidence,
		Error:      err,
	}
}

func MakeRetryResult(name string, err error) CheckResult {
	return CheckResult{
		Name:        name,
		Pass:        false,
		ShouldRetry: true,
		Error:       scorecarderrors.MakeRetryError(err),
	}
}

func MakeProportionalResult(name string, numerator int, denominator int,
	threshold float32) CheckResult {
	if denominator == 0 {
		return MakeInconclusiveResult(name, ErrorDemoninatorZero)
	}
	if numerator == 0 {
		return CheckResult{
			Name:       name,
			Pass:       false,
			Confidence: MaxResultConfidence,
		}
	}
	actual := float32(numerator) / float32(denominator)
	if actual >= threshold {
		return CheckResult{
			Name:       name,
			Pass:       true,
			Confidence: int(actual * MaxResultConfidence),
		}
	}

	return CheckResult{
		Name:       name,
		Pass:       false,
		Confidence: MaxResultConfidence - int(actual*MaxResultConfidence),
	}
}

// Given a min result, check if another result is worse.
func isMinResult(result, min CheckResult) bool {
	if Bool2int(result.Pass) < Bool2int(min.Pass) {
		return true
	}
	if result.Pass && result.Confidence < min.Confidence {
		return true
	} else if !result.Pass && result.Confidence > min.Confidence {
		return true
	}
	return false
}

// MakeAndResult means all checks must succeed. This returns a conservative result
// where the worst result is returned.
func MakeAndResult(checks ...CheckResult) CheckResult {
	minResult := CheckResult{
		Pass:       true,
		Confidence: MaxResultConfidence,
	}

	for _, result := range checks {
		if minResult.Name == "" {
			minResult.Name = result.Name
		}
		if isMinResult(result, minResult) {
			minResult = result
		}
	}
	return minResult
}
