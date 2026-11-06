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

package scorecard

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/config"
	"github.com/ossf/scorecard/v5/docs/checks"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/log"
	"github.com/ossf/scorecard/v5/options"
	spol "github.com/ossf/scorecard/v5/policy"
)

func mockScorecardResultCheck1(t *testing.T) *Result {
	t.Helper()
	// Helper variables to mock Scorecard results
	date, err := time.Parse(time.RFC3339, "2023-03-02T10:30:43-06:00")
	if err != nil {
		t.Fatalf("time.Parse: %v", err)
	}
	t.Logf("date: %v", date)

	return &Result{
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
		Config: config.Config{
			Annotations: []config.Annotation{
				{
					Checks: []string{"Check-Name"},
					Reasons: []config.ReasonGroup{
						{Reason: "test-data"},
						{Reason: "not-applicable"},
					},
				},
			},
		},
		Metadata: []string{},
	}
}

func Test_formatResults_outputToFile(t *testing.T) {
	t.Parallel()
	type args struct {
		opts    *options.Options
		results *Result
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
		{
			name: "output file with format default and annotations",
			args: args{
				opts: &options.Options{
					Format:          options.FormatDefault,
					ShowDetails:     true,
					LogLevel:        log.DebugLevel.String(),
					ShowAnnotations: true,
				},
				results: scorecardResults,
				doc:     checkDocs,
			},
			want: want{
				path: "check1_annotations.log",
				err:  false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// generate a unique result file in a temp directory for every test run to avoid race conditions in the test.
			// This can happen when ginkgo runs unit tests in parallel (with -p flag)
			tt.args.opts.ResultsFile = filepath.Join(t.TempDir(), "result-file")

			// Format results.
			formatErr := FormatResults(tt.args.opts, tt.args.results, tt.args.doc, tt.args.policy)
			if (formatErr != nil) != tt.want.err {
				t.Fatalf("FormatResults() error = %v, want error %v", formatErr, tt.want.err)
			}

			// Get output and wanted output.
			output, outputErr := os.ReadFile(tt.args.opts.ResultsFile)
			if outputErr != nil {
				t.Fatalf("cannot read file: %v", outputErr)
			}
			wantOutput, wantOutputErr := os.ReadFile("./testdata/" + tt.want.path)
			if wantOutputErr != nil {
				t.Fatalf("cannot read file: %v", wantOutputErr)
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
		})
	}
}

func testDocForFixSummary() *mockDoc {
	return &mockDoc{checks: map[string]mockCheck{
		"Check-Critical": {name: "Check-Critical", risk: "Critical", url: "https://example.com/critical"},
		"Check-High":     {name: "Check-High", risk: "High", url: "https://example.com/high"},
		"Check-Medium":   {name: "Check-Medium", risk: "Medium", url: "https://example.com/medium"},
		"Check-Low":      {name: "Check-Low", risk: "Low", url: "https://example.com/low"},
	}}
}

func TestStringResult_fixFirstSummary(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		checks []checker.CheckResult
		// want is the expected summary line; "" means the line must be absent.
		want string
	}{
		{
			name: "ranked by risk, then score, top 3",
			checks: []checker.CheckResult{
				{Name: "Check-Low", Score: 0, Reason: "r"},
				{Name: "Check-Medium", Score: 4, Reason: "r"},
				{Name: "Check-High", Score: 1, Reason: "r"},
				{Name: "Check-Critical", Score: 4, Reason: "r"},
			},
			want: "Highest-impact fixes: Check-Critical, Check-High, Check-Medium",
		},
		{
			name: "no failing checks means no summary line",
			checks: []checker.CheckResult{
				{Name: "Check-High", Score: 10, Reason: "r"},
				{Name: "Check-Low", Score: 7, Reason: "r"},
			},
			want: "",
		},
		{
			name: "inconclusive checks are not listed",
			checks: []checker.CheckResult{
				{Name: "Check-High", Score: checker.InconclusiveResultScore, Reason: "r"},
				{Name: "Check-Low", Score: 10, Reason: "r"},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := mockScorecardResultCheck1(t)
			r.Checks = tt.checks
			var buf bytes.Buffer
			if err := r.AsString(&buf, testDocForFixSummary(), nil); err != nil {
				t.Fatalf("AsString: %v", err)
			}
			out := buf.String()
			if tt.want == "" {
				if strings.Contains(out, "Highest-impact fixes:") {
					t.Errorf("unexpected summary line in output:\n%s", out)
				}
				return
			}
			if !strings.Contains(out, tt.want) {
				t.Errorf("output missing %q:\n%s", tt.want, out)
			}
		})
	}
}
