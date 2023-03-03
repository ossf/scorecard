// Copyright 2021 OpenSSF Scorecard Authors
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
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/go-github/v38/github"
	"github.com/shurcooL/githubv4"

	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/clients/githubrepo/roundtripper"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/log"
)

var (
	_                     clients.RepoClient = &Client{}
	errInputRepoType                         = errors.New("input repo should be of type repoURL")
	errDefaultBranchEmpty                    = errors.New("default branch name is empty")
)

// Client is GitHub-specific implementation of RepoClient.
type Client struct {
	repourl       *repoURL
	repo          *github.Repository
	repoClient    *github.Client
	graphClient   *graphqlHandler
	contributors  *contributorsHandler
	branches      *branchesHandler
	releases      *releasesHandler
	workflows     *workflowsHandler
	checkruns     *checkrunsHandler
	statuses      *statusesHandler
	search        *searchHandler
	searchCommits *searchCommitsHandler
	webhook       *webhookHandler
	languages     *languagesHandler
	licenses      *licensesHandler
	ctx           context.Context
	tarball       tarballHandler
	commitDepth   int
}

// InitRepo sets up the GitHub repo in local storage for improving performance and GitHub token usage efficiency.
func (client *Client) InitRepo(inputRepo clients.Repo, commitSHA string, commitDepth int) error {
	ghRepo, ok := inputRepo.(*repoURL)
	if !ok {
		return fmt.Errorf("%w: %v", errInputRepoType, inputRepo)
	}

	// Sanity check.
	repo, _, err := client.repoClient.Repositories.Get(client.ctx, ghRepo.owner, ghRepo.repo)
	if err != nil {
		return sce.WithMessage(sce.ErrRepoUnreachable, err.Error())
	}
	if commitDepth <= 0 {
		client.commitDepth = 30 // default
	} else {
		client.commitDepth = commitDepth
	}
	client.repo = repo
	client.repourl = &repoURL{
		owner:         repo.Owner.GetLogin(),
		repo:          repo.GetName(),
		defaultBranch: repo.GetDefaultBranch(),
		commitSHA:     commitSHA,
	}

	// Init tarballHandler.
	client.tarball.init(client.ctx, client.repo, commitSHA)

	// Setup GraphQL.
	client.graphClient.init(client.ctx, client.repourl, client.commitDepth)

	// Setup contributorsHandler.
	client.contributors.init(client.ctx, client.repourl)

	// Setup branchesHandler.
	client.branches.init(client.ctx, client.repourl)

	// Setup releasesHandler.
	client.releases.init(client.ctx, client.repourl)

	// Setup workflowsHandler.
	client.workflows.init(client.ctx, client.repourl)

	// Setup checkrunsHandler.
	client.checkruns.init(client.ctx, client.repourl, commitDepth)

	// Setup statusesHandler.
	client.statuses.init(client.ctx, client.repourl)

	// Setup searchHandler.
	client.search.init(client.ctx, client.repourl)

	// Setup searchCommitsHandler
	client.searchCommits.init(client.ctx, client.repourl)

	// Setup webhookHandler.
	client.webhook.init(client.ctx, client.repourl)

	// Setup languagesHandler.
	client.languages.init(client.ctx, client.repourl)

	// Setup licensesHandler.
	client.licenses.init(client.ctx, client.repourl)
	return nil
}

// URI implements RepoClient.URI.
func (client *Client) URI() string {
	return fmt.Sprintf("github.com/%s/%s", client.repourl.owner, client.repourl.repo)
}

// LocalPath implements RepoClient.LocalPath.
func (client *Client) LocalPath() (string, error) {
	return client.tarball.getLocalPath()
}

// ListFiles implements RepoClient.ListFiles.
func (client *Client) ListFiles(predicate func(string) (bool, error)) ([]string, error) {
	return client.tarball.listFiles(predicate)
}

// GetFileContent implements RepoClient.GetFileContent.
func (client *Client) GetFileContent(filename string) ([]byte, error) {
	return client.tarball.getFileContent(filename)
}

// ListCommits implements RepoClient.ListCommits.
func (client *Client) ListCommits() ([]clients.Commit, error) {
	return client.graphClient.getCommits()
}

