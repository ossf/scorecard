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
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	"github.com/ossf/scorecard/v5/finding"
	scut "github.com/ossf/scorecard/v5/utests"
)

func TestSbom(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		releases []clients.Release
		files    []string
		err      error
		expected checker.SBOMData
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
			expected: checker.SBOMData{
				SBOMFiles: []checker.SBOM{
					{
						Name: "test-sbom.cdx.json",
						File: checker.File{
							Type: finding.FileTypeURL,
							Path: "https://this.url",
						},
					},
				},
			},
			err: nil,
		},
		{
			name:     "With Sbom in source",
			releases: []clients.Release{},
			files:    []string{"test-sbom.spdx.json"},
			err:      nil,
			expected: checker.SBOMData{
				SBOMFiles: []checker.SBOM{
					{
						Name: "test-sbom.spdx.json",
						File: checker.File{
							Type: finding.FileTypeSource,
							Path: "test-sbom.spdx.json",
						},
					},
				},
			},
		},
		{
			name:     "Without SBOM",
			releases: []clients.Release{},
			files:    []string{},
			expected: checker.SBOMData{},
			err:      nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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
			res, err := SBOM(&req)
			if tt.err != nil {
				if err == nil {
					t.Fatalf("Expected error %v, got nil", tt.err)
				}
			}

			if !cmp.Equal(res, tt.expected) {
				t.Errorf("Expected %v, got %v for %v", tt.expected, res, tt.name)
			}
		})
	}
}
