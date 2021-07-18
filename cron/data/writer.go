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

package data

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"

	"github.com/jszwec/csvutil"

	"github.com/ossf/scorecard/repos"
)

func repoFormatFromRepoURL(repoURLs []repos.RepoURL) []repoFormat {
	repoentries := make([]repoFormat, 0)
	for _, repoURL := range repoURLs {
		repoentry := repoFormat{
			Repo:     repoURL.URL(),
			Metadata: repoURL.Metadata,
		}
		repoentries = append(repoentries, repoentry)
	}
	return repoentries
}

func SortAndAppendTo(out io.Writer, oldRepos, newRepos []repos.RepoURL) error {
	repoentries := repoFormatFromRepoURL(oldRepos)
	repoentries = append(repoentries, repoFormatFromRepoURL(newRepos)...)

	sort.SliceStable(repoentries, func(i, j int) bool {
		return repoentries[i].Repo < repoentries[j].Repo
	})
	csvWriter := csv.NewWriter(out)
	enc := csvutil.NewEncoder(csvWriter)
	if err := enc.Encode(repoentries); err != nil {
		return fmt.Errorf("error during Encode: %w", err)
	}
	csvWriter.Flush()
	return nil
}

func SortAndAppendFrom(in io.Reader, out io.Writer, newRepos []repos.RepoURL) error {
	iter, err := MakeIteratorFrom(in)
	if err != nil {
		return fmt.Errorf("error during MakeIterator: %w", err)
	}

	oldRepos := make([]repos.RepoURL, 0)
	for iter.HasNext() {
		repo, err := iter.Next()
		if err != nil {
			return fmt.Errorf("error during iter.Next: %w", err)
		}
		oldRepos = append(oldRepos, repo)
	}
	return SortAndAppendTo(out, oldRepos, newRepos)
}
