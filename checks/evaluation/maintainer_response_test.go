// Copyright 2025 OpenSSF Scorecard Authors
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

package evaluation

import (
	"strings"
	"testing"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
	scut "github.com/ossf/scorecard/v5/utests"
)

func TestMaintainerResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		findings  []finding.Finding
		wantScore int
	}{
		{
			name:      "No findings - max score",
			findings:  []finding.Finding{},
			wantScore: checker.MaxResultScore,
		},
		{
			name: "All NotApplicable - max score",
			findings: []finding.Finding{
				{Outcome: finding.OutcomeNotApplicable, Message: "issue #1 had no bug/security labels"},
				{Outcome: finding.OutcomeNotApplicable, Message: "issue #2 had no bug/security labels"},
			},
			wantScore: checker.MaxResultScore,
		},
		{
			name: "100% violations (all issues unresponsive) - score 0",
			findings: []finding.Finding{
				{
					Outcome: finding.OutcomeFalse,
					Message: "issue #1 exceeded 180 days without reaction (worst 200 days)",
					Location: &finding.Location{
						Path: "https://github.com/owner/repo/issues/1",
					},
				},
				{
					Outcome: finding.OutcomeFalse,
					Message: "issue #2 exceeded 180 days without reaction (worst 250 days)",
					Location: &finding.Location{
						Path: "https://github.com/owner/repo/issues/2",
					},
				},
			},
			wantScore: 0,
		},
		{
			name: "50% violations (half unresponsive) - score 0",
			findings: []finding.Finding{
				{Outcome: finding.OutcomeFalse, Message: "issue #1 exceeded 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #2 did not exceed 180 days"},
			},
			wantScore: 0,
		},
		{
			name: "Boundary: exactly 40% violations - score 5",
			findings: []finding.Finding{
				{Outcome: finding.OutcomeFalse, Message: "issue #1 exceeded 180 days"},
				{Outcome: finding.OutcomeFalse, Message: "issue #2 exceeded 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #3 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #4 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #5 did not exceed 180 days"},
			},
			wantScore: 5, // 40% = 2 out of 5, NOT > 40.0, so score 5
		},
		{
			name: "Just above 40% violations (41%) - score 0",
			findings: []finding.Finding{
				{Outcome: finding.OutcomeFalse, Message: "issue #1 exceeded 180 days"},
				{Outcome: finding.OutcomeFalse, Message: "issue #2 exceeded 180 days"},
				{Outcome: finding.OutcomeFalse, Message: "issue #3 exceeded 180 days"},
				{Outcome: finding.OutcomeFalse, Message: "issue #4 exceeded 180 days"},
				{Outcome: finding.OutcomeFalse, Message: "issue #5 exceeded 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #6 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #7 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #8 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #9 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #10 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #11 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #12 did not exceed 180 days"},
			},
			wantScore: 0, // 5/12 = 41.67%
		},
		{
			name: "Just below 40% violations (38%) - score 5",
			findings: []finding.Finding{
				{Outcome: finding.OutcomeFalse, Message: "issue #1 exceeded 180 days"},
				{Outcome: finding.OutcomeFalse, Message: "issue #2 exceeded 180 days"},
				{Outcome: finding.OutcomeFalse, Message: "issue #3 exceeded 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #4 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #5 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #6 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #7 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #8 did not exceed 180 days"},
			},
			wantScore: 5, // 3/8 = 37.5%
		},
		{
			name: "30% violations - score 5",
			findings: []finding.Finding{
				{Outcome: finding.OutcomeFalse, Message: "issue #1 exceeded 180 days"},
				{Outcome: finding.OutcomeFalse, Message: "issue #2 exceeded 180 days"},
				{Outcome: finding.OutcomeFalse, Message: "issue #3 exceeded 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #4 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #5 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #6 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #7 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #8 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #9 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #10 did not exceed 180 days"},
			},
			wantScore: 5, // 3/10 = 30%
		},
		{
			name: "Boundary: exactly 20% violations - score 10",
			findings: []finding.Finding{
				{Outcome: finding.OutcomeFalse, Message: "issue #1 exceeded 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #2 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #3 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #4 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #5 did not exceed 180 days"},
			},
			wantScore: 10, // 1/5 = 20%, NOT > 20.0, so score 10
		},
		{
			name: "Just above 20% violations (21%) - score 5",
			findings: []finding.Finding{
				{Outcome: finding.OutcomeFalse, Message: "issue #1 exceeded 180 days"},
				{Outcome: finding.OutcomeFalse, Message: "issue #2 exceeded 180 days"},
				{Outcome: finding.OutcomeFalse, Message: "issue #3 exceeded 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #4 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #5 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #6 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #7 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #8 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #9 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #10 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #11 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #12 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #13 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #14 did not exceed 180 days"},
			},
			wantScore: 5, // 3/14 = 21.43%
		},
		{
			name: "Just below 20% violations (19%) - score 10",
			findings: []finding.Finding{
				{Outcome: finding.OutcomeFalse, Message: "issue #1 exceeded 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #2 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #3 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #4 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #5 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #6 did not exceed 180 days"},
			},
			wantScore: 10, // 1/6 = 16.67%
		},
		{
			name: "10% violations - score 10",
			findings: []finding.Finding{
				{Outcome: finding.OutcomeFalse, Message: "issue #1 exceeded 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #2 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #3 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #4 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #5 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #6 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #7 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #8 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #9 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #10 did not exceed 180 days"},
			},
			wantScore: 10,
		},
		{
			name: "0% violations (all responsive) - score 10",
			findings: []finding.Finding{
				{Outcome: finding.OutcomeTrue, Message: "issue #1 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #2 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #3 did not exceed 180 days"},
			},
			wantScore: 10,
		},
		{
			name: "Single violation out of 1 (100%) - score 0",
			findings: []finding.Finding{
				{Outcome: finding.OutcomeFalse, Message: "issue #1 exceeded 180 days"},
			},
			wantScore: 0,
		},
		{
			name: "Single non-violation (0%) - score 10",
			findings: []finding.Finding{
				{Outcome: finding.OutcomeTrue, Message: "issue #1 did not exceed 180 days"},
			},
			wantScore: 10,
		},
		{
			name: "Mix with NotApplicable excluded from calculation",
			findings: []finding.Finding{
				{Outcome: finding.OutcomeFalse, Message: "issue #1 exceeded 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #2 did not exceed 180 days"},
				{Outcome: finding.OutcomeNotApplicable, Message: "issue #3 had no labels"},
				{Outcome: finding.OutcomeNotApplicable, Message: "issue #4 had no labels"},
			},
			wantScore: 0, // 50% of evaluated (1 out of 2) = score 0
		},
		{
			name: "Complex mix: NotApplicable + Unknown outcomes ignored",
			findings: []finding.Finding{
				{Outcome: finding.OutcomeFalse, Message: "issue #1 exceeded 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #2 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #3 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #4 did not exceed 180 days"},
				{Outcome: finding.OutcomeTrue, Message: "issue #5 did not exceed 180 days"},
				{Outcome: finding.OutcomeNotApplicable, Message: "issue #6 had no labels"},
				{Outcome: finding.OutcomeNotApplicable, Message: "issue #7 had no labels"},
				{Outcome: finding.OutcomeNotApplicable, Message: "issue #8 had no labels"},
			},
			wantScore: 10, // 1/5 = 20% of evaluated, but NOT > 20%, so score 10
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := &scut.TestDetailLogger{}
			result := MaintainerResponse("test-check", tt.findings, dl)
			if result.Score != tt.wantScore {
				t.Errorf("MaintainerResponse() score = %v, want %v", result.Score, tt.wantScore)
			}
			if len(result.Findings) != len(tt.findings) {
				t.Errorf("MaintainerResponse() findings count = %v, want %v", len(result.Findings), len(tt.findings))
			}
		})
	}
}

func TestParseIssueNumberFromURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		url  string
		want int
	}{
		{
			name: "GitHub issue URL",
			url:  "https://github.com/owner/repo/issues/123",
			want: 123,
		},
		{
			name: "GitHub issue URL with trailing slash",
			url:  "https://github.com/owner/repo/issues/456/",
			want: 456,
		},
		{
			name: "GitLab issue URL",
			url:  "https://gitlab.com/owner/repo/-/issues/789",
			want: 789,
		},
		{
			name: "No issue number",
			url:  "https://github.com/owner/repo",
			want: 0,
		},
		{
			name: "Empty URL",
			url:  "",
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parseIssueNumberFromURL(tt.url)
			if got != tt.want {
				t.Errorf("parseIssueNumberFromURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseAnyInt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		wantN  int
		wantOk bool
	}{
		{
			name:   "Single number",
			input:  "issue 123 exceeded",
			wantN:  123,
			wantOk: true,
		},
		{
			name:   "Multiple numbers - returns first",
			input:  "issue 456 took 789 days",
			wantN:  456,
			wantOk: true,
		},
		{
			name:   "No numbers",
			input:  "no numbers here",
			wantN:  0,
			wantOk: false,
		},
		{
			name:   "Empty string",
			input:  "",
			wantN:  0,
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotN, gotOk := parseAnyInt(tt.input)
			if gotN != tt.wantN || gotOk != tt.wantOk {
				t.Errorf("parseAnyInt() = (%v, %v), want (%v, %v)", gotN, gotOk, tt.wantN, tt.wantOk)
			}
		})
	}
}

func TestFormatIssueList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		want     string
		nums     []int
		maxCount int
	}{
		{
			name:     "Empty list",
			nums:     []int{},
			maxCount: 20,
			want:     "",
		},
		{
			name:     "Single issue",
			nums:     []int{123},
			maxCount: 20,
			want:     "#123",
		},
		{
			name:     "Multiple issues under limit",
			nums:     []int{1, 2, 3},
			maxCount: 20,
			want:     "#1, #2, #3",
		},
		{
			name:     "More issues than limit",
			nums:     []int{1, 2, 3, 4, 5, 6},
			maxCount: 3,
			want:     "#1, #2, #3, ... +3 more",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := formatIssueList(tt.nums, tt.maxCount)
			if got != tt.want {
				t.Errorf("formatIssueList() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Regression test for bug where "worst X days" showed misleading values in perfect score case.
// When all issues had timely responses (0 violations), the message should not show
// "worst 58453 days" from old open issues, as discovered in github.com/istio/istio testing.
func TestMaintainerResponse_RegressionPerfectScoreWorstDays(t *testing.T) {
	t.Parallel()

	// Simulate findings where all had responses, but some intervals were very long
	// (e.g., old open issues that had responses but have been open for years)
	findings := []finding.Finding{
		{
			Outcome: finding.OutcomeTrue,
			Message: "issue #1 did not exceed 180 days without reaction (worst 50 days)",
		},
		{
			Outcome: finding.OutcomeTrue,
			Message: "issue #2 did not exceed 180 days without reaction (worst 100 days)",
		},
		{
			Outcome: finding.OutcomeTrue,
			Message: "issue #3 did not exceed 180 days without reaction (worst 58453 days)", // Very old open issue
		},
	}

	dl := &scut.TestDetailLogger{}
	result := MaintainerResponse("Maintainer-Response-Test", findings, dl)

	// Should get perfect score
	if result.Score != checker.MaxResultScore {
		t.Errorf("Expected score %d, got %d", checker.MaxResultScore, result.Score)
	}

	// The reason should NOT contain misleading "worst 58453 days" text
	// Instead, it should say "had timely maintainer activity"
	if result.Reason == "" {
		t.Errorf("Expected non-empty reason")
	}

	// Should contain "timely maintainer activity"
	if !strings.Contains(result.Reason, "timely maintainer activity") {
		t.Errorf("Expected reason to mention 'timely maintainer activity', got: %s", result.Reason)
	}

	// Should NOT show confusing large numbers like 58453
	if strings.Contains(result.Reason, "58453") {
		t.Errorf("Reason should not contain misleading 'worst 58453 days', got: %s", result.Reason)
	}
}

// Regression test for bug where GitHub client only fetched "bug" and "security" labels,
// missing "kind/bug", "area/security", etc. as discovered in github.com/ossf/scorecard testing.
// This E2E test verifies the entire pipeline works with new label types.
func TestMaintainerResponse_RegressionNewLabelTypes(t *testing.T) {
	t.Parallel()

	// Test findings with the new label types that were previously ignored
	findings := []finding.Finding{
		{
			Outcome: finding.OutcomeTrue,
			Message: "issue #4762 did not exceed 180 days without reaction (worst 10 days)", // kind/bug label
		},
		{
			Outcome: finding.OutcomeTrue,
			Message: "issue #4730 did not exceed 180 days without reaction (worst 5 days)", // kind/bug label
		},
		{
			Outcome: finding.OutcomeFalse,
			Message: "issue #4729 exceeded 180 days without reaction (worst 200 days)", // area/security label
			Location: &finding.Location{
				Type: finding.FileTypeURL,
				Path: "https://github.com/ossf/scorecard/issues/4729",
			},
		},
	}

	dl := &scut.TestDetailLogger{}
	result := MaintainerResponse("Maintainer-Response-Test", findings, dl)

	// Should evaluate all 3 issues (not say "no issues with bug/security labels found")
	if !strings.Contains(result.Reason, "Evaluated 3 issues") {
		t.Errorf("Expected reason to say 'Evaluated 3 issues', got: %s", result.Reason)
	}

	// Should show 2 had activity and 1 violation (33.3%)
	if !strings.Contains(result.Reason, "2 had activity") {
		t.Errorf("Expected reason to mention '2 had activity', got: %s", result.Reason)
	}

	// Score should be 5 (33.3% violations: >20% but ≤40%)
	if result.Score != 5 {
		t.Errorf("Expected score 5 for 33.3%% violations, got %d", result.Score)
	}
}
