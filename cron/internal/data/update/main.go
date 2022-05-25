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

// Package main updates projects repositories with a projects dependencies.
package main

import (
	"bytes"
	"os"

	"github.com/ossf/scorecard/v4/cron/internal/data"
)

// Adds "project=${PROJECT},dependency=true" to the repositories metadata.
// Args:
//     file path to old_projects.csv new_projects.csv
func main() {
	if len(os.Args) != 3 {
		panic("must provide 2 arguments")
	}

	inFile, err := os.OpenFile(os.Args[1], os.O_RDONLY, 0o644)
	if err != nil {
		panic(err)
	}
	defer inFile.Close()
	oldRepos, newRepos, err := getDependencies(inFile)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	if err := data.SortAndAppendTo(&buf, oldRepos, newRepos); err != nil {
		panic(err)
	}

	projects, err := os.OpenFile(os.Args[2], os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		panic(err)
	}
	if _, err := projects.Write(buf.Bytes()); err != nil {
		panic(err)
	}
}
