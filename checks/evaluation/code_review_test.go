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

package evaluation

import (
	"testing"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
	scut "github.com/ossf/scorecard/v4/utests"
)

func TestCodeReview(t *testing.T) {
	t.Parallel()

	//nolint:govet // ignore since this is a test.
	tests := []struct {
		name     string
		expected scut.TestReturn
		rawData  *checker.CodeReviewData
	}{
		{
			name: "NullRawData",
			expected: scut.TestReturn{
				Error: sce.ErrScorecardInternal,
				Score: checker.InconclusiveResultScore,
			},
			rawData: nil,
		},
		{
			name: "NoCommits",
			expected: scut.TestReturn{
				Score: checker.InconclusiveResultScore,
			},
			rawData: &checker.CodeReviewData{},
		},
		{
			name: "NoReviews",
			expected: scut.TestReturn{
				Score: checker.MinResultScore,
			},
			rawData: &checker.CodeReviewData{
				DefaultBranchChangesets: []checker.Changeset{
					{
						Commits: []clients.Commit{{SHA: "1"}},
					},
					{
						Commits: []clients.Commit{{SHA: "1"}},
					},
				},
			},
		},
		{
			name: "all changesets reviewed",
			expected: scut.TestReturn{
				Score: checker.MaxResultScore,
			},
			rawData: &checker.CodeReviewData{
				DefaultBranchChangesets: []checker.Changeset{
					{
						ReviewPlatform: checker.ReviewPlatformGitHub,
						RevisionID:     "1",
						Commits: []clients.Commit{
							{
								SHA: "1",
								AssociatedMergeRequest: clients.PullRequest{
									Reviews: []clients.Review{
										{
											State: "APPROVED",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "all changesets reviewed outside github",
			expected: scut.TestReturn{
				Score: checker.MaxResultScore,
			},
			rawData: &checker.CodeReviewData{
				DefaultBranchChangesets: []checker.Changeset{
					{
						ReviewPlatform: checker.ReviewPlatformGerrit,
						RevisionID:     "1",
						Commits:        []clients.Commit{{SHA: "1"}},
					},
				},
			},
		},
		{
			name: "implicit maintainer approval through github merge",
			expected: scut.TestReturn{
				Score: checker.MaxResultScore,
			},
			rawData: &checker.CodeReviewData{
				DefaultBranchChangesets: []checker.Changeset{
					{
						ReviewPlatform: checker.ReviewPlatformGitHub,
						Commits:        []clients.Commit{{SHA: "1"}},
					},
					{
						ReviewPlatform: checker.ReviewPlatformGitHub,
						Commits:        []clients.Commit{{SHA: "2"}},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dl := &scut.TestDetailLogger{}
			res := CodeReview(tt.name, dl, tt.rawData)
			if !scut.ValidateTestReturn(t, tt.name, &tt.expected, &res, dl) {
				t.Error()
			}
		})
	}
}
