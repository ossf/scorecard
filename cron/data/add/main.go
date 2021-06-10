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
//     path to output.csv file
func main() {
	iter, err := data.MakeIterator()
	if err != nil {
		panic(err)
	}

	repoURLs := make([]repos.RepoURL, 0)
	repoMap := make(map[string]bool)
	for iter.HasNext() {
		repo, err := iter.Next()
		if err != nil {
			panic(err)
		}
		if _, ok := repoMap[repo.URL()]; ok {
			continue
		}
		repoURLs = append(repoURLs, repo)
		repoMap[repo.URL()] = true
	}

	var buf bytes.Buffer
	if err := data.SortAndAppendTo(&buf, repoURLs, nil); err != nil {
		panic(err)
	}
	projects, err := os.OpenFile(os.Args[1], os.O_WRONLY, 0o644)
	if err != nil {
		panic(err)
	}
	if _, err := projects.Write(buf.Bytes()); err != nil {
		panic(err)
	}
}
