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
	const testRepo = "test-repo"

	// fieldalignment lint issue. Ignoring it as it is not important for this test.
	// nolint: govet, goerr113
	tests := []struct {
		err       error
		name      string
		commiterr error
		commits   []clients.Commit
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
			commits: []clients.Commit{
				{
					SHA: "sha",
					Committer: clients.User{
						Login: "user",
					},
					AssociatedMergeRequest: clients.PullRequest{
						Repository: testRepo,
						Number:     1,
						MergedAt:   time.Now(),
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
			},
			expected: checker.CheckResult{
				Score: 10,
				Pass:  true,
			},
		},
		{
			name: "Valid PR's and commits as bot",
			commits: []clients.Commit{
				{
					SHA: "sha",
					Committer: clients.User{
						Login: "bot",
					},
					AssociatedMergeRequest: clients.PullRequest{
						Repository: testRepo,
						Number:     1,
						MergedAt:   time.Now(),
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
			},
			expected: checker.CheckResult{
				Score: 10,
				Pass:  true,
			},
		},
		{
			name: "Valid PR's and commits not yet merged",
			commits: []clients.Commit{
				{
					SHA: "sha",
					Committer: clients.User{
						Login: "bot",
					},
					AssociatedMergeRequest: clients.PullRequest{
						Repository: testRepo,
						Number:     1,
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
			},
			expected: checker.CheckResult{
				Score: -1,
			},
		},
		{
			name: "Valid PR's and commits with merged by someone else",
			commits: []clients.Commit{
				{
					SHA: "sha",
					Committer: clients.User{
						Login: "bot",
					},
					AssociatedMergeRequest: clients.PullRequest{
						Repository: testRepo,
						Number:     1,
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
			},
			expected: checker.CheckResult{
				Score: -1,
			},
		},
		{
			name: "Merged PR in a different repo",
			commits: []clients.Commit{
				{
					SHA: "sha",
					Committer: clients.User{
						Login: "bot",
					},
					AssociatedMergeRequest: clients.PullRequest{
						Repository: "does-not-exist",
						MergedAt:   time.Now(),
						Number:     1,
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
			mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
			mockRepoClient.EXPECT().ListCommits().DoAndReturn(func() ([]clients.Commit, error) {
				if tt.commiterr != nil {
					return tt.commits, tt.commiterr
				}
				return tt.commits, tt.err
			}).AnyTimes()
			mockRepo := mockrepo.NewMockRepo(ctrl)
			mockRepo.EXPECT().String().Return(testRepo).AnyTimes()

			req := checker.CheckRequest{
				RepoClient: mockRepoClient,
				Repo:       mockRepo,
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
