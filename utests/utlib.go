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
	"errors"
	"fmt"
	"log"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ossf/scorecard/v4/checker"
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

func getTestReturn(cr *checker.CheckResult, logger *TestDetailLogger) (*TestReturn, error) {
	ret := new(TestReturn)
	for _, v := range logger.messages {
		switch v.Type {
		default:
			//nolint: goerr113
			return nil, fmt.Errorf("invalid type %v", v.Type)
		case checker.DetailInfo:
			ret.NumberOfInfo++
		case checker.DetailDebug:
			ret.NumberOfDebug++
		case checker.DetailWarn:
			ret.NumberOfWarn++
		}
	}
	ret.Score = cr.Score
	ret.Error = cr.Error
	return ret, nil
}

func errCmp(e1, e2 error) bool {
	return errors.Is(e1, e2) || errors.Is(e2, e1)
}

// ValidateTestReturn validates expected TestReturn with actual checker.CheckResult values.
//
//nolint:thelper
func ValidateTestReturn(
	t *testing.T,
	name string,
	expected *TestReturn,
	actual *checker.CheckResult,
	logger *TestDetailLogger,
) bool {
	actualTestReturn, err := getTestReturn(actual, logger)
	if err != nil {
		panic(err)
	}
	if !cmp.Equal(*expected, *actualTestReturn, cmp.Comparer(errCmp)) {
		log.Println(name+":", cmp.Diff(*expected, *actualTestReturn))
		return false
	}
	return true
}

// ValidatePinningDependencies tests that at least one entry returns true for isExpectedMessage.
func ValidatePinningDependencies(isExpectedDependency func(checker.Dependency) bool,
	r *checker.PinningDependenciesData,
) bool {
	for _, dep := range r.Dependencies {
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
