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
	"testing"

	"github.com/ossf/scorecard/v2/repos"
)

func TestCsvWriter(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name     string
		out      string
		oldRepos []repos.RepoURL
		newRepos []repos.RepoURL
	}{
		{
			name: "Basic",
			oldRepos: []repos.RepoURL{
				{
					Host:     "github.com",
					Owner:    "owner1",
					Repo:     "repo1",
					Metadata: []string{"meta1"},
				},
			},
			newRepos: []repos.RepoURL{
				{
					Host:     "github.com",
					Owner:    "owner2",
					Repo:     "repo2",
					Metadata: []string{"meta2"},
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
