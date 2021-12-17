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
	"fmt"
)

// UPGRADEv2: messages2 will ultimately
// be renamed to messages.
type logger struct {
	messages2 []CheckDetail
}

func (l *logger) Info(desc string, args ...interface{}) {
	cd := CheckDetail{Type: DetailInfo, Msg: LogMessage{Text: fmt.Sprintf(desc, args...)}}
	l.messages2 = append(l.messages2, cd)
}

func (l *logger) Warn(desc string, args ...interface{}) {
	cd := CheckDetail{Type: DetailWarn, Msg: LogMessage{Text: fmt.Sprintf(desc, args...)}}
	l.messages2 = append(l.messages2, cd)
}

func (l *logger) Debug(desc string, args ...interface{}) {
	cd := CheckDetail{Type: DetailDebug, Msg: LogMessage{Text: fmt.Sprintf(desc, args...)}}
	l.messages2 = append(l.messages2, cd)
}

// UPGRADEv3: to rename.
func (l *logger) Info3(msg *LogMessage) {
	cd := CheckDetail{
		Type: DetailInfo,
		Msg:  *msg,
	}
	cd.Msg.Version = 3
	l.messages2 = append(l.messages2, cd)
}

func (l *logger) Warn3(msg *LogMessage) {
	cd := CheckDetail{
		Type: DetailWarn,
		Msg:  *msg,
	}
	cd.Msg.Version = 3
	l.messages2 = append(l.messages2, cd)
}

func (l *logger) Debug3(msg *LogMessage) {
	cd := CheckDetail{
		Type: DetailDebug,
		Msg:  *msg,
	}
	cd.Msg.Version = 3
	l.messages2 = append(l.messages2, cd)
}
