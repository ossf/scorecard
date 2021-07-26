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

package clients

import (
	"fmt"
	"time"
)

// UPGRADEv2: use ErrRepoUnreachable instead.
type ErrRepoUnavailable struct {
	innerError error
}

func (e *ErrRepoUnavailable) Error() string {
	return fmt.Sprintf("repo cannot be accessed: %v", e.innerError)
}

func (e *ErrRepoUnavailable) Unwrap() error {
	return e.innerError
}

func NewRepoUnavailableError(err error) error {
	return &ErrRepoUnavailable{
		innerError: err,
	}
}

type RepoClient interface {
	InitRepo(owner, repo string) error
	ListFiles(predicate func(string) bool) []string
	GetFileContent(filename string) ([]byte, error)
	ListMergedPRs() ([]PullRequest, error)
	GetDefaultBranch() (BranchRef, error)
	Close() error
}

type BranchRef struct {
	Name                 string
	BranchProtectionRule BranchProtectionRule
}

type BranchProtectionRule struct {
	RequiredApprovingReviewCount int
}

// nolint: govet
type PullRequest struct {
	MergedAt    time.Time
	MergeCommit MergeCommit
	Number      int
	Labels      []Label
	Reviews     []Review
}

type MergeCommit struct {
	AuthoredByCommitter bool
}

type Label struct {
	Name string
}

type Review struct {
	State string
}
