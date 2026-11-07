// Copyright 2026 OpenSSF Scorecard Authors
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

package checkdef

import (
	"errors"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v5/checker"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes"
)

var (
	errCollect = errors.New("collect failed")
	errProbe   = errors.New("probe failed")
)

// testDefinition returns a definition named Test whose collector returns
// one CI-Tests revision or collectErr, and whose probe returns one
// finding or probeErr. collected counts collector calls.
func testDefinition(collectErr, probeErr error, collected *int) *Definition {
	collect := func(*checker.CheckRequest) (checker.CITestData, error) {
		*collected++
		if collectErr != nil {
			return checker.CITestData{}, collectErr
		}
		return checker.CITestData{CIInfo: []checker.RevisionCIInfo{{PullRequestNumber: 1}}}, nil
	}
	probe := func(r *checker.RawResults) ([]finding.Finding, string, error) {
		if probeErr != nil {
			return nil, "testProbe", probeErr
		}
		return []finding.Finding{{Probe: "testProbe", Outcome: finding.OutcomeTrue}}, "testProbe", nil
	}
	evaluate := func(name string, findings []finding.Finding, dl checker.DetailLogger) checker.CheckResult {
		dl.Info(&checker.LogMessage{Text: "evaluated"})
		return checker.CreateMaxScoreResult(name, "ok")
	}
	return define(
		"Test",
		collect,
		func(r *checker.RawResults, data checker.CITestData) { r.CITestResults = data },
		[]probes.ProbeImpl{probe},
		evaluate,
	)
}

func TestRun(t *testing.T) {
	t.Parallel()
	collected := 0
	raw := &checker.RawResults{}
	dl := checker.NewLogger()
	got := testDefinition(nil, nil, &collected).Run(&checker.CheckRequest{RawResults: raw, Dlogger: dl})

	if collected != 1 {
		t.Errorf("collector ran %d times, want 1", collected)
	}
	if n := len(raw.CITestResults.CIInfo); n != 1 {
		t.Errorf("shared RawResults has %d CI-Tests revisions, want 1", n)
	}
	want := checker.CreateMaxScoreResult("Test", "ok")
	want.Findings = []finding.Finding{{Probe: "testProbe", Outcome: finding.OutcomeTrue}}
	if diff := cmp.Diff(want, got, cmpopts.IgnoreUnexported(finding.Finding{})); diff != "" {
		t.Errorf("result (-want +got):\n%s", diff)
	}
	if logs := dl.Flush(); len(logs) != 1 || logs[0].Msg.Text != "evaluated" {
		t.Errorf("details = %+v, want the evaluation's one detail", logs)
	}
}

func TestRunWithoutRawResults(t *testing.T) {
	t.Parallel()
	collected := 0
	got := testDefinition(nil, nil, &collected).Run(&checker.CheckRequest{Dlogger: checker.NewLogger()})
	if got.Error != nil || got.Score != checker.MaxResultScore {
		t.Errorf("Run = score %d, error %v; want %d, no error", got.Score, got.Error, checker.MaxResultScore)
	}
}

func TestRunErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		collectErr error
		probeErr   error
		name       string
		wantText   string
	}{
		{
			name:       "collector fails",
			collectErr: errCollect,
			wantText:   "internal error: collect failed",
		},
		{
			name:     "probe fails",
			probeErr: errProbe,
			wantText: "internal error: probe run failure: [probe failed]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			collected := 0
			d := testDefinition(tt.collectErr, tt.probeErr, &collected)
			got := d.Run(&checker.CheckRequest{Dlogger: checker.NewLogger()})
			if !errors.Is(got.Error, sce.ErrScorecardInternal) || got.Error.Error() != tt.wantText {
				t.Errorf("error = %v, want %q wrapping ErrScorecardInternal", got.Error, tt.wantText)
			}
			if got.Name != "Test" || got.Score != checker.InconclusiveResultScore || got.Findings != nil {
				t.Errorf("Run = %+v, want an inconclusive runtime error result for Test", got)
			}
		})
	}
}

func TestRunExperimental(t *testing.T) {
	const wantText = "SCORECARD_EXPERIMENTAL is not set, not running the Testing check"
	tests := []struct {
		name string
		set  bool
	}{
		{name: "unset"},
		{name: "set", set: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SCORECARD_EXPERIMENTAL", "1")
			if !tt.set {
				if err := os.Unsetenv("SCORECARD_EXPERIMENTAL"); err != nil {
					t.Fatalf("Unsetenv: %v", err)
				}
			}
			collected := 0
			dl := checker.NewLogger()
			got := testDefinition(nil, nil, &collected).experimental("Testing").Run(
				&checker.CheckRequest{Dlogger: dl})
			if tt.set {
				if got.Error != nil || collected != 1 {
					t.Errorf("Run = error %v after %d collections, want a run", got.Error, collected)
				}
				return
			}
			if collected != 0 {
				t.Errorf("collector ran %d times, want 0", collected)
			}
			if !errors.Is(got.Error, sce.ErrUnsupportedCheck) || got.Error.Error() != "check is not supported for this request: "+wantText {
				t.Errorf("error = %v, want %q wrapping ErrUnsupportedCheck", got.Error, wantText)
			}
			logs := dl.Flush()
			if len(logs) != 1 || logs[0].Type != checker.DetailWarn || logs[0].Msg.Text != wantText {
				t.Errorf("details = %+v, want one warning %q", logs, wantText)
			}
		})
	}
}
