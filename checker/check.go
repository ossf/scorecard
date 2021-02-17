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

type CheckResult struct {
	Pass        bool
	Details     []string
	Confidence  int
	ShouldRetry bool
	Error       error
}

var InconclusiveResult = CheckResult{
	Pass:       false,
	Confidence: 0,
}

var retryResult = CheckResult{
	Pass:        false,
	ShouldRetry: true,
}

var maxConfidence int = 10

func RetryResult(err error) CheckResult {
	r := retryResult
	r.Error = err
	return r
}

type CheckFn func(Checker) CheckResult

func Bool2int(b bool) int {
	if b {
		return 1
	}
	return 0
}

func MultiCheck(fns ...CheckFn) CheckFn {
	return func(c Checker) CheckResult {
		var maxResult CheckResult

		for _, fn := range fns {
			result := fn(c)
			if Bool2int(result.Pass) < Bool2int(maxResult.Pass) {
				continue
			}
			if result.Pass && result.Confidence >= maxConfidence {
				return result
			}
			if result.Confidence >= maxResult.Confidence {
				maxResult = result
			}
		}
		return maxResult
	}
}

func ProportionalResult(numerator int, denominator int, threshold float32) CheckResult {
	if numerator == 0 {
		return CheckResult{
			Pass:       false,
			Confidence: maxConfidence,
		}
	}

	actual := float32(numerator) / float32(denominator)
	const confidence = 10
	if actual >= threshold {
		return CheckResult{
			Pass:       true,
			Confidence: int(actual * confidence),
		}
	}

	return CheckResult{
		Pass:       false,
		Confidence: maxConfidence - int(actual*confidence),
	}
}

type NamedCheck struct {
	Name string
	Fn   CheckFn
}
