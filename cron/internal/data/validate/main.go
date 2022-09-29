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

// Package main validates an input file with a list of projects.
package main

import (
	"log"
	"os"

	"github.com/ossf/scorecard/v4/cron/data"
)

// Validates data.Iterator used by production PubSub cron job.
// * Check for no duplicates in repoURLs.
// * Check repoURL is a valid GitHub URL.
func main() {
	if len(os.Args) != 2 {
		panic("must provide single argument")
	}

	inFile, err := os.OpenFile(os.Args[1], os.O_RDONLY, 0o644)
	if err != nil {
		panic(err)
	}
	iter, err := data.MakeIteratorFrom(inFile)
	if err != nil {
		panic(err)
	}

	m := make(map[string]bool)
	for iter.HasNext() {
		repo, err := iter.Next()
		if err != nil {
			panic(err)
		}
		if _, ok := m[repo.Repo]; ok {
			log.Panicf("Item already in the list %s", repo.Repo)
		}
		m[repo.Repo] = true
	}
}
