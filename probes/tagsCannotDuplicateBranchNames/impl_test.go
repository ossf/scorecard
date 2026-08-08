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

package tagsCannotDuplicateBranchNames

import (
	"testing"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/finding"
)

func TestRun(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		rawResults       *checker.RawResults
		wantOutcome      finding.Outcome
		wantFindingCount int
		wantErr          bool
	}{
		{
			name:        "nil raw results",
			rawResults:  nil,
			wantOutcome: finding.OutcomeNotApplicable,
			wantErr:     true,
		},
		{
			name: "not a GitLab repository - no branches",
			rawResults: &checker.RawResults{
				TagProtectionResults: checker.TagProtectionsData{
					GitLabBranches: []string{},
					Tags:           []clients.TagRef{},
				},
			},
			wantOutcome:      finding.OutcomeNotApplicable,
			wantFindingCount: 1,
			wantErr:          false,
		},
		{
			name: "no protected patterns - all branches unprotected",
			rawResults: &checker.RawResults{
				TagProtectionResults: checker.TagProtectionsData{
					GitLabBranches:      []string{"main", "develop"},
					GitLabProtectedTags: []checker.GitLabProtectedTagInfo{},
					Tags:                []clients.TagRef{},
				},
			},
			wantOutcome:      finding.OutcomeFalse,
			wantFindingCount: 2,
			wantErr:          false,
		},
		{
			name: "all branches protected with No one - 2 points",
			rawResults: &checker.RawResults{
				TagProtectionResults: checker.TagProtectionsData{
					GitLabBranches: []string{"main", "develop"},
					GitLabProtectedTags: []checker.GitLabProtectedTagInfo{
						{Pattern: "*", CreateAccessLevel: AccessLevelNone},
					},
					Tags: []clients.TagRef{},
				},
			},
			wantOutcome:      finding.OutcomeTrue,
			wantFindingCount: 2,
			wantErr:          false,
		},
		{
			name: "branch with maintainer protection",
			rawResults: &checker.RawResults{
				TagProtectionResults: checker.TagProtectionsData{
					GitLabBranches: []string{"main"},
					GitLabProtectedTags: []checker.GitLabProtectedTagInfo{
						{Pattern: "main", CreateAccessLevel: AccessLevelMaintainer},
					},
					Tags: []clients.TagRef{},
				},
			},
			wantOutcome:      finding.OutcomeTrue,
			wantFindingCount: 1,
			wantErr:          false,
		},
		{
			name: "branch with weak developer protection",
			rawResults: &checker.RawResults{
				TagProtectionResults: checker.TagProtectionsData{
					GitLabBranches: []string{"main"},
					GitLabProtectedTags: []checker.GitLabProtectedTagInfo{
						{Pattern: "*", CreateAccessLevel: AccessLevelDeveloper},
					},
					Tags: []clients.TagRef{},
				},
			},
			wantOutcome:      finding.OutcomeFalse,
			wantFindingCount: 1,
			wantErr:          false,
		},
		{
			name: "mixed protection levels - multiple branches",
			rawResults: &checker.RawResults{
				TagProtectionResults: checker.TagProtectionsData{
					GitLabBranches: []string{"main", "develop", "feature"},
					GitLabProtectedTags: []checker.GitLabProtectedTagInfo{
						{Pattern: "main", CreateAccessLevel: AccessLevelNone},
						{Pattern: "develop", CreateAccessLevel: AccessLevelMaintainer},
					},
					Tags: []clients.TagRef{},
				},
			},
			wantOutcome:      finding.OutcomeTrue,
			wantFindingCount: 3,
			wantErr:          false,
		},
		{
			name: "prefix wildcard pattern matches specific branches",
			rawResults: &checker.RawResults{
				TagProtectionResults: checker.TagProtectionsData{
					GitLabBranches: []string{"release-v1", "release-v2", "main"},
					GitLabProtectedTags: []checker.GitLabProtectedTagInfo{
						{Pattern: "release-*", CreateAccessLevel: AccessLevelNone},
					},
					Tags: []clients.TagRef{},
				},
			},
			wantOutcome:      finding.OutcomeTrue,
			wantFindingCount: 3,
			wantErr:          false,
		},
		{
			name: "suffix wildcard pattern matches specific branches",
			rawResults: &checker.RawResults{
				TagProtectionResults: checker.TagProtectionsData{
					GitLabBranches: []string{"v1-main", "v2-main", "develop"},
					GitLabProtectedTags: []checker.GitLabProtectedTagInfo{
						{Pattern: "*-main", CreateAccessLevel: AccessLevelMaintainer},
					},
					Tags: []clients.TagRef{},
				},
			},
			wantOutcome:      finding.OutcomeTrue,
			wantFindingCount: 3,
			wantErr:          false,
		},
		{
			name: "multiple overlapping patterns - minimum access level used",
			rawResults: &checker.RawResults{
				TagProtectionResults: checker.TagProtectionsData{
					GitLabBranches: []string{"main"},
					GitLabProtectedTags: []checker.GitLabProtectedTagInfo{
						{Pattern: "main", CreateAccessLevel: AccessLevelNone},
						{Pattern: "*", CreateAccessLevel: AccessLevelDeveloper},
					},
					Tags: []clients.TagRef{},
				},
			},
			wantOutcome:      finding.OutcomeTrue,
			wantFindingCount: 1,
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			findings, probeName, err := Run(tt.rawResults)

			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if probeName != Probe {
				t.Errorf("Run() probeName = %v, want %v", probeName, Probe)
			}

			if len(findings) != tt.wantFindingCount {
				t.Errorf("Run() returned %d findings, want %d", len(findings), tt.wantFindingCount)
			}

			if len(findings) > 0 {
				hasExpectedOutcome := false
				for _, f := range findings {
					if f.Outcome == tt.wantOutcome {
						hasExpectedOutcome = true
						break
					}
				}
				if !hasExpectedOutcome && tt.wantOutcome != finding.OutcomeNotApplicable {
					t.Errorf("Run() no finding with outcome %v", tt.wantOutcome)
				}
			}
		})
	}
}

