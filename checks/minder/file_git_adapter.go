package minder

import (
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
		return nil, err
	}
	worktree, err := gitRepo.Worktree()
	if err != nil {
		return nil, err
	}
	workfs := worktree.Filesystem
	// Right now, because we're exposing git.Repository, not billy.Filesystem, we need to
	// copy all the files into the git repository.
	err = fileparser.OnMatchingFileContentDo(repo, fileparser.PathMatcher{
		Pattern: "*",
	}, func(path string, content []byte, args ...interface{}) (bool, error) {
		err := billyutil.WriteFile(workfs, path, content, 0644)
		if err != nil {
			return false, err
		}
		return true, nil
	})

	if err != nil {
		return nil, err
	}

	// Minder wants a head reference to exist, so create one
	if _, err := worktree.Add("."); err != nil {
		return nil, err
	}
	if _, err := worktree.Commit("Initial commit", &git.CommitOptions{}); err != nil {
		return nil, err
	}

	return gitRepo, nil
}
