// Copyright 2021 OpenSSF Scorecard Authors
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

// Copyright 2026 OpenSSF Scorecard Authors
// Licensed under the Apache License, Version 2.0


package checks

import (
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	scut "github.com/ossf/scorecard/v5/utests"
)


func TestArtifactIntegrity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		releases []mockRelease
		expected scut.TestReturn
	}{

		{
			name: "checksum present – stem match",
			releases: []mockRelease{
				{Tag: "v1.0", Assets: []string{"app.tar.gz", "app.sha256"}},
			},
			expected: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 1,
			},
		},
		{
			name: "catch-all checksum file (checksums.txt)",
			releases: []mockRelease{
				{Tag: "v1.0", Assets: []string{"app.tar.gz", "other.zip", "checksums.txt"}},
			},
			expected: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 1,
			},
		},
		{
			name: "catch-all SHA256SUMS",
			releases: []mockRelease{
				{Tag: "v1.0", Assets: []string{"app.tar.gz", "SHA256SUMS"}},
			},
			expected: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 1,
			},
		},

		{
			name: "detached signature (.sig)",
			releases: []mockRelease{
				{Tag: "v1.0", Assets: []string{"app.tar.gz", "app.sig"}},
			},
			expected: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 1,
			},
		},
		{
			name: "detached signature (.asc)",
			releases: []mockRelease{
				{Tag: "v1.0", Assets: []string{"app.tar.gz", "app.tar.gz.asc"}},
			},
			expected: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 1,
			},
		},

		{
			name: "SLSA provenance (.intoto.jsonl)",
			releases: []mockRelease{
				{Tag: "v1.0", Assets: []string{"app.tar.gz", "app.intoto.jsonl"}},
			},
			expected: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 1,
			},
		},
		{
			name: "SLSA multiple.intoto.jsonl",
			releases: []mockRelease{
				{Tag: "v1.0", Assets: []string{"app.tar.gz", "lib.tar.gz", "multiple.intoto.jsonl"}},
			},
			expected: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 1,
			},
		},

		{
			name: "multiple checksum formats",
			releases: []mockRelease{
				{Tag: "v1.0", Assets: []string{"checksums.txt", "app.sha512"}},
			},
			expected: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 1,
			},
		},

		{
			name: "no integrity files",
			releases: []mockRelease{
				{Tag: "v1.0", Assets: []string{"app.tar.gz"}},
			},
			expected: scut.TestReturn{
				Score:        checker.MinResultScore,
				NumberOfWarn: 1,
			},
		},
		{
			name: "binary asset only – no verification",
			releases: []mockRelease{
				{Tag: "v1.0", Assets: []string{"app.exe"}},
			},
			expected: scut.TestReturn{
				Score:        checker.MinResultScore,
				NumberOfWarn: 1,
			},
		},
		{
			name: "empty assets list",
			releases: []mockRelease{
				{Tag: "v1.0", Assets: []string{}},
			},
			expected: scut.TestReturn{
				Score:        checker.MinResultScore,
				NumberOfWarn: 1,
			},
		},

		{
			name: "mixed releases (proportional score ~5)",
			releases: []mockRelease{
				{Tag: "v1.0", Assets: []string{"app.sha256"}},
				{Tag: "v1.1", Assets: []string{"app.tar.gz"}},
			},
			expected: scut.TestReturn{
				Score:        5,
				NumberOfWarn: 1,
				NumberOfInfo: 1,
			},
		},

		{
			name: "source-only assets ignored, binary still covered",
			releases: []mockRelease{
				{Tag: "v1.0", Assets: []string{
					"source-code.tar.gz",
					"source-code.zip",
					"app.tar.gz",
					"app.tar.gz.sha256",
				}},
			},
			expected: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 1,
			},
		},
		{

			name: "release with only source-only assets – inconclusive",
			releases: []mockRelease{
				{Tag: "v1.0", Assets: []string{"source-code.tar.gz", "source-code.zip"}},
			},
			expected: scut.TestReturn{
			},
		},
		{
	
			name: "versioned source archive (v1.0.tar.gz) is source-only – inconclusive",
			releases: []mockRelease{
				{Tag: "v1.0", Assets: []string{"v1.0.tar.gz", "v1.0.zip"}},
			},
			expected: scut.TestReturn{
			},
		},

		{
			name:     "no releases (inconclusive)",
			releases: []mockRelease{},
			expected: scut.TestReturn{
			},
		},

		{

			name: "case-insensitive checksum pattern – uppercase",
			releases: []mockRelease{
				{Tag: "v1.0", Assets: []string{"app.tar.gz", "app.SHA256"}},
			},
			expected: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 1,
			},
		},
		{

			name: "case-insensitive signature pattern – uppercase .SIG",
			releases: []mockRelease{
				{Tag: "v1.0", Assets: []string{"app.tar.gz", "app.SIG"}},
			},
			expected: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 1,
			},
		},

		{
			name: "mismatched stem – no correlation",
			releases: []mockRelease{
				{Tag: "v1.0", Assets: []string{"app-v1.tar.gz", "unrelated.sha256"}},
			},
			expected: scut.TestReturn{
				Score:        checker.MinResultScore,
				NumberOfWarn: 1,
			},
		},

		{
			name: "multiple releases all covered",
			releases: []mockRelease{
				{Tag: "v1.0", Assets: []string{"app.tar.gz", "app.sha256"}},
				{Tag: "v1.1", Assets: []string{"app.tar.gz", "app.sha256"}},
				{Tag: "v1.2", Assets: []string{"app.tar.gz", "app.sha256"}},
			},
			expected: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 3,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mockrepo.NewMockRepoClient(ctrl)
			mockClient.EXPECT().
				ListReleases().
				Return(convertReleases(tt.releases), nil).
				Times(1)

			dl := scut.TestDetailLogger{}
			req := checker.CheckRequest{
				RepoClient: mockClient,
				Dlogger:    &dl,
			}

			result := EntryPointArtifactIntegrity(&req)
			scut.ValidateTestReturn(t, tt.name, &tt.expected, &result, &dl)
		})
	}
}

