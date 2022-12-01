// Copyright 2021 OpenSSF Scorecard Authors
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
)

// WriteTo writes `repos` to `out`.
func WriteTo(out io.Writer, repos []RepoFormat) error {
	csvWriter := csv.NewWriter(out)
	enc := csvutil.NewEncoder(csvWriter)
	if err := enc.Encode(repos); err != nil {
		return fmt.Errorf("error during Encode: %w", err)
	}
	csvWriter.Flush()
	return nil
}

// SortAndAppendTo appends `oldRepos` and `newRepos` before sorting and writing out the result to `out`.
func SortAndAppendTo(out io.Writer, oldRepos, newRepos []RepoFormat) error {
	oldRepos = append(oldRepos, newRepos...)

	sort.SliceStable(oldRepos, func(i, j int) bool {
		return oldRepos[i].Repo < oldRepos[j].Repo
	})
	return WriteTo(out, oldRepos)
}

// SortAndAppendFrom reads from `in`, appends to newRepos and writes the sorted output to `out`.
func SortAndAppendFrom(in io.Reader, out io.Writer, newRepos []RepoFormat) error {
	iter, err := MakeIteratorFrom(in)
	if err != nil {
		return fmt.Errorf("error during MakeIterator: %w", err)
	}

	oldRepos := make([]RepoFormat, 0)
	for iter.HasNext() {
		repo, err := iter.Next()
		if err != nil {
			return fmt.Errorf("error during iter.Next: %w", err)
		}
		oldRepos = append(oldRepos, repo)
	}
	return SortAndAppendTo(out, oldRepos, newRepos)
}
