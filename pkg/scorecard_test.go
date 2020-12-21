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

package pkg

import (
	"testing"
)

func TestRepoURL_Set(t *testing.T) {
	type fields struct {
		Host  string
		Owner string
		Repo  string
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
				Host:  "github.com",
				Owner: "kubeflow",
				Repo:  "kubeflow",
			},
			args:    args{s: "https://github.com/kubeflow/kubeflow"},
			wantErr: false,
		},
		{
			name: "Valid http address with trailing slash",
			fields: fields{
				Host:  "github.com",
				Owner: "kubeflow",
				Repo:  "kubeflow",
			},
			args:    args{s: "https://github.com/kubeflow/kubeflow/"},
			wantErr: false,
		},
		{
			name: "Non github repository",
			fields: fields{
				Host:  "gitlab.com",
				Owner: "kubeflow",
				Repo:  "kubeflow",
			},
			args:    args{s: "https://gitlab.com/kubeflow/kubeflow"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RepoURL{
				Host:  tt.fields.Host,
				Owner: tt.fields.Owner,
				Repo:  tt.fields.Repo,
			}
			if err := r.Set(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("RepoURL.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
