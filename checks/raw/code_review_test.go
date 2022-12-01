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
	"strings"
	"testing"
	"time"

	"golang.org/x/exp/slices"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
)

//nolint:gocritic
func assertCommitEq(t *testing.T, actual clients.Commit, expected clients.Commit) {
	t.Helper()

	if actual.SHA != expected.SHA {
		t.Fatalf("commit shas mismatched\na.sha: %s\nb.sha %s", actual.SHA, expected.SHA)
	}
}

func assertChangesetEq(t *testing.T, actual, expected *checker.Changeset) {
	t.Helper()

	if actual.ReviewPlatform != expected.ReviewPlatform {
		t.Fatalf(
			"changeset review platform mismatched\na.platform: %s\nb.platform: %s",
			actual.ReviewPlatform,
			expected.ReviewPlatform,
		)
	}

	if actual.RevisionID != expected.RevisionID {
		t.Fatalf(
			"changeset revisionID mismatched\na.revid: %s\nb.revid %s",
			actual.RevisionID,
			expected.RevisionID,
		)
	}

	if len(actual.Commits) != len(expected.Commits) {
		t.Fatalf(
			"changesets contain different numbers of commits\na:%d\nb:%d",
			len(actual.Commits),
			len(expected.Commits),
		)
	}

	for i := range actual.Commits {
		assertCommitEq(t, actual.Commits[i], expected.Commits[i])
	}
}

//nolint:gocritic
func csless(a, b checker.Changeset) bool {
	if cmp := strings.Compare(a.RevisionID, b.RevisionID); cmp != 0 {
		return cmp < 0
	}

	return a.ReviewPlatform < b.ReviewPlatform
}

func assertChangesetArrEq(t *testing.T, actual, expected []checker.Changeset) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Fatalf("different number of changesets\na:%d\nb:%d", len(actual), len(expected))
	}

	slices.SortFunc(actual, csless)
	slices.SortFunc(expected, csless)

	for i := range actual {
		assertChangesetEq(t, &actual[i], &expected[i])
	}
}

