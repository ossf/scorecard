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
	"bytes"
	"fmt"
	"testing"

	"github.com/ossf/scorecard/v3/repos"
)

func TestCsvWriter(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name     string
		out      string
		oldRepos []fields
		newRepos []fields
	}{
		{
			name: "Basic",
			oldRepos: []fields{
				{
					host:     "github.com",
					owner:    "owner1",
					repo:     "repo1",
					metadata: []string{"meta1"},
				},
			},
			newRepos: []fields{
				{
					host:     "github.com",
					owner:    "owner2",
					repo:     "repo2",
					metadata: []string{"meta2"},
				},
			},
			out: `repo,metadata
github.com/owner1/repo1,meta1
github.com/owner2/repo2,meta2
`,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			var oldRepos []repos.RepoURI
			var newRepos []repos.RepoURI
			for _, v := range testcase.oldRepos {
				r, err := repos.NewFromURL(fmt.Sprintf("%s/%s/%s", v.host, v.owner, v.repo))
				if err != nil {
					t.Errorf("repos.NewFromURL: %v", err)
				}
				if err = r.AppendMetadata(v.metadata...); err != nil {
					t.Errorf("r.AppendMetadata: %v", err)
				}
				oldRepos = append(oldRepos, *r)
			}
			for _, v := range testcase.newRepos {
				r, err := repos.NewFromURL(fmt.Sprintf("%s/%s/%s", v.host, v.owner, v.repo))
				if err != nil {
					t.Errorf("repos.NewFromURL: %v", err)
				}
				if err = r.AppendMetadata(v.metadata...); err != nil {
					t.Errorf("r.AppendMetadata: %v", err)
				}
				newRepos = append(newRepos, *r)
			}

			err := SortAndAppendTo(&buf, oldRepos, newRepos)
			if err != nil {
				t.Errorf("error while running testcase: %v", err)
			}
			if buf.String() != testcase.out {
				t.Errorf("\nexpected: \n%s \ngot: \n%s", testcase.out, buf.String())
			}
		})
	}
}
