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
	"bytes"
	"testing"
)

func TestCsvWriter(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name     string
		out      string
		oldRepos []RepoFormat
		newRepos []RepoFormat
	}{
		{
			name: "Basic",
			oldRepos: []RepoFormat{
				{
					Repo:     "github.com/owner1/repo1",
					Metadata: []string{"meta1"},
				},
				{
					Repo:     "gitlab.com/owner3/repo3",
					Metadata: []string{"meta3"},
				},
			},
			newRepos: []RepoFormat{
				{
					Repo:     "github.com/owner2/repo2",
					Metadata: []string{"meta2"},
				},
				{
					Repo:     "gitlab.com/owner4/repo4",
					Metadata: []string{"meta4"},
				},
			},
			out: `repo,metadata
github.com/owner1/repo1,meta1
github.com/owner2/repo2,meta2
gitlab.com/owner3/repo3,meta3
gitlab.com/owner4/repo4,meta4
`,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			err := SortAndAppendTo(&buf, testcase.oldRepos, testcase.newRepos)
			if err != nil {
				t.Errorf("error while running testcase: %v", err)
			}
			if buf.String() != testcase.out {
				t.Errorf("\nexpected: \n%s \ngot: \n%s", testcase.out, buf.String())
			}
		})
	}
}
