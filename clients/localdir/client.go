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
//

// Package localdir implements RepoClient on local source code.
package localdir

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	clients "github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/log"
)

var (
	_                clients.RepoClient = &Client{}
	errInputRepoType                    = errors.New("input repo should be of type repoLocal")
)

//nolint:govet
type Client struct {
	logger      *log.Logger
	ctx         context.Context
	path        string
	once        sync.Once
	errFiles    error
	files       []string
	commitDepth int
}

// InitRepo sets up the local repo.
func (client *Client) InitRepo(inputRepo clients.Repo, commitSHA string, commitDepth int) error {
	localRepo, ok := inputRepo.(*Repo)
	if !ok {
		return fmt.Errorf("%w: %v", errInputRepoType, inputRepo)
	}
	if commitDepth <= 0 {
		client.commitDepth = 30 // default
	} else {
		client.commitDepth = commitDepth
	}
	client.path = strings.TrimPrefix(localRepo.URI(), "file://")

	return nil
}

// URI implements RepoClient.URI.
func (client *Client) URI() string {
	return fmt.Sprintf("file://%s", client.path)
}

// IsArchived implements RepoClient.IsArchived.
func (client *Client) IsArchived() (bool, error) {
	return false, fmt.Errorf("IsArchived: %w", clients.ErrUnsupportedFeature)
}

func isDir(p string) (bool, error) {
	fileInfo, err := os.Stat(p)
	if err != nil {
		return false, fmt.Errorf("%w", err)
	}
	return fileInfo.IsDir(), nil
}

func trimPrefix(pathfn, clientPath string) string {
	cleanPath := path.Clean(pathfn)
	prefix := fmt.Sprintf("%s%s", clientPath, string(os.PathSeparator))
	return strings.TrimPrefix(cleanPath, prefix)
}

func listFiles(clientPath string) ([]string, error) {
	files := []string{}
	err := filepath.Walk(clientPath, func(pathfn string, info fs.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failure accessing path %q: %w", pathfn, err)
		}

		// Skip directories.
		d, err := isDir(pathfn)
		if err != nil {
			return err
		}
		if d {
			// Check if the directory is .git. Use filepath.Base for compatibility across different OS path separators.
			// ignoring the .git folder.
			if filepath.Base(pathfn) == ".git" {
				return fs.SkipDir
			}
			return nil
		}

		// Remove prefix of the folder.
		p := trimPrefix(pathfn, clientPath)
		files = append(files, p)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking the path %q: %w", clientPath, err)
	}

	return files, nil
}

func applyPredicate(
	clientFiles []string,
	errFiles error,
	predicate func(string) (bool, error),
) ([]string, error) {
	if errFiles != nil {
		return nil, errFiles
	}

	files := []string{}
	for _, pathfn := range clientFiles {
		matches, err := predicate(pathfn)
		if err != nil {
			return nil, err
		}

		if matches {
			files = append(files, pathfn)
		}
	}

	return files, nil
}

// LocalPath implements RepoClient.LocalPath.
func (client *Client) LocalPath() (string, error) {
	clientPath, err := filepath.Abs(client.path)
	if err != nil {
		return "", fmt.Errorf("error during filepath.Abs: %w", err)
	}
	return clientPath, nil
}

// ListFiles implements RepoClient.ListFiles.
func (client *Client) ListFiles(predicate func(string) (bool, error)) ([]string, error) {
	client.once.Do(func() {
		client.files, client.errFiles = listFiles(client.path)
	})
	return applyPredicate(client.files, client.errFiles, predicate)
}

func getFile(clientpath, filename string) (*os.File, error) {
	// Note: the filenames do not contain the original path - see ListFiles().
	fn := path.Join(clientpath, filename)
	f, err := os.Open(fn)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	return f, nil
}

// GetFileReader implements RepoClient.GetFileReader.
func (client *Client) GetFileReader(filename string) (io.ReadCloser, error) {
	return getFile(client.path, filename)
}

// GetBranch implements RepoClient.GetBranch.
func (client *Client) GetBranch(branch string) (*clients.BranchRef, error) {
	return nil, fmt.Errorf("ListBranches: %w", clients.ErrUnsupportedFeature)
}

