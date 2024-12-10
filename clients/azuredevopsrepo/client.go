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
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/audit"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/projectanalysis"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/search"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"

	"github.com/ossf/scorecard/v5/clients"
)

var (
	_                        clients.RepoClient = &Client{}
	errInputRepoType                            = errors.New("input repo should be of type azuredevopsrepo.Repo")
	errDefaultBranchNotFound                    = errors.New("default branch not found")
)

type Client struct {
	azdoClient   git.Client
	ctx          context.Context
	repourl      *Repo
	repo         *git.GitRepository
	audit        *auditHandler
	branches     *branchesHandler
	commits      *commitsHandler
	contributors *contributorsHandler
	languages    *languagesHandler
	search       *searchHandler
	workItems    *workItemsHandler
	zip          *zipHandler
	commitDepth  int
}

func (c *Client) InitRepo(inputRepo clients.Repo, commitSHA string, commitDepth int) error {
	azdoRepo, ok := inputRepo.(*Repo)
	if !ok {
		return fmt.Errorf("%w: %v", errInputRepoType, inputRepo)
	}

	repo, err := c.azdoClient.GetRepository(c.ctx, git.GetRepositoryArgs{
		Project:      &azdoRepo.project,
		RepositoryId: &azdoRepo.name,
	})
	if err != nil {
		return fmt.Errorf("could not get repository with error: %w", err)
	}

	c.repo = repo

	if commitDepth <= 0 {
		c.commitDepth = 30 // default
	} else {
		c.commitDepth = commitDepth
	}

	branch := strings.TrimPrefix(*repo.DefaultBranch, "refs/heads/")

	c.repourl = &Repo{
		scheme:        azdoRepo.scheme,
		host:          azdoRepo.host,
		organization:  azdoRepo.organization,
		project:       azdoRepo.project,
		name:          azdoRepo.name,
		id:            fmt.Sprint(*repo.Id),
		defaultBranch: branch,
		commitSHA:     commitSHA,
	}

	c.audit.init(c.ctx, c.repourl)

	c.branches.init(c.ctx, c.repourl)

	c.commits.init(c.ctx, c.repourl, c.commitDepth)

	c.contributors.init(c.ctx, c.repourl)

	c.languages.init(c.ctx, c.repourl)

	c.search.init(c.ctx, c.repourl)

	c.workItems.init(c.ctx, c.repourl)

	c.zip.init(c.ctx, c.repourl)

	return nil
}

func (c *Client) URI() string {
	return c.repourl.URI()
}

func (c *Client) IsArchived() (bool, error) {
	return *c.repo.IsDisabled, nil
}

func (c *Client) ListFiles(predicate func(string) (bool, error)) ([]string, error) {
	return c.zip.listFiles(predicate)
}

func (c *Client) LocalPath() (string, error) {
	return c.zip.getLocalPath()
}

func (c *Client) GetFileReader(filename string) (io.ReadCloser, error) {
	return c.zip.getFile(filename)
}

func (c *Client) GetBranch(branch string) (*clients.BranchRef, error) {
	return nil, clients.ErrUnsupportedFeature
}

func (c *Client) GetCreatedAt() (time.Time, error) {
	createdAt, err := c.audit.getRepsitoryCreatedAt()
	if err != nil {
		return time.Time{}, err
	}

	// The audit log may not be enabled on the repository
	if createdAt.IsZero() {
		return c.commits.getFirstCommitCreatedAt()
	}
	return createdAt, nil
}

func (c *Client) GetDefaultBranchName() (string, error) {
	if len(c.repourl.defaultBranch) > 0 {
		return c.repourl.defaultBranch, nil
	}

	return "", errDefaultBranchNotFound
}

func (c *Client) GetDefaultBranch() (*clients.BranchRef, error) {
	return c.branches.getDefaultBranch()
}

func (c *Client) GetOrgRepoClient(context.Context) (clients.RepoClient, error) {
	return nil, clients.ErrUnsupportedFeature
}

