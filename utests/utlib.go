// Copyright 2021 OpenSSF Scorecard Authors
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

// Package utests defines util fns for Scorecard unit testing.
package utests

import (
	"fmt"
	"log"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v5/checker"
)

// TestReturn encapsulates expected CheckResult return values.
type TestReturn struct {
	Error         error
	Score         int
	NumberOfWarn  int
	NumberOfInfo  int
	NumberOfDebug int
}

// TestDetailLogger implements `checker.DetailLogger`.
type TestDetailLogger struct {
	messages []checker.CheckDetail
}

// Info implements DetailLogger.Info.
func (l *TestDetailLogger) Info(msg *checker.LogMessage) {
	cd := checker.CheckDetail{
		Type: checker.DetailInfo,
		Msg:  *msg,
	}
	l.messages = append(l.messages, cd)
}

// Warn implements DetailLogger.Warn.
func (l *TestDetailLogger) Warn(msg *checker.LogMessage) {
	cd := checker.CheckDetail{
		Type: checker.DetailWarn,
		Msg:  *msg,
	}
	l.messages = append(l.messages, cd)
}

// Debug implements DetailLogger.Debug.
func (l *TestDetailLogger) Debug(msg *checker.LogMessage) {
	cd := checker.CheckDetail{
		Type: checker.DetailDebug,
		Msg:  *msg,
	}
	l.messages = append(l.messages, cd)
}

// Flush implements DetailLogger.Flush.
func (l *TestDetailLogger) Flush() []checker.CheckDetail {
	ret := l.messages
	l.messages = nil
	return ret
}

//nolint:gocritic // not worried about test helper param size
func logDetail(tb testing.TB, level string, detail checker.CheckDetail) {
	tb.Helper()
	if detail.Msg.Finding != nil {
		tb.Logf("%s: %s", level, detail.Msg.Finding.Message)
	} else {
		tb.Logf("%s: %s", level, detail.Msg.Text)
	}
}

func getTestReturn(tb testing.TB, cr *checker.CheckResult, logger *TestDetailLogger) (*TestReturn, error) {
	tb.Helper()
	ret := new(TestReturn)
	for _, v := range logger.messages {
		switch v.Type {
		default:
			//nolint:goerr113
			return nil, fmt.Errorf("invalid type %v", v.Type)
		case checker.DetailInfo:
			ret.NumberOfInfo++
			logDetail(tb, "INFO", v)
		case checker.DetailDebug:
			ret.NumberOfDebug++
			logDetail(tb, "DEBUG", v)
		case checker.DetailWarn:
			ret.NumberOfWarn++
			logDetail(tb, "WARN", v)
		}
	}
	ret.Score = cr.Score
	ret.Error = cr.Error
	return ret, nil
}

// ValidateTestReturn validates expected [TestReturn] with actual [checker.CheckResult] values.
// All test management is handled by this function.
func ValidateTestReturn(
	tb testing.TB,
	name string,
	expected *TestReturn,
	actual *checker.CheckResult,
	logger *TestDetailLogger,
) {
	tb.Helper()
	actualTestReturn, err := getTestReturn(tb, actual, logger)
	if err != nil {
		tb.Fatal(err)
	}
	if diff := cmp.Diff(*expected, *actualTestReturn, cmpopts.EquateErrors()); diff != "" {
		tb.Error(name + ": (-expected +actual)" + diff)
	}
}

// ValidatePinningDependencies tests that at least one entry returns true for isExpectedMessage.
func ValidatePinningDependencies(isExpectedDependency func(checker.Dependency) bool,
	r *checker.PinningDependenciesData,
) bool {
	for _, dep := range append(r.Dependencies, r.StagedDependencies...) {
		if isExpectedDependency(dep) {
			return true
		}
	}
	return false
}

// ValidateLogMessage tests that at least one log message returns true for isExpectedMessage.
func ValidateLogMessage(isExpectedMessage func(checker.LogMessage, checker.DetailType) bool,
	dl *TestDetailLogger,
) bool {
	for _, message := range dl.messages {
		if isExpectedMessage(message.Msg, message.Type) {
			return true
		}
	}
	return false
}

// ValidateLogMessageOffsets tests that the log message offsets match those
// in the passed in slice.
func ValidateLogMessageOffsets(dl *TestDetailLogger, offsets []uint) bool {
	if len(dl.messages) != len(offsets) {
		log.Println(cmp.Diff(dl.messages, offsets))
		return false
	}
	for i, message := range dl.messages {
		expectedOffset := offsets[i]
		if expectedOffset != message.Msg.Offset {
			log.Println(cmp.Diff(message.Msg.Offset, expectedOffset))
			return false
		}
	}
	return true
}
