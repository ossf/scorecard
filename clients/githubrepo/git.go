// Copyright 2024 OpenSSF Scorecard Authors
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

package githubrepo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/go-github/v53/github"

	"github.com/ossf/scorecard/v5/clients"
)

var errPathTraversal = errors.New("requested file outside repo directory")

type gitHandler struct {
	errSetup  error
	ctx       context.Context
	once      *sync.Once
	cloneURL  string
	gitRepo   *git.Repository
	tempDir   string
	commitSHA string
}

func (g *gitHandler) init(ctx context.Context, repo *github.Repository, commitSHA string) {
	g.errSetup = nil
	g.once = new(sync.Once)
	g.ctx = ctx
	g.cloneURL = repo.GetCloneURL()
	g.commitSHA = commitSHA
}

func (g *gitHandler) setup() error {
	g.once.Do(func() {
		tempDir, err := os.MkdirTemp("", repoDir)
		if err != nil {
			g.errSetup = err
			return
		}
		g.tempDir = tempDir
		g.gitRepo, err = git.PlainClone(g.tempDir, false, &git.CloneOptions{
			URL: g.cloneURL,
			// TODO: auth may be required for private repos
			Depth:        1, // currently only use the git repo for files, dont need history
			SingleBranch: true,
		})
		if err != nil {
			g.errSetup = err
			return
		}

		// assume the commit SHA is reachable from the default branch
		// this isn't as flexible as the tarball handler, but good enough for now
		if g.commitSHA != clients.HeadSHA {
			wt, err := g.gitRepo.Worktree()
			if err != nil {
				g.errSetup = err
				return
			}
			if err := wt.Checkout(&git.CheckoutOptions{Hash: plumbing.NewHash(g.commitSHA)}); err != nil {
				g.errSetup = fmt.Errorf("checkout specified commit: %w", err)
				return
			}
		}
	})
	return g.errSetup
}

func (g *gitHandler) getLocalPath() (string, error) {
	if err := g.setup(); err != nil {
		return "", fmt.Errorf("setup: %w", err)
	}
	return g.tempDir, nil
}

func (g *gitHandler) listFiles(predicate func(string) (bool, error)) ([]string, error) {
	if err := g.setup(); err != nil {
		return nil, fmt.Errorf("setup: %w", err)
	}
	ref, err := g.gitRepo.Head()
	if err != nil {
		return nil, fmt.Errorf("git.Head: %w", err)
	}

	commit, err := g.gitRepo.CommitObject(ref.Hash())
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

func (g *gitHandler) getFile(filename string) (*os.File, error) {
	if err := g.setup(); err != nil {
		return nil, fmt.Errorf("setup: %w", err)
	}

	// check for path traversal
	path := filepath.Join(g.tempDir, filename)
	if !strings.HasPrefix(path, filepath.Clean(g.tempDir)+string(os.PathSeparator)) {
		return nil, errPathTraversal
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	return f, nil
}

func (g *gitHandler) cleanup() error {
	if err := os.RemoveAll(g.tempDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("os.Remove: %w", err)
	}
	return nil
}
