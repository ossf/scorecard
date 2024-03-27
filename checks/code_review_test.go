// Copyright 2023 OpenSSF Scorecard Authors
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

package checks

import (
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
	sce "github.com/ossf/scorecard/v4/errors"
	scut "github.com/ossf/scorecard/v4/utests"
)

// TestCodeReview tests the code review checker.
func TestCodereview(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		commitErr error
		commits   []clients.Commit
		expected  scut.TestReturn
	}{
		{
			name: "no commits",
			expected: scut.TestReturn{
				Score: -1,
			},
		},
		{
			name:      "no commits due to error",
			commitErr: errors.New("error fetching commits"),
			expected: scut.TestReturn{
				Score: -1,
				Error: sce.ErrScorecardInternal,
			},
		},
		{
			name: "Valid GitHub PR",
			commits: []clients.Commit{
				{
					SHA: "sha",
					Committer: clients.User{
						Login: "bob",
					},
					AssociatedMergeRequest: clients.PullRequest{
						Number:   1,
						MergedAt: time.Now(),
						Reviews: []clients.Review{
							{
								Author: &clients.User{Login: "alice"},
								State:  "APPROVED",
							},
						},
					},
				},
			},
			expected: scut.TestReturn{
				Score: 10,
			},
		},
		{
			name: "Valid Prow PR as not a bot",
			commits: []clients.Commit{
				{
					SHA: "sha",
					Committer: clients.User{
						Login: "user",
					},
					AssociatedMergeRequest: clients.PullRequest{
						Number:   1,
						MergedAt: time.Now(),
						Labels: []clients.Label{
							{
								Name: "lgtm",
							},
						},
					},
				},
			},
			expected: scut.TestReturn{
				Score: 10,
			},
		},
		{
			name: "Valid Prow PR and commits as bot",
			commits: []clients.Commit{
				{
					SHA: "sha",
					Committer: clients.User{
						Login: "bot",
					},
					AssociatedMergeRequest: clients.PullRequest{
						Number:   1,
						MergedAt: time.Now(),
						Labels: []clients.Label{
							{
								Name: "lgtm",
							},
						},
					},
				},
			},
			expected: scut.TestReturn{
				Score: 10,
			},
		},
		{
			name: "Valid PR's and commits with merged by someone else",
			commits: []clients.Commit{
				{
					SHA: "sha",
					Committer: clients.User{
						Login: "bob",
					},
					AssociatedMergeRequest: clients.PullRequest{
						Number: 1,
						Labels: []clients.Label{
							{
								Name: "lgtm",
							},
						},
					},
				},
			},
			expected: scut.TestReturn{
				Score: 0,
			},
		},
		{
			name: "2 PRs 1 review on GitHub",
			commits: []clients.Commit{
				{
					SHA: "a",
					Committer: clients.User{
						Login: "bob",
					},
					AssociatedMergeRequest: clients.PullRequest{
						Number:   1,
						MergedAt: time.Now(),
						Reviews: []clients.Review{
							{
								Author: &clients.User{Login: "alice"},
								State:  "APPROVED",
							},
						},
					},
				},
				{
					SHA: "sha2",
					Committer: clients.User{
						Login: "bob",
					},
				},
			},
			expected: scut.TestReturn{
				Score: 5,
			},
		},
		{
			name: "implicit maintainer approval through merge",
			commits: []clients.Commit{
				{
					SHA: "abc",
					Committer: clients.User{
						Login: "bob",
					},
					AssociatedMergeRequest: clients.PullRequest{
						Number:   1,
						MergedAt: time.Now(),
						MergedBy: clients.User{Login: "alice"},
					},
				},
				{
					SHA: "def",
					Committer: clients.User{
						Login: "bob",
					},
				},
			},
			expected: scut.TestReturn{
				Score: 5,
			},
		},
		{
			name: "Valid Phabricator commit",
			commits: []clients.Commit{
				{
					SHA: "sha",
					Committer: clients.User{
						Login: "bob",
					},
					Message: "Title\nReviewed By: alice\nDifferential Revision: PHAB234",
				},
			},
			expected: scut.TestReturn{
				Score: 10,
			},
		},
		{
			name: "Phabricator like, missing differential",
			commits: []clients.Commit{
				{
					SHA: "sha",
					Committer: clients.User{
						Login: "bob",
					},
					Message: "Title\nReviewed By: alice",
				},
			},
			expected: scut.TestReturn{
				Score: 0,
			},
		},
		{
			name: "Phabricator like, missing reviewed by",
			commits: []clients.Commit{
				{
					SHA: "sha",
					Committer: clients.User{
						Login: "bob",
					},
					Message: "Title\nDifferential Revision: PHAB234",
				},
			},
			expected: scut.TestReturn{
				Score: checker.MaxResultScore,
			},
		},
		{
			name: "Valid piper commit",
			commits: []clients.Commit{
				{
					SHA: "sha",
					Committer: clients.User{
						Login: "",
					},
					Message: "Title\nPiperOrigin-RevId: 444529962",
				},
			},
			expected: scut.TestReturn{
				Score: 10,
			},
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockRepo := mockrepo.NewMockRepoClient(ctrl)
			mockRepo.EXPECT().ListCommits().Return(tt.commits, tt.commitErr).AnyTimes()

			var dl scut.TestDetailLogger
			req := checker.CheckRequest{
				RepoClient: mockRepo,
				Dlogger:    &dl,
			}
			res := CodeReview(&req)
			if tt.commitErr != nil && res.Error == nil {
				t.Fatalf("Expected error %v, got nil", tt.commitErr)
			}
			scut.ValidateTestReturn(t, tt.name, &tt.expected, &res, &dl)
		})
	}
}
