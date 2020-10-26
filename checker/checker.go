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
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v32/github"
)

type Checker struct {
	Ctx         context.Context
	Client      *github.Client
	HttpClient  *http.Client
	Owner, Repo string
	Logf        func(s string, f ...interface{})
}

type logger struct {
	messages []string
}

func (l *logger) Logf(s string, f ...interface{}) {
	l.messages = append(l.messages, fmt.Sprintf(s, f...))
}

type Runner struct {
	Checker Checker
}

func (r *Runner) Run(f CheckFn) CheckResult {
	var res CheckResult
	var l logger
	for retriesRemaining := 3; retriesRemaining > 0; retriesRemaining-- {
		checker := r.Checker
		l = logger{}
		checker.Logf = l.Logf
		res = f(checker)
		if res.ShouldRetry {
			checker.Logf("error, retrying: %s", res.Error)
			continue
		}
		break

	}
	res.Details = l.messages
	return res
}
