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
package pkg

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/pkg/errors"
)

type StoragePath struct {
	BucketPath         string // name of the bucket that is going store the content
	GitDir             string // directory in which the git folder would be used for pull/clone
	GitTarDir          string // directory for tar/gzip the git folder
	GitTarFile         string // the tar file name for the git folder
	BlobArchiveDir     string // directory which is going archive the git folder without .git
	BlobArchiveFile    string // the tar file for the git folder without  .git
	BlobLastCommitPath string // blob path for storing the last commit
	BlobLastSyncPath   string // blob path for storing the lasy sync time
	BlobGitFolderPath  string // blob path for storing the GitTarFile
	BlobArchivePath    string // blob path for storing the archive file BlobArchiveFile
}

// NewStoragePath returns path for blob, archiving and also creates temp directories for archiving.
func NewStoragePath(repo RepoURL) (StoragePath, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return StoragePath{}, errors.Wrap(err, "unable to the current working dir")
	}

	bucketPath := fmt.Sprintf("gitcache/%s/%s/%s", repo.Host, repo.Owner, repo.Repo)
	gitDir := repo.Host + repo.Owner + repo.Repo

	err = os.Mkdir(gitDir, 0755)
	if err != nil {
		return StoragePath{}, errors.Wrapf(err, "unable to create temp directory %s", gitDir)
	}
	gitTarPath := path.Join(gitDir, "gitfolder.tar.gz")

	blobArchiveDir, err := ioutil.TempDir(cwd, gitDir+"tar")
	if err != nil {
		return StoragePath{}, errors.Wrapf(err, "unable to create temp directory %s", gitDir+"tar")
	}
	blobArchivePath := path.Join(blobArchiveDir, fmt.Sprintf("%s.tar.gz", repo.Repo))

	return StoragePath{
		BucketPath:         bucketPath,
		GitDir:             gitDir,
		GitTarFile:         gitTarPath,
		BlobArchiveDir:     blobArchiveDir,
		BlobArchiveFile:    blobArchivePath,
		BlobLastCommitPath: fmt.Sprintf("%s/lastcommit", bucketPath),
		BlobLastSyncPath:   fmt.Sprintf("%s/lastsync", bucketPath),
		BlobGitFolderPath:  fmt.Sprintf("%s/gitfolder", bucketPath),
		BlobArchivePath:    fmt.Sprintf("%s/tar", bucketPath),
	}, nil
}

// Cleanup removes the directories that were created.
func (s *StoragePath) Cleanup() {
	os.RemoveAll(s.GitDir)
	os.RemoveAll(s.GitTarDir)
	os.RemoveAll(s.GitTarFile)
	os.RemoveAll(s.BlobArchiveDir)
}
