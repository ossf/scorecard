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

package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v4/cron/data"
)

func lessThanURI(x, y data.RepoFormat) bool {
	return fmt.Sprintf("%+v", x) < fmt.Sprintf("%+v", y)
}

func TestGetRepoURLs(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name, filename string
		outcome        []data.RepoFormat
	}{
		{
			name:     "NoChange",
			filename: "testdata/no_change.csv",
			outcome: []data.RepoFormat{
				{
					Repo:     "github.com/owner1/repo1",
					Metadata: []string{"meta1", "meta2"},
				},
				{
					Repo: "github.com/owner2/repo2",
				},
			},
		},
		{
			name:     "AddMetadata",
			filename: "testdata/add_metadata.csv",
			outcome: []data.RepoFormat{
				{
					Repo:     "github.com/owner1/repo1",
					Metadata: []string{"meta1", "meta2"},
				},
				{
					Repo:     "github.com/owner2/repo2",
					Metadata: []string{"meta1"},
				},
			},
		},
		{
			name:     "SkipLatest",
			filename: "testdata/skip_latest.csv",
			outcome: []data.RepoFormat{
				{
					Repo:     "github.com/owner1/repo1",
					Metadata: []string{"meta1", "meta2"},
				},
				{
					Repo: "github.com/owner2/repo2",
				},
			},
		},
		{
			name:     "SkipEmpty",
			filename: "testdata/skip_empty.csv",
			outcome: []data.RepoFormat{
				{
					Repo:     "github.com/owner1/repo1",
					Metadata: []string{"meta1", "meta2"},
				},
				{
					Repo:     "github.com/owner2/repo2",
					Metadata: []string{"meta3"},
				},
			},
		},
		{
			name:     "SkipEmpty_2",
			filename: "testdata/skip_empty_2.csv",
			outcome: []data.RepoFormat{
				{
					Repo:     "github.com/owner1/repo1",
					Metadata: []string{"meta1", "meta2"},
				},
				{
					Repo:     "github.com/owner2/repo2",
					Metadata: []string{"meta3"},
				},
			},
		},
		{
			name:     "SkipDuplicates",
			filename: "testdata/skip_duplicates.csv",
			outcome: []data.RepoFormat{
				{
					Repo:     "github.com/owner1/repo1",
					Metadata: []string{"meta1", "meta2"},
				},
				{
					Repo:     "github.com/owner2/repo2",
					Metadata: []string{"meta3", "meta4"},
				},
			},
		},
	}
	for _, testcase := range testcases {
		testcase := testcase

		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()
			testFile, err := os.OpenFile(testcase.filename, os.O_RDONLY, 0o644)
			if err != nil {
				t.Errorf("testcase failed to open %s: %v", testcase.filename, err)
			}
			defer testFile.Close()

			iter, err := data.MakeIteratorFrom(testFile)
			if err != nil {
				t.Errorf("testcase failed to MakeIterator: %v", err)
			}

			repoURLs, err := getRepoURLs(iter)
			if err != nil {
				t.Errorf("testcase failed: %v", err)
			}

			if !cmp.Equal(testcase.outcome, repoURLs, cmpopts.EquateEmpty(), cmpopts.SortSlices(lessThanURI)) {
				t.Errorf("testcase failed. expected equal, got diff: %s", cmp.Diff(testcase.outcome, repoURLs))
			}
		})
	}
}
