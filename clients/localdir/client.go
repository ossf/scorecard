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

	"go.uber.org/zap"

	clients "github.com/ossf/scorecard/v2/clients"
)

var errUnsupportedFeature = errors.New("unsupported feature")

type localDirClient struct {
	logger *zap.Logger
	ctx    context.Context
}

// InitRepo sets up the local repo.
func (client *localDirClient) InitRepo(repo clients.Repo) error {
	// TODO
	return nil
}

// URI implements RepoClient.URI.
func (client *localDirClient) URI() string {
	// TODO
	return errUnsupportedFeature.Error()
}

// IsArchived implements RepoClient.IsArchived.
func (client *localDirClient) IsArchived() (bool, error) {
	// TODO
	return false, nil
}

// ListFiles implements RepoClient.ListFiles.
func (client *localDirClient) ListFiles(predicate func(string) (bool, error)) ([]string, error) {
	// TODO
	return nil, nil
}

// GetFileContent implements RepoClient.GetFileContent.
func (client *localDirClient) GetFileContent(filename string) ([]byte, error) {
	// TODO
	return nil, nil
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
