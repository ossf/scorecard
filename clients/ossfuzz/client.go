// Copyright 2023 OpenSSF Scorecard Authors
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

package ossfuzz

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/ossf/scorecard/v4/clients"
)

var (
	errUnreachableStatusFile = errors.New("could not fetch OSS Fuzz status file")
	errMalformedURL          = errors.New("malformed repo url")
)

type commonClient struct {
	statusURL string
}

type lazyClient struct {
	commonClient
	err      error
	contents []byte
	once     sync.Once
}

type eagerClient struct {
	projects map[string]bool
	commonClient
}

// CreateOSSFuzzClient returns a client which implements RepoClient interface.
func CreateOSSFuzzClient(ctx context.Context, ossFuzzStatusURL string) clients.RepoClient {
	return &lazyClient{
		commonClient: commonClient{
			statusURL: ossFuzzStatusURL,
		},
	}
}

type ossFuzzStatus struct {
	Projects []struct {
		RepoURI string `json:"main_repo"`
	} `json:"projects"`
}

func CreateEagerOSSFuzzClient(ctx context.Context, ossFuzzStatusURL string) (clients.RepoClient, error) {
	contents, err := fetchStatusFile(ossFuzzStatusURL)
	if err != nil {
		return nil, err
	}
	status := ossFuzzStatus{}
	if err = json.Unmarshal(contents, &status); err != nil {
		return nil, fmt.Errorf("parse status file: %w", err)
	}

	c := eagerClient{
		commonClient: commonClient{
			statusURL: ossFuzzStatusURL,
		},
		projects: map[string]bool{},
	}

	for i := range status.Projects {
		repoURI := status.Projects[i].RepoURI
		normalizedRepoURI, err := normalize(repoURI)
		if err != nil {
			continue
		}
		c.projects[normalizedRepoURI] = true
	}
	return &c, nil
}

// Search implements RepoClient.Search.
func (c *lazyClient) Search(request clients.SearchRequest) (clients.SearchResponse, error) {
	c.once.Do(func() {
		c.contents, c.err = fetchStatusFile(c.statusURL)
	})
	if c.err != nil {
		return clients.SearchResponse{}, c.err
	}
	var sr clients.SearchResponse
	projectURI := []byte(request.Query)
	if bytes.Contains(c.contents, projectURI) {
		sr.Hits = 1
	}
	return sr, nil
}

// Search implements RepoClient.Search.
func (c *eagerClient) Search(request clients.SearchRequest) (clients.SearchResponse, error) {
	var sr clients.SearchResponse
	if c.projects[request.Query] {
		sr.Hits = 1
	}
	return sr, nil
}

func fetchStatusFile(uri string) ([]byte, error) {
	//nolint:gosec // URI comes from a constant or a test HTTP server, not user input
	resp, err := http.Get(uri)
	if err != nil {
		return nil, fmt.Errorf("http.Get: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("%s: %w", resp.Status, errUnreachableStatusFile)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %w", err)
	}
	return b, nil
}

func normalize(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("url.Parse: %w", err)
	}
	const splitLen = 2
	split := strings.SplitN(strings.Trim(u.Path, "/"), "/", splitLen)
	if len(split) != splitLen {
		return "", fmt.Errorf("%s: %w", rawURL, errMalformedURL)
	}
	org := split[0]
	repo := strings.TrimSuffix(split[1], ".git")
	return fmt.Sprintf("%s/%s/%s", u.Host, org, repo), nil
}

// URI implements RepoClient.URI.
func (c *commonClient) URI() string {
	return c.statusURL
}

// InitRepo sets up the local repo.
func (c *commonClient) InitRepo(inputRepo clients.Repo, commitSHA string, commitDepth int) error {
	return fmt.Errorf("InitRepo: %w", clients.ErrUnsupportedFeature)
}

// IsArchived implements RepoClient.IsArchived.
func (c *commonClient) IsArchived() (bool, error) {
	return false, fmt.Errorf("IsArchived: %w", clients.ErrUnsupportedFeature)
}

// LocalPath implements RepoClient.LocalPath.
func (c *commonClient) LocalPath() (string, error) {
	return "", fmt.Errorf("LocalPath: %w", clients.ErrUnsupportedFeature)
}

// ListFiles implements RepoClient.ListFiles.
func (c *commonClient) ListFiles(predicate func(string) (bool, error)) ([]string, error) {
	return nil, fmt.Errorf("ListFiles: %w", clients.ErrUnsupportedFeature)
}

