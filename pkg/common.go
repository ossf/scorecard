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

package pkg

import (
	"fmt"
	"strings"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/log"
)

func textToMarkdown(s string) string {
	return strings.ReplaceAll(s, "\n", "\n\n")
}

// DetailToString turns a detail information into a string.
func DetailToString(d *checker.CheckDetail, logLevel log.Level) string {
	if d.Type == checker.DetailDebug && logLevel != log.DebugLevel {
		return ""
	}

	switch {
	case d.Msg.Path != "" && d.Msg.Offset != 0 && d.Msg.EndOffset != 0 && d.Msg.Offset < d.Msg.EndOffset:
		return fmt.Sprintf("%s: %s: %s:%d-%d", typeToString(d.Type), d.Msg.Text, d.Msg.Path, d.Msg.Offset, d.Msg.EndOffset)
	case d.Msg.Path != "" && d.Msg.Offset != 0:
		return fmt.Sprintf("%s: %s: %s:%d", typeToString(d.Type), d.Msg.Text, d.Msg.Path, d.Msg.Offset)
	case d.Msg.Path != "" && d.Msg.Offset == 0:
		return fmt.Sprintf("%s: %s: %s", typeToString(d.Type), d.Msg.Text, d.Msg.Path)
	default:
		return fmt.Sprintf("%s: %s", typeToString(d.Type), d.Msg.Text)
	}
}

func detailsToString(details []checker.CheckDetail, logLevel log.Level) (string, bool) {
	// UPGRADEv2: change to make([]string, len(details)).
	var sa []string
	for i := range details {
		v := details[i]
		s := DetailToString(&v, logLevel)
		if s != "" {
			sa = append(sa, s)
		}
	}
	return strings.Join(sa, "\n"), len(sa) > 0
}

func typeToString(cd checker.DetailType) string {
	switch cd {
	default:
		panic("invalid detail")
	case checker.DetailInfo:
		return "Info"
	case checker.DetailWarn:
		return "Warn"
	case checker.DetailDebug:
		return "Debug"
	}
}
