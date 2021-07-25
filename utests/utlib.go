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

type TestDetailLogger struct {
	messages []checker.CheckDetail
}

type TestReturn struct {
	Errors        []error
	Score         int
	NumberOfWarn  int
	NumberOfInfo  int
	NumberOfDebug int
}

func (l *TestDetailLogger) Info(desc string, args ...interface{}) {
	cd := checker.CheckDetail{Type: checker.DetailInfo, Msg: fmt.Sprintf(desc, args...)}
	l.messages = append(l.messages, cd)
}

func (l *TestDetailLogger) Warn(desc string, args ...interface{}) {
	cd := checker.CheckDetail{Type: checker.DetailWarn, Msg: fmt.Sprintf(desc, args...)}
	l.messages = append(l.messages, cd)
}

func (l *TestDetailLogger) Debug(desc string, args ...interface{}) {
	cd := checker.CheckDetail{Type: checker.DetailDebug, Msg: fmt.Sprintf(desc, args...)}
	l.messages = append(l.messages, cd)
}

//nolint
func ValidateTestReturn(t *testing.T, name string, te *TestReturn,
	tr *checker.CheckResult, dl *TestDetailLogger) bool {
	for _, we := range te.Errors {
		if !errors.Is(tr.Error2, we) {
			if t != nil {
				t.Errorf("%v: invalid error returned: %v is not of type %v",
					name, tr.Error, we)
			}
			return false
		}
	}
	// UPGRADEv2: update name.
	if tr.Score != te.Score ||
		!validateDetailTypes(dl.messages, te.NumberOfWarn,
			te.NumberOfInfo, te.NumberOfDebug) {
		if t != nil {
			t.Errorf("%v: Got (score=%v) expected (%v)\n%v",
				name, tr.Score, te.Score, dl.messages)
		}
		return false
	}
	return true
}
