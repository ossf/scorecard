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

// Package githubrepo implements clients.RepoClient for GitHub.
package githubrepo

import (
	"context"
	"fmt"

	"github.com/google/go-github/v38/github"
	"github.com/shurcooL/githubv4"

	"github.com/ossf/scorecard/v2/clients"
)

// Client is GitHub-specific implementation of RepoClient.
type Client struct {
	owner        string
	repoName     string
	repo         *github.Repository
	repoClient   *github.Client
	graphClient  *graphqlHandler
	contributors *contributorsHandler
	search       *searchHandler
	ctx          context.Context
	tarball      tarballHandler
}

// InitRepo sets up the GitHub repo in local storage for improving performance and GitHub token usage efficiency.
func (client *Client) InitRepo(owner, repoName string) error {
	// Sanity check.
	repo, _, err := client.repoClient.Repositories.Get(client.ctx, owner, repoName)
	if err != nil {
		// nolint: wrapcheck
		return clients.NewRepoUnavailableError(err)
	}
	client.repo = repo
	client.owner = owner
	client.repoName = repoName

	// Init tarballHandler.
	if err := client.tarball.init(client.ctx, client.repo); err != nil {
		return fmt.Errorf("error during tarballHandler.init: %w", err)
	}

	// Setup GraphQL.
	if err := client.graphClient.init(client.ctx, owner, repoName); err != nil {
		return fmt.Errorf("error during graphqlHandler.init: %w", err)
	}

	// Setup contributors.
	if err := client.contributors.init(client.ctx, owner, repoName); err != nil {
		return fmt.Errorf("error during contributorsHandler.init: %w", err)
	}

	// Setup Search.
	client.search.init(client.ctx, owner, repoName)

	return nil
}

// URL implements RepoClient.URL.
func (client *Client) URL() string {
	return fmt.Sprintf("github.com/%s/%s", client.owner, client.repoName)
}

// ListFiles implements RepoClient.ListFiles.
func (client *Client) ListFiles(predicate func(string) (bool, error)) ([]string, error) {
	return client.tarball.listFiles(predicate)
}

// GetFileContent implements RepoClient.GetFileContent.
func (client *Client) GetFileContent(filename string) ([]byte, error) {
	return client.tarball.getFileContent(filename)
}

// ListMergedPRs implements RepoClient.ListMergedPRs.
func (client *Client) ListMergedPRs() ([]clients.PullRequest, error) {
	return client.graphClient.getMergedPRs()
}

// ListCommits implements RepoClient.ListCommits.
func (client *Client) ListCommits() ([]clients.Commit, error) {
	return client.graphClient.getCommits()
}

// ListReleases implements RepoClient.ListReleases.
func (client *Client) ListReleases() ([]clients.Release, error) {
	return client.graphClient.getReleases()
}

// ListContributors implements RepoClient.ListContributors.
func (client *Client) ListContributors() ([]clients.Contributor, error) {
	return client.contributors.getContributors()
}

// IsArchived implements RepoClient.IsArchived.
func (client *Client) IsArchived() (bool, error) {
	return client.graphClient.isArchived()
}

// GetDefaultBranch implements RepoClient.GetDefaultBranch.
func (client *Client) GetDefaultBranch() (clients.BranchRef, error) {
	return client.graphClient.getDefaultBranch()
}

// Search implements RepoClient.Search.
func (client *Client) Search(request clients.SearchRequest) (clients.SearchResponse, error) {
	return client.search.search(request)
}

// Close implements RepoClient.Close.
func (client *Client) Close() error {
	return client.tarball.cleanup()
}

// CreateGithubRepoClient returns a Client which implements RepoClient interface.
func CreateGithubRepoClient(ctx context.Context,
	client *github.Client, graphClient *githubv4.Client) clients.RepoClient {
	return &Client{
		ctx:        ctx,
		repoClient: client,
		graphClient: &graphqlHandler{
			client: graphClient,
		},
		contributors: &contributorsHandler{
			ghClient: client,
		},
		search: &searchHandler{
			ghClient: client,
		},
	}
}