// ListIssues implements RepoClient.ListIssues.
func (client *Client) ListIssues() ([]clients.Issue, error) {
	// here you would need to pass commitDepth or something
	return client.graphClient.getIssues()
}

// ListReleases implements RepoClient.ListReleases.
func (client *Client) ListReleases() ([]clients.Release, error) {
	return client.releases.getReleases()
}

// ListContributors implements RepoClient.ListContributors.
func (client *Client) ListContributors() ([]clients.User, error) {
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

// GetDefaultBranchName implements RepoClient.GetDefaultBranchName.
func (client *Client) GetDefaultBranchName() (string, error) {
	if len(client.repourl.defaultBranch) > 0 {
		return client.repourl.defaultBranch, nil
	}

	return "", fmt.Errorf("%w", errDefaultBranchEmpty)
}

// GetBranch implements RepoClient.GetBranch.
func (client *Client) GetBranch(branch string) (*clients.BranchRef, error) {
	return client.branches.getBranch(branch)
}

// GetCreatedAt is a getter for repo.CreatedAt.
func (client *Client) GetCreatedAt() (time.Time, error) {
	return client.repo.CreatedAt.Time, nil
}

// ListWebhooks implements RepoClient.ListWebhooks.
func (client *Client) ListWebhooks() ([]clients.Webhook, error) {
	return client.webhook.listWebhooks()
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

// ListProgrammingLanguages implements RepoClient.ListProgrammingLanguages.
func (client *Client) ListProgrammingLanguages() ([]clients.Language, error) {
	return client.languages.listProgrammingLanguages()
}

// ListLicenses implements RepoClient.ListLicenses.
func (client *Client) ListLicenses() ([]clients.License, error) {
	return client.licenses.listLicenses()
}

// Search implements RepoClient.Search.
func (client *Client) Search(request clients.SearchRequest) (clients.SearchResponse, error) {
	return client.search.search(request)
}

// SearchCommits implements RepoClient.SearchCommits.
func (client *Client) SearchCommits(request clients.SearchCommitsOptions) ([]clients.Commit, error) {
	return client.searchCommits.search(request)
}

// Close implements RepoClient.Close.
func (client *Client) Close() error {
	return client.tarball.cleanup()
}

// CreateGithubRepoClientWithTransport returns a Client which implements RepoClient interface.
func CreateGithubRepoClientWithTransport(ctx context.Context, rt http.RoundTripper) clients.RepoClient {
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
			client:      client,
			graphClient: graphClient,
		},
		statuses: &statusesHandler{
			client: client,
		},
		search: &searchHandler{
			ghClient: client,
		},
		searchCommits: &searchCommitsHandler{
			ghClient: client,
		},
		webhook: &webhookHandler{
			ghClient: client,
		},
		languages: &languagesHandler{
			ghclient: client,
		},
		licenses: &licensesHandler{
			ghclient: client,
		},
		tarball: tarballHandler{
			httpClient: httpClient,
		},
	}
}

// CreateGithubRepoClient returns a Client which implements RepoClient interface.
func CreateGithubRepoClient(ctx context.Context, logger *log.Logger) clients.RepoClient {
	// Use our custom roundtripper
	rt := roundtripper.NewTransport(ctx, logger)
	return CreateGithubRepoClientWithTransport(ctx, rt)
}

// CreateOssFuzzRepoClient returns a RepoClient implementation
// intialized to `google/oss-fuzz` GitHub repository.
//
// Deprecated: Searching the github.com/google/oss-fuzz repo for projects is flawed. Use a constructor
// from clients/ossfuzz instead. https://github.com/ossf/scorecard/issues/2670
func CreateOssFuzzRepoClient(ctx context.Context, logger *log.Logger) (clients.RepoClient, error) {
	ossFuzzRepo, err := MakeGithubRepo("google/oss-fuzz")
	if err != nil {
		return nil, fmt.Errorf("error during MakeGithubRepo: %w", err)
	}

	ossFuzzRepoClient := CreateGithubRepoClient(ctx, logger)
	if err := ossFuzzRepoClient.InitRepo(ossFuzzRepo, clients.HeadSHA, 0); err != nil {
		return nil, fmt.Errorf("error during InitRepo: %w", err)
	}
	return ossFuzzRepoClient, nil
}
