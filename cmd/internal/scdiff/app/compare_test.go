// Copyright 2023 OpenSSF Scorecard Authors
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

package app

import (
	"io"
	"os"
	"strings"
	"testing"
)

//nolint:lll // results are long
func Test_compare(t *testing.T) {
	t.Parallel()
	//nolint:govet // struct alignment
	tests := []struct {
		name          string
		x             string
		y             string
		match         bool
		wantErrSubstr string
	}{
		{
			name: "results match",
			x: `{"date":"2023-08-11T10:22:43-07:00","repo":{"name":"github.com/foo/bar","commit":"f0840f7158c8044af2bd9b8aa661d7942b1f29d2"},"scorecard":{"version":"","commit":"unknown"},"score":10.0,"checks":[{"details":null,"score":10,"reason":"no vulnerabilities detected","name":"Vulnerabilities","documentation":{"url":"https://github.com/ossf/scorecard/blob/main/docs/checks.md#vulnerabilities","short":"Determines if the project has open, known unfixed vulnerabilities."}}],"metadata":null}
`,
			y: `{"date":"2023-08-11T10:22:43-07:00","repo":{"name":"github.com/foo/bar","commit":"f0840f7158c8044af2bd9b8aa661d7942b1f29d2"},"scorecard":{"version":"","commit":"unknown"},"score":10.0,"checks":[{"details":null,"score":10,"reason":"no vulnerabilities detected","name":"Vulnerabilities","documentation":{"url":"https://github.com/ossf/scorecard/blob/main/docs/checks.md#vulnerabilities","short":"Determines if the project has open, known unfixed vulnerabilities."}}],"metadata":null}
`,
			match: true,
		},
		{
			name: "unequal number of results",
			x: `{"date":"2023-08-11T10:22:43-07:00","repo":{"name":"github.com/foo/bar","commit":"f0840f7158c8044af2bd9b8aa661d7942b1f29d2"},"scorecard":{"version":"","commit":"unknown"},"score":10.0,"checks":[{"details":null,"score":10,"reason":"no vulnerabilities detected","name":"Vulnerabilities","documentation":{"url":"https://github.com/ossf/scorecard/blob/main/docs/checks.md#vulnerabilities","short":"Determines if the project has open, known unfixed vulnerabilities."}}],"metadata":null}
{"date":"2023-08-11T10:22:43-07:00","repo":{"name":"github.com/foo/bar","commit":"f0840f7158c8044af2bd9b8aa661d7942b1f29d2"},"scorecard":{"version":"","commit":"unknown"},"score":10.0,"checks":[{"details":null,"score":10,"reason":"no vulnerabilities detected","name":"Vulnerabilities","documentation":{"url":"https://github.com/ossf/scorecard/blob/main/docs/checks.md#vulnerabilities","short":"Determines if the project has open, known unfixed vulnerabilities."}}],"metadata":null}
`,
			y: `{"date":"2023-08-11T10:22:43-07:00","repo":{"name":"github.com/foo/bar","commit":"f0840f7158c8044af2bd9b8aa661d7942b1f29d2"},"scorecard":{"version":"","commit":"unknown"},"score":10.0,"checks":[{"details":null,"score":10,"reason":"no vulnerabilities detected","name":"Vulnerabilities","documentation":{"url":"https://github.com/ossf/scorecard/blob/main/docs/checks.md#vulnerabilities","short":"Determines if the project has open, known unfixed vulnerabilities."}}],"metadata":null}
`,
			wantErrSubstr: "number of results",
		},
		{
			name: "results differ",
			x: `{"date":"2023-08-11T10:22:43-07:00","repo":{"name":"github.com/foo/bar","commit":"f0840f7158c8044af2bd9b8aa661d7942b1f29d2"},"scorecard":{"version":"","commit":"unknown"},"score":10.0,"checks":[{"details":null,"score":10,"reason":"no vulnerabilities detected","name":"Vulnerabilities","documentation":{"url":"https://github.com/ossf/scorecard/blob/main/docs/checks.md#vulnerabilities","short":"Determines if the project has open, known unfixed vulnerabilities."}}],"metadata":null}
`,
			y: `{"date":"2023-08-11T10:22:43-07:00","repo":{"name":"github.com/foo/bar","commit":"f0840f7158c8044af2bd9b8aa661d7942b1f29d2"},"scorecard":{"version":"","commit":"unknown"},"score":7.0,"checks":[{"details":null,"score":7,"reason":"3 existing vulnerabilities detected","name":"Vulnerabilities","documentation":{"url":"https://github.com/ossf/scorecard/blob/main/docs/checks.md#vulnerabilities","short":"Determines if the project has open, known unfixed vulnerabilities."}}],"metadata":null}
`,
			wantErrSubstr: "results differ",
		},
		{
			name: "x results fail to parse",
			x: `not a scorecard result
`,
			y: `{"date":"2023-08-11T10:22:43-07:00","repo":{"name":"github.com/foo/bar","commit":"f0840f7158c8044af2bd9b8aa661d7942b1f29d2"},"scorecard":{"version":"","commit":"unknown"},"score":7.0,"checks":[{"details":null,"score":7,"reason":"3 existing vulnerabilities detected","name":"Vulnerabilities","documentation":{"url":"https://github.com/ossf/scorecard/blob/main/docs/checks.md#vulnerabilities","short":"Determines if the project has open, known unfixed vulnerabilities."}}],"metadata":null}
`,
			wantErrSubstr: "parsing first",
		},
		{
			name: "y results fail to parse",
			x: `{"date":"2023-08-11T10:22:43-07:00","repo":{"name":"github.com/foo/bar","commit":"f0840f7158c8044af2bd9b8aa661d7942b1f29d2"},"scorecard":{"version":"","commit":"unknown"},"score":7.0,"checks":[{"details":null,"score":7,"reason":"3 existing vulnerabilities detected","name":"Vulnerabilities","documentation":{"url":"https://github.com/ossf/scorecard/blob/main/docs/checks.md#vulnerabilities","short":"Determines if the project has open, known unfixed vulnerabilities."}}],"metadata":null}
`,
			y: `not a scorecard result
`,
			wantErrSubstr: "parsing second",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			x := strings.NewReader(tt.x)
			y := strings.NewReader(tt.y)
			err := compareReaders(x, y, os.Stderr)
			if (err != nil) == tt.match {
				t.Errorf("wanted match: %t, but got err: %v", tt.match, err)
			}
			if !tt.match && !strings.Contains(err.Error(), tt.wantErrSubstr) {
				t.Errorf("wanted err: %v, got err: %v", tt.wantErrSubstr, err)
			}
		})
	}
}

