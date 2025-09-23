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

package azuredevopsrepo

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"

	"github.com/ossf/scorecard/v5/clients"
)

func Test_listStatuses(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		ref         string
		getRefs     fnGetRefs
		getStatuses fnGetStatuses
		want        []clients.Status
		wantErr     bool
	}{
		{
			name: "sha - no statuses",
			ref:  "4b825dc642cb6eb9a060e54bf8d69288fbee4904",
			getStatuses: func(ctx context.Context, args git.GetStatusesArgs) (*[]git.GitStatus, error) {
				return &[]git.GitStatus{}, nil
			},
			want:    []clients.Status{},
			wantErr: false,
		},
		{
			name: "sha - single status",
			ref:  "4b825dc642cb6eb9a060e54bf8d69288fbee4904",
			getStatuses: func(ctx context.Context, args git.GetStatusesArgs) (*[]git.GitStatus, error) {
				return &[]git.GitStatus{
					{
						Context: &git.GitStatusContext{
							Name: toPtr("test"),
						},
						State:     toPtr(git.GitStatusStateValues.Succeeded),
						TargetUrl: toPtr("https://example.com"),
					},
				}, nil
			},
			want: []clients.Status{
				{
					Context:   "test",
					State:     "success",
					TargetURL: "https://example.com",
					URL:       "https://example.com",
				},
			},
			wantErr: false,
		},
		{
			name: "main - no statuses",
			ref:  "main",
			getRefs: func(ctx context.Context, args git.GetRefsArgs) (*git.GetRefsResponseValue, error) {
				return &git.GetRefsResponseValue{
					Value: []git.GitRef{
						{
							Statuses: &[]git.GitStatus{},
						},
					},
				}, nil
			},
			want:    []clients.Status{},
			wantErr: false,
		},
		{
			name: "main - single status",
			ref:  "main",
			getRefs: func(ctx context.Context, args git.GetRefsArgs) (*git.GetRefsResponseValue, error) {
				return &git.GetRefsResponseValue{
					Value: []git.GitRef{
						{
							Statuses: &[]git.GitStatus{
								{
									Context: &git.GitStatusContext{
										Name: toPtr("test"),
									},
									State:     toPtr(git.GitStatusStateValues.Succeeded),
									TargetUrl: toPtr("https://example.com"),
								},
							},
						},
					},
				}, nil
			},
			want: []clients.Status{
				{
					Context:   "test",
					State:     "success",
					TargetURL: "https://example.com",
					URL:       "https://example.com",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &commitsHandler{
				ctx: t.Context(),
				repourl: &Repo{
					id: "id",
				},
				getRefs:     tt.getRefs,
				getStatuses: tt.getStatuses,
			}
			got, err := c.listStatuses(tt.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("listStatuses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("listStatuses() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
