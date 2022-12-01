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

// Package clients defines the interface for RepoClient and related structs.
package clients

import (
	"errors"
	"time"
)

// ErrUnsupportedFeature indicates an API that is not supported by the client.
var ErrUnsupportedFeature = errors.New("unsupported feature")

// HeadSHA is default commitSHA value used to denote git HEAD.
const HeadSHA = "HEAD"

// RepoClient interface is used by Scorecard checks to access a repo.
type RepoClient interface {
	InitRepo(repo Repo, commitSHA string, commitDepth int) error
	URI() string
	IsArchived() (bool, error)
	ListFiles(predicate func(string) (bool, error)) ([]string, error)
	GetFileContent(filename string) ([]byte, error)
	GetBranch(branch string) (*BranchRef, error)
	GetCreatedAt() (time.Time, error)
	GetDefaultBranchName() (string, error)
	GetDefaultBranch() (*BranchRef, error)
	ListCommits() ([]Commit, error)
	ListIssues() ([]Issue, error)
	ListLicenses() ([]License, error)
	ListReleases() ([]Release, error)
	ListContributors() ([]User, error)
	ListSuccessfulWorkflowRuns(filename string) ([]WorkflowRun, error)
	ListCheckRunsForRef(ref string) ([]CheckRun, error)
	ListStatuses(ref string) ([]Status, error)
	ListWebhooks() ([]Webhook, error)
	ListProgrammingLanguages() ([]Language, error)
	Search(request SearchRequest) (SearchResponse, error)
	SearchCommits(request SearchCommitsOptions) ([]Commit, error)
	Close() error
}
