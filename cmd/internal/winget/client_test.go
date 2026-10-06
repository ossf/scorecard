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

package winget

import (
	"bytes"
	"cmp"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	pmc "github.com/ossf/scorecard/v5/cmd/internal/packagemanager"
)

func TestRepoFromURL(t *testing.T) {
	t.Parallel()
	tests := []struct{ url, want string }{
		{"https://github.com/jqlang/jq/blob/master/COPYING", "https://github.com/jqlang/jq"},
		{"https://github.com/sharkdp/bat#license", "https://github.com/sharkdp/bat"},
		{"https://github.com/JanDeDobbeleer", ""},
		{"https://gitlab.com/inkscape/inkscape", "https://gitlab.com/inkscape/inkscape"},
		{"https://code.videolan.org/videolan/vlc/-/blob/HEAD/COPYING", "https://code.videolan.org/videolan/vlc"},
		{"https://gitlab.gnome.org/GNOME/gtk/-/issues", "https://gitlab.gnome.org/gnome/gtk"},
		{"https://gitlab.com/gitlab-org/cli/-/releases/v1.120.0/downloads/glab.exe", "https://gitlab.com/gitlab-org/cli"},
		{"https://gitlab.com/api/v4/projects/4207231/packages/generic/graphviz-releases/16.1.0/g.exe", ""},
		{"https://gitlab.com/api/v4/projects/gitlab-org%2Fcli/packages/generic/glab/1.0/glab.exe", "https://gitlab.com/gitlab-org/cli"},
		{"https://gitlab.com/-/project/4207231/uploads/abc/file.exe", ""},
		{"https://sharkdp.github.io/hyperfine/", "https://github.com/sharkdp/hyperfine"},
		{"https://sharkdp.github.io/", ""},
		{"https://dev.azure.com/org/project/_git/repo?path=/README.md", "https://dev.azure.com/org/project/_git/repo"},
		{"https://dev.azure.com/org/project", ""},
		{"https://www.python.org/ftp/python/3.11.9/python-3.11.9.exe", ""},
	}
	for _, tt := range tests {
		if got := repoFromURL(tt.url); got != tt.want {
			t.Errorf("repoFromURL(%q) = %q, want %q", tt.url, got, tt.want)
		}
	}
}

