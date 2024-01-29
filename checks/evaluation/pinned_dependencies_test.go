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

package evaluation

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	scut "github.com/ossf/scorecard/v4/utests"
)

var testLineEnd = uint(124)

func Test_createScoreForGitHubActionsWorkflow(t *testing.T) {
	t.Parallel()
	//nolint:govet
	tests := []struct {
		name   string
		r      workflowPinningResult
		scores []checker.ProportionalScoreWeighted
	}{
		{
			name: "GitHub-owned and Third-Party actions pinned",
			r: workflowPinningResult{
				gitHubOwned: pinnedResult{
					pinned: 1,
					total:  1,
				},
				thirdParties: pinnedResult{
					pinned: 1,
					total:  1,
				},
			},
			scores: []checker.ProportionalScoreWeighted{
				{
					Success: 1,
					Total:   1,
					Weight:  2,
				},
				{
					Success: 1,
					Total:   1,
					Weight:  8,
				},
			},
		},
		{
			name: "only GitHub-owned actions pinned",
			r: workflowPinningResult{
				gitHubOwned: pinnedResult{
					pinned: 1,
					total:  1,
				},
				thirdParties: pinnedResult{
					pinned: 0,
					total:  1,
				},
			},
			scores: []checker.ProportionalScoreWeighted{
				{
					Success: 1,
					Total:   1,
					Weight:  2,
				},
				{
					Success: 0,
					Total:   1,
					Weight:  8,
				},
			},
		},
		{
			name: "only Third-Party actions pinned",
			r: workflowPinningResult{
				gitHubOwned: pinnedResult{
					pinned: 0,
					total:  1,
				},
				thirdParties: pinnedResult{
					pinned: 1,
					total:  1,
				},
			},
			scores: []checker.ProportionalScoreWeighted{
				{
					Success: 0,
					Total:   1,
					Weight:  2,
				},
				{
					Success: 1,
					Total:   1,
					Weight:  8,
				},
			},
		},
		{
			name: "no GitHub actions pinned",
			r: workflowPinningResult{
				gitHubOwned: pinnedResult{
					pinned: 0,
					total:  1,
				},
				thirdParties: pinnedResult{
					pinned: 0,
					total:  1,
				},
			},
			scores: []checker.ProportionalScoreWeighted{
				{
					Success: 0,
					Total:   1,
					Weight:  2,
				},
				{
					Success: 0,
					Total:   1,
					Weight:  8,
				},
			},
		},
		{
			name: "no GitHub-owned actions and Third-party actions unpinned",
			r: workflowPinningResult{
				gitHubOwned: pinnedResult{
					pinned: 0,
					total:  0,
				},
				thirdParties: pinnedResult{
					pinned: 0,
					total:  1,
				},
			},
			scores: []checker.ProportionalScoreWeighted{
				{
					Success: 0,
					Total:   1,
					Weight:  10,
				},
			},
		},
		{
			name: "no Third-party actions and GitHub-owned actions unpinned",
			r: workflowPinningResult{
				gitHubOwned: pinnedResult{
					pinned: 0,
					total:  1,
				},
				thirdParties: pinnedResult{
					pinned: 0,
					total:  0,
				},
			},
			scores: []checker.ProportionalScoreWeighted{
				{
					Success: 0,
					Total:   1,
					Weight:  10,
				},
			},
		},
		{
			name: "no GitHub-owned actions and Third-party actions pinned",
			r: workflowPinningResult{
				gitHubOwned: pinnedResult{
					pinned: 0,
					total:  0,
				},
				thirdParties: pinnedResult{
					pinned: 1,
					total:  1,
				},
			},
			scores: []checker.ProportionalScoreWeighted{
				{
					Success: 1,
					Total:   1,
					Weight:  10,
				},
			},
		},
		{
			name: "no Third-party actions and GitHub-owned actions pinned",
			r: workflowPinningResult{
				gitHubOwned: pinnedResult{
					pinned: 1,
					total:  1,
				},
				thirdParties: pinnedResult{
					pinned: 0,
					total:  0,
				},
			},
			scores: []checker.ProportionalScoreWeighted{
				{
					Success: 1,
					Total:   1,
					Weight:  10,
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			actual := createScoreForGitHubActionsWorkflow(&tt.r, &dl)
			diff := cmp.Diff(tt.scores, actual)
			if diff != "" {
				t.Errorf("createScoreForGitHubActionsWorkflow (-want,+got) %+v", diff)
			}
		})
	}
}

func Test_PinningDependencies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		findings []finding.Finding
		result   scut.TestReturn
	}{
		{
			name: "pinned pip dependency scores 10 and shows no warn message",
			findings: []finding.Finding{
				{
					Probe:   "pinsDependencies",
					Outcome: finding.OutcomePositive,
					Location: &finding.Location{
						Type:      finding.FileTypeText,
						Path:      "test-file",
						LineStart: &testLineStart,
						Snippet:   &testSnippet,
					},
					Values: map[string]int{
						"dependencyType": 6, // pip type
					},
				},
			},
			result: scut.TestReturn{
				Score:        10,
				NumberOfInfo: 1,
			},
		},
		{
			name: "unpinned pip dependency scores 0 and shows warn message",
			findings: []finding.Finding{
				{
					Probe:   "pinsDependencies",
					Outcome: finding.OutcomeNegative,
					Location: &finding.Location{
						Type:      finding.FileTypeText,
						Path:      "test-file",
						LineStart: &testLineStart,
						LineEnd:   &testLineEnd,
						Snippet:   &testSnippet,
					},
					Values: map[string]int{
						"dependencyType": 6, // pip type
					},
				},
			},
			result: scut.TestReturn{
				Score:        0,
				NumberOfInfo: 1,
				NumberOfWarn: 1,
			},
		},
		{
			name: "dependency missing Pinned info does not count for score and shows debug message",
			findings: []finding.Finding{
				{
					Probe:   "pinsDependencies",
					Outcome: finding.OutcomeNotApplicable,
					Location: &finding.Location{
						Type:      finding.FileTypeText,
						Path:      "test-file",
						LineStart: &testLineStart,
						LineEnd:   &testLineEnd,
						Snippet:   &testSnippet,
					},
					Values: map[string]int{
						"dependencyType": 6, // pip type
					},
				},
			},
			result: scut.TestReturn{
				Score:         -1,
				NumberOfDebug: 1,
			},
		},
		{
			name: "2 unpinned dependencies for 1 ecosystem shows 2 warn messages",
			findings: []finding.Finding{
				{
					Probe:   "pinsDependencies",
					Outcome: finding.OutcomeNegative,
					Location: &finding.Location{
						Type:      finding.FileTypeText,
						Path:      "test-file",
						LineStart: &testLineStart,
						LineEnd:   &testLineEnd,
						Snippet:   &testSnippet,
					},
					Values: map[string]int{
						"dependencyType": 6, // pip type
					},
				},
				{
					Probe:   "pinsDependencies",
					Outcome: finding.OutcomeNegative,
					Location: &finding.Location{
						Type:      finding.FileTypeText,
						Path:      "test-file",
						LineStart: &testLineStart,
						LineEnd:   &testLineEnd,
						Snippet:   &testSnippet,
					},
					Values: map[string]int{
						"dependencyType": 6, // pip type
					},
				},
			},
			result: scut.TestReturn{
				Score:        0,
				NumberOfWarn: 2,
				NumberOfInfo: 1,
			},
		},
		{
			name: "2 unpinned dependencies for 1 ecosystem shows 2 warn messages",
			findings: []finding.Finding{
				{
					Probe:   "pinsDependencies",
					Outcome: finding.OutcomeNegative,
					Location: &finding.Location{
						Type:      finding.FileTypeText,
						Path:      "test-file",
						LineStart: &testLineStart,
						LineEnd:   &testLineEnd,
						Snippet:   &testSnippet,
					},
					Values: map[string]int{
						"dependencyType": 6, // pip type
					},
				},
				{
					Probe:   "pinsDependencies",
					Outcome: finding.OutcomeNegative,
					Location: &finding.Location{
						Type:      finding.FileTypeText,
						Path:      "test-file",
						LineStart: &testLineStart,
						LineEnd:   &testLineEnd,
						Snippet:   &testSnippet,
					},
					Values: map[string]int{
						"dependencyType": 3, // go type
					},
				},
			},
			result: scut.TestReturn{
				Score:        0,
				NumberOfWarn: 2,
				NumberOfInfo: 2,
			},
		},
		{
			name: "GitHub Actions ecosystem with GitHub-owned pinned",
			findings: []finding.Finding{
				{
					Probe:   "pinsDependencies",
					Outcome: finding.OutcomePositive,
					Location: &finding.Location{
						Type:      finding.FileTypeText,
						Path:      "test-file",
						LineStart: &testLineStart,
						LineEnd:   &testLineEnd,
						Snippet:   &testSnippet,
					},
					Values: map[string]int{
						"dependencyType": 0, // GH Action type
					},
				},
			},
			result: scut.TestReturn{
				Score:        10,
				NumberOfInfo: 1,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			got := PinningDependencies(tt.name, tt.findings, &dl)
			scut.ValidateTestReturn(t, tt.name, &tt.expected, &actual, &dl)
		})
	}
}

