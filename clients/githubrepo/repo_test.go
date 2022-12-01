// Copyright 2020 OpenSSF Scorecard Authors
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
			name: "Valid http address",
			expected: repoURL{
				host:  "github.com",
				owner: "foo",
				repo:  "kubeflow",
			},
			inputURL: "https://github.com/foo/kubeflow",
			wantErr:  false,
		},
		{
			name: "Valid http address with trailing slash",
			expected: repoURL{
				host:  "github.com",
				owner: "foo",
				repo:  "kubeflow",
			},
			inputURL: "https://github.com/foo/kubeflow/",
			wantErr:  false,
		},
		{
			name: "Non github repository",
			expected: repoURL{
				host:  "gitlab.com",
				owner: "foo",
				repo:  "kubeflow",
			},
			inputURL: "https://gitlab.com/foo/kubeflow",
			wantErr:  true,
		},
		{
			name: "github repository",
			expected: repoURL{
				host:  "github.com",
				owner: "foo",
				repo:  "kubeflow",
			},
			inputURL: "foo/kubeflow",
			wantErr:  false,
		},
		{
			name: "github repository",
			expected: repoURL{
				host:  "github.com",
				owner: "foo",
				repo:  "kubeflow",
			},
			inputURL: "https://github.com/foo/kubeflow",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := repoURL{
				host:  tt.expected.host,
				owner: tt.expected.owner,
				repo:  tt.expected.repo,
			}
			if err := r.parse(tt.inputURL); err != nil {
				t.Errorf("repoURL.parse() error = %v", err)
			}
			if err := r.IsValid(); (err != nil) != tt.wantErr {
				t.Errorf("repoURL.IsValid() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !cmp.Equal(tt.expected, r, cmp.AllowUnexported(repoURL{})) {
				t.Errorf("Got diff: %s", cmp.Diff(tt.expected, r))
			}
		})
	}
}
