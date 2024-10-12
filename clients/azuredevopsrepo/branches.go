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

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/ossf/scorecard/v5/clients"
)

type branchesHandler struct {
	gitClient        git.Client
	ctx              context.Context
	once             *sync.Once
	errSetup         error
	repourl          *Repo
	defaultBranchRef *clients.BranchRef
	queryBranch      fnQueryBranch
}

func (handler *branchesHandler) init(ctx context.Context, repourl *Repo) {
	handler.ctx = ctx
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
	handler.queryBranch = handler.gitClient.GetBranch
}

type (
	fnQueryBranch func(ctx context.Context, args git.GetBranchArgs) (*git.GitBranchStats, error)
)

func (handler *branchesHandler) setup() error {
	handler.once.Do(func() {
		branch, err := handler.queryBranch(handler.ctx, git.GetBranchArgs{
			RepositoryId: &handler.repourl.ID,
			Name:         &handler.repourl.defaultBranch,
		})
		if err != nil {
			handler.errSetup = fmt.Errorf("request for default branch failed with error %w", err)
			return
		}
		handler.defaultBranchRef = &clients.BranchRef{
			Name: branch.Name,
		}

		handler.errSetup = nil
	})
	return handler.errSetup

}

func (handler *branchesHandler) getDefaultBranch() (*clients.BranchRef, error) {
	err := handler.setup()
	if err != nil {
		return nil, fmt.Errorf("error during branchesHandler.setup: %w", err)
	}

	return handler.defaultBranchRef, nil
}