func stringAsPointer(s string) *string {
	return &s
}

func Test_addWorkflowPinnedResult(t *testing.T) {
	t.Parallel()
	type args struct {
		w        *workflowPinningResult
		outcome  finding.Outcome
		isGitHub bool
	}
	tests := []struct {
		name string
		want *workflowPinningResult
		args args
	}{
		{
			name: "add pinned GitHub-owned action dependency",
			args: args{
				outcome:  finding.OutcomePositive,
				w:        &workflowPinningResult{},
				isGitHub: true,
			},
			want: &workflowPinningResult{
				thirdParties: pinnedResult{
					pinned: 0,
					total:  0,
				},
				gitHubOwned: pinnedResult{
					pinned: 1,
					total:  1,
				},
			},
		},
		{
			name: "add unpinned GitHub-owned action dependency",
			args: args{
				outcome:  finding.OutcomeNegative,
				w:        &workflowPinningResult{},
				isGitHub: true,
			},
			want: &workflowPinningResult{
				thirdParties: pinnedResult{
					pinned: 0,
					total:  0,
				},
				gitHubOwned: pinnedResult{
					pinned: 0,
					total:  1,
				},
			},
		},
		{
			name: "add pinned Third-Party action dependency",
			args: args{
				outcome:  finding.OutcomePositive,
				w:        &workflowPinningResult{},
				isGitHub: false,
			},
			want: &workflowPinningResult{
				thirdParties: pinnedResult{
					pinned: 1,
					total:  1,
				},
				gitHubOwned: pinnedResult{
					pinned: 0,
					total:  0,
				},
			},
		},
		{
			name: "add unpinned Third-Party action dependency",
			args: args{
				outcome:  finding.OutcomeNegative,
				w:        &workflowPinningResult{},
				isGitHub: false,
			},
			want: &workflowPinningResult{
				thirdParties: pinnedResult{
					pinned: 0,
					total:  1,
				},
				gitHubOwned: pinnedResult{
					pinned: 0,
					total:  0,
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			addWorkflowPinnedResult(tt.args.outcome, tt.args.w, tt.args.isGitHub)
			if tt.want.thirdParties != tt.args.w.thirdParties {
				t.Errorf("addWorkflowPinnedResult Third-party GitHub actions mismatch (-want +got):"+
					"\nThird-party pinned: %s\nThird-party total: %s",
					cmp.Diff(tt.want.thirdParties.pinned, tt.args.w.thirdParties.pinned),
					cmp.Diff(tt.want.thirdParties.total, tt.args.w.thirdParties.total))
			}
			if tt.want.gitHubOwned != tt.args.w.gitHubOwned {
				t.Errorf("addWorkflowPinnedResult GitHub-owned GitHub actions mismatch (-want +got):"+
					"\nGitHub-owned pinned: %s\nGitHub-owned total: %s",
					cmp.Diff(tt.want.gitHubOwned.pinned, tt.args.w.gitHubOwned.pinned),
					cmp.Diff(tt.want.gitHubOwned.total, tt.args.w.gitHubOwned.total))
			}
		})
	}
}

func TestUpdatePinningResults(t *testing.T) {
	t.Parallel()
	type args struct {
		snippet        *string
		w              *workflowPinningResult
		pr             map[checker.DependencyUseType]pinnedResult
		dependencyType checker.DependencyUseType
		outcome        finding.Outcome
	}
	type want struct {
		w  *workflowPinningResult
		pr map[checker.DependencyUseType]pinnedResult
	}
	tests := []struct { //nolint:govet
		name string
		args args
		want want
	}{
		{
			name: "add pinned GitHub-owned action",
			args: args{
				dependencyType: checker.DependencyUseTypeGHAction,
				outcome:        finding.OutcomePositive,
				snippet:        stringAsPointer("actions/checkout@a81bbbf8298c0fa03ea29cdc473d45769f953675"),
				w:              &workflowPinningResult{},
				pr:             make(map[checker.DependencyUseType]pinnedResult),
			},
			want: want{
				w: &workflowPinningResult{
					thirdParties: pinnedResult{
						pinned: 0,
						total:  0,
					},
					gitHubOwned: pinnedResult{
						pinned: 1,
						total:  1,
					},
				},
				pr: make(map[checker.DependencyUseType]pinnedResult),
			},
		},
		{
			name: "add unpinned GitHub-owned action",
			args: args{
				dependencyType: checker.DependencyUseTypeGHAction,
				outcome:        finding.OutcomeNegative,
				snippet:        stringAsPointer("actions/checkout@v2"),
				w:              &workflowPinningResult{},
				pr:             make(map[checker.DependencyUseType]pinnedResult),
			},
			want: want{
				w: &workflowPinningResult{
					thirdParties: pinnedResult{
						pinned: 0,
						total:  0,
					},
					gitHubOwned: pinnedResult{
						pinned: 0,
						total:  1,
					},
				},
				pr: make(map[checker.DependencyUseType]pinnedResult),
			},
		},
		{
			name: "add pinned Third-party action",
			args: args{
				dependencyType: checker.DependencyUseTypeGHAction,
				outcome:        finding.OutcomePositive,
				w:              &workflowPinningResult{},
				snippet:        stringAsPointer("other/checkout@ffa6706ff2127a749973072756f83c532e43ed02"),
				pr:             make(map[checker.DependencyUseType]pinnedResult),
			},
			want: want{
				w: &workflowPinningResult{
					thirdParties: pinnedResult{
						pinned: 1,
						total:  1,
					},
					gitHubOwned: pinnedResult{
						pinned: 0,
						total:  0,
					},
				},
				pr: make(map[checker.DependencyUseType]pinnedResult),
			},
		},
		{
			name: "add unpinned Third-party action",
			args: args{
				dependencyType: checker.DependencyUseTypeGHAction,
				snippet:        stringAsPointer("other/checkout@v2"),
				outcome:        finding.OutcomeNegative,
				w:              &workflowPinningResult{},
				pr:             make(map[checker.DependencyUseType]pinnedResult),
			},
			want: want{
				w: &workflowPinningResult{
					thirdParties: pinnedResult{
						pinned: 0,
						total:  1,
					},
					gitHubOwned: pinnedResult{
						pinned: 0,
						total:  0,
					},
				},
				pr: make(map[checker.DependencyUseType]pinnedResult),
			},
		},
		{
			name: "add pinned pip install",
			args: args{
				dependencyType: checker.DependencyUseTypePipCommand,
				outcome:        finding.OutcomePositive,
				w:              &workflowPinningResult{},
				pr:             make(map[checker.DependencyUseType]pinnedResult),
			},
			want: want{
				w: &workflowPinningResult{},
				pr: map[checker.DependencyUseType]pinnedResult{
					checker.DependencyUseTypePipCommand: {
						pinned: 1,
						total:  1,
					},
				},
			},
		},
		{
			name: "add unpinned pip install",
			args: args{
				dependencyType: checker.DependencyUseTypePipCommand,
				outcome:        finding.OutcomeNegative,
				w:              &workflowPinningResult{},
				pr:             make(map[checker.DependencyUseType]pinnedResult),
			},
			want: want{
				w: &workflowPinningResult{},
				pr: map[checker.DependencyUseType]pinnedResult{
					checker.DependencyUseTypePipCommand: {
						pinned: 0,
						total:  1,
					},
				},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			updatePinningResults(tc.args.dependencyType, tc.args.outcome, tc.args.snippet, tc.args.w, tc.args.pr)
			if tc.want.w.thirdParties != tc.args.w.thirdParties {
				t.Errorf("updatePinningResults Third-party GitHub actions mismatch (-want +got):"+
					"\nThird-party pinned: %s\nThird-party total: %s",
					cmp.Diff(tc.want.w.thirdParties.pinned, tc.args.w.thirdParties.pinned),
					cmp.Diff(tc.want.w.thirdParties.total, tc.args.w.thirdParties.total))
			}
			if tc.want.w.gitHubOwned != tc.args.w.gitHubOwned {
				t.Errorf("updatePinningResults GitHub-owned GitHub actions mismatch (-want +got):"+
					"\nGitHub-owned pinned: %s\nGitHub-owned total: %s",
					cmp.Diff(tc.want.w.gitHubOwned.pinned, tc.args.w.gitHubOwned.pinned),
					cmp.Diff(tc.want.w.gitHubOwned.total, tc.args.w.gitHubOwned.total))
			}
			for dependencyUseType := range tc.want.pr {
				if tc.want.pr[dependencyUseType] != tc.args.pr[dependencyUseType] {
					t.Errorf("updatePinningResults %s mismatch (-want +got):\npinned: %s\ntotal: %s",
						dependencyUseType,
						cmp.Diff(tc.want.pr[dependencyUseType].pinned, tc.args.pr[dependencyUseType].pinned),
						cmp.Diff(tc.want.pr[dependencyUseType].total, tc.args.pr[dependencyUseType].total))
				}
			}
		})
	}
}
