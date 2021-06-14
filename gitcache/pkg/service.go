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

package pkg

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/mholt/archiver"
	"github.com/pkg/errors"
)

type cacheService struct {
	BlobURL string
	TempDir string
	Logf    func(s string, f ...interface{})
}

// CacheService updates git cache.
type CacheService interface {
	// UpdateCache updates the git cache into the gocloud.dev blob storage.
	UpdateCache(string) error
}

// NewCacheService returns new CacheService.
func NewCacheService(blobURL, tempDir string, logf func(s string, f ...interface{})) (CacheService, error) {
	if blobURL == "" {
		return nil, errors.New("BLOB_URL env cannot be empty")
	}
	if tempDir == "" {
		return nil, errors.New("TEMP_DIR  env cannot be empty")
	}
	if logf == nil {
		return nil, errors.New("Log function cannot be nil")
	}
	return cacheService{
		BlobURL: blobURL,
		Logf:    logf,
		TempDir: tempDir,
	}, nil
}

// UpdateCache updates the git cache into the gocloud.dev blob storage.
func (c cacheService) UpdateCache(s string) error {
	var gitRepo *git.Repository
	// opening the blob
	bucket, err := NewBucket(c.BlobURL)
	if err != nil {
		return errors.Wrapf(err, "unable to open the bucket %s", c.BlobURL)
	}

	repo := RepoURL{}
	if err := repo.Set(s); err != nil {
		return errors.Wrapf(err, "unable parse the URL %s", s)
	}

	// gets all the path configuration.
	storage, err := NewStoragePath(repo, c.TempDir)
	if err != nil {
		return errors.Wrapf(err, "unable get storage")
	}

	defer storage.Cleanup()

	alreadyUptoDate := false
	// checks if there is an existing git repo in the bucket
	if data, exists := bucket.Get(storage.BlobGitFolderPath); exists {
		c.Logf("bucket ", c.BlobURL, " already has git folder")
		gitRepo, alreadyUptoDate, err = fetchGitRepo(&storage, data, repo, c.TempDir)
	} else {
		c.Logf("bucket ", c.BlobURL, " does not have a git folder")
		gitRepo, err = cloneGitRepo(&storage, repo)
	}

	if err != nil {
		return errors.Wrapf(err, "unable open the git repo")
	}

	// case where the git repository hasn't changed and it need not be updated into the blob
	if alreadyUptoDate {
		c.Logf("bucketb ", c.BlobURL, " git folder is already up to date.")
		// Just update the last sync time and return
		err = bucket.Set(storage.BlobLastSyncPath, []byte(fmt.Sprintf("%64b", time.Now().Unix())))
		if err != nil {
			return errors.Wrapf(err, "unable set storage last sync path")
		}
		c.Logf("Finished processing ", repo)
		return nil
	}

	lastRef, err := gitRepo.Head()
	if err != nil {
		return errors.Wrapf(err, "unable get last ref")
	}

	commit, err := gitRepo.CommitObject(lastRef.Hash())
	if err != nil {
		return errors.Wrapf(err, "unable get commit object")
	}

	data, err := archiveFolder(storage.GitDir, storage.GitTarFile)
	if err != nil {
		return errors.Wrapf(err, "unable archive folder %s %s", storage.GitDir, storage.GitTarFile)
	}

	err = bucket.Set(storage.BlobGitFolderPath, data)
	if err != nil {
		return errors.Wrapf(err, "unable set blob contnet for %s ", storage.BlobGitFolderPath)
	}

	// removing the .git folder as it is not required for the tar ball
	err = os.RemoveAll(path.Join(storage.GitDir, ".git"))
	if err != nil {
		return errors.Wrapf(err, "unable .git folder %s ", storage.GitDir)
	}

	data, err = archiveFolder(storage.GitDir, storage.BlobArchiveFile)
	if err != nil {
		return errors.Wrapf(err, "unable archive folder %s %s", storage.GitDir, storage.BlobArchiveFile)
	}

	err = bucket.Set(storage.BlobArchivePath, data)
	if err != nil {
		return errors.Wrapf(err, "unable set blob contnet for %s ", storage.BlobArchivePath)
	}

	c.Logf("Storing the last commit ", commit.Author.When)
	// storing the last commit to the blob
	err = bucket.Set(storage.BlobLastCommitPath, []byte(fmt.Sprintf("%64b", commit.Author.When.Unix())))
	if err != nil {
		return errors.Wrapf(err, "unable set storage last commit")
	}
	err = bucket.Set(storage.BlobLastSyncPath, []byte(fmt.Sprintf("%64b", time.Now().Unix())))
	if err != nil {
		return errors.Wrapf(err, "unable set storage last sync")
	}
	c.Logf("Finished processing for the first time ", repo)
	return nil
}

// clones the repository if it does not exists. If the git repo already exists then it fetches from the remote.
func cloneGitRepo(storagePath *StoragePath, repo RepoURL) (*git.Repository, error) {
	// Clone the given repository to the given directory
	gitRepo, err := git.PlainClone(storagePath.GitDir, false, &git.CloneOptions{
		URL: fmt.Sprintf("http://%s/%s/%s", repo.Host, repo.Owner, repo.Repo),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to clone %+v", repo)
	}
	return gitRepo, nil
}

func archiveFolder(folderToArchive, archivePath string) ([]byte, error) {
	tarformat := archiver.DefaultTarGz
	err := tarformat.Archive([]string{folderToArchive}, archivePath)
	if err != nil {
		return nil,
			errors.Wrapf(err, "unable to read the archive the folder %s to the archive path %s", folderToArchive, archivePath)
	}

	data, err := ioutil.ReadFile(archivePath)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read the archive file %s", archivePath)
	}
	return data, nil
}

// fetchGitRepo fetches the git repo. Returns git repository, bool if it is already up to date and error.
func fetchGitRepo(storagePath *StoragePath, data []byte, repo RepoURL, tempDir string) (*git.Repository, bool, error) {
	const fileMode os.FileMode = 0o600
	gitZipFile := path.Join(tempDir, "gitfolder.tar.gz")
	if err := ioutil.WriteFile(gitZipFile, data, fileMode); err != nil {
		return nil, false, errors.Wrapf(err, "unable write targz file %s", storagePath.BlobArchiveFile)
	}
	if err := archiver.Unarchive(gitZipFile, storagePath.GitDir); err != nil {
		return nil, false,
			errors.Wrapf(err, "unable unarchive targz file %s in %s", storagePath.BlobArchiveFile, storagePath.BlobArchiveDir)
	}
	p := path.Join(storagePath.GitDir, repo.NonURLString())
	gitRepo, err := git.PlainOpen(p)
	if err != nil {
		return nil, false, errors.Wrapf(err, "unable to open the git dir %s", p)
	}
	err = gitRepo.Fetch(&git.FetchOptions{RemoteName: "origin"})
	// the repo is up to date and can return
	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		return gitRepo, true, nil
	}

	if err != nil {
		return nil, false, errors.Wrap(err, "unable to fetch from git repo")
	}
	return gitRepo, false, nil
}
