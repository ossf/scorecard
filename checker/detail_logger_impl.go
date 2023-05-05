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

// Logger is an implementation of the `DetailLogger` interface.
type logger struct {
	logs []CheckDetail
}

// NewLogger creates a new instance of `DetailLogger`.
func NewLogger() DetailLogger {
	return &logger{}
}

// Info emits info level logs.
func (l *logger) Info(msg *LogMessage) {
	cd := CheckDetail{
		Type: DetailInfo,
		Msg:  *msg,
	}
	l.logs = append(l.logs, cd)
}

// Warn emits warn level logs.
func (l *logger) Warn(msg *LogMessage) {
	cd := CheckDetail{
		Type: DetailWarn,
		Msg:  *msg,
	}
	l.logs = append(l.logs, cd)
}

// Debug emits debug level logs.
func (l *logger) Debug(msg *LogMessage) {
	cd := CheckDetail{
		Type: DetailDebug,
		Msg:  *msg,
	}
	l.logs = append(l.logs, cd)
}

// Flush returns existing logs and resets the logger instance.
func (l *logger) Flush() []CheckDetail {
	ret := l.Logs()
	l.logs = nil
	return ret
}

// Logs returns existing logs.
func (l *logger) Logs() []CheckDetail {
	return l.logs
}
