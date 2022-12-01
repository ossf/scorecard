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
			name: "valid http address",
			expected: repoURL{
				hostname:  "gitlab.example.com",
				owner:     "foo",
				projectID: "1234",
			},
			inputURL: "http://gitlab.example.com/foo/1234",
			wantErr:  false,
		},
		{
			name: "valid https address",
			expected: repoURL{
				hostname:  "gitlab.example.com",
				owner:     "foo",
				projectID: "1234",
			},
			inputURL: "https://gitlab.example.com/foo/1234",
			wantErr:  false,
		},
		{
			name: "valid http address with trailing slash",
			expected: repoURL{
				hostname:  "gitlab.example.com",
				owner:     "foo",
				projectID: "1234",
			},
			inputURL: "http://gitlab.example.com/foo/1234/",
			wantErr:  false,
		},
		{
			name: "valid https address with trailing slash",
			expected: repoURL{
				hostname:  "gitlab.example.com",
				owner:     "foo",
				projectID: "1234",
			},
			inputURL: "https://gitlab.example.com/foo/1234/",
			wantErr:  false,
		},
		{
			name: "non gitlab repository",
			expected: repoURL{
				hostname:  "github.com",
				owner:     "foo",
				projectID: "1234",
			},
			inputURL: "https://github.com/foo/1234",
			wantErr:  true,
		},
		{
			name: "GitLab project with wrong projectID",
			expected: repoURL{
				hostname:  "gitlab.example.com",
				owner:     "foo",
				projectID: "bar",
			},
			inputURL: "https://gitlab.example.com/foo/bar",
			wantErr:  true,
		},
		{
			name: "GitHub project with 'gitlab.' in the title",
			expected: repoURL{
				hostname:  "github.com",
				owner:     "foo",
				projectID: "gitlab.test",
			},
			inputURL: "http://github.com/foo/gitlab.test",
			wantErr:  true,
		},
		{
			name: "valid gitlab project without http or https",
			expected: repoURL{
				hostname:  "gitlab.example.com",
				owner:     "foo",
				projectID: "1234",
			},
			inputURL: "gitlab.example.com/foo/1234",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure blow
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := repoURL{
				hostname:  tt.expected.hostname,
				owner:     tt.expected.owner,
				projectID: tt.expected.projectID,
			}
			if err := r.parse(tt.inputURL); err != nil {
				t.Errorf("repoURL.parse() error = %v", err)
			}
			if err := r.IsValid(); (err != nil) != tt.wantErr {
				t.Errorf("repoURL.IsValid() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !cmp.Equal(tt.expected, r, cmp.AllowUnexported(repoURL{})) {
				fmt.Println("expected: " + tt.expected.hostname + " GOT: " + r.hostname)
				fmt.Println("expected: " + tt.expected.owner + " GOT: " + r.owner)
				fmt.Println("expected: " + tt.expected.projectID + " GOT: " + r.projectID)
				t.Errorf("Got diff: %s", cmp.Diff(tt.expected, r))
			}
		})
	}
}
