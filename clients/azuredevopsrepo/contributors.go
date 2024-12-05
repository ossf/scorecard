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
	"sync"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"

	"github.com/ossf/scorecard/v5/clients"
)

type contributorsHandler struct {
	ctx          context.Context
	once         *sync.Once
	repourl      *Repo
	gitClient    git.Client
	errSetup     error
	getCommits   fnGetCommits
	contributors []clients.User
}

func (c *contributorsHandler) init(ctx context.Context, repourl *Repo) {
	c.ctx = ctx
	c.once = new(sync.Once)
	c.repourl = repourl
	c.errSetup = nil
	c.getCommits = c.gitClient.GetCommits
	c.contributors = nil
}

func (c *contributorsHandler) setup() error {
	c.once.Do(func() {
		contributors := make(map[string]clients.User)
		commitsPageSize := 1000
		skip := 0
		for {
			args := git.GetCommitsArgs{
				RepositoryId: &c.repourl.id,
				SearchCriteria: &git.GitQueryCommitsCriteria{
					Top:  &commitsPageSize,
					Skip: &skip,
				},
			}
			commits, err := c.getCommits(c.ctx, args)
			if err != nil {
				c.errSetup = err
				return
			}

			if commits == nil || len(*commits) == 0 {
				break
			}

			for i := range *commits {
				commit := (*commits)[i]
				email := *commit.Author.Email
				if _, ok := contributors[email]; ok {
					user := contributors[email]
					user.NumContributions++
					contributors[email] = user
				} else {
					contributors[email] = clients.User{
						Login:            email,
						NumContributions: 1,
						Companies:        []string{c.repourl.organization},
					}
				}
			}

			skip += commitsPageSize
		}

		for _, contributor := range contributors {
			c.contributors = append(c.contributors, contributor)
		}
	})
	return c.errSetup
}

func (c *contributorsHandler) listContributors() ([]clients.User, error) {
	if err := c.setup(); err != nil {
		return nil, err
	}

	return c.contributors, nil
}
