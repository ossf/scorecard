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
	"strings"

	"github.com/google/go-github/v32/github"
	"github.com/ossf/scorecard/checker"
)

func init() {
	registerCheck("Pull-Requests", PullRequests)
}

func PullRequests(c checker.Checker) checker.CheckResult {
	commits, _, err := c.Client.Repositories.ListCommits(c.Ctx, c.Owner, c.Repo, &github.CommitsListOptions{})
	if err != nil {
		return checker.RetryResult(err)
	}

	total := 0
	totalWithPrs := 0
	for _, commit := range commits {
		isBot := false
		committer := commit.GetCommitter().GetLogin()
		for _, substring := range []string{"bot", "gardener"} {
			if strings.Contains(committer, substring) {
				isBot = true
				break
			}
		}
		if isBot {
			c.Logf("skip commit from bot account: %s", committer)
			continue
		}

		total++

		// check for gerrit use via Reviewed-on
		commitMessage := commit.GetCommit().GetMessage()
		if strings.Contains(commitMessage, "\nReviewed-on: ") {
			totalWithPrs++
			c.Logf("found gerrit reviewed commit: %s", commit.GetSHA())
			continue
		}

		prs, _, err := c.Client.PullRequests.ListPullRequestsWithCommit(c.Ctx, c.Owner, c.Repo, commit.GetSHA(), &github.PullRequestListOptions{})
		if err != nil {
			return checker.RetryResult(err)
		}
		if len(prs) > 0 {
			totalWithPrs++
			c.Logf("found commit with PR: %s", commit.GetSHA())
		} else {
			c.Logf("!! found commit without PR: %s, committer: %s", commit.GetSHA(), commit.GetCommitter().GetLogin())
		}
	}
	c.Logf("found PRs for %d out of %d commits", totalWithPrs, total)
	return checker.ProportionalResult(totalWithPrs, total, .75)
}
