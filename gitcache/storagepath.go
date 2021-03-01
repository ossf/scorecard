package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/pkg/errors"
)

type StoragePath struct {
	BucketName         string // name of the bucket that is going store the content
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

// Returns path for blob, archiving and also creates temp directories for archiving.
func getStoragePath(repo RepoURL) (StoragePath, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return StoragePath{}, errors.Wrap(err, "unable to the current working dir")
	}

	bucketName := fmt.Sprintf("gitcache/%s/%s/%s", repo.Host, repo.Owner, repo.Repo)
	gitDir := fmt.Sprintf(repo.Host + repo.Owner + repo.Repo)

	err = os.Mkdir(gitDir, 0755)
	if err != nil {
		return StoragePath{}, errors.Wrapf(err, "unable to create temp directory %s", gitDir)
	}
	gitTarPath := path.Join(gitDir, fmt.Sprintf("%s.tar.gz", "gitfolder"))

	blobArchiveDir, err := ioutil.TempDir(cwd, gitDir+"tar")
	if err != nil {
		return StoragePath{}, errors.Wrapf(err, "unable to create temp directory %s", gitDir+"tar")
	}
	blobArchivePath := path.Join(blobArchiveDir, fmt.Sprintf("%s.tar.gz", repo.Repo))

	return StoragePath{
		BucketName:         bucketName,
		GitDir:             gitDir,
		GitTarFile:         gitTarPath,
		BlobArchiveDir:     blobArchiveDir,
		BlobArchiveFile:    blobArchivePath,
		BlobLastCommitPath: fmt.Sprintf("%s/lastcommit", bucketName),
		BlobLastSyncPath:   fmt.Sprintf("%s/lastsync", bucketName),
		BlobGitFolderPath:  fmt.Sprintf("%s/gitfolder", bucketName),
		BlobArchivePath:    fmt.Sprintf("%s/tar", bucketName),
	}, nil
}
