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

package checks

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	"github.com/ossf/scorecard/v5/internal/packageclient"
	scut "github.com/ossf/scorecard/v5/utests"
)

func TestSignedRelease(t *testing.T) {
	t.Parallel()
	tests := []struct {
		err      error
		name     string
		releases []clients.Release
		expected checker.CheckResult
	}{
		{
			name: "NoReleases",
			expected: checker.CheckResult{
				Score: -1,
			},
		},
		{
			name: "Releases with no assets",
			releases: []clients.Release{
				{
					TagName:         "v1.0.0",
					URL:             "http://foo.com/v1.0.0",
					TargetCommitish: "master",
					Assets:          []clients.ReleaseAsset{},
				},
			},
			expected: checker.CheckResult{
				Score: -1,
			},
		},
		{
			name: "Releases with assets without signed artifacts",
			releases: []clients.Release{
				{
					TagName:         "v1.0.0",
					URL:             "http://foo.com/v1.0.0",
					TargetCommitish: "master",
					Assets: []clients.ReleaseAsset{
						{
							Name: "foo.txt",
							URL:  "http://foo.com/v1.0.0/foo.txt",
						},
					},
				},
			},
			expected: checker.CheckResult{
				Score: 0,
			},
		},
		{
			name: "Releases with assets with signed artifacts-asc",
			releases: []clients.Release{
				{
					TagName:         "v1.0.0",
					URL:             "http://foo.com/v1.0.0",
					TargetCommitish: "master",
					Assets: []clients.ReleaseAsset{
						{
							Name: "foo.asc",
							URL:  "http://foo.com/v1.0.0/foo.asc",
						},
					},
				},
			},
			expected: checker.CheckResult{
				Score: 8,
			},
		},
		{
			name: "Releases with assets with intoto SLSA provenance",
			releases: []clients.Release{
				{
					TagName:         "v1.0.0",
					URL:             "http://foo.com/v1.0.0",
					TargetCommitish: "master",
					Assets: []clients.ReleaseAsset{
						{
							Name: "foo.intoto.jsonl",
							URL:  "http://foo.com/v1.0.0/foo.intoto.jsonl",
						},
					},
				},
			},
			expected: checker.CheckResult{
				Score: 10,
			},
		},
		{
			name: "Releases with assets with signed artifacts-sig",
			releases: []clients.Release{
				{
					TagName:         "v1.0.0",
					URL:             "http://foo.com/v1.0.0",
					TargetCommitish: "master",
					Assets: []clients.ReleaseAsset{
						{
							Name: "foo.sig",
							URL:  "http://foo.com/v1.0.0/foo.sig",
						},
					},
				},
			},
			expected: checker.CheckResult{
				Score: 8,
			},
		},
		{
			name: "Releases with assets with signed artifacts-sign",
			releases: []clients.Release{
				{
					TagName:         "v1.0.0",
					URL:             "http://foo.com/v1.0.0",
					TargetCommitish: "master",
					Assets: []clients.ReleaseAsset{
						{
							Name: "foo.sign",
							URL:  "http://foo.com/v1.0.0/foo.sign",
						},
					},
				},
			},
			expected: checker.CheckResult{
				Score: 8,
			},
		},
		{
			name: "Releases with assets with signed artifacts-minisig",
			releases: []clients.Release{
				{
					TagName:         "v1.0.0",
					URL:             "http://foo.com/v1.0.0",
					TargetCommitish: "master",
					Assets: []clients.ReleaseAsset{
						{
							Name: "foo.minisig",
							URL:  "http://foo.com/v1.0.0/foo.minisig",
						},
					},
				},
			},
			expected: checker.CheckResult{
				Score: 8,
			},
		},
		{
			name: "Releases with assets with signed artifacts-sigstore",
			releases: []clients.Release{
				{
					TagName:         "v1.0.0",
					URL:             "http://foo.com/v1.0.0",
					TargetCommitish: "master",
					Assets: []clients.ReleaseAsset{
						{
							Name: "foo.sigstore",
							URL:  "http://foo.com/v1.0.0/foo.sigstore",
						},
					},
				},
			},
			expected: checker.CheckResult{
				Score: 8,
			},
		},
		{
			name: "Releases with assets with signed and unsigned artifacts",
			releases: []clients.Release{
				{
					TagName:         "v1.0.0",
					URL:             "http://foo.com/v1.0.0",
					TargetCommitish: "master",
					Assets: []clients.ReleaseAsset{
						{
							Name: "foo.minisig",
							URL:  "http://foo.com/v1.0.0/foo.minisig",
						},
						{
							Name: "foo.txt",
							URL:  "http://foo.com/v1.0.0/foo.txt",
						},
					},
				},
			},
			expected: checker.CheckResult{
				Score: 8,
			},
		},
		{
			name: "Multiple Releases with assets with signed and unsigned artifacts",
			releases: []clients.Release{
				{
					TagName:         "v1.0.0",
					URL:             "http://foo.com/v1.0.0",
					TargetCommitish: "master",
					Assets: []clients.ReleaseAsset{
						{
							Name: "foo.minisig",
							URL:  "http://foo.com/v1.0.0/foo.minisig",
						},
						{
							Name: "foo.txt",
							URL:  "http://foo.com/v1.0.0/foo.txt",
						},
					},
				},
				{
					TagName:         "v2.0.0",
					URL:             "http://foo.com/v2.0.0",
					TargetCommitish: "master",
					Assets: []clients.ReleaseAsset{
						{
							Name: "foo.sig",
							URL:  "http://foo.com/v2.0.0/foo.sig",
						},
						{
							Name: "foo.txt",
							URL:  "http://foo.com/v2.0.0/foo.txt",
						},
					},
				},
			},
			expected: checker.CheckResult{
				Score: 8,
			},
		},
		{
			name: "Some releases with assets with signed and unsigned artifacts",
			releases: []clients.Release{
				{
					TagName:         "v1.0.0",
					URL:             "http://foo.com/v1.0.0",
					TargetCommitish: "master",
					Assets: []clients.ReleaseAsset{
						{
							Name: "foo.txt",
							URL:  "http://foo.com/v1.0.0/foo.txt",
						},
					},
				},
				{
					TagName:         "v2.0.0",
					URL:             "http://foo.com/v2.0.0",
					TargetCommitish: "master",
					Assets: []clients.ReleaseAsset{
						{
							Name: "foo.sig",
							URL:  "http://foo.com/v2.0.0/foo.sig",
						},
						{
							Name: "foo.txt",
							URL:  "http://foo.com/v2.0.0/foo.txt",
						},
					},
				},
			},
			expected: checker.CheckResult{
				Score: 4,
			},
		},
		{
			name: "6 Releases with assets with signed artifacts",
			releases: []clients.Release{
				{
					TagName:         "v1.0.0",
					URL:             "http://foo.com/v1.0.0",
					TargetCommitish: "master",
					Assets: []clients.ReleaseAsset{
						{
							Name: "foo.minisig",
							URL:  "http://foo.com/v1.0.0/foo.minisig",
						},
						{
							Name: "foo.txt",
							URL:  "http://foo.com/v1.0.0/foo.txt",
						},
					},
				},
				{
					TagName:         "v2.0.0",
					URL:             "http://foo.com/v2.0.0",
					TargetCommitish: "master",
					Assets: []clients.ReleaseAsset{
						{
							Name: "foo.sig",
							URL:  "http://foo.com/v2.0.0/foo.sig",
						},
						{
							Name: "foo.txt",
							URL:  "http://foo.com/v2.0.0/foo.txt",
						},
					},
				},
				{
					TagName:         "v3.0.0",
					URL:             "http://foo.com/v2.0.0",
					TargetCommitish: "master",
					Assets: []clients.ReleaseAsset{
						{
							Name: "foo.sig",
							URL:  "http://foo.com/v2.0.0/foo.sig",
						},
						{
							Name: "foo.txt",
							URL:  "http://foo.com/v2.0.0/foo.txt",
						},
					},
				},
				{
					TagName:         "v4.0.0",
					URL:             "http://foo.com/v2.0.0",
					TargetCommitish: "master",
					Assets: []clients.ReleaseAsset{
						{
							Name: "foo.sig",
							URL:  "http://foo.com/v2.0.0/foo.sig",
						},
						{
							Name: "foo.txt",
							URL:  "http://foo.com/v2.0.0/foo.txt",
						},
					},
				},
				{
					TagName:         "v5.0.0",
					URL:             "http://foo.com/v2.0.0",
					TargetCommitish: "master",
					Assets: []clients.ReleaseAsset{
						{
							Name: "foo.sig",
							URL:  "http://foo.com/v2.0.0/foo.sig",
						},
						{
							Name: "foo.txt",
							URL:  "http://foo.com/v2.0.0/foo.txt",
						},
					},
				},
				{
					TagName:         "v6.0.0",
					URL:             "http://foo.com/v2.0.0",
					TargetCommitish: "master",
					Assets: []clients.ReleaseAsset{
						{
							Name: "foo.sig",
							URL:  "http://foo.com/v2.0.0/foo.sig",
						},
						{
							Name: "foo.txt",
							URL:  "http://foo.com/v2.0.0/foo.txt",
						},
					},
				},
			},
			expected: checker.CheckResult{
				Score: 8,
			},
		},

		{
			name: "Error getting releases",
			err:  errors.New("Error getting releases"),
			expected: checker.CheckResult{
				Score: -1,
				Error: errors.New("Error getting releases"),
			},
		},
		{
			name: "9 Releases with assets with signed artifacts",
			releases: []clients.Release{
				release("v0.8.5"),
				release("v0.8.4"),
				release("v0.8.3"),
				release("v0.8.2"),
				release("v0.8.1"),
				release("v0.8.0"),
				release("v0.7.0"),
				release("v0.6.0"),
				release("v0.5.0"),
				release("v0.4.0"),
				release("v0.3.0"),
				release("v0.2.0"),
				release("v0.1.0"),
				release("v0.0.6"),
				release("v0.0.5"),
				release("v0.0.4"),
				release("v0.0.3"),
				release("v0.0.2"),
				release("v0.0.1"),
			},
			expected: checker.CheckResult{
				Score: 8,
			},
		},

		{
			name: "Error getting releases",
			err:  errors.New("Error getting releases"),
			expected: checker.CheckResult{
				Score: -1,
				Error: errors.New("Error getting releases"),
			},
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockRepoC := mockrepo.NewMockRepoClient(ctrl)
			mockRepoC.EXPECT().ListReleases().DoAndReturn(
				func() ([]clients.Release, error) {
					if tt.err != nil {
						return nil, tt.err
					}
					return tt.releases, tt.err
				},
			).MinTimes(1)

			mockRepo := mockrepo.NewMockRepo(ctrl)
			mockRepo.EXPECT().Host().DoAndReturn(
				func() string {
					return ""
				},
			).AnyTimes()
			mockRepo.EXPECT().Path().DoAndReturn(
				func() string {
					return ""
				},
			).AnyTimes()

			mockPkgC := mockrepo.NewMockProjectPackageClient(ctrl)
			mockPkgC.EXPECT().GetProjectPackageVersions(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, host, project string) (*packageclient.ProjectPackageVersions, error) {
					v := packageclient.ProjectPackageVersions{}
					return &v, nil
				},
			).AnyTimes()

			req := checker.CheckRequest{
				RepoClient:    mockRepoC,
				Repo:          mockRepo,
				ProjectClient: mockPkgC,
			}
			req.Dlogger = &scut.TestDetailLogger{}
			res := SignedReleases(&req)

			if tt.err != nil {
				if res.Error == nil {
					t.Errorf("Expected error %v, got nil", tt.err)
				}
				// return as we don't need to check the rest of the fields.
				return
			}

			if res.Score != tt.expected.Score {
				t.Errorf("Expected score %d, got %d for %v", tt.expected.Score, res.Score, tt.name)
			}
			ctrl.Finish()
		})
	}
}

