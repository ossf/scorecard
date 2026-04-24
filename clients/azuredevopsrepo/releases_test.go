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
	"errors"
	"sync"
	"testing"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
)

func Test_releasesHandler_listReleases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		getRefs fnGetRefs
		name    string
		wantTag string
		wantSHA string
		wantLen int
		wantErr bool
	}{
		{
			name: "lightweight tag",
			getRefs: func(ctx context.Context, args git.GetRefsArgs) (*git.GetRefsResponseValue, error) {
				return &git.GetRefsResponseValue{
					Value: []git.GitRef{
						{
							Name:     strptr("refs/tags/v1.0.0"),
							ObjectId: strptr("abc123def456789012345678901234567890abcd"),
						},
					},
				}, nil
			},
			wantErr: false,
			wantLen: 1,
			wantTag: "v1.0.0",
			wantSHA: "abc123def456789012345678901234567890abcd",
		},
		{
			name: "annotated tag uses peeled object id",
			getRefs: func(ctx context.Context, args git.GetRefsArgs) (*git.GetRefsResponseValue, error) {
				return &git.GetRefsResponseValue{
					Value: []git.GitRef{
						{
							Name:           strptr("refs/tags/v2.0.0"),
							ObjectId:       strptr("tag-object-sha-not-a-commit-sha1234567890"),
							PeeledObjectId: strptr("real-commit-sha1234567890abcdef01234567"),
						},
					},
				}, nil
			},
			wantErr: false,
			wantLen: 1,
			wantTag: "v2.0.0",
			wantSHA: "real-commit-sha1234567890abcdef01234567",
		},
		{
			name: "multiple tags",
			getRefs: func(ctx context.Context, args git.GetRefsArgs) (*git.GetRefsResponseValue, error) {
				return &git.GetRefsResponseValue{
					Value: []git.GitRef{
						{
							Name:     strptr("refs/tags/v1.0.0"),
							ObjectId: strptr("sha1"),
						},
						{
							Name:     strptr("refs/tags/v2.0.0"),
							ObjectId: strptr("sha2"),
						},
					},
				}, nil
			},
			wantErr: false,
			wantLen: 2,
			wantTag: "v1.0.0",
			wantSHA: "sha1",
		},
		{
			name: "no tags",
			getRefs: func(ctx context.Context, args git.GetRefsArgs) (*git.GetRefsResponseValue, error) {
				return &git.GetRefsResponseValue{
					Value: []git.GitRef{},
				}, nil
			},
			wantErr: false,
			wantLen: 0,
		},
		{
			name: "skips refs with nil name",
			getRefs: func(ctx context.Context, args git.GetRefsArgs) (*git.GetRefsResponseValue, error) {
				return &git.GetRefsResponseValue{
					Value: []git.GitRef{
						{
							ObjectId: strptr("sha1"),
						},
						{
							Name:     strptr("refs/tags/v1.0.0"),
							ObjectId: strptr("sha2"),
						},
					},
				}, nil
			},
			wantErr: false,
			wantLen: 1,
			wantTag: "v1.0.0",
			wantSHA: "sha2",
		},
		{
			name: "API error",
			getRefs: func(ctx context.Context, args git.GetRefsArgs) (*git.GetRefsResponseValue, error) {
				return nil, errors.New("API error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			handler := &releasesHandler{
				once:    new(sync.Once),
				getRefs: tt.getRefs,
				repourl: &Repo{id: "test-repo-id"},
			}
			releases, err := handler.listReleases()
			if (err != nil) != tt.wantErr {
				t.Fatalf("listReleases() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(releases) != tt.wantLen {
				t.Fatalf("listReleases() returned %d releases, want %d", len(releases), tt.wantLen)
			}
			if tt.wantLen > 0 {
				if releases[0].TagName != tt.wantTag {
					t.Errorf("TagName = %q, want %q", releases[0].TagName, tt.wantTag)
				}
				if releases[0].TargetCommitish != tt.wantSHA {
					t.Errorf("TargetCommitish = %q, want %q", releases[0].TargetCommitish, tt.wantSHA)
				}
			}
		})
	}
}
