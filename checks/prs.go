package checks

import (
	"github.com/dlorenc/scorecard/checker"
	"github.com/google/go-github/v32/github"
)

func init() {
	AllChecks = append(AllChecks, NamedCheck{
		Name: "PullRequests",
		Fn:   PullRequests,
	})
}

func PullRequests(c *checker.Checker) CheckResult {
	commits, _, err := c.Client.Repositories.ListCommits(c.Ctx, c.Owner, c.Repo, &github.CommitsListOptions{})
	if err != nil {
		return RetryResult(err)
	}

	total := 0
	totalWithPrs := 0
	for _, commit := range commits {
		prs, _, err := c.Client.PullRequests.ListPullRequestsWithCommit(c.Ctx, c.Owner, c.Repo, commit.GetSHA(), &github.PullRequestListOptions{})
		if err != nil {
			return RetryResult(err)
		}
		total++
		if len(prs) > 0 {
			totalWithPrs++
		}
	}
	actual := float32(totalWithPrs) / float32(total)
	if actual >= .9 {
		return CheckResult{
			Pass:       true,
			Confidence: int(actual * 10),
		}
	}
	return CheckResult{
		Pass:       false,
		Confidence: int(10 - int(actual*10)),
	}
}
