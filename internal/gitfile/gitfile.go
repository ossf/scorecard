// Copyright 2025 OpenSSF Scorecard Authors
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

// Package gitfile defines functionality to list and fetch files after temporarily cloning a git repo.
package gitfile

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

	"github.com/ossf/scorecard/v5/clients"
)

var errPathTraversal = errors.New("requested file outside repo")

const repoDir = "repo*"

type Handler struct {
	errSetup  error
	ctx       context.Context
	once      *sync.Once
	cloneURL  string
	gitRepo   *git.Repository
	tempDir   string
	commitSHA string
	files     []string
}

func (h *Handler) Init(ctx context.Context, cloneURL, commitSHA string) {
	h.errSetup = nil
	h.once = new(sync.Once)
	h.ctx = ctx
	h.cloneURL = cloneURL
	h.commitSHA = commitSHA
	h.files = nil
}

func (h *Handler) setup() error {
	h.once.Do(func() {
		tempDir, err := os.MkdirTemp("", repoDir)
		if err != nil {
			h.errSetup = err
			return
		}
		h.tempDir = tempDir
		h.gitRepo, err = git.PlainClone(h.tempDir, false, &git.CloneOptions{
			URL: h.cloneURL,
			// TODO: auth may be required for private repos
			Depth:        1, // currently only use the git repo for files, dont need history
			SingleBranch: true,
		})
		if err != nil {
			h.errSetup = err
			return
		}

		// assume the commit SHA is reachable from the default branch
		// this isn't as flexible as the tarball handler, but good enough for now
		if h.commitSHA != clients.HeadSHA {
			wt, err := h.gitRepo.Worktree()
			if err != nil {
				h.errSetup = err
				return
			}
			if err := wt.Checkout(&git.CheckoutOptions{Hash: plumbing.NewHash(h.commitSHA)}); err != nil {
				h.errSetup = fmt.Errorf("checkout specified commit: %w", err)
				return
			}
		}

		// go-git is not thread-safe so list the files inside this sync.Once and save them
		// https://github.com/go-git/go-git/issues/773
		files, err := enumerateFiles(h.gitRepo)
		if err != nil {
			h.errSetup = err
			return
		}
		h.files = files
	})
	return h.errSetup
}

func (h *Handler) GetLocalPath() (string, error) {
	if err := h.setup(); err != nil {
		return "", fmt.Errorf("setup: %w", err)
	}
	return h.tempDir, nil
}

func (h *Handler) ListFiles(predicate func(string) (bool, error)) ([]string, error) {
	if err := h.setup(); err != nil {
		return nil, fmt.Errorf("setup: %w", err)
	}
	var files []string
	for _, f := range h.files {
		shouldInclude, err := predicate(f)
		if err != nil {
			return nil, fmt.Errorf("error applying predicate to file %s: %w", f, err)
		}

		if shouldInclude {
			files = append(files, f)
		}
	}
	return files, nil
}

func enumerateFiles(repo *git.Repository) ([]string, error) {
	ref, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("git.Head: %w", err)
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, fmt.Errorf("git.CommitObject: %w", err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("git.Commit.Tree: %w", err)
	}

	var files []string
	err = tree.Files().ForEach(func(f *object.File) error {
		files = append(files, f.Name)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("git.Tree.Files: %w", err)
	}

	return files, nil
}

func (h *Handler) GetFile(filename string) (*os.File, error) {
	if err := h.setup(); err != nil {
		return nil, fmt.Errorf("setup: %w", err)
	}

	// check for path traversal
	path := filepath.Join(h.tempDir, filename)
	if !strings.HasPrefix(path, filepath.Clean(h.tempDir)+string(os.PathSeparator)) {
		return nil, errPathTraversal
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	return f, nil
}

func (h *Handler) Cleanup() error {
	if err := os.RemoveAll(h.tempDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("os.Remove: %w", err)
	}
	return nil
}
