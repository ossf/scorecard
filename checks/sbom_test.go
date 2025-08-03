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
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	scut "github.com/ossf/scorecard/v5/utests"
)

func TestSbom(t *testing.T) {
	tests := []struct {
		name     string
		releases []clients.Release
		files    []string
		err      error
		expected scut.TestReturn
	}{
		{
			name: "With Sbom in release artifacts",
			releases: []clients.Release{
				{
					Assets: []clients.ReleaseAsset{
						{
							Name: "test-sbom.cdx.json",
							URL:  "https://this.url",
						},
					},
				},
			},
			files: []string{},
			expected: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 2,
				NumberOfWarn: 0,
			},
			err: nil,
		},
		{
			name:     "With Sbom in source",
			releases: []clients.Release{},
			files:    []string{"test-sbom.spdx.json"},
			err:      nil,
			expected: scut.TestReturn{
				Score:        5,
				NumberOfInfo: 1,
				NumberOfWarn: 1,
			},
		},
		{
			name:     "Without SBOM",
			releases: []clients.Release{},
			files:    []string{},
			expected: scut.TestReturn{
				Score:        checker.MinResultScore,
				NumberOfInfo: 0,
				NumberOfWarn: 2,
			},
			err: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SCORECARD_EXPERIMENTAL", "true")
			ctrl := gomock.NewController(t)
			mockRepo := mockrepo.NewMockRepoClient(ctrl)

			mockRepo.EXPECT().ListReleases().DoAndReturn(
				func() ([]clients.Release, error) {
					if tt.err != nil {
						return nil, tt.err
					}
					return tt.releases, tt.err
				},
			).MaxTimes(1)

			mockRepo.EXPECT().ListFiles(gomock.Any()).DoAndReturn(func(predicate func(string) (bool, error)) ([]string, error) {
				return tt.files, nil
			}).AnyTimes()

			dl := scut.TestDetailLogger{}
			req := checker.CheckRequest{
				RepoClient: mockRepo,
				Ctx:        context.Background(),
				Dlogger:    &dl,
			}
			res := SBOM(&req)
			if tt.err != nil {
				if res.Error == nil {
					t.Fatalf("Expected error %v, got nil", tt.err)
				}
			}

			scut.ValidateTestReturn(t, tt.name, &tt.expected, &res, &dl)
		})
	}
}
