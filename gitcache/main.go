package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/mholt/archiver"
)

func main() {
	blob := os.Getenv("BLOB_URL")

	if blob == "" {
		log.Panic("BLOB_URL env is not set.")
	}

	repo := &RepoURL{}
	err := repo.Set(os.Args[1])
	if err != nil {
		log.Panic(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Panic(err)
	}

	dir, err := ioutil.TempDir(path, repo.Owner+repo.Owner)
	if err != nil {
		log.Panic(err)
	}

	tarDir, err := ioutil.TempDir(cwd, repo.Owner+repo.Owner+"tar")
	if err != nil {
		log.Panic(err)
	}

	defer os.RemoveAll(dir)
	defer os.RemoveAll(tarDir)

	// Clone the given repository to the given directory
	r, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL:   fmt.Sprintf("http://%s/%s/%s", repo.Host, repo.Owner, repo.Repo),
		Depth: 1, // Just fetch the last commit
	})
	if err != nil {
		log.Panic(err)
	}
	lastRef, err := r.Head()
	if err != nil {
		log.Panic(err)
	}

	commit, err := r.CommitObject(lastRef.Hash())
	if err != nil {
		log.Panic(err)
	}

	// opening the blob
	bucket, err := New(blob)
	if err != nil {
		log.Panic(err)
	}

	archivePath := path.Join(tarDir, fmt.Sprintf("%s.tar.gz", repo.Repo))
	// storing the last commit to the blob
	err = bucket.Set(fmt.Sprintf("gitcache/%s/%s/lastcommit", repo.Owner, repo.Repo), []byte(fmt.Sprintf("%64b\n", commit.Author.When.Unix())))
	if err != nil {
		log.Panic(err)
	}

	now := time.Now().Unix()
	err = bucket.Set(fmt.Sprintf("gitcache/%s/%s/lastsync", repo.Owner, repo.Repo), []byte(fmt.Sprintf("%64b\n", now)))
	if err != nil {
		log.Panic(err)
	}

	// removing the .git folder as it is not required for the tar ball
	err = os.RemoveAll(path.Join(dir, ".git"))
	if err != nil {
		log.Panic(err)
	}

	// creating an archive
	archive := archiver.NewTarGz()

	err = archive.Archive([]string{dir}, archivePath)
	if err != nil {
		log.Panic(err)
	}

	data, err := ioutil.ReadFile(archivePath)
	if err != nil {
		log.Panic(err)
	}

	// storing the archive to the blob
	err = bucket.Set(fmt.Sprintf("gitcache/%s/%s/tar", repo.Owner, repo.Repo), data)
	if err != nil {
		log.Panic(err)
	}
}

type RepoURL struct {
	Host, Owner, Repo string
}

func (r *RepoURL) String() string {
	return fmt.Sprintf("%s/%s/%s", r.Host, r.Owner, r.Repo)
}

func (r *RepoURL) Type() string {
	return "repo"
}

func (r *RepoURL) Set(s string) error {
	// Allow skipping scheme for ease-of-use, default to https.
	if !strings.Contains(s, "://") {
		s = "https://" + s
	}

	parsedURL, err := url.Parse(s)
	if err != nil {
		return err
	}

	const splitLen = 2
	split := strings.SplitN(strings.Trim(parsedURL.Path, "/"), "/", splitLen)
	if len(split) != splitLen {
		log.Fatalf("invalid repo flag: [%s], pass the full repository URL", s)
	}

	r.Host, r.Owner, r.Repo = parsedURL.Host, split[0], split[1]
	return nil
}
