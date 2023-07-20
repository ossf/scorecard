// Copyright 2020 OpenSSF Scorecard Authors
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
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/metric"

	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/stats"
)

const checkRetries = 3

// Runner runs a check with retries.
type Runner struct {
	CheckName    string
	Repo         string
	CheckRequest CheckRequest
}

// NewRunner creates a new instance of `Runner`.
func NewRunner(checkName, repo string, checkReq *CheckRequest) *Runner {
	return &Runner{
		CheckName:    checkName,
		Repo:         repo,
		CheckRequest: *checkReq,
	}
}

// SetCheckName sets the check name.
func (r *Runner) SetCheckName(check string) {
	r.CheckName = check
}

// SetRepo sets the repository.
func (r *Runner) SetRepo(repo string) {
	r.Repo = repo
}

// SetCheckRequest sets the check request.
func (r *Runner) SetCheckRequest(checkReq *CheckRequest) {
	r.CheckRequest = *checkReq
}

// CheckFn defined for convenience.
type CheckFn func(*CheckRequest) CheckResult

// Check defines a Scorecard check fn and its supported request types.
type Check struct {
	Fn                    CheckFn
	SupportedRequestTypes []RequestType
}

// CheckNameToFnMap defined here for convenience.
type CheckNameToFnMap map[string]Check

func logStats(ctx context.Context, startTime time.Time, result *CheckResult) error {
	runTimeInSecs := time.Now().Unix() - startTime.Unix()
	attr := stats.CheckName.String(result.Name)
	stats.Metrics.CheckRuntimeInSec.Record(ctx, runTimeInSecs, metric.WithAttributes(attr))

	if result.Error != nil {
		attr = stats.ErrorName.String(sce.GetName(result.Error))
		stats.Metrics.CheckErrors.Record(ctx, 1, metric.WithAttributes(attr))
	}
	return nil
}

// Run runs a given check.
func (r *Runner) Run(ctx context.Context, c Check) CheckResult {
	l := NewLogger()

	// Sanity check.
	unsupported := ListUnsupported(r.CheckRequest.RequiredTypes, c.SupportedRequestTypes)
	if len(unsupported) != 0 {
		return CreateRuntimeErrorResult(r.CheckName,
			sce.WithMessage(sce.ErrorUnsupportedCheck,
				fmt.Sprintf("requiredType: %s not supported by check %s", fmt.Sprint(unsupported), r.CheckName)))
	}

	err := stats.InitMetrics()
	if err != nil {
		panic(err)
	}

	if err != nil {
		l.Warn(&LogMessage{Text: fmt.Sprintf("tag.New: %v", err)})
	}

	ctx = context.WithValue(ctx, stats.RepoHost, r.CheckRequest.Repo.Host())
	if err != nil {
		l.Warn(&LogMessage{Text: fmt.Sprintf("tag.New: %v", err)})
	}

	startTime := time.Now()

	var res CheckResult
	for retriesRemaining := checkRetries; retriesRemaining > 0; retriesRemaining-- {
		checkRequest := r.CheckRequest
		checkRequest.Ctx = ctx
		checkRequest.Dlogger = l
		res = c.Fn(&checkRequest)
		if res.Error != nil && errors.Is(res.Error, sce.ErrRepoUnreachable) {
			checkRequest.Dlogger.Warn(&LogMessage{
				Text: fmt.Sprintf("%v", res.Error),
			})
			continue
		}
		break
	}

	// Set details.
	// TODO(#1393): Remove.
	res.Details = l.Flush()

	if err := logStats(ctx, startTime, &res); err != nil {
		panic(err)
	}
	return res
}
