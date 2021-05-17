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
	"fmt"
	"strings"
)

const checkRetries = 3

type Runner struct {
	CheckRequest CheckRequest
}

type CheckFn func(*CheckRequest) CheckResult

type CheckNameToFnMap map[string]CheckFn

type logger struct {
	messages []string
}

func (l *logger) Logf(s string, f ...interface{}) {
	l.messages = append(l.messages, fmt.Sprintf(s, f...))
}

func (r *Runner) Run(f CheckFn) CheckResult {
	var res CheckResult
	var l logger
	for retriesRemaining := checkRetries; retriesRemaining > 0; retriesRemaining-- {
		checkRequest := r.CheckRequest
		l = logger{}
		checkRequest.Logf = l.Logf
		res = f(&checkRequest)
		if res.ShouldRetry && !strings.Contains(res.Error.Error(), "invalid header field value") {
			checkRequest.Logf("error, retrying: %s", res.Error)
			continue
		}
		break
	}
	res.Details = l.messages
	return res
}

func Bool2int(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Returns the best check result out of several ones performed.
func MultiCheckOr(fns ...CheckFn) CheckFn {
	return func(c *CheckRequest) CheckResult {
		var maxResult CheckResult

		for _, fn := range fns {
			result := fn(c)
			if Bool2int(result.Pass) < Bool2int(maxResult.Pass) {
				continue
			}
			if result.Pass && result.Confidence >= MaxResultConfidence {
				return result
			}
			if result.Confidence >= maxResult.Confidence {
				maxResult = result
			}
		}
		return maxResult
	}
}

// All checks must succeed. This returns a conservative result
// where the worst result is returned.
func MultiCheckAnd(fns ...CheckFn) CheckFn {
	return func(c *CheckRequest) CheckResult {
		minResult := CheckResult{
			Pass:       true,
			Confidence: MaxResultConfidence,
		}

		for _, fn := range fns {
			result := fn(c)
			if len(minResult.Name) == 0 {
				minResult.Name = result.Name
			}
			if Bool2int(result.Pass) < Bool2int(minResult.Pass) ||
				result.Confidence < MaxResultConfidence {
				minResult = result
			}
		}
		return minResult
	}
}
