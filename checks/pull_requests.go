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
	"fmt"
	"strings"

	"github.com/google/go-github/v32/github"

	"github.com/ossf/scorecard/v2/checker"
	sce "github.com/ossf/scorecard/v2/errors"
)

// CheckPullRequests is the registered name for PullRequests.
const CheckPullRequests = "Pull-Requests"

//nolint:gochecknoinits
func init() {
	registerCheck(CheckPullRequests, PullRequests)
}

func PullRequests(c *checker.CheckRequest) checker.CheckResult {
	commits, _, err := c.Client.Repositories.ListCommits(c.Ctx, c.Owner, c.Repo, &github.CommitsListOptions{})
	if err != nil {
		e := sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("Client.Repositories.ListCommits: %v", err))
		return checker.CreateRuntimeErrorResult(CheckPullRequests, e)
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
			c.Dlogger.Debug("skip commit from bot account: %s", committer)
			continue
		}

		total++

		// check for gerrit use via Reviewed-on
		commitMessage := commit.GetCommit().GetMessage()
		if strings.Contains(commitMessage, "\nReviewed-on: ") {
			totalWithPrs++
			c.Dlogger.Debug("Gerrit reviewed commit: %s", commit.GetSHA())
			continue
		}

		prs, _, err := c.Client.PullRequests.ListPullRequestsWithCommit(c.Ctx, c.Owner, c.Repo, commit.GetSHA(),
			&github.PullRequestListOptions{})
		if err != nil {
			e := sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("Client.PullRequests.ListPullRequestsWithCommit: %v", err))
			return checker.CreateRuntimeErrorResult(CheckPullRequests, e)
		}
		if len(prs) > 0 {
			totalWithPrs++
			c.Dlogger.Debug("commit with PR: %s", commit.GetSHA())
		} else {
			c.Dlogger.Debug("found commit without PR: %s, committer: %s", commit.GetSHA(), commit.GetCommitter().GetLogin())
		}
	}

	reason := fmt.Sprintf("%d ouf of %d commits have a PR", totalWithPrs, total)
	return checker.CreateProportionalScoreResult(CheckPullRequests, reason, totalWithPrs, total)
}