// GetFileContent implements RepoClient.GetFileContent.
func (c *commonClient) GetFileContent(filename string) ([]byte, error) {
	return nil, fmt.Errorf("GetFileContent: %w", clients.ErrUnsupportedFeature)
}

// GetBranch implements RepoClient.GetBranch.
func (c *commonClient) GetBranch(branch string) (*clients.BranchRef, error) {
	return nil, fmt.Errorf("GetBranch: %w", clients.ErrUnsupportedFeature)
}

// GetDefaultBranch implements RepoClient.GetDefaultBranch.
func (c *commonClient) GetDefaultBranch() (*clients.BranchRef, error) {
	return nil, fmt.Errorf("GetDefaultBranch: %w", clients.ErrUnsupportedFeature)
}

// GetDefaultBranchName implements RepoClient.GetDefaultBranchName.
func (c *commonClient) GetDefaultBranchName() (string, error) {
	return "", fmt.Errorf("GetDefaultBranchName: %w", clients.ErrUnsupportedFeature)
}

// ListCommits implements RepoClient.ListCommits.
func (c *commonClient) ListCommits() ([]clients.Commit, error) {
	return nil, fmt.Errorf("ListCommits: %w", clients.ErrUnsupportedFeature)
}

// ListIssues implements RepoClient.ListIssues.
func (c *commonClient) ListIssues() ([]clients.Issue, error) {
	return nil, fmt.Errorf("ListIssues: %w", clients.ErrUnsupportedFeature)
}

// ListReleases implements RepoClient.ListReleases.
func (c *commonClient) ListReleases() ([]clients.Release, error) {
	return nil, fmt.Errorf("ListReleases: %w", clients.ErrUnsupportedFeature)
}

// ListContributors implements RepoClient.ListContributors.
func (c *commonClient) ListContributors() ([]clients.User, error) {
	return nil, fmt.Errorf("ListContributors: %w", clients.ErrUnsupportedFeature)
}

// ListSuccessfulWorkflowRuns implements RepoClient.WorkflowRunsByFilename.
func (c *commonClient) ListSuccessfulWorkflowRuns(filename string) ([]clients.WorkflowRun, error) {
	return nil, fmt.Errorf("ListSuccessfulWorkflowRuns: %w", clients.ErrUnsupportedFeature)
}

// ListCheckRunsForRef implements RepoClient.ListCheckRunsForRef.
func (c *commonClient) ListCheckRunsForRef(ref string) ([]clients.CheckRun, error) {
	return nil, fmt.Errorf("ListCheckRunsForRef: %w", clients.ErrUnsupportedFeature)
}

// ListStatuses implements RepoClient.ListStatuses.
func (c *commonClient) ListStatuses(ref string) ([]clients.Status, error) {
	return nil, fmt.Errorf("ListStatuses: %w", clients.ErrUnsupportedFeature)
}

// ListWebhooks implements RepoClient.ListWebhooks.
func (c *commonClient) ListWebhooks() ([]clients.Webhook, error) {
	return nil, fmt.Errorf("ListWebhooks: %w", clients.ErrUnsupportedFeature)
}

// SearchCommits implements RepoClient.SearchCommits.
func (c *commonClient) SearchCommits(request clients.SearchCommitsOptions) ([]clients.Commit, error) {
	return nil, fmt.Errorf("SearchCommits: %w", clients.ErrUnsupportedFeature)
}

// Close implements RepoClient.Close.
func (c *commonClient) Close() error {
	return nil
}

// ListProgrammingLanguages implements RepoClient.ListProgrammingLanguages.
func (c *commonClient) ListProgrammingLanguages() ([]clients.Language, error) {
	return nil, fmt.Errorf("ListProgrammingLanguages: %w", clients.ErrUnsupportedFeature)
}

// ListLicenses implements RepoClient.ListLicenses.
func (c *commonClient) ListLicenses() ([]clients.License, error) {
	return nil, fmt.Errorf("ListLicenses: %w", clients.ErrUnsupportedFeature)
}

// GetCreatedAt implements RepoClient.GetCreatedAt.
func (c *commonClient) GetCreatedAt() (time.Time, error) {
	return time.Time{}, fmt.Errorf("GetCreatedAt: %w", clients.ErrUnsupportedFeature)
}
