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

package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/conformance"
	"github.com/ossf/scorecard/v5/internal/conformance/catalog"
	"github.com/ossf/scorecard/v5/options"
	"github.com/ossf/scorecard/v5/pkg/scorecard"
)

func scanResult(defaultBranch string, findings ...finding.Finding) *scorecard.Result {
	r := &scorecard.Result{
		Repo:      scorecard.RepoInfo{Name: "github.com/owner/repo", CommitSHA: "0123456789abcdef"},
		Scorecard: scorecard.ScorecardInfo{Version: "v5.7.0", CommitSHA: "fedcba9876543210"},
		Date:      time.Date(2026, time.November, 8, 12, 0, 0, 0, time.UTC),
		Findings:  findings,
	}
	r.RawResults.Metadata.Metadata = map[string]string{"repository.defaultBranch": defaultBranch}
	return r
}

func onBranch(probe, branch string, outcome finding.Outcome) finding.Finding {
	return finding.Finding{Probe: probe, Outcome: outcome, Values: map[string]string{"branchName": branch}}
}

func TestConformanceReport(t *testing.T) {
	t.Parallel()
	release, err := catalog.Lookup("osps-baseline")
	if err != nil {
		t.Fatal(err)
	}
	result := scanResult("main",
		onBranch("requiresPRsToChangeCode", "main", finding.OutcomeTrue),
		onBranch("branchProtectionAppliesToAdmins", "main", finding.OutcomeFalse),
		onBranch("blocksDeleteOnBranches", "main", finding.OutcomeTrue),
		onBranch("blocksDeleteOnBranches", "release/v1", finding.OutcomeFalse),
	)
	report := conformanceReport(result, release, "maturity-1")

	want := map[string]conformance.Result{
		"OSPS-AC-03.01": conformance.NeedsReview,
		"OSPS-AC-03.02": conformance.Pass,
	}
	if len(report.Results) != len(want) {
		t.Fatalf("got %d results, want %d: %v", len(report.Results), len(want), report.Results)
	}
	for _, r := range report.Results {
		if r.Result != want[r.Requirement] {
			t.Errorf("%s = %s, want %s", r.Requirement, r.Result, want[r.Requirement])
		}
	}
	if got := len(report.NotEvaluated); got != 23 {
		t.Errorf("got %d requirements not evaluated, want 23", got)
	}
	if report.Repo != "github.com/owner/repo" || report.Commit != "0123456789abcdef" ||
		report.ScorecardVersion != "v5.7.0" || report.Profile != "maturity-1" {
		t.Errorf("report identity = %+v", report)
	}
}

// TestConformanceReportEmptyRepository covers a scan that returns before
// recording a default branch, as Scorecard does for an empty repository.
func TestConformanceReportEmptyRepository(t *testing.T) {
	t.Parallel()
	release, err := catalog.Lookup("osps-baseline")
	if err != nil {
		t.Fatal(err)
	}
	result := &scorecard.Result{Repo: scorecard.RepoInfo{Name: "github.com/owner/empty"}}
	report := conformanceReport(result, release, "maturity-1")
	for _, r := range report.Results {
		if r.Result != conformance.Unknown {
			t.Errorf("%s = %s, want UNKNOWN", r.Requirement, r.Result)
		}
	}
}

func TestWriteConformanceJSON(t *testing.T) {
	t.Parallel()
	release, err := catalog.Lookup("osps-baseline")
	if err != nil {
		t.Fatal(err)
	}
	report := conformanceReport(scanResult("main"), release, "maturity-1")
	var b bytes.Buffer
	if err := writeConformance(&b, options.FormatJSON, report); err != nil {
		t.Fatalf("writeConformance() = %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(b.Bytes(), &out); err != nil {
		t.Fatalf("output is not JSON: %v", err)
	}
	for _, key := range []string{"experimental", "date", "repo", "scorecard", "catalog", "profile", "results", "notEvaluated"} {
		if _, ok := out[key]; !ok {
			t.Errorf("output lacks %q", key)
		}
	}
	for _, key := range []string{"checks", "score", "aggregate_score"} {
		if _, ok := out[key]; ok {
			t.Errorf("output has check score key %q", key)
		}
	}
	if out["experimental"] != true {
		t.Errorf("experimental = %v, want true", out["experimental"])
	}
}

func TestWriteConformanceGemara(t *testing.T) {
	t.Parallel()
	release, err := catalog.Lookup("osps-baseline@v2026.08.28")
	if err != nil {
		t.Fatal(err)
	}
	report := conformanceReport(scanResult("main"), release, "maturity-1")
	var b bytes.Buffer
	if err := writeConformance(&b, options.FormatGemara, report); err != nil {
		t.Fatalf("writeConformance() = %v", err)
	}
	var out struct {
		Metadata struct {
			MappingReferences []struct {
				Version string `json:"version"`
			} `json:"mapping-references"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal(b.Bytes(), &out); err != nil {
		t.Fatalf("output is not JSON: %v", err)
	}
	refs := out.Metadata.MappingReferences
	if len(refs) != 1 || refs[0].Version != "v2026.08.28" {
		t.Errorf("mapping-references = %+v, want one entry with version v2026.08.28", refs)
	}
}

func TestWriteConformanceRejectsOtherFormats(t *testing.T) {
	t.Parallel()
	release, err := catalog.Lookup("osps-baseline")
	if err != nil {
		t.Fatal(err)
	}
	report := conformanceReport(scanResult("main"), release, "maturity-1")
	var b bytes.Buffer
	if err := writeConformance(&b, options.FormatDefault, report); !errors.Is(err, errCatalogFormat) {
		t.Errorf("writeConformance(default) = %v, want %v", err, errCatalogFormat)
	}
}
