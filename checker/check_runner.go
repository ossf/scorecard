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
	"errors"
	"fmt"
	"time"

	opencensusstats "go.opencensus.io/stats"
	"go.opencensus.io/tag"

	sce "github.com/ossf/scorecard/v2/errors"
	"github.com/ossf/scorecard/v2/stats"
)

const checkRetries = 3

// Runner runs a check with retries.
type Runner struct {
	CheckName    string
	Repo         string
	CheckRequest CheckRequest
}

// CheckFn defined for convenience.
type CheckFn func(*CheckRequest) CheckResult

// CheckNameToFnMap defined here for convenience.
type CheckNameToFnMap map[string]CheckFn

// UPGRADEv2: messages2 will ultimately
// be renamed to messages.
type logger struct {
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

// UPGRADEv3: rename.
type logger3 struct {
	messages3 []CheckDetail3
}

func (l *logger3) Info(msg *LogMessage) {
	cd := CheckDetail3{
		Type: DetailInfo,
		Msg:  *msg,
	}
	l.messages3 = append(l.messages3, cd)
}

func (l *logger3) Warn(msg *LogMessage) {
	cd := CheckDetail3{
		Type: DetailWarn,
		Msg:  *msg,
	}
	l.messages3 = append(l.messages3, cd)
}

func (l *logger3) Debug(msg *LogMessage) {
	cd := CheckDetail3{
		Type: DetailDebug,
		Msg:  *msg,
	}
	l.messages3 = append(l.messages3, cd)
}

func logStats(ctx context.Context, startTime time.Time, result *CheckResult) error {
	runTimeInSecs := time.Now().Unix() - startTime.Unix()
	opencensusstats.Record(ctx, stats.CheckRuntimeInSec.M(runTimeInSecs))

	if result.Error != nil {
		ctx, err := tag.New(ctx, tag.Upsert(stats.ErrorName, sce.GetName(result.Error2)))
		if err != nil {
			//nolint:wrapcheck
			return sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("tag.New: %v", err))
		}
		opencensusstats.Record(ctx, stats.CheckErrors.M(1))
	}
	return nil
}

// Run runs a given check.
func (r *Runner) Run(ctx context.Context, f CheckFn) CheckResult {
	ctx, err := tag.New(ctx, tag.Upsert(stats.CheckName, r.CheckName))
	if err != nil {
		panic(err)
	}
	startTime := time.Now()

	var res CheckResult
	var l logger
	var l3 logger3
	for retriesRemaining := checkRetries; retriesRemaining > 0; retriesRemaining-- {
		checkRequest := r.CheckRequest
		checkRequest.Ctx = ctx
		l = logger{}
		l3 = logger3{}
		checkRequest.Dlogger = &l
		res = f(&checkRequest)
		if res.Error2 != nil && errors.Is(res.Error2, sce.ErrRepoUnreachable) {
			checkRequest.Dlogger.Warn("%v", res.Error2)
			continue
		}
		break
	}

	res.Details3 = l3.messages3
	res.Details2 = l.messages2
	for _, d := range l.messages2 {
		res.Details = append(res.Details, d.Msg)
	}
	if err := logStats(ctx, startTime, &res); err != nil {
		panic(err)
	}
	return res
}
