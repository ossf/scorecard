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

package gitlabReleaseTagsAreProtected

import (
	"testing"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/finding"
)

func strPtr(s string) *string {
	return &s
}

func TestRun(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		raw      *checker.RawResults
		expected []finding.Outcome
		wantErr  bool
	}{
		{
			name:    "nil raw results",
			raw:     nil,
			wantErr: true,
		},
		{
			name: "not a GitLab repository",
			raw: &checker.RawResults{
				TagProtectionResults: checker.TagProtectionsData{
					GitLabBranches:      []string{},
					GitLabProtectedTags: []checker.GitLabProtectedTagInfo{},
				},
			},
			expected: []finding.Outcome{finding.OutcomeNotApplicable},
		},
		{
			name: "no release tags",
			raw: &checker.RawResults{
				TagProtectionResults: checker.TagProtectionsData{
					Tags:                []clients.TagRef{},
					GitLabBranches:      []string{"main"},
					GitLabProtectedTags: []checker.GitLabProtectedTagInfo{},
				},
			},
			expected: []finding.Outcome{finding.OutcomeNotApplicable},
		},
		{
			name: "unprotected release tag",
			raw: &checker.RawResults{
				TagProtectionResults: checker.TagProtectionsData{
					Tags: []clients.TagRef{
						{Name: strPtr("v1.0.0")},
					},
					GitLabBranches:      []string{"main"},
					GitLabProtectedTags: []checker.GitLabProtectedTagInfo{},
				},
			},
			expected: []finding.Outcome{finding.OutcomeFalse},
		},
		{
			name: "release tag with strongest protection (no one)",
			raw: &checker.RawResults{
				TagProtectionResults: checker.TagProtectionsData{
					Tags: []clients.TagRef{
						{Name: strPtr("v1.0.0")},
					},
					GitLabBranches: []string{"main"},
					GitLabProtectedTags: []checker.GitLabProtectedTagInfo{
						{Pattern: "v*", CreateAccessLevel: AccessLevelNone},
					},
				},
			},
			expected: []finding.Outcome{finding.OutcomeTrue},
		},
		{
			name: "release tag with strong protection (maintainer)",
			raw: &checker.RawResults{
				TagProtectionResults: checker.TagProtectionsData{
					Tags: []clients.TagRef{
						{Name: strPtr("v1.0.0")},
					},
					GitLabBranches: []string{"main"},
					GitLabProtectedTags: []checker.GitLabProtectedTagInfo{
						{Pattern: "v*", CreateAccessLevel: AccessLevelMaintainer},
					},
				},
			},
			expected: []finding.Outcome{finding.OutcomeTrue},
		},
		{
			name: "release tag with weak protection (developer)",
			raw: &checker.RawResults{
				TagProtectionResults: checker.TagProtectionsData{
					Tags: []clients.TagRef{
						{Name: strPtr("v1.0.0")},
					},
					GitLabBranches: []string{"main"},
					GitLabProtectedTags: []checker.GitLabProtectedTagInfo{
						{Pattern: "v*", CreateAccessLevel: AccessLevelDeveloper},
					},
				},
			},
			expected: []finding.Outcome{finding.OutcomeFalse},
		},
		{
			name: "multiple tags with mixed protection",
			raw: &checker.RawResults{
				TagProtectionResults: checker.TagProtectionsData{
					Tags: []clients.TagRef{
						{Name: strPtr("v1.0.0")},
						{Name: strPtr("v2.0.0")},
						{Name: strPtr("release-1.0")},
					},
					GitLabBranches: []string{"main"},
					GitLabProtectedTags: []checker.GitLabProtectedTagInfo{
						{Pattern: "v*", CreateAccessLevel: AccessLevelNone},
						{Pattern: "release-*", CreateAccessLevel: AccessLevelDeveloper},
					},
				},
			},
			expected: []finding.Outcome{
				finding.OutcomeTrue,  // v1.0.0 - no one
				finding.OutcomeTrue,  // v2.0.0 - no one
				finding.OutcomeFalse, // release-1.0 - developer
			},
		},
		{
			name: "multiple overlapping patterns - minimum access level",
			raw: &checker.RawResults{
				TagProtectionResults: checker.TagProtectionsData{
					Tags: []clients.TagRef{
						{Name: strPtr("v1.0.0")},
					},
					GitLabBranches: []string{"main"},
					GitLabProtectedTags: []checker.GitLabProtectedTagInfo{
						{Pattern: "v*", CreateAccessLevel: AccessLevelNone},
						{Pattern: "*", CreateAccessLevel: AccessLevelMaintainer},
					},
				},
			},
			expected: []finding.Outcome{finding.OutcomeTrue},
		},
		{
			name: "exact match takes precedence",
			raw: &checker.RawResults{
				TagProtectionResults: checker.TagProtectionsData{
					Tags: []clients.TagRef{
						{Name: strPtr("v1.0.0")},
					},
					GitLabBranches: []string{"main"},
					GitLabProtectedTags: []checker.GitLabProtectedTagInfo{
						{Pattern: "v1.0.0", CreateAccessLevel: AccessLevelNone},
						{Pattern: "v*", CreateAccessLevel: AccessLevelDeveloper},
					},
				},
			},
			expected: []finding.Outcome{finding.OutcomeTrue},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			findings, probeName, err := Run(tt.raw)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil {
				return
			}

			if probeName != Probe {
				t.Errorf("Run() probeName = %v, want %v", probeName, Probe)
			}

			if len(findings) != len(tt.expected) {
				t.Errorf("Run() returned %d findings, want %d", len(findings), len(tt.expected))
				return
			}

			for i, f := range findings {
				if f.Outcome != tt.expected[i] {
					t.Errorf("Finding %d: outcome = %v, want %v", i, f.Outcome, tt.expected[i])
				}
			}
		})
	}
}

