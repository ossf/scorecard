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
	"os"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	scut "github.com/ossf/scorecard/v5/utests"
)

func TestPinningDependencies(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		path    string
		files   []string
		want    scut.TestReturn
		wantErr bool
	}{
		{
			name: "Dockerfile",
			path: "./raw/testdata/Dockerfile-script-ok",
			files: []string{
				"Dockerfile-script-ok",
			},
			want: scut.TestReturn{
				Score:        10,
				NumberOfInfo: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockRepo := mockrepo.NewMockRepoClient(ctrl)
			mockRepo.EXPECT().GetDefaultBranchName().Return("main", nil).AnyTimes()
			mockRepo.EXPECT().URI().Return("github.com/ossf/scorecard").AnyTimes()
			mockRepo.EXPECT().ListFiles(gomock.Any()).DoAndReturn(
				func(predicate func(string) (bool, error)) ([]string, error) {
					var matched []string
					for _, file := range tt.files {
						ok, err := predicate(file)
						if err != nil {
							return nil, err
						}
						if ok {
							matched = append(matched, file)
						}
					}
					return matched, nil
				},
			).AnyTimes()

			mockRepo.EXPECT().GetFileReader(gomock.Any()).DoAndReturn(func(fn string) (io.ReadCloser, error) {
				if tt.path == "" {
					return nil, nil
				}
				return os.Open(tt.path)
			}).AnyTimes()

			dl := scut.TestDetailLogger{}
			c := &checker.CheckRequest{
				RepoClient: mockRepo,
				Dlogger:    &dl,
			}

			res := PinningDependencies(c)
			scut.ValidateTestReturn(t, tt.name, &tt.want, &res, &dl)
		})
	}
}

func TestPinningDependenciesNpmLockfile(t *testing.T) {
	t.Parallel()

	const integrity = "sha512-MJTUg1kjuLeQCJ+ccE4Vpa6kKVXkPYJ2mOCQyUuKLcLQsdrMCpBPUi8qVE6+YuaJkozeA9NusTAw3hLr8Xe5EQ=="
	pinnedPackage := `{"integrity":"` + integrity + `"}`

	tests := []struct {
		name     string
		packages string
		score    int
	}{
		{
			name:     "all pinned",
			packages: `"node_modules/foo":` + pinnedPackage,
			score:    10,
		},
		{
			name: "mixed",
			packages: `"node_modules/foo":` + pinnedPackage +
				`,"node_modules/bar":{}`,
			score: 5,
		},
		{
			name:     "all unpinned",
			packages: `"node_modules/foo":{}`,
			score:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			content := `{"lockfileVersion":3,"packages":{"":{},` +
				tt.packages + `}}`

			ctrl := gomock.NewController(t)
			repo := mockrepo.NewMockRepoClient(ctrl)
			repo.EXPECT().GetDefaultBranchName().Return("main", nil).AnyTimes()
			repo.EXPECT().URI().Return("github.com/ossf/scorecard").AnyTimes()
			repo.EXPECT().ListFiles(gomock.Any()).DoAndReturn(
				func(predicate func(string) (bool, error)) ([]string, error) {
					ok, err := predicate("package-lock.json")
					if err != nil {
						return nil, err
					}
					if ok {
						return []string{"package-lock.json"}, nil
					}
					return nil, nil
				},
			).AnyTimes()
			repo.EXPECT().GetFileReader("package-lock.json").DoAndReturn(
				func(_ string) (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader(content)), nil
				},
			).AnyTimes()

			dl := scut.TestDetailLogger{}
			req := &checker.CheckRequest{
				RepoClient: repo,
				Dlogger:    &dl,
			}

			got := PinningDependencies(req)
			if got.Score != tt.score {
				t.Errorf("Score = %d, want %d; result: %+v",
					got.Score, tt.score, got)
			}
		})
	}
}
