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
	"net/http"

	"github.com/google/go-github/v38/github"
	"github.com/shurcooL/githubv4"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/ossf/scorecard/v2/clients"
	"github.com/ossf/scorecard/v2/clients/githubrepo/roundtripper"
	sce "github.com/ossf/scorecard/v2/errors"
)

// Client is GitHub-specific implementation of RepoClient.
type Client struct {
	owner        string
	repoName     string
	repo         *github.Repository
	repoClient   *github.Client
	graphClient  *graphqlHandler
	contributors *contributorsHandler
	branches     *branchesHandler
	releases     *releasesHandler
	workflows    *workflowsHandler
	checkruns    *checkrunsHandler
	statuses     *statusesHandler
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
		return sce.Create(sce.ErrRepoUnreachable, err.Error())
	}
	client.repo = repo
	client.owner = repo.Owner.GetLogin()
	client.repoName = repo.GetName()

	// Init tarballHandler.
	if err := client.tarball.init(client.ctx, client.repo); err != nil {
		return fmt.Errorf("error during tarballHandler.init: %w", err)
	}

	// Setup GraphQL.
	client.graphClient.init(client.ctx, client.owner, client.repoName)

	// Setup contributorsHandler.
	client.contributors.init(client.ctx, client.owner, client.repoName)

	// Setup branchesHandler.
	client.branches.init(client.ctx, client.owner, client.repoName)

	// Setup releasesHandler.
	client.releases.init(client.ctx, client.owner, client.repoName)

	// Setup workflowsHandler.
	client.workflows.init(client.ctx, client.owner, client.repoName)

	// Setup checkrunsHandler.
	client.checkruns.init(client.ctx, client.owner, client.repoName)

	// Setup statusesHandler.
	client.statuses.init(client.ctx, client.owner, client.repoName)

	// Setup searchHandler.
	client.search.init(client.ctx, client.owner, client.repoName)

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
	return client.releases.getReleases()
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
func (client *Client) GetDefaultBranch() (*clients.BranchRef, error) {
	return client.branches.getDefaultBranch()
}

// ListBranches implements RepoClient.ListBranches.
func (client *Client) ListBranches() ([]*clients.BranchRef, error) {
	return client.branches.listBranches()
}

// ListSuccessfulWorkflowRuns implements RepoClient.WorkflowRunsByFilename.
func (client *Client) ListSuccessfulWorkflowRuns(filename string) ([]clients.WorkflowRun, error) {
	return client.workflows.listSuccessfulWorkflowRuns(filename)
}

// ListCheckRunsForRef implements RepoClient.ListCheckRunsForRef.
func (client *Client) ListCheckRunsForRef(ref string) ([]clients.CheckRun, error) {
	return client.checkruns.listCheckRunsForRef(ref)
}

// ListStatuses implements RepoClient.ListStatuses.
func (client *Client) ListStatuses(ref string) ([]clients.Status, error) {
	return client.statuses.listStatuses(ref)
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
func CreateGithubRepoClient(ctx context.Context, logger *zap.Logger) clients.RepoClient {
	// Use our custom roundtripper
	rt := roundtripper.NewTransport(ctx, logger.Sugar())
	httpClient := &http.Client{
		Transport: rt,
	}
	client := github.NewClient(httpClient)
	graphClient := githubv4.NewClient(httpClient)

	return &Client{
		ctx:        ctx,
		repoClient: client,
		graphClient: &graphqlHandler{
			client: graphClient,
		},
		contributors: &contributorsHandler{
			ghClient: client,
		},
		branches: &branchesHandler{
			ghClient:    client,
			graphClient: graphClient,
		},
		releases: &releasesHandler{
			client: client,
		},
		workflows: &workflowsHandler{
			client: client,
		},
		checkruns: &checkrunsHandler{
			client: client,
		},
		statuses: &statusesHandler{
			client: client,
		},
		search: &searchHandler{
			ghClient: client,
		},
	}
}

// NewLogger creates an instance of *zap.Logger.
func NewLogger(logLevel zapcore.Level) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.Level.SetLevel(logLevel)
	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("cfg.Build: %w", err)
	}
	return logger, nil
}
