// Copyright 2023 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package checker

import (
	"testing"
)

func Test_logger_Info(t *testing.T) {
	l := &logger{
		logs: []CheckDetail{},
	}
	l.Info(&LogMessage{Text: "test"})
	if len(l.logs) != 1 && l.logs[0].Type != DetailInfo {
		t.Errorf("expected 1 log, got %d", len(l.logs))
	}
}

func Test_logger_Warn(t *testing.T) {
	l := &logger{
		logs: []CheckDetail{},
	}
	l.Warn(&LogMessage{Text: "test"})
	if len(l.logs) != 1 && l.logs[0].Type != DetailWarn {
		t.Errorf("expected 1 log, got %d", len(l.logs))
	}
}

func Test_logger_Flush(t *testing.T) {
	l := &logger{
		logs: []CheckDetail{},
	}
	l.Warn(&LogMessage{Text: "test"})
	l.Flush()
	if len(l.logs) != 0 {
		t.Errorf("expected 0 log, got %d", len(l.logs))
	}
}

func Test_logger_Logs(t *testing.T) {
	l := &logger{
		logs: []CheckDetail{},
	}
	l.Warn(&LogMessage{Text: "test"})
	if len(l.Logs()) != 1 {
		t.Errorf("expected 1 log, got %d", len(l.logs))
	}
}

func Test_logger_Debug(t *testing.T) {
	l := &logger{
		logs: []CheckDetail{},
	}
	l.Debug(&LogMessage{Text: "test"})
	if len(l.logs) != 1 && l.logs[0].Type != DetailDebug {
		t.Errorf("expected 1 log, got %d", len(l.logs))
	}
}

func TestNewLogger(t *testing.T) {
	l := NewLogger()
	if l == nil {
		t.Errorf("expected non-nil logger, got nil")
	}
}
