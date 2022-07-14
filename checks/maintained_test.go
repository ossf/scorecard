// Copyright 2020 Security Scorecard Authors
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

// ignoring the linter for cyclomatic complexity because it is a test func
// TestMaintained tests the maintained check.
//nolint
func Test_Maintained(t *testing.T) {
	t.Parallel()
	threeHundredDaysAgo := time.Now().AddDate(0, 0, -300)
	twoHundredDaysAgo := time.Now().AddDate(0, 0, -200)
	fiveDaysAgo := time.Now().AddDate(0, 0, -5)
	oneDayAgo := time.Now().AddDate(0, 0, -1)
	ownerAssociation := clients.RepoAssociationOwner
	noneAssociation := clients.RepoAssociationNone
	// fieldalignment lint issue. Ignoring it as it is not important for this test.
	someone := clients.User{
		Login: "someone",
	}
	otheruser := clients.User{
		Login: "someone-else",
	}
	//nolint
	tests := []struct {
		err        error
		name       string
		isarchived bool
		archiveerr error
		commits    []clients.Commit
		commiterr  error
		issues     []clients.Issue
		issueerr   error
		expected   checker.CheckResult
	}{
		{
			name:       "archived",
			isarchived: true,
			expected: checker.CheckResult{
				Score: 0,
			},
		},
		{
			name:       "archived",
			isarchived: true,
			archiveerr: errors.New("error"),
			expected: checker.CheckResult{
				Score: -1,
			},
		},
		{
			name:       "commit lookup error",
			isarchived: false,
			commits:    []clients.Commit{},
			commiterr:  errors.New("error"),
			issues:     []clients.Issue{},
			expected: checker.CheckResult{
				Score: -1,
			},
		},
		{
			name:       "issue lookup error",
			isarchived: false,
			issueerr:   errors.New("error"),
			issues:     []clients.Issue{},
			expected: checker.CheckResult{
				Score: -1,
			},
		},
		{
			name:       "repo with no commits or issues",
			isarchived: false,
			commits:    []clients.Commit{},
			issues:     []clients.Issue{},
			expected: checker.CheckResult{
				Score: 0,
			},
		},
		{
			name:       "repo with valid commits",
			isarchived: false,
			commits: []clients.Commit{
				{
					CommittedDate: time.Now().AddDate(0, 0, -1),
				},
				{
					CommittedDate: time.Now().AddDate(0, 0, -10),
				},
				{
					CommittedDate: time.Now().AddDate(0, 0, -11),
				},
				{
					CommittedDate: time.Now().AddDate(0, 0, -12),
				},
			},
			issues: []clients.Issue{},
			expected: checker.CheckResult{
				Score: 3,
			},
		},
		{
			name:       "old issues, no comments",
			isarchived: false,
			commits:    []clients.Commit{},
			issues: []clients.Issue{
				{
					CreatedAt:         &threeHundredDaysAgo,
					AuthorAssociation: &ownerAssociation,
					Author:            &someone,
				},
				{
					CreatedAt:         &twoHundredDaysAgo,
					AuthorAssociation: &noneAssociation,
					Author:            &someone,
				},
			},
			expected: checker.CheckResult{
				Score: 0,
			},
		},
		{
			name:       "new issues by non-associated users",
			isarchived: false,
			commits:    []clients.Commit{},
			issues: []clients.Issue{
				{
					CreatedAt:         &fiveDaysAgo,
					AuthorAssociation: &noneAssociation,
					Author:            &someone,
				},
				{
					CreatedAt:         &oneDayAgo,
					AuthorAssociation: &noneAssociation,
					Author:            &someone,
				},
			},
			expected: checker.CheckResult{
				Score: 0,
			},
		},
		{
			name:       "new issues with comments by non-associated users",
			isarchived: false,
			commits:    []clients.Commit{},
			issues: []clients.Issue{
				{
					CreatedAt:         &fiveDaysAgo,
					AuthorAssociation: &noneAssociation,
					Comments: []clients.IssueComment{
						{
							CreatedAt:         &oneDayAgo,
							AuthorAssociation: &noneAssociation,
							Author:            &someone,
						},
					},
				},
				{
					CreatedAt:         &oneDayAgo,
					AuthorAssociation: &noneAssociation,
					Comments: []clients.IssueComment{
						{
							CreatedAt:         &oneDayAgo,
							AuthorAssociation: &noneAssociation,
							Author:            &someone,
						},
					},
				},
			},
			expected: checker.CheckResult{
				Score: 0,
			},
		},
		{
			name:       "old issues with old comments by owner",
			isarchived: false,
			commits:    []clients.Commit{},
			issues: []clients.Issue{
				{
					CreatedAt:         &twoHundredDaysAgo,
					AuthorAssociation: &noneAssociation,
					Comments: []clients.IssueComment{
						{
							CreatedAt:         &twoHundredDaysAgo,
							AuthorAssociation: &ownerAssociation,
							Author:            &someone,
						},
					},
				},
				{
					CreatedAt:         &threeHundredDaysAgo,
					AuthorAssociation: &noneAssociation,
					Comments: []clients.IssueComment{
						{
							CreatedAt:         &twoHundredDaysAgo,
							AuthorAssociation: &ownerAssociation,
							Author:            &someone,
						},
					},
				},
			},
			expected: checker.CheckResult{
				Score: 0,
			},
		},
		{
			name:       "old issues with new comments by owner",
			isarchived: false,
			commits:    []clients.Commit{},
			issues: []clients.Issue{
				{
					CreatedAt:         &twoHundredDaysAgo,
					AuthorAssociation: &noneAssociation,
					Comments: []clients.IssueComment{
						{
							CreatedAt:         &fiveDaysAgo,
							AuthorAssociation: &ownerAssociation,
							Author:            &someone,
						},
					},
				},
				{
					CreatedAt:         &threeHundredDaysAgo,
					AuthorAssociation: &noneAssociation,
					Comments: []clients.IssueComment{
						{
							CreatedAt:         &oneDayAgo,
							AuthorAssociation: &ownerAssociation,
							Author:            &someone,
						},
					},
				},
			},
			expected: checker.CheckResult{
				Score: 1,
			},
		},
		{
			name:       "new issues by owner",
			isarchived: false,
			commits:    []clients.Commit{},
			issues: []clients.Issue{
				{
					CreatedAt:         &fiveDaysAgo,
					AuthorAssociation: &ownerAssociation,
					Author:            &someone,
				},
				{
					CreatedAt:         &oneDayAgo,
					AuthorAssociation: &ownerAssociation,
					Author:            &someone,
				},
			},
			expected: checker.CheckResult{
				Score: 1,
			},
		},
		{
			name:       "new issues by non-owner",
			isarchived: false,
			commits:    []clients.Commit{},
			issues: []clients.Issue{
				{
					CreatedAt:         &fiveDaysAgo,
					AuthorAssociation: &noneAssociation,
					Author:            &otheruser,
				},
				{
					CreatedAt:         &oneDayAgo,
					AuthorAssociation: &noneAssociation,
					Author:            &otheruser,
				},
			},
			expected: checker.CheckResult{
				Score: 0,
			},
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockRepo := mockrepo.NewMockRepoClient(ctrl)

			mockRepo.EXPECT().IsArchived().DoAndReturn(func() (bool, error) {
				if tt.archiveerr != nil {
					return false, tt.archiveerr
				}
				return tt.isarchived, nil
			})

			if tt.archiveerr == nil {
				mockRepo.EXPECT().ListCommits().DoAndReturn(
					func() ([]clients.Commit, error) {
						if tt.commiterr != nil {
							return nil, tt.commiterr
						}
						return tt.commits, tt.err
					},
				).MinTimes(1)

				if tt.commiterr == nil {
					mockRepo.EXPECT().ListIssues().DoAndReturn(
						func() ([]clients.Issue, error) {
							if tt.issueerr != nil {
								return nil, tt.issueerr
							}
							return tt.issues, tt.err
						},
					).MinTimes(1)
				}
			}

			req := checker.CheckRequest{
				RepoClient: mockRepo,
			}
			req.Dlogger = &scut.TestDetailLogger{}
			res := Maintained(&req)

			if tt.err != nil {
				if res.Error == nil {
					t.Errorf("Expected error %v, got nil", tt.err)
				}
				// return as we don't need to check the rest of the fields.
				return
			}

			if res.Score != tt.expected.Score {
				t.Errorf("Expected score %d, got %d for %v", tt.expected.Score, res.Score, tt.name)
			}
			ctrl.Finish()
		})
	}
}
