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

func TestListSBOMs(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name         string
		responsePath string
		want         []clients.SBOM
		wantError    bool
	}{
		{
			name:         "Request Error",
			responsePath: "testdata/invalid-response",
			want:         nil,
			wantError:    true,
		},
		{
			name:         "Artifact List has matches",
			responsePath: "testdata/artifact-matches-response",
			want: []clients.SBOM{
				{
					Name: "sbom.spdx.json",
					URL:  "https://api.github.com/repos/test/actions/artifacts/1357350465/zip",
					Path: "https://api.github.com/repos/test/actions/artifacts/1357350465",
				},
				{
					Name: "trivy-bom.cdx.xml",
					URL:  "https://api.github.com/repos/test/actions/artifacts/1373743022/zip",
					Path: "https://api.github.com/repos/test/actions/artifacts/1373743022",
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

			handler := &SBOMHandler{
				ghclient: client,
				ctx:      ctx,
			}

			repoURL := repoURL{
				owner:     "ossf-tests",
				repo:      "foo",
				commitSHA: clients.HeadSHA,
			}

			handler.init(ctx, &repoURL)
			SBOMs, err := handler.listSBOMs()

			if tt.wantError && err == nil {
				t.Fatalf("listSBOMs() - expected error did not occur")
			}
			if !tt.wantError && err != nil {
				t.Fatalf("listSBOMs() - unexpected error occurred: %v", err)
			}
			if !cmp.Equal(SBOMs, tt.want) {
				t.Errorf("listSBOMs() = %v, want %v, diff %v", SBOMs, tt.want, cmp.Diff(SBOMs, tt.want))
			}
		})
	}
}

func TestCheckCICDArtifacts(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name         string
		responsePath string
		want         []clients.SBOM
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
			want: []clients.SBOM{
				{
					Name: "sbom.spdx.json",
					URL:  "https://api.github.com/repos/test/actions/artifacts/1357350465/zip",
					Path: "https://api.github.com/repos/test/actions/artifacts/1357350465",
				},
				{
					Name: "trivy-bom.cdx.xml",
					URL:  "https://api.github.com/repos/test/actions/artifacts/1373743022/zip",
					Path: "https://api.github.com/repos/test/actions/artifacts/1373743022",
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

			handler := &SBOMHandler{
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
			SBOMs := handler.SBOMs

			if tt.wantError && err == nil {
				t.Fatalf("checkCICDArtifacts() - expected error did not occur")
			}
			if !tt.wantError && err != nil {
				t.Fatalf("checkCICDArtifacts() - unexpected error occurred: %v", err)
			}
			if !cmp.Equal(SBOMs, tt.want) {
				t.Errorf("checkCICDArtifacts() = %v, want %v, diff %v", SBOMs, tt.want, cmp.Diff(SBOMs, tt.want))
			}
		})
	}
}