func TestGitRepositoryByPackageName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, id      string
		dirs          string // space-separated directories in the package folder
		version       string // manifests are only served for this version
		defaultLocale string
		locale        string
		installer     string
		want          string
		wantErr       bool
	}{
		{
			name: "repo link in locale manifest", id: "Notepad++.Notepad++", dirs: "8.9.6.2 8.9.6.4", version: "8.9.6.4",
			locale: "LicenseUrl: https://github.com/notepad-plus-plus/notepad-plus-plus/blob/HEAD/LICENSE",
			want:   "https://github.com/notepad-plus-plus/notepad-plus-plus",
		},
		{
			name: "every dot in the ID is a directory", id: "BurntSushi.ripgrep.MSVC", dirs: "15.2.0", version: "15.2.0",
			locale: "PackageUrl: https://github.com/BurntSushi/ripgrep",
			want:   "https://github.com/burntsushi/ripgrep",
		},
		{
			name: "newest numeric version, sub-packages ignored", id: "Foo.Bar", dirs: "10.0 9.0 Preview zh-TW", version: "10.0",
			locale: "PackageUrl: https://github.com/foo/bar",
			want:   "https://github.com/foo/bar",
		},
		{
			name: "most linked repo wins", id: "Git.Git", dirs: "2.55.0", version: "2.55.0",
			locale: "LicenseUrl: https://github.com/git-for-windows/build-extra/blob/HEAD/LICENSE.txt\n" +
				"PublisherSupportUrl: https://github.com/git-for-windows/git/issues\n" +
				"ReleaseNotesUrl: https://github.com/git-for-windows/git/releases/tag/v2.55.0.windows.1",
			want: "https://github.com/git-for-windows/git",
		},
		{
			name: "gitlab package registry files do not outvote the project link", id: "Graphviz.Graphviz",
			dirs: "16.1.0", version: "16.1.0",
			locale: "ReleaseNotesUrl: https://gitlab.com/graphviz/graphviz/-/releases/16.1.0",
			installer: "- InstallerUrl: https://gitlab.com/api/v4/projects/4207231/packages/generic/graphviz-releases/16.1.0/a.exe\n" +
				"- InstallerUrl: https://gitlab.com/api/v4/projects/4207231/packages/generic/graphviz-releases/16.1.0/b.exe",
			want: "https://gitlab.com/graphviz/graphviz",
		},
		{
			name: "release link in installer manifest", id: "Rustlang.Rustup", dirs: "1.29.1", version: "1.29.1",
			locale:    "PackageUrl: https://rustup.rs/",
			installer: "- InstallerUrl: https://github.com/rust-lang/rustup/releases/download/1.29.1/rustup-init.exe",
			want:      "https://github.com/rust-lang/rustup",
		},
		{
			name: "default locale other than en-US", id: "Foo.Bar", dirs: "1.0", version: "1.0", defaultLocale: "zh-CN",
			locale: "PackageUrl: https://github.com/foo/bar",
			want:   "https://github.com/foo/bar",
		},
		{
			name: "no repo link", id: "Python.Python.3.11", dirs: "3.11.9", version: "3.11.9",
			locale: "PackageUrl: https://www.python.org/", wantErr: true,
		},
		{
			name: "v-prefixed versions", id: "Pantelis23.MLRift", dirs: "v1.1.0 v1.10.0 v1.9.0", version: "v1.10.0",
			locale: "PackageUrl: https://github.com/Pantelis23/MLRift",
			want:   "https://github.com/pantelis23/mlrift",
		},
		{
			name: "non-version directory when it is the only one", id: "KDE.Falkon.Nightly", dirs: "master", version: "master",
			locale: "LicenseUrl: https://invent.kde.org/network/falkon/-/blob/master/COPYING",
			want:   "https://invent.kde.org/network/falkon",
		},
		{
			name: "one installer per architecture does not outvote the package page", id: "PromptExecution.ledgrrr",
			dirs: "1.9.0", version: "1.9.0",
			locale: "PackageUrl: https://github.com/PromptExecution/ledgrrr\n" +
				"PublisherSupportUrl: https://github.com/PromptExecution/ledgrrr/issues",
			installer: "- InstallerUrl: https://github.com/elasticdotventures/_b00t_/releases/download/v1.9.0/x64.msi\n" +
				"- InstallerUrl: https://github.com/elasticdotventures/_b00t_/releases/download/v1.9.0/arm64.msi\n" +
				"- InstallerUrl: https://github.com/elasticdotventures/_b00t_/releases/download/v1.9.0/x86.msi",
			want: "https://github.com/promptexecution/ledgrrr",
		},
		{name: "manifests missing", id: "Foo.Bar", dirs: "1.0", wantErr: true},
		{name: "only sub-packages", id: "Foo.Bar", dirs: "Preview", wantErr: true},
		{name: "package not found", id: "XAMPP.XAMPP", wantErr: true},
		{name: "no dot", id: "NotepadPlusPlus", wantErr: true},
		{name: "empty publisher", id: ".Notepad++", wantErr: true},
		{name: "empty name", id: "Notepad++.", wantErr: true},
		{name: "empty segment", id: "Foo..Bar", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pkgPath := strings.ToLower(tt.id[:1]) + "/" + strings.ReplaceAll(tt.id, ".", "/")
			var entries []string
			for _, d := range strings.Fields(tt.dirs) {
				entries = append(entries, fmt.Sprintf(`{"name":%q,"type":"dir"}`, d))
			}
			locale := cmp.Or(tt.defaultLocale, "en-US")
			prefix := pkgPath + "/" + tt.version + "/" + tt.id
			files := map[string]string{
				contentsURL + pkgPath:                           "[" + strings.Join(entries, ",") + "]",
				rawURL + prefix + ".yaml":                       "DefaultLocale: " + locale,
				rawURL + prefix + ".locale." + locale + ".yaml": tt.locale,
				rawURL + prefix + ".installer.yaml":             tt.installer,
			}
			ctrl := gomock.NewController(t)
			p := pmc.NewMockClient(ctrl)
			p.EXPECT().GetURI(gomock.Any()).
				DoAndReturn(func(uri string) (*http.Response, error) {
					body, ok := files[uri]
					if !ok || body == "" || body == "[]" {
						return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(&bytes.Buffer{})}, nil
					}
					return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString(body))}, nil
				}).AnyTimes()
			client := WingetClient{Manager: p}
			got, err := client.GitRepositoryByPackageName(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitRepositoryByPackageName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GitRepositoryByPackageName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		a, b string
		sign int
	}{
		{"10.0", "9.0", 1},
		{"1.2.3", "1.2.3", 0},
		{"1.2", "1.2.1", -1},
		{"8.9.6.4", "8.9.6.2", 1},
		{"1.0.0-beta", "1.0.0-alpha", 1},
		{"v1.10.0", "v1.9.0", 1},
		{"v2.0", "1.9", 1},
	}
	for _, tt := range tests {
		got := compareVersions(tt.a, tt.b)
		if (got > 0) != (tt.sign > 0) || (got < 0) != (tt.sign < 0) {
			t.Errorf("compareVersions(%q, %q) = %d, want sign %d", tt.a, tt.b, got, tt.sign)
		}
	}
}
