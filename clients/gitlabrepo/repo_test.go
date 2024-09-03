// Copyright 2022 OpenSSF Scorecard Authors
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
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestRepoURL_IsValid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		inputURL     string
		expected     Repo
		wantErr      bool
		flagRequired bool
	}{
		{
			name: "github repository",
			expected: Repo{
				scheme:  "https",
				host:    "https://github.com",
				owner:   "ossf",
				project: "scorecard",
			},
			inputURL: "https://github.com/ossf/scorecard",
			wantErr:  true,
		},
		{
			name: "GitHub project with 'gitlab.' in the title",
			expected: Repo{
				scheme:  "http",
				host:    "github.com",
				owner:   "foo",
				project: "gitlab.test",
			},
			inputURL: "http://github.com/foo/gitlab.test",
			wantErr:  true,
		},
		{
			name: "valid gitlab project",
			expected: Repo{
				host:    "gitlab.com",
				owner:   "ossf-test",
				project: "scorecard-check-binary-artifacts-e2e",
			},
			inputURL: "https://gitlab.com/ossf-test/scorecard-check-binary-artifacts-e2e",
			wantErr:  false,
		},
		{
			name: "valid gitlab project",
			expected: Repo{
				host:    "gitlab.com",
				owner:   "ossf-test",
				project: "scorecard-check-binary-artifacts-e2e",
			},
			inputURL: "https://gitlab.com/ossf-test/scorecard-check-binary-artifacts-e2e",
			wantErr:  false,
		},
		{
			name: "valid https address with trailing slash",
			expected: Repo{
				scheme:  "https",
				host:    "gitlab.haskell.org",
				owner:   "haskell",
				project: "filepath",
			},
			inputURL: "https://gitlab.haskell.org/haskell/filepath",
			wantErr:  false,
		},
		{
			name: "valid hosted gitlab project",
			expected: Repo{
				scheme:  "https",
				host:    "salsa.debian.org",
				owner:   "webmaster-team",
				project: "webml",
			},
			inputURL:     "https://salsa.debian.org/webmaster-team/webwml",
			wantErr:      false,
			flagRequired: true,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure blow
		if tt.flagRequired && os.Getenv("TEST_GITLAB_EXTERNAL") == "" {
			continue
		}
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := Repo{
				host:    tt.expected.host,
				owner:   tt.expected.owner,
				project: tt.expected.project,
			}
			if err := r.parse(tt.inputURL); err != nil {
				t.Errorf("repoURL.parse() error = %v", err)
			}
			if err := r.IsValid(); (err != nil) != tt.wantErr {
				t.Errorf("repoURL.IsValid() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			t.Log(r.URI())
			if !tt.wantErr && !cmp.Equal(tt.expected, r, cmpopts.IgnoreUnexported(Repo{})) {
				t.Logf("expected: %s GOT: %s", tt.expected.host, r.host)
				t.Logf("expected: %s GOT: %s", tt.expected.owner, r.owner)
				t.Logf("expected: %s GOT: %s", tt.expected.project, r.project)
				t.Errorf("Got diff: %s", cmp.Diff(tt.expected, r))
			}
			if !cmp.Equal(r.Host(), tt.expected.host) {
				t.Errorf("%s expected host: %s got host %s", tt.inputURL, tt.expected.host, r.Host())
			}
		})
	}
}

func TestRepoURL_MakeGitLabRepo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		repouri      string
		expected     bool
		flagRequired bool
	}{
		{
			repouri:  "github.com/ossf/scorecard",
			expected: false,
		},
		{
			repouri:  "ossf/scorecard",
			expected: false,
		},
		{
			repouri:  "https://github.com/ossf/scorecard",
			expected: false,
		},
		{
			repouri:  "gitlab.com/gitlab-org/gitlab",
			expected: true,
		},
		{
			repouri:      "https://salsa.debian.org/webmaster-team/webml",
			expected:     true,
			flagRequired: true,
		},
		{
			// Invalid repo
			repouri:  "https://abcdef.foo.txt/nilproj/nilrepo",
			expected: false,
		},
	}

	for _, tt := range tests {
		if tt.flagRequired && os.Getenv("TEST_GITLAB_EXTERNAL") == "" {
			continue
		}
		g, err := MakeGitlabRepo(tt.repouri)
		if (g != nil) != (err == nil) {
			t.Errorf("got gitlabrepo: %s with err %s", g, err)
		}
		isGitlab := g != nil && err == nil
		if isGitlab != tt.expected {
			t.Errorf("got %s isgitlab: %t expected %t", tt.repouri, isGitlab, tt.expected)
		}
	}
}

func TestRepoURL_parse_GL_HOST(t *testing.T) {
	tests := []struct {
		name                 string
		url                  string
		host, owner, project string
		glHost               string
		wantErr              bool
	}{
		{
			name:    "GL_HOST ends with slash",
			glHost:  "https://foo.com/gitlab/",
			url:     "foo.com/gitlab/ssdlc/scorecard-scanner",
			host:    "foo.com/gitlab",
			owner:   "ssdlc",
			project: "scorecard-scanner",
		},
		{
			name:    "GL_HOST doesn't end with slash",
			glHost:  "https://foo.com/gitlab",
			url:     "foo.com/gitlab/ssdlc/scorecard-scanner",
			host:    "foo.com/gitlab",
			owner:   "ssdlc",
			project: "scorecard-scanner",
		},
		{
			name:    "GL_HOST doesn't match url",
			glHost:  "https://foo.com/gitlab",
			url:     "bar.com/gitlab/ssdlc/scorecard-scanner",
			host:    "bar.com",
			owner:   "gitlab",
			project: "ssdlc/scorecard-scanner",
		},
		{
			name:    "GL_HOST has no path component",
			glHost:  "https://foo.com",
			url:     "foo.com/ssdlc/scorecard-scanner",
			host:    "foo.com",
			owner:   "ssdlc",
			project: "scorecard-scanner",
		},
		{
			name:    "GL_HOST path has multiple slashes",
			glHost:  "https://foo.com/bar/baz/",
			url:     "foo.com/bar/baz/ssdlc/scorecard-scanner",
			host:    "foo.com/bar/baz",
			owner:   "ssdlc",
			project: "scorecard-scanner",
		},
		{
			name:    "GL_HOST has no scheme",
			glHost:  "foo.com/bar/",
			url:     "foo.com/bar/ssdlc/scorecard-scanner",
			host:    "foo.com/bar",
			owner:   "ssdlc",
			project: "scorecard-scanner",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GL_HOST", tt.glHost)
			var r Repo
			err := r.parse(tt.url)
			if (err != nil) != tt.wantErr {
				t.Fatalf("wanted err: %t, got: %v", tt.wantErr, err)
			}
			if r.host != tt.host {
				t.Errorf("got host: %s, want: %s", r.host, tt.host)
			}
			if r.owner != tt.owner {
				t.Errorf("got owner: %s, want: %s", r.owner, tt.owner)
			}
			if r.project != tt.project {
				t.Errorf("got project: %s, want: %s", r.project, tt.project)
			}
		})
	}
}
