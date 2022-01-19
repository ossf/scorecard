// Copyright 2022 Security Scorecard Authors
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
	scut "github.com/ossf/scorecard/v4/utests"
)

// TestCodeReview tests the code review checker.
func TestCodereview(t *testing.T) {
	t.Parallel()
	//fieldalignment lint issue. Ignoring it as it is not important for this test.
	//nolint
	tests := []struct {
		err       error
		name      string
		commiterr error
		commits   []clients.Commit
		prs       []clients.PullRequest
		expected  checker.CheckResult
	}{
		{
			name: "no commits",
			expected: checker.CheckResult{
				Score: -1,
			},
		},
		{
			name:      "no commits with error",
			commiterr: errors.New("error"),
			expected: checker.CheckResult{
				Score: -1,
			},
		},
		{
			name: "no PR's with error",
			err:  errors.New("error"),
			expected: checker.CheckResult{
				Score: -1,
			},
		},
		{
			name:      "no PR's with error as well as commits",
			err:       errors.New("error"),
			commiterr: errors.New("error"),
			expected: checker.CheckResult{
				Score: -1,
			},
		},
		{
			name: "Valid PR's and commits as not a bot",
			prs: []clients.PullRequest{
				{
					Number:   1,
					MergedAt: time.Now(),
					Reviews: []clients.Review{
						{
							State: "APPROVED",
						},
					},
					Labels: []clients.Label{
						{
							Name: "lgtm",
						},
					},
				},
			},
			commits: []clients.Commit{
				{
					SHA: "sha",
					Committer: clients.User{
						Login: "user",
					},
				},
			},

			expected: checker.CheckResult{
				Score: 10,
				Pass:  true,
			},
		},

		{
			name: "Valid PR's and commits as bot",
			prs: []clients.PullRequest{
				{
					Number:   1,
					MergedAt: time.Now(),
					Reviews: []clients.Review{
						{
							State: "APPROVED",
						},
					},
					Labels: []clients.Label{
						{
							Name: "lgtm",
						},
					},
				},
			},
			commits: []clients.Commit{
				{
					SHA: "sha",
					Committer: clients.User{
						Login: "bot",
					},
				},
			},

			expected: checker.CheckResult{
				Score: 10,
				Pass:  true,
			},
		},

		{
			name: "Valid PR's and commits not yet merged",
			prs: []clients.PullRequest{
				{
					Number: 1,
					Reviews: []clients.Review{
						{
							State: "APPROVED",
						},
					},
					Labels: []clients.Label{
						{
							Name: "lgtm",
						},
					},
				},
			},
			commits: []clients.Commit{
				{
					SHA: "sha",
					Committer: clients.User{
						Login: "bot",
					},
				},
			},

			expected: checker.CheckResult{
				Score: -1,
			},
		},
		{
			name: "Valid PR's and commits with merged by someone else",
			prs: []clients.PullRequest{
				{
					Number: 1,
					Reviews: []clients.Review{
						{
							State: "APPROVED",
						},
					},
					Labels: []clients.Label{
						{
							Name: "lgtm",
						},
					},
					MergeCommit: clients.Commit{
						SHA: "sha",
						Committer: clients.User{
							Login: "bob",
						},
					},
				},
			},
			commits: []clients.Commit{
				{
					SHA: "sha",
					Committer: clients.User{
						Login: "bot",
					},
				},
			},

			expected: checker.CheckResult{
				Score: -1,
			},
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockRepo := mockrepo.NewMockRepoClient(ctrl)
			mockRepo.EXPECT().ListMergedPRs().DoAndReturn(func() ([]clients.PullRequest, error) {
				if tt.err != nil {
					return tt.prs, tt.err
				}
				return tt.prs, tt.err
			}).AnyTimes()
			mockRepo.EXPECT().ListCommits().DoAndReturn(func() ([]clients.Commit, error) {
				if tt.commiterr != nil {
					return tt.commits, tt.commiterr
				}
				return tt.commits, tt.err
			}).AnyTimes()

			req := checker.CheckRequest{
				RepoClient: mockRepo,
			}
			req.Dlogger = &scut.TestDetailLogger{}
			res := DoesCodeReview(&req)

			if tt.err != nil {
				if res.Error2 == nil {
					t.Errorf("Expected error %v, got nil", tt.err)
				}
				// return as we don't need to check the rest of the fields.
				return
			}

			if res.Score != tt.expected.Score {
				t.Errorf("Expected score %d, got %d for %v", tt.expected.Score, res.Score, tt.name)
			}
			if res.Pass != tt.expected.Pass {
				t.Errorf("Expected pass %t, got %t for %v", tt.expected.Pass, res.Pass, tt.name)
			}
			ctrl.Finish()
		})
	}
}
