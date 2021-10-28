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
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v3/cron/data"
	"github.com/ossf/scorecard/v3/repos"
)

type fields struct {
	host     string
	owner    string
	repo     string
	metadata []string
}

//nolint: gocritic
func lessThan(x, y repos.RepoURI) bool {
	return fmt.Sprintf("%+v", x) < fmt.Sprintf("%+v", y)
}

func TestGetRepoURLs(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name, filename string
		outcome        []fields
	}{
		{
			name:     "NoChange",
			filename: "testdata/no_change.csv",
			outcome: []fields{
				{
					host:     "github.com",
					owner:    "owner1",
					repo:     "repo1",
					metadata: []string{"meta1", "meta2"},
				},
				{
					host:  "github.com",
					owner: "owner2",
					repo:  "repo2",
				},
			},
		},
		{
			name:     "AddMetadata",
			filename: "testdata/add_metadata.csv",
			outcome: []fields{
				{
					host:     "github.com",
					owner:    "owner1",
					repo:     "repo1",
					metadata: []string{"meta1", "meta2"},
				},
				{
					host:     "github.com",
					owner:    "owner2",
					repo:     "repo2",
					metadata: []string{"meta1"},
				},
			},
		},
		{
			name:     "SkipLatest",
			filename: "testdata/skip_latest.csv",
			outcome: []fields{
				{
					host:     "github.com",
					owner:    "owner1",
					repo:     "repo1",
					metadata: []string{"meta1", "meta2"},
				},
				{
					host:  "github.com",
					owner: "owner2",
					repo:  "repo2",
				},
			},
		},
		{
			name:     "SkipEmpty",
			filename: "testdata/skip_empty.csv",
			outcome: []fields{
				{
					host:     "github.com",
					owner:    "owner1",
					repo:     "repo1",
					metadata: []string{"meta1", "meta2"},
				},
				{
					host:     "github.com",
					owner:    "owner2",
					repo:     "repo2",
					metadata: []string{"meta3"},
				},
			},
		},
		{
			name:     "SkipEmpty_2",
			filename: "testdata/skip_empty_2.csv",
			outcome: []fields{
				{
					host:     "github.com",
					owner:    "owner1",
					repo:     "repo1",
					metadata: []string{"meta1", "meta2"},
				},
				{
					host:     "github.com",
					owner:    "owner2",
					repo:     "repo2",
					metadata: []string{"meta3"},
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

			// Create the list of RepoURL from the outcome.
			var rs []repos.RepoURI
			for _, r := range testcase.outcome {
				u := fmt.Sprintf("%s/%s/%s", r.host, r.owner, r.repo)
				outcomeRepo, err := repos.NewFromURL(u)
				if err != nil {
					t.Errorf("repos.NewFromURL: %v", err)
				}
				if err := outcomeRepo.AppendMetadata(r.metadata...); err != nil {
					t.Errorf("outcomeRepo.AppendMetadata: %v", err)
				}
				rs = append(rs, *outcomeRepo)
			}

			// Export all private fields for comparison.
			exp := cmp.Exporter(func(t reflect.Type) bool {
				return true
			})

			if !cmp.Equal(rs, repoURLs, exp, cmpopts.EquateEmpty(), cmpopts.SortSlices(lessThan)) {
				t.Errorf("testcase failed. expected %+v, got %+v", rs, repoURLs)
			}
		})
	}
}
