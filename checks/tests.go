package checks

import (
	"strings"

	"github.com/dlorenc/scorecard/checker"
	"github.com/google/go-github/v32/github"
)

func init() {
	AllChecks = append(AllChecks, NamedCheck{
		Name: "CI-Tests",
		Fn:   GithubChecks,
	})
}

func GithubChecks(c *checker.Checker) CheckResult {
	prs, _, err := c.Client.PullRequests.List(c.Ctx, c.Owner, c.Repo, &github.PullRequestListOptions{
		State: "closed",
	})
	if err != nil {
		return RetryResult(err)
	}

	totalMerged := 0
	totalTested := 0
	for _, pr := range prs {
		if pr.MergedAt == nil {
			continue
		}
		totalMerged++
		statuses, _, err := c.Client.Repositories.ListStatuses(c.Ctx, c.Owner, c.Repo, pr.GetHead().GetSHA(), &github.ListOptions{})
		if err != nil {
			return RetryResult(err)
		}
		for _, status := range statuses {
			if status.GetState() != "success" {
				continue
			}
			c := status.GetContext()
			hadTest := false
			for _, pattern := range []string{"travis-ci", "buildkite", "e2e"} {
				if strings.Contains(c, pattern) {
					hadTest = true
					break
				}
			}
			if hadTest {
				totalTested++
				break
			}
		}
	}
	// Threshold is 3/4 of merged PRs
	actual := float32(totalTested) / float32(totalMerged)
	if actual >= .75 {
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
