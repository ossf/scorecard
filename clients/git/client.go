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
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	cp "github.com/otiai10/copy"

	"github.com/ossf/scorecard/v4/clients"
)

const repoDir = "repo*"

var (
	errNilCommitFound = errors.New("nil commit found")
	errEmptyQuery     = errors.New("query is empty")
	errDefaultBranch  = errors.New("default branch name could not be determined")

	// ensure Client implements clients.RepoClient.
	_ clients.RepoClient = (*Client)(nil)
)

type Client struct {
	repo           clients.Repo
	errListCommits error
	gitRepo        *git.Repository
	worktree       *git.Worktree
	listCommits    *sync.Once
	tempDir        string
	commits        []clients.Commit
	commitDepth    int
}

func (c *Client) InitRepo(repo clients.Repo, commitSHA string, commitDepth int) error {
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
	uri := repo.URI()
	c.tempDir = tempDir
	const filePrefix = "file://"
	if strings.HasPrefix(uri, filePrefix) { //nolint:nestif
		if err := cp.Copy(strings.TrimPrefix(uri, filePrefix), tempDir); err != nil {
			return fmt.Errorf("cp.Copy: %w", err)
		}
		c.gitRepo, err = git.PlainOpen(tempDir)
		if err != nil {
			return fmt.Errorf("git.PlainOpen: %w", err)
		}
	} else {
		if !strings.HasPrefix(uri, "https://") && !strings.HasPrefix(uri, "ssh://") {
			uri = "https://" + uri
		}
		if !strings.HasSuffix(uri, ".git") {
			uri += ".git"
		}
		c.gitRepo, err = git.PlainClone(tempDir, false /*isBare*/, &git.CloneOptions{
			URL:      uri,
			Progress: os.Stdout,
		})
	}
	if err != nil {
		return fmt.Errorf("git.PlainClone: %w %s", err, uri)
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

func (c *Client) ListCommits() ([]clients.Commit, error) {
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
	return c.commits, c.errListCommits
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

func (c *Client) ListFiles(predicate func(string) (bool, error)) ([]string, error) {
	ref, err := c.gitRepo.Head()
	if err != nil {
		return nil, fmt.Errorf("git.Head: %w", err)
	}

	commit, err := c.gitRepo.CommitObject(ref.Hash())
	if err != nil {
		return nil, fmt.Errorf("git.CommitObject: %w", err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("git.Commit.Tree: %w", err)
	}

	var files []string
	err = tree.Files().ForEach(func(f *object.File) error {
		shouldInclude, err := predicate(f.Name)
		if err != nil {
			return fmt.Errorf("error applying predicate to file %s: %w", f.Name, err)
		}

		if shouldInclude {
			files = append(files, f.Name)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("git.Tree.Files: %w", err)
	}

	return files, nil
}

func (c *Client) GetFileReader(filename string) (io.ReadCloser, error) {
	// Create the full path of the file
	fullPath := filepath.Join(c.tempDir, filename)

	// Read the file
	f, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("os.Open: %w", err)
	}

	return f, nil
}

func (c *Client) IsArchived() (bool, error) {
	return false, clients.ErrUnsupportedFeature
}

func (c *Client) URI() string {
	return c.repo.URI()
}

func (c *Client) Close() error {
	if err := os.RemoveAll(c.tempDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("os.RemoveAll: %w", err)
	}
	return nil
}

func (c *Client) GetBranch(branch string) (*clients.BranchRef, error) {
	// Get the branch reference
	ref, err := c.gitRepo.Branch(branch)
	if err != nil {
		return nil, fmt.Errorf("git.Branch: %w", err)
	}

	// Get the commit object
	if err != nil {
		return nil, fmt.Errorf("git.CommitObject: %w", err)
	}
	f := false
	// Create the BranchRef object
	branchRef := &clients.BranchRef{
		Name:      &ref.Name,
		Protected: &f,
	}

	return branchRef, nil
}

func (c *Client) GetCreatedAt() (time.Time, error) {
	// Retrieve the first commit of the repository
	commitIter, err := c.gitRepo.Log(&git.LogOptions{Order: git.LogOrderCommitterTime})
	if err != nil {
		return time.Time{}, fmt.Errorf("git.Log: %w", err)
	}
	defer commitIter.Close()

	// Iterate through the commits to find the first one
	var firstCommit *object.Commit
	err = commitIter.ForEach(func(c *object.Commit) error {
		firstCommit = c
		return storer.ErrStop
	})
	if err != nil && !errors.Is(err, storer.ErrStop) {
		return time.Time{}, fmt.Errorf("commitIter.ForEach: %w", err)
	}

	if firstCommit == nil {
		return time.Time{}, errNilCommitFound
	}

	// Return the commit time of the first commit
	return firstCommit.Committer.When, nil
}

func (c *Client) GetDefaultBranchName() (string, error) {
	headRef, err := c.gitRepo.Head()
	if err != nil {
		return "", fmt.Errorf("git.Head: %w", err)
	}

	// Extract the branch name from the Head reference
	defaultBranch := headRef.Name()
	if defaultBranch == "" {
		return "", errDefaultBranch
	}

	return string(defaultBranch), nil
}

func (c *Client) GetDefaultBranch() (*clients.BranchRef, error) {
	// TODO: Implement this
	return nil, clients.ErrUnsupportedFeature
}

func (c *Client) GetOrgRepoClient(ctx context.Context) (clients.RepoClient, error) {
	return nil, clients.ErrUnsupportedFeature
}

func (c *Client) ListIssues() ([]clients.Issue, error) {
	return nil, clients.ErrUnsupportedFeature
}

func (c *Client) ListLicenses() ([]clients.License, error) {
	return nil, clients.ErrUnsupportedFeature
}

func (c *Client) ListReleases() ([]clients.Release, error) {
	return nil, clients.ErrUnsupportedFeature
}

func (c *Client) ListContributors() ([]clients.User, error) {
	// TODO: Implement this
	return nil, clients.ErrUnsupportedFeature
}

func (c *Client) ListSuccessfulWorkflowRuns(filename string) ([]clients.WorkflowRun, error) {
	return nil, clients.ErrUnsupportedFeature
}

func (c *Client) ListCheckRunsForRef(ref string) ([]clients.CheckRun, error) {
	return nil, clients.ErrUnsupportedFeature
}

func (c *Client) ListStatuses(ref string) ([]clients.Status, error) {
	return nil, clients.ErrUnsupportedFeature
}

func (c *Client) ListWebhooks() ([]clients.Webhook, error) {
	return nil, clients.ErrUnsupportedFeature
}

func (c *Client) ListProgrammingLanguages() ([]clients.Language, error) {
	return nil, clients.ErrUnsupportedFeature
}

func (c *Client) SearchCommits(request clients.SearchCommitsOptions) ([]clients.Commit, error) {
	return nil, clients.ErrUnsupportedFeature
}

func (c *Client) LocalPath() (string, error) {
	return c.tempDir, nil
}
