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
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/google/go-github/v33/github"
	"github.com/jszwec/csvutil"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/vcs"
)

type RepositoryDepsURL struct {
	Owner, Repo, File string
	Vendor            bool
}

type Repository struct {
	Repo     string `csv:"repo"`
	Metadata string `csv:"metadata,omitempty"`
}

type gomod struct {
	Require []struct {
		Path string `json:"Path"`
	} `json:"Require"`
}

// Programmatically gets Envoy's dependencies and add to projects.
// Re-using a checker type.
func GetBazelDeps(repo RepositoryDepsURL) []Repository {
	client := github.NewClient(nil)
	ctx := context.Background()
	repos := []Repository{}
	fo, _, _, err := client.Repositories.GetContents(ctx, repo.Owner, repo.Repo, repo.File, nil)
	if err != nil {
		// If we can't get content, gracefully fail but alert.
		log.Panicf("Failed to get repository content %s", err)
		return repos
	}

	fc, err := fo.GetContent()
	if err != nil {
		// If we can't get content, gracefully fail, but alert.
		log.Panicf("Failed to get repository content %s", err)
		return repos
	}

	// Match all patterns of github.com/{}/{}.
	re := regexp.MustCompile(`github.com/[^\/]*/[^\/"]*`)

	// TODO: Replace with a starlark interpreter that can be used for any project.
	for _, match := range re.FindAllString(fc, -1) {
		repos = append(repos, Repository{strings.TrimSuffix(match, ".git"), ""})
	}

	return repos
}

// GetGoDeps returns go repo dependencies.
func GetGoDeps(repo RepositoryDepsURL) []Repository {
	repos := []Repository{}
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
			parseGoModURL(dependency, repos)
		} else {
			dependency = getVanityRepoURL(dependency)
			parseGoModURL(dependency, repos)
		}
	}
	return repos
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

// Runs scripts to update projects.txt with a projects dependencies.
// Adds "project=${PROJECT},dependency=true" to the repositories metadata.
// Args:
//     file path to projects.txt
func main() {
	// TODO = move them outside the sourcecode
	bazelRepos := []RepositoryDepsURL{
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
	// TODO = move them outside the sourcecode
	gorepos := []RepositoryDepsURL{
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

	projects, err := os.OpenFile(os.Args[1], os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	defer projects.Close()

	repos := []Repository{}
	data, err := ioutil.ReadAll(projects)
	if err != nil {
		panic(err)
	}
	err = csvutil.Unmarshal(data, &repos)
	if err != nil {
		panic(err)
	}

	// Read all projects.txt repositores into a map.
	m := make(map[string]string)
	for _, item := range repos {
		// We do not expect any duplicates.
		m[item.Repo] = item.Metadata
	}
	// Create a list of project dependencies that are not already present.
	newRepos := []Repository{}
	for _, repo := range bazelRepos {
		for _, item := range GetBazelDeps(repo) {
			if _, ok := m[item.Repo]; !ok {
				// Also add to m to avoid dupes.
				m[item.Repo] = item.Metadata
				newRepos = append(newRepos, item)
			}
		}
	}
	for _, repo := range gorepos {
		for _, item := range GetGoDeps(repo) {
			if _, ok := m[item.Repo]; !ok {
				// Also add to m to avoid dupes.
				m[item.Repo] = item.Metadata
				newRepos = append(newRepos, item)
			}
		}
	}

	// Append new repos to projects.txt without the header.
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	enc := csvutil.NewEncoder(w)
	enc.AutoHeader = false
	if err := enc.Encode(newRepos); err != nil {
		// This shouldn't happen.
		panic(err)
	}
	w.Flush()

	if _, err := projects.Write(buf.Bytes()); err != nil {
		panic(err)
	}
}

// RepoURL parses and stores URL into fields.
type RepoURL struct {
	Host  string // Host where the repo is stored. Example GitHub.com
	Owner string // Owner of the repo. Example ossf.
	Repo  string // The actual repo. Example scorecard.
}

func (r *RepoURL) String() string {
	return fmt.Sprintf("%s/%s/%s", r.Host, r.Owner, r.Repo)
}

func (r *RepoURL) NonURLString() string {
	return fmt.Sprintf("%s-%s-%s", r.Host, r.Owner, r.Repo)
}

func (r *RepoURL) Set(s string) error {
	// Allow skipping scheme for ease-of-use, default to https.
	if !strings.Contains(s, "://") {
		s = "https://" + s
	}

	parsedURL, err := url.Parse(s)
	if err != nil {
		return errors.Wrap(err, "unable to parse the URL")
	}

	const splitLen = 2
	split := strings.SplitN(strings.Trim(parsedURL.Path, "/"), "/", splitLen)
	if len(split) != splitLen {
		return errors.Errorf("invalid repo flag: [%s], pass the full repository URL", s)
	}

	r.Host, r.Owner, r.Repo = parsedURL.Host, split[0], split[1]
	return nil
}

func parseGoModURL(dependency string, repos []Repository) {
	repourl := RepoURL{}
	splitURL := strings.Split(dependency, "/")
	if len(splitURL) < 3 {
		return
	}
	u := fmt.Sprintf("%s/%s/%s", splitURL[0], splitURL[1], splitURL[2])
	var err error
	if err = repourl.Set(u); err == nil {
		repos = append(repos, Repository{repourl.String(), ""})
	}
}
