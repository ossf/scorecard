// Copyright 2020 Security Scorecard Authors
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

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/mholt/archiver"
	"github.com/ossf/scorecard/gitcache/pkg"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func main() {
	var gitRepo *git.Repository

	logLevel := zap.LevelFlag("verbosity", zap.InfoLevel, "override the default log level")

	cfg := zap.NewProductionConfig()
	cfg.Level.SetLevel(*logLevel)
	logger, err := cfg.Build()
	if err != nil {
		log.Panic(err)
	}
	//nolint
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()
	blob := os.Getenv("BLOB_URL")
	sugar.Info("BLOB_URL ", blob)

	if blob == "" {
		sugar.Panic("BLOB_URL env is not set.")
	}

	// opening the blob
	bucket, e := pkg.NewBucket(blob)
	if e != nil {
		sugar.Panic(e)
	}

	sugar.Debugf("The bucket was opened %s", blob)

	repo := pkg.RepoURL{}
	err = repo.Set(os.Args[1])
	if err != nil {
		sugar.Panic(err)
	}

	sugar.Debugf("Fetching %+v", repo)

	// gets all the path configuration.
	storage, err := pkg.NewStoragePath(repo)
	if err != nil {
		sugar.Panic(err)
	}

	defer storage.Cleanup()

	alreadyUptoDate := false
	// checks if there is an existing git repo in the bucket
	if data, exists := bucket.Get(storage.BlobGitFolderPath); exists {
		sugar.Info("bucket ", blob, " already has git folder")
		gitRepo, alreadyUptoDate, err = fetchGitRepo(&storage, data)
	} else {
		sugar.Info("bucket ", blob, " does not have a git folder")
		gitRepo, err = cloneGitRepo(&storage, repo)
	}

	if err != nil {
		sugar.Panic(err)
	}

	// case where the git repository hasn't changed and it need not be updated into the blob
	if alreadyUptoDate {
		sugar.Info("bucketb ", blob, " git folder is already up to date.")
		// Just update the last sync time and return
		err = bucket.Set(storage.BlobLastSyncPath, []byte(fmt.Sprintf("%64b", time.Now().Unix())))
		if err != nil {
			sugar.Panic(err)
		}
		sugar.Info("Finished processing ", repo)
		return
	}

	lastRef, err := gitRepo.Head()
	if err != nil {
		sugar.Panic(err)
	}

	commit, err := gitRepo.CommitObject(lastRef.Hash())
	if err != nil {
		sugar.Panic(err)
	}

	data, err := archiveFolder(storage.GitDir, storage.GitTarFile)
	if err != nil {
		log.Panic(err)
	}

	err = bucket.Set(storage.BlobGitFolderPath, data)
	if err != nil {
		log.Panic(err)
	}

	// removing the .git folder as it is not required for the tar ball
	err = os.RemoveAll(path.Join(storage.GitDir, ".git"))
	if err != nil {
		log.Panic(err)
	}

	data, err = archiveFolder(storage.GitDir, storage.BlobArchiveFile)
	if err != nil {
		log.Panic(err)
	}

	err = bucket.Set(storage.BlobArchivePath, data)
	if err != nil {
		log.Panic(err)
	}

	sugar.Info("Storing the last commit ", commit.Author.When)
	// storing the last commit to the blob
	err = bucket.Set(storage.BlobLastCommitPath, []byte(fmt.Sprintf("%64b", commit.Author.When.Unix())))
	if err != nil {
		log.Panic(err)
	}
	err = bucket.Set(storage.BlobLastSyncPath, []byte(fmt.Sprintf("%64b", time.Now().Unix())))
	if err != nil {
		log.Panic(err)
	}
	sugar.Info("Finished processing for the first time ", repo)
}

// clones the repository if it does not exists. If the git repo already exists then it fetches from the remote.
func cloneGitRepo(storagePath *pkg.StoragePath, repo pkg.RepoURL) (*git.Repository, error) {
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
func fetchGitRepo(storagePath *pkg.StoragePath, data []byte) (*git.Repository, bool, error) {
	const fileMode os.FileMode = 0600
	if err := ioutil.WriteFile("gitfolder.tar.gz", data, fileMode); err != nil {
		return nil, false, errors.Wrapf(err, "unable write targz file %s", storagePath.BlobArchiveFile)
	}
	if err := archiver.Unarchive("gitfolder.tar.gz", storagePath.GitDir); err != nil {
		return nil, false,
			errors.Wrapf(err, "unable unarchive targz file %s in %s", storagePath.BlobArchiveFile, storagePath.BlobArchiveDir)
	}
	p := path.Join(storagePath.GitDir, storagePath.GitDir)
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
