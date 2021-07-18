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
	"fmt"
	"os"

	"github.com/ossf/scorecard/cron/data"
	"github.com/ossf/scorecard/repos"
)

// Script to add new project repositories to the projects.csv file:
// * Removes any duplicates
// * Repos are read in order and the first entry is honored in case of duplicates
// * Sort and output all projects
// Usage: add all new dependencies to the projects.csv file before running this script
// Args:
//     path to input.csv output.csv
func main() {
	// nolint: gomnd
	if len(os.Args) != 3 {
		panic("must provide 2 arguments")
	}
	// nolint: gomnd
	inFile, err := os.OpenFile(os.Args[1], os.O_RDONLY, 0o644)
	if err != nil {
		panic(err)
	}
	iter, err := data.MakeIteratorFrom(inFile)
	if err != nil {
		panic(err)
	}

	repoURLs, err := getRepoURLs(iter)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	if err := data.SortAndAppendTo(&buf, repoURLs, nil); err != nil {
		panic(err)
	}
	// nolint: gomnd
	projects, err := os.OpenFile(os.Args[2], os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		panic(err)
	}
	if _, err := projects.Write(buf.Bytes()); err != nil {
		panic(err)
	}
}

func getRepoURLs(iter data.Iterator) ([]repos.RepoURL, error) {
	repoURLs := make(map[string]*repos.RepoURL)
	repoMap := make(map[string]map[string]bool)
	for iter.HasNext() {
		repo, err := iter.Next()
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
		if _, ok := repoMap[repo.URL()]; !ok {
			repoURLs[repo.URL()] = new(repos.RepoURL)
			*repoURLs[repo.URL()] = repo
			repoMap[repo.URL()] = make(map[string]bool)
			for _, metadata := range repo.Metadata {
				repoMap[repo.URL()][metadata] = true
			}
			continue
		}
		for _, metadata := range repo.Metadata {
			if _, ok := repoMap[repo.URL()][metadata]; !ok && metadata != "" {
				repoURLs[repo.URL()].Metadata = append(repoURLs[repo.URL()].Metadata, metadata)
				repoMap[repo.URL()][metadata] = true
			}
		}
	}

	newRepoURLs := make([]repos.RepoURL, 0)
	for _, repoURL := range repoURLs {
		newRepoURLs = append(newRepoURLs, *repoURL)
	}
	return newRepoURLs, nil
}
