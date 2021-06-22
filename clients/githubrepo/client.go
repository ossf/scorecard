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
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v32/github"

	"github.com/ossf/scorecard/clients"
)

const repoFilename = "githubrepo*.tar.gz"

type Client struct {
	repo       *github.Repository
	repoClient *github.Client
	ctx        context.Context
	owner      string
	repoName   string
	tarball    string
}

func (client *Client) InitRepo(owner, repoName string) error {
	client.owner = owner
	client.repoName = repoName

	// Remove older files so the function can be called multiple times.
	// Note: tarball may be the empty string on first call.
	err := os.Remove(client.tarball)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("os.Remove: %w", err)
	}

	repo, _, err := client.repoClient.Repositories.Get(client.ctx, client.owner, client.repoName)
	if err != nil {
		// nolint: wrapcheck
		return clients.NewRepoUnavailableError(err)
	}
	client.repo = repo

	url := client.repo.GetArchiveURL()
	url = strings.Replace(url, "{archive_format}", "tarball/", 1)
	url = strings.Replace(url, "{/ref}", client.repo.GetDefaultBranch(), 1)
	req, err := http.NewRequestWithContext(client.ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("http.NewRequestWithContext: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http.DefaultClient.Do: %w", err)
	}
	defer resp.Body.Close()

	// Create a temp file. This automaticlly appends a random number to the name.
	repoFile, err := ioutil.TempFile("", repoFilename)
	if err != nil {
		return fmt.Errorf("ioutil.TempFile: %w", err)
	}
	defer repoFile.Close()

	client.tarball = repoFile.Name()

	if _, err := io.Copy(repoFile, resp.Body); err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}
	return nil
}

func (client *Client) GetRepoArchiveReader() (io.ReadCloser, error) {
	archiveReader, err := os.OpenFile(client.tarball, os.O_RDONLY, 0o644)
	if err != nil {
		return archiveReader, fmt.Errorf("os.OpenFile: %w", err)
	}
	return archiveReader, nil
}

func CreateGithubRepoClient(ctx context.Context, client *github.Client) clients.RepoClient {
	return &Client{
		ctx:        ctx,
		repoClient: client,
	}
}