type alwaysErrorReader struct{}

func (a alwaysErrorReader) Read(b []byte) (n int, err error) {
	return 0, io.ErrClosedPipe
}

//nolint:lll // results are long
func Test_compare_reader_err(t *testing.T) {
	t.Parallel()
	//nolint:govet // struct alignment
	tests := []struct {
		name string
		x    io.Reader
		y    io.Reader
	}{
		{
			name: "error in y reader",
			x: strings.NewReader(`{"date":"2023-08-11T10:22:43-07:00","repo":{"name":"github.com/foo/bar","commit":"f0840f7158c8044af2bd9b8aa661d7942b1f29d2"},"scorecard":{"version":"","commit":"unknown"},"score":10.0,"checks":[{"details":null,"score":10,"reason":"no vulnerabilities detected","name":"Vulnerabilities","documentation":{"url":"https://github.com/ossf/scorecard/blob/main/docs/checks.md#vulnerabilities","short":"Determines if the project has open, known unfixed vulnerabilities."}}],"metadata":null}
`),
			y: alwaysErrorReader{},
		},
		{
			name: "error in x reader",
			x:    alwaysErrorReader{},
			y: strings.NewReader(`{"date":"2023-08-11T10:22:43-07:00","repo":{"name":"github.com/foo/bar","commit":"f0840f7158c8044af2bd9b8aa661d7942b1f29d2"},"scorecard":{"version":"","commit":"unknown"},"score":10.0,"checks":[{"details":null,"score":10,"reason":"no vulnerabilities detected","name":"Vulnerabilities","documentation":{"url":"https://github.com/ossf/scorecard/blob/main/docs/checks.md#vulnerabilities","short":"Determines if the project has open, known unfixed vulnerabilities."}}],"metadata":null}
`),
		},
		{
			name: "error in both readesr",
			x:    alwaysErrorReader{},
			y:    alwaysErrorReader{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := compareReaders(tt.x, tt.y, os.Stderr); err == nil { // if NO error
				t.Errorf("wanted error, got none")
			}
		})
	}
}
