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

// Package main implements the PubSub controller.
package main

import (
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/ossf/scorecard/v4/cron/internal/data"
)

func main() {
	if len(os.Args) != 4 {
		panic("must provide exactly 3 arguments")
	}

	n, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(err)
	}

	inFile, err := os.OpenFile(os.Args[2], os.O_RDONLY, 0o644)
	if err != nil {
		panic(err)
	}
	iter, err := data.MakeIteratorFrom(inFile)
	if err != nil {
		panic(err)
	}

	outFile, err := os.OpenFile(os.Args[3], os.O_WRONLY|os.O_CREATE, 0o755)
	if err != nil {
		panic(err)
	}
	var repoURLs []data.RepoFormat
	for iter.HasNext() {
		repo, err := iter.Next()
		if err != nil {
			panic(err)
		}
		repoURLs = append(repoURLs, repo)
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(repoURLs), func(i, j int) {
		repoURLs[i], repoURLs[j] = repoURLs[j], repoURLs[i]
	})
	if err := data.WriteTo(outFile, repoURLs[:n]); err != nil {
		panic(err)
	}
}
