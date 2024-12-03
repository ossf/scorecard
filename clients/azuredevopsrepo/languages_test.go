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
	"reflect"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/projectanalysis"

	"github.com/ossf/scorecard/v5/clients"
)

func Test_listProgrammingLanguages(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		projectAnalysis func(ctx context.Context, args projectanalysis.GetProjectLanguageAnalyticsArgs) (*projectanalysis.ProjectLanguageAnalytics, error)
		want            []clients.Language
		wantErr         bool
	}{
		{
			name: "empty response",
			projectAnalysis: func(ctx context.Context, args projectanalysis.GetProjectLanguageAnalyticsArgs) (*projectanalysis.ProjectLanguageAnalytics, error) {
				return &projectanalysis.ProjectLanguageAnalytics{
					RepositoryLanguageAnalytics: &[]projectanalysis.RepositoryLanguageAnalytics{},
				}, nil
			},
			want:    []clients.Language(nil),
			wantErr: false,
		},
		{
			name: "single response",
			projectAnalysis: func(ctx context.Context, args projectanalysis.GetProjectLanguageAnalyticsArgs) (*projectanalysis.ProjectLanguageAnalytics, error) {
				return &projectanalysis.ProjectLanguageAnalytics{
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l := &languagesHandler{
				once: new(sync.Once),
				ctx:  context.Background(),
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("listProgrammingLanguages() got = %v, want %v", got, tt.want)
			}
		})
	}
}
