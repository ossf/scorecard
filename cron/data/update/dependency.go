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

package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/google/go-github/v33/github"
	"golang.org/x/tools/go/vcs"

	"github.com/ossf/scorecard/cron/data"
	"github.com/ossf/scorecard/repos"
)

var (
	// TODO = move them outside the sourcecode.
	bazelRepos = []RepositoryDepsURL{
		{
			Owner: "envoyproxy",
			Repo:  "envoy",
			File:  "bazel/repository_locations.bzl",
		},
		{
			Owner: "envoyproxy",
			Repo:  "envoy",
			File:  "api/bazel/repository_locations.bzl",
		},
		{
			Owner: "grpc",
			Repo:  "grpc",
			File:  "bazel/grpc_deps.bzl",
		},
	}
	// TODO = move them outside the sourcecode.
	gorepos = []RepositoryDepsURL{
		{
			Owner: "ossf",
			Repo:  "scorecard",
		},
		{
			Owner: "sigstore",
			Repo:  "cosign",
		},
		{
			Owner:  "kubernetes",
			Repo:   "kubernetes",
			Vendor: true,
		},
	}
)

type RepositoryDepsURL struct {
	Owner, Repo, File string
	Vendor            bool
}

// Programmatically gets Envoy's dependencies and add to projects.
// Re-using a checker type.
func getBazelDeps(repo RepositoryDepsURL) []repos.RepoURL {
	client := github.NewClient(nil)
	ctx := context.Background()
	depRepos := []repos.RepoURL{}
	fo, _, _, err := client.Repositories.GetContents(ctx, repo.Owner, repo.Repo, repo.File, nil)
	if err != nil {
		// If we can't get content, gracefully fail but alert.
		log.Panicf("Failed to get repository content %s", err)
		return depRepos
	}

	fc, err := fo.GetContent()
	if err != nil {
		// If we can't get content, gracefully fail, but alert.
		log.Panicf("Failed to get repository content %s", err)
		return depRepos
	}

	// Match all patterns of github.com/{}/{}.
	re := regexp.MustCompile(`github.com/[^\/]*/[^\/"]*`)

	// TODO: Replace with a starlark interpreter that can be used for any project.
	for _, match := range re.FindAllString(fc, -1) {
		repo := repos.RepoURL{}
		if err := repo.Set(strings.TrimSuffix(match, ".git")); err != nil {
			log.Panicf("error during repo.Set: %v", err)
			return depRepos
		}
		depRepos = append(depRepos, repo)
	}
	return depRepos
}

// GetGoDeps returns go repo dependencies.
func getGoDeps(repo RepositoryDepsURL) []repos.RepoURL {
	repoURLs := []repos.RepoURL{}
	pwd, err := os.Getwd()
	if err != nil {
		log.Default().Println(err)
		return nil
	}
	//nolint
	defer os.Chdir(pwd)
	// creating temp dir for git clone
	gitDir, err := ioutil.TempDir(pwd, "")
	if err != nil {
		log.Default().Println("Cannot create temporary dir", err)
		return nil
	}
	defer os.RemoveAll(gitDir)

	// cloning git repo to get `go list -m all` out for getting all the dependencies
	_, err = git.PlainClone(gitDir, false,
		&git.CloneOptions{URL: fmt.Sprintf("http://github.com/%s/%s", repo.Owner, repo.Repo)})
	if err != nil {
		log.Default().Println(err)
		return nil
	}

	if err := os.Chdir(gitDir); err != nil {
		log.Default().Println(err)
		return nil
	}

	var cmd *exec.Cmd
	if repo.Vendor {
		cmd = exec.Command("go", "list", "-e", "mod=vendor", "all")
	} else {
		cmd = exec.Command("go", "list", "-m", "all")
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		log.Default().Println(err)
		return nil
	}

	/*
		   example output of go list -m all
			gopkg.in/resty.v1 v1.12.0
			gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7
	*/
	for _, l := range strings.Split(out.String(), "\n") {
		dependency := strings.Split(l, " ")[0]
		if strings.HasPrefix(dependency, "github.com") {
			repoURLs = parseGoModURL(dependency, repoURLs)
		} else {
			dependency = getVanityRepoURL(dependency)
			repoURLs = parseGoModURL(dependency, repoURLs)
		}
	}
	return repoURLs
}

// getVanityRepoURL returns actual git repository for the go vanity URL
// https://github.com/GoogleCloudPlatform/govanityurls.
func getVanityRepoURL(u string) string {
	repo, err := vcs.RepoRootForImportDynamic(u, false)
	if err != nil {
		log.Default().Println("unable to parse the vanity URL", u, err)
		return ""
	}
	return repo.Repo
}

func parseGoModURL(dependency string, repoURLs []repos.RepoURL) []repos.RepoURL {
	repoURL := repos.RepoURL{}
	splitURL := strings.Split(dependency, "/")
	// nolint:gomnd
	if len(splitURL) < 3 {
		return repoURLs
	}
	u := fmt.Sprintf("%s/%s/%s", splitURL[0], splitURL[1], splitURL[2])
	if err := repoURL.Set(u); err != nil {
		return repoURLs
	}
	repoURLs = append(repoURLs, repoURL)
	return repoURLs
}

func getDependencies(in io.Reader) (oldRepos, newRepos []repos.RepoURL, e error) {
	iter, err := data.MakeIteratorFrom(in)
	if err != nil {
		return nil, nil, fmt.Errorf("error during data.MakeIterator: %w", err)
	}

	// Read all project repositores into a map.
	m := make(map[string][]string)
	oldRepos = make([]repos.RepoURL, 0)
	for iter.HasNext() {
		repo, err := iter.Next()
		if err != nil {
			return nil, nil, fmt.Errorf("error during iter.Next: %w", err)
		}
		oldRepos = append(oldRepos, repo)
		// We do not handle duplicates.
		m[repo.URL()] = repo.Metadata
	}

	// Create a list of project dependencies that are not already present.
	newRepos = []repos.RepoURL{}
	for _, repo := range bazelRepos {
		for _, item := range getBazelDeps(repo) {
			if _, ok := m[item.URL()]; !ok {
				// Also add to m to avoid dupes.
				m[item.URL()] = item.Metadata
				newRepos = append(newRepos, item)
			}
		}
	}
	for _, repo := range gorepos {
		for _, item := range getGoDeps(repo) {
			if _, ok := m[item.URL()]; !ok {
				// Also add to m to avoid dupes.
				m[item.URL()] = item.Metadata
				newRepos = append(newRepos, item)
			}
		}
	}
	return oldRepos, newRepos, nil
}
