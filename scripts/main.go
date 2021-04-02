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
	"io/ioutil"
	"log"
	"os"

	"github.com/jszwec/csvutil"
)

type Repository struct {
	Repo     string `csv:"repo"`
	Metadata string `csv:"metadata,omitempty"`
}

// Checks for duplicate item in the projects.txt
// This is used in the builds to validate there aren't duplicates in projects.txt.
func main() {
	projects, err := os.OpenFile(os.Args[1], os.O_RDONLY, 0644)
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

	m := make(map[string]bool)
	for _, item := range repos {
		if _, ok := m[item.Repo]; ok {
			log.Panicf("Item already in the list %s", item.Repo)
		}
		m[item.Repo] = true
	}
}
