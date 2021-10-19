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
<<<<<<< HEAD
// Package localdir implements RepoClient on local source code.
=======
// Package localdir is local repo containing source code.
>>>>>>> 376995a (docker file)
=======
// Package localdir implements RepoClient on local source code.
>>>>>>> 251f88d (comments)
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
=======
// InitRepo sets up the local repo.
>>>>>>> 251f88d (comments)
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

<<<<<<< HEAD
>>>>>>> 376995a (docker file)
=======
// ListMergedPRs implements RepoClient.ListMergedPRs.
>>>>>>> 251f88d (comments)
func (client *localDirClient) ListMergedPRs() ([]clients.PullRequest, error) {
	return nil, fmt.Errorf("ListMergedPRs: %w", errUnsupportedFeature)
}

<<<<<<< HEAD
<<<<<<< HEAD
// ListBranches implements RepoClient.ListBranches.
=======
>>>>>>> 376995a (docker file)
=======
// ListBranches implements RepoClient.ListBranches.
>>>>>>> 251f88d (comments)
func (client *localDirClient) ListBranches() ([]*clients.BranchRef, error) {
	return nil, fmt.Errorf("ListBranches: %w", errUnsupportedFeature)
}

<<<<<<< HEAD
<<<<<<< HEAD
// GetDefaultBranch implements RepoClient.GetDefaultBranch.
=======
>>>>>>> 376995a (docker file)
=======
// GetDefaultBranch implements RepoClient.GetDefaultBranch.
>>>>>>> 251f88d (comments)
func (client *localDirClient) GetDefaultBranch() (*clients.BranchRef, error) {
	return nil, fmt.Errorf("GetDefaultBranch: %w", errUnsupportedFeature)
}

func (client *localDirClient) ListCommits() ([]clients.Commit, error) {
	return nil, fmt.Errorf("ListCommits: %w", errUnsupportedFeature)
}

<<<<<<< HEAD
<<<<<<< HEAD
// ListReleases implements RepoClient.ListReleases.
=======
>>>>>>> 376995a (docker file)
=======
// ListReleases implements RepoClient.ListReleases.
>>>>>>> 251f88d (comments)
func (client *localDirClient) ListReleases() ([]clients.Release, error) {
	return nil, fmt.Errorf("ListReleases: %w", errUnsupportedFeature)
}

<<<<<<< HEAD
<<<<<<< HEAD
// ListContributors implements RepoClient.ListContributors.
=======
>>>>>>> 376995a (docker file)
=======
// ListContributors implements RepoClient.ListContributors.
>>>>>>> 251f88d (comments)
func (client *localDirClient) ListContributors() ([]clients.Contributor, error) {
	return nil, fmt.Errorf("ListContributors: %w", errUnsupportedFeature)
}

<<<<<<< HEAD
<<<<<<< HEAD
// ListSuccessfulWorkflowRuns implements RepoClient.WorkflowRunsByFilename.
=======
>>>>>>> 376995a (docker file)
=======
// ListSuccessfulWorkflowRuns implements RepoClient.WorkflowRunsByFilename.
>>>>>>> 251f88d (comments)
func (client *localDirClient) ListSuccessfulWorkflowRuns(filename string) ([]clients.WorkflowRun, error) {
	return nil, fmt.Errorf("ListSuccessfulWorkflowRuns: %w", errUnsupportedFeature)
}

<<<<<<< HEAD
<<<<<<< HEAD
// ListCheckRunsForRef implements RepoClient.ListCheckRunsForRef.
=======
>>>>>>> 376995a (docker file)
=======
// ListCheckRunsForRef implements RepoClient.ListCheckRunsForRef.
>>>>>>> 251f88d (comments)
func (client *localDirClient) ListCheckRunsForRef(ref string) ([]clients.CheckRun, error) {
	return nil, fmt.Errorf("ListCheckRunsForRef: %w", errUnsupportedFeature)
}

<<<<<<< HEAD
<<<<<<< HEAD
// ListStatuses implements RepoClient.ListStatuses.
=======
>>>>>>> 376995a (docker file)
=======
// ListStatuses implements RepoClient.ListStatuses.
>>>>>>> 251f88d (comments)
func (client *localDirClient) ListStatuses(ref string) ([]clients.Status, error) {
	return nil, fmt.Errorf("ListStatuses: %w", errUnsupportedFeature)
}

<<<<<<< HEAD
<<<<<<< HEAD
// Search implements RepoClient.Search.
=======
>>>>>>> 376995a (docker file)
=======
// Search implements RepoClient.Search.
>>>>>>> 251f88d (comments)
func (client *localDirClient) Search(request clients.SearchRequest) (clients.SearchResponse, error) {
	return clients.SearchResponse{}, fmt.Errorf("Search: %w", errUnsupportedFeature)
}

func (client *localDirClient) Close() error {
<<<<<<< HEAD
<<<<<<< HEAD
	// TODO
=======
>>>>>>> 376995a (docker file)
=======
	// TODO
>>>>>>> 251f88d (comments)
	return nil
}

// CreateLocalDirClient returns a client which implements RepoClient interface.
func CreateLocalDirClient(ctx context.Context, logger *zap.Logger) clients.RepoClient {
	return &localDirClient{
		ctx:    ctx,
		logger: logger,
	}
}
