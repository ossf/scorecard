// Copyright OpenSSF Authors
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

package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/ossf/scorecard/v4/clients/githubrepo/roundtripper"
	"github.com/ossf/scorecard/v4/log"
)

const (
	baseRepoURL = "https://api.github.com/repos/"
)

// RepoInfo is a struct for repository information.
type RepoInfo struct {
	Repo      repo `json:"repository"`
	respBytes []byte
}

type repo struct {
	/*
		https://docs.github.com/en/actions/reference/workflow-commands-for-github-actions#github_repository_is_fork

		GITHUB_REPOSITORY_IS_FORK is true if the repository is a fork.
	*/
	DefaultBranch string `json:"default_branch"`
	Fork          bool   `json:"fork"`
	Private       bool   `json:"private"`
}

// Client holds a context and roundtripper for querying repo info from GitHub.
type Client struct {
	ctx context.Context
	rt  http.RoundTripper
}

// NewClient returns a new Client for querying repo info from GitHub.
func NewClient(ctx context.Context) *Client {
	c := &Client{}

	defaultCtx := context.Background()
	if ctx == nil {
		ctx = defaultCtx
	}

	c.SetContext(ctx)
	c.SetDefaultTransport()
	return c
}

// SetContext sets a context for a GitHub client.
func (c *Client) SetContext(ctx context.Context) {
	c.ctx = ctx
}

// SetTransport sets a http.RoundTripper for a GitHub client.
func (c *Client) SetTransport(rt http.RoundTripper) {
	c.rt = rt
}

// SetDefaultTransport sets the scorecard roundtripper for a GitHub client.
func (c *Client) SetDefaultTransport() {
	logger := log.NewLogger(log.DefaultLevel)
	rt := roundtripper.NewTransport(c.ctx, logger)
	c.rt = rt
}

// WriteRepoInfo queries GitHub for repo info and writes it to a file.
func WriteRepoInfo(ctx context.Context, repoName, path string) error {
	c := NewClient(ctx)
	repoInfo, err := c.RepoInfo(repoName)
	if err != nil {
		return fmt.Errorf("getting repo info: %w", err)
	}

	repoFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating repo info file: %w", err)
	}
	defer repoFile.Close()

	resp := repoInfo.respBytes
	_, writeErr := repoFile.Write(resp)
	if writeErr != nil {
		return fmt.Errorf("writing repo info: %w", writeErr)
	}

	return nil
}

// RepoInfo is a function to get the repository information.
// It is decided to not use the golang GitHub library because of the
// dependency on the github.com/google/go-github/github library
// which will in turn require other dependencies.
func (c *Client) RepoInfo(repoName string) (RepoInfo, error) {
	var r RepoInfo

	baseURL, err := url.Parse(baseRepoURL)
	if err != nil {
		return r, fmt.Errorf("parsing base repo URL: %w", err)
	}

	repoURL, err := baseURL.Parse(repoName)
	if err != nil {
		return r, fmt.Errorf("parsing repo endpoint: %w", err)
	}

	method := "GET"
	req, err := http.NewRequestWithContext(
		c.ctx,
		method,
		repoURL.String(),
		nil,
	)
	if err != nil {
		return r, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return r, fmt.Errorf("error creating request: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return r, fmt.Errorf("error reading response body: %w", err)
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return r, fmt.Errorf("error reading response body: %w", err)
	}

	r.respBytes = respBytes

	err = json.Unmarshal(respBytes, &r)
	if err != nil {
		return r, fmt.Errorf("error decoding response body: %w", err)
	}

	return r, nil
}