func release(version string) clients.Release {
	return clients.Release{
		TagName:         version,
		URL:             fmt.Sprintf("https://github.com/test/test_artifact/releases/tag/%s", version),
		TargetCommitish: "master",
		Assets: []clients.ReleaseAsset{
			{
				Name: fmt.Sprintf("%s_checksums.txt", version),
				URL:  fmt.Sprintf("https://github.com/test/repo/releases/%s/%s_checksums.txt", version, version),
			},
			{
				Name: fmt.Sprintf("%s_checksums.txt.sig", version),
				URL:  fmt.Sprintf("https://github.com/test/repo/releases/%s/%s_checksums.txt.sig", version, version),
			},
			{
				Name: fmt.Sprintf("%s_darwin_x86_64.tar.gz", version),
				URL:  fmt.Sprintf("https://github.com/test/repo/releases/%s/%s_darwin_x86_64.tar.gz", version, version),
			},
			{
				Name: fmt.Sprintf("%s_Linux_arm64.tar.gz", version),
				URL:  fmt.Sprintf("https://github.com/test/repo/releases/%s/%s_Linux_arm64.tar.gz", version, version),
			},
			{
				Name: fmt.Sprintf("%s_Linux_i386.tar.gz", version),
				URL:  fmt.Sprintf("https://github.com/test/repo/releases/%s/%s_Linux_i386.tar.gz", version, version),
			},
			{
				Name: fmt.Sprintf("%s_Linux_x86_64.tar.gz", version),
				URL:  fmt.Sprintf("https://github.com/test/repo/releases/%s/%s_Linux_x86_64.tar.gz", version, version),
			},
			{
				Name: fmt.Sprintf("%s_windows_i386.tar.gz", version),
				URL:  fmt.Sprintf("https://github.com/test/repo/releases/%s/%s_windows_i386.tar.gz", version, version),
			},
			{
				Name: fmt.Sprintf("%s_windows_x86_64.tar.gz", version),
				URL:  fmt.Sprintf("https://github.com/test/repo/releases/%s/%s_windows_x86_64.tar.gz", version, version),
			},
		},
	}
}
