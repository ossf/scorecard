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

package gitlabrepo

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ossf/scorecard/v4/clients"
)

func TestListSboms(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		sbomData graphqlSbomData
		want     []clients.Sbom
		wantErr  bool
	}{
		{
			name:     "Empty sbom data",
			sbomData: graphqlSbomData{},
			want:     nil,
			wantErr:  true,
		},
		{
			name: "sbomData.Project.Releases.Nodes is nil",
			sbomData: graphqlSbomData{
				Project: graphqlProject{
					Releases: graphqlReleases{
						Nodes: nil,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "sbomData.Project.Releases.Nodes[0].Assets.Links.Nodes is nil",
			sbomData: graphqlSbomData{
				Project: graphqlProject{
					Releases: graphqlReleases{
						Nodes: []graphqlReleaseNode{
							{
								Name: "v1.2.3",
								Assets: graphqlReleaseAsset{
									Links: graphqlReleaseAssetLinks{
										Nodes: nil,
									},
								},
							},
						},
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "sbomData.Project.Pipelines.Nodes is nil",
			sbomData: graphqlSbomData{
				Project: graphqlProject{
					Releases: graphqlReleases{
						Nodes: []graphqlReleaseNode{
							{
								Name: "v1.2.3",
								Assets: graphqlReleaseAsset{
									Links: graphqlReleaseAssetLinks{
										Nodes: testassetlinks,
									},
								},
							},
						},
					},
					Pipelines: graphqlPipelines{
						Nodes: nil,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "successful run",
			sbomData: graphqlSbomData{
				Project: graphqlProject{
					Releases: graphqlReleases{
						Nodes: []graphqlReleaseNode{
							{
								Name: "v1.2.3",
								Assets: graphqlReleaseAsset{
									Links: graphqlReleaseAssetLinks{
										Nodes: testassetlinks,
									},
								},
							},
						},
					},
					Pipelines: graphqlPipelines{
						Nodes: testpipelines,
					},
				},
			},
			want: []clients.Sbom{
				{
					Name:   "tool-v1.23.4.cdx.json",
					Origin: "repositoryRelease",
					URL:    "https://test-url.com/uploads/bef0f126121567f3d1e11499c2f96e49/tool-v1.23.4.cdx.json",
				},
				{
					Name:   "gl-dependency-scanning-report.json",
					Origin: "repositoryCICD",
					URL:    "/repo/-/jobs/6470404028/artifacts/download?file_type=dependency_scanning",
				},
				{
					Name:   "gl-sbom.cdx.json.gz",
					Origin: "repositoryCICD",
					URL:    "/repo/-/jobs/6470404028/artifacts/download?file_type=cyclonedx",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := &sbomHandler{}

			repoURL := repoURL{
				owner:     "ossf-tests",
				commitSHA: clients.HeadSHA,
			}
			handler.init(&repoURL)
			got, err := handler.listSboms(tt.sbomData)
			if (err != nil) != tt.wantErr {
				t.Fatalf("listSboms error: %v, wantedErr: %t", err, tt.wantErr)
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("listSboms() = %v, want %v", got, cmp.Diff(got, tt.want))
			}
		})
	}
}

var testassetlinks = []graphqlReleaseAssetLinksNode{
	{
		Name:            "LICENSE",
		URL:             "https://test-url.com/uploads/bef0f126121567f3d1e11499c2f96e49/LICENSE",
		LinkType:        "OTHER",
		DirectAssetPath: "",
		DirectAssetURL:  "https://test-url.com/uploads/bef0f126121567f3d1e11499c2f96e49/LICENSE",
	},
	{
		Name:            "repo-v1.23.4.whl",
		URL:             "https://test-url.com/uploads/bef0f126121567f3d1e11499c2f96e49/repo-v1.23.4.whl",
		LinkType:        "PACKAGE",
		DirectAssetPath: "",
		DirectAssetURL:  "https://test-url.com/uploads/bef0f126121567f3d1e11499c2f96e49/repo-v1.23.4.whl",
	},
	{
		Name:            "tool-v1.23.4.cdx.json",
		URL:             "https://test-url.com/uploads/bef0f126121567f3d1e11499c2f96e49/tool-v1.23.4.cdx.json",
		LinkType:        "OTHER",
		DirectAssetPath: "",
		DirectAssetURL:  "https://test-url.com/uploads/bef0f126121567f3d1e11499c2f96e49/tool-v1.23.4.cdx.json",
	},
}

func TestCheckReleaseArtifacts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		assetlinks []graphqlReleaseAssetLinksNode
		want       []clients.Sbom
	}{
		{
			name:       "no release links",
			assetlinks: []graphqlReleaseAssetLinksNode{},
			want:       nil,
		},
		{
			name: "release links without matches",
			assetlinks: []graphqlReleaseAssetLinksNode{
				{
					Name:            "LICENSE",
					URL:             "https://test-url.com/uploads/bef0f126121567f3d1e11499c2f96e49/LICENSE",
					LinkType:        "OTHER",
					DirectAssetPath: "",
					DirectAssetURL:  "https://test-url.com/uploads/bef0f126121567f3d1e11499c2f96e49/LICENSE",
				},
				{
					Name:            "repo-v1.23.4.whl",
					URL:             "https://test-url.com/uploads/bef0f126121567f3d1e11499c2f96e49/repo-v1.23.4.whl",
					LinkType:        "PACKAGE",
					DirectAssetPath: "",
					DirectAssetURL:  "https://test-url.com/uploads/bef0f126121567f3d1e11499c2f96e49/repo-v1.23.4.whl",
				},
			},
			want: nil,
		},
		{
			name:       "release links with matches",
			assetlinks: testassetlinks,
			want: []clients.Sbom{
				{
					Name:   "tool-v1.23.4.cdx.json",
					Origin: "repositoryRelease",
					URL:    "https://test-url.com/uploads/bef0f126121567f3d1e11499c2f96e49/tool-v1.23.4.cdx.json",
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := &sbomHandler{}

			repoURL := repoURL{
				owner:     "ossf-tests",
				commitSHA: clients.HeadSHA,
			}
			handler.init(&repoURL)
			handler.checkReleaseArtifacts(tt.assetlinks)

			sboms := handler.sboms
			if !cmp.Equal(sboms, tt.want) {
				t.Errorf("checkReleaseArtifacts() = %v, want %v, diff %v", sboms, tt.want, cmp.Diff(sboms, tt.want))
			}
		})
	}
}

var testpipelines = []graphqlPipelineNode{
	{
		Status: "SUCCESS",
		JobArtifacts: []graphqlJobArtifact{
			{
				Name:         "6470404015.zip",
				FileType:     "ARCHIVE",
				DownloadPath: "/repo/-/jobs/6470404015/artifacts/download?file_type=archive",
			},
			{
				Name:         "metadata.gz",
				FileType:     "METADATA",
				DownloadPath: "/repo/-/jobs/6470404015/artifacts/download?file_type=metadata",
			},
			{
				Name:         "job.log",
				FileType:     "TRACE",
				DownloadPath: "/repo/-/jobs/6470404015/artifacts/download?file_type=trace",
			},
			{
				Name:         "gl-dependency-scanning-report.json",
				FileType:     "DEPENDENCY_SCANNING",
				DownloadPath: "/repo/-/jobs/6470404028/artifacts/download?file_type=dependency_scanning",
			},
			{
				Name:         "gl-sbom.cdx.json.gz",
				FileType:     "CYCLONEDX",
				DownloadPath: "/repo/-/jobs/6470404028/artifacts/download?file_type=cyclonedx",
			},
		},
	},
	{
		Status: "SUCCESS",
		JobArtifacts: []graphqlJobArtifact{
			{
				Name:         "6470404015.zip",
				FileType:     "ARCHIVE",
				DownloadPath: "/repo/-/jobs/6470404015/artifacts/download?file_type=archive",
			},
			{
				Name:         "metadata.gz",
				FileType:     "METADATA",
				DownloadPath: "/repo/-/jobs/6470404015/artifacts/download?file_type=metadata",
			},
			{
				Name:         "job.log",
				FileType:     "TRACE",
				DownloadPath: "/repo/-/jobs/6470404015/artifacts/download?file_type=trace",
			},
		},
	},
	{
		Status:       "FAILED",
		JobArtifacts: []graphqlJobArtifact{},
	},
	{
		Status:       "CANCELED",
		JobArtifacts: []graphqlJobArtifact{},
	},
	{
		Status:       "RUNNING",
		JobArtifacts: []graphqlJobArtifact{},
	},
	{
		Status:       "SKIPPED",
		JobArtifacts: []graphqlJobArtifact{},
	},
}

func TestCheckCICDArtifacts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		pipelines []graphqlPipelineNode
		want      []clients.Sbom
	}{
		{
			name:      "no release links",
			pipelines: []graphqlPipelineNode{},
			want:      nil,
		},
		{
			name: "release links without matches",
			pipelines: []graphqlPipelineNode{
				{
					Status: "SUCCESS",
					JobArtifacts: []graphqlJobArtifact{
						{
							Name:         "6470404015.zip",
							FileType:     "ARCHIVE",
							DownloadPath: "/repo/-/jobs/6470404015/artifacts/download?file_type=archive",
						},
						{
							Name:         "metadata.gz",
							FileType:     "METADATA",
							DownloadPath: "/repo/-/jobs/6470404015/artifacts/download?file_type=metadata",
						},
						{
							Name:         "job.log",
							FileType:     "TRACE",
							DownloadPath: "/repo/-/jobs/6470404015/artifacts/download?file_type=trace",
						},
					},
				},
				{
					Status:       "FAILED",
					JobArtifacts: []graphqlJobArtifact{},
				},
				{
					Status:       "CANCELED",
					JobArtifacts: []graphqlJobArtifact{},
				},
				{
					Status:       "RUNNING",
					JobArtifacts: []graphqlJobArtifact{},
				},
				{
					Status:       "SKIPPED",
					JobArtifacts: []graphqlJobArtifact{},
				},
			},
			want: nil,
		},
		{
			name:      "release links with matches",
			pipelines: testpipelines,
			want: []clients.Sbom{
				{
					Name:   "gl-dependency-scanning-report.json",
					Origin: "repositoryCICD",
					URL:    "/repo/-/jobs/6470404028/artifacts/download?file_type=dependency_scanning",
				},
				{
					Name:   "gl-sbom.cdx.json.gz",
					Origin: "repositoryCICD",
					URL:    "/repo/-/jobs/6470404028/artifacts/download?file_type=cyclonedx",
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := &sbomHandler{}

			repoURL := repoURL{
				owner:     "ossf-tests",
				commitSHA: clients.HeadSHA,
			}
			handler.init(&repoURL)
			handler.checkCICDArtifacts(tt.pipelines)

			sboms := handler.sboms
			if !cmp.Equal(sboms, tt.want) {
				t.Errorf("checkCICDArtifacts() = %v, want %v, diff %v", sboms, tt.want, cmp.Diff(sboms, tt.want))
			}
		})
	}
}
