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

package raw

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
)

// TestCodeReviews tests the CodeReviews function.
func Test_getChangesets(t *testing.T) {
	t.Parallel()
	var (
		commitC = clients.Commit{
			SHA: "c",
			AssociatedMergeRequest: clients.PullRequest{
				Number: 3,
				MergedAt: time.Date(2023 /*year*/, time.March, 21, /*day*/
					13 /*hour*/, 42 /*min*/, 0 /*sec*/, 0 /*nsec*/, time.UTC),
			},
			Message: "merge commitSHA c form GitHub",
		}
		commitB = clients.Commit{
			SHA: "b",
			AssociatedMergeRequest: clients.PullRequest{
				Number: 2,
				MergedAt: time.Date(2023 /*year*/, time.March, 21, /*day*/
					13 /*hour*/, 41 /*min*/, 0 /*sec*/, 0 /*nsec*/, time.UTC),
			},
			Message: "merge commitSHA b from GitHub",
		}
		commitBUnsquashed = clients.Commit{
			SHA: "b_unsquashed",
			AssociatedMergeRequest: clients.PullRequest{
				Number: 2,
				MergedAt: time.Date(2023 /*year*/, time.March, 21, /*day*/
					13 /*hour*/, 40 /*min*/, 0 /*sec*/, 0 /*nsec*/, time.UTC),
			},
			Message: "unsquashed commitSHA b_unsquashed from GitHub",
		}
		commitA = clients.Commit{
			SHA: "a",
			AssociatedMergeRequest: clients.PullRequest{
				Number: 1,
				MergedAt: time.Date(2023 /*year*/, time.March, 21, /*day*/
					13 /*hour*/, 39 /*min*/, 0 /*sec*/, 0 /*nsec*/, time.UTC),
			},
			Message: "merge commitSHA a from GitHub",
		}

		phabricatorCommitA = clients.Commit{
			Message: "\nDifferential Revision: D123\nReviewed By: user-123",
			SHA:     "abc",
		}
		phabricatorCommitAUnsquashed = clients.Commit{
			Message: "\nDifferential Revision: D123\nReviewed By: user-123",
			SHA:     "adef",
		}
		phabricatorCommitAUnsquashed2 = clients.Commit{
			Message: "\nDifferential Revision: D123\nReviewed By: user-456",
			SHA:     "afab",
		}
		phabricatorCommitB = clients.Commit{
			Message: "\nDifferential Revision: D158\nReviewed By: user-123",
			SHA:     "def",
		}
		phabricatorCommitC = clients.Commit{
			Message: "\nDifferential Revision: D2000\nReviewed By: user-456",
			SHA:     "fab",
		}
		phabricatorCommitD = clients.Commit{
			Message: "\nDifferential Revision: D2\nReviewed By: user-456",
			SHA:     "d",
		}
		phabricatorCommitE = clients.Commit{
			Message: "\nDifferential Revision: https://reviews.foo.org/D123 \nReviewed By: user-123",
			SHA:     "e",
		}
		phabricatorCommitF = clients.Commit{
			Message: "\nDifferential Revision: https://foo.bar.example.com/D456 \nReviewed By: user-123",
			SHA:     "f",
		}
		phabricatorCommitG = clients.Commit{
			Message: "\nDifferential Revision: https://reviews.bar.org/D78910 \nReviewed By: user-123",
			SHA:     "g",
		}

		gerritCommitB = clients.Commit{
			Message: "first change\nReviewed-on: server.url \nReviewed-by:user-123",
			SHA:     "abc",
		}
		gerritCommitA = clients.Commit{
			Message: "followup\nReviewed-on: server.url \nReviewed-by:user-123",
			SHA:     "def",
		}

		prowCommitSelfMerged = clients.Commit{
			SHA: "prow_self_merged",
			AssociatedMergeRequest: clients.PullRequest{
				Number: 4,
				MergedAt: time.Date(2023 /*year*/, time.March, 21, /*day*/
					13 /*hour*/, 43 /*min*/, 0 /*sec*/, 0 /*nsec*/, time.UTC),
				Author:   clients.User{Login: "kratos"},
				MergedBy: clients.User{Login: "kratos"},
				Labels:   []clients.Label{{Name: "lgtm"}},
			},
			Message: "merge commitSHA prow_self_merged from GitHub",
		}
		prowCommitSelfMergedUnsquashed = clients.Commit{
			SHA: "prow_self_merged_unsquashed",
			AssociatedMergeRequest: clients.PullRequest{
				Number: 4,
				MergedAt: time.Date(2023 /*year*/, time.March, 21, /*day*/
					13 /*hour*/, 43 /*min*/, 0 /*sec*/, 0 /*nsec*/, time.UTC),
				Author:   clients.User{Login: "kratos"},
				MergedBy: clients.User{Login: "kratos"},
				Labels:   []clients.Label{{Name: "lgtm"}},
			},
			Message: "unsquashed commitSHA prow_self_merged_unsquashed from GitHub",
		}
		prowCommitApprovedLabel = clients.Commit{
			SHA: "prow_approved_label",
			AssociatedMergeRequest: clients.PullRequest{
				Number: 5,
				MergedAt: time.Date(2023 /*year*/, time.March, 21, /*day*/
					13 /*hour*/, 44 /*min*/, 0 /*sec*/, 0 /*nsec*/, time.UTC),
				Author:   clients.User{Login: "hades"},
				MergedBy: clients.User{Login: "hades"},
				Labels:   []clients.Label{{Name: "approved"}},
			},
			Message: "merge commitSHA prow_approved_label from GitHub",
		}
		prowCommitReviewedByOther = clients.Commit{
			SHA: "prow_reviewed",
			AssociatedMergeRequest: clients.PullRequest{
				Number: 6,
				MergedAt: time.Date(2023 /*year*/, time.March, 21, /*day*/
					13 /*hour*/, 45 /*min*/, 0 /*sec*/, 0 /*nsec*/, time.UTC),
				Author:   clients.User{Login: "kratos"},
				MergedBy: clients.User{Login: "kratos"},
				Labels:   []clients.Label{{Name: "lgtm"}},
				Reviews: []clients.Review{
					{
						Author: &clients.User{Login: "athena"},
						State:  "APPROVED",
					},
				},
			},
			Message: "merge commitSHA prow_reviewed from GitHub",
		}
		unknownCommit = clients.Commit{
			SHA:     "unknown_platform",
			Message: "commit without any revision markers",
		}
		piperCommit = clients.Commit{
			Message: "a piper change\nPiperOrigin-RevId: 123456789",
			SHA:     "piper_sha",
		}
		prowCommitLabelledNotMerged = clients.Commit{
			SHA: "prow_labelled_not_merged",
			AssociatedMergeRequest: clients.PullRequest{
				Number:   7,
				Author:   clients.User{Login: "kratos"},
				MergedBy: clients.User{Login: "kratos"},
				Labels:   []clients.Label{{Name: "lgtm"}},
			},
			Message: "commitSHA prow_labelled_not_merged from GitHub",
		}
	)

	tests := []struct {
		name     string
		commits  []clients.Commit
		expected []checker.Changeset
	}{
		{
			name:    "github: merge with squash",
			commits: []clients.Commit{commitC, commitB, commitA},
			expected: []checker.Changeset{
				{
					ReviewPlatform: checker.ReviewPlatformGitHub,
					RevisionID:     "3",
					Commits:        []clients.Commit{commitC},
					Reviews: []clients.Review{
						{
							Author: &clients.User{},
							State:  "APPROVED",
						},
					},
				},
				{
					ReviewPlatform: checker.ReviewPlatformGitHub,
					RevisionID:     "2",
					Commits:        []clients.Commit{commitB},
					Reviews: []clients.Review{
						{
							Author: &clients.User{},
							State:  "APPROVED",
						},
					},
				},
				{
					ReviewPlatform: checker.ReviewPlatformGitHub,
					RevisionID:     "1",
					Commits:        []clients.Commit{commitA},
					Reviews: []clients.Review{
						{
							Author: &clients.User{},
							State:  "APPROVED",
						},
					},
				},
			},
		},
		{
			name:    "github: merge with squash reverse chronological order",
			commits: []clients.Commit{commitA, commitB, commitC},
			expected: []checker.Changeset{
				{
					ReviewPlatform: checker.ReviewPlatformGitHub,
					RevisionID:     "1",
					Commits:        []clients.Commit{commitA},
					Reviews: []clients.Review{
						{
							Author: &clients.User{},
							State:  "APPROVED",
						},
					},
				},
				{
					ReviewPlatform: checker.ReviewPlatformGitHub,
					RevisionID:     "2",
					Commits:        []clients.Commit{commitB},
					Reviews: []clients.Review{
						{
							Author: &clients.User{},
							State:  "APPROVED",
						},
					},
				},
				{
					ReviewPlatform: checker.ReviewPlatformGitHub,
					RevisionID:     "3",
					Commits:        []clients.Commit{commitC},
					Reviews: []clients.Review{
						{
							Author: &clients.User{},
							State:  "APPROVED",
						},
					},
				},
			},
		},
		{
			name:    "github: merge without squash",
			commits: []clients.Commit{commitC, commitB, commitBUnsquashed},
			expected: []checker.Changeset{
				{
					ReviewPlatform: checker.ReviewPlatformGitHub,
					RevisionID:     "3",
					Commits:        []clients.Commit{commitC},
					Reviews: []clients.Review{
						{
							Author: &clients.User{},
							State:  "APPROVED",
						},
					},
				},
				{
					ReviewPlatform: checker.ReviewPlatformGitHub,
					RevisionID:     "2",
					Commits:        []clients.Commit{commitB, commitBUnsquashed},
					Reviews: []clients.Review{
						{
							Author: &clients.User{},
							State:  "APPROVED",
						},
					},
				},
			},
		},
		{
			name:    "github: merge without squash reverse chronological order",
			commits: []clients.Commit{commitA, commitBUnsquashed, commitB, commitC},
			expected: []checker.Changeset{
				{
					ReviewPlatform: checker.ReviewPlatformGitHub,
					RevisionID:     "1",
					Commits:        []clients.Commit{commitA},
					Reviews: []clients.Review{
						{
							Author: &clients.User{},
							State:  "APPROVED",
						},
					},
				},
				{
					ReviewPlatform: checker.ReviewPlatformGitHub,
					RevisionID:     "2",
					Commits:        []clients.Commit{commitB, commitBUnsquashed},
					Reviews: []clients.Review{
						{
							Author: &clients.User{},
							State:  "APPROVED",
						},
					},
				},
				{
					ReviewPlatform: checker.ReviewPlatformGitHub,
					RevisionID:     "3",
					Commits:        []clients.Commit{commitC},
					Reviews: []clients.Review{
						{
							Author: &clients.User{},
							State:  "APPROVED",
						},
					},
				},
			},
		},
		{
			name:    "phabricator: merge with squash",
			commits: []clients.Commit{phabricatorCommitA, phabricatorCommitB, phabricatorCommitC},
			expected: []checker.Changeset{
				{
					RevisionID:     "D123",
					ReviewPlatform: checker.ReviewPlatformPhabricator,
					Commits:        []clients.Commit{phabricatorCommitA},
				},
				{
					RevisionID:     "D158",
					ReviewPlatform: checker.ReviewPlatformPhabricator,
					Commits:        []clients.Commit{phabricatorCommitB},
				},
				{
					RevisionID:     "D2000",
					ReviewPlatform: checker.ReviewPlatformPhabricator,
					Commits:        []clients.Commit{phabricatorCommitC},
				},
			},
		},
		{
			name:    "phabricator: merge without squash",
			commits: []clients.Commit{phabricatorCommitA, phabricatorCommitAUnsquashed, phabricatorCommitAUnsquashed2},
			expected: []checker.Changeset{
				{
					RevisionID:     "D123",
					ReviewPlatform: checker.ReviewPlatformPhabricator,
					Commits:        []clients.Commit{phabricatorCommitA, phabricatorCommitAUnsquashed, phabricatorCommitAUnsquashed2},
				},
			},
		},
		{
			name:    "gerrit: merge with squash",
			commits: []clients.Commit{gerritCommitB, gerritCommitA},
			expected: []checker.Changeset{
				{
					ReviewPlatform: checker.ReviewPlatformGerrit,
					RevisionID:     "abc",
					Commits:        []clients.Commit{gerritCommitB},
				},
				{
					ReviewPlatform: checker.ReviewPlatformGerrit,
					RevisionID:     "def",
					Commits:        []clients.Commit{gerritCommitA},
				},
			},
		},
		{
			name:    "mixed: phabricator + gh",
			commits: []clients.Commit{phabricatorCommitA, phabricatorCommitD, commitB, commitBUnsquashed},
			expected: []checker.Changeset{
				{
					ReviewPlatform: checker.ReviewPlatformPhabricator,
					RevisionID:     "D123",
					Commits:        []clients.Commit{phabricatorCommitA},
				},
				{
					ReviewPlatform: checker.ReviewPlatformPhabricator,
					RevisionID:     "D2",
					Commits:        []clients.Commit{phabricatorCommitD},
				},
				{
					ReviewPlatform: checker.ReviewPlatformGitHub,
					RevisionID:     "2",
					Commits:        []clients.Commit{commitB, commitBUnsquashed},
					Reviews: []clients.Review{{
						Author: &clients.User{},
						State:  "APPROVED",
					}},
				},
			},
		},
		{
			name:    "mixed: gerrit + gh",
			commits: []clients.Commit{gerritCommitB, gerritCommitA, commitC},
			expected: []checker.Changeset{
				{
					ReviewPlatform: checker.ReviewPlatformGerrit,
					RevisionID:     "abc",
					Commits:        []clients.Commit{gerritCommitB},
				},
				{
					ReviewPlatform: checker.ReviewPlatformGerrit,
					RevisionID:     "def",
					Commits:        []clients.Commit{gerritCommitA},
				},
				{
					ReviewPlatform: checker.ReviewPlatformGitHub,
					RevisionID:     "3",
					Commits:        []clients.Commit{commitC},
					Reviews: []clients.Review{
						{
							Author: &clients.User{},
							State:  "APPROVED",
						},
					},
				},
			},
		},
		{
			name:    "phabricator with URL for differential revision",
			commits: []clients.Commit{phabricatorCommitE, phabricatorCommitF, phabricatorCommitG},
			expected: []checker.Changeset{
				{
					ReviewPlatform: checker.ReviewPlatformPhabricator,
					RevisionID:     "D123",
					Commits:        []clients.Commit{phabricatorCommitE},
				},
				{
					ReviewPlatform: checker.ReviewPlatformPhabricator,
					RevisionID:     "D456",
					Commits:        []clients.Commit{phabricatorCommitF},
				},
				{
					ReviewPlatform: checker.ReviewPlatformPhabricator,
					RevisionID:     "D78910",
					Commits:        []clients.Commit{phabricatorCommitG},
				},
			},
		},
		{
			name:    "prow: label keeps the pull request author and reviews",
			commits: []clients.Commit{prowCommitSelfMerged},
			expected: []checker.Changeset{
				{
					ReviewPlatform: checker.ReviewPlatformProw,
					RevisionID:     "4",
					Commits:        []clients.Commit{prowCommitSelfMerged},
					Author:         clients.User{Login: "kratos"},
					Reviews: []clients.Review{
						{
							Author: &clients.User{Login: "kratos"},
							State:  "APPROVED",
						},
					},
				},
			},
		},
		{
			name:    "prow: label keeps the pull request author and reviews without squash",
			commits: []clients.Commit{prowCommitSelfMerged, prowCommitSelfMergedUnsquashed},
			expected: []checker.Changeset{
				{
					ReviewPlatform: checker.ReviewPlatformProw,
					RevisionID:     "4",
					Commits:        []clients.Commit{prowCommitSelfMerged, prowCommitSelfMergedUnsquashed},
					Author:         clients.User{Login: "kratos"},
					Reviews: []clients.Review{
						{
							Author: &clients.User{Login: "kratos"},
							State:  "APPROVED",
						},
					},
				},
			},
		},
		{
			name:    "prow: label keeps the pull request author and reviews in reverse chronological order",
			commits: []clients.Commit{prowCommitSelfMergedUnsquashed, prowCommitSelfMerged},
			expected: []checker.Changeset{
				{
					ReviewPlatform: checker.ReviewPlatformProw,
					RevisionID:     "4",
					Commits:        []clients.Commit{prowCommitSelfMerged, prowCommitSelfMergedUnsquashed},
					Author:         clients.User{Login: "kratos"},
					Reviews: []clients.Review{
						{
							Author: &clients.User{Login: "kratos"},
							State:  "APPROVED",
						},
					},
				},
			},
		},
		{
			name:    "prow: approved label keeps the pull request author and reviews",
			commits: []clients.Commit{prowCommitApprovedLabel},
			expected: []checker.Changeset{
				{
					ReviewPlatform: checker.ReviewPlatformProw,
					RevisionID:     "5",
					Commits:        []clients.Commit{prowCommitApprovedLabel},
					Author:         clients.User{Login: "hades"},
					Reviews: []clients.Review{
						{
							Author: &clients.User{Login: "hades"},
							State:  "APPROVED",
						},
					},
				},
			},
		},
		{
			name:    "prow: pull request reviews are kept next to the merge approval",
			commits: []clients.Commit{prowCommitReviewedByOther},
			expected: []checker.Changeset{
				{
					ReviewPlatform: checker.ReviewPlatformProw,
					RevisionID:     "6",
					Commits:        []clients.Commit{prowCommitReviewedByOther},
					Author:         clients.User{Login: "kratos"},
					Reviews: []clients.Review{
						{
							Author: &clients.User{Login: "athena"},
							State:  "APPROVED",
						},
						{
							Author: &clients.User{Login: "kratos"},
							State:  "APPROVED",
						},
					},
				},
			},
		},
		{
			name:    "mixed: prow + gh",
			commits: []clients.Commit{prowCommitSelfMerged, commitC},
			expected: []checker.Changeset{
				{
					ReviewPlatform: checker.ReviewPlatformProw,
					RevisionID:     "4",
					Commits:        []clients.Commit{prowCommitSelfMerged},
					Author:         clients.User{Login: "kratos"},
					Reviews: []clients.Review{
						{
							Author: &clients.User{Login: "kratos"},
							State:  "APPROVED",
						},
					},
				},
				{
					ReviewPlatform: checker.ReviewPlatformGitHub,
					RevisionID:     "3",
					Commits:        []clients.Commit{commitC},
					Reviews: []clients.Review{
						{
							Author: &clients.User{},
							State:  "APPROVED",
						},
					},
				},
			},
		},
		{
			name:    "unknown platform keeps no review data",
			commits: []clients.Commit{unknownCommit},
			expected: []checker.Changeset{
				{
					ReviewPlatform: checker.ReviewPlatformUnknown,
					RevisionID:     "unknown_platform",
					Commits:        []clients.Commit{unknownCommit},
				},
			},
		},
		{
			name:    "piper: changesets keep no review data (control)",
			commits: []clients.Commit{piperCommit},
			expected: []checker.Changeset{
				{
					ReviewPlatform: checker.ReviewPlatformPiper,
					RevisionID:     "123456789",
					Commits:        []clients.Commit{piperCommit},
				},
			},
		},
		{
			name:    "prow: an unmerged labelled pull request is not a prow changeset (control)",
			commits: []clients.Commit{prowCommitLabelledNotMerged},
			expected: []checker.Changeset{
				{
					ReviewPlatform: checker.ReviewPlatformUnknown,
					RevisionID:     "prow_labelled_not_merged",
					Commits:        []clients.Commit{prowCommitLabelledNotMerged},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			changesets := getChangesets(tt.commits)
			if !cmp.Equal(tt.expected, changesets,
				cmpopts.SortSlices(func(x, y checker.Changeset) bool {
					if x.RevisionID == y.RevisionID {
						return x.ReviewPlatform < y.ReviewPlatform
					}
					return x.RevisionID < y.RevisionID
				}),
				cmpopts.SortSlices(func(x, y clients.Commit) bool {
					return x.SHA < y.SHA
				})) {
				t.Log(cmp.Diff(tt.expected, changesets))
				t.Fail()
			}
		})
	}
}
