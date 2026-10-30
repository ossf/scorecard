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

package githubrepo

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v82/github"

	"github.com/ossf/scorecard/v5/clients"
)

type releasesRoundTripper func(*http.Request) (*http.Response, error)

func (f releasesRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func Test_releasesFrom(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data []*github.RepositoryRelease
		want []clients.Release
	}{
		{
			name: "assets with browser download URLs",
			data: []*github.RepositoryRelease{
				{
					TagName:         asPtr("v1.0.0"),
					URL:             asPtr("https://api.github.com/repos/owner/repo/releases/1"),
					HTMLURL:         asPtr("https://github.com/owner/repo/releases/tag/v1.0.0"),
					TargetCommitish: asPtr("main"),
					Assets: []*github.ReleaseAsset{
						{
							Name:               asPtr("binary.tar.gz"),
							BrowserDownloadURL: asPtr("https://github.com/owner/repo/releases/download/v1.0.0/binary.tar.gz"),
							URL:                asPtr("https://api.github.com/repos/owner/repo/releases/assets/101"),
						},
						{
							Name:               asPtr("binary.tar.gz.sig"),
							BrowserDownloadURL: asPtr("https://github.com/owner/repo/releases/download/v1.0.0/binary.tar.gz.sig"),
							URL:                asPtr("https://api.github.com/repos/owner/repo/releases/assets/102"),
						},
					},
				},
			},
			want: []clients.Release{
				{
					TagName:         "v1.0.0",
					URL:             "https://api.github.com/repos/owner/repo/releases/1",
					TargetCommitish: "main",
					Assets: []clients.ReleaseAsset{
						{
							Name: "binary.tar.gz",
							URL:  "https://github.com/owner/repo/releases/download/v1.0.0/binary.tar.gz",
						},
						{
							Name: "binary.tar.gz.sig",
							URL:  "https://github.com/owner/repo/releases/download/v1.0.0/binary.tar.gz.sig",
						},
					},
				},
			},
		},
		{
			name: "fallback to asset URL when browser download URL is empty",
			data: []*github.RepositoryRelease{
				{
					TagName:         asPtr("v2.0.0"),
					URL:             asPtr("https://api.github.com/repos/owner/repo/releases/2"),
					HTMLURL:         asPtr("https://github.com/owner/repo/releases/tag/v2.0.0"),
					TargetCommitish: asPtr("main"),
					Assets: []*github.ReleaseAsset{
						{
							Name:               asPtr("artifact.zip"),
							BrowserDownloadURL: asPtr(""),
							URL:                asPtr("https://api.github.com/repos/owner/repo/releases/assets/201"),
						},
					},
				},
			},
			want: []clients.Release{
				{
					TagName:         "v2.0.0",
					URL:             "https://api.github.com/repos/owner/repo/releases/2",
					TargetCommitish: "main",
					Assets: []clients.ReleaseAsset{
						{
							Name: "artifact.zip",
							URL:  "https://api.github.com/repos/owner/repo/releases/assets/201",
						},
					},
				},
			},
		},
		{
			name: "release without assets",
			data: []*github.RepositoryRelease{
				{
					TagName:         asPtr("v3.0.0"),
					URL:             asPtr("https://api.github.com/repos/owner/repo/releases/3"),
					HTMLURL:         asPtr("https://github.com/owner/repo/releases/tag/v3.0.0"),
					TargetCommitish: asPtr("main"),
					Assets:          nil,
				},
			},
			want: []clients.Release{
				{
					TagName:         "v3.0.0",
					URL:             "https://api.github.com/repos/owner/repo/releases/3",
					TargetCommitish: "main",
					Assets:          nil,
				},
			},
		},
		{
			name: "nil release and nil asset are ignored gracefully",
			data: []*github.RepositoryRelease{
				nil,
				{
					TagName: asPtr("v4.0.0"),
					Assets: []*github.ReleaseAsset{
						nil,
						{
							Name:               asPtr("bundle.tar.gz"),
							BrowserDownloadURL: asPtr("https://github.com/owner/repo/releases/download/v4.0.0/bundle.tar.gz"),
						},
					},
				},
			},
			want: []clients.Release{
				{
					TagName: "v4.0.0",
					Assets: []clients.ReleaseAsset{
						{
							Name: "bundle.tar.gz",
							URL:  "https://github.com/owner/repo/releases/download/v4.0.0/bundle.tar.gz",
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
			if !cmp.Equal(got, tt.want) {
				t.Errorf("releasesFrom() diff: %v", cmp.Diff(got, tt.want))
			}
		})
	}
}

func Test_releasesHandler_getReleases(t *testing.T) {
	t.Parallel()

	t.Run("non-HEAD query returns ErrUnsupportedFeature", func(t *testing.T) {
		t.Parallel()

		handler := &releasesHandler{}
		repoURL := &Repo{
			owner:     "ossf",
			repo:      "scorecard",
			commitSHA: "1234567890abcdef",
		}
		handler.init(t.Context(), repoURL)

		releases, err := handler.getReleases()
		if !errors.Is(err, clients.ErrUnsupportedFeature) {
			t.Errorf("expected ErrUnsupportedFeature, got %v", err)
		}
		if releases != nil {
			t.Errorf("expected nil releases, got %v", releases)
		}
	})

	t.Run("error from GitHub API propagates", func(t *testing.T) {
		t.Parallel()

		httpClient := &http.Client{
			Transport: releasesRoundTripper(func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					Status:     "500 Internal Server Error",
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(strings.NewReader(`{"message":"server error"}`)),
				}, nil
			}),
		}
		client := github.NewClient(httpClient)
		handler := &releasesHandler{client: client}
		repoURL := &Repo{
			owner:     "ossf",
			repo:      "scorecard",
			commitSHA: clients.HeadSHA,
		}
		handler.init(t.Context(), repoURL)

		releases, err := handler.getReleases()
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if releases != nil {
			t.Errorf("expected nil releases, got %v", releases)
		}
	})

	t.Run("successful release fetching and asset parsing", func(t *testing.T) {
		t.Parallel()

		responseJSON := `[
  {
    "tag_name": "v1.0.0",
    "url": "https://api.github.com/repos/ossf/scorecard/releases/1",
    "html_url": "https://github.com/ossf/scorecard/releases/tag/v1.0.0",
    "target_commitish": "main",
    "assets": [
      {
        "name": "scorecard.tar.gz",
        "browser_download_url": "https://github.com/ossf/scorecard/releases/download/v1.0.0/scorecard.tar.gz",
        "url": "https://api.github.com/repos/ossf/scorecard/releases/assets/101"
      }
    ]
  }
]`

		httpClient := &http.Client{
			Transport: releasesRoundTripper(func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					Status:     "200 OK",
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(responseJSON)),
				}, nil
			}),
		}
		client := github.NewClient(httpClient)
		handler := &releasesHandler{client: client}
		repoURL := &Repo{
			owner:     "ossf",
			repo:      "scorecard",
			commitSHA: clients.HeadSHA,
		}
		handler.init(t.Context(), repoURL)

		releases, err := handler.getReleases()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(releases) != 1 {
			t.Fatalf("expected 1 release, got %d", len(releases))
		}
		if len(releases[0].Assets) != 1 {
			t.Fatalf("expected 1 asset, got %d", len(releases[0].Assets))
		}
		wantAssetURL := "https://github.com/ossf/scorecard/releases/download/v1.0.0/scorecard.tar.gz"
		if releases[0].Assets[0].URL != wantAssetURL {
			t.Errorf("got asset URL %q, want %q", releases[0].Assets[0].URL, wantAssetURL)
		}
	})
}
