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

	"github.com/google/uuid"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"

	"github.com/ossf/scorecard/v5/clients"
)

type branchesHandler struct {
	gitClient               git.Client
	ctx                     context.Context
	once                    *sync.Once
	errSetup                error
	repourl                 *Repo
	defaultBranchRef        *clients.RepoRef
	queryBranch             fnQueryBranch
	getPolicyConfigurations fnGetPolicyConfigurations
}

func (b *branchesHandler) init(ctx context.Context, repourl *Repo) {
	b.ctx = ctx
	b.repourl = repourl
	b.errSetup = nil
	b.once = new(sync.Once)
	b.queryBranch = b.gitClient.GetBranch
	b.getPolicyConfigurations = b.gitClient.GetPolicyConfigurations
}

type (
	fnQueryBranch             func(ctx context.Context, args git.GetBranchArgs) (*git.GitBranchStats, error)
	fnGetPolicyConfigurations func(
		ctx context.Context,
		args git.GetPolicyConfigurationsArgs,
	) (*git.GitPolicyConfigurationResponse, error)
)

func (b *branchesHandler) setup() error {
	b.once.Do(func() {
		args := git.GetBranchArgs{
			RepositoryId: &b.repourl.id,
			Name:         &b.repourl.defaultBranch,
		}
		branch, err := b.queryBranch(b.ctx, args)
		if err != nil {
			b.errSetup = fmt.Errorf("request for default branch failed with error %w", err)
			return
		}
		b.defaultBranchRef = &clients.RepoRef{
			Name: branch.Name,
		}

		b.errSetup = nil
	})
	return b.errSetup
}

func (b *branchesHandler) getDefaultBranch() (*clients.RepoRef, error) {
	if err := b.setup(); err != nil {
		return nil, fmt.Errorf("error during branchesHandler.setup: %w", err)
	}

	return b.defaultBranchRef, nil
}

func (b *branchesHandler) getBranch(branchName string) (*clients.RepoRef, error) {
	branch, err := b.queryBranch(b.ctx, git.GetBranchArgs{
		RepositoryId: &b.repourl.id,
		Name:         &branchName,
	})
	if err != nil {
		return nil, fmt.Errorf("request for branch %s failed with error %w", branchName, err)
	}

	refName := fmt.Sprintf("refs/heads/%s", *branch.Name)
	repositoryID, err := uuid.Parse(b.repourl.id)
	if err != nil {
		return nil, fmt.Errorf("error parsing repository ID %s: %w", b.repourl.id, err)
	}
	args := git.GetPolicyConfigurationsArgs{
		RepositoryId: &repositoryID,
		RefName:      &refName,
	}
	response, err := b.getPolicyConfigurations(b.ctx, args)
	if err != nil {
		return nil, fmt.Errorf("request for policy configurations failed with error %w", err)
	}

	isBranchProtected := false
	if len(*response.PolicyConfigurations) > 0 {
		isBranchProtected = true
	}

	// TODO: map Azure DevOps branch protection to Scorecard branch protection
	return &clients.RepoRef{
		Name:      branch.Name,
		Protected: &isBranchProtected,
	}, nil
}
