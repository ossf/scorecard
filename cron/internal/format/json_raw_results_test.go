// Copyright 2022 OpenSSF Scorecard Authors
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

package format

import (
	"reflect"
	"testing"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
)

func TestAddCodeReviewRawResults(t *testing.T) {
	t.Parallel()

	r := jsonScorecardRawResult{}
	cr := checker.CodeReviewData{
		DefaultBranchChangesets: []checker.Changeset{
			{
				RevisionID: "id1",
				Commits: []clients.Commit{
					{
						Committer: clients.User{
							Login: "user1",
						},
						Message: "commit1",
						SHA:     "sha1",
					},
					{
						Committer: clients.User{
							Login: "user2",
						},
						Message: "commit2",
						SHA:     "sha2",
					},
				},
			},
		},
	}
	if err := addCodeReviewRawResults(&r, &cr); err != nil {
		t.Errorf("addCodeReviewRawResults: %v", err)
	}
	want := []jsonDefaultBranchChangeset{
		{
			RevisionID: "id1",
			Commits: []jsonCommit{
				{
					Committer: jsonUser{
						Login: "user1",
					},
					Message: "commit1",
					SHA:     "sha1",
				},
				{
					Committer: jsonUser{
						Login: "user2",
					},
					Message: "commit2",
					SHA:     "sha2",
				},
			},
		},
	}
	if !reflect.DeepEqual(r.Results.DefaultBranchChangesets, want) {
		t.Errorf("Expected %v, got %v", want, r.Results.DefaultBranchChangesets)
	}
}
