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

import (
	"fmt"
)

// ErrRepoUnavailable is returned when RepoClient is unable to reach the repo.
// UPGRADEv2: use ErrRepoUnreachable instead.
type ErrRepoUnavailable struct {
	innerError error
}

// Error returns the error string.
func (e *ErrRepoUnavailable) Error() string {
	return fmt.Sprintf("repo cannot be accessed: %v", e.innerError)
}

// Unwrap returns the wrapped error.
func (e *ErrRepoUnavailable) Unwrap() error {
	return e.innerError
}

// NewRepoUnavailableError returns a wrapped error of type ErrRepoUnavailable.
func NewRepoUnavailableError(err error) error {
	return &ErrRepoUnavailable{
		innerError: err,
	}
}

// RepoClient interface is used by Scorecard checks to access a repo.
type RepoClient interface {
	InitRepo(owner, repo string) error
	URL() string
	IsArchived() (bool, error)
	ListFiles(predicate func(string) (bool, error)) ([]string, error)
	GetFileContent(filename string) ([]byte, error)
	ListMergedPRs() ([]PullRequest, error)
	GetDefaultBranch() (BranchRef, error)
	ListCommits() ([]Commit, error)
	ListReleases() ([]Release, error)
	ListContributors() ([]Contributor, error)
	Search(request SearchRequest) (SearchResponse, error)
	Close() error
}
