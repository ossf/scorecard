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
	"io/ioutil"
	"log"
	"os"
	"regexp"

	"github.com/google/go-github/v33/github"
	"github.com/jszwec/csvutil"
)

type RepositoryDepsURL struct {
	Owner, Repo, File string
}

type Repository struct {
	Repo     string `csv:"repo"`
	Metadata string `csv:"metadata,omitempty"`
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
