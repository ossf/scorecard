// Copyright 2020 Security Scorecard Authors
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
	"errors"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
	scut "github.com/ossf/scorecard/v4/utests"
)

func TestSignedRelease(t *testing.T) {
	t.Parallel()
	//fieldalignment lint issue. Ignoring it as it is not important for this test.
	//nolint
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
			name: "Releases with no assests",
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
			name: "Releases with assests without signed artifacts",
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
			name: "Releases with assests with signed artifacts-asc",
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
				Score: 10,
			},
		},
		{
			name: "Releases with assests with signed artifacts-sig",
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
				Score: 10,
			},
		},
		{
			name: "Releases with assests with signed artifacts-sign",
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
				Score: 10,
			},
		},
		{
			name: "Releases with assests with signed artifacts-minisig",
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
				Score: 10,
			},
		},
		{
			name: "Releases with assests with signed and unsigned artifacts",
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
				Score: 10,
			},
		},
		{
			name: "Multiple Releases with assests with signed and unsigned artifacts",
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
				Score: 10,
			},
		},
		{
			name: "Some releases with assests with signed and unsigned artifacts",
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
				Score: 5,
			},
		},
		{
			name: "6 Releases with assests with signed artifacts",
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
				Score: 10,
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

			mockRepo := mockrepo.NewMockRepoClient(ctrl)
			mockRepo.EXPECT().ListReleases().DoAndReturn(
				func() ([]clients.Release, error) {
					if tt.err != nil {
						return nil, tt.err
					}
					return tt.releases, tt.err
				},
			).MinTimes(1)

			req := checker.CheckRequest{
				RepoClient: mockRepo,
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
