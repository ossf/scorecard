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

// Package git defines helper functions for clients.RepoClient interface.
package git

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	cp "github.com/otiai10/copy"

	"github.com/ossf/scorecard/v4/clients"
)

const repoDir = "repo*"

var (
	errNilCommitFound = errors.New("nil commit found")
	errEmptyQuery     = errors.New("query is empty")
)

type Client struct {
	gitRepo        *git.Repository
	worktree       *git.Worktree
	listCommits    *sync.Once
	tempDir        string
	errListCommits error
	commits        []clients.Commit
	commitDepth    int
}

func (c *Client) InitRepo(uri, commitSHA string, commitDepth int) error {
	// cleanup previous state, if any.
	c.Close()
	c.listCommits = new(sync.Once)
	c.commits = nil

	// init
	c.commitDepth = commitDepth
	tempDir, err := os.MkdirTemp("", repoDir)
	if err != nil {
		return fmt.Errorf("os.MkdirTemp: %w", err)
	}

	// git clone
	const filePrefix = "file://"
	if strings.HasPrefix(uri, filePrefix) {
		if err := cp.Copy(strings.TrimPrefix(uri, filePrefix), tempDir); err != nil {
			return fmt.Errorf("cp.Copy: %w", err)
		}
		c.gitRepo, err = git.PlainOpen(tempDir)
		if err != nil {
			return fmt.Errorf("git.PlainOpen: %w", err)
		}
	} else {
		c.gitRepo, err = git.PlainClone(tempDir, false /*isBare*/, &git.CloneOptions{
			URL:      uri,
			Progress: os.Stdout,
		})
		if err != nil {
			return fmt.Errorf("git.PlainClone: %w", err)
		}
	}
	c.tempDir = tempDir
	c.worktree, err = c.gitRepo.Worktree()
	if err != nil {
		return fmt.Errorf("git.Worktree: %w", err)
	}

	// git checkout
	if commitSHA != clients.HeadSHA {
		if err := c.worktree.Checkout(&git.CheckoutOptions{
			Hash:  plumbing.NewHash(commitSHA),
			Force: true, // throw away any unsaved changes.
		}); err != nil {
			return fmt.Errorf("git.Worktree: %w", err)
		}
	}

	return nil
}

func (c *Client) ListCommits() (clients.CommitIterator, error) {
	c.listCommits.Do(func() {
		commitIter, err := c.gitRepo.Log(&git.LogOptions{
			Order: git.LogOrderCommitterTime,
		})
		if err != nil {
			c.errListCommits = fmt.Errorf("git.CommitObjects: %w", err)
			return
		}
		c.commits = make([]clients.Commit, 0, c.commitDepth)
		for i := 0; i < c.commitDepth; i++ {
			commit, err := commitIter.Next()
			if err != nil && !errors.Is(err, io.EOF) {
				c.errListCommits = fmt.Errorf("commitIter.Next: %w", err)
				return
			}
			// No more commits.
			if errors.Is(err, io.EOF) {
				break
			}

			if commit == nil {
				// Not sure in what case a nil commit is returned. Fail explicitly.
				c.errListCommits = fmt.Errorf("%w", errNilCommitFound)
				return
			}

			c.commits = append(c.commits, clients.Commit{
				SHA:           commit.Hash.String(),
				Message:       commit.Message,
				CommittedDate: commit.Committer.When,
				Committer: clients.User{
					Login: commit.Committer.Email,
				},
			})
		}
	})
	return clients.NewSliceBackedCommitIterator(c.commits), c.errListCommits
}

func (c *Client) Search(request clients.SearchRequest) (clients.SearchResponse, error) {
	// Pattern
	if request.Query == "" {
		return clients.SearchResponse{}, errEmptyQuery
	}
	queryRegexp, err := regexp.Compile(request.Query)
	if err != nil {
		return clients.SearchResponse{}, fmt.Errorf("regexp.Compile: %w", err)
	}
	grepOpts := &git.GrepOptions{
		Patterns: []*regexp.Regexp{queryRegexp},
	}

	// path/filename
	var pathExpr string
	switch {
	case request.Path != "" && request.Filename != "":
		pathExpr = filepath.Join(fmt.Sprintf("^%s", request.Path),
			fmt.Sprintf(".*%s$", request.Filename))
	case request.Path != "":
		pathExpr = fmt.Sprintf("^%s", request.Path)
	case request.Filename != "":
		pathExpr = filepath.Join(".*", fmt.Sprintf("%s$", request.Filename))
	}
	if pathExpr != "" {
		pathRegexp, err := regexp.Compile(pathExpr)
		if err != nil {
			return clients.SearchResponse{}, fmt.Errorf("regexp.Compile: %w", err)
		}
		grepOpts.PathSpecs = append(grepOpts.PathSpecs, pathRegexp)
	}

	// Grep
	grepResults, err := c.worktree.Grep(grepOpts)
	if err != nil {
		return clients.SearchResponse{}, fmt.Errorf("git.Grep: %w", err)
	}

	ret := clients.SearchResponse{}
	for _, grepResult := range grepResults {
		ret.Results = append(ret.Results, clients.SearchResult{
			Path: grepResult.FileName,
		})
	}
	ret.Hits = len(grepResults)
	return ret, nil
}

// TODO(#1709): Implement below fns using go-git.
func (c *Client) SearchCommits(request clients.SearchCommitsOptions) ([]clients.Commit, error) {
	return nil, nil
}

func (c *Client) ListFiles(predicate func(string) (bool, error)) ([]string, error) {
	return nil, nil
}

func (c *Client) GetFileContent(filename string) ([]byte, error) {
	return nil, nil
}

func (c *Client) Close() error {
	if err := os.RemoveAll(c.tempDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("os.RemoveAll: %w", err)
	}
	return nil
}
