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
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v53/github"
	"github.com/shurcooL/githubv4"

	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/clients/githubrepo/roundtripper"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/internal/gitfile"
	"github.com/ossf/scorecard/v5/log"
)

var (
	_                     clients.RepoClient = &Client{}
	errInputRepoType                         = errors.New("input repo should be of type repoURL")
	errDefaultBranchEmpty                    = errors.New("default branch name is empty")
	errNoCodeOwners                          = errors.New("repo has no CODEOWNERS file")
)

type Option func(*repoClientConfig) error

// Client is GitHub-specific implementation of RepoClient.
type Client struct {
	repourl       *Repo
	repo          *github.Repository
	repoClient    *github.Client
	graphClient   *graphqlHandler
	contributors  *contributorsHandler
	owners        *ownersHandler
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
	git           *gitfile.Handler
	ctx           context.Context
	tarball       tarballHandler
	commitDepth   int
	gitMode       bool
}

// WithFileModeGit configures the repo client to fetch files using git.
func WithFileModeGit() Option {
	return func(c *repoClientConfig) error {
		c.gitMode = true
		return nil
	}
}

// WithRoundTripper configures the repo client to use the specified http.RoundTripper.
func WithRoundTripper(rt http.RoundTripper) Option {
	return func(c *repoClientConfig) error {
		c.rt = rt
		return nil
	}
}

type repoClientConfig struct {
	rt      http.RoundTripper
	gitMode bool
}

const defaultGhHost = "github.com"

// InitRepo sets up the GitHub repo in local storage for improving performance and GitHub token usage efficiency.
func (client *Client) InitRepo(inputRepo clients.Repo, commitSHA string, commitDepth int) error {
	ghRepo, ok := inputRepo.(*Repo)
	if !ok {
		return fmt.Errorf("%w: %v", errInputRepoType, inputRepo)
	}

	// Sanity check.
	repo, _, err := client.repoClient.Repositories.Get(client.ctx, ghRepo.owner, ghRepo.repo)
	if err != nil {
		return sce.WithMessage(sce.ErrRepoUnreachable, err.Error())
	}
	if commitDepth <= 0 {
		commitDepth = 30 // default
	}
	client.commitDepth = commitDepth
	client.repo = repo
	client.repourl = &Repo{
		owner:         repo.Owner.GetLogin(),
		repo:          repo.GetName(),
		defaultBranch: repo.GetDefaultBranch(),
		commitSHA:     commitSHA,
	}

	if client.gitMode {
		client.git.Init(client.ctx, client.repo.GetCloneURL(), commitSHA)
	} else {
		// Init tarballHandler.
		client.tarball.init(client.ctx, client.repo, commitSHA)
	}

	// Setup GraphQL.
	client.graphClient.init(client.ctx, client.repourl, client.commitDepth)

	// Setup contributorsHandler.
	client.contributors.init(client.ctx, client.repourl)

	// Setup ownersHandler.
	client.owners.init(client.ctx, client.repourl)

	// Setup branchesHandler.
	client.branches.init(client.ctx, client.repourl)

	// Setup releasesHandler.
	client.releases.init(client.ctx, client.repourl)

	// Setup workflowsHandler.
	client.workflows.init(client.ctx, client.repourl)

	// Setup checkrunsHandler.
	client.checkruns.init(client.ctx, client.repourl, client.commitDepth)

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
	host, isHost := os.LookupEnv("GH_HOST")
	if !isHost {
		host = defaultGhHost
	}
	return fmt.Sprintf("%s/%s/%s", host, client.repourl.owner, client.repourl.repo)
}

// RepoOwner implements RepoClient.RepoOwner
func (client *Client) RepoOwner() (string, error) {
	return client.repourl.owner, nil
}

// LocalPath implements RepoClient.LocalPath.
func (client *Client) LocalPath() (string, error) {
	if client.gitMode {
		path, err := client.git.GetLocalPath()
		if err != nil {
			return "", fmt.Errorf("git local path: %w", err)
		}
		return path, nil
	}
	return client.tarball.getLocalPath()
}

// ListFiles implements RepoClient.ListFiles.
func (client *Client) ListFiles(predicate func(string) (bool, error)) ([]string, error) {
	if client.gitMode {
		files, err := client.git.ListFiles(predicate)
		if err != nil {
			return nil, fmt.Errorf("git listfiles: %w", err)
		}
		return files, nil
	}
	return client.tarball.listFiles(predicate)
}

