package checks

import (
	"strings"

	"github.com/dlorenc/scorecard/checker"
	"github.com/google/go-github/v32/github"
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

		prs, _, err := c.Client.PullRequests.ListPullRequestsWithCommit(c.Ctx, c.Owner, c.Repo, commit.GetSHA(), &github.PullRequestListOptions{})
		if err != nil {
			return checker.RetryResult(err)
		}
		total++
		if len(prs) > 0 {
			c.Logf("found PRs, example: #%v", prs[0].GetNumber())
			totalWithPrs++
		}
	}
	return checker.ProportionalResult(totalWithPrs, total, .75)
}
