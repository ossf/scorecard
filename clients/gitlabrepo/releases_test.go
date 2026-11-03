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

func Test_releasesFrom(t *testing.T) {
	t.Parallel()
	commitSHA := "7b5c3cc8fc409613f1089f85d1687e32d47f30db"
	tests := []struct {
		name string
		data []*gitlab.Release
		want []clients.Release
	}{
		{
			name: "populated Commit.ID",
			data: []*gitlab.Release{
				{
					TagName:    "v1.0.0",
					CommitPath: "/group/project/-/commit/" + commitSHA,
					Commit: gitlab.Commit{
						ID: commitSHA,
					},
					Assets: gitlab.ReleaseAssets{
						Links: []*gitlab.ReleaseLink{
							{
								DirectAssetURL: "https://gitlab.com/group/project/-/releases/v1.0.0/downloads/binary",
							},
						},
						Sources: []gitlab.ReleaseAssetsSource{
							{
								Format: "tar.gz",
								URL:    "https://gitlab.com/group/project/-/archive/v1.0.0/project-v1.0.0.tar.gz",
							},
						},
					},
				},
			},
			want: []clients.Release{
				{
					TagName:         "v1.0.0",
					TargetCommitish: commitSHA,
					URL:             "https://gitlab.com/group/project/-/releases/v1.0.0/downloads/binary",
					Assets: []clients.ReleaseAsset{
						{
							Name: "tar.gz",
							URL:  "https://gitlab.com/group/project/-/archive/v1.0.0/project-v1.0.0.tar.gz",
						},
					},
				},
			},
		},
		{
			name: "empty Commit.ID",
			data: []*gitlab.Release{
				{
					TagName:    "v2.0.0",
					CommitPath: "/owner/repo/-/commit/" + commitSHA,
					Commit:     gitlab.Commit{},
				},
			},
			want: []clients.Release{
				{
					TagName:         "v2.0.0",
					TargetCommitish: "",
				},
			},
		},
		{
			name: "empty Commit and empty CommitPath",
			data: []*gitlab.Release{
				{
					TagName: "v3.0.0",
				},
			},
			want: []clients.Release{
				{
					TagName:         "v3.0.0",
					TargetCommitish: "",
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
