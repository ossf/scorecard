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

// Package clients defines the interface for RepoClient and related structs.
package clients

import "errors"

// ErrUnsupportedFeature indicates an API that is not supported by the client.
var ErrUnsupportedFeature = errors.New("unsupported feature")

// RepoClient interface is used by Scorecard checks to access a repo.
type RepoClient interface {
	InitRepo(repo Repo) error
	URI() string
	IsArchived() (bool, error)
	ListFiles(predicate func(string) (bool, error)) ([]string, error)
	GetFileContent(filename string) ([]byte, error)
	ListBranches() ([]*BranchRef, error)
	GetDefaultBranch() (*BranchRef, error)
	ListCommits() ([]Commit, error)
	ListIssues() ([]Issue, error)
	ListReleases() ([]Release, error)
	ListContributors() ([]Contributor, error)
	ListSuccessfulWorkflowRuns(filename string) ([]WorkflowRun, error)
	ListCheckRunsForRef(ref string) ([]CheckRun, error)
	ListStatuses(ref string) ([]Status, error)
	Search(request SearchRequest) (SearchResponse, error)
	Close() error
}