// TestCodeReviews tests the CodeReviews function.
func Test_getChangesets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		commits  []clients.Commit
		expected []checker.Changeset
	}{
		{
			name: "commit history squashed during merge",
			commits: []clients.Commit{
				{
					SHA:                    "a",
					AssociatedMergeRequest: clients.PullRequest{Number: 3, MergedAt: time.Now()},
				},
				{
					SHA:                    "b",
					AssociatedMergeRequest: clients.PullRequest{Number: 2, MergedAt: time.Now()},
				},
				{
					SHA:                    "c",
					AssociatedMergeRequest: clients.PullRequest{Number: 1, MergedAt: time.Now()},
				},
			},
			expected: []checker.Changeset{
				{
					ReviewPlatform: checker.ReviewPlatformGitHub,
					RevisionID:     "3",
					Commits:        []clients.Commit{{SHA: "a"}},
				},
				{
					ReviewPlatform: checker.ReviewPlatformGitHub,
					RevisionID:     "2",
					Commits:        []clients.Commit{{SHA: "b"}},
				},
				{
					ReviewPlatform: checker.ReviewPlatformGitHub,
					RevisionID:     "1",
					Commits:        []clients.Commit{{SHA: "c"}},
				},
			},
		},
		{
			name: "commits in reverse chronological order",
			commits: []clients.Commit{
				{
					SHA:                    "a",
					AssociatedMergeRequest: clients.PullRequest{Number: 1, MergedAt: time.Now()},
				},
				{
					SHA:                    "b",
					AssociatedMergeRequest: clients.PullRequest{Number: 2, MergedAt: time.Now()},
				},
				{
					SHA:                    "c",
					AssociatedMergeRequest: clients.PullRequest{Number: 3, MergedAt: time.Now()},
				},
			},
			expected: []checker.Changeset{
				{
					ReviewPlatform: checker.ReviewPlatformGitHub,
					RevisionID:     "1",
					Commits: []clients.Commit{
						{
							SHA:                    "a",
							AssociatedMergeRequest: clients.PullRequest{Number: 1},
						},
					},
				},
				{
					ReviewPlatform: checker.ReviewPlatformGitHub,
					RevisionID:     "2",
					Commits: []clients.Commit{
						{
							SHA:                    "b",
							AssociatedMergeRequest: clients.PullRequest{Number: 2},
						},
					},
				},
				{
					ReviewPlatform: checker.ReviewPlatformGitHub,
					RevisionID:     "3",
					Commits: []clients.Commit{
						{
							SHA:                    "c",
							AssociatedMergeRequest: clients.PullRequest{Number: 3},
						},
					},
				},
			},
		},
		{
			name: "without commit squashing",
			commits: []clients.Commit{
				{
					SHA:                    "foo",
					AssociatedMergeRequest: clients.PullRequest{Number: 1, MergedAt: time.Now()},
				},
				{
					SHA:                    "bar",
					AssociatedMergeRequest: clients.PullRequest{Number: 2, MergedAt: time.Now()},
				},
				{
					SHA:                    "baz",
					AssociatedMergeRequest: clients.PullRequest{Number: 2, MergedAt: time.Now()},
				},
			},
			expected: []checker.Changeset{
				{
					ReviewPlatform: checker.ReviewPlatformGitHub,
					RevisionID:     "1",
					Commits: []clients.Commit{
						{
							SHA:                    "foo",
							AssociatedMergeRequest: clients.PullRequest{Number: 1},
						},
					},
				},
				{
					ReviewPlatform: checker.ReviewPlatformGitHub,
					RevisionID:     "2",
					Commits: []clients.Commit{
						{
							SHA:                    "bar",
							AssociatedMergeRequest: clients.PullRequest{Number: 2},
						},
						{
							SHA:                    "baz",
							AssociatedMergeRequest: clients.PullRequest{Number: 2},
						},
					},
				},
			},
		},
		{
			name: "some commits from external scm",
			commits: []clients.Commit{
				{
					Message: "\nDifferential Revision: 123\nReviewed By: user-123",
					SHA:     "abc",
				},
				{
					Message: "\nDifferential Revision: 158\nReviewed By: user-456",
					SHA:     "def",
				},
				{
					Message:                "this one from github, but unrelated to prev",
					AssociatedMergeRequest: clients.PullRequest{Number: 158, MergedAt: time.Now()},
					SHA:                    "fab",
				},
				{
					Message:                "another from gh",
					AssociatedMergeRequest: clients.PullRequest{Number: 158, MergedAt: time.Now()},
					SHA:                    "dab",
				},
			},
			expected: []checker.Changeset{
				{
					ReviewPlatform: checker.ReviewPlatformPhabricator,
					RevisionID:     "123",
					Commits:        []clients.Commit{{SHA: "abc"}},
				},
				{
					ReviewPlatform: checker.ReviewPlatformPhabricator,
					RevisionID:     "158",
					Commits:        []clients.Commit{{SHA: "def"}},
				},
				{
					ReviewPlatform: checker.ReviewPlatformGitHub,
					RevisionID:     "158",
					Commits: []clients.Commit{
						{SHA: "fab"},
						{SHA: "dab"},
					},
				},
			},
		},
		{
			name: "some commits from external scm with no revision id",
			commits: []clients.Commit{
				{
					Message: "first change\nReviewed-on: server.url \nReviewed-by:user-123",
					SHA:     "abc",
				},
				{
					Message: "followup\nReviewed-on: server.url \nReviewed-by:user-123",
					SHA:     "def",
				},
				{
					Message:                "commit thru gh",
					AssociatedMergeRequest: clients.PullRequest{Number: 3, MergedAt: time.Now()},
					SHA:                    "fab",
				},
			},
			expected: []checker.Changeset{
				{
					ReviewPlatform: checker.ReviewPlatformGerrit,
					RevisionID:     "abc",
					Commits:        []clients.Commit{{SHA: "abc"}},
				},
				{
					ReviewPlatform: checker.ReviewPlatformGerrit,
					RevisionID:     "def",
					Commits:        []clients.Commit{{SHA: "def"}},
				},
				{
					ReviewPlatform: checker.ReviewPlatformGitHub,
					RevisionID:     "3",
					Commits:        []clients.Commit{{SHA: "fab"}},
				},
			},
		},
		{
			name: "external scm mirrored to github",
			commits: []clients.Commit{
				{
					Message: "\nDifferential Revision: 123\nReviewed By: user-123",
					SHA:     "abc",
				},
				{
					Message: "\nDifferential Revision: 158\nReviewed By: user-123",
					SHA:     "def",
				},
				{
					Message: "\nDifferential Revision: 2000\nReviewed By: user-456",
					SHA:     "fab",
				},
			},
			expected: []checker.Changeset{
				{
					RevisionID:     "123",
					ReviewPlatform: checker.ReviewPlatformPhabricator,
					Commits:        []clients.Commit{{SHA: "abc"}},
				},
				{
					RevisionID:     "158",
					ReviewPlatform: checker.ReviewPlatformPhabricator,
					Commits:        []clients.Commit{{SHA: "def"}},
				},
				{
					RevisionID:     "2000",
					ReviewPlatform: checker.ReviewPlatformPhabricator,
					Commits:        []clients.Commit{{SHA: "fab"}},
				},
			},
		},
		{
			name: "external scm no squash",
			commits: []clients.Commit{
				{
					Message: "\nDifferential Revision: 123\nReviewed By: user-123",
					SHA:     "abc",
				},
				{
					Message: "\nDifferential Revision: 123\nReviewed By: user-123",
					SHA:     "def",
				},
				{
					Message: "\nDifferential Revision: 123\nReviewed By: user-456",
					SHA:     "fab",
				},
			},
			expected: []checker.Changeset{
				{
					RevisionID:     "123",
					ReviewPlatform: checker.ReviewPlatformPhabricator,
					Commits:        []clients.Commit{{SHA: "abc"}, {SHA: "def"}, {SHA: "fab"}},
				},
			},
		},
		{
			name: "single changeset",
			commits: []clients.Commit{
				{
					SHA:                    "abc",
					AssociatedMergeRequest: clients.PullRequest{Number: 1, MergedAt: time.Now()},
				},
			},
			expected: []checker.Changeset{
				{
					ReviewPlatform: checker.ReviewPlatformGitHub,
					RevisionID:     "1",
					Commits:        []clients.Commit{{SHA: "abc"}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Logf("test: %s", tt.name)
		changesets := getChangesets(tt.commits)
		assertChangesetArrEq(t, changesets, tt.expected)
	}
}