func (c *Client) ListCommits() ([]clients.Commit, error) {
	return c.commits.listCommits()
}

func (c *Client) ListIssues() ([]clients.Issue, error) {
	return c.workItems.listIssues()
}

func (c *Client) ListLicenses() ([]clients.License, error) {
	return nil, clients.ErrUnsupportedFeature
}

func (c *Client) ListReleases() ([]clients.Release, error) {
	return nil, clients.ErrUnsupportedFeature
}

func (c *Client) ListContributors() ([]clients.User, error) {
	return c.contributors.listContributors()
}

func (c *Client) ListSuccessfulWorkflowRuns(filename string) ([]clients.WorkflowRun, error) {
	return nil, clients.ErrUnsupportedFeature
}

func (c *Client) ListCheckRunsForRef(ref string) ([]clients.CheckRun, error) {
	return nil, clients.ErrUnsupportedFeature
}

func (c *Client) ListStatuses(ref string) ([]clients.Status, error) {
	return nil, clients.ErrUnsupportedFeature
}

func (c *Client) ListWebhooks() ([]clients.Webhook, error) {
	return nil, clients.ErrUnsupportedFeature
}

func (c *Client) ListProgrammingLanguages() ([]clients.Language, error) {
	return c.languages.listProgrammingLanguages()
}

func (c *Client) Search(request clients.SearchRequest) (clients.SearchResponse, error) {
	return c.search.search(request)
}

func (c *Client) SearchCommits(request clients.SearchCommitsOptions) ([]clients.Commit, error) {
	return nil, clients.ErrUnsupportedFeature
}

func (c *Client) Close() error {
	return c.zip.cleanup()
}

func CreateAzureDevOpsClient(ctx context.Context, repo clients.Repo) (*Client, error) {
	token := os.Getenv("AZURE_DEVOPS_AUTH_TOKEN")
	return CreateAzureDevOpsClientWithToken(ctx, token, repo)
}

func CreateAzureDevOpsClientWithToken(ctx context.Context, token string, repo clients.Repo) (*Client, error) {
	// https://dev.azure.com/<org>
	url := "https://" + repo.Host() + "/" + strings.Split(repo.Path(), "/")[0]
	connection := azuredevops.NewPatConnection(url, token)

	client := connection.GetClientByUrl(url)

	auditClient, err := audit.NewClient(ctx, connection)
	if err != nil {
		return nil, fmt.Errorf("could not create azure devops audit client with error: %w", err)
	}

	gitClient, err := git.NewClient(ctx, connection)
	if err != nil {
		return nil, fmt.Errorf("could not create azure devops git client with error: %w", err)
	}

	projectAnalysisClient, err := projectanalysis.NewClient(ctx, connection)
	if err != nil {
		return nil, fmt.Errorf("could not create azure devops project analysis client with error: %w", err)
	}

	searchClient, err := search.NewClient(ctx, connection)
	if err != nil {
		return nil, fmt.Errorf("could not create azure devops search client with error: %w", err)
	}

	workItemsClient, err := workitemtracking.NewClient(ctx, connection)
	if err != nil {
		return nil, fmt.Errorf("could not create azure devops work item tracking client with error: %w", err)
	}

	return &Client{
		ctx:        ctx,
		azdoClient: gitClient,
		audit: &auditHandler{
			auditClient: auditClient,
		},
		branches: &branchesHandler{
			gitClient: gitClient,
		},
		commits: &commitsHandler{
			gitClient: gitClient,
		},
		contributors: &contributorsHandler{
			gitClient: gitClient,
		},
		languages: &languagesHandler{
			projectAnalysisClient: projectAnalysisClient,
		},
		search: &searchHandler{
			searchClient: searchClient,
		},
		workItems: &workItemsHandler{
			workItemsClient: workItemsClient,
		},
		zip: &zipHandler{
			client: client,
		},
	}, nil
}
