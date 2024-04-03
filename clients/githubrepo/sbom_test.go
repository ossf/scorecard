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

package githubrepo

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v59/github"

	"github.com/ossf/scorecard/v4/clients"
)

func TestListSboms(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name         string
		responsePath string
		want         []clients.Sbom
		wantError    bool
	}{
		{
			name:         "Request Error",
			responsePath: "testdata/invalid-response",
			want:         nil,
			wantError:    true,
		},
		{
			name:         "Asset List has matches",
			responsePath: "testdata/asset-matches-response",
			want: []clients.Sbom{
				{
					Name:   "trivy_0.50.1_tar.gz.spdx",
					Origin: "repositoryRelease",
					URL:    "https://github.com/test/releases/download/v0.50.1/trivy_0.50.1_tar.gz.spdx",
					Path:   "https://api.github.com/repos/test/releases/assets/158838506",
				},
				{
					Name:   "trivy_0.50.1_tar.gz.spdx.xml",
					Origin: "repositoryRelease",
					URL:    "https://github.com/test/releases/download/v0.50.1/trivy_0.50.1_tar.gz.spdx.xml",
					Path:   "https://api.github.com/repos/test/releases/assets/158838450",
				},
			},
			wantError: false,
		},
	}
	for _, tt := range testcases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			httpClient := &http.Client{
				Transport: stubTripper{
					responsePath: tt.responsePath,
				},
			}

			client := github.NewClient(httpClient)

			handler := &sbomHandler{
				ghclient: client,
				ctx:      ctx,
			}

			repoURL := repoURL{
				owner:     "ossf-tests",
				repo:      "foo",
				commitSHA: clients.HeadSHA,
			}

			handler.init(ctx, &repoURL)
			sboms, err := handler.listSboms()

			if tt.wantError && err == nil {
				t.Fatalf("listSboms() - expected error did not occur")
			}
			if !tt.wantError && err != nil {
				t.Fatalf("listSboms() - unexpected error occurred: %v", err)
			}
			if !cmp.Equal(sboms, tt.want) {
				t.Errorf("listSboms() = %v, want %v, diff %v", sboms, tt.want, cmp.Diff(sboms, tt.want))
			}
		})
	}
}

func TestFetchGithubAPISbom(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name         string
		responsePath string
		want         []clients.Sbom
		wantError    bool
	}{
		{
			name:         "Request Error",
			responsePath: "testdata/invalid-response",
			want:         nil,
			wantError:    true,
		},
		{
			name:         "Successful Return",
			responsePath: "testdata/sbom-response",
			want: []clients.Sbom{
				{
					ID:            "SPDXRef-DOCUMENT",
					Name:          "github/example",
					Origin:        "repositoryRelease",
					URL:           "https://github.com/github/example/dependency_graph/sbom-abcdef123456",
					Path:          "github/example",
					Tool:          "Tool: GitHub.com-Dependency-Graph",
					Schema:        "SPDX",
					SchemaVersion: "SPDX-2.3",
				},
			},
			wantError: false,
		},
	}
	for _, tt := range testcases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			httpClient := &http.Client{
				Transport: stubTripper{
					responsePath: tt.responsePath,
				},
			}

			client := github.NewClient(httpClient)

			handler := &sbomHandler{
				ghclient: client,
				ctx:      ctx,
			}

			repoURL := repoURL{
				owner:     "ossf-tests",
				repo:      "foo",
				commitSHA: clients.HeadSHA,
			}

			handler.init(ctx, &repoURL)
			err := handler.fetchGithubAPISbom()
			sboms := handler.sboms

			if tt.wantError && err == nil {
				t.Fatalf("fetchGithubAPISbom() - expected error did not occur")
			}
			if !tt.wantError && err != nil {
				t.Fatalf("fetchGithubAPISbom() - unexpected error occurred: %v", err)
			}
			if !cmp.Equal(sboms, tt.want) {
				t.Errorf("fetchGithubAPISbom() = %v, want %v, diff %v", sboms, tt.want, cmp.Diff(sboms, tt.want))
			}
		})
	}
}

