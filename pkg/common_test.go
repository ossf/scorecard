// Copyright 2022 Security Scorecard Authors
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
	"testing"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/log"
	"github.com/ossf/scorecard/v4/remediation"
)

func TestDetailString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		detail checker.CheckDetail
		log    log.Level
		want   string
	}{
		{
			name: "ignoreDebug",
			detail: checker.CheckDetail{
				Msg: checker.LogMessage{
					Text: "should not appear",
				},
				Type: checker.DetailDebug,
			},
			log:  log.DefaultLevel,
			want: "",
		},
		{
			name: "includeDebug",
			detail: checker.CheckDetail{
				Msg: checker.LogMessage{
					Text: "should appear",
				},
				Type: checker.DetailDebug,
			},
			log:  log.DebugLevel,
			want: "Debug: should appear",
		},
		{
			name: "onlyType",
			detail: checker.CheckDetail{
				Msg: checker.LogMessage{
					Text: "some meaningful text",
				},
				Type: checker.DetailWarn,
			},
			log:  log.DefaultLevel,
			want: "Warn: some meaningful text",
		},
		{
			name: "displayPath",
			detail: checker.CheckDetail{
				Msg: checker.LogMessage{
					Text: "some meaningful text",
					Path: "Dockerfile",
				},
				Type: checker.DetailWarn,
			},
			log:  log.DefaultLevel,
			want: "Warn: some meaningful text: Dockerfile",
		},
		{
			name: "displayStartOffset",
			detail: checker.CheckDetail{
				Msg: checker.LogMessage{
					Text:   "some meaningful text",
					Path:   "Dockerfile",
					Offset: 1,
				},
				Type: checker.DetailWarn,
			},
			log:  log.DefaultLevel,
			want: "Warn: some meaningful text: Dockerfile:1",
		},
		{
			name: "displayEndOffset",
			detail: checker.CheckDetail{
				Msg: checker.LogMessage{
					Text:      "some meaningful text",
					Path:      "Dockerfile",
					Offset:    1,
					EndOffset: 7,
				},
				Type: checker.DetailWarn,
			},
			log:  log.DefaultLevel,
			want: "Warn: some meaningful text: Dockerfile:1-7",
		},
		{
			name: "ignoreInvalidEndOffset",
			detail: checker.CheckDetail{
				Msg: checker.LogMessage{
					Text:      "some meaningful text",
					Path:      "Dockerfile",
					Offset:    3,
					EndOffset: 2,
				},
				Type: checker.DetailWarn,
			},
			log:  log.DefaultLevel,
			want: "Warn: some meaningful text: Dockerfile:3",
		},
		{
			name: "includeRemediation",
			detail: checker.CheckDetail{
				Msg: checker.LogMessage{
					Text: "some meaningful text",
					Path: "Dockerfile",
					Remediation: &remediation.Remediation{
						HelpText: "fix x by doing y",
					},
				},
				Type: checker.DetailWarn,
			},
			log:  log.DefaultLevel,
			want: "Warn: some meaningful text: Dockerfile: fix x by doing y",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := DetailToString(&tt.detail, tt.log)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
