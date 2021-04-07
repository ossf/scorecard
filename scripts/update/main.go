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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/google/go-github/v33/github"
	"github.com/jszwec/csvutil"
	"github.com/pkg/errors"
)

type RepositoryDepsURL struct {
	Owner, Repo, File string
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
		repos = append(repos, Repository{match, ""})
	}

	return repos
}

// GetGoDeps returns go repo dependencies.
func GetGoDeps(repo RepositoryDepsURL) []Repository {
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

	tmpFile, err := ioutil.TempFile(os.TempDir(), "")
	if err != nil {
		log.Fatal("Cannot create temporary file", err)
	}

	// Remember to clean up the file afterwards
	defer os.Remove(tmpFile.Name())

	// Example writing to the file
	if _, err = tmpFile.Write([]byte(fc)); err != nil {
		log.Panicf("Failed to write to temporary file %s", err)
	}

	//nolint
	cmd := exec.Command("go", "mod", "edit", "--json", tmpFile.Name())

	var out bytes.Buffer
	cmd.Stdout = &out

	err = cmd.Run()
	if err != nil {
		log.Panic(err)
	}

	mods := &gomod{}
	if err := json.Unmarshal(out.Bytes(), mods); err != nil {
		log.Panic(err)
	}
	// Close the file
	if err := tmpFile.Close(); err != nil {
		log.Panic(err)
	}

	for _, match := range mods.Require {
		if strings.HasPrefix(match.Path, "github.com") {
			repourl := RepoURL{}
			if err := repourl.Set(match.Path); err == nil {
				repos = append(repos, Repository{repourl.String(), ""})
			}
		}
	}
	return repos
}

// Runs scripts to update projects.txt with a projects dependencies.
// Adds "project=${PROJECT},dependency=true" to the repositories metadata.
// Args:
//     file path to projects.txt
func main() {
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
	gorepos := []RepositoryDepsURL{
		{
			Owner: "kubernetes",
			Repo:  "kubernetes",
			File:  "go.mod",
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
