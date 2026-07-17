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

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/projectanalysis"

	"github.com/ossf/scorecard/v5/clients"
)

func Test_listProgrammingLanguages(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		projectAnalysis fnGetProjectLanguageAnalytics
		want            []clients.Language
		wantErr         bool
	}{
		{
			name: "empty response",
			projectAnalysis: func(ctx context.Context, args projectanalysis.GetProjectLanguageAnalyticsArgs) (*projectanalysis.ProjectLanguageAnalytics, error) {
				return &projectanalysis.ProjectLanguageAnalytics{
					RepositoryLanguageAnalytics: &[]projectanalysis.RepositoryLanguageAnalytics{},
					ResultPhase:                 toPtr(projectanalysis.ResultPhaseValues.Full),
				}, nil
			},
			want:    []clients.Language(nil),
			wantErr: false,
		},
		{
			name: "single response",
			projectAnalysis: func(ctx context.Context, args projectanalysis.GetProjectLanguageAnalyticsArgs) (*projectanalysis.ProjectLanguageAnalytics, error) {
				return &projectanalysis.ProjectLanguageAnalytics{
					ResultPhase: toPtr(projectanalysis.ResultPhaseValues.Full),
					RepositoryLanguageAnalytics: &[]projectanalysis.RepositoryLanguageAnalytics{
						{
							Id: toPtr(uuid.Nil),
							LanguageBreakdown: &[]projectanalysis.LanguageStatistics{
								{
									Name:               toPtr("test"),
									LanguagePercentage: toPtr(float64(100)),
								},
							},
						},
					},
				}, nil
			},
			want: []clients.Language{
				{
					Name:     "test",
					NumLines: 100,
				},
			},
			wantErr: false,
		},
		{
			name: "multiple response",
			projectAnalysis: func(ctx context.Context, args projectanalysis.GetProjectLanguageAnalyticsArgs) (*projectanalysis.ProjectLanguageAnalytics, error) {
				return &projectanalysis.ProjectLanguageAnalytics{
					ResultPhase: toPtr(projectanalysis.ResultPhaseValues.Full),
					RepositoryLanguageAnalytics: &[]projectanalysis.RepositoryLanguageAnalytics{
						{
							Id: toPtr(uuid.Nil),
							LanguageBreakdown: &[]projectanalysis.LanguageStatistics{
								{
									Name:               toPtr("test1"),
									LanguagePercentage: toPtr(float64(50)),
								},
								{
									Name:               toPtr("test2"),
									LanguagePercentage: toPtr(float64(50)),
								},
							},
						},
					},
				}, nil
			},
			want: []clients.Language{
				{
					Name:     "test1",
					NumLines: 50,
				},
				{
					Name:     "test2",
					NumLines: 50,
				},
			},
			wantErr: false,
		},
		{
			name: "API error",
			projectAnalysis: func(ctx context.Context, args projectanalysis.GetProjectLanguageAnalyticsArgs) (*projectanalysis.ProjectLanguageAnalytics, error) {
				return nil, errors.New("network error")
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l := &languagesHandler{
				once: new(sync.Once),
				ctx:  t.Context(),
				repourl: &Repo{
					id:      uuid.Nil.String(),
					project: "project",
				},
				projectAnalysis: tt.projectAnalysis,
			}
			got, err := l.listProgrammingLanguages()
			if (err != nil) != tt.wantErr {
				t.Errorf("listProgrammingLanguages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("languages mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_isLanguageAnalyticsUnavailable(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet
		name string
		err  error
		want bool
	}{
		{
			name: "forbidden status",
			err: azuredevops.WrappedError{
				Message:    toPtr("forbidden"),
				StatusCode: toPtr(403),
			},
			want: true,
		},
		{
			name: "not found status",
			err: azuredevops.WrappedError{
				Message:    toPtr("not found"),
				StatusCode: toPtr(404),
			},
			want: true,
		},
		{
			name: "unauthorized pointer status",
			err: &azuredevops.WrappedError{
				Message:    toPtr("unauthorized"),
				StatusCode: toPtr(401),
			},
			want: true,
		},
		{
			name: "access denied message",
			err:  errors.New("Access Denied: missing permission"),
			want: true,
		},
		{
			name: "server error",
			err: azuredevops.WrappedError{
				Message:    toPtr("server error"),
				StatusCode: toPtr(500),
			},
			want: false,
		},
		{
			name: "network error",
			err:  errors.New("network error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isLanguageAnalyticsUnavailable(tt.err); got != tt.want {
				t.Errorf("isLanguageAnalyticsUnavailable() = %t, want %t", got, tt.want)
			}
		})
	}
}

func Test_listProgrammingLanguagesFallback(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet
		name            string
		projectAnalysis fnGetProjectLanguageAnalytics
	}{
		{
			name: "permission denied",
			projectAnalysis: func(ctx context.Context, args projectanalysis.GetProjectLanguageAnalyticsArgs) (*projectanalysis.ProjectLanguageAnalytics, error) {
				return nil, azuredevops.WrappedError{
					Message:    toPtr("Access Denied"),
					StatusCode: toPtr(403),
				}
			},
		},
		{
			name: "nil phase",
			projectAnalysis: func(ctx context.Context, args projectanalysis.GetProjectLanguageAnalyticsArgs) (*projectanalysis.ProjectLanguageAnalytics, error) {
				return &projectanalysis.ProjectLanguageAnalytics{}, nil
			},
		},
		{
			name: "preliminary phase",
			projectAnalysis: func(ctx context.Context, args projectanalysis.GetProjectLanguageAnalyticsArgs) (*projectanalysis.ProjectLanguageAnalytics, error) {
				return &projectanalysis.ProjectLanguageAnalytics{
					ResultPhase: toPtr(projectanalysis.ResultPhaseValues.Preliminary),
				}, nil
			},
		},
	}
	want := []clients.Language{{Name: clients.All, NumLines: 1}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			projectCalls := 0
			l := &languagesHandler{
				once: new(sync.Once),
				ctx:  t.Context(),
				repourl: &Repo{
					id:      uuid.Nil.String(),
					project: "project",
				},
				projectAnalysis: func(ctx context.Context, args projectanalysis.GetProjectLanguageAnalyticsArgs) (*projectanalysis.ProjectLanguageAnalytics, error) {
					projectCalls++
					return tt.projectAnalysis(ctx, args)
				},
			}

			got, err := l.listProgrammingLanguages()
			if err != nil {
				t.Fatalf("listProgrammingLanguages() error = %v", err)
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("languages mismatch (-want +got):\n%s", diff)
			}

			got, err = l.listProgrammingLanguages()
			if err != nil {
				t.Fatalf("second listProgrammingLanguages() error = %v", err)
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("second languages mismatch (-want +got):\n%s", diff)
			}
			if projectCalls != 1 {
				t.Errorf("project analysis calls = %d, want 1", projectCalls)
			}
		})
	}
}

