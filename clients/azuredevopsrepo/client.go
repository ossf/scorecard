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
	"log"
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

func (client *Client) InitRepo(inputRepo clients.Repo, commitSHA string, commitDepth int) error {
	azdoRepo, ok := inputRepo.(*Repo)
	if !ok {
		return fmt.Errorf("%w: %v", errInputRepoType, inputRepo)
	}

	repo, err := client.azdoClient.GetRepository(context.Background(), git.GetRepositoryArgs{
		Project:      &azdoRepo.project,
		RepositoryId: &azdoRepo.name,
	})
	if err != nil {
		return fmt.Errorf("could not get repository with error: %w", err)
	}

	if commitDepth <= 0 {
		client.commitDepth = 30 // default
	} else {
		client.commitDepth = commitDepth
	}

	branch, _ := strings.CutPrefix(*repo.DefaultBranch, "refs/heads/")

	client.repourl = &Repo{
		scheme:        azdoRepo.scheme,
		host:          azdoRepo.host,
		organization:  azdoRepo.organization,
		project:       azdoRepo.project,
		name:          azdoRepo.name,
		ID:            fmt.Sprint(*repo.Id),
		defaultBranch: branch,
		commitSHA:     commitSHA,
	}

	client.branches.init(client.repourl)

	client.commits.init(client.repourl, client.commitDepth)

	return nil
}

func (client *Client) URI() string {
	return fmt.Sprintf("%s/%s/_git/%s", client.repourl.host, client.repourl.organization, client.repourl.project)
}

func (client *Client) IsArchived() (bool, error) {
	repo, err := client.azdoClient.GetRepository(context.Background(), git.GetRepositoryArgs{RepositoryId: &client.repourl.ID})
	if err != nil {
		return false, fmt.Errorf("could not get repository with error: %w", err)
	}

	return *repo.IsDisabled, nil
}

func (client *Client) ListFiles(predicate func(string) (bool, error)) ([]string, error) {
	return []string{}, nil
}

func (client *Client) LocalPath() (string, error) {
	return "", nil
}

func (client *Client) GetFileReader(filename string) (io.ReadCloser, error) {
	return nil, nil
}

func (client *Client) GetBranch(branch string) (*clients.BranchRef, error) {
	return nil, nil
}

func (client *Client) GetCreatedAt() (time.Time, error) {
	return time.Time{}, nil
}

func (client *Client) GetDefaultBranchName() (string, error) {
	if len(client.repourl.defaultBranch) > 0 {
		return client.repourl.defaultBranch, nil
	}

	return "", fmt.Errorf("%s", "default branch name is empty")
}

func (client *Client) GetDefaultBranch() (*clients.BranchRef, error) {
	return client.branches.getDefaultBranch()
}

func (client *Client) GetOrgRepoClient(context.Context) (clients.RepoClient, error) {
	return nil, fmt.Errorf("GetOrgRepoClient (GitLab): %w", clients.ErrUnsupportedFeature)
}

func (client *Client) ListCommits() ([]clients.Commit, error) {
	return client.commits.listCommits()
}

func (client *Client) ListIssues() ([]clients.Issue, error) {
	return nil, nil
}

func (client *Client) ListLicenses() ([]clients.License, error) {
	return nil, nil
}

func (client *Client) ListReleases() ([]clients.Release, error) {
	return nil, nil
}

func (client *Client) ListContributors() ([]clients.User, error) {
	return nil, nil
}

func (client *Client) ListSuccessfulWorkflowRuns(filename string) ([]clients.WorkflowRun, error) {
	return nil, nil
}

func (client *Client) ListCheckRunsForRef(ref string) ([]clients.CheckRun, error) {
	return nil, nil
}

func (client *Client) ListStatuses(ref string) ([]clients.Status, error) {
	return nil, nil
}

func (client *Client) ListWebhooks() ([]clients.Webhook, error) {
	return nil, nil
}

func (client *Client) ListProgrammingLanguages() ([]clients.Language, error) {
	return nil, nil
}

func (client *Client) Search(request clients.SearchRequest) (clients.SearchResponse, error) {
	return clients.SearchResponse{}, nil
}

func (client *Client) SearchCommits(request clients.SearchCommitsOptions) ([]clients.Commit, error) {
	return nil, nil
}

func (client *Client) Close() error {
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

func CreateOssFuzzRepoClient(ctx context.Context, logger *log.Logger) (clients.RepoClient, error) {
	return nil, fmt.Errorf("%w, oss fuzz currently only supported for github repos", clients.ErrUnsupportedFeature)
}
