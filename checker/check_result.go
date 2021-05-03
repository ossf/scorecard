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

const MaxResultConfidence = 10

type CheckResult struct {
	Error       error `json:"-"`
	Name        string
	Details     []string `json:",omitempty"`
	Confidence  int
	Pass        bool
	ShouldRetry bool `json:"-"`
}

func MakeInconclusiveResult(name string) CheckResult {
	return CheckResult{
		Name:       name,
		Pass:       false,
		Confidence: 0,
	}
}

func MakePassResult(name string) CheckResult {
	return CheckResult{
		Name:       name,
		Pass:       true,
		Confidence: MaxResultConfidence,
	}
}

func MakeRetryResult(name string, err error) CheckResult {
	return CheckResult{
		Name:        name,
		Pass:        false,
		ShouldRetry: true,
		Error:       err,
	}
}

func MakeProportionalResult(name string, numerator int, denominator int,
	threshold float32) CheckResult {
	if denominator == 0 {
		return MakeInconclusiveResult(name)
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
