// Copyright 2024 OpenSSF Scorecard Authors
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

package raw

import (
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	"github.com/ossf/scorecard/v5/finding"
	scut "github.com/ossf/scorecard/v5/utests"
)

func TestHasSubstantiveBody(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		body     string
		expected bool
	}{
		{"empty body", "", false},
		{"whitespace only", "  \n  \n  ", false},
		{"auto-generated full changelog link", "**Full Changelog**: https://github.com/org/repo/compare/v1...v2", false},
		{"real content", "## What's Changed\n- Added feature X", true},
		{"single line content", "Initial release", true},
		{"content after boilerplate", "**Full Changelog**: https://link\n\n- But also real notes", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := hasSubstantiveBody(tt.body)
			if got != tt.expected {
				t.Errorf("hasSubstantiveBody(%q) = %v, want %v", tt.body, got, tt.expected)
			}
		})
	}
}

func TestExtractVersions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name: "Keep a Changelog format with content",
			content: `# Changelog
## [2.0.0] - 2024-01-15
### Added
- Feature X
## [1.0.0] - 2023-06-01
### Fixed
- Bug Y`,
			expected: []string{"2.0.0", "1.0.0"},
		},
		{
			name: "simple markdown headers with content",
			content: `# Changelog
## 3.1.0
- Feature
## 3.0.0
- Initial`,
			expected: []string{"3.1.0", "3.0.0"},
		},
		{
			name: "GNU NEWS format with content",
			content: `Version 2.0 (2024-01-15)
* New feature

Version 1.0 (2023-06-01)
* Initial release`,
			expected: []string{"2.0", "1.0"},
		},
		{
			name: "Python CHANGES.rst format with content",
			content: `3.0.0 (2024-01-15)
===================
- New feature

2.0.0 (2023-06-01)
===================
- Previous`,
			expected: []string{"3.0.0", "2.0.0"},
		},
		{
			name: "pre-release versions with content",
			content: `## [2.0.0-rc.1] - 2024-01-10
- Release candidate changes
## [1.0.0] - 2023-06-01
- Initial release`,
			expected: []string{"2.0.0-rc.1", "1.0.0"},
		},
		{
			name:     "no versions found",
			content:  "This is just a readme with no version info.",
			expected: nil,
		},
		{
			name: "v-prefixed versions with content",
			content: `## v1.2.3
- Something`,
			expected: []string{"1.2.3"},
		},
		{
			name: "version headers without content are excluded",
			content: `## [2.0.0] - 2024-01-15
## [1.0.0] - 2023-06-01
- Has content`,
			expected: []string{"1.0.0"},
		},
		{
			name: "version with only sub-headers but no content is excluded",
			content: `## [2.0.0] - 2024-01-15
### Added
### Fixed
## [1.0.0] - 2023-06-01
### Added
- Real content here`,
			expected: []string{"1.0.0"},
		},
		{
			name: "version with only empty lines is excluded",
			content: `## [2.0.0] - 2024-01-15


## [1.0.0] - 2023-06-01
- Content`,
			expected: []string{"1.0.0"},
		},
		{
			name: "all versions have content",
			content: `## [2.0.0] - 2024-01-15
- Feature A
## [1.0.0] - 2023-06-01
- Feature B`,
			expected: []string{"2.0.0", "1.0.0"},
		},
		{
			name: "RST underlines are not content",
			content: `2.0.0 (2024-01-15)
===================

1.0.0 (2023-06-01)
===================
- Actual content`,
			expected: []string{"1.0.0"},
		},
		{
			name: "duplicate versions only counted once",
			content: `## [1.0.0] - 2024-01-15
- First entry
## [1.0.0] - 2024-01-15
- Duplicate entry`,
			expected: []string{"1.0.0"},
		},
		{
			name: "version in prose is not a header",
			content: `# Changelog
## [2.0.0] - 2024-01-15
This requires version 3.0 of Python
- Actual change`,
			expected: []string{"2.0.0"},
		},
		{
			name: "gaming attempt: empty version entries",
			content: `## 2.0.0
## 1.0.0`,
			expected: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := extractVersions(strings.NewReader(tt.content))
			if diff := cmp.Diff(tt.expected, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestChangelog(t *testing.T) {
	t.Parallel()

	changelogContent := `# Changelog
## [2.0.0] - 2024-01-15
### Added
- Feature
## [1.0.0] - 2023-06-01
### Fixed
- Bug
`

	tests := []struct {
		name             string
		releases         []clients.Release
		files            []string
		changelogContent string
		expected         checker.ChangelogData
	}{
		{
			name:             "changelog covers all releases",
			changelogContent: changelogContent,
			releases: []clients.Release{
				{TagName: "v2.0.0"},
				{TagName: "v1.0.0"},
			},
			files: []string{"CHANGELOG.md"},
			expected: checker.ChangelogData{
				ChangelogFiles: []checker.File{
					{Path: "CHANGELOG.md", Type: finding.FileTypeSource},
				},
				ChangelogVersions:     []string{"2.0.0", "1.0.0"},
				TotalReleases:         2,
				ReleasesWithChangelog: 2,
			},
		},
		{
			name:             "changelog missing a release",
			changelogContent: changelogContent,
			releases: []clients.Release{
				{TagName: "v3.0.0"},
				{TagName: "v2.0.0"},
				{TagName: "v1.0.0"},
			},
			files: []string{"CHANGELOG.md"},
			expected: checker.ChangelogData{
				ChangelogFiles: []checker.File{
					{Path: "CHANGELOG.md", Type: finding.FileTypeSource},
				},
				ChangelogVersions:     []string{"2.0.0", "1.0.0"},
				TotalReleases:         3,
				ReleasesWithChangelog: 2,
			},
		},
		{
			name:             "changelog file but no releases",
			changelogContent: changelogContent,
			releases:         []clients.Release{},
			files:            []string{"CHANGELOG.md"},
			expected: checker.ChangelogData{
				ChangelogFiles: []checker.File{
					{Path: "CHANGELOG.md", Type: finding.FileTypeSource},
				},
				ChangelogVersions: []string{"2.0.0", "1.0.0"},
			},
		},
		{
			name:     "no changelog file, no releases",
			releases: []clients.Release{},
			files:    []string{},
			expected: checker.ChangelogData{},
		},
		{
			name:             "RELEASE-NOTES file detected",
			changelogContent: "Release 1.0.0\n- Initial\n",
			releases:         []clients.Release{},
			files:            []string{"RELEASE-NOTES.md"},
			expected: checker.ChangelogData{
				ChangelogFiles: []checker.File{
					{Path: "RELEASE-NOTES.md", Type: finding.FileTypeSource},
				},
				ChangelogVersions: []string{"1.0.0"},
			},
		},
		{
			name: "empty changelog entries don't count",
			changelogContent: `## [2.0.0]
## [1.0.0]
`,
			releases: []clients.Release{
				{TagName: "v2.0.0"},
				{TagName: "v1.0.0"},
			},
			files: []string{"CHANGELOG.md"},
			expected: checker.ChangelogData{
				ChangelogFiles: []checker.File{
					{Path: "CHANGELOG.md", Type: finding.FileTypeSource},
				},
				TotalReleases:         2,
				ReleasesWithChangelog: 0,
			},
		},
		{
			name:             "release body counts as changelog even without file entry",
			changelogContent: changelogContent,
			releases: []clients.Release{
				{TagName: "v3.0.0", Body: "## What's Changed\n- New feature\n"},
				{TagName: "v2.0.0"},
			},
			files: []string{"CHANGELOG.md"},
			expected: checker.ChangelogData{
				ChangelogFiles: []checker.File{
					{Path: "CHANGELOG.md", Type: finding.FileTypeSource},
				},
				ChangelogVersions:     []string{"2.0.0", "1.0.0"},
				TotalReleases:         2,
				ReleasesWithChangelog: 2, // v3.0.0 via body, v2.0.0 via changelog file
			},
		},
		{
			name: "release body with only auto-generated link does not count",
			changelogContent: `## [1.0.0]
- Content
`,
			releases: []clients.Release{
				{TagName: "v2.0.0", Body: "**Full Changelog**: https://github.com/org/repo/compare/v1.0.0...v2.0.0"},
				{TagName: "v1.0.0"},
			},
			files: []string{"CHANGELOG.md"},
			expected: checker.ChangelogData{
				ChangelogFiles: []checker.File{
					{Path: "CHANGELOG.md", Type: finding.FileTypeSource},
				},
				ChangelogVersions:     []string{"1.0.0"},
				TotalReleases:         2,
				ReleasesWithChangelog: 1, // only v1.0.0 via changelog file
			},
		},
		{
			name: "no changelog file but releases have bodies",
			releases: []clients.Release{
				{TagName: "v2.0.0", Body: "## Changes\n- Feature A\n"},
				{TagName: "v1.0.0", Body: "Initial release\n"},
			},
			files: []string{},
			expected: checker.ChangelogData{
				TotalReleases:         2,
				ReleasesWithChangelog: 2,
			},
		},
		{
			name:             "lookback limited to 5 releases",
			changelogContent: changelogContent,
			releases: []clients.Release{
				{TagName: "v6.0.0"},
				{TagName: "v5.0.0"},
				{TagName: "v4.0.0"},
				{TagName: "v3.0.0"},
				{TagName: "v2.0.0"},
				{TagName: "v1.0.0"},
			},
			files: []string{"CHANGELOG.md"},
			expected: checker.ChangelogData{
				ChangelogFiles: []checker.File{
					{Path: "CHANGELOG.md", Type: finding.FileTypeSource},
				},
				ChangelogVersions:     []string{"2.0.0", "1.0.0"},
				TotalReleases:         5,
				ReleasesWithChangelog: 1, // only v2.0.0 is in the changelog within lookback
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockRepo := mockrepo.NewMockRepoClient(ctrl)

			mockRepo.EXPECT().ListReleases().DoAndReturn(
				func() ([]clients.Release, error) {
					return tt.releases, nil
				},
			).MaxTimes(1)

			mockRepo.EXPECT().ListFiles(gomock.Any()).DoAndReturn(
				func(predicate func(string) (bool, error)) ([]string, error) {
					var matched []string
					for _, f := range tt.files {
						ok, err := predicate(f)
						if err != nil {
							return nil, err
						}
						if ok {
							matched = append(matched, f)
						}
					}
					return matched, nil
				},
			).AnyTimes()

			if tt.changelogContent != "" {
				mockRepo.EXPECT().GetFileReader(gomock.Any()).DoAndReturn(
					func(path string) (io.ReadCloser, error) {
						return io.NopCloser(strings.NewReader(tt.changelogContent)), nil
					},
				).MaxTimes(1)
			}

			dl := scut.TestDetailLogger{}
			req := checker.CheckRequest{
				RepoClient: mockRepo,
				Ctx:        t.Context(),
				Dlogger:    &dl,
			}
			res, err := Changelog(&req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !cmp.Equal(res, tt.expected) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(tt.expected, res))
			}
		})
	}
}
