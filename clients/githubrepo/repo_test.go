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
	tests := []struct {
		name     string
		inputURL string
		expected Repo
		wantErr  bool
		ghHost   bool
	}{
		{
			name: "Valid http address",
			expected: Repo{
				host:  "github.com",
				owner: "foo",
				repo:  "kubeflow",
			},
			inputURL: "https://github.com/foo/kubeflow",
			wantErr:  false,
		},
		{
			name: "Valid http address with trailing slash",
			expected: Repo{
				host:  "github.com",
				owner: "foo",
				repo:  "kubeflow",
			},
			inputURL: "https://github.com/foo/kubeflow/",
			wantErr:  false,
		},
		{
			name: "Non GitHub repository",
			expected: Repo{
				host:  "gitlab.com",
				owner: "foo",
				repo:  "kubeflow",
			},
			inputURL: "https://gitlab.com/foo/kubeflow",
			wantErr:  true,
		},
		{
			name: "GitHub repository",
			expected: Repo{
				host:  "github.com",
				owner: "foo",
				repo:  "kubeflow",
			},
			inputURL: "foo/kubeflow",
			wantErr:  false,
		},
		{
			name: "GitHub repository with host",
			expected: Repo{
				host:  "github.com",
				owner: "foo",
				repo:  "kubeflow",
			},
			inputURL: "https://github.com/foo/kubeflow",
			wantErr:  false,
		},
		{
			name: "Enterprise github repository with host",
			expected: Repo{
				host:  "github.corp.com",
				owner: "corpfoo",
				repo:  "kubeflow",
			},
			inputURL: "https://github.corp.com/corpfoo/kubeflow",
			wantErr:  false,
			ghHost:   true,
		},
		{
			name: "Enterprise github repository",
			expected: Repo{
				host:  "github.corp.com",
				owner: "corpfoo",
				repo:  "kubeflow",
			},
			inputURL: "corpfoo/kubeflow",
			wantErr:  false,
			ghHost:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ghHost {
				t.Setenv("GH_HOST", "github.corp.com")
			}

			r := Repo{
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
			if !tt.wantErr && !cmp.Equal(tt.expected, r, cmp.AllowUnexported(Repo{})) {
				t.Errorf("Got diff: %s", cmp.Diff(tt.expected, r))
			}
			if !cmp.Equal(r.Host(), tt.expected.host) {
				t.Errorf("%s expected host: %s got host %s", tt.inputURL, tt.expected.host, r.Host())
			}
		})
	}
}
