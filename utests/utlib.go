// Copyright 2021 Security Scorecard Authors
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
	"testing"

	"github.com/ossf/scorecard/v2/checker"
)

func validateDetailTypes(messages []checker.CheckDetail, nw, ni, nd int) bool {
	enw := 0
	eni := 0
	end := 0
	for _, v := range messages {
		switch v.Type {
		default:
			panic(fmt.Sprintf("invalid type %v", v.Type))
		case checker.DetailInfo:
			eni++
		case checker.DetailDebug:
			end++
		case checker.DetailWarn:
			enw++
		}
	}
	return enw == nw &&
		eni == ni &&
		end == nd
}

// TestDetailLogger implements `checker.DetailLogger`.
type TestDetailLogger struct {
	messages []checker.CheckDetail
}

// TestReturn encapsulates expected CheckResult return values.
type TestReturn struct {
	Errors        []error
	Score         int
	NumberOfWarn  int
	NumberOfInfo  int
	NumberOfDebug int
}

// Info implements DetailLogger.Info.
func (l *TestDetailLogger) Info(desc string, args ...interface{}) {
	cd := checker.CheckDetail{Type: checker.DetailInfo, Msg: checker.LogMessage{Text: fmt.Sprintf(desc, args...)}}
	l.messages = append(l.messages, cd)
}

// Warn implements DetailLogger.Warn.
func (l *TestDetailLogger) Warn(desc string, args ...interface{}) {
	cd := checker.CheckDetail{Type: checker.DetailWarn, Msg: checker.LogMessage{Text: fmt.Sprintf(desc, args...)}}
	l.messages = append(l.messages, cd)
}

// Debug implements DetailLogger.Debug.
func (l *TestDetailLogger) Debug(desc string, args ...interface{}) {
	cd := checker.CheckDetail{Type: checker.DetailDebug, Msg: checker.LogMessage{Text: fmt.Sprintf(desc, args...)}}
	l.messages = append(l.messages, cd)
}

// UPGRADEv3: to rename.
//nolint:revive
func (l *TestDetailLogger) Info3(msg *checker.LogMessage) {
	cd := checker.CheckDetail{
		Type: checker.DetailInfo,
		Msg:  *msg,
	}
	cd.Msg.Version = 3
	l.messages = append(l.messages, cd)
}

//nolint:revive
func (l *TestDetailLogger) Warn3(msg *checker.LogMessage) {
	cd := checker.CheckDetail{
		Type: checker.DetailWarn,
		Msg:  *msg,
	}
	cd.Msg.Version = 3
	l.messages = append(l.messages, cd)
}

//nolint:revive
func (l *TestDetailLogger) Debug3(msg *checker.LogMessage) {
	cd := checker.CheckDetail{
		Type: checker.DetailDebug,
		Msg:  *msg,
	}
	cd.Msg.Version = 3
	l.messages = append(l.messages, cd)
}

// ValidateTestValues validates returned score and log values.
// nolint: thelper
func ValidateTestValues(t *testing.T, name string, te *TestReturn,
	score int, err error, dl *TestDetailLogger) bool {
	for _, we := range te.Errors {
		if !errors.Is(err, we) {
			if t != nil {
				t.Errorf("%v: invalid error returned: %v is not of type %v",
					name, err, we)
			}
			fmt.Printf("%v: invalid error returned: %v is not of type %v",
				name, err, we)
			return false
		}
	}
	if score != te.Score ||
		!validateDetailTypes(dl.messages, te.NumberOfWarn,
			te.NumberOfInfo, te.NumberOfDebug) {
		if t != nil {
			t.Errorf("%v: Got (score=%v) expected (%v)\n%v",
				name, score, te.Score, dl.messages)
		}
		return false
	}
	return true
}

// ValidateTestReturn validates expected TestReturn with actual checker.CheckResult values.
// nolint: thelper
func ValidateTestReturn(t *testing.T, name string, te *TestReturn,
	tr *checker.CheckResult, dl *TestDetailLogger) bool {
	return ValidateTestValues(t, name, te, tr.Score, tr.Error2, dl)
}
