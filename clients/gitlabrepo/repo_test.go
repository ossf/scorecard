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
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestRepoURL_IsValid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		inputURL string
		expected repoURL
		wantErr  bool
	}{
		{
			name: "github repository",
			expected: repoURL{
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
			expected: repoURL{
				scheme:  "http",
				host:    "http://github.com",
				owner:   "foo",
				project: "gitlab.test",
			},
			inputURL: "http://github.com/foo/gitlab.test",
			wantErr:  true,
		},
		{
			name: "valid gitlab project",
			expected: repoURL{
				host:    "https://gitlab.com",
				owner:   "ossf-test",
				project: "scorecard-check-binary-artifacts-e2e",
			},
			inputURL: "https://gitlab.com/ossf-test/scorecard-check-binary-artifacts-e2e",
			wantErr:  false,
		},
		{
			name: "valid https address with trailing slash",
			expected: repoURL{
				scheme:  "https",
				host:    "https://gitlab.com",
				owner:   "ossf-test",
				project: "scorecard-check-binary-artifacts-e2e",
			},
			inputURL: "https://gitlab.com/ossf-test/scorecard-check-binary-artifacts-e2e/",
			wantErr:  false,
		},

		{
			name: "valid hosted gitlab project",
			expected: repoURL{
				scheme:  "https",
				host:    "https://salsa.debian.org",
				owner:   "webmaster-team",
				project: "webml",
			},
			inputURL: "https://salsa.debian.org/webmaster-team/webwml",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure blow
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := repoURL{
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
			if !tt.wantErr && !cmp.Equal(tt.expected, r, cmpopts.IgnoreUnexported(repoURL{})) {
				fmt.Println("expected: " + tt.expected.host + " GOT: " + r.host)
				fmt.Println("expected: " + tt.expected.owner + " GOT: " + r.owner)
				fmt.Println("expected: " + tt.expected.project + " GOT: " + r.project)
				t.Errorf("Got diff: %s", cmp.Diff(tt.expected, r))
			}
			if !cmp.Equal(r.Host(), tt.expected.host) {
				t.Errorf("%s expected host: %s got host %s", tt.inputURL, tt.expected.host, r.Host())
			}
		})
	}
}

func TestRepoURL_DetectGitlab(t *testing.T) {
	tests := []struct {
		repouri  string
		expected bool
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
			repouri:  "https://salsa.debian.org/webmaster-team/webml",
			expected: true,
		},
		{
			// Invalid repo
			repouri:  "https://abcdef.foo.txt/nilproj/nilrepo",
			expected: false,
		},
	}

	for _, tt := range tests {
		g := DetectGitLab(tt.repouri)
		if g != tt.expected {
			t.Errorf("got %s isgitlab: %t expected %t", tt.repouri, g, tt.expected)
		}
	}
}
