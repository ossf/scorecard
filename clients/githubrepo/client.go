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
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v32/github"

	"github.com/ossf/scorecard/clients"
)

const repoFilename = "./githubrepo.tar.gz"

type Client struct {
	repo       *github.Repository
	repoClient *github.Client
	ctx        context.Context
	owner      string
	repoName   string
}

func (client *Client) InitRepo(owner, repoName string) error {
	client.owner = owner
	client.repoName = repoName
	repo, _, err := client.repoClient.Repositories.Get(client.ctx, client.owner, client.repoName)
	if err != nil {
		return fmt.Errorf("error during Repositories.Get: %w", err)
	}
	client.repo = repo

	url := client.repo.GetArchiveURL()
	url = strings.Replace(url, "{archive_format}", "tarball/", 1)
	url = strings.Replace(url, "{/ref}", client.repo.GetDefaultBranch(), 1)
	req, err := http.NewRequestWithContext(client.ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("error during http.NewRequestWithContext: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error during HTTP call: %w", err)
	}
	defer resp.Body.Close()

	repoFile, err := os.OpenFile(repoFilename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("error opening file %s for write: %w", repoFilename, err)
	}
	if _, err := io.Copy(repoFile, resp.Body); err != nil {
		return fmt.Errorf("error during io.Copy: %w", err)
	}
	if err := repoFile.Close(); err != nil {
		return fmt.Errorf("error during file Close: %w", err)
	}
	return nil
}

func (client *Client) GetRepoArchiveReader() (io.ReadCloser, error) {
	archiveReader, err := os.OpenFile(repoFilename, os.O_RDONLY, 0o644)
	if err != nil {
		return archiveReader, fmt.Errorf("error opening file %s for read: %w", repoFilename, err)
	}
	return archiveReader, nil
}

func CreateGithubRepoClient(ctx context.Context, client *github.Client) clients.RepoClient {
	return &Client{
		ctx:        ctx,
		repoClient: client,
	}
}
