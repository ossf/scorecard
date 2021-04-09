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

package lib

import (
	"fmt"
	"strings"
)

const numRetries = 3

type Runner struct {
	CheckRequest CheckRequest
}

type CheckFn func(CheckRequest) CheckResult

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
	for retriesRemaining := numRetries; retriesRemaining > 0; retriesRemaining-- {
		checkRequest := r.CheckRequest
		l = logger{}
		checkRequest.Logf = l.Logf
		res = f(checkRequest)
		if res.ShouldRetry && !strings.Contains(res.Error.Error(), "invalid header field value") {
			checkRequest.Logf("error, retrying: %s", res.Error)
			continue
		}
		break

	}
	res.Details = l.messages
	return res
}