func TestFindMatchingPatterns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		tagName  string
		patterns []checker.GitLabProtectedTagInfo
		want     int
	}{
		{
			name:     "no patterns",
			tagName:  "v1.0.0",
			patterns: []checker.GitLabProtectedTagInfo{},
			want:     0,
		},
		{
			name:    "exact match",
			tagName: "v1.0.0",
			patterns: []checker.GitLabProtectedTagInfo{
				{Pattern: "v1.0.0", CreateAccessLevel: AccessLevelNone},
			},
			want: 1,
		},
		{
			name:    "prefix wildcard match",
			tagName: "v1.0.0",
			patterns: []checker.GitLabProtectedTagInfo{
				{Pattern: "v*", CreateAccessLevel: AccessLevelNone},
			},
			want: 1,
		},
		{
			name:    "suffix wildcard match",
			tagName: "release-v1.0.0",
			patterns: []checker.GitLabProtectedTagInfo{
				{Pattern: "*v1.0.0", CreateAccessLevel: AccessLevelNone},
			},
			want: 1,
		},
		{
			name:    "wildcard all",
			tagName: "anything",
			patterns: []checker.GitLabProtectedTagInfo{
				{Pattern: "*", CreateAccessLevel: AccessLevelNone},
			},
			want: 1,
		},
		{
			name:    "multiple matches",
			tagName: "v1.0.0",
			patterns: []checker.GitLabProtectedTagInfo{
				{Pattern: "v*", CreateAccessLevel: AccessLevelNone},
				{Pattern: "*", CreateAccessLevel: AccessLevelMaintainer},
				{Pattern: "v1.0.0", CreateAccessLevel: AccessLevelOwner},
			},
			want: 3,
		},
		{
			name:    "no match",
			tagName: "release-1.0",
			patterns: []checker.GitLabProtectedTagInfo{
				{Pattern: "v*", CreateAccessLevel: AccessLevelNone},
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			matches := findMatchingPatterns(tt.tagName, tt.patterns)
			if len(matches) != tt.want {
				t.Errorf("findMatchingPatterns() returned %d matches, want %d", len(matches), tt.want)
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
		{name: "wildcard all", pattern: "*", tagName: "anything", want: true},
		{name: "prefix match", pattern: "v*", tagName: "v1.0.0", want: true},
		{name: "prefix no match", pattern: "v*", tagName: "release-1.0", want: false},
		{name: "suffix match", pattern: "*-v1", tagName: "release-v1", want: true},
		{name: "suffix no match", pattern: "*-v1", tagName: "release-v2", want: false},
		{name: "no wildcard exact", pattern: "v1.0.0", tagName: "v1.0.0", want: false},
		{name: "no wildcard different", pattern: "v1.0.0", tagName: "v2.0.0", want: false},
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

func TestGetMinimumAccessLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		patterns []checker.GitLabProtectedTagInfo
		want     int
	}{
		{
			name:     "empty patterns",
			patterns: []checker.GitLabProtectedTagInfo{},
			want:     AccessLevelDeveloper,
		},
		{
			name: "single pattern",
			patterns: []checker.GitLabProtectedTagInfo{
				{Pattern: "v*", CreateAccessLevel: AccessLevelMaintainer},
			},
			want: AccessLevelMaintainer,
		},
		{
			name: "multiple patterns - minimum is used",
			patterns: []checker.GitLabProtectedTagInfo{
				{Pattern: "v*", CreateAccessLevel: AccessLevelOwner},
				{Pattern: "*", CreateAccessLevel: AccessLevelMaintainer},
				{Pattern: "v1*", CreateAccessLevel: AccessLevelDeveloper},
			},
			want: AccessLevelDeveloper,
		},
		{
			name: "none access level",
			patterns: []checker.GitLabProtectedTagInfo{
				{Pattern: "v*", CreateAccessLevel: AccessLevelNone},
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