func TestArtifactIntegrity_ListReleasesError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockrepo.NewMockRepoClient(ctrl)
	mockClient.EXPECT().
		ListReleases().
		Return(nil, errors.New("transport error")).
		Times(1)

	dl := scut.TestDetailLogger{}
	req := checker.CheckRequest{
		RepoClient: mockClient,
		Dlogger:    &dl,
	}

	result := EntryPointArtifactIntegrity(&req)

	if result.Error == nil {
		t.Errorf("expected a non-nil Error in CheckResult when ListReleases fails, got score=%d", result.Score)
	}
}

func TestArtifactIntegrity_TierOrdering(t *testing.T) {
	t.Parallel()

	type tierCase struct {
		name   string
		assets []string
	}

	cases := []tierCase{
		{
			name:   "tier=SLSA (intoto)",
			assets: []string{"app.tar.gz", "app.intoto.jsonl"},
		},
		{
			name:   "tier=signature (asc)",
			assets: []string{"app.tar.gz", "app.tar.gz.asc"},
		},
		{
			name:   "tier=checksum (sha256)",
			assets: []string{"app.tar.gz", "app.sha256"},
		},
	}

	scores := make(map[string]int, len(cases))

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mockrepo.NewMockRepoClient(ctrl)
			mockClient.EXPECT().
				ListReleases().
				Return(convertReleases([]mockRelease{
					{Tag: "v1.0", Assets: tc.assets},
				}), nil).
				Times(1)

			dl := scut.TestDetailLogger{}
			req := checker.CheckRequest{
				RepoClient: mockClient,
				Dlogger:    &dl,
			}

			result := EntryPointArtifactIntegrity(&req)
			scores[tc.name] = result.Score

			if result.Score != checker.MaxResultScore {
				t.Errorf("%s: expected score %d, got %d", tc.name, checker.MaxResultScore, result.Score)
			}
		})
	}
}

func TestArtifactIntegrity_TierQualityInMixedReleases(t *testing.T) {
	t.Parallel()

	runMixed := func(t *testing.T, coveredAssets []string) int {
		t.Helper()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := mockrepo.NewMockRepoClient(ctrl)
		mockClient.EXPECT().
			ListReleases().
			Return(convertReleases([]mockRelease{
				{Tag: "v1.0", Assets: coveredAssets},       
				{Tag: "v1.1", Assets: []string{"app.tar.gz"}}, 
			}), nil).
			Times(1)

		dl := scut.TestDetailLogger{}
		req := checker.CheckRequest{
			RepoClient: mockClient,
			Dlogger:    &dl,
		}
		return EntryPointArtifactIntegrity(&req).Score
	}

	slsaScore := runMixed(t, []string{"app.tar.gz", "app.intoto.jsonl"})
	sigScore := runMixed(t, []string{"app.tar.gz", "app.tar.gz.asc"})
	csumScore := runMixed(t, []string{"app.tar.gz", "app.sha256"})

	if slsaScore < sigScore {
		t.Errorf("SLSA score (%d) should be >= signature score (%d)", slsaScore, sigScore)
	}
	if sigScore < csumScore {
		t.Errorf("signature score (%d) should be >= checksum score (%d)", sigScore, csumScore)
	}
}

func TestArtifactIntegrity_WeightDecay(t *testing.T) {
	t.Parallel()

	runDecay := func(releases []mockRelease) int {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := mockrepo.NewMockRepoClient(ctrl)
		mockClient.EXPECT().
			ListReleases().
			Return(convertReleases(releases), nil).
			Times(1)

		dl := scut.TestDetailLogger{}
		req := checker.CheckRequest{
			RepoClient: mockClient,
			Dlogger:    &dl,
		}
		return EntryPointArtifactIntegrity(&req).Score
	}

	recentGoodScore := runDecay([]mockRelease{
		{Tag: "v2.0", Assets: []string{"app.sha256"}},  
		{Tag: "v1.0", Assets: []string{"app.tar.gz"}},  
	})

	recentBadScore := runDecay([]mockRelease{
		{Tag: "v2.0", Assets: []string{"app.tar.gz"}},  
		{Tag: "v1.0", Assets: []string{"app.sha256"}},  
	})

	if recentGoodScore <= recentBadScore {
		t.Errorf(
			"expected recentGoodScore (%d) > recentBadScore (%d) due to release weight decay",
			recentGoodScore, recentBadScore,
		)
	}
}

type mockRelease struct {
	Tag    string
	Assets []string
}

func convertReleases(in []mockRelease) []clients.Release {
	var out []clients.Release
	for _, r := range in {
		var assets []clients.ReleaseAsset
		for _, name := range r.Assets {
			assets = append(assets, clients.ReleaseAsset{Name: name})
		}
		out = append(out, clients.Release{
			TagName: r.Tag,
			Assets:  assets,
		})
	}
	return out
}