// GetDefaultBranch implements RepoClient.GetDefaultBranch.
func (client *Client) GetDefaultBranch() (*clients.BranchRef, error) {
	return nil, fmt.Errorf("GetDefaultBranch: %w", clients.ErrUnsupportedFeature)
}

// GetDefaultBranchName implements RepoClient.GetDefaultBranchName.
func (client *Client) GetDefaultBranchName() (string, error) {
	return "", fmt.Errorf("GetDefaultBranchName: %w", clients.ErrUnsupportedFeature)
}

// ListCommits implements RepoClient.ListCommits.
func (client *Client) ListCommits() ([]clients.Commit, error) {
	return nil, fmt.Errorf("ListCommits: %w", clients.ErrUnsupportedFeature)
}

// ListIssues implements RepoClient.ListIssues.
func (client *Client) ListIssues() ([]clients.Issue, error) {
	return nil, fmt.Errorf("ListIssues: %w", clients.ErrUnsupportedFeature)
}

// ListReleases implements RepoClient.ListReleases.
func (client *Client) ListReleases() ([]clients.Release, error) {
	return nil, fmt.Errorf("ListReleases: %w", clients.ErrUnsupportedFeature)
}

// ListContributors implements RepoClient.ListContributors.
func (client *Client) ListContributors() ([]clients.User, error) {
	return nil, fmt.Errorf("ListContributors: %w", clients.ErrUnsupportedFeature)
}

// ListSuccessfulWorkflowRuns implements RepoClient.WorkflowRunsByFilename.
func (client *Client) ListSuccessfulWorkflowRuns(filename string) ([]clients.WorkflowRun, error) {
	return nil, fmt.Errorf("ListSuccessfulWorkflowRuns: %w", clients.ErrUnsupportedFeature)
}

// ListCheckRunsForRef implements RepoClient.ListCheckRunsForRef.
func (client *Client) ListCheckRunsForRef(ref string) ([]clients.CheckRun, error) {
	return nil, fmt.Errorf("ListCheckRunsForRef: %w", clients.ErrUnsupportedFeature)
}

// ListStatuses implements RepoClient.ListStatuses.
func (client *Client) ListStatuses(ref string) ([]clients.Status, error) {
	return nil, fmt.Errorf("ListStatuses: %w", clients.ErrUnsupportedFeature)
}

// ListWebhooks implements RepoClient.ListWebhooks.
func (client *Client) ListWebhooks() ([]clients.Webhook, error) {
	return nil, fmt.Errorf("ListWebhooks: %w", clients.ErrUnsupportedFeature)
}

// Search implements RepoClient.Search.
func (client *Client) Search(request clients.SearchRequest) (clients.SearchResponse, error) {
	return clients.SearchResponse{}, fmt.Errorf("Search: %w", clients.ErrUnsupportedFeature)
}

// SearchCommits implements RepoClient.SearchCommits.
func (client *Client) SearchCommits(request clients.SearchCommitsOptions) ([]clients.Commit, error) {
	return nil, fmt.Errorf("Search: %w", clients.ErrUnsupportedFeature)
}

func (client *Client) Close() error {
	return nil
}

// ListProgrammingLanguages implements RepoClient.ListProgrammingLanguages.
// TODO: add ListProgrammingLanguages support for local directories.
func (client *Client) ListProgrammingLanguages() ([]clients.Language, error) {
	// for now just return all programming languages
	return []clients.Language{{Name: clients.All, NumLines: 1}}, nil
}

// ListLicenses implements RepoClient.ListLicenses.
// TODO: add ListLicenses support for local directories.
func (client *Client) ListLicenses() ([]clients.License, error) {
	return nil, fmt.Errorf("ListLicenses: %w", clients.ErrUnsupportedFeature)
}

func (client *Client) GetCreatedAt() (time.Time, error) {
	return time.Time{}, fmt.Errorf("GetCreatedAt: %w", clients.ErrUnsupportedFeature)
}

func (client *Client) GetOrgRepoClient(ctx context.Context) (clients.RepoClient, error) {
	return nil, fmt.Errorf("GetOrgRepoClient: %w", clients.ErrUnsupportedFeature)
}

// CreateLocalDirClient returns a client which implements RepoClient interface.
func CreateLocalDirClient(ctx context.Context, logger *log.Logger) clients.RepoClient {
	return &Client{
		ctx:    ctx,
		logger: logger,
	}
}
