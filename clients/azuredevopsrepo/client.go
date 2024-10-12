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
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/ossf/scorecard/v5/clients"
)

var (
	_                clients.RepoClient = &Client{}
	errInputRepoType                    = errors.New("input repo should be of type repoURL")
)

type Client struct {
	repourl     *Repo
	azdoClient  git.Client
	branches    *branchesHandler
	commits     *commitsHandler
	ctx         context.Context
	commitDepth int
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
		ID:            fmt.Sprint(*repo.Id),
		defaultBranch: branch,
		commitSHA:     commitSHA,
	}

	c.branches.init(c.ctx, c.repourl)

	c.commits.init(c.ctx, c.repourl, c.commitDepth)

	return nil
}

func (c *Client) URI() string {
	return fmt.Sprintf("%s/%s", c.repourl.host, c.repourl.Path())
}

func (c *Client) IsArchived() (bool, error) {
	repo, err := c.azdoClient.GetRepository(c.ctx, git.GetRepositoryArgs{RepositoryId: &c.repourl.ID})
	if err != nil {
		return false, fmt.Errorf("could not get repository with error: %w", err)
	}

	return *repo.IsDisabled, nil
}

func (c *Client) ListFiles(predicate func(string) (bool, error)) ([]string, error) {
	return []string{}, clients.ErrUnsupportedFeature
}

func (c *Client) LocalPath() (string, error) {
	return "", clients.ErrUnsupportedFeature
}

func (c *Client) GetFileReader(filename string) (io.ReadCloser, error) {
	return nil, clients.ErrUnsupportedFeature
}

func (c *Client) GetBranch(branch string) (*clients.BranchRef, error) {
	return nil, clients.ErrUnsupportedFeature
}

func (c *Client) GetCreatedAt() (time.Time, error) {
	return time.Time{}, clients.ErrUnsupportedFeature
}

func (c *Client) GetDefaultBranchName() (string, error) {
	if len(c.repourl.defaultBranch) > 0 {
		return c.repourl.defaultBranch, nil
	}

	return "", fmt.Errorf("%s", "default branch name is empty")
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
	return nil, clients.ErrUnsupportedFeature
}

func (c *Client) ListLicenses() ([]clients.License, error) {
	return nil, clients.ErrUnsupportedFeature
}

func (c *Client) ListReleases() ([]clients.Release, error) {
	return nil, clients.ErrUnsupportedFeature
}

func (c *Client) ListContributors() ([]clients.User, error) {
	return nil, clients.ErrUnsupportedFeature
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
	return nil, clients.ErrUnsupportedFeature
}

func (c *Client) Search(request clients.SearchRequest) (clients.SearchResponse, error) {
	return clients.SearchResponse{}, clients.ErrUnsupportedFeature
}

func (c *Client) SearchCommits(request clients.SearchCommitsOptions) ([]clients.Commit, error) {
	return nil, clients.ErrUnsupportedFeature
}

func (c *Client) Close() error {
	return nil
}

func CreateAzureDevOpsClient(ctx context.Context, repo clients.Repo) (*Client, error) {
	token := os.Getenv("AZURE_DEVOPS_AUTH_TOKEN")
	return CreateAzureDevOpsClientWithToken(ctx, token, repo)
}

func CreateAzureDevOpsClientWithToken(ctx context.Context, token string, repo clients.Repo) (*Client, error) {
	// https://dev.azure.com/<org>
	url := "https://" + repo.Host() + "/" + strings.Split(repo.Path(), "/")[0]
	connection := azuredevops.NewPatConnection(url, token)

	gitClient, err := git.NewClient(ctx, connection)
	if err != nil {
		return nil, fmt.Errorf("could not create azure devops git client with error: %w", err)
	}

	return &Client{
		ctx:        ctx,
		azdoClient: gitClient,
		branches: &branchesHandler{
			gitClient: gitClient,
		},
		commits: &commitsHandler{
			gitClient: gitClient,
		},
	}, nil
}