func TestMatchesWildcard(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		pattern string
		tagName string
		want    bool
	}{
		{
			name:    "catch-all wildcard",
			pattern: "*",
			tagName: "anything",
			want:    true,
		},
		{
			name:    "prefix match",
			pattern: "release-*",
			tagName: "release-1.0",
			want:    true,
		},
		{
			name:    "prefix no match",
			pattern: "release-*",
			tagName: "hotfix-1.0",
			want:    false,
		},
		{
			name:    "suffix match",
			pattern: "*-final",
			tagName: "v1.0.0-final",
			want:    true,
		},
		{
			name:    "suffix no match",
			pattern: "*-final",
			tagName: "v1.0.0-beta",
			want:    false,
		},
		{
			name:    "no wildcard - no match",
			pattern: "main",
			tagName: "develop",
			want:    false,
		},
		{
			name:    "empty pattern",
			pattern: "",
			tagName: "main",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := matchesWildcard(tt.pattern, tt.tagName)

			if got != tt.want {
				t.Errorf("matchesWildcard(%q, %q) = %v, want %v", tt.pattern, tt.tagName, got, tt.want)
			}
		})
	}
}

func TestFindMatchingPatterns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		branchName string
		patterns   []checker.GitLabProtectedTagInfo
		want       int
	}{
		{
			name:       "no patterns",
			branchName: "main",
			patterns:   []checker.GitLabProtectedTagInfo{},
			want:       0,
		},
		{
			name:       "exact match",
			branchName: "main",
			patterns: []checker.GitLabProtectedTagInfo{
				{Pattern: "main", CreateAccessLevel: AccessLevelNone},
			},
			want: 1,
		},
		{
			name:       "wildcard all matches everything",
			branchName: "anything",
			patterns: []checker.GitLabProtectedTagInfo{
				{Pattern: "*", CreateAccessLevel: AccessLevelNone},
			},
			want: 1,
		},
		{
			name:       "prefix wildcard match",
			branchName: "release-v1",
			patterns: []checker.GitLabProtectedTagInfo{
				{Pattern: "release-*", CreateAccessLevel: AccessLevelMaintainer},
			},
			want: 1,
		},
		{
			name:       "suffix wildcard match",
			branchName: "v1-main",
			patterns: []checker.GitLabProtectedTagInfo{
				{Pattern: "*-main", CreateAccessLevel: AccessLevelMaintainer},
			},
			want: 1,
		},
		{
			name:       "multiple matching patterns",
			branchName: "main",
			patterns: []checker.GitLabProtectedTagInfo{
				{Pattern: "main", CreateAccessLevel: AccessLevelNone},
				{Pattern: "*", CreateAccessLevel: AccessLevelDeveloper},
				{Pattern: "m*", CreateAccessLevel: AccessLevelMaintainer},
			},
			want: 3,
		},
		{
			name:       "no matching patterns",
			branchName: "develop",
			patterns: []checker.GitLabProtectedTagInfo{
				{Pattern: "main", CreateAccessLevel: AccessLevelNone},
				{Pattern: "release-*", CreateAccessLevel: AccessLevelMaintainer},
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := findMatchingPatterns(tt.branchName, tt.patterns)

			if len(got) != tt.want {
				t.Errorf("findMatchingPatterns(%q) returned %d matches, want %d", tt.branchName, len(got), tt.want)
			}
		})
	}
}

func TestGetMinimumAccessLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		patterns []checker.GitLabProtectedTagInfo
		want     int
	}{
		{
			name:     "empty patterns returns developer (weak)",
			patterns: []checker.GitLabProtectedTagInfo{},
			want:     AccessLevelDeveloper,
		},
		{
			name: "single pattern",
			patterns: []checker.GitLabProtectedTagInfo{
				{Pattern: "main", CreateAccessLevel: AccessLevelMaintainer},
			},
			want: AccessLevelMaintainer,
		},
		{
			name: "multiple patterns - minimum is used",
			patterns: []checker.GitLabProtectedTagInfo{
				{Pattern: "main", CreateAccessLevel: AccessLevelOwner},
				{Pattern: "*", CreateAccessLevel: AccessLevelMaintainer},
				{Pattern: "m*", CreateAccessLevel: AccessLevelDeveloper},
			},
			want: AccessLevelDeveloper,
		},
		{
			name: "none access level is minimum",
			patterns: []checker.GitLabProtectedTagInfo{
				{Pattern: "main", CreateAccessLevel: AccessLevelNone},
				{Pattern: "*", CreateAccessLevel: AccessLevelMaintainer},
			},
			want: AccessLevelNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := getMinimumAccessLevel(tt.patterns)

			if got != tt.want {
				t.Errorf("getMinimumAccessLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}
