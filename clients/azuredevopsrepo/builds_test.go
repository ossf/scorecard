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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/build"

	"github.com/ossf/scorecard/v5/clients"
)

func Test_listSuccessfulBuilds(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                string
		getBuildDefinitions fnListBuildDefinitions
		getBuilds           fnGetBuilds
		want                []clients.WorkflowRun
		wantErr             bool
	}{
		{
			name: "no build definitions",
			getBuildDefinitions: func(ctx context.Context, args build.GetDefinitionsArgs) (*build.GetDefinitionsResponseValue, error) {
				return &build.GetDefinitionsResponseValue{
					Value: []build.BuildDefinitionReference{},
				}, nil
			},
			getBuilds: func(ctx context.Context, args build.GetBuildsArgs) (*build.GetBuildsResponseValue, error) {
				return &build.GetBuildsResponseValue{}, nil
			},
			want:    []clients.WorkflowRun{},
			wantErr: false,
		},
		{
			name: "no builds",
			getBuildDefinitions: func(ctx context.Context, args build.GetDefinitionsArgs) (*build.GetDefinitionsResponseValue, error) {
				return &build.GetDefinitionsResponseValue{
					Value: []build.BuildDefinitionReference{
						{Id: toPtr(123)},
					},
				}, nil
			},
			getBuilds: func(ctx context.Context, args build.GetBuildsArgs) (*build.GetBuildsResponseValue, error) {
				return &build.GetBuildsResponseValue{
					Value: []build.Build{},
				}, nil
			},
			want:    []clients.WorkflowRun{},
			wantErr: false,
		},
		{
			name: "single definition and build",
			getBuildDefinitions: func(ctx context.Context, args build.GetDefinitionsArgs) (*build.GetDefinitionsResponseValue, error) {
				return &build.GetDefinitionsResponseValue{
					Value: []build.BuildDefinitionReference{
						{Id: toPtr(123)},
					},
				}, nil
			},
			getBuilds: func(ctx context.Context, args build.GetBuildsArgs) (*build.GetBuildsResponseValue, error) {
				return &build.GetBuildsResponseValue{
					Value: []build.Build{
						{
							Url:           toPtr("https://example.com"),
							SourceVersion: toPtr("abc123"),
						},
					},
				}, nil
			},
			want: []clients.WorkflowRun{
				{
					URL:     "https://example.com",
					HeadSHA: toPtr("abc123"),
				},
			},
			wantErr: false,
		},
		{
			name: "multiple definitions and builds",
			getBuildDefinitions: func(ctx context.Context, args build.GetDefinitionsArgs) (*build.GetDefinitionsResponseValue, error) {
				return &build.GetDefinitionsResponseValue{
					Value: []build.BuildDefinitionReference{
						{Id: toPtr(123)},
						{Id: toPtr(456)},
					},
				}, nil
			},
			getBuilds: func(ctx context.Context, args build.GetBuildsArgs) (*build.GetBuildsResponseValue, error) {
				return &build.GetBuildsResponseValue{
					Value: []build.Build{
						{
							Url:           toPtr("https://example.com"),
							SourceVersion: toPtr("abc123"),
						},
						{
							Url:           toPtr("https://test.com"),
							SourceVersion: toPtr("def456"),
						},
					},
				}, nil
			},
			want: []clients.WorkflowRun{
				{
					URL:     "https://example.com",
					HeadSHA: toPtr("abc123"),
				},
				{
					URL:     "https://test.com",
					HeadSHA: toPtr("def456"),
				},
			},
			wantErr: false,
		},
		{
			name: "multiple definitions and builds with continuation token",
			getBuildDefinitions: func(ctx context.Context, args build.GetDefinitionsArgs) (*build.GetDefinitionsResponseValue, error) {
				if args.ContinuationToken == nil {
					return &build.GetDefinitionsResponseValue{
						Value: []build.BuildDefinitionReference{
							{Id: toPtr(123)},
						},
						ContinuationToken: "abc123",
					}, nil
				}
				return &build.GetDefinitionsResponseValue{
					Value: []build.BuildDefinitionReference{
						{Id: toPtr(789)},
					},
				}, nil
			},
			getBuilds: func(ctx context.Context, args build.GetBuildsArgs) (*build.GetBuildsResponseValue, error) {
				return &build.GetBuildsResponseValue{
					Value: []build.Build{
						{
							Url:           toPtr("https://example.com"),
							SourceVersion: toPtr("abc123"),
						},
					},
				}, nil
			},
			want: []clients.WorkflowRun{
				{
					URL:     "https://example.com",
					HeadSHA: toPtr("abc123"),
				},
			},
			wantErr: false,
		},
		{
			name: "build definitions error",
			getBuildDefinitions: func(ctx context.Context, args build.GetDefinitionsArgs) (*build.GetDefinitionsResponseValue, error) {
				return nil, errors.New("error")
			},
			getBuilds: func(ctx context.Context, args build.GetBuildsArgs) (*build.GetBuildsResponseValue, error) {
				return &build.GetBuildsResponseValue{}, nil
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "builds error",
			getBuildDefinitions: func(ctx context.Context, args build.GetDefinitionsArgs) (*build.GetDefinitionsResponseValue, error) {
				return &build.GetDefinitionsResponseValue{
					Value: []build.BuildDefinitionReference{
						{Id: toPtr(123)},
					},
				}, nil
			},
			getBuilds: func(ctx context.Context, args build.GetBuildsArgs) (*build.GetBuildsResponseValue, error) {
				return nil, errors.New("error")
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			b := &buildsHandler{
				repourl: &Repo{
					project: "test",
					id:      "123",
				},
				getBuildDefinitions: tt.getBuildDefinitions,
				getBuilds:           tt.getBuilds,
			}
			got, err := b.listSuccessfulBuilds("test.yaml")

			if (err != nil) != tt.wantErr {
				t.Errorf("listSuccessfulBuilds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("listSuccessfulBuilds() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
