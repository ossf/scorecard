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
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/ossf/scorecard/v4/clients"
)

type ossFuzzClient struct {
	statusURL string
	err       error
	contents  []byte
	once      sync.Once
}

// CreateOSSFuzzClient returns a client which implements RepoClient interface.
func CreateOSSFuzzClient(ctx context.Context, ossFuzzStatusURL string) clients.RepoClient {
	return &ossFuzzClient{
		statusURL: ossFuzzStatusURL,
	}
}

// Search implements RepoClient.Search.
func (c *ossFuzzClient) Search(request clients.SearchRequest) (clients.SearchResponse, error) {
	c.once.Do(func() {
		c.contents, c.err = parseStatusJSON(c.statusURL)
	})
	if c.err != nil {
		return clients.SearchResponse{}, c.err
	}
	projectURI := []byte(request.Query)
	sr := clients.SearchResponse{}
	if bytes.Contains(c.contents, projectURI) {
		sr.Hits = 1
	}
	return sr, nil
}

func parseStatusJSON(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http.Get: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("fetch OSS-Fuzz project list: %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}

// URI implements RepoClient.URI.
func (c *ossFuzzClient) URI() string {
	return c.statusURL
}

// InitRepo sets up the local repo.
func (c *ossFuzzClient) InitRepo(inputRepo clients.Repo, commitSHA string, commitDepth int) error {
	return fmt.Errorf("InitRepo: %w", clients.ErrUnsupportedFeature)
}

// IsArchived implements RepoClient.IsArchived.
func (c *ossFuzzClient) IsArchived() (bool, error) {
	return false, fmt.Errorf("IsArchived: %w", clients.ErrUnsupportedFeature)
}

// LocalPath implements RepoClient.LocalPath.
func (c *ossFuzzClient) LocalPath() (string, error) {
	return "", fmt.Errorf("LocalPath: %w", clients.ErrUnsupportedFeature)
}

// ListFiles implements RepoClient.ListFiles.
func (c *ossFuzzClient) ListFiles(predicate func(string) (bool, error)) ([]string, error) {
	return nil, fmt.Errorf("ListFiles: %w", clients.ErrUnsupportedFeature)
}

// GetFileContent implements RepoClient.GetFileContent.
func (c *ossFuzzClient) GetFileContent(filename string) ([]byte, error) {
	return nil, fmt.Errorf("GetFileContent: %w", clients.ErrUnsupportedFeature)
}

// GetBranch implements RepoClient.GetBranch.
func (c *ossFuzzClient) GetBranch(branch string) (*clients.BranchRef, error) {
	return nil, fmt.Errorf("GetBranch: %w", clients.ErrUnsupportedFeature)
}

// GetDefaultBranch implements RepoClient.GetDefaultBranch.
func (c *ossFuzzClient) GetDefaultBranch() (*clients.BranchRef, error) {
	return nil, fmt.Errorf("GetDefaultBranch: %w", clients.ErrUnsupportedFeature)
}

// GetDefaultBranchName implements RepoClient.GetDefaultBranchName.
func (c *ossFuzzClient) GetDefaultBranchName() (string, error) {
	return "", fmt.Errorf("GetDefaultBranchName: %w", clients.ErrUnsupportedFeature)
}

// ListCommits implements RepoClient.ListCommits.
func (c *ossFuzzClient) ListCommits() ([]clients.Commit, error) {
	return nil, fmt.Errorf("ListCommits: %w", clients.ErrUnsupportedFeature)
}

// ListIssues implements RepoClient.ListIssues.
func (c *ossFuzzClient) ListIssues() ([]clients.Issue, error) {
	return nil, fmt.Errorf("ListIssues: %w", clients.ErrUnsupportedFeature)
}

// ListReleases implements RepoClient.ListReleases.
func (c *ossFuzzClient) ListReleases() ([]clients.Release, error) {
	return nil, fmt.Errorf("ListReleases: %w", clients.ErrUnsupportedFeature)
}

// ListContributors implements RepoClient.ListContributors.
func (c *ossFuzzClient) ListContributors() ([]clients.User, error) {
	return nil, fmt.Errorf("ListContributors: %w", clients.ErrUnsupportedFeature)
}

// ListSuccessfulWorkflowRuns implements RepoClient.WorkflowRunsByFilename.
func (c *ossFuzzClient) ListSuccessfulWorkflowRuns(filename string) ([]clients.WorkflowRun, error) {
	return nil, fmt.Errorf("ListSuccessfulWorkflowRuns: %w", clients.ErrUnsupportedFeature)
}

// ListCheckRunsForRef implements RepoClient.ListCheckRunsForRef.
func (c *ossFuzzClient) ListCheckRunsForRef(ref string) ([]clients.CheckRun, error) {
	return nil, fmt.Errorf("ListCheckRunsForRef: %w", clients.ErrUnsupportedFeature)
}

// ListStatuses implements RepoClient.ListStatuses.
func (c *ossFuzzClient) ListStatuses(ref string) ([]clients.Status, error) {
	return nil, fmt.Errorf("ListStatuses: %w", clients.ErrUnsupportedFeature)
}

// ListWebhooks implements RepoClient.ListWebhooks.
func (c *ossFuzzClient) ListWebhooks() ([]clients.Webhook, error) {
	return nil, fmt.Errorf("ListWebhooks: %w", clients.ErrUnsupportedFeature)
}

// SearchCommits implements RepoClient.SearchCommits.
func (c *ossFuzzClient) SearchCommits(request clients.SearchCommitsOptions) ([]clients.Commit, error) {
	return nil, fmt.Errorf("SearchCommits: %w", clients.ErrUnsupportedFeature)
}

// Close implements RepoClient.Close.
func (c *ossFuzzClient) Close() error {
	return nil
}

// ListProgrammingLanguages implements RepoClient.ListProgrammingLanguages.
func (c *ossFuzzClient) ListProgrammingLanguages() ([]clients.Language, error) {
	return nil, fmt.Errorf("ListProgrammingLanguages: %w", clients.ErrUnsupportedFeature)
}

// ListLicenses implements RepoClient.ListLicenses.
func (c *ossFuzzClient) ListLicenses() ([]clients.License, error) {
	return nil, fmt.Errorf("ListLicenses: %w", clients.ErrUnsupportedFeature)
}

// GetCreatedAt implements RepoClient.GetCreatedAt.
func (c *ossFuzzClient) GetCreatedAt() (time.Time, error) {
	return time.Time{}, fmt.Errorf("GetCreatedAt: %w", clients.ErrUnsupportedFeature)
}