// GetFileReader implements RepoClient.GetFileReader.
func (client *Client) GetFileReader(filename string) (io.ReadCloser, error) {
	if client.gitMode {
		f, err := client.git.GetFile(filename)
		if err != nil {
			return nil, fmt.Errorf("git getfile: %w", err)
		}
		return f, nil
	}
	return client.tarball.getFile(filename)
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

// ListCodeOwners implements RepoClient.ListCodeOwners.
func (client *Client) ListCodeOwners() ([]clients.User, error) {
	var fileReader io.ReadCloser
	var err error
	for _, path := range CodeOwnerPaths {
		fileReader, err = client.GetFileReader(path)
		if err == nil {
			break
		}
	}
	if err != nil {
		return []clients.User{}, errNoCodeOwners
	}

	return client.owners.getOwners(fileReader)
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

func (client *Client) GetOrgRepoClient(ctx context.Context) (clients.RepoClient, error) {
	dotGithubRepo, err := MakeGithubRepo(fmt.Sprintf("%s/.github", client.repourl.owner))
	if err != nil {
		return nil, fmt.Errorf("error during MakeGithubRepo: %w", err)
	}

	options := []Option{WithRoundTripper(client.repoClient.Client().Transport)}
	if client.gitMode {
		options = append(options, WithFileModeGit())
	}
	c, err := NewRepoClient(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("create org repoclient: %w", err)
	}
	if err := c.InitRepo(dotGithubRepo, clients.HeadSHA, 0); err != nil {
		return nil, fmt.Errorf("error during InitRepo: %w", err)
	}

	return c, nil
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
	if client.gitMode {
		if err := client.git.Cleanup(); err != nil {
			return fmt.Errorf("git cleanup: %w", err)
		}
		return nil
	}
	return client.tarball.cleanup()
}

// CreateGithubRepoClientWithTransport returns a Client which implements RepoClient interface.
func CreateGithubRepoClientWithTransport(ctx context.Context, rt http.RoundTripper) clients.RepoClient {
	//nolint:errcheck // need to suppress because this method doesn't return an error
	rc, _ := NewRepoClient(ctx, WithRoundTripper(rt))
	return rc
}

// NewRepoClient returns a Client which implements RepoClient interface.
// It can be configured with various [Option]s.
func NewRepoClient(ctx context.Context, opts ...Option) (clients.RepoClient, error) {
	var config repoClientConfig

	for _, option := range opts {
		if err := option(&config); err != nil {
			return nil, err
		}
	}

	if config.rt == nil {
		logger := log.NewLogger(log.DefaultLevel)
		config.rt = roundtripper.NewTransport(ctx, logger)
	}

	httpClient := &http.Client{
		Transport: config.rt,
	}

	var client *github.Client
	var graphClient *githubv4.Client
	githubHost, isGhHost := os.LookupEnv("GH_HOST")

	if isGhHost && githubHost != defaultGhHost {
		githubRestURL := fmt.Sprintf("https://%s/api/v3", strings.TrimSpace(githubHost))
		githubGraphqlURL := fmt.Sprintf("https://%s/api/graphql", strings.TrimSpace(githubHost))

		var err error
		client, err = github.NewEnterpriseClient(githubRestURL, githubRestURL, httpClient)
		if err != nil {
			panic(fmt.Errorf("error during CreateGithubRepoClientWithTransport:EnterpriseClient: %w", err))
		}

		graphClient = githubv4.NewEnterpriseClient(githubGraphqlURL, httpClient)
	} else {
		client = github.NewClient(httpClient)
		graphClient = githubv4.NewClient(httpClient)
	}

	return &Client{
		ctx:        ctx,
		repoClient: client,
		graphClient: &graphqlHandler{
			client: graphClient,
		},
		contributors: &contributorsHandler{
			ghClient: client,
		},
		owners: &ownersHandler{
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
		gitMode: config.gitMode,
		git:     &gitfile.Handler{},
	}, nil
}

// CreateGithubRepoClient returns a Client which implements RepoClient interface.
func CreateGithubRepoClient(ctx context.Context, logger *log.Logger) clients.RepoClient {
	// Use our custom roundtripper
	rt := roundtripper.NewTransport(ctx, logger)
	return CreateGithubRepoClientWithTransport(ctx, rt)
}

// CreateOssFuzzRepoClient returns a RepoClient implementation
// initialized to `google/oss-fuzz` GitHub repository.
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
