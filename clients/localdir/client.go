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
//

// Package localdir implements RepoClient on local source code.
package localdir

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"go.uber.org/zap"

	clients "github.com/ossf/scorecard/v3/clients"
)

var (
	errUnsupportedFeature = errors.New("unsupported feature")
	errInputRepoType      = errors.New("input repo should be of type repoLocal")
)

type localDirClient struct {
	logger *zap.Logger
	ctx    context.Context
	path   string
}

// InitRepo sets up the local repo.
func (client *localDirClient) InitRepo(inputRepo clients.Repo) error {
	// TODO
	// panic("invalid InitRepo()")
	localRepo, ok := inputRepo.(*repoLocal)
	if !ok {
		return fmt.Errorf("%w: %v", errInputRepoType, inputRepo)
	}

	client.path = localRepo.Path()

	return nil
}

// URI implements RepoClient.URI.
func (client *localDirClient) URI() string {
	return fmt.Sprintf("file://%s", client.path)
}

// IsArchived implements RepoClient.IsArchived.
func (client *localDirClient) IsArchived() (bool, error) {
	// TODO
	// panic("invalid IsArchived()")
	return false, nil
}

// ListFiles implements RepoClient.ListFiles.
func (client *localDirClient) ListFiles(predicate func(string) (bool, error)) ([]string, error) {
	files := []string{}
	// fmt.Println("folder:", client.path)

	err := filepath.Walk(client.path, func(pathfn string, info fs.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failure accessing path %q: %w", pathfn, err)
		}

		cleanPath := path.Clean(pathfn)
		if cleanPath == client.path {
			// fmt.Println("skipping", cleanPath)
			return nil
		}
		prefix := fmt.Sprintf("%v%v", client.path, string(os.PathSeparator))
		p := strings.TrimPrefix(cleanPath, prefix)
		files = append(files, p)
		// fmt.Printf("visited file or dir: %q\n", p)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking the path %q: %w", client.path, err)
	}
	// fmt.Printf("%v", len(files))
	// panic("ba")
	return files, nil
}

// GetFileContent implements RepoClient.GetFileContent.
func (client *localDirClient) GetFileContent(filename string) ([]byte, error) {
	content, err := os.ReadFile(filename)
	return content, fmt.Errorf("%w", err)
}

// ListMergedPRs implements RepoClient.ListMergedPRs.
func (client *localDirClient) ListMergedPRs() ([]clients.PullRequest, error) {
	return nil, fmt.Errorf("ListMergedPRs: %w", errUnsupportedFeature)
}

// ListBranches implements RepoClient.ListBranches.
func (client *localDirClient) ListBranches() ([]*clients.BranchRef, error) {
	return nil, fmt.Errorf("ListBranches: %w", errUnsupportedFeature)
}

// GetDefaultBranch implements RepoClient.GetDefaultBranch.
func (client *localDirClient) GetDefaultBranch() (*clients.BranchRef, error) {
	return nil, fmt.Errorf("GetDefaultBranch: %w", errUnsupportedFeature)
}

func (client *localDirClient) ListCommits() ([]clients.Commit, error) {
	return nil, fmt.Errorf("ListCommits: %w", errUnsupportedFeature)
}

// ListReleases implements RepoClient.ListReleases.
func (client *localDirClient) ListReleases() ([]clients.Release, error) {
	return nil, fmt.Errorf("ListReleases: %w", errUnsupportedFeature)
}

// ListContributors implements RepoClient.ListContributors.
func (client *localDirClient) ListContributors() ([]clients.Contributor, error) {
	return nil, fmt.Errorf("ListContributors: %w", errUnsupportedFeature)
}

// ListSuccessfulWorkflowRuns implements RepoClient.WorkflowRunsByFilename.
func (client *localDirClient) ListSuccessfulWorkflowRuns(filename string) ([]clients.WorkflowRun, error) {
	return nil, fmt.Errorf("ListSuccessfulWorkflowRuns: %w", errUnsupportedFeature)
}

// ListCheckRunsForRef implements RepoClient.ListCheckRunsForRef.
func (client *localDirClient) ListCheckRunsForRef(ref string) ([]clients.CheckRun, error) {
	return nil, fmt.Errorf("ListCheckRunsForRef: %w", errUnsupportedFeature)
}

// ListStatuses implements RepoClient.ListStatuses.
func (client *localDirClient) ListStatuses(ref string) ([]clients.Status, error) {
	return nil, fmt.Errorf("ListStatuses: %w", errUnsupportedFeature)
}

// Search implements RepoClient.Search.
func (client *localDirClient) Search(request clients.SearchRequest) (clients.SearchResponse, error) {
	return clients.SearchResponse{}, fmt.Errorf("Search: %w", errUnsupportedFeature)
}

func (client *localDirClient) Close() error {
	// TODO
	return nil
}

// CreateLocalDirClient returns a client which implements RepoClient interface.
func CreateLocalDirClient(ctx context.Context, logger *zap.Logger) clients.RepoClient {
	return &localDirClient{
		ctx:    ctx,
		logger: logger,
	}
}
