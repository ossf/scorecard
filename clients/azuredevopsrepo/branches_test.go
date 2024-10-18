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
	"fmt"
	"sync"
	"testing"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
)

func TestGetDefaultBranch(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func() fnQueryBranch
		expectedError bool
		expectedName  string
	}{
		{
			name: "successful branch retrieval",
			setupMock: func() fnQueryBranch {
				return func(ctx context.Context, args git.GetBranchArgs) (*git.GitBranchStats, error) {
					return &git.GitBranchStats{Name: args.Name}, nil
				}
			},
			expectedError: false,
			expectedName:  "main",
		},
		{
			name: "error during branch retrieval",
			setupMock: func() fnQueryBranch {
				return func(ctx context.Context, args git.GetBranchArgs) (*git.GitBranchStats, error) {
					return nil, fmt.Errorf("error")
				}
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &branchesHandler{
				queryBranch: tt.setupMock(),
				once:        new(sync.Once),
				repourl: &Repo{
					id:            "repo-id",
					defaultBranch: "main",
				},
			}

			branch, err := handler.getDefaultBranch()
			if (err != nil) != tt.expectedError {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			}
			if branch != nil && *branch.Name != tt.expectedName {
				t.Errorf("expected branch name: %v, got: %v", tt.expectedName, *branch.Name)
			}
		})
	}
}
