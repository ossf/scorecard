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

package utests

import (
	"errors"
	"fmt"
	"testing"

	"github.com/ossf/scorecard/checker"
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
			eni += 1
		case checker.DetailDebug:
			end += 1
		case checker.DetailWarn:
			enw += 1
		}
	}

	return enw == nw &&
		eni == ni &&
		end == nd
}

type TestDetailLogger struct {
	messages []checker.CheckDetail
}

type TestArgs struct {
	Dl       TestDetailLogger
	Filename string
}

type TestReturn struct {
	Errors        []error
	Score         int
	NumberOfWarn  int
	NumberOfInfo  int
	NumberOfDebug int
}

type TestInfo struct {
	Args     TestArgs
	Expected TestReturn
	Name     string
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

func ValidateTest(t *testing.T, ti *TestInfo, tr checker.CheckResult) {
	for _, we := range ti.Expected.Errors {
		if !errors.Is(tr.Error2, we) {
			t.Errorf("TestDockerfileScriptDownload:\"%v\": invalid error returned: %v is not of type %v",
				ti.Name, tr.Error, we)
		}
	}
	// UPGRADEv2: update name.
	if tr.Score2 != ti.Expected.Score ||
		!validateDetailTypes(ti.Args.Dl.messages, ti.Expected.NumberOfWarn,
			ti.Expected.NumberOfInfo, ti.Expected.NumberOfDebug) {
		t.Errorf("TestDockerfileScriptDownload:\"%v\": %v. Got (score=%v) expected (%v)\n%v",
			ti.Name, ti.Args.Filename, tr.Score2, ti.Expected.Score, ti.Args.Dl.messages)
	}
}