func Test_listProgrammingLanguagesMalformedResponse(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet
		name     string
		response *projectanalysis.ProjectLanguageAnalytics
	}{
		{
			name:     "nil response",
			response: nil,
		},
		{
			name: "missing repositories",
			response: &projectanalysis.ProjectLanguageAnalytics{
				ResultPhase: toPtr(projectanalysis.ResultPhaseValues.Full),
			},
		},
		{
			name: "missing repository ID",
			response: &projectanalysis.ProjectLanguageAnalytics{
				ResultPhase: toPtr(projectanalysis.ResultPhaseValues.Full),
				RepositoryLanguageAnalytics: &[]projectanalysis.RepositoryLanguageAnalytics{
					{},
				},
			},
		},
		{
			name: "missing language breakdown",
			response: &projectanalysis.ProjectLanguageAnalytics{
				ResultPhase: toPtr(projectanalysis.ResultPhaseValues.Full),
				RepositoryLanguageAnalytics: &[]projectanalysis.RepositoryLanguageAnalytics{
					{Id: toPtr(uuid.Nil)},
				},
			},
		},
		{
			name: "missing language name",
			response: &projectanalysis.ProjectLanguageAnalytics{
				ResultPhase: toPtr(projectanalysis.ResultPhaseValues.Full),
				RepositoryLanguageAnalytics: &[]projectanalysis.RepositoryLanguageAnalytics{
					{
						Id: toPtr(uuid.Nil),
						LanguageBreakdown: &[]projectanalysis.LanguageStatistics{
							{},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			l := &languagesHandler{
				once: new(sync.Once),
				ctx:  t.Context(),
				repourl: &Repo{
					id:      uuid.Nil.String(),
					project: "project",
				},
				projectAnalysis: func(ctx context.Context, args projectanalysis.GetProjectLanguageAnalyticsArgs) (*projectanalysis.ProjectLanguageAnalytics, error) {
					return tt.response, nil
				},
			}

			_, err := l.listProgrammingLanguages()
			if !errors.Is(err, errLanguageAnalyticsMalformed) {
				t.Errorf("listProgrammingLanguages() error = %v, want malformed analytics error", err)
			}
		})
	}
}
