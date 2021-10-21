// Copyright 2020 Security Scorecard Authors
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

package repos

import (
	"testing"
)

func TestRepoURI_ValidGitHubUrl(t *testing.T) {
	t.Parallel()
	type fields struct {
		host  string
		owner string
		repo  string
	}
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Valid http address",
			fields: fields{
				host:  "github.com",
				owner: "foo",
				repo:  "kubeflow",
			},
			args:    args{s: "https://github.com/foo/kubeflow"},
			wantErr: false,
		},
		{
			name: "Valid http address with trailing slash",
			fields: fields{
				host:  "github.com",
				owner: "foo",
				repo:  "kubeflow",
			},
			args:    args{s: "https://github.com/foo/kubeflow/"},
			wantErr: false,
		},
		{
			name: "Non github repository",
			fields: fields{
				host:  "gitlab.com",
				owner: "foo",
				repo:  "kubeflow",
			},
			args:    args{s: "https://gitlab.com/foo/kubeflow"},
			wantErr: true,
		},
		{
			name: "github repository",
			fields: fields{
				host:  "github.com",
				owner: "foo",
				repo:  "kubeflow",
			},
			args:    args{s: "foo/kubeflow"},
			wantErr: false,
		},
		{
			name: "github repository",
			fields: fields{
				host:  "github.com",
				owner: "foo",
				repo:  "kubeflow",
			},
			args:    args{s: "https://github.com/foo/kubeflow"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &RepoURI{
				url: repoURL{
					host:  tt.fields.host,
					owner: tt.fields.owner,
					repo:  tt.fields.repo,
				},
			}
			t.Log("Test")
			if err := r.Set(tt.args.s); err != nil {
				t.Errorf("repoURI.Set() error = %v", err)
			}
			if err := r.IsValidGitHubURL(); (err != nil) != tt.wantErr {
				t.Errorf("repoURI.ValidGitHubUrl() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if tt.fields.host != r.url.host {
					t.Errorf("repo host expected to be %s but got %s", tt.fields.host, r.url.host)
				}
				if tt.fields.owner != r.url.owner {
					t.Errorf("repo owner expected to be %s but got %s", tt.fields.owner, r.url.owner)
				}
				if tt.fields.repo != r.url.repo {
					t.Errorf("repo expected to be %s but got %s", tt.fields.repo, r.url.repo)
				}
			}
		})
	}
}
