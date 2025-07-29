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

package minder

import (
	"fmt"

	"github.com/go-git/go-billy/v5/memfs"
	billyutil "github.com/go-git/go-billy/v5/util"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"

	"github.com/ossf/scorecard/v5/checks/fileparser"
	"github.com/ossf/scorecard/v5/clients"
)

func repositoryFromClient(repo clients.RepoClient) (*git.Repository, error) {
	// We only need to return a working filesystem and a plumbing hash for the evaluator.
	// The storage is used by remediation actions, but not during evaluation.
	// The hash is used in checkpoints to identify the checkpoint, but not otherwise used.
	storer := memory.NewStorage()
	fs := memfs.New()
	gitRepo, err := git.Init(storer, fs)
	if err != nil {
		return nil, fmt.Errorf("error initializing git repository: %w", err)
	}
	worktree, err := gitRepo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("error getting git worktree: %w", err)
	}
	workfs := worktree.Filesystem
	// Right now, because we're exposing git.Repository, not billy.Filesystem, we need to
	// copy all the files into the git repository.
	err = fileparser.OnMatchingFileContentDo(
		repo,
		fileparser.PathMatcher{Pattern: "*"},
		func(path string, content []byte, args ...any) (bool, error) {
			err := billyutil.WriteFile(workfs, path, content, 0o644)
			if err != nil {
				return false, fmt.Errorf("error writing file to git worktree: %w", err)
			}
			return true, nil
		})
	if err != nil {
		return nil, fmt.Errorf("error processing file content: %w", err)
	}

	// Minder wants a head reference to exist, so create one
	if _, err := worktree.Add("."); err != nil {
		return nil, fmt.Errorf("error adding files to git worktree: %w", err)
	}
	if _, err := worktree.Commit("Initial commit", &git.CommitOptions{}); err != nil {
		return nil, fmt.Errorf("error committing changes to git worktree: %w", err)
	}

	return gitRepo, nil
}
