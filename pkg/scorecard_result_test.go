// Copyright 2020 OpenSSF Scorecard Authors
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
package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/docs/checks"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/log"
	"github.com/ossf/scorecard/v4/options"
	spol "github.com/ossf/scorecard/v4/policy"
)

func mockScorecardResultCheck1(t *testing.T) *ScorecardResult {
	t.Helper()
	// Helper variables to mock Scorecard results
	date, e := time.Parse(time.RFC3339, "2023-03-02T10:30:43-06:00")
	t.Logf("date: %v", date)
	if e != nil {
		panic(fmt.Errorf("time.Parse: %w", e))
	}

	return &ScorecardResult{
		Repo: RepoInfo{
			Name:      "org/name",
			CommitSHA: "68bc59901773ab4c051dfcea0cc4201a1567ab32",
		},
		Scorecard: ScorecardInfo{
			Version:   "1.2.3",
			CommitSHA: "ccbc59901773ab4c051dfcea0cc4201a1567abdd",
		},
		Date: date,
		Checks: []checker.CheckResult{
			{
				Details: []checker.CheckDetail{
					{
						Type: checker.DetailWarn,
						Msg: checker.LogMessage{
							Text:    "warn message",
							Path:    "src/file1.cpp",
							Type:    finding.FileTypeSource,
							Offset:  5,
							Snippet: "if (bad) {BUG();}",
						},
					},
				},
				Score:  5,
				Reason: "half score reason",
				Name:   "Check-Name",
			},
		},
		Metadata: []string{},
	}
}

func Test_formatResults_outputToFile(t *testing.T) {
	t.Parallel()
	type args struct {
		opts    *options.Options
		results *ScorecardResult
		doc     checks.Doc
		policy  *spol.ScorecardPolicy
	}
	type want struct {
		path string
		err  bool
	}

	// Helper variables to mock scorecard results and checks doc
	scorecardResults := mockScorecardResultCheck1(t)
	checkDocs := jsonMockDocRead()

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "output file with format json",
			args: args{
				opts: &options.Options{
					Format:      options.FormatJSON,
					Output:      "result.json",
					ShowDetails: true,
					LogLevel:    log.DebugLevel.String(),
				},
				results: scorecardResults,
				doc:     checkDocs,
			},
			want: want{
				path: "check1.json",
				err:  false,
			},
		},
		{
			name: "output file with format default",
			args: args{
				opts: &options.Options{
					Format:      options.FormatDefault,
					Output:      "result.log",
					ShowDetails: true,
					LogLevel:    log.DebugLevel.String(),
				},
				results: scorecardResults,
				doc:     checkDocs,
			},
			want: want{
				path: "check1.log",
				err:  false,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Format results.
			formatErr := FormatResults(tt.args.opts, tt.args.results, tt.args.doc, tt.args.policy)
			if (formatErr != nil) != tt.want.err {
				t.Errorf("FormatResults() error = %v, want error %v", formatErr, tt.want.err)
				return
			}

			// Get output and wanted output.
			output, outputErr := os.ReadFile(tt.args.opts.Output)
			if outputErr != nil {
				t.Errorf("cannot read file: %v", outputErr)
			}
			wantOutput, wantOutputErr := os.ReadFile("./testdata/" + tt.want.path)
			if wantOutputErr != nil {
				t.Errorf("cannot read file: %v", wantOutputErr)
			}

			// Unmarshal if comparing JSON output.
			if tt.args.opts.Format == options.FormatJSON {
				// Unmarshal expected output.
				var js JSONScorecardResultV2
				if err := json.Unmarshal(wantOutput, &js); err != nil {
					t.Fatalf("%s: json.Unmarshal: %s", tt.name, err)
				}

				// Marshal.
				var es bytes.Buffer
				encoder := json.NewEncoder(&es)
				if err := encoder.Encode(js); err != nil {
					t.Fatalf("%s: Encode: %s", tt.name, err)
				}
				wantOutput = es.Bytes()
			}

			// Compare outputs.
			if !bytes.Equal(output, wantOutput) {
				t.Errorf("%v\nGOT\n-------\n%s\nWANT\n-------\n%s", tt.name, string(output), string(wantOutput))
			}

			// Delete generated output file.
			os.Remove(tt.args.opts.Output)
		})
	}
}