func TestCheckReleaseArtifacts(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name         string
		responsePath string
		want         []clients.Sbom
		wantError    bool
	}{
		{
			name:         "Request Error",
			responsePath: "testdata/invalid-response",
			want:         nil,
			wantError:    true,
		},
		{
			name:         "Zero Length Asset List",
			responsePath: "testdata/empty-asset-response",
			want:         nil,
			wantError:    false,
		},
		{
			name:         "Asset List has no matches",
			responsePath: "testdata/no-asset-matches-response",
			want:         nil,
			wantError:    false,
		},
		{
			name:         "Asset List has matches",
			responsePath: "testdata/asset-matches-response",
			want: []clients.Sbom{
				{
					Name:   "trivy_0.50.1_tar.gz.spdx",
					Origin: "repositoryRelease",
					URL:    "https://github.com/test/releases/download/v0.50.1/trivy_0.50.1_tar.gz.spdx",
					Path:   "https://api.github.com/repos/test/releases/assets/158838506",
				},
				{
					Name:   "trivy_0.50.1_tar.gz.spdx.xml",
					Origin: "repositoryRelease",
					URL:    "https://github.com/test/releases/download/v0.50.1/trivy_0.50.1_tar.gz.spdx.xml",
					Path:   "https://api.github.com/repos/test/releases/assets/158838450",
				},
			},
			wantError: false,
		},
	}
	for _, tt := range testcases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			httpClient := &http.Client{
				Transport: stubTripper{
					responsePath: tt.responsePath,
				},
			}

			client := github.NewClient(httpClient)

			handler := &sbomHandler{
				ghclient: client,
				ctx:      ctx,
			}

			repoURL := repoURL{
				owner:     "ossf-tests",
				repo:      "foo",
				commitSHA: clients.HeadSHA,
			}

			handler.init(ctx, &repoURL)
			err := handler.checkReleaseArtifacts()
			sboms := handler.sboms

			if tt.wantError && err == nil {
				t.Fatalf("checkReleaseArtifacts() - expected error did not occur")
			}
			if !tt.wantError && err != nil {
				t.Fatalf("checkReleaseArtifacts() - unexpected error occurred: %v", err)
			}
			if !cmp.Equal(sboms, tt.want) {
				t.Errorf("checkReleaseArtifacts() = %v, want %v, diff %v", sboms, tt.want, cmp.Diff(sboms, tt.want))
			}
		})
	}
}

func TestCheckCICDArtifacts(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name         string
		responsePath string
		want         []clients.Sbom
		wantError    bool
	}{
		{
			name:         "Request Error",
			responsePath: "testdata/invalid-response",
			want:         nil,
			wantError:    true,
		},
		{
			name:         "Zero Length Artifact List",
			responsePath: "testdata/empty-artifact-response",
			want:         nil,
			wantError:    false,
		},
		{
			name:         "Artifact List has no matches",
			responsePath: "testdata/no-artifact-matches-response",
			want:         nil,
			wantError:    false,
		},
		{
			name:         "Artifact List has matches",
			responsePath: "testdata/artifact-matches-response",
			want: []clients.Sbom{
				{
					Name:   "sbom.spdx.json",
					Origin: "repositoryCICD",
					URL:    "https://api.github.com/repos/test/actions/artifacts/1357350465/zip",
					Path:   "https://api.github.com/repos/test/actions/artifacts/1357350465",
				},
				{
					Name:   "trivy-bom.cdx.xml",
					Origin: "repositoryCICD",
					URL:    "https://api.github.com/repos/test/actions/artifacts/1373743022/zip",
					Path:   "https://api.github.com/repos/test/actions/artifacts/1373743022",
				},
			},
			wantError: false,
		},
	}
	for _, tt := range testcases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			httpClient := &http.Client{
				Transport: stubTripper{
					responsePath: tt.responsePath,
				},
			}

			client := github.NewClient(httpClient)

			handler := &sbomHandler{
				ghclient: client,
				ctx:      ctx,
			}

			repoURL := repoURL{
				owner:     "ossf-tests",
				repo:      "foo",
				commitSHA: clients.HeadSHA,
			}

			handler.init(ctx, &repoURL)
			err := handler.checkCICDArtifacts()
			sboms := handler.sboms

			if tt.wantError && err == nil {
				t.Fatalf("checkCICDArtifacts() - expected error did not occur")
			}
			if !tt.wantError && err != nil {
				t.Fatalf("checkCICDArtifacts() - unexpected error occurred: %v", err)
			}
			if !cmp.Equal(sboms, tt.want) {
				t.Errorf("checkCICDArtifacts() = %v, want %v, diff %v", sboms, tt.want, cmp.Diff(sboms, tt.want))
			}
		})
	}
}
