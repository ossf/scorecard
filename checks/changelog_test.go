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

package checks

import (
	"io"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	scut "github.com/ossf/scorecard/v5/utests"
)

const testChangelogContent = `# Changelog
## [2.0.0] - 2024-01-15
### Added
- Feature
## [1.0.0] - 2023-06-01
### Fixed
- Bug
`

func TestChangelog(t *testing.T) {
	tests := []struct {
		name     string
		releases []clients.Release
		files    []string
		expected scut.TestReturn
	}{
		{
			name: "changelog covers release",
			releases: []clients.Release{
				{TagName: "v2.0.0"},
				{TagName: "v1.0.0"},
			},
			files: []string{"CHANGELOG.md"},
			expected: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 2,
			},
		},
		{
			name:     "changelog file only, no releases",
			releases: []clients.Release{},
			files:    []string{"CHANGELOG.md"},
			expected: scut.TestReturn{
				Score:         3, // 3 for file, 0 for releases (not applicable)
				NumberOfInfo:  1,
				NumberOfDebug: 1,
			},
		},
		{
			name:     "no changelog file, no releases",
			releases: []clients.Release{},
			files:    []string{},
			expected: scut.TestReturn{
				Score:         checker.InconclusiveResultScore,
				NumberOfWarn:  1,
				NumberOfDebug: 1,
			},
		},
		{
			name: "changelog file, release not in changelog",
			releases: []clients.Release{
				{TagName: "v3.0.0"},
			},
			files: []string{"CHANGELOG.md"},
			expected: scut.TestReturn{
				Score:        3, // 3 for file + 0 for releases
				NumberOfInfo: 1,
				NumberOfWarn: 1,
			},
		},
		{
			name: "no changelog file, releases have body text. 10/10",
			releases: []clients.Release{
				{TagName: "v2.0.0", Body: "## Changes\n- Feature A\n"},
				{TagName: "v1.0.0", Body: "Initial release\n"},
			},
			files: []string{},
			expected: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfWarn: 1,
				NumberOfInfo: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SCORECARD_EXPERIMENTAL", "true")
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

			if len(tt.files) > 0 {
				mockRepo.EXPECT().GetFileReader(gomock.Any()).DoAndReturn(
					func(path string) (io.ReadCloser, error) {
						return io.NopCloser(strings.NewReader(testChangelogContent)), nil
					},
				).MaxTimes(1)
			}

			dl := scut.TestDetailLogger{}
			req := checker.CheckRequest{
				RepoClient: mockRepo,
				Ctx:        t.Context(),
				Dlogger:    &dl,
			}
			res := Changelog(&req)
			scut.ValidateTestReturn(t, tt.name, &tt.expected, &res, &dl)
		})
	}
}
