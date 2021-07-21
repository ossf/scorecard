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
	"strings"
	"time"

	opencensusstats "go.opencensus.io/stats"
	"go.opencensus.io/tag"

	scorecarderrors "github.com/ossf/scorecard/errors"
	"github.com/ossf/scorecard/stats"
)

const checkRetries = 3

type Runner struct {
	CheckName    string
	Repo         string
	CheckRequest CheckRequest
}

type CheckFn func(*CheckRequest) CheckResult

type CheckNameToFnMap map[string]CheckFn

// UPGRADEv2: messages2 will ultimately
// be renamed to messages.
type logger struct {
	messages  []string
	messages2 []CheckDetail
}

func (l *logger) Info(desc string, args ...interface{}) {
	cd := CheckDetail{Type: DetailInfo, Msg: fmt.Sprintf(desc, args...)}
	l.messages2 = append(l.messages2, cd)
}

func (l *logger) Warn(desc string, args ...interface{}) {
	cd := CheckDetail{Type: DetailWarn, Msg: fmt.Sprintf(desc, args...)}
	l.messages2 = append(l.messages2, cd)
}

func (l *logger) Debug(desc string, args ...interface{}) {
	cd := CheckDetail{Type: DetailDebug, Msg: fmt.Sprintf(desc, args...)}
	l.messages2 = append(l.messages2, cd)
}

// UPGRADEv2: to remove.
func (l *logger) Logf(s string, f ...interface{}) {
	l.messages = append(l.messages, fmt.Sprintf(s, f...))
}

func logStats(ctx context.Context, startTime time.Time, result *CheckResult) error {
	runTimeInSecs := time.Now().Unix() - startTime.Unix()
	opencensusstats.Record(ctx, stats.CheckRuntimeInSec.M(runTimeInSecs))

	if result.Error != nil {
		ctx, err := tag.New(ctx, tag.Upsert(stats.ErrorName, scorecarderrors.GetErrorName(result.Error)))
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		opencensusstats.Record(ctx, stats.CheckErrors.M(1))
	}
	return nil
}

func (r *Runner) Run(ctx context.Context, f CheckFn) CheckResult {
	ctx, err := tag.New(ctx, tag.Upsert(stats.CheckName, r.CheckName))
	if err != nil {
		panic(err)
	}
	startTime := time.Now()

	var res CheckResult
	var l logger
	for retriesRemaining := checkRetries; retriesRemaining > 0; retriesRemaining-- {
		checkRequest := r.CheckRequest
		checkRequest.Ctx = ctx
		l = logger{}
		// UPGRADEv2: to remove.
		checkRequest.Logf = l.Logf
		checkRequest.Dlogger = &l
		res = f(&checkRequest)
		// UPGRADEv2: to fix using proper error check.
		if res.ShouldRetry && !strings.Contains(res.Error.Error(), "invalid header field value") {
			checkRequest.Logf("error, retrying: %s", res.Error)
			continue
		}
		break
	}
	// UPGRADEv2: to remove.
	res.Details = l.messages
	res.Details2 = l.messages2

	if err := logStats(ctx, startTime, &res); err != nil {
		panic(err)
	}
	return res
}

func Bool2int(b bool) int {
	if b {
		return 1
	}
	return 0
}

func MultiCheckOr2(fns ...CheckFn) CheckFn {
	return func(c *CheckRequest) CheckResult {
		var checks []CheckResult
		for _, fn := range fns {
			res := fn(c)
			checks = append(checks, res)
		}
		return MakeOrResult(c, checks...)
	}
}

func MultiCheckAnd2(fns ...CheckFn) CheckFn {
	return func(c *CheckRequest) CheckResult {
		var checks []CheckResult
		for _, fn := range fns {
			res := fn(c)
			checks = append(checks, res)
		}
		return MakeAndResult2(checks...)
	}
}

// UPGRADEv2: will be removed.
// MultiCheckOr returns the best check result out of several ones performed.
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

// MultiCheckAnd means all checks must succeed. This returns a conservative result
// where the worst result is returned.
func MultiCheckAnd(fns ...CheckFn) CheckFn {
	return func(c *CheckRequest) CheckResult {
		var checks []CheckResult
		for _, fn := range fns {
			res := fn(c)
			checks = append(checks, res)
		}
		return MakeAndResult(checks...)
	}
}
