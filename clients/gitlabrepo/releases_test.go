// Copyright 2026 OpenSSF Scorecard Authors
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

package gitlabrepo

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/ossf/scorecard/v5/clients"
)

func TestReleasesFrom(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data []*gitlab.Release
		want []clients.Release
	}{
		{
			name: "assets links and sources populated",
			data: []*gitlab.Release{
				{
					TagName:    "v1.0.0",
					CommitPath: "/owner/repo/-/commit/abcdef1234567890",
					Assets: gitlab.ReleaseAssets{
						Sources: []gitlab.ReleaseAssetsSource{
							{
								Format: "tar.gz",
								URL:    "https://gitlab.example.com/owner/repo/-/archive/v1.0.0/repo-v1.0.0.tar.gz",
							},
						},
						Links: []*gitlab.ReleaseLink{
							{
								Name:           "app_v1.0.0_linux_amd64.tar.gz.sig",
								DirectAssetURL: "https://gitlab.example.com/owner/repo/-/releases/v1.0.0/downloads/app.sig",
							},
							{
								Name: "sbom.spdx.json",
								URL:  "https://gitlab.example.com/owner/repo/-/releases/v1.0.0/downloads/sbom.spdx.json",
							},
						},
					},
				},
			},
			want: []clients.Release{
				{
					TagName:         "v1.0.0",
					TargetCommitish: "/owner/repo/-/commit/abcdef1234567890",
					URL:             "https://gitlab.example.com/owner/repo/-/releases/v1.0.0/downloads/app.sig",
					Assets: []clients.ReleaseAsset{
						{
							Name: "tar.gz",
							URL:  "https://gitlab.example.com/owner/repo/-/archive/v1.0.0/repo-v1.0.0.tar.gz",
						},
						{
							Name: "app_v1.0.0_linux_amd64.tar.gz.sig",
							URL:  "https://gitlab.example.com/owner/repo/-/releases/v1.0.0/downloads/app.sig",
						},
						{
							Name: "sbom.spdx.json",
							URL:  "https://gitlab.example.com/owner/repo/-/releases/v1.0.0/downloads/sbom.spdx.json",
						},
					},
				},
			},
		},
		{
			name: "fallback to release links self when assets links is empty",
			data: []*gitlab.Release{
				{
					TagName:    "v2.0.0",
					CommitPath: "/owner/repo/-/commit/1234567890abcdef",
					Links: gitlab.ReleaseLinks{
						Self: "https://gitlab.example.com/api/v4/projects/1/releases/v2.0.0",
					},
					Assets: gitlab.ReleaseAssets{
						Sources: []gitlab.ReleaseAssetsSource{
							{
								Format: "zip",
								URL:    "https://gitlab.example.com/owner/repo/-/archive/v2.0.0/repo-v2.0.0.zip",
							},
						},
					},
				},
			},
			want: []clients.Release{
				{
					TagName:         "v2.0.0",
					TargetCommitish: "/owner/repo/-/commit/1234567890abcdef",
					URL:             "https://gitlab.example.com/api/v4/projects/1/releases/v2.0.0",
					Assets: []clients.ReleaseAsset{
						{
							Name: "zip",
							URL:  "https://gitlab.example.com/owner/repo/-/archive/v2.0.0/repo-v2.0.0.zip",
						},
					},
				},
			},
		},
		{
			name: "nil link entries handled safely",
			data: []*gitlab.Release{
				{
					TagName:    "v3.0.0",
					CommitPath: "/owner/repo/-/commit/0000000000000000",
					Assets: gitlab.ReleaseAssets{
						Links: []*gitlab.ReleaseLink{
							nil,
							{
								Name:           "valid.sig",
								DirectAssetURL: "https://gitlab.example.com/valid.sig",
							},
						},
					},
				},
			},
			want: []clients.Release{
				{
					TagName:         "v3.0.0",
					TargetCommitish: "/owner/repo/-/commit/0000000000000000",
					Assets: []clients.ReleaseAsset{
						{
							Name: "valid.sig",
							URL:  "https://gitlab.example.com/valid.sig",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := releasesFrom(tt.data)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("releasesFrom mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
