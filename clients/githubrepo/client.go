// Copyright 2021 Security Scorecard Authors
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

package githubrepo

import (
	"context"
	"fmt"

	"github.com/google/go-github/v32/github"
	"github.com/shurcooL/githubv4"

	"github.com/ossf/scorecard/clients"
)

type Client struct {
	repo        *github.Repository
	repoClient  *github.Client
	graphClient *graphqlHandler
	ctx         context.Context
	tarball     tarballHandler
}

func (client *Client) InitRepo(owner, repoName string) error {
	// Sanity check
	repo, _, err := client.repoClient.Repositories.Get(client.ctx, owner, repoName)
	if err != nil {
		// nolint: wrapcheck
		return clients.NewRepoUnavailableError(err)
	}
	client.repo = repo

	// Init tarballHandler.
	if err := client.tarball.init(client.ctx, client.repo); err != nil {
		return fmt.Errorf("error during tarballHandler.init: %w", err)
	}

	// Setup GraphQL
	if err := client.graphClient.init(client.ctx, owner, repoName); err != nil {
		return fmt.Errorf("error during graphqlHandler.init: %w", err)
	}

	return nil
}

func (client *Client) ListFiles(predicate func(string) bool) []string {
	return client.tarball.listFiles(predicate)
}

func (client *Client) GetFileContent(filename string) ([]byte, error) {
	return client.tarball.getFileContent(filename)
}

func (client *Client) ListMergedPRs() []clients.PullRequest {
	return client.graphClient.getMergedPRs()
}

func (client *Client) GetDefaultBranch() clients.BranchRef {
	return client.graphClient.getDefaultBranch()
}

func (client *Client) Close() error {
	return client.tarball.cleanup()
}

func CreateGithubRepoClient(ctx context.Context,
	client *github.Client, graphClient *githubv4.Client) clients.RepoClient {
	return &Client{
		ctx:        ctx,
		repoClient: client,
		graphClient: &graphqlHandler{
			client: graphClient,
		},
	}
}
