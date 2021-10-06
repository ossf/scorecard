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

<<<<<<< HEAD
// Package localdir implements RepoClient on local source code.
=======
// Package localdir is local repo containing source code.
>>>>>>> 376995a (docker file)
package localdir

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

<<<<<<< HEAD
	clients "github.com/ossf/scorecard/v3/clients"
=======
	clients "github.com/ossf/scorecard/v2/clients"
>>>>>>> 376995a (docker file)
)

var errUnsupportedFeature = errors.New("unsupported feature")

type localDirClient struct {
	logger *zap.Logger
	ctx    context.Context
}

<<<<<<< HEAD
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
=======
func (client *localDirClient) InitRepo(repo clients.Repo) error {
	return nil
}

func (client *localDirClient) URL() string {
	return errUnsupportedFeature.Error()
}

func (client *localDirClient) IsArchived() (bool, error) {
	return false, nil
}

func (client *localDirClient) ListFiles(predicate func(string) (bool, error)) ([]string, error) {
	return nil, nil
}

func (client *localDirClient) GetFileContent(filename string) ([]byte, error) {
	return nil, nil
}

>>>>>>> 376995a (docker file)
func (client *localDirClient) ListMergedPRs() ([]clients.PullRequest, error) {
	return nil, fmt.Errorf("ListMergedPRs: %w", errUnsupportedFeature)
}

<<<<<<< HEAD
// ListBranches implements RepoClient.ListBranches.
=======
>>>>>>> 376995a (docker file)
func (client *localDirClient) ListBranches() ([]*clients.BranchRef, error) {
	return nil, fmt.Errorf("ListBranches: %w", errUnsupportedFeature)
}

<<<<<<< HEAD
// GetDefaultBranch implements RepoClient.GetDefaultBranch.
=======
>>>>>>> 376995a (docker file)
func (client *localDirClient) GetDefaultBranch() (*clients.BranchRef, error) {
	return nil, fmt.Errorf("GetDefaultBranch: %w", errUnsupportedFeature)
}

func (client *localDirClient) ListCommits() ([]clients.Commit, error) {
	return nil, fmt.Errorf("ListCommits: %w", errUnsupportedFeature)
}

<<<<<<< HEAD
// ListReleases implements RepoClient.ListReleases.
=======
>>>>>>> 376995a (docker file)
func (client *localDirClient) ListReleases() ([]clients.Release, error) {
	return nil, fmt.Errorf("ListReleases: %w", errUnsupportedFeature)
}

<<<<<<< HEAD
// ListContributors implements RepoClient.ListContributors.
=======
>>>>>>> 376995a (docker file)
func (client *localDirClient) ListContributors() ([]clients.Contributor, error) {
	return nil, fmt.Errorf("ListContributors: %w", errUnsupportedFeature)
}

<<<<<<< HEAD
// ListSuccessfulWorkflowRuns implements RepoClient.WorkflowRunsByFilename.
=======
>>>>>>> 376995a (docker file)
func (client *localDirClient) ListSuccessfulWorkflowRuns(filename string) ([]clients.WorkflowRun, error) {
	return nil, fmt.Errorf("ListSuccessfulWorkflowRuns: %w", errUnsupportedFeature)
}

<<<<<<< HEAD
// ListCheckRunsForRef implements RepoClient.ListCheckRunsForRef.
=======
>>>>>>> 376995a (docker file)
func (client *localDirClient) ListCheckRunsForRef(ref string) ([]clients.CheckRun, error) {
	return nil, fmt.Errorf("ListCheckRunsForRef: %w", errUnsupportedFeature)
}

<<<<<<< HEAD
// ListStatuses implements RepoClient.ListStatuses.
=======
>>>>>>> 376995a (docker file)
func (client *localDirClient) ListStatuses(ref string) ([]clients.Status, error) {
	return nil, fmt.Errorf("ListStatuses: %w", errUnsupportedFeature)
}

<<<<<<< HEAD
// Search implements RepoClient.Search.
=======
>>>>>>> 376995a (docker file)
func (client *localDirClient) Search(request clients.SearchRequest) (clients.SearchResponse, error) {
	return clients.SearchResponse{}, fmt.Errorf("Search: %w", errUnsupportedFeature)
}

func (client *localDirClient) Close() error {
<<<<<<< HEAD
	// TODO
=======
>>>>>>> 376995a (docker file)
	return nil
}

// CreateLocalDirClient returns a client which implements RepoClient interface.
func CreateLocalDirClient(ctx context.Context, logger *zap.Logger) clients.RepoClient {
	return &localDirClient{
		ctx:    ctx,
		logger: logger,
	}
}
