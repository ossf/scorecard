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

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
)

type mockGitClient struct {
	git.Client
	isDisabled bool
	err        error
}

func (m *mockGitClient) GetRepository(ctx context.Context, args git.GetRepositoryArgs) (*git.GitRepository, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &git.GitRepository{IsDisabled: &m.isDisabled}, nil
}

func TestIsArchived(t *testing.T) {
	tests := []struct {
		name       string
		isDisabled bool
		err        error
		want       bool
		wantErr    bool
	}{
		{
			name:       "repository is archived",
			isDisabled: true,
			want:       true,
			wantErr:    false,
		},
		{
			name:       "repository is not archived",
			isDisabled: false,
			want:       false,
			wantErr:    false,
		},
		{
			name:    "error getting repository",
			err:     errors.New("some error"),
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				azdoClient: &mockGitClient{
					isDisabled: tt.isDisabled,
					err:        tt.err,
				},
				repourl: &Repo{ID: "test-repo-id"},
			}

			got, err := client.IsArchived()
			if (err != nil) != tt.wantErr {
				t.Errorf("IsArchived() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsArchived() = %v, want %v", got, tt.want)
			}
		})
	}
}