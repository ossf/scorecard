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

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/policy"

	"github.com/ossf/scorecard/v5/clients"
)

func Test_getDefaultBranch(t *testing.T) {
	t.Parallel()
	tests := []struct {
		queryBranch   fnQueryBranch
		name          string
		expectedName  string
		expectedError bool
	}{
		{
			name: "successful branch retrieval",
			queryBranch: func(ctx context.Context, args git.GetBranchArgs) (*git.GitBranchStats, error) {
				return &git.GitBranchStats{Name: args.Name}, nil
			},
			expectedError: false,
			expectedName:  "main",
		},
		{
			name: "error during branch retrieval",
			queryBranch: func(ctx context.Context, args git.GetBranchArgs) (*git.GitBranchStats, error) {
				return nil, fmt.Errorf("error")
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			handler := &branchesHandler{
				queryBranch: tt.queryBranch,
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

func Test_getBranch(t *testing.T) {
	t.Parallel()
	tests := []struct {
		getBranch               fnQueryBranch
		getPolicyConfigurations fnGetPolicyConfigurations
		expected                *clients.BranchRef
		name                    string
		branchName              string
		expectedError           bool
	}{
		{
			name:       "successful branch retrieval",
			branchName: "main",
			getBranch: func(ctx context.Context, args git.GetBranchArgs) (*git.GitBranchStats, error) {
				return &git.GitBranchStats{Name: args.Name}, nil
			},
			getPolicyConfigurations: func(ctx context.Context, args git.GetPolicyConfigurationsArgs) (*git.GitPolicyConfigurationResponse, error) {
				return &git.GitPolicyConfigurationResponse{
					PolicyConfigurations: &[]policy.PolicyConfiguration{},
				}, nil
			},
			expected: &clients.BranchRef{
				Name:      toPtr("main"),
				Protected: toPtr(false),
			},
			expectedError: false,
		},
		{
			name:       "error during branch retrieval",
			branchName: "main",
			getBranch: func(ctx context.Context, args git.GetBranchArgs) (*git.GitBranchStats, error) {
				return nil, fmt.Errorf("error")
			},
			getPolicyConfigurations: func(ctx context.Context, args git.GetPolicyConfigurationsArgs) (*git.GitPolicyConfigurationResponse, error) {
				return &git.GitPolicyConfigurationResponse{}, nil
			},
			expected:      nil,
			expectedError: true,
		},
		{
			name:       "error during policy configuration retrieval",
			branchName: "main",
			getBranch: func(ctx context.Context, args git.GetBranchArgs) (*git.GitBranchStats, error) {
				return &git.GitBranchStats{Name: args.Name}, nil
			},
			getPolicyConfigurations: func(ctx context.Context, args git.GetPolicyConfigurationsArgs) (*git.GitPolicyConfigurationResponse, error) {
				return nil, fmt.Errorf("error")
			},
			expected:      nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			handler := &branchesHandler{
				queryBranch:             tt.getBranch,
				getPolicyConfigurations: tt.getPolicyConfigurations,
				once:                    new(sync.Once),
				repourl: &Repo{
					id:            uuid.Nil.String(),
					defaultBranch: "main",
				},
			}

			branch, err := handler.getBranch(tt.branchName)
			if (err != nil) != tt.expectedError {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			}
			if diff := cmp.Diff(branch, tt.expected); diff != "" {
				t.Errorf("mismatch in branch ref (-want +got):\n%s", diff)
			}
		})
	}
}
