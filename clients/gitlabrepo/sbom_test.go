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

func TestListsboms(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		SBOMData graphqlSBOMData
		want     []clients.SBOM
		wantErr  bool
	}{
		{
			name:     "Empty SBOM data",
			SBOMData: graphqlSBOMData{},
			want:     nil,
			wantErr:  true,
		},
		{
			name: "SBOMData.Project.Pipelines.Nodes is nil",
			SBOMData: graphqlSBOMData{
				Project: graphqlProject{
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
			SBOMData: graphqlSBOMData{
				Project: graphqlProject{
					Pipelines: graphqlPipelines{
						Nodes: testpipelines,
					},
				},
			},
			want: []clients.SBOM{
				{
					Name: "gl-dependency-scanning-report.json",
					URL:  "/repo/-/jobs/6470404028/artifacts/download?file_type=dependency_scanning",
				},
				{
					Name: "gl-SBOM.cdx.json.gz",
					URL:  "/repo/-/jobs/6470404028/artifacts/download?file_type=cyclonedx",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := &SBOMHandler{}

			repoURL := repoURL{
				owner:     "ossf-tests",
				commitSHA: clients.HeadSHA,
			}
			handler.init(&repoURL)
			got, err := handler.listSBOMs(tt.SBOMData)
			if (err != nil) != tt.wantErr {
				t.Fatalf("listsboms error: %v, wantedErr: %t", err, tt.wantErr)
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("listsboms() = %v, want %v", got, cmp.Diff(got, tt.want))
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
				Name:         "gl-SBOM.cdx.json.gz",
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
		want      []clients.SBOM
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
			want: []clients.SBOM{
				{
					Name: "gl-dependency-scanning-report.json",
					URL:  "/repo/-/jobs/6470404028/artifacts/download?file_type=dependency_scanning",
				},
				{
					Name: "gl-SBOM.cdx.json.gz",
					URL:  "/repo/-/jobs/6470404028/artifacts/download?file_type=cyclonedx",
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := &SBOMHandler{}

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
